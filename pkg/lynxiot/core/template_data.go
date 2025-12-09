// Package core 定义IoT设备管理的核心领域模型
//
// 文件说明：template_data.go
// 职责：业务数据处理配置（字段映射、数据验证、时间戳提取）
//
// Etcd Key格式: sensor-data/{category}/{model}/{topicSuffix}
// 例如: sensor-data/ai_box/AI-200/data
package core

import (
	"fmt"
)

// TimestampFormat 时间戳格式类型
type TimestampFormat string

const (
	TimestampFormatUnix      TimestampFormat = "unix"       // Unix时间戳（秒）
	TimestampFormatUnixMs    TimestampFormat = "unix_ms"    // Unix时间戳（毫秒）
	TimestampFormatUnixMicro TimestampFormat = "unix_micro" // Unix时间戳（微秒）
	TimestampFormatISO8601   TimestampFormat = "iso8601"    // ISO 8601格式
	TimestampFormatRFC3339   TimestampFormat = "rfc3339"    // RFC 3339格式
)

// FieldMapping 字段映射规则
type FieldMapping struct {
	StandardField FieldName `json:"standardField" title:"标准字段名"`                 // 平台预定义的标准字段名（如：temperature, humidity）
	SourcePath    string    `json:"sourcePath" title:"源字段路径"`                    // 原始数据中的字段路径（点分隔，如：data.temp）
	FieldType     FieldType `json:"fieldType,optional,omitempty" title:"字段类型"`   // 字段类型（用于类型转换，如时间戳转日期）
	Scale         float64   `json:"scale,optional,omitempty" title:"缩放因子"`       // 缩放因子（value * scale，如：0.1）
	Offset        float64   `json:"offset,optional,omitempty" title:"偏移量"`       // 偏移量（value + offset，如：-273.15）
	DefaultValue  string    `json:"defaultValue,optional,omitempty" title:"默认值"` // 字段缺失时的默认值
}

// FilterRule 数据过滤规则（用于分发前筛选，简化版无嵌套）
type FilterRule struct {
	Conditions []Condition `json:"conditions" title:"条件列表"`                     // 过滤条件列表
	Logic      string      `json:"logic,optional,omitempty" title:"逻辑关系"` // AND/OR，默认AND
}

// Condition 单个过滤条件
type Condition struct {
	Field    string `json:"field" title:"字段名"`        // 字段名（对应StandardField）
	Operator string `json:"operator" title:"操作符"`     // eq/ne/gt/lt/gte/lte/contains/in/not_in
	Value    any    `json:"value" title:"比较值"`        // 比较值
}

// 操作符说明：
// - eq: 等于
// - ne: 不等于
// - gt: 大于
// - lt: 小于
// - gte: 大于等于
// - lte: 小于等于
// - contains: 包含（字符串）
// - in: 在列表中
// - not_in: 不在列表中

// DataProcessingConfig 业务数据处理配置（存储在Etcd中）
//
// Etcd Key格式: sensor-data/{category}/{model}/{topicSuffix}
// 例如: sensor-data/ai_box/AI-200/data
//
// 说明：
// - 不需要 id、template_id 字段（Etcd key 本身已唯一标识）
// - category/model/topicSuffix 组合即为 Etcd key，实现 O(1) 查询
// - 一个主题对应一个数据处理配置（通过 category/model/topicSuffix 唯一标识）
// - 配置变更可触发数据处理流程的自动更新
// - TenantID 用于数据存储时的租户隔离
type DataProcessingConfig struct {
	Category    string `json:"category"`    // 设备类别
	Model       string `json:"model"`       // 设备型号（关联 iot_sensor_templates.model）
	TopicSuffix string `json:"topicSuffix"` // 主题后缀
	TenantID    string `json:"tenantId"`    // 租户ID（用于数据存储时的租户隔离）

	// 数据提取配置
	TimestampPath   string          `json:"timestampPath,omitempty"`   // 时间戳字段路径（点分隔，如：data.timestamp，可选）
	TimestampFormat TimestampFormat `json:"timestampFormat,omitempty"` // 时间戳格式（unix/unix_ms/unix_nano/iso8601/rfc3339）

	// 字段映射配置
	FieldMappings []FieldMapping `json:"fieldMappings"` // 字段映射规则列表

	// 数据过滤配置（用于分发前筛选）
	FilterRules *FilterRule `json:"filterRules,optional,omitempty" title:"过滤规则"` // 数据分发过滤规则

	// 数据分发配置
	DispatchConfigs []DispatchConfig `json:"dispatchConfigs,omitempty"` // 数据分发配置列表
}

