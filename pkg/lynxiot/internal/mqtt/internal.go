package mqtt

import (
	"context"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core/fields"

	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/logx"
)

// ===============================
// 订阅主题管理（内存缓存 + Redis持久化双层架构）
// ===============================

// isTopicSubscribed 检查主题是否已订阅
// 使用内存Set（go-zero 1.9泛型特性）作为缓存层，提升查询性能
func (m *mqttManager) isTopicSubscribed(topic string) (bool, error) {
	// 先从内存缓存查询（O(1)时间复杂度）
	m.mu.RLock()
	cached := m.subscribedTopics.Contains(topic)
	m.mu.RUnlock()

	if cached {
		return true, nil
	}

	// 缓存未命中，查询Redis（防止Pod重启后缓存丢失）
	key := core.BuildSubscribedTopicsKey(m.config.PodName)
	exists, err := m.redisClient.Sismember(key, topic)
	if err != nil {
		return false, err
	}

	// 如果Redis中存在，更新内存缓存
	if exists {
		m.mu.Lock()
		m.subscribedTopics.Add(topic)
		m.mu.Unlock()
	}

	return exists, nil
}

// markTopicSubscribed 标记主题已订阅
// 同时更新内存缓存和Redis持久化存储
func (m *mqttManager) markTopicSubscribed(topic string) error {
	// 1. 写入Redis（持久化，跨Pod共享）
	key := core.BuildSubscribedTopicsKey(m.config.PodName)
	_, err := m.redisClient.Sadd(key, topic)
	if err != nil {
		return err
	}

	// 2. 更新内存缓存（加速后续查询）
	m.mu.Lock()
	m.subscribedTopics.Add(topic)
	m.mu.Unlock()

	return nil
}

// clearSubscribedTopics 清除订阅记录
// 同时清除内存缓存和Redis存储
func (m *mqttManager) clearSubscribedTopics() error {
	// 1. 清除Redis
	key := core.BuildSubscribedTopicsKey(m.config.PodName)
	_, err := m.redisClient.Del(key)
	if err != nil {
		return err
	}

	// 2. 清除内存缓存
	m.mu.Lock()
	m.subscribedTopics = collection.NewSet[string]()
	m.mu.Unlock()

	return nil
}

// ===============================
// 分布式锁管理
// ===============================

// tryAcquireLock 尝试获取分布式锁
// 使用Redis SET NX EX命令，返回是否成功获取锁
func (m *mqttManager) tryAcquireLock(ctx context.Context, key string, ttl int) (bool, error) {
	// SET key value NX EX ttl
	// NX: 只在键不存在时设置
	// EX: 设置过期时间（秒）
	ok, err := m.redisClient.SetnxExCtx(ctx, key, "1", ttl)
	if err != nil {
		return false, fmt.Errorf("获取分布式锁失败: %w", err)
	}
	return ok, nil
}

// releaseLock 释放分布式锁
func (m *mqttManager) releaseLock(ctx context.Context, key string) error {
	_, err := m.redisClient.DelCtx(ctx, key)
	if err != nil {
		return fmt.Errorf("释放分布式锁失败: %w", err)
	}
	return nil
}

// ===============================
// 字段提取和时间戳解析
// ===============================

