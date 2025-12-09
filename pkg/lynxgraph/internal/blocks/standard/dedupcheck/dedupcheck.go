package dedupcheck

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/zeromicro/go-zero/core/logx"
)

const DedupCheckVersion = "v1"

// ResetMode 重置模式
type ResetMode string

const (
	ResetModeDaily  ResetMode = "daily"  // 每天指定时间重置
	ResetModeHourly ResetMode = "hourly" // 每小时重置
	ResetModeNone   ResetMode = "none"   // 不自动重置，完全依赖 GraphContextTTL
)

// Config 去重检查积木配置
type Config struct {
	DedupFields       []string  `json:"dedupFields" check:"must"` // 去重字段列表，从信息原子 payload 中提取
	ResetMode         ResetMode `json:"resetMode"`                // 重置模式：daily/hourly/none，默认 daily
	ResetTime         string    `json:"resetTime"`                // 每日重置时间，格式 HH:MM，默认 "00:00"
	Timezone          string    `json:"timezone"`                 // 时区，默认 Asia/Shanghai
	TerminateIfExists *bool     `json:"terminateIfExists"`        // 已存在时终止流程（默认 true）
}

// DedupCheckBlock 去重检查积木
type DedupCheckBlock struct {
	core.BaseLogicBlock
}

// NewDedupCheckBlock 创建一个新的 DedupCheckBlock 实例
func NewDedupCheckBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &DedupCheckBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeDedupCheck,
			RawConfig: config,
		},
	}, nil
}

func (b *DedupCheckBlock) GetID() string                { return b.ID }
func (b *DedupCheckBlock) GetType() core.LogicBlockType { return b.Type }
func (b *DedupCheckBlock) GetConfigure() map[string]any { return b.RawConfig }

func (b *DedupCheckBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 验证去重字段不能为空
	if len(cfg.DedupFields) == 0 {
		return fmt.Errorf("dedupFields 不能为空")
	}

	// 设置默认值
	if cfg.ResetMode == "" {
		cfg.ResetMode = ResetModeDaily
	}
	if cfg.ResetTime == "" {
		cfg.ResetTime = "00:00"
	}
	if cfg.Timezone == "" {
		cfg.Timezone = "Asia/Shanghai"
	}
	if cfg.TerminateIfExists == nil {
		defaultTrue := true
		cfg.TerminateIfExists = &defaultTrue
	}

	// 验证 resetMode
	if cfg.ResetMode != ResetModeDaily && cfg.ResetMode != ResetModeHourly && cfg.ResetMode != ResetModeNone {
		return fmt.Errorf("resetMode 必须是 daily、hourly 或 none")
	}

	// 验证 resetTime 格式
	if cfg.ResetMode == ResetModeDaily {
		if _, err := time.Parse("15:04", cfg.ResetTime); err != nil {
			return fmt.Errorf("resetTime 格式无效，应为 HH:MM")
		}
	}

	b.TypedConfig = &cfg
	return nil
}

// Execute 执行去重检查逻辑
func (b *DedupCheckBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	config, ok := b.TypedConfig.(*Config)
	if !ok {
		return false, core.ErrInvalidConfig
	}

	// 1. 获取当前时间（用于时间周期计算）
	loc, err := time.LoadLocation(config.Timezone)
	if err != nil {
		logx.WithContext(ctx).Debugf("[DedupCheck] 无效的时区 %s，使用 UTC", config.Timezone)
		loc = time.UTC
	}
	now := time.Now().In(loc)

	// 2. 计算时间周期前缀
	timePeriod := getTimePeriod(now, config.ResetMode, config.ResetTime)

	// 3. 从信息原子 payload 中提取去重字段值，生成 contextKey
	fieldKey := buildFieldKey(config.DedupFields, execCtx.GetInfoAtom())
	contextKey := fmt.Sprintf("%s|%s", timePeriod, fieldKey)

	tenantId := execCtx.GetTenantId()
	graphKey := execCtx.GetGraphKey()

	logx.WithContext(ctx).Debugf("[DedupCheck] 检查去重 key: %s", contextKey)

	// 4. 检查是否已存在
	existingCtx, err := datastore.GetGraphContext(ctx, tenantId, graphKey, contextKey)
	if err != nil && err != core.ErrItemNotFound {
		return false, fmt.Errorf("[DedupCheck] 查询图上下文失败: %w", err)
	}

	if existingCtx != nil {
		// 检查结果：已存在，使用 Info 级别
		logx.WithContext(ctx).Infof("[DedupCheck] 重复，跳过: %s", contextKey)
		if *config.TerminateIfExists {
			return false, nil
		}
		return true, nil
	}

	// 5. 不存在，创建新记录
	newContext := &core.BaseGraphContext{
		TenantId:   tenantId,
		GraphKey:   graphKey,
		ContextKey: contextKey,
		Payload: map[string]any{
			"checked":   true,
			"checkedAt": now.Format(time.RFC3339),
		},
	}

	if err := datastore.SaveGraphContext(ctx, newContext); err != nil {
		return false, fmt.Errorf("[DedupCheck] 保存图上下文失败: %w", err)
	}

	// 检查结果：首次处理，使用 Info 级别
	logx.WithContext(ctx).Infof("[DedupCheck] 首次处理: %s", contextKey)

	return true, nil
}

