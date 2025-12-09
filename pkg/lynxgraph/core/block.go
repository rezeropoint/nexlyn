package core

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// LogicBlockType 表示逻辑块的类型
type LogicBlockType string

// BlockFactory 逻辑块工厂函数类型
type BlockFactory func(id string, config map[string]any) (LogicBlock, error)

type BlockKey struct {
	BlockType LogicBlockType
	Version   string // 允许为空，表示默认版本
}

// LogicBlock 表示可执行的逻辑块接口
type LogicBlock interface {
	GetID() string
	GetType() LogicBlockType
	GetConfigure() map[string]any
	SetConfigure(config map[string]any) error
	// Execute 执行逻辑块
	// execCtx: 执行上下文，包含InfoAtom、Graph、Node、TenantId等信息
	// datastore: 数据存储接口，用于读写GraphContext和InfoAtom
	// service: 外部服务接口，用于调用外部系统（如Skylark）
	// 返回值: (success bool, err error) - success=true继续后续节点，false中断流程
	Execute(ctx context.Context, execCtx ExecutionContext, datastore Store, service Service) (bool, error)
}

// BaseLogicBlock 提供 LogicBlock 接口的基本实现
type BaseLogicBlock struct {
	ID        string         `json:"id"`
	Type      LogicBlockType `json:"type"`
	RawConfig map[string]any `json:"config"`

	TypedConfig any `json:"-"`
}

// 预定义的逻辑块类型
const (
	BlockTypeFilter       LogicBlockType = "filter"        // 过滤器
	BlockTypeStateMachine LogicBlockType = "state_machine" // 状态机
	BlockTypeStateUpdate  LogicBlockType = "state_update"  // 状态更新
	BlockTypeStateCheck   LogicBlockType = "state_check"   // 状态检查
	BlockTypeAction       LogicBlockType = "action"        // 动作执行
	BlockTypeExpression   LogicBlockType = "expression"    // 表达式计算
	BlockTypeAggregator   LogicBlockType = "aggregator"    // 事件聚合器
	BlockTypeTimer        LogicBlockType = "timer"         // 定时器/延迟
)

const (
	BlockTypeSkylarkJourneyCreate LogicBlockType = "skylark_journey_create" // 创建skylark流程记录
	BlockTypeLog                  LogicBlockType = "log"                    // 日志打印
	BlockTypeRuleMatcher          LogicBlockType = "rule_matcher"           // 规则匹配器
	BlockTypeFetchSensorData      LogicBlockType = "fetch_sensor_data"      // 拉取传感器数据（替代旧的 query_history_data）
	BlockTypeStatisticalAnalyzer  LogicBlockType = "statistical_analyzer"   // 统计分析（均值、标准差、最大值等）
	BlockTypeRateOfChange         LogicBlockType = "rate_of_change"         // 变化率计算（绝对/相对变化）
	BlockTypeAnomalyDetector      LogicBlockType = "anomaly_detector"       // 异常检测（Z-score、IQR方法）
	BlockTypeTrendAnalyzer        LogicBlockType = "trend_analyzer"         // 趋势分析（线性回归）
	BlockTypeHttpRequest          LogicBlockType = "http_request"           // HTTP 请求
	BlockTypeTimeWindowCheck      LogicBlockType = "time_window_check"      // 时间窗口检查
	BlockTypeSwitch               LogicBlockType = "switch"                 // 条件路由
	BlockTypeHolidayCheck         LogicBlockType = "holiday_check"          // 假期检查
	BlockTypeQueryDatabase        LogicBlockType = "query_database"         // 外部数据库查询
	BlockTypeDedupCheck           LogicBlockType = "dedup_check"            // 去重检查
)

type CreateBlockFunc func(id string, blockKey BlockKey, config map[string]any) (LogicBlock, error)

// GetSpecFunc 获取逻辑块规格的函数类型
type GetSpecFunc func(blockKey BlockKey) (BlockSpec, error)

// toStringAnyMap 将各种 map 类型转换为 map[string]any
// 支持 map[string]any、map[string]interface{}、bson.M 等
func toStringAnyMap(v any) map[string]any {
	if v == nil {
		return nil
	}
	// 直接类型断言
	if m, ok := v.(map[string]any); ok {
		return m
	}
	// 通过反射处理 bson.M 等类型
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
		result := make(map[string]any, rv.Len())
		for _, key := range rv.MapKeys() {
			result[key.String()] = rv.MapIndex(key).Interface()
		}
		return result
	}
	return nil
}

