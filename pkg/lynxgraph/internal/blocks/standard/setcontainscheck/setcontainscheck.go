package setcontainscheck

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/zeromicro/go-zero/core/logx"
)

const SetContainsCheckVersion = "v1"

// Config SetContainsCheck 积木配置
type Config struct {
	SetContextKey     string `json:"setContextKey" check:"must"` // 集合数据在 GraphContext 中的 key（GetDedupSet 的输出）
	CheckField        string `json:"checkField" check:"must"`    // 去重字段名（如 "person_name"），用于从 records 中匹配
	ValueExpr         string `json:"valueExpr" check:"must"`     // 检查值表达式（如 "{{atom.name}}"）
	TerminateIfExists *bool  `json:"terminateIfExists"`          // 已存在则终止（默认 true）
	SaveResultTo      string `json:"saveResultTo"`               // 可选，保存检查结果到 GraphContext
}

// SetContainsCheckBlock 集合包含检查积木
type SetContainsCheckBlock struct {
	core.BaseLogicBlock
}

// NewSetContainsCheckBlock 创建一个新的 SetContainsCheckBlock 实例
func NewSetContainsCheckBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &SetContainsCheckBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeSetContainsCheck,
			RawConfig: config,
		},
	}, nil
}

func (b *SetContainsCheckBlock) GetID() string                { return b.ID }
func (b *SetContainsCheckBlock) GetType() core.LogicBlockType { return b.Type }
func (b *SetContainsCheckBlock) GetConfigure() map[string]any { return b.RawConfig }

func (b *SetContainsCheckBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 设置默认值
	if cfg.TerminateIfExists == nil {
		defaultTrue := true
		cfg.TerminateIfExists = &defaultTrue
	}

	b.TypedConfig = &cfg
	return nil
}

// Execute 执行集合包含检查逻辑
func (b *SetContainsCheckBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	config, ok := b.TypedConfig.(*Config)
	if !ok {
		return false, core.ErrInvalidConfig
	}

	tenantId := execCtx.GetTenantId()
	graphKey := execCtx.GetGraphKey()
	atom := execCtx.GetInfoAtom()

	// 1. 从 GraphContext 获取集合数据
	setContext, err := datastore.GetGraphContext(ctx, tenantId, graphKey, config.SetContextKey)
	if err != nil {
		if err == core.ErrItemNotFound {
			logx.WithContext(ctx).Infof("[SetContainsCheck] 集合数据不存在: %s，视为不包含", config.SetContextKey)
			return true, nil
		}
		return false, fmt.Errorf("[SetContainsCheck] 获取集合数据失败: %w", err)
	}

	payload := setContext.GetPayload()
	if payload == nil {
		logx.WithContext(ctx).Infof("[SetContainsCheck] 集合数据为空: %s，视为不包含", config.SetContextKey)
		return true, nil
	}

	// 2. 获取 records 数组
	recordsRaw, ok := payload["records"]
	if !ok {
		logx.WithContext(ctx).Infof("[SetContainsCheck] 集合数据缺少 records 字段: %s，视为不包含", config.SetContextKey)
		return true, nil
	}

	records, ok := recordsRaw.([]any)
	if !ok {
		// 尝试转换为 []map[string]string
		if recordsTyped, ok := recordsRaw.([]map[string]string); ok {
			records = make([]any, len(recordsTyped))
			for i, r := range recordsTyped {
				records[i] = r
			}
		} else {
			return false, fmt.Errorf("[SetContainsCheck] records 字段类型错误: %T", recordsRaw)
		}
	}

	// 3. 解析 valueExpr 获取要检查的值
	checkValue := resolveExpression(config.ValueExpr, atom)
	logx.WithContext(ctx).Debugf("[SetContainsCheck] 检查值: %s (来自表达式 %s)", checkValue, config.ValueExpr)

	// 4. 遍历 records 检查是否包含
	exists := false
	for _, recordRaw := range records {
		record, ok := toStringMap(recordRaw)
		if !ok {
			continue
		}

		if fieldValue, ok := record[config.CheckField]; ok && fieldValue == checkValue {
			exists = true
			break
		}
	}

	logx.WithContext(ctx).Infof("[SetContainsCheck] 检查结果: %s=%s, 存在=%v", config.CheckField, checkValue, exists)

	// 5. 可选：保存检查结果
	if config.SaveResultTo != "" {
		resultPayload := map[string]any{
			"exists":     exists,
			"checkField": config.CheckField,
			"checkValue": checkValue,
		}

		resultContext := &core.BaseGraphContext{
			TenantId:   tenantId,
			GraphKey:   graphKey,
			ContextKey: config.SaveResultTo,
			Payload:    resultPayload,
		}

		if err := datastore.SaveGraphContext(ctx, resultContext); err != nil {
			logx.WithContext(ctx).Errorf("[SetContainsCheck] 保存检查结果失败: %v", err)
		}
	}

	// 6. 根据配置决定是否终止流程
	if exists && *config.TerminateIfExists {
		return false, nil // 终止流程
	}

	return true, nil
}

