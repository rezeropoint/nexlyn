package mqtt

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
)

// applyFilter 对数据应用过滤规则（简化版）
func applyFilter(ctx context.Context, data *map[string]core.TypedValue, rule core.FilterRule) (bool, error) {
	// 如果没有条件，默认通过
	if len(rule.Conditions) == 0 {
		return true, nil
	}

	// 默认逻辑为 AND
	logic := strings.ToUpper(rule.Logic)
	if logic == "" {
		logic = "AND"
	}

	// 评估所有条件
	conditionResults := make([]bool, 0, len(rule.Conditions))
	for _, condition := range rule.Conditions {
		result, err := evaluateCondition(*data, condition)
		if err != nil {
			return false, err
		}
		conditionResults = append(conditionResults, result)
	}

	// 根据逻辑关系合并结果
	if logic == "OR" {
		// OR 逻辑：任一条件为 true 则通过
		for _, result := range conditionResults {
			if result {
				return true, nil
			}
		}
		return false, nil
	} else {
		// AND 逻辑：所有条件都为 true 才通过
		for _, result := range conditionResults {
			if !result {
				return false, nil
			}
		}
		return true, nil
	}
}

// evaluateCondition 评估单个条件
func evaluateCondition(data map[string]core.TypedValue, condition core.Condition) (bool, error) {
	// 获取字段值
	fieldValue, exists := data[condition.Field]
	if !exists {
		// 字段不存在，根据操作符决定结果
		if condition.Operator == "ne" || condition.Operator == "not_in" {
			return true, nil
		}
		return false, nil
	}

	// 根据操作符比较
	switch strings.ToLower(condition.Operator) {
	case "eq", "=", "==":
		return compareEqual(fieldValue.Value, condition.Value), nil

	case "ne", "!=", "<>":
		return !compareEqual(fieldValue.Value, condition.Value), nil

	case "gt", ">":
		return compareGreater(fieldValue.Value, condition.Value)

	case "gte", ">=":
		cmp, err := compareValues(fieldValue.Value, condition.Value)
		if err != nil {
			return false, err
		}
		return cmp >= 0, nil

	case "lt", "<":
		return compareLess(fieldValue.Value, condition.Value)

	case "lte", "<=":
		cmp, err := compareValues(fieldValue.Value, condition.Value)
		if err != nil {
			return false, err
		}
		return cmp <= 0, nil

	case "contains":
		return contains(fieldValue.Value, condition.Value), nil

	case "in":
		return inList(fieldValue.Value, condition.Value), nil

	case "not_in":
		return !inList(fieldValue.Value, condition.Value), nil

	default:
		return false, fmt.Errorf("不支持的操作符: %s", condition.Operator)
	}
}

// compareEqual 比较相等
func compareEqual(a, b any) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// compareGreater 比较大于
func compareGreater(a, b any) (bool, error) {
	cmp, err := compareValues(a, b)
	if err != nil {
		return false, err
	}
	return cmp > 0, nil
}

// compareLess 比较小于
func compareLess(a, b any) (bool, error) {
	cmp, err := compareValues(a, b)
	if err != nil {
		return false, err
	}
	return cmp < 0, nil
}

// compareValues 比较两个值，返回 -1, 0, 1
func compareValues(a, b any) (int, error) {
	// 尝试转换为数字比较
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)

	if aOk && bOk {
		if aFloat < bFloat {
			return -1, nil
		} else if aFloat > bFloat {
			return 1, nil
		}
		return 0, nil
	}

	// 字符串比较
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	if aStr < bStr {
		return -1, nil
	} else if aStr > bStr {
		return 1, nil
	}
	return 0, nil
}

// toFloat64 尝试转换为 float64
func toFloat64(v any) (float64, bool) {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(val.Uint()), true
	case reflect.Float32, reflect.Float64:
		return val.Float(), true
	}
	return 0, false
}

// contains 检查是否包含
func contains(haystack, needle any) bool {
	haystackStr := fmt.Sprintf("%v", haystack)
	needleStr := fmt.Sprintf("%v", needle)
	return strings.Contains(haystackStr, needleStr)
}

// inList 检查是否在列表中
func inList(value any, list any) bool {
	listVal := reflect.ValueOf(list)
	if listVal.Kind() != reflect.Slice && listVal.Kind() != reflect.Array {
		// 不是列表，直接比较
		return compareEqual(value, list)
	}

	valueStr := fmt.Sprintf("%v", value)
	for i := 0; i < listVal.Len(); i++ {
		item := listVal.Index(i).Interface()
		if fmt.Sprintf("%v", item) == valueStr {
			return true
		}
	}
	return false
}
