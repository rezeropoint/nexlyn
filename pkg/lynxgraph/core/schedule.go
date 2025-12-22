package core

// ScheduleConfig 定时配置
type ScheduleConfig struct {
	CronExpr       string         `json:"cronExpr"`                 // cron 表达式，如 "31 9 * * 1-5"
	Timezone       string         `json:"timezone,omitempty"`       // 时区，默认 Asia/Shanghai
	InitialPayload map[string]any `json:"initialPayload,omitempty"` // 初始载荷
}

// ScheduleTriggerFunc 定时触发回调函数类型
// 参数：graphKey 图标识, scheduleConfig 定时配置
type ScheduleTriggerFunc func(graphKey GraphKey, scheduleConfig *ScheduleConfig) error

// ScheduleRegisterFunc 定时任务注册函数类型
// 参数：graphKey 图标识, nodes 节点配置列表（从中提取定时积木）
type ScheduleRegisterFunc func(graphKey GraphKey, nodes []NodeConfig) error

// ScheduleUnregisterFunc 定时任务注销函数类型
type ScheduleUnregisterFunc func(graphKey GraphKey) error

// HasScheduleNodes 检查节点配置中是否包含定时积木
func HasScheduleNodes(nodes []NodeConfig) bool {
	for _, node := range nodes {
		if node.BlockType == BlockTypeSchedule && node.IsEntryPoint {
			return true
		}
	}
	return false
}

// ExtractScheduleConfigsFromNodes 从节点配置中提取定时积木的配置
// 返回 nodeID -> ScheduleConfig 的映射
func ExtractScheduleConfigsFromNodes(nodes []NodeConfig) map[string]*ScheduleConfig {
	result := make(map[string]*ScheduleConfig)
	for _, node := range nodes {
		if node.BlockType != BlockTypeSchedule || !node.IsEntryPoint {
			continue
		}

		config := &ScheduleConfig{}

		// 从 BlockConfig 中提取 ScheduleConfig 字段
		if cronExpr, ok := node.BlockConfig["cronExpr"].(string); ok {
			config.CronExpr = cronExpr
		}
		if timezone, ok := node.BlockConfig["timezone"].(string); ok {
			config.Timezone = timezone
		}
		if payload, ok := node.BlockConfig["initialPayload"].(map[string]any); ok {
			config.InitialPayload = payload
		}

		// 只有有效的 cron 表达式才添加
		if config.CronExpr != "" {
			result[node.ID] = config
		}
	}
	return result
}