// resolveExpression 解析表达式，从 InfoAtom 获取值
// 支持 {{atom.xxx}} 格式
func resolveExpression(expr string, atom core.InfoAtom) string {
	if atom == nil {
		return expr
	}

	// 匹配 {{atom.xxx}} 格式
	re := regexp.MustCompile(`\{\{atom\.([^}]+)\}\}`)
	result := re.ReplaceAllStringFunc(expr, func(match string) string {
		// 提取字段名
		fieldName := strings.TrimPrefix(match, "{{atom.")
		fieldName = strings.TrimSuffix(fieldName, "}}")

		payload := atom.GetPayload()
		if payload == nil {
			return ""
		}

		// 支持嵌套路径（如 employee.name）
		parts := strings.Split(fieldName, ".")
		var current any = payload

		for _, part := range parts {
			if m, ok := current.(map[string]any); ok {
				current = m[part]
			} else {
				return ""
			}
		}

		if current == nil {
			return ""
		}

		return fmt.Sprintf("%v", current)
	})

	return result
}

// toStringMap 将 any 转换为 map[string]string
func toStringMap(v any) (map[string]string, bool) {
	switch m := v.(type) {
	case map[string]string:
		return m, true
	case map[string]any:
		result := make(map[string]string)
		for k, val := range m {
			result[k] = fmt.Sprintf("%v", val)
		}
		return result, true
	default:
		return nil, false
	}
}

// GetSetContainsCheckSpec 返回 SetContainsCheckBlock 的规格
func GetSetContainsCheckSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"SetContainsCheck",
		SetContainsCheckVersion,
		"检查集合是否包含指定值，配合 GetDedupSet 和 ForEach 使用",
		[]string{"input", "output"},
		map[string]any{
			"setContextKey": map[string]any{
				"type":        "string",
				"title":       "集合数据位置",
				"description": "选择 GetDedupSet 积木保存结果的位置",
				"check":       "must",
				"source":      "contextKeys",
				"placeholder": "checked_employees",
			},
			"checkField": map[string]any{
				"type":        "string",
				"title":       "检查字段",
				"description": "去重字段名，用于在 records 中匹配",
				"check":       "must",
				"placeholder": "person_name",
			},
			"valueExpr": map[string]any{
				"type":        "string",
				"title":       "检查值表达式",
				"description": "从信息原子获取检查值的表达式，如 {{atom.name}}",
				"check":       "must",
				"placeholder": "{{atom.name}}",
			},
			"terminateIfExists": map[string]any{
				"type":        "boolean",
				"title":       "已存在时终止",
				"description": "如果值已存在于集合中，是否终止流程",
				"default":     true,
			},
			"saveResultTo": map[string]any{
				"type":        "string",
				"title":       "结果保存位置",
				"description": "（可选）保存检查结果到 GraphContext 的 key",
				"placeholder": "contains_check_result",
			},
		},
		[]core.ServiceType{},
	)
}
