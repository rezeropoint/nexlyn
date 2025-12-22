package scheduler

// Config 调度器配置
type Config struct {
	// DefaultTimezone 默认时区，默认 Asia/Shanghai
	DefaultTimezone string
	// EnableSeconds 是否启用秒级精度（默认 false，使用分钟级）
	EnableSeconds bool
}
