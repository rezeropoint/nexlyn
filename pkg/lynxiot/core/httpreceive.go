// Package core 定义IoT设备管理的核心领域模型
//
// 文件说明：http_receive.go
// 职责：HTTP 数据接收配置模型
//
// Etcd Key格式: http-receive/{configId}
// 例如: http-receive/550e8400-e29b-41d4-a716-446655440000
package core

import (
	"fmt"
)

// HttpReceiveFieldMapping HTTP 接收字段映射规则
// 说明：
// - HTTP 接收处理任意格式的 JSON 数据，不需要映射到预定义的标准字段
// - 用户自定义字段名，从 JSON 路径提取数据
// - 不同于 MQTT 的 FieldMapping，不需要缩放因子和偏移量
type HttpReceiveFieldMapping struct {
	FieldName  string    `json:"fieldName"`  // 自定义字段名（用户自定义，如：temperature、deviceName）
	SourcePath string    `json:"sourcePath"` // JSON 字段路径，支持：
	//   - 点分隔：data.temp
	//   - 数组索引：Result.Tags[0]
	//   - 混合使用：data.items[2].name
	FieldType    FieldType `json:"fieldType"`              // 字段类型（string/number/boolean）
	DefaultValue string    `json:"defaultValue,omitempty"` // 默认值（可选，字段不存在时使用）
}

// HttpReceiveConfig HTTP 数据接收配置（存储在 Etcd）
//
// Etcd Key 格式: http-receive/{configId}
// 例如: http-receive/550e8400-e29b-41d4-a716-446655440000
//
// 说明：
// - ConfigID 是 UUID，由系统生成
// - 通过 configId 路径参数区分不同配置
// - 与 MQTT 的 DataProcessingConfig 独立，使用独立的字段映射类型
// - 这是纯配置对象，不包含审计信息（遵循 Core 层黄金法则）
// - HTTP 接收处理任意格式的 JSON 数据，不需要设备类别和标准字段的概念
type HttpReceiveConfig struct {
	// 基本信息
	ConfigID string `json:"configId"` // 配置ID（UUID，系统生成）

	// 租户信息
	TenantID string `json:"tenantId"` // 租户ID

	// 数据提取配置
	TimestampPath   string          `json:"timestampPath,omitempty"`   // 时间戳字段路径（点分隔，如：data.timestamp，可选）
	TimestampFormat TimestampFormat `json:"timestampFormat,omitempty"` // 时间戳格式（unix/unix_ms/unix_nano/iso8601/rfc3339）
	DeviceIDPath    string          `json:"deviceIdPath,omitempty"`    // 设备ID字段路径（从请求体提取，可选）

	// 字段映射配置（使用 HTTP Receive 专用的 FieldMapping）
	FieldMappings []HttpReceiveFieldMapping `json:"fieldMappings"` // 字段映射规则列表

	// 数据分发配置（复用现有类型）
	DispatchConfigs []DispatchConfig `json:"dispatchConfigs,omitempty"` // 数据分发配置列表
}

// Etcd配置前缀常量
const (
	HttpReceiveConfigPrefix = "http-receive/" // HTTP接收配置根前缀
)

// BuildHttpReceiveConfigKey 构建HTTP接收配置的Etcd key
// 格式: http-receive/{configId}
// 例如: http-receive/550e8400-e29b-41d4-a716-446655440000
func BuildHttpReceiveConfigKey(configId string) string {
	return fmt.Sprintf("%s%s", HttpReceiveConfigPrefix, configId)
}

