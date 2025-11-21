// Package core 定义IoT设备管理的核心领域模型
//
// 文件说明：platform.go
// 职责：平台对接配置（Skylark、Webhook、Kafka等）
//
// Etcd Key格式: platform-{platformType}/{id}
// 例如: platform-skylark/550e8400-e29b-41d4-a716-446655440000
package core

import "fmt"

// PlatformType 平台类型
type PlatformType string

const (
	PlatformTypeSkylark PlatformType = "skylark" // Skylark 平台
	PlatformTypeWebhook PlatformType = "webhook" // Webhook（预留）
	PlatformTypeKafka   PlatformType = "kafka"   // Kafka（预留）
)

// PlatformMetadata 平台配置元数据（用于创建平台配置，不含ID）
type PlatformMetadata struct {
	Type        PlatformType `json:"type" title:"平台类型"`                       // 平台类型
	Name        string       `json:"name" title:"配置名称"`                       // 配置显示名称
	Description string       `json:"description,optional,omitempty" title:"描述"` // 配置描述
	TenantID    string       `json:"tenantID" title:"租户ID"`                    // 租户隔离
	Enabled     bool         `json:"enabled" title:"是否启用"`                    // 是否启用该配置

	// 审计字段
	CreatedBy string `json:"createdBy,optional,omitempty" title:"创建者"` // 创建者
	UpdatedBy string `json:"updatedBy,optional,omitempty" title:"更新者"` // 最后修改者
}

// PlatformConfig 平台配置（存储在Etcd，包含敏感信息）
// 多态配置，通过 Type 区分平台类型
type PlatformConfig struct {
	// 基本信息（与元数据一致，用于校验）
	Type     PlatformType `json:"type" title:"平台类型"`   // 平台类型（必须与元数据一致）
	TenantID string       `json:"tenantID" title:"租户ID"` // 租户隔离（必须与元数据一致）

	// Skylark 平台配置
	SkylarkDomain     string `json:"skylarkDomain,optional,omitempty" title:"Skylark域名"`      // Skylark 推送的域名
	SkylarkAuthHeader string `json:"skylarkAuthHeader,optional,omitempty" title:"Skylark认证头"` // Skylark 推送的认证头
	SkylarkUserID     int64  `json:"skylarkUserID,optional,omitempty" title:"Skylark用户ID"`    // 发起流程使用的用户ID

	// Webhook 平台配置（预留）
	WebhookURL     string            `json:"webhookUrl,optional,omitempty" title:"Webhook URL"`        // Webhook 推送地址
	WebhookHeaders map[string]string `json:"webhookHeaders,optional,omitempty" title:"Webhook请求头"`    // 自定义请求头
	WebhookMethod  string            `json:"webhookMethod,optional,omitempty" title:"Webhook HTTP方法"` // HTTP 方法（GET/POST等）

	// Kafka 平台配置（预留）
	KafkaBrokers []string `json:"kafkaBrokers,optional,omitempty" title:"Kafka Brokers"` // Kafka broker 地址列表
	KafkaTopic   string   `json:"kafkaTopic,optional,omitempty" title:"Kafka Topic"`     // Kafka 主题
}

// Platform 平台配置完整信息（用于查询返回）
type Platform struct {
	ID          string       `json:"id"`                         // 配置UUID（系统自动生成）
	Type        PlatformType `json:"type"`                       // 平台类型
	Name        string       `json:"name"`                       // 配置显示名称
	Description string       `json:"description,omitempty"`      // 配置描述
	Enabled     bool         `json:"enabled"`                    // 是否启用该配置
	TenantID    string       `json:"tenantId"`                   // 租户隔离
	CreatedBy   string       `json:"createdBy,omitempty"`        // 创建者
	UpdatedBy   string       `json:"updatedBy,omitempty"`        // 最后修改者
	CreatedAt   string       `json:"createdAt,omitempty"`        // 创建时间
	UpdatedAt   string       `json:"updatedAt,omitempty"`        // 更新时间
	Config      *PlatformConfig `json:"config,omitempty"`       // 平台配置（包含敏感信息，可选）
}

