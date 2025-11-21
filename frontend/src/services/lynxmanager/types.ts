/**
 * LynxGraph Manager API 类型定义
 * @description 信息原子类型、逻辑图配置等相关类型
 */

// ========== 通用类型 ==========

/** 基础响应 */
export interface BaseResponse {
  code: number;
  message: string;
  msg?: string; // 兼容不同的错误消息字段
}

/** 分页参数 */
export interface PageParams {
  page: number;
  pageSize: number;
  total: number;
}

/** 分页请求参数 */
export interface PageParamsRequest {
  page?: number;
  pageSize?: number;
}

// ========== 信息原子类型相关 ==========

/** 字段类型枚举 */
export enum FieldType {
  String = 'string',
  Int = 'int',
  Float = 'float',
  Bool = 'bool',
}

/** 字段配置 */
export interface FieldConfig {
  fieldKey: string; // 字段键
  fieldPath: string; // 字段路径（支持JSONPath）
  fieldType: FieldType | string; // 字段类型
}

/** 数据格式定义 */
export interface DataFormat {
  dataPlural: boolean; // 数据是否为数组
  fieldStart?: string; // 数组数据的起始字段路径
  fields: FieldConfig[]; // 字段列表
}

/** 信息原子类型 */
export interface InfoAtomType {
  id: string; // 信息原子类型ID（UUID）
  tenantId: string; // 租户ID
  name: string; // 类型名称（如 sensor.temperature）
  version: string; // 版本号（如 v1）
  tags?: LynxTagSummary[]; // 标签列表（Phase 2.5: 改为完整对象）
  dataFormat: DataFormat; // 数据格式定义
  createdAt?: number; // 创建时间（Unix毫秒）
  updatedAt?: number; // 更新时间（Unix毫秒）
  createdBy?: string; // 创建人ID
  updatedBy?: string; // 更新人ID
}

// ========== 信息原子类型 API 请求/响应 ==========

/** 创建信息原子类型请求 */
export interface CreateInfoAtomTypeRequest {
  tenantId: string; // 租户ID
  name: string; // 类型名称
  version: string; // 版本号
  tagIds?: string[]; // 标签ID列表
  dataFormat: DataFormat; // 数据格式定义
}

/** 创建信息原子类型响应 */
export interface CreateInfoAtomTypeResponse extends BaseResponse {
  data?: {
    id: string; // 创建的信息原子类型ID
  };
}

/** 查询信息原子类型列表请求 */
export interface GetInfoAtomTypeListRequest extends PageParamsRequest {
  tenantId: string; // 租户ID（必填）
  name?: string; // 类型名称筛选（模糊匹配）
  tags?: string[]; // 标签筛选
}

/** 查询信息原子类型列表响应 */
export interface GetInfoAtomTypeListResponse extends BaseResponse, PageParams {
  data?: {
    list: InfoAtomType[];
  };
}

/** 获取信息原子类型详情请求 */
export interface GetInfoAtomTypeRequest {
  id: string; // 信息原子类型ID
}

/** 获取信息原子类型详情响应 */
export interface GetInfoAtomTypeResponse extends BaseResponse {
  data?: InfoAtomType;
}

/** 更新信息原子类型请求 */
export interface UpdateInfoAtomTypeRequest {
  id: string; // 信息原子类型ID
  tenantId: string; // 租户ID
  name: string; // 类型名称
  version: string; // 版本号
  tagIds?: string[]; // 标签ID列表
  dataFormat: DataFormat; // 数据格式定义
}

/** 更新信息原子类型响应 */
export interface UpdateInfoAtomTypeResponse extends BaseResponse {}

/** 删除信息原子类型请求 */
export interface DeleteInfoAtomTypeRequest {
  id: string; // 信息原子类型ID
}

/** 删除信息原子类型响应 */
export interface DeleteInfoAtomTypeResponse extends BaseResponse {}

// ========== 逻辑图配置相关（预留） ==========

