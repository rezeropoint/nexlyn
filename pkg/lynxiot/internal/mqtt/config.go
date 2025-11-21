package mqtt

import (
	"time"
)

// Config MQTT管理器配置
type Config struct {
	// MQTT连接配置
	Broker               string        // MQTT Broker地址（如：tcp://localhost:1883）
	ClientID             string        // 客户端ID
	Username             string        // 用户名
	Password             string        // 密码
	ConnectTimeout       time.Duration // 连接超时时间
	PingTimeout          time.Duration // Ping超时时间
	KeepAlive            time.Duration // 保活时间
	MaxReconnectInterval time.Duration // 最大重连间隔

	// 服务标识
	ServiceName string // 服务名称（用于日志）
	PodName     string // Pod名称（用于日志）

	// Redis配置
	RedisTTL int // Redis键过期时间（秒，0表示不过期）

	// 分布式锁配置
	DedupWindowMs  int64 // 消息去重时间窗口（毫秒），默认100ms。同一主题在此窗口内的消息只处理一次
	LockTTLSeconds int   // 分布式锁TTL（秒），默认30秒。防止死锁

}
