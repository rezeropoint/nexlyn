package cleardedupcontext

import (
	"context"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/zeromicro/go-zero/core/logx"
)

const ClearDedupContextVersion = "v1"

// ResetMode 重置模式（与 DedupCheck/GetDedupSet 一致）
type ResetMode string

const (
	ResetModeDaily  ResetMode = "daily"  // 每天指定时间重置
	ResetModeHourly ResetMode = "hourly" // 每小时重置
	ResetModeNone   ResetMode = "none"   // 不自动重置
)

// Config ClearDedupContext 积木配置
type Config struct {
	SourceNodeId *string   `json:"sourceNodeId"`  // 关联的 DedupCheck 节点 ID（可选，前端用于自动复制配置）
	DedupFields  []string  `json:"dedupFields"`   // 去重字段列表（与 DedupCheck 配置一致，用于文档说明）
	ResetMode    ResetMode `json:"resetMode"`     // 重置模式：daily/hourly/none，默认 daily
	ResetTime    string    `json:"resetTime"`     // 每日重置时间，格式 HH:MM，默认 "00:00"
	Timezone     string    `json:"timezone"`      // 时区，默认 Asia/Shanghai
}

// ClearDedupContextBlock 清空去重记录积木
type ClearDedupContextBlock struct {
	core.BaseLogicBlock
}

// NewClearDedupContextBlock 创建一个新的 ClearDedupContextBlock 实例
func NewClearDedupContextBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &ClearDedupContextBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeClearDedupContext,
			RawConfig: config,
		},
	}, nil
}

func (b *ClearDedupContextBlock) GetID() string                { return b.ID }
func (b *ClearDedupContextBlock) GetType() core.LogicBlockType { return b.Type }
func (b *ClearDedupContextBlock) GetConfigure() map[string]any { return b.RawConfig }

func (b *ClearDedupContextBlock) SetConfigure(config map[string]any) error {
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

// Execute 执行清空去重记录逻辑
func (b *ClearDedupContextBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	config, ok := b.TypedConfig.(*Config)
	if !ok {
		return false, core.ErrInvalidConfig
	}

	// 1. 获取当前时间（用于时间周期计算）
	loc, err := time.LoadLocation(config.Timezone)
	if err != nil {
		logx.WithContext(ctx).Debugf("[ClearDedupContext] 无效的时区 %s，使用 UTC", config.Timezone)
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
		return false, fmt.Errorf("[ClearDedupContext] 扫描图上下文失败: %w", err)
	}

	if len(contextKeys) == 0 {
		logx.WithContext(ctx).Infof("[ClearDedupContext] 无需清理，周期 %s 没有去重记录", timePeriod)
		return true, nil
	}

	logx.WithContext(ctx).Infof("[ClearDedupContext] 开始清理 %d 条去重记录，周期: %s", len(contextKeys), timePeriod)

	// 4. 遍历删除每个 key
	deletedCount := 0
	failedCount := 0
	for _, contextKey := range contextKeys {
		if err := datastore.DeleteGraphContext(ctx, tenantId, graphKey, contextKey); err != nil {
			logx.WithContext(ctx).Errorf("[ClearDedupContext] 删除失败 %s: %v", contextKey, err)
			failedCount++
			continue
		}
		deletedCount++
	}

	logx.WithContext(ctx).Infof("[ClearDedupContext] 清理完成，删除 %d 条，失败 %d 条", deletedCount, failedCount)

	// 即使有部分失败也继续流程
	return true, nil
}

// getTimePeriod 根据重置模式计算时间周期标识（与 DedupCheck/GetDedupSet 保持一致）
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

// GetClearDedupContextSpec 返回 ClearDedupContextBlock 的规格
func GetClearDedupContextSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"ClearDedupContext",
		ClearDedupContextVersion,
		"清空指定周期的所有去重记录，通常在定时任务结束时使用",
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
				"description": "与 DedupCheck 积木配置一致（用于文档说明，实际清理基于时间周期前缀）",
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
		},
		[]core.ServiceType{},
	)
}
