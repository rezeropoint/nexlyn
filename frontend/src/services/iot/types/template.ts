/**
 * 设备模板相关类型
 */

// 字段映射规则
export interface FieldMapping {
  standardField: string; // 平台标准字段名
  sourcePath: string; // 源字段路径（点分隔）
  fieldType?: string; // 字段类型
  scale?: number; // 缩放因子
  offset?: number; // 偏移量
  defaultValue?: string; // 默认值
}

// 字段检查规则（用于在线检测）
export interface FieldCheck {
  field: string; // 字段路径
  operator?: string; // 操作符：exists/equals/in
  value?: string; // 期望值（operator为equals时使用）
  values?: string[]; // 期望值列表（operator为in时使用）
}

// 在线检测配置（独立配置）
export interface OnlineDetectionConfig {
  category: string; // 设备类别
  model: string; // 设备型号
  topicSuffix: string; // MQTT主题后缀
  strategy: string; // 检测策略：field_check/any_data
  timeoutSeconds: number; // 超时时间（秒）
  fieldChecks?: FieldCheck[]; // 字段检查规则（strategy为field_check时使用）
}

// 分发配置
export interface DispatchConfig {
  type: string; // 分发类型：skylark_flows/skylark_forms/lynxgraph/log
  platformID?: string; // 平台配置ID（关联平台配置，Skylark使用）
  flowID?: number; // Skylark流程ID（type=skylark_flows时必填）
  formID?: number; // Skylark表单ID（type=skylark_forms时必填）
  infoAtomTypeID?: string; // LynxGraph信息原子类型ID（type=lynxgraph时必填）
  extraParams?: Record<string, any>; // 扩展参数（其他平台使用）
}

// 过滤条件
export interface Condition {
  field: string; // 字段名（对应StandardField）
  operator: string; // 操作符：eq/ne/gt/lt/gte/lte/contains/in/not_in
  value: any; // 比较值
}

// 过滤规则（简化版，无嵌套）
export interface FilterRule {
  conditions: Condition[]; // 过滤条件列表
  logic?: string; // 逻辑关系：AND/OR，默认AND
}

// 业务数据处理配置（独立配置）
export interface DataProcessingConfig {
  category: string; // 设备类别
  model: string; // 设备型号
  topicSuffix: string; // MQTT主题后缀
  timestampPath?: string; // 时间戳路径
  timestampFormat?: string; // 时间戳格式：unix/unix_ms/iso8601/rfc3339
  fieldMappings: FieldMapping[]; // 字段映射规则
  filterRules?: FilterRule | null; // 数据过滤规则（用于分发前筛选）
  dispatchConfigs?: DispatchConfig[]; // 分发配置
}

// 设备模板实体（PostgreSQL中的元数据）
export interface SensorTemplate {
  id: string; // 模板UUID
  model: string; // 设备型号标识（业务标识）
  name: string; // 模板显示名称
  category: string; // 设备类别
  manufacturer: string; // 设备厂商
  description: string; // 模板描述
  version: string; // 模板版本号
  enabled: boolean; // 是否启用
  deviceCount: number; // 使用该模板的设备数量
  tenantId: string; // 租户ID
  createdBy: string; // 创建者用户ID
  updatedBy: string; // 最后修改者用户ID
  createdAt: string; // 创建时间（ISO 8601）
  updatedAt: string; // 更新时间（ISO 8601）
}

// 设备控制配置（独立配置）
export interface DeviceControlConfig {
  category: string; // 设备类别
  model: string; // 设备型号
  commandSuffix: string; // 命令主题后缀
  responseSuffix: string; // 响应主题后缀
  tenantId?: string; // 租户ID（可选）
}

// 设备模板完整信息（包含Etcd配置）
export interface SensorTemplateDetail {
  id: string;
  model: string;
  name: string;
  category: string;
  manufacturer: string;
  description: string;
  version: string;
  enabled: boolean;
  deviceCount: number;
  tenantId: string;
  createdBy: string;
  updatedBy: string;
  createdAt: string;
  updatedAt: string;
  onlineConfig?: OnlineDetectionConfig; // 在线检测配置（可选）
  businessConfig?: DataProcessingConfig; // 业务数据处理配置（可选）
  controlConfig?: DeviceControlConfig; // 设备控制配置（可选）
}
