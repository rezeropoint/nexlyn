package core

import "fmt"

// ===============================
// Redis键格式常量和工具函数
// ===============================

// RedisKeyPrefixMQTT MQTT相关Redis键前缀
const (
	RedisKeyPrefixMQTT           = "iot:mqtt"            // MQTT基础前缀
	RedisKeyPrefixMQTTSubscribed = "iot:mqtt:subscribed" // 订阅主题前缀
	RedisKeyPrefixMQTTLock       = "iot:mqtt:lock"       // 消息处理锁前缀
)

// BuildSubscribedTopicsKey 构建订阅主题集合的Redis键
// 格式: iot:mqtt:subscribed:{pod_name}:topics
// 用途: 存储每个Pod已订阅的MQTT主题列表（Set类型）
func BuildSubscribedTopicsKey(podName string) string {
	return fmt.Sprintf("%s:%s:topics", RedisKeyPrefixMQTTSubscribed, podName)
}

// BuildMessageLockKey 构建消息处理锁的Redis键
// 格式: iot:mqtt:lock:{topic}:{time_window_ms}
// 参数:
//   - topic: MQTT主题
//   - timestampMs: 当前时间戳（毫秒）
//   - windowMs: 时间窗口大小（毫秒）
// 用途: 基于时间窗口的分布式锁，防止多副本重复处理同一消息
// 原理: 将时间戳向下取整到时间窗口，同一窗口内的消息生成相同的锁键
func BuildMessageLockKey(topic string, timestampMs int64, windowMs int64) string {
	// 默认时间窗口100ms
	if windowMs <= 0 {
		windowMs = 100
	}
	// 将时间戳向下取整到时间窗口
	timeWindow := (timestampMs / windowMs) * windowMs
	return fmt.Sprintf("%s:%s:%d", RedisKeyPrefixMQTTLock, topic, timeWindow)
}