// Etcd配置前缀常量
const (
	DataProcessingConfigPrefix = "sensor-data/" // 业务数据配置根前缀
)

// BuildDataConfigPrefix 构建业务数据配置的型号级前缀
// 格式: sensor-data/{category}/{model}/
// 用于: GetAllKeys 前缀查询、AddPrefixWatcher 监听
func BuildDataConfigPrefix(category, model string) string {
	return fmt.Sprintf("%s%s/%s/", DataProcessingConfigPrefix, category, model)
}

// BuildDataProcessingKey 构建业务数据处理配置的Etcd key
// 格式: sensor-data/{category}/{model}/{topicSuffix}
// 例如: sensor-data/ai_box/AI-200/data
func BuildDataProcessingKey(category, model, topicSuffix string) string {
	return fmt.Sprintf("%s%s/%s/%s", DataProcessingConfigPrefix, category, model, topicSuffix)
}

// Validate 验证业务数据处理配置的有效性
func (config *DataProcessingConfig) Validate() error {
	// 验证必填字段
	if config.Category == "" {
		return fmt.Errorf("设备类别不能为空")
	}
	if config.Model == "" {
		return fmt.Errorf("设备型号不能为空")
	}
	if config.TopicSuffix == "" {
		return fmt.Errorf("主题后缀不能为空")
	}
	if config.TenantID == "" {
		return fmt.Errorf("租户ID不能为空")
	}

	// 验证字段映射
	if len(config.FieldMappings) == 0 {
		return fmt.Errorf("字段映射列表不能为空")
	}

	// 验证字段映射是否符合设备类别的标准字段
	if err := ValidateFieldMappings(DeviceCategory(config.Category), config.FieldMappings); err != nil {
		return fmt.Errorf("字段映射验证失败: %w", err)
	}

	// 验证时间戳格式
	if config.TimestampFormat != "" {
		validFormats := map[TimestampFormat]bool{
			TimestampFormatUnix:      true,
			TimestampFormatUnixMs:    true,
			TimestampFormatUnixMicro: true,
			TimestampFormatISO8601:   true,
			TimestampFormatRFC3339:   true,
		}
		if !validFormats[config.TimestampFormat] {
			return fmt.Errorf("无效的时间戳格式: %s", config.TimestampFormat)
		}
	}

	// 验证分发配置
	for i, dispatchConfig := range config.DispatchConfigs {
		if err := dispatchConfig.Validate(); err != nil {
			return fmt.Errorf("分发配置[%d]验证失败: %w", i, err)
		}
	}

	return nil
}

// ValidateFieldMappings 验证字段映射的有效性
// 检查映射的标准字段是否属于指定的设备类别
func ValidateFieldMappings(category DeviceCategory, mappings []FieldMapping) error {
	if len(mappings) == 0 {
		return fmt.Errorf("字段映射列表不能为空")
	}

	// 获取该类别的所有标准字段
	standardFields := GetStandardFieldsByCategory(category)
	if len(standardFields) == 0 {
		return fmt.Errorf("设备类别 %s 没有定义标准字段", category)
	}

	// 构建标准字段名集合
	validFieldNames := make(map[FieldName]bool)
	requiredFields := make(map[FieldName]bool)
	for _, field := range standardFields {
		validFieldNames[FieldName(field.Name)] = true
		if field.Required {
			requiredFields[FieldName(field.Name)] = true
		}
	}

	// 验证每个映射
	mappedFields := make(map[FieldName]bool)
	for i, mapping := range mappings {
		// 1. 检查标准字段是否存在
		if !validFieldNames[mapping.StandardField] {
			return fmt.Errorf("字段映射[%d]: 标准字段 '%s' 不属于设备类别 '%s'",
				i, mapping.StandardField, category)
		}

		// 2. 检查源字段路径是否为空
		if mapping.SourcePath == "" {
			return fmt.Errorf("字段映射[%d]: 标准字段 '%s' 的源字段路径不能为空",
				i, mapping.StandardField)
		}

		// 3. 记录已映射的字段
		mappedFields[mapping.StandardField] = true
	}

	// 4. 检查必填字段是否都已映射
	for fieldName := range requiredFields {
		if !mappedFields[fieldName] {
			return fmt.Errorf("必填字段 '%s' 缺少映射配置", fieldName)
		}
	}

	return nil
}
