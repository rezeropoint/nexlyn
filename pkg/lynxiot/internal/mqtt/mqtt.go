package mqtt

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/rezeropoint/etcdtrigger/v2/engine"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// Manager MQTT管理器接口
type Manager interface {
	// Start 启动MQTT监控（连接Broker并订阅主题）
	Start(ctx context.Context) error

	// Stop 停止MQTT监控
	Stop()

	// IsConnected 检查MQTT连接状态
	IsConnected() bool

	// 设备在线状态缓存接口（提供给Device Manager使用）

	// GetDeviceOnlineStatus 获取设备在线状态
	// 返回: 在线状态（true/false）
	GetDeviceOnlineStatus(deviceID string) bool

	// GetAllOnlineDevices 获取所有在线设备列表
	// 返回: 在线设备列表、错误
	GetAllOnlineDevices() ([]core.UnboundDevice, error)
}

// NewManager 创建MQTT管理器实例
func NewManager(config Config, configStore engine.Engine, redisClient *redis.Redis, dispatchInfoFunc core.DispatchInfoFunc, storeSensorDataFunc core.StoreSensorDataFunc) (Manager, error) {
	return newMqttManager(config, configStore, redisClient, dispatchInfoFunc, storeSensorDataFunc)
}
