package engine

import (
	"time"
)

// Config 引擎配置
type Config struct {
	// Etcd配置
	EtcdHosts []string // Etcd主机列表
	EtcdUser  string   // Etcd用户名
	EtcdPass  string   // Etcd密码

	// MQTT配置
	MQTTBroker               string        // MQTT Broker地址（如：tcp://localhost:1883）
	MQTTClientID             string        // MQTT客户端ID
	MQTTUsername             string        // MQTT用户名
	MQTTPassword             string        // MQTT密码
	MQTTConnectTimeout       time.Duration // MQTT连接超时时间（可选，默认10秒）
	MQTTPingTimeout          time.Duration // MQTT Ping超时时间（可选，默认10秒）
	MQTTKeepAlive            time.Duration // MQTT保活时间（可选，默认60秒）
	MQTTMaxReconnectInterval time.Duration // MQTT最大重连间隔（可选，默认10分钟）
	MQTTRedisTTL             int           // MQTT在线状态Redis键过期时间（秒，可选，默认300秒即5分钟）

	// 控制配置
	CapabilitiesCacheTTL int // 算法能力缓存时长（秒，可选，默认7200秒即2小时，设置为0表示禁用缓存）

	// Workgroup 配置
	WorkerCount int // 工作协程数量（可选，默认10，用于异步数据分发）

	// Storage 配置
	ClickHouseDatabase string         // ClickHouse数据库名称（用于Storage Manager）
	TTLOverrides       map[string]int // TTL配置覆盖（字段名 -> TTL天数，可选）

	// LynxGraph 配置
	LynxGraphGRPCURL string // LynxGraph gRPC服务地址（如：lynxengine:9999）

	// Pod信息（用于Etcd配置管理和日志）
	PodName     string // Pod名称
	ServiceName string // 服务名称
}
