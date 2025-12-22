package getdedupset

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/zeromicro/go-zero/core/logx"
)

const GetDedupSetVersion = "v1"

// ResetMode 重置模式（与 DedupCheck 一致）
type ResetMode string

const (
	ResetModeDaily  ResetMode = "daily"  // 每天指定时间重置
	ResetModeHourly ResetMode = "hourly" // 每小时重置
	ResetModeNone   ResetMode = "none"   // 不自动重置
)

// Config GetDedupSet 积木配置
type Config struct {
	SourceNodeId *string   `json:"sourceNodeId"`             // 关联的 DedupCheck 节点 ID（可选，前端用于自动复制配置）
	DedupFields  []string  `json:"dedupFields"`              // 去重字段列表（与 DedupCheck 配置一致）
	ResetMode    ResetMode `json:"resetMode"`                // 重置模式：daily/hourly/none，默认 daily
	ResetTime    string    `json:"resetTime"`                // 每日重置时间，格式 HH:MM，默认 "00:00"
	Timezone     string    `json:"timezone"`                 // 时区，默认 Asia/Shanghai
	SaveResultTo string    `json:"saveResultTo" check:"must"` // 结果保存到 GraphContext 的 key
}

// GetDedupSetBlock 获取去重集合积木
type GetDedupSetBlock struct {
	core.BaseLogicBlock
}

// NewGetDedupSetBlock 创建一个新的 GetDedupSetBlock 实例
func NewGetDedupSetBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &GetDedupSetBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeGetDedupSet,
			RawConfig: config,
		},
	}, nil
}

func (b *GetDedupSetBlock) GetID() string                { return b.ID }
func (b *GetDedupSetBlock) GetType() core.LogicBlockType { return b.Type }
func (b *GetDedupSetBlock) GetConfigure() map[string]any { return b.RawConfig }

func (b *GetDedupSetBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 验证去重字段不能为空（除非指定了 sourceNodeId，此时前端会自动复制配置）
	if (cfg.SourceNodeId == nil || *cfg.SourceNodeId == "") && len(cfg.DedupFields) == 0 {
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

// Execute 执行获取去重集合逻辑
func (b *GetDedupSetBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	config, ok := b.TypedConfig.(*Config)
	if !ok {
		return false, core.ErrInvalidConfig
	}

	// 1. 获取当前时间（用于时间周期计算）
	loc, err := time.LoadLocation(config.Timezone)
	if err != nil {
		logx.WithContext(ctx).Debugf("[GetDedupSet] 无效的时区 %s，使用 UTC", config.Timezone)
		loc = time.UTC
	}
	now := time.Now().In(loc)

	// 2. 计算时间周期前缀
	timePeriod := getTimePeriod(now, config.ResetMode, config.ResetTime)

	tenantId := execCtx.GetTenantId()
	graphKey := execCtx.GetGraphKey()

	// 3. 使用 SCAN 查询当前周期的所有去重 key
	prefix := timePeriod + "|"
	contextKeys, err := datastore.ScanGraphContextKeys(ctx, tenantId, graphKey, prefix)
	if err != nil {
		return false, fmt.Errorf("[GetDedupSet] 扫描图上下文失败: %w", err)
	}

	logx.WithContext(ctx).Infof("[GetDedupSet] 查询到 %d 条去重记录，周期: %s", len(contextKeys), timePeriod)

	// 4. 解析 contextKey 提取去重字段值
	// contextKey 格式: {timePeriod}|{field1}:{value1}|{field2}:{value2}|...
	dedupRecords := make([]map[string]string, 0, len(contextKeys))
	for _, key := range contextKeys {
		// 移除时间周期前缀
		fieldPart := strings.TrimPrefix(key, prefix)
		if fieldPart == "" {
			continue
		}

		// 解析字段值
		record := make(map[string]string)
		parts := strings.Split(fieldPart, "|")
		for _, part := range parts {
			kv := strings.SplitN(part, ":", 2)
			if len(kv) == 2 {
				record[kv[0]] = kv[1]
			}
		}

		if len(record) > 0 {
			dedupRecords = append(dedupRecords, record)
		}
	}

	// 5. 保存结果到 GraphContext
	resultPayload := map[string]any{
		"keys":       contextKeys,
		"records":    dedupRecords,
		"count":      len(contextKeys),
		"timePeriod": timePeriod,
		"queriedAt":  now.Format(time.RFC3339),
	}

	newContext := &core.BaseGraphContext{
		TenantId:   tenantId,
		GraphKey:   graphKey,
		ContextKey: config.SaveResultTo,
		Payload:    resultPayload,
	}

	if err := datastore.SaveGraphContext(ctx, newContext); err != nil {
		return false, fmt.Errorf("[GetDedupSet] 保存结果失败: %w", err)
	}

	return true, nil
}

// getTimePeriod 根据重置模式计算时间周期标识（与 DedupCheck 保持一致）
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

// buildFieldKey 根据去重字段构建字段部分的 key（供调试使用）
func buildFieldKey(fields []string, values map[string]string) string {
	// 排序字段名，确保相同字段组合生成相同的 key
	sortedFields := make([]string, len(fields))
	copy(sortedFields, fields)
	sort.Strings(sortedFields)

	var parts []string
	for _, field := range sortedFields {
		value := values[field]
		parts = append(parts, fmt.Sprintf("%s:%s", field, value))
	}

	return strings.Join(parts, "|")
}

// GetGetDedupSetSpec 返回 GetDedupSetBlock 的规格
func GetGetDedupSetSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"GetDedupSet",
		GetDedupSetVersion,
		"获取当前周期所有已去重记录，配合 ForEach 实现员工遍历等场景",
		[]string{"input", "output"},
		map[string]any{
			"sourceNodeId": map[string]any{
				"type":        "string",
				"title":       "关联去重节点",
				"description": "选择图中的 DedupCheck 节点，自动使用其配置",
				"source":      "dedupCheckNodeSelector",
			},
			"dedupFields": map[string]any{
				"type":        "array",
				"title":       "去重字段",
				"description": "与 DedupCheck 积木配置一致，用于匹配去重记录",
				"items": map[string]any{
					"type": "string",
				},
				"hideWhen": map[string]any{
					"field":    "sourceNodeId",
					"hasValue": true,
				},
			},
			"resetMode": map[string]any{
				"type":        "string",
				"title":       "重置周期",
				"description": "与 DedupCheck 积木配置一致",
				"enum":        []string{"daily", "hourly", "none"},
				"enumLabels":  []string{"每天重置", "每小时重置", "不自动重置"},
				"default":     "daily",
				"hideWhen": map[string]any{
					"field":    "sourceNodeId",
					"hasValue": true,
				},
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
				"hideWhen": map[string]any{
					"field":    "sourceNodeId",
					"hasValue": true,
				},
			},
			"timezone": map[string]any{
				"type":        "string",
				"title":       "时区",
				"description": "用于计算重置时间的时区",
				"default":     "Asia/Shanghai",
				"placeholder": "Asia/Shanghai",
				"hideWhen": map[string]any{
					"field":    "sourceNodeId",
					"hasValue": true,
				},
			},
			"saveResultTo": map[string]any{
				"type":        "string",
				"title":       "结果保存位置",
				"description": "查询结果保存到 GraphContext 的 key",
				"check":       "must",
				"placeholder": "dedup_set_result",
			},
		},
		[]core.ServiceType{},
	)
}
