package device

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Manager 设备绑定管理器接口
type Manager interface {
	// Bind 绑定设备（创建设备实例）
	Bind(ctx context.Context, metadata core.DeviceBindingMetadata) (string, error)

	// Get 获取设备详情
	Get(ctx context.Context, deviceID string, tenantID string, orgIDs []string) (*core.DeviceBinding, error)

	// List 查询设备列表
	List(ctx context.Context, query core.DeviceBindingQuery) ([]*core.DeviceBindingSummary, int64, error)

	// Update 更新设备信息
	Update(ctx context.Context, deviceID string, tenantID string, orgIDs []string, update core.DeviceBindingUpdate) error

	// Delete 删除设备（物理删除，解绑即删除）
	Delete(ctx context.Context, deviceID string, tenantID string, orgIDs []string) error

	// ListUnboundDevices 获取未绑定设备列表（当前在线但未绑定的设备）
	ListUnboundDevices(ctx context.Context, tenantID string) ([]core.UnboundDevice, error)

	// GetDeviceInfoForController 获取设备基本信息（用于Controller Manager）
	// 只查询设备的category、model、tenant_id等基本信息，不进行组织权限验证
	// 注意：权限验证应在Engine层的AI Box控制方法中统一处理
	GetDeviceInfoForController(ctx context.Context, deviceID string) (*core.DeviceBinding, error)
}

// NewManager 创建设备管理器实例
// getDeviceOnlineStatus: 获取设备在线状态函数（由MQTT Manager提供）
// getAllOnlineDevices: 获取所有在线设备函数（由MQTT Manager提供）
func NewManager(dbConn sqlx.SqlConn, redisClient *redis.Redis, config Config, getDeviceOnlineStatus core.GetDeviceOnlineStatusFunc, getAllOnlineDevices core.GetAllOnlineDevicesFunc) (Manager, error) {
	return newDeviceManager(dbConn, redisClient, config, getDeviceOnlineStatus, getAllOnlineDevices)
}
