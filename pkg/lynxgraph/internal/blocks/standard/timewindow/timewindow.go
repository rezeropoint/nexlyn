package timewindow

import (
	"context"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/zeromicro/go-zero/core/logx"
)

const TimeWindowCheckVersion = "v1"

// TimeWindow 时间窗口配置
type TimeWindow struct {
	Name      string `json:"name"`      // 窗口名称
	StartTime string `json:"startTime"` // 开始时间 HH:MM 格式
	EndTime   string `json:"endTime"`   // 结束时间 HH:MM 格式
}

// Config 时间窗口检查积木的配置
type Config struct {
	TimeSource string       `json:"timeSource"` // 时间来源: atom（默认）或 now
	Windows    []TimeWindow `json:"windows"`    // 时间窗口列表
	Timezone   string       `json:"timezone"`   // 时区，默认 Asia/Shanghai
	ResultKey  string       `json:"resultKey"`  // 保存匹配窗口名到 GraphContext 的 key
}

// TimeWindowCheckBlock 实现时间窗口检查逻辑块
type TimeWindowCheckBlock struct {
	core.BaseLogicBlock
}

// NewTimeWindowCheckBlock 创建一个新的 TimeWindowCheckBlock 实例
func NewTimeWindowCheckBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &TimeWindowCheckBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeTimeWindowCheck,
			RawConfig: config,
		},
	}, nil
}

func (b *TimeWindowCheckBlock) GetID() string                 { return b.ID }
func (b *TimeWindowCheckBlock) GetType() core.LogicBlockType  { return b.Type }
func (b *TimeWindowCheckBlock) GetConfigure() map[string]any  { return b.RawConfig }

func (b *TimeWindowCheckBlock) SetConfigure(config map[string]any) error {
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

	// 验证时间窗口格式
	for i, window := range cfg.Windows {
		if window.Name == "" {
			return fmt.Errorf("时间窗口 %d 缺少名称", i)
		}
		if _, err := parseTimeOfDay(window.StartTime); err != nil {
			return fmt.Errorf("时间窗口 %s 的开始时间格式无效: %s", window.Name, window.StartTime)
		}
		if _, err := parseTimeOfDay(window.EndTime); err != nil {
			return fmt.Errorf("时间窗口 %s 的结束时间格式无效: %s", window.Name, window.EndTime)
		}
	}

	b.TypedConfig = &cfg
	return nil
}

// Execute 执行时间窗口检查
func (b *TimeWindowCheckBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	config, ok := b.TypedConfig.(*Config)
	if !ok {
		return false, core.ErrInvalidConfig
	}

	// 获取时间
	var checkTime time.Time
	if config.TimeSource == "atom" {
		// 从 InfoAtom 获取时间戳（毫秒）
		timestamp := execCtx.GetInfoAtom().GetTimestamp()
		checkTime = time.UnixMilli(timestamp)
	} else {
		checkTime = time.Now()
	}

	// 转换到指定时区
	loc, _ := time.LoadLocation(config.Timezone) // 已在 SetConfigure 验证
	checkTime = checkTime.In(loc)

	// 提取时分
	hour := checkTime.Hour()
	minute := checkTime.Minute()
	currentMinutes := hour*60 + minute

	logx.WithContext(ctx).Infof("[TimeWindowCheck] 检查时间: %02d:%02d (时区: %s)", hour, minute, config.Timezone)

	// 找出匹配的窗口
	var matchedWindow string
	for _, window := range config.Windows {
		startMinutes, _ := parseTimeOfDay(window.StartTime)
		endMinutes, _ := parseTimeOfDay(window.EndTime)

		var matched bool
		if startMinutes <= endMinutes {
			// 正常窗口 (如 09:00 - 17:00)
			matched = currentMinutes >= startMinutes && currentMinutes <= endMinutes
		} else {
			// 跨午夜窗口 (如 22:00 - 06:00)
			matched = currentMinutes >= startMinutes || currentMinutes <= endMinutes
		}

		if matched {
			matchedWindow = window.Name
			logx.WithContext(ctx).Infof("[TimeWindowCheck] 匹配窗口: %s (%s - %s)", window.Name, window.StartTime, window.EndTime)
			break
		}
	}

	// 保存结果到 GraphContext（窗口名称为键，布尔值表示是否匹配）
	if config.ResultKey != "" {
		payload := make(map[string]any)
		for _, window := range config.Windows {
			payload[window.Name] = window.Name == matchedWindow
		}
		graphContext := &core.BaseGraphContext{
			TenantId:   execCtx.GetTenantId(),
			GraphKey:   execCtx.GetGraphKey(),
			ContextKey: config.ResultKey,
			Payload:    payload,
		}
		if err := datastore.SaveGraphContext(ctx, graphContext); err != nil {
			logx.WithContext(ctx).Errorf("[TimeWindowCheck] 保存结果到 GraphContext 失败: %v", err)
		}
	}

	if matchedWindow == "" {
		logx.WithContext(ctx).Infof("[TimeWindowCheck] 未匹配任何时间窗口")
	}
	return matchedWindow != "", nil
}

// parseTimeOfDay 解析 HH:MM 格式的时间，返回分钟数
func parseTimeOfDay(timeStr string) (int, error) {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return 0, err
	}
	return t.Hour()*60 + t.Minute(), nil
}

// GetTimeWindowCheckSpec 返回 TimeWindowCheckBlock 的规格
func GetTimeWindowCheckSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"TimeWindowCheck",
		TimeWindowCheckVersion,
		"检查事件时间是否在指定的时间窗口内，支持多个窗口配置，匹配首个满足条件的窗口",
		[]string{"condition", "time", "schedule"},
		map[string]any{
			"timeSource": map[string]any{
				"type":        "string",
				"title":       "时间来源",
				"description": "检查的时间来源：atom=信息原子时间戳（推荐），now=服务器当前时间",
				"check":       "",
				"enum":        []string{"atom", "now"},
				"enumLabels":  []string{"信息原子时间戳", "当前时间"},
			},
			"windows": map[string]any{
				"type":        "array",
				"title":       "时间窗口",
				"description": "时间窗口列表，按顺序检查，匹配首个满足条件的窗口",
				"check":       "must",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type":        "string",
							"title":       "窗口名称",
							"description": "用于标识匹配的窗口",
						},
						"startTime": map[string]any{
							"type":        "string",
							"title":       "开始时间",
							"description": "HH:MM 格式，如 09:00",
						},
						"endTime": map[string]any{
							"type":        "string",
							"title":       "结束时间",
							"description": "HH:MM 格式，如 17:30",
						},
					},
				},
			},
			"timezone": map[string]any{
				"type":        "string",
				"title":       "时区",
				"description": "时区名称，默认 Asia/Shanghai",
				"placeholder": "Asia/Shanghai",
				"check":       "",
			},
			"resultKey": map[string]any{
				"type":        "string",
				"title":       "结果键名",
				"description": "保存到图上下文的键名，边条件使用 {键名}.{窗口名} == true 判断",
				"placeholder": "time_window",
				"check":       "",
			},
		},
		[]core.ServiceType{}, // 不依赖任何外部服务
	).WithOutputSchema(map[string]any{
		"contextKey": "config.resultKey",          // 上下文键名来自配置
		"type":       "dynamic-keys",              // 键是动态的
		"keySource":  "config.windows[].name",     // 键来自配置中的窗口名称
		"valueType":  "boolean",                   // 值类型为布尔
		"description": "以窗口名称为键，布尔值表示该窗口是否匹配",
	})
}
