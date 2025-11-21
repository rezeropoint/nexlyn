package core

// Redis键格式常量
const (
	// RedisKeyDeviceOnlinePrefix 设备在线状态Redis键前缀
	RedisKeyDeviceOnlinePrefix = "iot:device:online:"
)

// DeviceOnlineInfo 设备在线信息
type DeviceOnlineInfo struct {
	DeviceID              string `json:"deviceId"`                        // 设备唯一标识符
	IsOnline              bool   `json:"isOnline"`                        // 当前在线状态
	OnlineStatusUpdatedAt string `json:"onlineStatusUpdatedAt,omitempty"` // 在线状态最后更新时间
	LastOnlineAt          string `json:"lastOnlineAt,omitempty"`          // 最后在线时间
	LastOfflineAt         string `json:"lastOfflineAt,omitempty"`         // 最后离线时间
	LastDataAt            string `json:"lastDataAt,omitempty"`            // 最后接收数据时间
}

// DeviceStatusMessage MQTT在线状态消息
type DeviceStatusMessage struct {
	DeviceID  string `json:"device_id"` // 设备ID（必填）
	IsOnline  bool   `json:"is_online"` // 在线状态（必填）
	Timestamp int64  `json:"timestamp"` // 消息时间戳（Unix秒，必填）
}

// DeviceOnlineRedisData Redis存储的设备在线数据
type DeviceOnlineRedisData struct {
	DeviceID  string `json:"device_id"` // 设备ID
	Category  string `json:"category"`  // 设备类别
	IsOnline  bool   `json:"is_online"` // 在线状态
	Timestamp int64  `json:"timestamp"` // 时间戳（Unix秒）
}

// UnboundDevice 未绑定设备信息
type UnboundDevice struct {
	DeviceID string `json:"deviceId"` // 设备ID
	Category string `json:"category"` // 设备类别
	IsOnline bool   `json:"isOnline"` // 在线状态
	LastSeen string `json:"lastSeen"` // 最后上报时间
}

// 设备在线状态缓存相关函数类型（用于解耦MQTT Manager和Device Manager）
// 这些函数由MQTT Manager实现，在Engine层注入到Device Manager

// GetDeviceOnlineStatusFunc 获取设备在线状态的函数类型
// 参数：deviceID - 设备ID
// 返回：在线状态（true/false）
type GetDeviceOnlineStatusFunc func(deviceID string) bool

// GetAllOnlineDevicesFunc 获取所有在线设备的函数类型
// 返回：在线设备列表、错误
type GetAllOnlineDevicesFunc func() ([]UnboundDevice, error)

// MQTT主题常量
const (
	// MQTTTopicPrefix MQTT主题前缀
	MQTTTopicPrefix = "github.com/rezeropoint/nexlyn/iot"

	// MQTTTopicStatusOnline 在线状态主题后缀
	MQTTTopicStatusOnline = "status/online"
)

// BuildDeviceStatusTopic 构建设备在线状态主题
// 格式: nexlyn/iot/{category}/{device_id}/status/online
func BuildDeviceStatusTopic(category DeviceCategory, deviceID string) string {
	return MQTTTopicPrefix + "/" + string(category) + "/" + deviceID + "/" + MQTTTopicStatusOnline
}

// BuildDeviceStatusTopicWildcard 构建设备在线状态主题通配符
// 格式: nexlyn/iot/+/+/status/online （订阅所有设备的在线状态）
func BuildDeviceStatusTopicWildcard() string {
	return MQTTTopicPrefix + "/+/+/" + MQTTTopicStatusOnline
}

// TopicInfo 解析后的主题信息
type TopicInfo struct {
	Category DeviceCategory // 设备类别
	DeviceID string         // 设备ID
	Type     string         // 消息类型（status/online）
}

// BuildDeviceOnlineRedisKey 构建设备在线状态Redis键
func BuildDeviceOnlineRedisKey(deviceID string) string {
	return RedisKeyDeviceOnlinePrefix + deviceID
}

// BuildDeviceTopicPattern 构建设备主题模式（用于订阅）
// 格式: nexlyn/iot/{category}/{model}/+/{topicSuffix}
func BuildDeviceTopicPattern(category DeviceCategory, model string, topicSuffix string) string {
	return MQTTTopicPrefix + "/" + string(category) + "/" + model + "/+/" + topicSuffix
}
