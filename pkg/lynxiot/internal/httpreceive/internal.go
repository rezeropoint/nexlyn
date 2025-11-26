package httpreceive

import (
	"context"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/zeromicro/go-zero/core/logx"
)

// ===============================
// 字段提取和时间戳解析
// ===============================

// extractFieldsByMapping 根据字段映射配置提取所有字段
// 返回: 自定义字段名 -> 映射后的值
// 注意：HTTP Receive 不需要标准字段验证，处理任意格式的 JSON 数据
func (m *httpReceiveManager) extractFieldsByMapping(ctx context.Context, data map[string]any, mappings []core.HttpReceiveFieldMapping) map[string]any {
	result := make(map[string]any)

	for _, mapping := range mappings {
		// 1. 使用 getFieldValue 提取源字段值
		rawValue, exists := getFieldValue(data, mapping.SourcePath)

		// 2. 如果字段不存在，使用默认值
		if !exists {
			if mapping.DefaultValue != "" {
				result[mapping.FieldName] = mapping.DefaultValue
				logx.WithContext(ctx).WithFields(
					logx.Field("service", m.config.ServiceName),
					logx.Field("pod", m.config.PodName),
					logx.Field("module", "http_receive"),
					logx.Field("operation", "extract_field"),
					logx.Field("field_name", mapping.FieldName),
					logx.Field("source_path", mapping.SourcePath),
					logx.Field("default_value", mapping.DefaultValue),
				).Debug("字段不存在，使用默认值")
			} else {
				logx.WithContext(ctx).WithFields(
					logx.Field("service", m.config.ServiceName),
					logx.Field("pod", m.config.PodName),
					logx.Field("module", "http_receive"),
					logx.Field("operation", "extract_field"),
					logx.Field("status", "skipped"),
					logx.Field("field_name", mapping.FieldName),
					logx.Field("source_path", mapping.SourcePath),
				).Debug("字段不存在且无默认值，跳过")
			}
			continue
		}

		// 3. 直接存储原始值（不进行缩放和偏移处理）
		result[mapping.FieldName] = rawValue
	}

	return result
}

// parseTimestamp 解析时间戳字段
// 返回: Unix时间戳（秒）
func (m *httpReceiveManager) parseTimestamp(ctx context.Context, data map[string]any, timestampPath string, format core.TimestampFormat) int64 {
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
			logx.Field("module", "http_receive"),
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

	case core.TimestampFormatUnixNano:
		// Unix时间戳（纳秒），转换为秒
		var ns int64
		ns, err = convertToInt64(rawValue)
		if err == nil {
			timestamp = ns / 1e9
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
			logx.Field("module", "http_receive"),
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

// convertToTypedValue 将字段值映射转换为 TypedValue 格式
// 用于数据分发
// 参数：extractedFields - 自定义字段名到值的映射
func (m *httpReceiveManager) convertToTypedValue(extractedFields map[string]any) map[string]core.TypedValue {
	result := make(map[string]core.TypedValue)
	for fieldName, value := range extractedFields {
		result[fieldName] = core.TypedValue{
			Value: value,
			Type:  detectValueType(value),
		}
	}
	return result
}

// detectValueType 检测值的类型
func detectValueType(value any) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int32, int64:
		return "int"
	case float32, float64:
		return "float"
	case bool:
		return "bool"
	default:
		return "unknown"
	}
}
