package controller

import (
	"time"
)

// Config Controller管理器配置（只包含配置数据，不包含运行时实例）
type Config struct {
	// MQTT连接配置（独立客户端）
	Broker               string        // MQTT Broker地址（如：tcp://localhost:1883）
	ClientID             string        // 客户端ID
	Username             string        // 用户名
	Password             string        // 密码
	ConnectTimeout       time.Duration // 连接超时时间
	PingTimeout          time.Duration // Ping超时时间
	KeepAlive            time.Duration // 保活时间
	MaxReconnectInterval time.Duration // 最大重连间隔

	// 缓存配置
	CapabilitiesCacheTTL int // 算法能力缓存时长（秒），默认7200秒（2小时），设置为0表示禁用缓存

	// 服务标识
	ServiceName string // 服务名称（用于日志）
	PodName     string // Pod名称（用于日志）
}
