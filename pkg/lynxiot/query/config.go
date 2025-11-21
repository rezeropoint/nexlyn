package query

import "time"

// Config 查询引擎配置
type Config struct {
	// ClickHouse配置
	ClickHouseDatabase string // 数据库名称（默认：nexlyn）

	// 查询限制配置
	DefaultLimit    int           // 默认查询限制（默认：1000）
	MaxLimit        int           // 最大查询限制（默认：10000）
	QueryTimeout    time.Duration // 查询超时时间（默认：30秒）
	MaxFieldsPerReq int           // 单次请求最大字段数（默认：10）

	// 服务信息（用于日志）
	ServiceName string // 服务名称
	PodName     string // Pod名称
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		ClickHouseDatabase: "nexlyn",
		DefaultLimit:       1000,
		MaxLimit:           10000,
		QueryTimeout:       30 * time.Second,
		MaxFieldsPerReq:    10,
		ServiceName:        "iot-query",
		PodName:            "",
	}
}
