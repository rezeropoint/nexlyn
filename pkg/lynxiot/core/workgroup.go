package core

// TaskInfo 任务信息结构体
type TaskInfo struct {
	TaskId     string
	ConfigType string
	TenantID   string // 租户ID（用于分发到外部系统）
	DeviceID   string // 设备ID（用于分发到外部系统）
	Timestamp  int64  // 时间戳（毫秒），用于分发到 LynxGraph
}

// TaskResult 任务结果结构体
type TaskResult struct {
	Error error
	Index int // 用于标识是哪个任务的结果
}

// SubmitAsyncFunc 异步提交任务函数
type SubmitAsyncFunc func(tasks []func() error) []error
