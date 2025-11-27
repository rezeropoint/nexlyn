package core

import (
	"context"
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
)

type CreateBlockFunc func(id string, blockKey BlockKey, config map[string]any) (LogicBlock, error)

// GetSpecFunc 获取逻辑块规格的函数类型
type GetSpecFunc func(blockKey BlockKey) (BlockSpec, error)

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

		rawVal, ok := cfg[configKey]
		if !ok || reflect.ValueOf(rawVal).IsZero() {
			if checkTag == "must" {
				return fmt.Errorf("%w: %s", ErrFieldMissingOrZero, configKey)
			}
			continue
		}

		fieldVal := v.Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		val := reflect.ValueOf(rawVal)
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
