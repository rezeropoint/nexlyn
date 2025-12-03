package mqtt

import (
	"context"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	etcdtriggercore "github.com/rezeropoint/etcdtrigger/v2/core"

	"github.com/zeromicro/go-zero/core/logx"
)

// onDataConfigChange 业务数据配置变化回调
func (m *mqttManager) onDataConfigChange(key string, eventType etcdtriggercore.EventType) {
	ctx := context.Background()

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_mqtt_manager"),
		logx.Field("operation", "on_data_config_change"),
		logx.Field("key", key),
		logx.Field("event_type", eventType.String()),
	).Info("检测到业务数据配置变化")

	if eventType.IsPut() { // 配置创建或更新
		// 从ConfigManager获取配置
		var config core.DataProcessingConfig
		if !m.configStore.GetConfig(key, &config) {
			logx.WithContext(ctx).WithFields(
				logx.Field("service", m.config.ServiceName),
				logx.Field("pod", m.config.PodName),
				logx.Field("module", "iot_mqtt_manager"),
				logx.Field("operation", "load_data_config"),
				logx.Field("status", "failed"),
				logx.Field("key", key),
				logx.Field("error", "未找到对应的配置"),
			).Error("未找到业务数据配置")
			return
		}

		// 构建订阅主题: nexlyn/iot/{category}/{model}/+/{topicSuffix}
		topic := core.BuildDeviceTopicPattern(core.DeviceCategory(config.Category), config.Model, config.TopicSuffix)

		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "load_data_config"),
			logx.Field("config_key", key),
			logx.Field("category", config.Category),
			logx.Field("model", config.Model),
			logx.Field("topic_suffix", config.TopicSuffix),
			logx.Field("topic_pattern", topic),
		).Info("成功加载业务数据配置")

		// 订阅主题
		m.subscribeIfNeeded(topic)

	} else if eventType.IsDelete() { // 配置删除
		// 配置已删除，不需要特殊处理
		// MQTT订阅会继续有效，但processBusinessData会因为找不到配置而跳过处理
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "delete_data_config"),
			logx.Field("key", key),
		).Info("业务数据配置已删除")
	}
}

