/**
 * 事件管理相关的TypeScript类型定义
 */

// ===== 基础响应类型 =====
export interface BaseResponse {
  code: number;
  msg: string;
  message?: string;
}

export interface PageParams {
  page?: number;
  pageSize?: number;
  total?: number;
}

// ===== 平台配置相关类型 =====
export interface PlatformConfig {
  id: string;
  tenantId: string;
  host: string;
  port: number;
  database: string;
  username: string;
  password?: string;
  namespaceId: number;
  apiBaseUrl: string;
  apiToken?: string;
  createdBy?: string;
  updatedBy?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface SavePlatformConfigRequest {
  host: string;
  port: number;
  database: string;
  username: string;
  password: string;
  namespaceId: number;
  apiBaseUrl: string;
  apiToken: string;
}

export interface TestConnectionRequest {
  host: string;
  port: number;
  database: string;
  username: string;
  password: string;
  namespaceId: number;
  apiBaseUrl: string;
  apiToken: string;
}

export interface FlowInfo {
  id: number;
  title: string;
  namespaceId: number;
}

export interface FieldMetadata {
  fieldName: string;
  dataType: string;
  isSystem: boolean;
}

// ===== 事件配置相关类型 =====
export interface EventConfig {
  id: string;
  name: string;
  flowId: number;
  flowTitle?: string;
  orgFieldName?: string;
  description?: string;
  enabled: boolean;
  tenantId: string;
  createdBy?: string;
  updatedBy?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface FieldConfig {
  id: string;
  eventConfigId: string;
  fieldName: string;
  displayName: string;
  fieldType: string;
  isVisible: boolean;
  displayOrder: number;
  isSearchable: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface EventConfigWithFields extends EventConfig {
  fields: FieldConfig[];
}

export interface FieldConfigInput {
  fieldName: string;
  displayName: string;
  fieldType: string;
  isVisible: boolean;
  displayOrder: number;
  isSearchable: boolean;
}

export interface CreateEventConfigRequest {
  name: string;
  flowId: number;
  orgFieldName?: string;
  description?: string;
  enabled?: boolean;
  fields: FieldConfigInput[];
}

export interface UpdateEventConfigRequest {
  name?: string;
  orgFieldName?: string;
  description?: string;
  enabled?: boolean;
  fields: FieldConfigInput[]; // 必填，完整替换策略
}

// ===== 组织映射相关类型 =====
export interface OrgMapping {
  id: string;
  tenantId: string;
  remoteOrgValue: string;
  localOrgId: string;
  localOrgName?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface CreateOrgMappingRequest {
  remoteOrgValue: string;
  localOrgId: string;
}

export interface UpdateOrgMappingRequest {
  remoteOrgValue?: string;
  localOrgId?: string;
}

// ===== 事件数据查询相关类型 =====
export interface ColumnInfo {
  field: string;
  displayName: string;
  type: string;
}

export interface FlowNode {
  vertexId: number;
  vertexName: string;
  vertexAlias: string;
  userIds: string[]; // 改为数组，支持同一节点多人处理
  userNames: string[]; // 改为数组，与userIds一一对应
  // status 字段已移除（历史节点的状态无意义，只有currentStatus有意义）
  // assignmentId 字段已移除（不需要暴露给前端）
  createdAt: string;
  updatedAt: string;
}

export interface QueryEventDataRequest {
  eventId: string; // 事件配置ID（路径参数）
  orgId: string; // 组织ID（必填筛选条件）
  page?: number;
  pageSize?: number;
  keyword?: string;
  sortField?: string;
  sortOrder?: string;
  status?: string[]; // 状态筛选（可选，支持多个状态：processing/finished/cancelled等）
}

export interface QueryEventDataResponse extends BaseResponse {
  data?: {
    columns: ColumnInfo[];
    records: Record<string, any>[];
    total: number;
  };
}

export interface EventDetailResponse extends BaseResponse {
  data?: {
    journeyId: number;
    currentStatus: string;
    initiatorUserId: string;
    initiatorUserName: string;
    initiatedAt: string;
    latestBusinessData: Record<string, any>;
    flowHistory: FlowNode[];
  };
}

// ===== 列表请求参数类型 =====
export interface ListEventConfigsRequest extends PageParams {
  keyword?: string;
  enabled?: boolean;
}

export interface ListOrgMappingsRequest extends PageParams {
  keyword?: string;
}

// ===== 字段类型枚举 =====
export enum FieldType {
  STRING = "string",
  NUMBER = "number",
  DATE = "date",
  DATETIME = "datetime",
  BOOLEAN = "boolean",
  IMAGE_BASE64 = "imageBase64",
}

// ===== 流程状态枚举 =====
export enum FlowStatus {
  PROCESSING = "processing",
  COMPLETED = "completed",
  REJECTED = "rejected",
  PENDING = "pending",
}
