package query

import "time"

// Config Query Manager配置
type Config struct {
	// ClickHouse配置
	Database string `json:",default=nexlyn"` // 数据库名称

	// 查询限制配置
	DefaultLimit    int           `json:",default=1000"`  // 默认查询限制（防止大量数据返回）
	MaxLimit        int           `json:",default=10000"` // 最大查询限制
	QueryTimeout    time.Duration `json:",default=30"`    // 查询超时时间（秒）
	MaxFieldsPerReq int           `json:",default=10"`    // 单次请求最大字段数

	// 服务信息（用于日志）
	ServiceName string `json:",optional"` // 服务名称
	PodName     string `json:",optional"` // Pod名称
}