// processBusinessData 处理业务数据提取（已解析参数，无重复解析）
func (m *mqttManager) processBusinessData(
	ctx context.Context,
	dataConfig *core.DataProcessingConfig,
	category core.DeviceCategory,
	model string,
	deviceID string,
	topicSuffix string,
	configKey string,
	data map[string]any,
) {
	// 1. 记录DEBUG日志
	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_business_data"),
		logx.Field("operation", "process_business_data"),
		logx.Field("category", category),
		logx.Field("model", model),
		logx.Field("device_id", deviceID),
		logx.Field("topic_suffix", topicSuffix),
		logx.Field("config_key", configKey),
	).Debug("开始处理业务数据")

	// 2. 提取时间戳（如果配置了）
	timestamp := m.parseTimestamp(ctx, data, dataConfig.TimestampPath, dataConfig.TimestampFormat)

	// 3. 提取所有字段
	extractedFields := m.extractFieldsByMapping(ctx, data, dataConfig.FieldMappings)

	if len(extractedFields) == 0 {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_business_data"),
			logx.Field("operation", "extract_fields"),
			logx.Field("status", "warning"),
			logx.Field("device_id", deviceID),
			logx.Field("category", category),
			logx.Field("model", model),
		).Error("未提取到任何字段")
		return
	}

	// 4. 记录成功日志（包含提取的字段）
	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_business_data"),
		logx.Field("operation", "extract_fields"),
		logx.Field("status", "success"),
		logx.Field("device_id", deviceID),
		logx.Field("category", category),
		logx.Field("model", model),
		logx.Field("config_key", configKey),
		logx.Field("timestamp", timestamp),
		logx.Field("extracted_fields", extractedFields),
		logx.Field("field_count", len(extractedFields)),
	).Info("成功提取业务数据字段")

	// 8. 将提取的字段转换为 TypedValue 格式（用于过滤和分发）
	// 构建字段名到类型的映射
	fieldTypeMap := make(map[core.FieldName]core.FieldType)
	for _, mapping := range dataConfig.FieldMappings {
		fieldTypeMap[mapping.StandardField] = mapping.FieldType
	}

	typedData := make(map[string]core.TypedValue)
	for fieldName, value := range extractedFields {
		// 优先使用配置中的字段类型，如果没有配置则使用 Go 类型名
		fieldType := string(fieldTypeMap[fieldName])
		if fieldType == "" {
			fieldType = fmt.Sprintf("%T", value)
		}
		typedData[string(fieldName)] = core.TypedValue{
			Type:  fieldType,
			Value: value,
		}
	}

	// 9. 存储数据到ClickHouse（在过滤和分发之前）
	if m.StoreSensorDataFunc != nil {
		record := &core.SensorDataRecord{
			Timestamp:      time.Unix(timestamp, 0),
			DeviceID:       deviceID,
			DeviceModel:    model,
			DeviceCategory: category,
			TenantID:       dataConfig.TenantID,
			Fields:         extractedFields,
		}

		if err := m.StoreSensorDataFunc(ctx, record); err != nil {
			logx.WithContext(ctx).WithFields(
				logx.Field("service", m.config.ServiceName),
				logx.Field("pod", m.config.PodName),
				logx.Field("module", "iot_business_data"),
				logx.Field("operation", "store_sensor_data"),
				logx.Field("status", "failed"),
				logx.Field("device_id", deviceID),
				logx.Field("category", category),
				logx.Field("model", model),
				logx.Field("tenant_id", dataConfig.TenantID),
				logx.Field("error", err.Error()),
			).Error("存储传感器数据失败")
			// 存储失败不阻塞后续的数据分发，继续执行
		}
	}

	// 10. 应用全局过滤规则（如果配置了FilterRules）
	if dataConfig.FilterRules != nil {
		passed, err := applyFilter(ctx, &typedData, *dataConfig.FilterRules)
		if err != nil {
			logx.WithContext(ctx).WithFields(
				logx.Field("service", m.config.ServiceName),
				logx.Field("pod", m.config.PodName),
				logx.Field("module", "iot_business_data"),
				logx.Field("operation", "apply_filter"),
				logx.Field("status", "failed"),
				logx.Field("device_id", deviceID),
				logx.Field("category", category),
				logx.Field("model", model),
				logx.Field("error", err.Error()),
			).Error("过滤规则执行失败")
			return // 过滤失败，跳过分发
		}

		if !passed {
			logx.WithContext(ctx).WithFields(
				logx.Field("service", m.config.ServiceName),
				logx.Field("pod", m.config.PodName),
				logx.Field("module", "iot_business_data"),
				logx.Field("operation", "apply_filter"),
				logx.Field("status", "filtered"),
				logx.Field("device_id", deviceID),
				logx.Field("category", category),
				logx.Field("model", model),
				logx.Field("condition_count", len(dataConfig.FilterRules.Conditions)),
			).Info("数据未通过过滤规则，跳过分发")
			return // 未通过过滤，跳过分发
		}

		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_business_data"),
			logx.Field("operation", "apply_filter"),
			logx.Field("status", "passed"),
			logx.Field("device_id", deviceID),
			logx.Field("category", category),
			logx.Field("model", model),
			logx.Field("condition_count", len(dataConfig.FilterRules.Conditions)),
		).Info("数据通过过滤规则，准备分发")
	}

	// 11. 数据分发（如果配置了DispatchConfigs）
	if len(dataConfig.DispatchConfigs) > 0 && m.DispatchInfoFunc != nil {
		// 构建任务信息（将设备信息编码到TaskId中）
		taskInfo := core.TaskInfo{
			TaskId:     fmt.Sprintf("%s/%s/%s-%d", category, model, deviceID, timestamp),
			ConfigType: "business_data",
			TenantID:   dataConfig.TenantID,
			DeviceID:   deviceID,
			Timestamp:  timestamp * 1000, // 转为毫秒
		}

		// 调用分发函数
		if err := m.DispatchInfoFunc(ctx, dataConfig.DispatchConfigs, &typedData, taskInfo); err != nil {
			logx.WithContext(ctx).WithFields(
				logx.Field("service", m.config.ServiceName),
				logx.Field("pod", m.config.PodName),
				logx.Field("module", "iot_business_data"),
				logx.Field("operation", "dispatch_data"),
				logx.Field("status", "failed"),
				logx.Field("device_id", deviceID),
				logx.Field("category", category),
				logx.Field("model", model),
				logx.Field("dispatch_config_count", len(dataConfig.DispatchConfigs)),
				logx.Field("error", err.Error()),
			).Error("数据分发失败")
		} else {
			logx.WithContext(ctx).WithFields(
				logx.Field("service", m.config.ServiceName),
				logx.Field("pod", m.config.PodName),
				logx.Field("module", "iot_business_data"),
				logx.Field("operation", "dispatch_data"),
				logx.Field("status", "success"),
				logx.Field("device_id", deviceID),
				logx.Field("category", category),
				logx.Field("model", model),
				logx.Field("dispatch_config_count", len(dataConfig.DispatchConfigs)),
			).Info("成功提交数据分发任务")
		}
	}
}