// getTimePeriod 根据重置模式计算时间周期标识
func getTimePeriod(now time.Time, mode ResetMode, resetTime string) string {
	switch mode {
	case ResetModeDaily:
		// 解析重置时间
		resetHour, resetMin := 0, 0
		fmt.Sscanf(resetTime, "%d:%d", &resetHour, &resetMin)

		// 计算当前属于哪个"天"（以重置时间为分界）
		resetToday := time.Date(now.Year(), now.Month(), now.Day(), resetHour, resetMin, 0, 0, now.Location())
		if now.Before(resetToday) {
			// 还没到今天的重置时间，属于"昨天"的周期
			return now.AddDate(0, 0, -1).Format("2006-01-02")
		}
		return now.Format("2006-01-02")

	case ResetModeHourly:
		return now.Format("2006-01-02T15")

	case ResetModeNone:
		return "static"

	default:
		return now.Format("2006-01-02")
	}
}

// buildFieldKey 根据去重字段从信息原子中提取值，构建字段部分的 key
func buildFieldKey(fields []string, atom core.InfoAtom) string {
	// 排序字段名，确保相同字段组合生成相同的 key
	sortedFields := make([]string, len(fields))
	copy(sortedFields, fields)
	sort.Strings(sortedFields)

	payload := atom.GetPayload()
	labels := atom.GetLabels()

	var parts []string
	for _, field := range sortedFields {
		var value string

		// 优先从 payload 获取
		if payload != nil {
			if val, ok := payload[field]; ok {
				value = fmt.Sprintf("%v", val)
			}
		}

		// 如果 payload 中没有，尝试从 labels 获取
		if value == "" && labels != nil {
			if val, ok := labels[field]; ok {
				value = val
			}
		}

		parts = append(parts, fmt.Sprintf("%s:%s", field, value))
	}

	return strings.Join(parts, "|")
}

// GetDedupCheckSpec 返回 DedupCheckBlock 的规格
func GetDedupCheckSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"DedupCheck",
		DedupCheckVersion,
		"去重检查，根据指定字段和时间周期判断是否已处理过",
		[]string{"input", "output"},
		map[string]any{
			"dedupFields": map[string]any{
				"type":        "array",
				"title":       "去重字段",
				"description": "选择用于判断重复的字段，相同字段值组合在同一时间周期内只处理一次",
				"check":       "must",
				"items": map[string]any{
					"type": "string",
				},
				"source": "infoAtomFields", // 前端从信息原子字段列表中选择
			},
			"resetMode": map[string]any{
				"type":        "string",
				"title":       "重置周期",
				"description": "去重记录的重置周期",
				"enum":        []string{"daily", "hourly", "none"},
				"enumLabels":  []string{"每天重置", "每小时重置", "不自动重置"},
				"default":     "daily",
			},
			"resetTime": map[string]any{
				"type":        "string",
				"title":       "每日重置时间",
				"description": "当重置周期为「每天」时，指定重置时间点",
				"placeholder": "00:00",
				"default":     "00:00",
				"showWhen": map[string]any{
					"field": "resetMode",
					"value": "daily",
				},
			},
			"timezone": map[string]any{
				"type":        "string",
				"title":       "时区",
				"description": "用于计算重置时间的时区",
				"default":     "Asia/Shanghai",
				"placeholder": "Asia/Shanghai",
			},
			"terminateIfExists": map[string]any{
				"type":        "boolean",
				"title":       "已存在时终止",
				"description": "如果已处理过，是否终止流程",
				"default":     true,
			},
		},
		[]core.ServiceType{},
	)
}