// extractFieldsByMapping 根据字段映射配置提取所有字段
// 返回: 标准字段名 -> 映射后的值
func (m *mqttManager) extractFieldsByMapping(ctx context.Context, data map[string]any, mappings []core.FieldMapping) map[core.FieldName]any {
	result := make(map[core.FieldName]any)

	for i, mapping := range mappings {
		// 0. 验证字段是否在注册表中（运行时验证）
		if err := fields.ValidateFieldExists(mapping.StandardField); err != nil {
			logx.WithContext(ctx).WithFields(
				logx.Field("service", m.config.ServiceName),
				logx.Field("pod", m.config.PodName),
				logx.Field("module", "iot_business_data"),
				logx.Field("operation", "validate_field"),
				logx.Field("status", "failed"),
				logx.Field("mapping_index", i),
				logx.Field("field_name", string(mapping.StandardField)),
				logx.Field("error", err.Error()),
			).Error("字段验证失败：字段未在注册表中注册")
			// 跳过未注册的字段，继续处理其他字段
			continue
		}

		// 1. 使用 core.GetFieldValue 提取源字段值
		rawValue, exists := getFieldValue(data, mapping.SourcePath)

		// 2. 如果字段不存在，使用默认值
		if !exists {
			if mapping.DefaultValue != "" {
				result[mapping.StandardField] = mapping.DefaultValue
				logx.WithContext(ctx).WithFields(
					logx.Field("service", m.config.ServiceName),
					logx.Field("pod", m.config.PodName),
					logx.Field("module", "iot_business_data"),
					logx.Field("operation", "extract_field"),
					logx.Field("standard_field", mapping.StandardField),
					logx.Field("source_path", mapping.SourcePath),
					logx.Field("default_value", mapping.DefaultValue),
				).Debug("字段不存在，使用默认值")
			} else {
				logx.WithContext(ctx).WithFields(
					logx.Field("service", m.config.ServiceName),
					logx.Field("pod", m.config.PodName),
					logx.Field("module", "iot_business_data"),
					logx.Field("operation", "extract_field"),
					logx.Field("status", "skipped"),
					logx.Field("standard_field", mapping.StandardField),
					logx.Field("source_path", mapping.SourcePath),
				).Debug("字段不存在且无默认值，跳过")
			}
			continue
		}

		// 3. 尝试将值转换为 float64（用于 Scale 和 Offset 计算）
		var finalValue any = rawValue

		// 如果配置了 Scale 或 Offset，尝试转换为数值类型
		if mapping.Scale != 0 || mapping.Offset != 0 {
			numValue, err := convertToFloat64(rawValue)
			if err != nil {
				logx.WithContext(ctx).WithFields(
					logx.Field("service", m.config.ServiceName),
					logx.Field("pod", m.config.PodName),
					logx.Field("module", "iot_business_data"),
					logx.Field("operation", "extract_field"),
					logx.Field("status", "warning"),
					logx.Field("standard_field", mapping.StandardField),
					logx.Field("source_path", mapping.SourcePath),
					logx.Field("raw_value", rawValue),
					logx.Field("error", err.Error()),
				).Error("无法将字段值转换为数值类型，跳过Scale/Offset处理")
			} else {
				// 应用 Scale 和 Offset
				if mapping.Scale != 0 {
					numValue = numValue * mapping.Scale
				}
				numValue = numValue + mapping.Offset
				finalValue = numValue
			}
		}

		// 4. 存储映射后的值
		result[mapping.StandardField] = finalValue
	}

	return result
}

// parseTimestamp 解析时间戳字段
// 返回: Unix时间戳（秒）
func (m *mqttManager) parseTimestamp(ctx context.Context, data map[string]any, timestampPath string, format core.TimestampFormat) int64 {
	// 如果未配置时间戳路径，使用当前时间
	if timestampPath == "" {
		return time.Now().Unix()
	}

	// 提取时间戳字段
	rawValue, exists := getFieldValue(data, timestampPath)
	if !exists {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_business_data"),
			logx.Field("operation", "parse_timestamp"),
			logx.Field("status", "warning"),
			logx.Field("timestamp_path", timestampPath),
		).Error("时间戳字段不存在，使用当前时间")
		return time.Now().Unix()
	}

	// 根据格式解析时间戳
	var timestamp int64
	var err error

	switch format {
	case core.TimestampFormatUnix:
		// Unix时间戳（秒）
		timestamp, err = convertToInt64(rawValue)

	case core.TimestampFormatUnixMs:
		// Unix时间戳（毫秒），转换为秒
		var ms int64
		ms, err = convertToInt64(rawValue)
		if err == nil {
			timestamp = ms / 1000
		}

	case core.TimestampFormatUnixMicro:
		// Unix时间戳（微秒），转换为秒
		var us int64
		us, err = convertToInt64(rawValue)
		if err == nil {
			timestamp = us / 1e6
		}

	case core.TimestampFormatISO8601, core.TimestampFormatRFC3339:
		// ISO8601 和 RFC3339 格式字符串
		strValue, ok := rawValue.(string)
		if !ok {
			err = fmt.Errorf("时间戳字段不是字符串类型")
		} else {
			var t time.Time
			t, err = time.Parse(time.RFC3339, strValue)
			if err == nil {
				timestamp = t.Unix()
			}
		}

	default:
		// 未知格式，尝试作为Unix时间戳处理
		timestamp, err = convertToInt64(rawValue)
	}

	// 解析失败，使用当前时间
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_business_data"),
			logx.Field("operation", "parse_timestamp"),
			logx.Field("status", "warning"),
			logx.Field("timestamp_path", timestampPath),
			logx.Field("format", format),
			logx.Field("raw_value", rawValue),
			logx.Field("error", err.Error()),
		).Error("时间戳解析失败，使用当前时间")
		return time.Now().Unix()
	}

	return timestamp
}