/** 节点配置 */
export interface NodeConfig {
  id: string; // 节点ID
  type: string; // 节点类型（entry/action/filter等）
  blockType: string; // 逻辑积木类型
  blockVersion: string; // 逻辑积木版本
  isEntryPoint: boolean; // 是否为入口节点
  subscribedInfoAtomTypeIDs?: string[]; // 订阅的信息原子类型ID列表
  subscribedSource?: string; // 订阅的数据来源
  subscribedLabels?: string[]; // 订阅的标签列表
  blockConfig?: string; // 逻辑积木配置（JSON字符串）
  x?: number; // 画布X坐标
  y?: number; // 画布Y坐标
}

/** 边配置 */
export interface EdgeConfig {
  id: string; // 边ID
  sourceID: string; // 源节点ID
  targetID: string; // 目标节点ID
  condition?: string; // 条件表达式
}

/** 逻辑图配置元数据 */
export interface GraphConfigMetadata {
  id: string; // 逻辑图ID（UUID）
  tenantId: string; // 租户ID
  name: string; // 图名称
  version: string; // 版本号
  description?: string; // 描述
  tags?: LynxTagSummary[]; // 标签列表（Phase 2.5: 改为完整对象）
  isEnabled: boolean; // 是否启用
  icon?: string; // 图标名称（如 'FaRobot', 'MdSensors'）
  iconColor?: string; // 图标颜色（可选，默认使用主题色）
  // 运行统计字段（后续实现）
  nodeCount?: number; // 节点数量
  edgeCount?: number; // 边数量
  executionCount?: number; // 总执行次数
  successRate?: number; // 成功率（%）
  lastExecutedAt?: number; // 最后执行时间
  // 审计字段
  createdAt?: number; // 创建时间（Unix毫秒）
  updatedAt?: number; // 更新时间（Unix毫秒）
  createdBy?: string; // 创建人ID
  updatedBy?: string; // 更新人ID
}

/** 逻辑图配置详情 */
export interface GraphConfigDetail extends GraphConfigMetadata {
  nodes: NodeConfig[]; // 节点列表
  edges: EdgeConfig[]; // 边列表
}

// ========== 逻辑图配置 API 请求/响应 ==========

/** 创建逻辑图配置请求 */
export interface CreateGraphConfigRequest {
  tenantId: string; // 租户ID
  name: string; // 图名称
  version: string; // 版本号
  description?: string; // 描述
  tagIds?: string[]; // 标签ID列表
  isEnabled?: boolean; // 是否启用（默认true）
  icon?: string; // 图标名称
  iconColor?: string; // 图标颜色
  nodes: NodeConfig[]; // 节点列表
  edges: EdgeConfig[]; // 边列表
}

/** 创建逻辑图配置响应 */
export interface CreateGraphConfigResponse extends BaseResponse {
  data?: {
    id: string; // 创建的逻辑图ID
  };
}

/** 查询逻辑图配置列表请求 */
export interface GetGraphConfigListRequest extends PageParamsRequest {
  tenantId: string; // 租户ID（必填）
  name?: string; // 图名称筛选（模糊匹配）
  tags?: string[]; // 标签筛选
  isEnabled?: boolean; // 启用状态筛选
}

/** 查询逻辑图配置列表响应 */
export interface GetGraphConfigListResponse extends BaseResponse, PageParams {
  data?: {
    list: GraphConfigMetadata[];
  };
}

/** 获取逻辑图配置详情请求 */
export interface GetGraphConfigRequest {
  id: string; // 逻辑图ID
}

/** 获取逻辑图配置详情响应 */
export interface GetGraphConfigResponse extends BaseResponse {
  data?: GraphConfigDetail;
}

/** 更新逻辑图配置请求 */
export interface UpdateGraphConfigRequest {
  id: string; // 逻辑图ID
  tenantId: string; // 租户ID
  name: string; // 图名称
  version: string; // 版本号
  description?: string; // 描述
  tagIds?: string[]; // 标签ID列表
  isEnabled?: boolean; // 是否启用
  icon?: string; // 图标名称
  iconColor?: string; // 图标颜色
  nodes: NodeConfig[]; // 节点列表
  edges: EdgeConfig[]; // 边列表
}

