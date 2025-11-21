// Package core 定义IoT设备管理的核心领域模型
//
// 文件说明：controller.go
// 职责：控制相关的函数类型定义（极简版，参考MQTT Manager设计）
package core

import "fmt"

// 控制主题和Redis键格式常量
const (
	// ControlTopicFormat 控制主题格式
	// 格式: nexlyn/iot/{category}/{model}/{device_id}/{suffix}
	// 示例: nexlyn/iot/ai_box/AI-200/box001/control
	ControlTopicFormat = "github.com/rezeropoint/nexlyn/iot/%s/%s/%s/%s"

	// ControlDeviceRedisKeyFormat 控制设备Redis键格式（存储当前正在执行的命令）
	// 格式: iot:control:device:{device_id}
	ControlDeviceRedisKeyFormat = "iot:control:device:%s"

	// ControlLockRedisKeyFormat 控制命令分布式锁Redis键格式
	// 格式: iot:control:lock:{device_id}
	// 用途：多Pod环境下协调同一设备的命令执行，防止并发冲突
	ControlLockRedisKeyFormat = "iot:control:lock:%s"

	// CapabilitiesCacheRedisKeyFormat 算法能力缓存Redis键格式
	// 格式: iot:capabilities:device:{device_id}
	CapabilitiesCacheRedisKeyFormat = "iot:capabilities:device:%s"
)

// BuildControlTopic 构建控制主题
func BuildControlTopic(category, model, deviceID, suffix string) string {
	return fmt.Sprintf(ControlTopicFormat, category, model, deviceID, suffix)
}

// BuildControlDeviceRedisKey 构建控制设备Redis键
func BuildControlDeviceRedisKey(deviceID string) string {
	return fmt.Sprintf(ControlDeviceRedisKeyFormat, deviceID)
}

// BuildControlLockKey 构建控制命令分布式锁Redis键
func BuildControlLockKey(deviceID string) string {
	return fmt.Sprintf(ControlLockRedisKeyFormat, deviceID)
}

// BuildCapabilitiesCacheKey 构建算法能力缓存Redis键
func BuildCapabilitiesCacheKey(deviceID string) string {
	return fmt.Sprintf(CapabilitiesCacheRedisKeyFormat, deviceID)
}

// GetDeviceInfoFunc 获取设备绑定信息的函数类型
// 用于Controller Manager获取设备的category、model、tenant_id等信息
// 参数：deviceID - 设备ID
// 返回：设备绑定信息、错误
type GetDeviceInfoFunc func(deviceID string) (*DeviceBinding, error)