// FillConfig 填充配置
func FillConfig(cfg map[string]any, tmpl any) error {
	v := reflect.ValueOf(tmpl)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return ErrTmplNotStruct
	}
	v = v.Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		// 只使用 json tag 作为配置 key，没有 json tag 的字段跳过
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		// 处理 json tag 中的选项（如 "name,omitempty"）
		configKey := jsonTag
		if idx := strings.Index(jsonTag, ","); idx != -1 {
			configKey = jsonTag[:idx]
		}
		checkTag := field.Tag.Get("check")

		// 嵌套结构体支持
		if field.Type.Kind() == reflect.Struct {
			subVal := v.Field(i)
			subRaw, ok := cfg[configKey]
			if !ok {
				if checkTag == "must" {
					return fmt.Errorf("%w: %s", ErrNestedFieldMissing, configKey)
				}
				continue
			}
			subMap, ok := subRaw.(map[string]any)
			if !ok {
				return fmt.Errorf("%w: %s", ErrNestedFieldType, configKey)
			}
			if err := FillConfig(subMap, subVal.Addr().Interface()); err != nil {
				return fmt.Errorf("%w: %s: %w", ErrNestedFieldCheckFailed, configKey, err)
			}
			continue
		}

		// 切片类型支持（如 []TimeWindow）
		if field.Type.Kind() == reflect.Slice {
			subRaw, ok := cfg[configKey]
			if !ok {
				if checkTag == "must" {
					return fmt.Errorf("%w: %s", ErrNestedFieldMissing, configKey)
				}
				continue
			}

			// 将各种切片类型（[]any, bson.A 等）统一转换为 []any
			var rawSlice []any
			subVal := reflect.ValueOf(subRaw)
			if subVal.Kind() == reflect.Slice {
				rawSlice = make([]any, subVal.Len())
				for i := 0; i < subVal.Len(); i++ {
					rawSlice[i] = subVal.Index(i).Interface()
				}
			} else {
				return fmt.Errorf("%w: %s: want slice, got %T", ErrNestedFieldType, configKey, subRaw)
			}

			elemType := field.Type.Elem()
			newSlice := reflect.MakeSlice(field.Type, len(rawSlice), len(rawSlice))

			for j, item := range rawSlice {
				elemVal := newSlice.Index(j)
				// 如果元素是结构体，递归填充
				if elemType.Kind() == reflect.Struct {
					itemMap := toStringAnyMap(item)
					if itemMap == nil {
						return fmt.Errorf("%w: %s[%d]: want map, got %T", ErrNestedFieldType, configKey, j, item)
					}
					if err := FillConfig(itemMap, elemVal.Addr().Interface()); err != nil {
						return fmt.Errorf("%w: %s[%d]: %w", ErrNestedFieldCheckFailed, configKey, j, err)
					}
				} else {
					// 基础类型直接赋值
					itemVal := reflect.ValueOf(item)
					if itemVal.Type().AssignableTo(elemType) {
						elemVal.Set(itemVal)
					} else if itemVal.Type().ConvertibleTo(elemType) {
						elemVal.Set(itemVal.Convert(elemType))
					} else {
						return fmt.Errorf("%w: %s[%d]: want %s, got %s", ErrFieldTypeMismatch, configKey, j, elemType, itemVal.Type())
					}
				}
			}
			v.Field(i).Set(newSlice)
			continue
		}

		rawVal, ok := cfg[configKey]
		if !ok || rawVal == nil {
			if checkTag == "must" {
				return fmt.Errorf("%w: %s", ErrFieldMissingOrZero, configKey)
			}
			continue
		}
		// 仅对 check:"must" 字段检查零值，其他字段允许零值（由调用方设置默认值）
		if checkTag == "must" {
			rv := reflect.ValueOf(rawVal)
			if rv.IsValid() && rv.IsZero() {
				return fmt.Errorf("%w: %s", ErrFieldMissingOrZero, configKey)
			}
		}

		fieldVal := v.Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		// 处理 json.Number 类型（go-zero 使用 UseNumber 解码 JSON）
		if num, ok := rawVal.(json.Number); ok {
			switch fieldVal.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if n, err := num.Int64(); err == nil {
					fieldVal.SetInt(n)
					continue
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if n, err := num.Int64(); err == nil && n >= 0 {
					fieldVal.SetUint(uint64(n))
					continue
				}
			case reflect.Float32, reflect.Float64:
				if f, err := num.Float64(); err == nil {
					fieldVal.SetFloat(f)
					continue
				}
			}
		}

		val := reflect.ValueOf(rawVal)

		// 处理指针类型字段（如 *bool、*int）
		if fieldVal.Kind() == reflect.Ptr {
			elemType := fieldVal.Type().Elem()
			if val.Type().AssignableTo(elemType) {
				ptr := reflect.New(elemType)
				ptr.Elem().Set(val)
				fieldVal.Set(ptr)
				continue
			} else if val.Type().ConvertibleTo(elemType) {
				ptr := reflect.New(elemType)
				ptr.Elem().Set(val.Convert(elemType))
				fieldVal.Set(ptr)
				continue
			}
		}

		if val.Type().AssignableTo(fieldVal.Type()) {
			fieldVal.Set(val)
		} else if val.Type().ConvertibleTo(fieldVal.Type()) {
			fieldVal.Set(val.Convert(fieldVal.Type()))
		} else {
			return fmt.Errorf("%w: %s: want %s, got %s", ErrFieldTypeMismatch, configKey, field.Type, val.Type())
		}
	}
	return nil
}
