package latetimecalculate

import (
	"context"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/zeromicro/go-zero/core/logx"
)

const LateTimeCalculateVersion = "v1"

// Config 迟到时间计算积木的配置
type Config struct {
	TimeSource    string `json:"timeSource"`                  // 时间来源: atom（默认）或 now
	WorkStartTime string `json:"workStartTime" check:"must"`  // 上班时间，HH:MM 格式
	Timezone      string `json:"timezone"`                    // 时区，默认 Asia/Shanghai
	SaveResultTo  string `json:"saveResultTo" check:"must"`   // 保存到 GraphContext 的键名
}

// LateTimeCalculateBlock 实现迟到时间计算逻辑块
type LateTimeCalculateBlock struct {
	core.BaseLogicBlock
}

// NewLateTimeCalculateBlock 创建一个新的 LateTimeCalculateBlock 实例
func NewLateTimeCalculateBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &LateTimeCalculateBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeLateTimeCalculate,
			RawConfig: config,
		},
	}, nil
}

func (b *LateTimeCalculateBlock) GetID() string                { return b.ID }
func (b *LateTimeCalculateBlock) GetType() core.LogicBlockType { return b.Type }
func (b *LateTimeCalculateBlock) GetConfigure() map[string]any { return b.RawConfig }

func (b *LateTimeCalculateBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 设置默认值
	if cfg.TimeSource == "" {
		cfg.TimeSource = "atom"
	}
	if cfg.Timezone == "" {
		cfg.Timezone = "Asia/Shanghai"
	}

	// 验证时间来源
	if cfg.TimeSource != "atom" && cfg.TimeSource != "now" {
		return fmt.Errorf("不支持的时间来源: %s (仅支持 atom/now)", cfg.TimeSource)
	}

	// 验证时区
	_, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return fmt.Errorf("无效的时区: %s", cfg.Timezone)
	}

	// 验证上班时间格式
	if _, err := parseTimeOfDay(cfg.WorkStartTime); err != nil {
		return fmt.Errorf("上班时间格式无效: %s (应为 HH:MM 格式)", cfg.WorkStartTime)
	}

	b.TypedConfig = &cfg
	return nil
}

// Execute 执行迟到时间计算
func (b *LateTimeCalculateBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	config, ok := b.TypedConfig.(*Config)
	if !ok {
		return false, core.ErrInvalidConfig
	}

	// 1. 获取打卡时间
	var checkTime time.Time
	if config.TimeSource == "atom" {
		timestamp := execCtx.GetInfoAtom().GetTimestamp()
		if timestamp == 0 {
			checkTime = time.Now()
			logx.WithContext(ctx).Infof("[LateTimeCalculate] 信息原子时间戳为0，使用当前时间")
		} else {
			checkTime = time.UnixMilli(timestamp)
		}
	} else {
		checkTime = time.Now()
	}

	// 2. 转换到指定时区
	loc, _ := time.LoadLocation(config.Timezone) // 已在 SetConfigure 验证
	checkTime = checkTime.In(loc)

	// 3. 解析上班时间
	workStartMinutes, _ := parseTimeOfDay(config.WorkStartTime) // 已在 SetConfigure 验证

	// 4. 计算迟到分钟数
	hour := checkTime.Hour()
	minute := checkTime.Minute()
	currentMinutes := hour*60 + minute

	lateMinutes := 0
	isLate := false
	if currentMinutes > workStartMinutes {
		lateMinutes = currentMinutes - workStartMinutes
		isLate = true
	}

	checkTimeStr := checkTime.Format("15:04")

	logx.WithContext(ctx).Infof("[LateTimeCalculate] 打卡时间: %s, 上班时间: %s, 迟到: %v, 迟到分钟数: %d",
		checkTimeStr, config.WorkStartTime, isLate, lateMinutes)

	// 5. 保存结果到 GraphContext
	payload := map[string]any{
		"lateMinutes":   lateMinutes,          // int: 迟到分钟数（不迟到为0）
		"isLate":        isLate,               // bool: 是否迟到
		"checkTime":     checkTimeStr,         // string: 打卡时间 HH:MM
		"workStartTime": config.WorkStartTime, // string: 上班时间 HH:MM
	}

	graphContext := &core.BaseGraphContext{
		TenantId:   execCtx.GetTenantId(),
		GraphKey:   execCtx.GetGraphKey(),
		ContextKey: config.SaveResultTo,
		Payload:    payload,
	}

	if err := datastore.SaveGraphContext(ctx, graphContext); err != nil {
		logx.WithContext(ctx).Errorf("[LateTimeCalculate] 保存结果到 GraphContext 失败: %v", err)
		return false, fmt.Errorf("保存结果失败: %w", err)
	}

	return true, nil
}

// parseTimeOfDay 解析 HH:MM 格式的时间，返回分钟数
func parseTimeOfDay(timeStr string) (int, error) {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return 0, err
	}
	return t.Hour()*60 + t.Minute(), nil
}

// GetLateTimeCalculateSpec 返回 LateTimeCalculateBlock 的规格
func GetLateTimeCalculateSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"LateTimeCalculate",
		LateTimeCalculateVersion,
		"计算打卡时间与上班时间的差值，输出迟到分钟数",
		[]string{"time", "attendance", "calculate"},
		map[string]any{
			"timeSource": map[string]any{
				"type":        "string",
				"title":       "时间来源",
				"description": "打卡时间来源：atom=信息原子时间戳（推荐），now=服务器当前时间",
				"check":       "",
				"enum":        []string{"atom", "now"},
				"enumLabels":  []string{"信息原子时间戳", "当前时间"},
			},
			"workStartTime": map[string]any{
				"type":        "string",
				"title":       "上班时间",
				"description": "规定的上班时间，HH:MM 格式",
				"placeholder": "09:30",
				"check":       "must",
			},
			"timezone": map[string]any{
				"type":        "string",
				"title":       "时区",
				"description": "时区名称，默认 Asia/Shanghai",
				"placeholder": "Asia/Shanghai",
			},
			"saveResultTo": map[string]any{
				"type":        "string",
				"title":       "结果保存键名",
				"description": "保存到图上下文的键名",
				"placeholder": "late_time",
				"check":       "must",
			},
		},
		[]core.ServiceType{}, // 不依赖任何外部服务
	).WithOutputSchema(map[string]any{
		"contextKey": "config.saveResultTo",
		"type":       "object",
		"properties": map[string]any{
			"lateMinutes": map[string]any{
				"type":        "integer",
				"description": "迟到分钟数（整数，不迟到为0）",
			},
			"isLate": map[string]any{
				"type":        "boolean",
				"description": "是否迟到",
			},
			"checkTime": map[string]any{
				"type":        "string",
				"description": "打卡时间（HH:MM格式）",
			},
			"workStartTime": map[string]any{
				"type":        "string",
				"description": "上班时间（HH:MM格式）",
			},
		},
	})
}