/** 更新逻辑图配置响应 */
export interface UpdateGraphConfigResponse extends BaseResponse {}

/** 删除逻辑图配置请求 */
export interface DeleteGraphConfigRequest {
  id: string; // 逻辑图ID
}

/** 删除逻辑图配置响应 */
export interface DeleteGraphConfigResponse extends BaseResponse {}

// ========== 逻辑积木规格相关 ==========

/** 逻辑积木规格 */
export interface BlockSpec {
  blockType: string; // 逻辑积木类型
  version: string; // 版本
  name: string; // 显示名称
  description: string; // 描述
  category: string; // 分类
  inputSchema?: string; // 输入Schema（JSON字符串）
  outputSchema?: string; // 输出Schema（JSON字符串）
  configSchema?: string; // 配置Schema（JSON字符串）
}

/** 查询逻辑积木规格列表响应 */
export interface GetBlockSpecsResponse extends BaseResponse {
  data?: {
    list: BlockSpec[];
  };
}

// ========== 标签管理相关 ==========

/** 标签作用域枚举 */
export enum TagScope {
  InfoAtom = 'info_atom',
  LogicGraph = 'logic_graph',
}

/** LynxGraph 标签 */
export interface LynxTag {
  id: string; // 标签ID（UUID）
  name: string; // 标签名称
  description?: string; // 标签描述
  scope: TagScope | string; // 标签作用域
  tenantId: string; // 租户ID
  createdBy?: string; // 创建者用户ID
  updatedBy?: string; // 最后修改者用户ID
  createdAt?: number; // 创建时间（Unix秒）
  updatedAt?: number; // 更新时间（Unix秒）
}

/** LynxGraph 标签摘要（用于列表展示和引用） - Phase 2.5 新增 */
export interface LynxTagSummary {
  id: string; // 标签ID
  name: string; // 标签名称
  description?: string; // 标签描述
  scope: TagScope | string; // 标签作用域
  createdAt?: number; // 创建时间（Unix秒）
  updatedAt?: number; // 更新时间（Unix秒）
}

/** 创建标签请求 */
export interface CreateTagRequest {
  name: string; // 标签名称（必填）
  description?: string; // 标签描述
  scope: TagScope | string; // 标签作用域（必填）
  tenantId: string; // 租户ID（必填）
  createdBy?: string; // 创建者用户ID
}

/** 创建标签响应 */
export interface CreateTagResponse extends BaseResponse {
  data?: {
    id: string; // 创建的标签ID
  };
}

/** 查询标签列表请求 */
export interface ListTagsRequest extends PageParamsRequest {
  tenantId: string; // 租户ID（必填）
  scope?: TagScope | string; // 作用域筛选
  keyword?: string; // 关键词搜索（名称）
}

/** 查询标签列表响应 */
export interface ListTagsResponse extends BaseResponse, PageParams {
  data?: {
    list: LynxTag[];
  };
}

/** 获取标签详情请求 */
export interface GetTagRequest {
  id: string; // 标签ID
}

/** 获取标签详情响应 */
export interface GetTagResponse extends BaseResponse {
  data?: LynxTag;
}

/** 更新标签请求 */
export interface UpdateTagRequest {
  id: string; // 标签ID
  name?: string; // 标签名称
  description?: string; // 标签描述
}

/** 更新标签响应 */
export interface UpdateTagResponse extends BaseResponse {}

/** 删除标签请求 */
export interface DeleteTagRequest {
  id: string; // 标签ID
}

/** 删除标签响应 */
export interface DeleteTagResponse extends BaseResponse {}

/** 批量查询标签请求 */
export interface BatchQueryTagsRequest {
  tenantId: string; // 租户ID
  scope?: TagScope | string; // 作用域（可选）
  ids?: string[]; // 标签ID列表
}

/** 批量查询标签响应 */
export interface BatchQueryTagsResponse extends BaseResponse {
  data?: {
    list: LynxTag[];
  };
}
