/**
 * 平台配置相关类型
 */

// 平台配置元数据（PostgreSQL存储）
export interface PlatformMetadata {
  id: string; // 配置ID（业务标识）
  type: string; // 平台类型：skylark/webhook/kafka
  name: string; // 配置显示名称
  description?: string; // 配置描述
  enabled: boolean; // 是否启用
  tenantId: string; // 租户ID
  createdBy?: string; // 创建者
  updatedBy?: string; // 更新者
  createdAt?: string; // 创建时间
  updatedAt?: string; // 更新时间
}

// Skylark平台配置
export interface SkylarkConfig {
  skylarkDomain: string; // Skylark域名
  skylarkAuthHeader: string; // Skylark认证头
  skylarkUserID: number; // Skylark用户ID
}

// Webhook平台配置
export interface WebhookConfig {
  webhookUrl: string; // Webhook URL
  webhookHeaders?: Record<string, string>; // 自定义请求头
  webhookMethod?: string; // HTTP方法
}

// Kafka平台配置
export interface KafkaConfig {
  kafkaBrokers: string[]; // Kafka Brokers
  kafkaTopic: string; // Kafka Topic
}

// 平台配置完整信息
export interface PlatformDetail extends PlatformMetadata {
  config?: SkylarkConfig | WebhookConfig | KafkaConfig; // 平台特定配置（根据type解析）
}
