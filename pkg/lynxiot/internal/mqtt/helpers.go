package mqtt

import (
	"fmt"
	"strings"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
)

// convertToFloat64 将任意类型转换为 float64
func convertToFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case string:
		// 尝试解析字符串为数字
		var num float64
		_, err := fmt.Sscanf(v, "%f", &num)
		return num, err
	default:
		return 0, fmt.Errorf("无法将类型 %T 转换为 float64", value)
	}
}

// convertToInt64 将任意类型转换为 int64
func convertToInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case string:
		// 尝试解析字符串为整数
		var num int64
		_, err := fmt.Sscanf(v, "%d", &num)
		return num, err
	default:
		return 0, fmt.Errorf("无法将类型 %T 转换为 int64", value)
	}
}

// formatTimestamp 将Unix时间戳格式化为字符串
func formatTimestamp(timestamp int64) string {
	if timestamp == 0 {
		return ""
	}
	// 使用time包格式化（需要import "time"）
	// 这里简化处理，只返回时间戳字符串
	return fmt.Sprintf("%d", timestamp)
}

// parseDeviceTopic 解析设备主题，提取category、model、device_id和suffix
// 期望格式: nexlyn/iot/{category}/{model}/{device_id}/{suffix...}
// 返回: category, model, deviceID, suffix, error
func parseDeviceTopic(topic string) (core.DeviceCategory, string, string, string, error) {
	// 去除前后空格
	topic = strings.TrimSpace(topic)

	// 分割主题
	parts := strings.Split(topic, "/")

	// 至少需要6段：nexlyn/iot/{category}/{model}/{device_id}/{suffix}
	if len(parts) < 6 {
		return "", "", "", "", fmt.Errorf("无效的主题格式: %s，期望格式: nexlyn/iot/{category}/{model}/{device_id}/{suffix}", topic)
	}

	// 验证前缀
	if parts[0] != "nexlyn" || parts[1] != "iot" {
		return "", "", "", "", fmt.Errorf("无效的主题前缀: %s/%s，期望: nexlyn/iot", parts[0], parts[1])
	}

	// 提取信息
	category := core.DeviceCategory(parts[2])
	model := parts[3]
	deviceID := parts[4]
	suffix := strings.Join(parts[5:], "/")

	// 验证必填字段
	if category == "" {
		return "", "", "", "", fmt.Errorf("设备类别不能为空")
	}
	if model == "" {
		return "", "", "", "", fmt.Errorf("设备型号不能为空")
	}
	if deviceID == "" {
		return "", "", "", "", fmt.Errorf("设备ID不能为空")
	}

	return category, model, deviceID, suffix, nil
}

// getFieldValue 从JSON数据中提取字段值（支持点分隔路径）
// fieldPath: 字段路径，如 "status.online", "Key", "data.temperature"
// 返回: 字段值（any），是否存在（bool）
func getFieldValue(data map[string]any, fieldPath string) (any, bool) {
	if fieldPath == "" {
		return nil, false
	}

	// 分割路径
	parts := strings.Split(fieldPath, ".")

	// 遍历路径
	current := data
	for i, part := range parts {
		value, exists := current[part]
		if !exists {
			return nil, false
		}

		// 如果不是最后一段，必须是map类型
		if i < len(parts)-1 {
			nextMap, ok := value.(map[string]any)
			if !ok {
				return nil, false
			}
			current = nextMap
		} else {
			// 最后一段，返回值
			return value, true
		}
	}

	return nil, false
}

// checkFieldValue 检查字段值是否符合预期
// 支持任意字段类型：string, int, float, bool等
func checkFieldValue(data map[string]any, check core.FieldCheck) bool {
	// 获取字段值
	value, exists := getFieldValue(data, check.Field)

	// 根据操作符检查
	switch check.Operator {
	case core.FieldCheckOperatorExists:
		return exists

	case core.FieldCheckOperatorEquals:
		if !exists {
			return false
		}
		// 将字段值标准化为字符串进行比较
		// 支持 string, int, float, bool 等基本类型
		return normalizeValue(value) == check.Value

	case core.FieldCheckOperatorIn:
		if !exists {
			return false
		}
		// 将字段值标准化为字符串，检查是否在列表中
		normalizedValue := normalizeValue(value)
		for _, expectedValue := range check.Values {
			if normalizedValue == expectedValue {
				return true
			}
		}
		return false

	default:
		return false
	}
}

// normalizeValue 将任意类型的值标准化为字符串
// 支持: string, int, int64, float64, bool, nil 等
func normalizeValue(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		// JSON数字默认解析为float64
		// 如果是整数，不显示小数点
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%g", v)
	case float32:
		if v == float32(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%g", v)
	default:
		// 其他类型（map, slice等）使用默认格式化
		return fmt.Sprintf("%v", v)
	}
}

// checkOnlineStatus 根据策略和字段检查判断设备是否在线
// data: MQTT消息体（已解析为JSON）
// config: 在线检测配置
// 返回: 是否在线
func checkOnlineStatus(data map[string]any, config core.OnlineDetectionConfig) bool {
	switch config.Strategy {
	case core.OnlineStrategyAnyData:
		// 收到任何数据即认为在线
		return true

	case core.OnlineStrategyFieldCheck:
		// 检查所有字段（所有检查都必须通过）
		for _, check := range config.FieldChecks {
			if !checkFieldValue(data, check) {
				return false
			}
		}
		return true

	default:
		return false
	}
}
