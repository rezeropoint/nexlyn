package schedule

import (
	"context"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
)

const ScheduleVersion = "v1"

// Config 定时触发积木的配置
type Config struct {
	CronExpr       string         `json:"cronExpr" check:"must"`    // cron 表达式，如 "31 9 * * 1-5"
	Timezone       string         `json:"timezone"`                 // 时区，默认 Asia/Shanghai
	InitialPayload map[string]any `json:"initialPayload,omitempty"` // 初始载荷
}

// ScheduleBlock 定时触发积木
// 注意：此积木不执行实际逻辑，由 scheduler 特殊处理
// 它的作用是在画布上可视化定时触发配置，并作为入口节点
type ScheduleBlock struct {
	core.BaseLogicBlock
}

// NewScheduleBlock 创建定时积木实例
func NewScheduleBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &ScheduleBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeSchedule,
			RawConfig: config,
		},
	}, nil
}

func (b *ScheduleBlock) GetID() string                 { return b.ID }
func (b *ScheduleBlock) GetType() core.LogicBlockType  { return b.Type }
func (b *ScheduleBlock) GetConfigure() map[string]any  { return b.RawConfig }

func (b *ScheduleBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 设置默认时区
	if cfg.Timezone == "" {
		cfg.Timezone = "Asia/Shanghai"
	}

	// 验证时区
	if _, err := time.LoadLocation(cfg.Timezone); err != nil {
		return fmt.Errorf("无效的时区: %s", cfg.Timezone)
	}

	b.TypedConfig = &cfg
	return nil
}

// Execute 定时积木的执行方法
// 定时积木不执行实际逻辑，直接返回 true 继续后续节点
// 触发时机由 scheduler 控制
func (b *ScheduleBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	// 定时积木作为入口点，不执行实际逻辑
	// 其作用是标识该图需要定时触发，并携带定时配置
	return true, nil
}

// GetScheduleSpec 返回定时积木的规格
func GetScheduleSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"Schedule",
		ScheduleVersion,
		"定时触发器，按 cron 表达式周期性触发逻辑图执行",
		[]string{"trigger", "schedule", "entry"},
		map[string]any{
			"cronExpr": map[string]any{
				"type":        "string",
				"title":       "Cron 表达式",
				"description": "标准 cron 表达式（分 时 日 月 周），如 '31 9 * * 1-5' 表示工作日9:31触发",
				"placeholder": "31 9 * * 1-5",
				"check":       "must",
			},
			"timezone": map[string]any{
				"type":        "string",
				"title":       "时区",
				"description": "执行时区，默认 Asia/Shanghai",
				"placeholder": "Asia/Shanghai",
			},
			"initialPayload": map[string]any{
				"type":        "object",
				"title":       "初始载荷",
				"description": "触发时携带的初始数据，会注入到 InfoAtom 的 payload 中",
			},
		},
		[]core.ServiceType{}, // 不依赖外部服务
	)
}