// ValidateHttpReceiveFieldMappings 验证HTTP接收字段映射
// 说明：HTTP 接收处理任意格式的 JSON 数据，不需要验证标准字段
func ValidateHttpReceiveFieldMappings(mappings []HttpReceiveFieldMapping) error {
	if len(mappings) == 0 {
		return fmt.Errorf("字段映射列表不能为空")
	}

	// 验证每个映射
	for i, mapping := range mappings {
		// 验证字段名不为空
		if mapping.FieldName == "" {
			return fmt.Errorf("字段映射[%d]: 字段名不能为空", i)
		}

		// 验证源路径不为空
		if mapping.SourcePath == "" {
			return fmt.Errorf("字段映射[%d]: 源路径不能为空", i)
		}

		// 验证字段类型
		validTypes := map[FieldType]bool{
			FieldTypeString:  true,
			FieldTypeNumber:  true,
			FieldTypeBoolean: true,
		}
		if !validTypes[mapping.FieldType] {
			return fmt.Errorf("字段映射[%d]: 无效的字段类型 %s", i, mapping.FieldType)
		}
	}

	return nil
}

// Validate 验证HTTP接收配置的有效性
func (config *HttpReceiveConfig) Validate() error {
	// 验证必填字段
	if config.TenantID == "" {
		return fmt.Errorf("租户ID不能为空")
	}

	// 验证字段映射
	if len(config.FieldMappings) == 0 {
		return fmt.Errorf("字段映射列表不能为空")
	}

	// 验证字段映射
	if err := ValidateHttpReceiveFieldMappings(config.FieldMappings); err != nil {
		return fmt.Errorf("字段映射验证失败: %w", err)
	}

	// 验证时间戳格式
	if config.TimestampFormat != "" {
		validFormats := map[TimestampFormat]bool{
			TimestampFormatUnix:     true,
			TimestampFormatUnixMs:   true,
			TimestampFormatUnixNano: true,
			TimestampFormatISO8601:  true,
			TimestampFormatRFC3339:  true,
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

// HttpReceiveMetadata HTTP接收配置的元数据（用于创建/更新操作）
// 遵循 Template Manager 的设计模式，只包含创建者，不包含时间戳
type HttpReceiveMetadata struct {
	Name        string `json:"name"`                  // 配置名称
	Description string `json:"description,omitempty"` // 配置描述
	Enabled     bool   `json:"enabled"`               // 是否启用
	TenantID    string `json:"tenantId"`              // 租户ID
	CreatedBy   string `json:"createdBy"`             // 创建者（仅创建时使用）
}

// HttpReceive 完整的HTTP接收配置信息（元数据 + 配置 + 审计信息）
// 用于 Get 方法返回
type HttpReceive struct {
	ID          string `json:"id"`                    // 配置ID（UUID）
	Name        string `json:"name"`                  // 配置名称
	Description string `json:"description,omitempty"` // 配置描述
	Enabled     bool   `json:"enabled"`               // 是否启用
	TenantID    string `json:"tenantId"`              // 租户ID
	CreatedBy   string `json:"createdBy,omitempty"`   // 创建者
	CreatedAt   string `json:"createdAt,omitempty"`   // 创建时间
	UpdatedAt   string `json:"updatedAt,omitempty"`   // 更新时间

	// 配置信息（从 Etcd 查询）
	Config *HttpReceiveConfig `json:"config,omitempty"` // HTTP接收配置
}

// HttpReceiveSummary HTTP接收配置的摘要信息（用于列表展示）
type HttpReceiveSummary struct {
	ID          string `json:"id"`                    // 配置ID（UUID）
	Name        string `json:"name"`                  // 配置名称
	Description string `json:"description,omitempty"` // 配置描述
	Enabled     bool   `json:"enabled"`               // 是否启用
	CreatedAt   string `json:"createdAt,omitempty"`   // 创建时间
	UpdatedAt   string `json:"updatedAt,omitempty"`   // 更新时间
}

// HttpReceiveQuery HTTP接收配置的查询条件
type HttpReceiveQuery struct {
	TenantID string // 租户ID（必填）
	Keyword  string // 关键字搜索（可选，匹配名称和描述）
	Page     int    // 页码（从1开始）
	PageSize int    // 每页数量
}
