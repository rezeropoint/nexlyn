package querydatabase

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/zeromicro/go-zero/core/logx"
)

const QueryDatabaseVersion = "v1"

// ExternalDBService 外部数据库服务接口（避免直接依赖 service 包，防止循环引用）
type ExternalDBService interface {
	Query(ctx context.Context, dataSource, query string, args ...any) ([]map[string]any, error)
}

// Config QueryDatabase 积木配置
type Config struct {
	DataSource   string   `json:"dataSource" check:"must"`   // 数据源名称（引用配置中的外部数据源）
	Query        string   `json:"query" check:"must"`        // SQL 查询（仅支持 SELECT）
	Params       []string `json:"params"`                    // 参数列表，按顺序对应 $1, $2...，支持 {{atom.xxx}} 变量
	SaveResultTo string   `json:"saveResultTo" check:"must"` // 保存到 GraphContext 的键名
	TimeoutSec   int      `json:"timeoutSec"`                // 查询超时（秒），默认 30
	MaxRows      int      `json:"maxRows"`                   // 最大返回行数，默认 1000，上限 10000
	SingleRow    bool     `json:"singleRow"`                 // 只返回第一行结果（简化后续访问）
}

// QueryDatabaseBlock 外部数据库查询积木
type QueryDatabaseBlock struct {
	core.BaseLogicBlock
}

// NewQueryDatabaseBlock 创建实例
func NewQueryDatabaseBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &QueryDatabaseBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeQueryDatabase,
			RawConfig: config,
		},
	}, nil
}

func (b *QueryDatabaseBlock) GetID() string                { return b.ID }
func (b *QueryDatabaseBlock) GetType() core.LogicBlockType { return b.Type }
func (b *QueryDatabaseBlock) GetConfigure() map[string]any { return b.RawConfig }

func (b *QueryDatabaseBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 设置默认值
	if cfg.TimeoutSec <= 0 {
		cfg.TimeoutSec = 30
	}
	if cfg.MaxRows <= 0 {
		cfg.MaxRows = 1000
	}
	if cfg.MaxRows > 10000 {
		cfg.MaxRows = 10000
	}

	b.TypedConfig = &cfg
	return nil
}

// Execute 执行数据库查询
func (b *QueryDatabaseBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	config, ok := b.TypedConfig.(*Config)
	if !ok {
		return false, core.ErrInvalidConfig
	}

	// 获取外部数据库服务
	dbServiceRaw, err := service.GetByType(core.ServiceTypeExternalDB)
	if err != nil {
		return false, fmt.Errorf("获取外部数据库服务失败: %w", err)
	}

	dbService, ok := dbServiceRaw.(ExternalDBService)
	if !ok {
		return false, fmt.Errorf("服务类型不匹配，期望 ExternalDBService")
	}

	// 获取 InfoAtom 用于变量替换
	infoAtom := execCtx.GetInfoAtom()

	// 处理参数：替换变量（支持 {{atom.xxx}}、{{timestamp}}、{{date}}）
	args := make([]any, len(config.Params))
	for i, param := range config.Params {
		args[i] = replaceVariables(param, infoAtom)
	}

	// 创建带超时的上下文
	queryCtx, cancel := context.WithTimeout(ctx, time.Duration(config.TimeoutSec)*time.Second)
	defer cancel()

	// 执行查询
	logx.WithContext(ctx).Infof("[QueryDatabase] 执行查询: dataSource=%s, query=%s, args=%v",
		config.DataSource, truncateString(config.Query, 100), args)

	results, err := dbService.Query(queryCtx, config.DataSource, config.Query, args...)
	if err != nil {
		logx.WithContext(ctx).Errorf("[QueryDatabase] 查询失败: %v", err)
		return false, fmt.Errorf("数据库查询失败: %w", err)
	}

	// 限制返回行数
	if len(results) > config.MaxRows {
		results = results[:config.MaxRows]
	}

	// 构建 Payload
	payload := map[string]any{
		"rowCount": len(results),
	}

	if config.SingleRow {
		// singleRow 模式：将第一行的字段直接提升到顶层，简化边条件访问
		// 例如：leave_check.on_leave 而不是 leave_check.data.on_leave
		if len(results) > 0 {
			for k, v := range results[0] {
				payload[k] = v
			}
		}
	} else {
		// 非 singleRow 模式：结果放在 data 字段中
		payload["data"] = results
	}

	// 保存到 GraphContext
	graphContext := &core.BaseGraphContext{
		TenantId:   execCtx.GetTenantId(),
		GraphKey:   execCtx.GetGraphKey(),
		ContextKey: config.SaveResultTo,
		Payload:    payload,
	}

	if err := datastore.SaveGraphContext(ctx, graphContext); err != nil {
		logx.WithContext(ctx).Errorf("[QueryDatabase] 保存结果失败: %v", err)
		return false, fmt.Errorf("保存查询结果失败: %w", err)
	}

	logx.WithContext(ctx).Infof("[QueryDatabase] 查询完成，返回 %d 行，保存到 %s",
		len(results), config.SaveResultTo)

	return true, nil
}

