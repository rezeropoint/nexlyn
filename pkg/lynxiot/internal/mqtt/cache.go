package mqtt

import (
	"encoding/json"
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
)

// SetDeviceOnlineStatus 设置设备在线状态到Redis
// category: 设备类别
// isOnline: true表示在线，false表示离线
// timestamp: 消息时间戳（Unix秒）
// ttl: 过期时间（秒），0表示不过期
func (m *mqttManager) SetDeviceOnlineStatus(deviceID string, category string, isOnline bool, timestamp int64, ttl int) error {

	redisKey := core.BuildDeviceOnlineRedisKey(deviceID)

	if isOnline {
		// 设备在线：写入Redis并设置过期时间
		data := core.DeviceOnlineRedisData{
			DeviceID:  deviceID,
			Category:  category,
			IsOnline:  true,
			Timestamp: timestamp,
		}

		valueBytes, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("序列化数据失败: %w", err)
		}

		if ttl > 0 {
			err = m.redisClient.Setex(redisKey, string(valueBytes), ttl)
		} else {
			err = m.redisClient.Set(redisKey, string(valueBytes))
		}

		if err != nil {
			return fmt.Errorf("写入Redis失败: %w", err)
		}
	} else {
		// 设备离线：删除Redis键
		_, err := m.redisClient.Del(redisKey)
		if err != nil {
			return fmt.Errorf("删除Redis键失败: %w", err)
		}
	}

	return nil
}

// GetDeviceOnlineStatus 从Redis获取设备在线状态
// 返回: 在线状态（true/false）
// 如果Redis中没有记录或已过期，返回false（离线）
func (m *mqttManager) GetDeviceOnlineStatus(deviceID string) bool {
	if m.redisClient == nil {
		return false
	}

	redisKey := core.BuildDeviceOnlineRedisKey(deviceID)

	// 读取Redis值
	value, err := m.redisClient.Get(redisKey)
	if err != nil || value == "" {
		// Redis中没有记录或已过期，认为离线
		return false
	}

	// 解析JSON值
	var data core.DeviceOnlineRedisData
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		// 解析失败，认为离线
		return false
	}

	return data.IsOnline
}

// GetAllOnlineDevices 获取所有在线设备列表
// 扫描Redis中所有在线设备键，返回设备详细信息
func (m *mqttManager) GetAllOnlineDevices() ([]core.UnboundDevice, error) {
	if m.redisClient == nil {
		return nil, fmt.Errorf("Redis客户端未初始化")
	}

	// 使用SCAN命令扫描所有匹配的键（避免阻塞）
	pattern := core.RedisKeyDeviceOnlinePrefix + "*"
	keys, err := m.redisClient.Keys(pattern)
	if err != nil {
		return nil, fmt.Errorf("扫描Redis键失败: %w", err)
	}

	devices := make([]core.UnboundDevice, 0, len(keys))

	// 遍历所有键，获取详细信息
	for _, key := range keys {
		value, err := m.redisClient.Get(key)
		if err != nil || value == "" {
			// 键已过期或读取失败，跳过
			continue
		}

		// 解析数据
		var data core.DeviceOnlineRedisData
		if err := json.Unmarshal([]byte(value), &data); err != nil {
			// 解析失败，跳过
			continue
		}

		// 只返回在线设备
		if !data.IsOnline {
			continue
		}

		// 构建未绑定设备信息
		device := core.UnboundDevice{
			DeviceID: data.DeviceID,
			Category: data.Category,
			IsOnline: data.IsOnline,
			LastSeen: formatTimestamp(data.Timestamp),
		}

		devices = append(devices, device)
	}

	return devices, nil
}
