/**
 * HTTP 接收配置相关类型
 *
 * 说明：HTTP Receive 处理任意格式的 JSON 数据，不需要映射到标准字段
 * 用户可以自定义字段名，从 JSON 路径提取数据
 */

// HTTP 接收字段映射（使用自定义字段名）
export interface HttpReceiveFieldMapping {
  fieldName: string; // 自定义字段名（如：temperature、deviceName）
  sourcePath: string; // JSON 字段路径，支持：
  //   - 点分隔：data.temp
  //   - 数组索引：Result.Tags[0]
  //   - 混合使用：data.items[2].name
  fieldType?: string; // 字段类型（string/imageURL/imageBase64）
  defaultValue?: string; // 默认值（可选，字段不存在时使用）
}

// HTTP 接收配置元数据（列表展示）
export interface HttpReceiveMetadata {
  id: string; // 配置ID
  name: string; // 配置名称
  description?: string; // 配置描述
  enabled: boolean; // 是否启用
  tenantId: string; // 租户ID
  createdBy?: string; // 创建者
  createdAt?: string; // 创建时间
  updatedAt?: string; // 更新时间
}

// HTTP 接收分发配置
export interface HttpReceiveDispatchConfig {
  type: string; // 分发类型：skylark_flows/skylark_forms/lynxgraph/log
  platformId?: string; // 平台配置ID
  flowId?: number; // Skylark 流程ID
  formId?: number; // Skylark 表单ID
  infoAtomTypeId?: string; // LynxGraph 信息原子类型ID
  extraParams?: Record<string, any>; // 扩展参数
}

// HTTP 接收配置完整详情
export interface HttpReceiveDetail extends HttpReceiveMetadata {
  timestampPath?: string; // 时间戳字段路径
  timestampFormat?: string; // 时间戳格式
  deviceIdPath?: string; // 设备ID字段路径
  fieldMappings: HttpReceiveFieldMapping[]; // 字段映射配置
  dispatchConfigs?: HttpReceiveDispatchConfig[]; // 分发配置
}
