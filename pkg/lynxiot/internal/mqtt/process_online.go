package mqtt

import (
	"context"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core/fields"

	etcdtriggercore "github.com/rezeropoint/etcdtrigger/v2/core"

	"github.com/zeromicro/go-zero/core/logx"
)

// onOnlineConfigChange 在线检测配置变化回调
func (m *mqttManager) onOnlineConfigChange(key string, eventType etcdtriggercore.EventType) {
	ctx := context.Background()

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_mqtt_manager"),
		logx.Field("operation", "on_online_config_change"),
		logx.Field("key", key),
		logx.Field("event_type", eventType.String()),
	).Info("检测到在线检测配置变化")

	if eventType.IsPut() { // 配置创建或更新
		// 从ConfigManager获取配置
		var config core.OnlineDetectionConfig
		if !m.configStore.GetConfig(key, &config) {
			logx.WithContext(ctx).WithFields(
				logx.Field("service", m.config.ServiceName),
				logx.Field("pod", m.config.PodName),
				logx.Field("module", "iot_mqtt_manager"),
				logx.Field("operation", "load_online_config"),
				logx.Field("status", "failed"),
				logx.Field("key", key),
				logx.Field("error", "未找到对应的配置"),
			).Error("未找到在线检测配置")
			return
		}

		// 构建订阅主题: nexlyn/iot/{category}/{model}/+/{topicSuffix}
		topic := core.BuildDeviceTopicPattern(core.DeviceCategory(config.Category), config.Model, config.TopicSuffix)

		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "load_online_config"),
			logx.Field("config_key", key),
			logx.Field("category", config.Category),
			logx.Field("model", config.Model),
			logx.Field("topic_suffix", config.TopicSuffix),
			logx.Field("topic_pattern", topic),
		).Info("成功加载在线检测配置")

		// 订阅主题
		m.subscribeIfNeeded(topic)

	} else if eventType.IsDelete() { // 配置删除
		// 配置已删除，不需要特殊处理
		// MQTT订阅会继续有效，但processMessage会因为找不到配置而跳过处理
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "delete_online_config"),
			logx.Field("key", key),
		).Info("在线检测配置已删除")
	}
}

// processOnlineDetection 处理在线检测（已解析参数，无重复解析）
func (m *mqttManager) processOnlineDetection(
	ctx context.Context,
	onlineConfig *core.OnlineDetectionConfig,
	category core.DeviceCategory,
	model string,
	deviceID string,
	configKey string,
	data map[string]any,
) {
	// 1. 记录DEBUG日志
	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_online_detection"),
		logx.Field("operation", "process_online_detection"),
		logx.Field("category", category),
		logx.Field("model", model),
		logx.Field("device_id", deviceID),
		logx.Field("config_key", configKey),
	).Debug("开始处理在线检测")

	// 2. 根据策略检查设备是否在线
	isOnline := checkOnlineStatus(data, *onlineConfig)

	if !isOnline {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_online_detection"),
			logx.Field("operation", "process_online_detection"),
			logx.Field("status", "skipped"),
			logx.Field("device_id", deviceID),
			logx.Field("category", category),
			logx.Field("reason", "字段检查未通过"),
		).Debug("设备在线检查未通过，不更新在线状态")
		return
	}

	// 3. 更新Redis在线状态
	timestamp := time.Now().Unix()
	ttl := onlineConfig.TimeoutSeconds
	if err := m.SetDeviceOnlineStatus(deviceID, string(category), true, timestamp, ttl); err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_online_detection"),
			logx.Field("operation", "update_online_status"),
			logx.Field("status", "failed"),
			logx.Field("device_id", deviceID),
			logx.Field("category", category),
			logx.Field("error", err.Error()),
		).Error("写入设备在线状态到Redis失败")
		return
	}

	// 4. 记录成功日志
	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_online_detection"),
		logx.Field("operation", "update_online_status"),
		logx.Field("status", "success"),
		logx.Field("device_id", deviceID),
		logx.Field("category", category),
		logx.Field("model", model),
		logx.Field("config_key", configKey),
		logx.Field("strategy", onlineConfig.Strategy),
		logx.Field("timestamp", timestamp),
	).Info("成功更新设备在线状态")

	// 5. 存储在线状态到ClickHouse
	if m.StoreSensorDataFunc != nil {
		record := &core.SensorDataRecord{
			Timestamp:      time.Unix(timestamp, 0),
			DeviceID:       deviceID,
			DeviceModel:    model,
			DeviceCategory: category,
			TenantID:       onlineConfig.TenantID,
			Fields: map[core.FieldName]any{
				fields.FieldNameIsOnline: true, // 使用平台统一的标准字段名
			},
		}

		if err := m.StoreSensorDataFunc(ctx, record); err != nil {
			logx.WithContext(ctx).WithFields(
				logx.Field("service", m.config.ServiceName),
				logx.Field("pod", m.config.PodName),
				logx.Field("module", "iot_online_detection"),
				logx.Field("operation", "store_online_status"),
				logx.Field("status", "failed"),
				logx.Field("device_id", deviceID),
				logx.Field("category", category),
				logx.Field("model", model),
				logx.Field("tenant_id", onlineConfig.TenantID),
				logx.Field("error", err.Error()),
			).Error("存储在线状态到ClickHouse失败")
			// 存储失败不影响在线状态更新，只记录日志
		} else {
			logx.WithContext(ctx).WithFields(
				logx.Field("service", m.config.ServiceName),
				logx.Field("pod", m.config.PodName),
				logx.Field("module", "iot_online_detection"),
				logx.Field("operation", "store_online_status"),
				logx.Field("status", "success"),
				logx.Field("device_id", deviceID),
				logx.Field("category", category),
				logx.Field("model", model),
				logx.Field("tenant_id", onlineConfig.TenantID),
			).Debug("成功存储在线状态到ClickHouse")
		}
	}
}