// replaceVariables 替换模板变量
// 支持的变量：
//   - {{atom.xxx}} - 从 InfoAtom Payload 获取字段值
//   - {{timestamp}} - InfoAtom 的毫秒时间戳
//   - {{date}} - InfoAtom 时间戳转换为日期字符串（YYYY-MM-DD，Asia/Shanghai 时区）
//   - {{datetime}} - InfoAtom 时间戳转换为日期时间字符串（YYYY-MM-DD HH:MM:SS）
func replaceVariables(template string, infoAtom core.InfoAtom) any {
	if template == "" {
		return ""
	}

	payload := infoAtom.GetPayload()
	timestamp := infoAtom.GetTimestamp()

	// 检查是否整个字符串就是一个变量（保持原始类型）
	// 1. {{atom.xxx}} - 返回 payload 中的原始类型
	reAtom := regexp.MustCompile(`^\{\{atom\.(\w+)\}\}$`)
	if match := reAtom.FindStringSubmatch(template); match != nil {
		fieldName := match[1]
		if value, ok := payload[fieldName]; ok {
			return value
		}
		return template
	}

	// 2. {{timestamp}} - 返回毫秒时间戳（int64）
	if template == "{{timestamp}}" {
		return timestamp
	}

	// 3. {{date}} - 返回日期字符串 YYYY-MM-DD
	if template == "{{date}}" {
		return formatTimestampToDate(timestamp)
	}

	// 4. {{datetime}} - 返回日期时间字符串 YYYY-MM-DD HH:MM:SS
	if template == "{{datetime}}" {
		return formatTimestampToDatetime(timestamp)
	}

	// 部分替换：将所有变量替换为字符串
	result := template

	// 替换 {{atom.xxx}}
	reAtomPartial := regexp.MustCompile(`\{\{atom\.(\w+)\}\}`)
	result = reAtomPartial.ReplaceAllStringFunc(result, func(match string) string {
		fieldName := reAtomPartial.FindStringSubmatch(match)[1]
		if value, ok := payload[fieldName]; ok {
			return fmt.Sprintf("%v", value)
		}
		return match
	})

	// 替换 {{timestamp}}
	result = regexp.MustCompile(`\{\{timestamp\}\}`).ReplaceAllString(result, fmt.Sprintf("%d", timestamp))

	// 替换 {{date}}
	result = regexp.MustCompile(`\{\{date\}\}`).ReplaceAllString(result, formatTimestampToDate(timestamp))

	// 替换 {{datetime}}
	result = regexp.MustCompile(`\{\{datetime\}\}`).ReplaceAllString(result, formatTimestampToDatetime(timestamp))

	return result
}

// formatTimestampToDate 将毫秒时间戳转换为日期字符串（YYYY-MM-DD）
func formatTimestampToDate(timestampMs int64) string {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	t := time.UnixMilli(timestampMs).In(loc)
	return t.Format("2006-01-02")
}

// formatTimestampToDatetime 将毫秒时间戳转换为日期时间字符串（YYYY-MM-DD HH:MM:SS）
func formatTimestampToDatetime(timestampMs int64) string {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	t := time.UnixMilli(timestampMs).In(loc)
	return t.Format("2006-01-02 15:04:05")
}

// truncateString 截断字符串用于日志
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GetQueryDatabaseSpec 返回积木规格
func GetQueryDatabaseSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"QueryDatabase",
		QueryDatabaseVersion,
		"查询外部数据库（仅支持 SELECT），支持参数化查询防止 SQL 注入",
		[]string{"data", "query", "database"},
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"dataSource": map[string]any{
					"type":        "string",
					"title":       "数据源名称",
					"description": "配置的外部数据源名称（如 skylark）",
					"check":       "must",
				},
				"query": map[string]any{
					"type":        "string",
					"title":       "SQL 查询",
					"description": "SELECT 查询语句，使用 $1, $2... 作为参数占位符",
					"check":       "must",
					"format":      "sql", // 前端校验：仅允许 SELECT 语句
					"placeholder": "SELECT EXISTS (\n  SELECT 1 FROM assignments\n  WHERE user_id = $1 AND status = 'active'\n) AS is_active",
				},
				"params": map[string]any{
					"type":        "array",
					"title":       "查询参数",
					"description": "按顺序对应 $1, $2... 的参数值。支持变量：{{atom.xxx}}（Payload字段）、{{timestamp}}（毫秒时间戳）、{{date}}（YYYY-MM-DD）、{{datetime}}（YYYY-MM-DD HH:MM:SS）",
					"itemAddable": true, // 前端渲染为动态添加列表
					"items": map[string]any{
						"type":        "string",
						"placeholder": "{{atom.name}} 或 {{date}}",
					},
				},
				"saveResultTo": map[string]any{
					"type":        "string",
					"title":       "结果保存键名",
					"description": "保存查询结果到 GraphContext 的键名",
					"check":       "must",
				},
				"timeoutSec": map[string]any{
					"type":        "integer",
					"title":       "查询超时",
					"description": "查询超时时间（秒），默认 30",
					"default":     30,
				},
				"maxRows": map[string]any{
					"type":        "integer",
					"title":       "最大行数",
					"description": "最大返回行数，默认 1000，上限 10000",
					"default":     1000,
				},
				"singleRow": map[string]any{
					"type":        "boolean",
					"title":       "单行模式",
					"description": "只返回第一行结果（简化后续访问）",
					"default":     false,
				},
			},
			"required": []string{"dataSource", "query", "saveResultTo"},
		},
		[]core.ServiceType{core.ServiceTypeExternalDB},
	).WithOutputSchema(map[string]any{
		"contextKey": "config.saveResultTo", // 上下文键名来自配置
		"type":       "object",
		"properties": map[string]any{
			"data": map[string]any{
				"type":        "query-result",                          // 查询结果，字段取决于 SQL 列
				"description": "查询结果（单行模式为对象，否则为数组）",
			},
			"rowCount": map[string]any{
				"type":        "number",
				"description": "返回的行数",
			},
		},
	})
}