// Validate 验证平台配置的有效性
func (c *PlatformConfig) Validate() error {
	// 验证必填字段
	if c.Type == "" {
		return fmt.Errorf("平台类型不能为空")
	}
	if c.TenantID == "" {
		return fmt.Errorf("租户ID不能为空")
	}

	// 验证平台类型
	validTypes := map[PlatformType]bool{
		PlatformTypeSkylark: true,
		PlatformTypeWebhook: true,
		PlatformTypeKafka:   true,
	}
	if !validTypes[c.Type] {
		return fmt.Errorf("无效的平台类型: %s", c.Type)
	}

	// 验证平台特定配置，并确保其他平台字段为空（字段互斥检查）
	switch c.Type {
	case PlatformTypeSkylark:
		// 检查其他平台字段必须为空
		if err := c.ensureOtherPlatformFieldsEmpty(PlatformTypeSkylark); err != nil {
			return err
		}
		return c.validateSkylark()
	case PlatformTypeWebhook:
		// 检查其他平台字段必须为空
		if err := c.ensureOtherPlatformFieldsEmpty(PlatformTypeWebhook); err != nil {
			return err
		}
		return c.validateWebhook()
	case PlatformTypeKafka:
		// 检查其他平台字段必须为空
		if err := c.ensureOtherPlatformFieldsEmpty(PlatformTypeKafka); err != nil {
			return err
		}
		return c.validateKafka()
	default:
		return fmt.Errorf("未知的平台类型: %s", c.Type)
	}
}

// ensureOtherPlatformFieldsEmpty 确保非当前平台类型的字段为空
func (c *PlatformConfig) ensureOtherPlatformFieldsEmpty(currentType PlatformType) error {
	switch currentType {
	case PlatformTypeSkylark:
		// Webhook 字段必须为空
		if c.WebhookURL != "" || c.WebhookMethod != "" || len(c.WebhookHeaders) > 0 {
			return fmt.Errorf("平台类型为 skylark 时，不能设置 Webhook 相关字段")
		}
		// Kafka 字段必须为空
		if len(c.KafkaBrokers) > 0 || c.KafkaTopic != "" {
			return fmt.Errorf("平台类型为 skylark 时，不能设置 Kafka 相关字段")
		}
	case PlatformTypeWebhook:
		// Skylark 字段必须为空
		if c.SkylarkDomain != "" || c.SkylarkAuthHeader != "" || c.SkylarkUserID != 0 {
			return fmt.Errorf("平台类型为 webhook 时，不能设置 Skylark 相关字段")
		}
		// Kafka 字段必须为空
		if len(c.KafkaBrokers) > 0 || c.KafkaTopic != "" {
			return fmt.Errorf("平台类型为 webhook 时，不能设置 Kafka 相关字段")
		}
	case PlatformTypeKafka:
		// Skylark 字段必须为空
		if c.SkylarkDomain != "" || c.SkylarkAuthHeader != "" || c.SkylarkUserID != 0 {
			return fmt.Errorf("平台类型为 kafka 时，不能设置 Skylark 相关字段")
		}
		// Webhook 字段必须为空
		if c.WebhookURL != "" || c.WebhookMethod != "" || len(c.WebhookHeaders) > 0 {
			return fmt.Errorf("平台类型为 kafka 时，不能设置 Webhook 相关字段")
		}
	}
	return nil
}

// validateSkylark 验证 Skylark 平台配置
func (c *PlatformConfig) validateSkylark() error {
	if c.SkylarkDomain == "" {
		return fmt.Errorf("Skylark域名不能为空")
	}
	if c.SkylarkAuthHeader == "" {
		return fmt.Errorf("Skylark认证头不能为空")
	}
	if c.SkylarkUserID <= 0 {
		return fmt.Errorf("Skylark用户ID必须大于0")
	}
	return nil
}

// validateWebhook 验证 Webhook 平台配置（预留）
func (c *PlatformConfig) validateWebhook() error {
	// 暂不实现，预留扩展点
	return fmt.Errorf("Webhook平台暂未实现")
}

// validateKafka 验证 Kafka 平台配置（预留）
func (c *PlatformConfig) validateKafka() error {
	// 暂不实现，预留扩展点
	return fmt.Errorf("Kafka平台暂未实现")
}

// BuildPlatformKey 构建平台配置的 Etcd key
func BuildPlatformKey(platformType PlatformType, id string) string {
	return fmt.Sprintf("platform-%s/%s", platformType, id)
}
