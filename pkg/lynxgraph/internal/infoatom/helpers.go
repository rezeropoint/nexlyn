package infoatom

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
)

// generateCacheKey 生成缓存键
func generateCacheKey(key core.InfoAtomTypeKey) string {
	return fmt.Sprintf("infoatom:%s", key.ID)
}

// nullStringToString 将 sql.NullString 转换为 string
// 如果 NULL 则返回空字符串
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// extractFields 根据 DataFormat 提取字段并验证
//
// 这是纯函数辅助方法（无 receiver），符合 DEVELOPMENT.md 的 helpers.go 规范
//
// 职责：
//  1. 从 rawPayload 中根据 FieldPath 提取值
//  2. 验证字段类型
//  3. 返回提取后的载荷
func extractFields(rawPayload map[string]any, dataFormat core.DataFormat) (map[string]any, error) {
	payload := make(map[string]any)

	// 遍历所有字段配置
	for _, fieldCfg := range dataFormat.Fields {
		// 1. 从 rawPayload 中根据 FieldPath 提取值
		value, exists := extractFieldByPath(rawPayload, fieldCfg.FieldPath)
		if !exists {
			return nil, fmt.Errorf("MISSING_FIELD: 缺失必填字段: %s (路径: %s)",
				fieldCfg.FieldKey, fieldCfg.FieldPath)
		}

		// 2. 验证字段类型
		if err := validateFieldType(value, fieldCfg.FieldType); err != nil {
			return nil, fmt.Errorf("TYPE_MISMATCH: 字段类型错误: %s (期望 %s，实际 %T): %v",
				fieldCfg.FieldKey, fieldCfg.FieldType, value, err)
		}

		// 3. 保存到 payload
		payload[fieldCfg.FieldKey] = value
	}

	return payload, nil
}

// extractFieldByPath 从嵌套map中根据路径提取值
//
// 这是纯函数辅助方法（无 receiver），符合 DEVELOPMENT.md 的 helpers.go 规范
//
// 支持格式：
//   - "field" - 直接访问
//   - "parent.child" - 嵌套访问
func extractFieldByPath(data map[string]any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		value, exists := current[part]
		if !exists {
			return nil, false
		}

		// 最后一级，直接返回
		if i == len(parts)-1 {
			return value, true
		}

		// 中间级，继续向下查找
		nextMap, ok := value.(map[string]any)
		if !ok {
			return nil, false
		}
		current = nextMap
	}

	return nil, false
}

// validateFieldType 验证字段类型
//
// 这是纯函数辅助方法（无 receiver），符合 DEVELOPMENT.md 的 helpers.go 规范
//
// 注意：JSON 数字默认解析为 float64，需要特殊处理整数类型
func validateFieldType(value any, expectedType core.FieldType) error {
	switch expectedType {
	case core.FieldTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("期望字符串类型")
		}
	case core.FieldTypeInt:
		// JSON数字默认解析为float64，需要判断是否为整数
		if f, ok := value.(float64); !ok {
			return fmt.Errorf("期望整数类型，但不是数字")
		} else if f != float64(int64(f)) {
			return fmt.Errorf("期望整数类型，但包含小数部分: %v", f)
		}
	case core.FieldTypeFloat:
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("期望浮点数类型")
		}
	case core.FieldTypeBool:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("期望布尔类型")
		}
	default:
		return fmt.Errorf("未知的字段类型: %s", expectedType)
	}
	return nil
}

// parseTags 将 tags 列表转换为 labels 映射
//
// 这是纯函数辅助方法（无 receiver），符合 DEVELOPMENT.md 的 helpers.go 规范
//
// 支持两种格式：
//  1. "key:value" → labels["key"] = "value"
//  2. "simple_tag" → labels["simple_tag"] = "true"
func parseTags(tags []string) map[string]string {
	labels := make(map[string]string)
	for _, tag := range tags {
		if strings.Contains(tag, ":") {
			parts := strings.SplitN(tag, ":", 2)
			labels[parts[0]] = parts[1]
		} else {
			labels[tag] = "true"
		}
	}
	return labels
}
