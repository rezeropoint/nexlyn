package storage

// Config Storage Manager配置
type Config struct {
	// ClickHouse连接配置
	Database string `json:",default=nexlyn"` // 数据库名称

	// TTL配置覆盖（可选）
	// 格式：字段名 -> TTL天数
	// 优先级：TTLOverrides > 字段定义TTL > 系统默认值180天
	// 注意：仅在表创建时生效，已存在的表不受影响
	TTLOverrides map[string]int `json:",optional"` // TTL覆盖配置

	// 服务信息（用于日志）
	ServiceName string `json:",optional"` // 服务名称
	PodName     string `json:",optional"` // Pod名称
}
