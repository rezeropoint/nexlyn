package httpreceive

import (
	"fmt"
	"strconv"
	"strings"
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

// getFieldValue 从JSON数据中提取字段值（支持点分隔路径和数组索引）
//
// 支持的路径格式：
// - "Key" - 简单字段
// - "data.temperature" - 嵌套对象
// - "Result.Tags[0]" - 数组索引
// - "data.items[2].name" - 混合使用（数组中对象的字段）
// - "matrix[0][1]" - 多维数组
//
// 示例：
//   data := map[string]any{
//     "Result": map[string]any{
//       "Tags": []any{"tag1", "tag2", "tag3"},
//     },
//   }
//   value, exists := getFieldValue(data, "Result.Tags[0]")  // value = "tag1", exists = true
//   value, exists := getFieldValue(data, "Result.Tags[10]") // value = nil, exists = false
//
// 返回: 字段值（any），是否存在（bool）
func getFieldValue(data map[string]any, fieldPath string) (any, bool) {
	if fieldPath == "" {
		return nil, false
	}

	// 解析路径（支持数组索引）
	parts := parseFieldPath(fieldPath)
	if len(parts) == 0 {
		return nil, false
	}

	// 遍历路径
	var current any = data
	for _, part := range parts {
		switch part.Type {
		case pathTypeField:
			// 字段访问
			currentMap, ok := current.(map[string]any)
			if !ok {
				return nil, false
			}
			value, exists := currentMap[part.Value]
			if !exists {
				return nil, false
			}
			current = value

		case pathTypeArrayIndex:
			// 数组索引访问
			currentArray, ok := current.([]any)
			if !ok {
				return nil, false
			}
			index, err := strconv.Atoi(part.Value)
			if err != nil || index < 0 || index >= len(currentArray) {
				return nil, false
			}
			current = currentArray[index]
		}
	}

	return current, true
}

// pathPartType 路径部分类型
type pathPartType int

const (
	pathTypeField      pathPartType = iota // 字段访问（如：data）
	pathTypeArrayIndex                     // 数组索引访问（如：[0]）
)

// pathPart 路径部分
type pathPart struct {
	Type  pathPartType
	Value string
}

// parseFieldPath 解析字段路径为路径部分列表
// 示例：
//   "data.items[0].name" -> [{Field, "data"}, {Field, "items"}, {ArrayIndex, "0"}, {Field, "name"}]
//   "Result.Tags[2]" -> [{Field, "Result"}, {Field, "Tags"}, {ArrayIndex, "2"}]
func parseFieldPath(path string) []pathPart {
	var parts []pathPart
	var currentField strings.Builder
	inBracket := false

	for i := 0; i < len(path); i++ {
		ch := path[i]

		switch ch {
		case '.':
			// 点号：分隔符
			if inBracket {
				// 在括号内，作为普通字符
				currentField.WriteByte(ch)
			} else {
				// 保存当前字段
				if currentField.Len() > 0 {
					parts = append(parts, pathPart{
						Type:  pathTypeField,
						Value: currentField.String(),
					})
					currentField.Reset()
				}
			}

		case '[':
			// 左方括号：开始数组索引
			if inBracket {
				// 嵌套括号，不支持
				return nil
			}
			// 保存当前字段
			if currentField.Len() > 0 {
				parts = append(parts, pathPart{
					Type:  pathTypeField,
					Value: currentField.String(),
				})
				currentField.Reset()
			}
			inBracket = true

		case ']':
			// 右方括号：结束数组索引
			if !inBracket {
				// 未匹配的右括号
				return nil
			}
			// 保存数组索引
			if currentField.Len() > 0 {
				parts = append(parts, pathPart{
					Type:  pathTypeArrayIndex,
					Value: currentField.String(),
				})
				currentField.Reset()
			}
			inBracket = false

		default:
			// 普通字符
			currentField.WriteByte(ch)
		}
	}

	// 保存最后一个字段
	if currentField.Len() > 0 {
		if inBracket {
			// 未闭合的括号
			return nil
		}
		parts = append(parts, pathPart{
			Type:  pathTypeField,
			Value: currentField.String(),
		})
	}

	return parts
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
