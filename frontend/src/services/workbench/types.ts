/**
 * 工作台业务流程管理类型定义
 */

// ===== 通用响应类型 =====
export interface BaseResponse {
  code: number;
  msg: string;
}

// ===== 核心数据类型 =====

/**
 * 流程用户信息
 */
export interface FlowUser {
  id: number; // 用户ID
  name: string; // 用户姓名
  nickname?: string; // 用户昵称
  identifier?: string; // 工号/身份证号
}

/**
 * Assignment 任务
 */
export interface Assignment {
  id: number; // Assignment ID
  assigneeId: number; // 处理人远程用户ID
  status: string; // 状态：processing/completed
  category: string; // 类别：processed（待办）/proposed（我发起的）/cc（抄送）
  vertexId: number; // 节点ID
  journeyId: number; // Journey ID
  createdAt: string; // 创建时间
  updatedAt: string; // 更新时间
  flowId?: number; // 流程ID（SDK自动补充）
  flowTitle?: string; // 流程标题（SDK自动补充）
}

/**
 * Journey 流程记录
 */
export interface Journey {
  id: number; // Journey ID
  sn: string; // 流程编号
  status: string; // 流程状态：processing/completed/aborted/stashed
  currentVertexId: number; // 当前节点ID
  flowId: number; // 流程ID
  createdAt: string; // 创建时间
  updatedAt: string; // 更新时间
  user?: FlowUser; // 发起人信息
}

/**
 * 附件信息
 */
export interface Attachment {
  id: number; // 附件ID
  name: string; // 文件名
  url: string; // 下载URL
  size: number; // 文件大小（字节）
  mimeType: string; // MIME类型
  uploadedBy: number; // 上传者用户ID
  createdAt: string; // 上传时间
}

/**
 * JourneyDetail 流程详情（用于 JourneyFullDetail）
 */
export interface JourneyDetail {
  id: number; // Journey ID
  sn: string; // 流程编号
  status: string; // 流程状态
  currentVertexId: number; // 当前节点ID
  flowId: number; // 流程ID
  createdAt: string; // 创建时间
  updatedAt: string; // 更新时间
  journeyUrl?: string; // Skylark流程详情页URL
  reviewerVertexIds?: number[]; // 审批节点ID列表
  currentDurationThreshold?: string; // 当前节点超时阈值
  initiator?: FlowUser; // 发起人信息
  businessData?: Record<string, any>; // 业务字段值（图片字段为URL）
}

/**
 * Moment 审批历史节点（用于 JourneyFullDetail）
 */
export interface Moment {
  id: number; // Moment ID
  assignmentId: number; // Assignment ID
  journeyId: number; // Journey ID
  vertexId: number; // 节点ID
  vertexName?: string; // 节点名称
  status: string; // 状态：approved/refused/transferred/cancelled
  operatorId: number; // 操作人远程用户ID
  operatorName?: string; // 操作人姓名
  comment?: string; // 处理意见
  createdAt: string; // 操作时间
  updatedAt: string; // 更新时间
}

/**
 * ProcessingUser 当前处理人（用于前端展示）
 */
export interface ProcessingUser {
  id: string; // 用户UUID
  name: string; // 用户姓名
  nickname?: string; // 用户昵称
  phone?: string; // 手机号
  identifier?: string; // 工号/身份证号
  headimgurl?: string; // 头像URL
  tags?: string[]; // 用户标签
}

/**
 * FieldOption 字段选项
 */
export interface FieldOption {
  id: number; // 选项ID
  value: string; // 选项值
  settings?: Record<string, any>; // 选项设置
  position: number; // 选项位置
}

/**
 * VertexField 节点字段
 */
export interface VertexField {
  id: number; // 字段ID
  identityKey: string; // 字段唯一标识键（用于构建Data参数）
  title: string; // 字段标题
  type: string; // 字段类型（如 Field::RadioButton, Field::TextField）
  required: boolean; // 是否必填
  editable: boolean; // 是否可编辑
  maxLength: number; // 字数限制（0表示无限制）
  options: FieldOption[]; // 字段可选项
}

/**
 * PendingNode 待处理节点
 */
export interface PendingNode {
  vertexId: number; // 节点ID
  vertexName: string; // 节点名称
  assigneeIds: string[]; // 待处理人ID列表（调用方可根据ID自行查询用户信息）
  createdAt: string; // 任务创建时间
  fields: VertexField[]; // 节点字段列表
}

/**
 * FlowVertex 流程节点信息
 */
export interface FlowVertex {
  id: number; // 节点ID
  title: string; // 节点名称
  type: string; // 节点类型
  description?: string; // 节点描述
}

/**
 * JourneyFullDetail 流程完整详情
 */
export interface JourneyFullDetail {
  basicInfo: JourneyDetail; // 基础信息
  history: Moment[]; // 审批历史
  pendingNodes: PendingNode[]; // 待处理节点
  vertices: Record<string, FlowVertex>; // 节点映射（key为节点ID字符串）
}

// ===== 请求参数类型 =====

/**
 * 获取用户任务列表请求参数
 */
export interface GetUserAssignmentsParams {
  category: string; // 类别：processed（待办）/proposed（我发起的）/cc（抄送）
  page: number; // 页码
  pageSize: number; // 每页数量
}

/**
 * 获取用户发起的流程请求参数
 */
export interface GetProposedJourneysParams {
  flowId: number; // 流程ID
  page: number; // 页码
  pageSize: number; // 每页数量
}

// ===== 响应数据类型 =====

/**
 * 任务列表响应数据
 */
export interface AssignmentListData {
  list: Assignment[]; // 任务列表
  total: number; // 总数
}

/**
 * 流程列表响应数据
 */
export interface JourneyListData {
  list: Journey[]; // 流程列表
  total: number; // 总数
}

/**
 * 获取用户任务列表响应
 */
export interface GetUserAssignmentsResponse extends BaseResponse {
  data: AssignmentListData;
}

/**
 * 获取用户发起的流程响应
 */
export interface GetProposedJourneysResponse extends BaseResponse {
  data: JourneyListData;
}

// ===== 流程元数据类型 =====

/**
 * TypedValue 类型化值（Skylark字段系统）
 */
export interface TypedValue {
  type: string; // string/number/boolean/date/array/object/file
  value: any; // 字段值
}

/**
 * 流程字段
 */
export interface FlowField {
  name: string; // 字段名
  title: string; // 字段标题
  type: string; // 字段类型：TextField/NumberField/DateField等
  required: boolean; // 是否必填
  description?: string; // 字段描述
}

/**
 * 流程节点
 */
export interface FlowVertex {
  id: number; // 节点ID
  title: string; // 节点标题
  type: string; // 节点类型：start/end/approve/cc等
  description?: string; // 节点描述
}

/**
 * 流程边
 */
export interface FlowEdge {
  id: number; // 边ID
  sourceVertex: number; // 源节点ID
  targetVertex: number; // 目标节点ID
  condition?: string; // 条件表达式
}

/**
 * 流程元数据
 */
export interface FlowDetailData {
  id: number; // 流程ID
  title: string; // 流程标题
  fields: FlowField[]; // 字段列表
  vertices: FlowVertex[]; // 节点列表
  edges: FlowEdge[]; // 边列表
}

/**
 * 获取流程元数据响应
 */
export interface GetFlowDetailResponse extends BaseResponse {
  data: FlowDetailData;
}

// ===== 流程Journey操作请求参数类型 =====

/**
 * 终止流程请求参数
 */
export interface AbortJourneyParams {
  flowId: number; // 流程ID
  journeyId: number; // Journey ID
}

/**
 * 获取流程完整详情请求参数
 */
export interface GetJourneyFullDetailParams {
  flowId: number; // 流程ID
  journeyId: number; // Journey ID
}

/**
 * 流程搜索条件
 */
export interface SearchJourneysParams {
  flowId: number; // 流程ID（路径参数）
  status?: string; // 流程状态
  keyword?: string; // 关键词
  initiatorId?: string; // 发起人ID（本地用户UUID）
  createdFrom?: string; // 创建时间起始
  createdTo?: string; // 创建时间结束
  page: number; // 页码
  pageSize: number; // 每页数量
}

/**
 * 更新流程状态请求参数
 */
export interface UpdateFlowJourneyStatusParams {
  flowId: number; // 流程ID
  journeyId: number; // Journey ID
  assignmentId: number; // Assignment ID（必填）
  operation: string; // 操作类型：approve/refuse/transfer/cancel
  comment?: string; // 处理意见（可选）
  carbonCopyUserIds?: string[]; // 抄送用户UUID列表（可选）
  data?: Record<string, TypedValue>; // 字段数据更新（可选）
}

// ===== 流程Journey操作响应类型 =====

/**
 * 终止流程响应
 */
export interface AbortJourneyResponse extends BaseResponse {}

/**
 * 获取流程完整详情响应
 */
export interface GetJourneyFullDetailResponse extends BaseResponse {
  data: JourneyFullDetail;
}

/**
 * 搜索流程响应数据
 */
export interface SearchJourneysData {
  list: Journey[]; // 流程列表
  total: number; // 总数
}

/**
 * 搜索流程响应
 */
export interface SearchJourneysResponse extends BaseResponse {
  data: SearchJourneysData;
}

/**
 * 更新流程状态响应
 */
export interface UpdateFlowJourneyStatusResponse extends BaseResponse {}

// ===== 统计数据类型 =====

/**
 * 统计查询通用参数
 */
export interface StatsQueryParams {
  eventIds?: string[]; // 事件配置ID列表（可选，支持单个或多个ID）
  orgId: string; // 组织ID（必填）
  dateFrom?: string; // 开始日期（ISO格式，可选）
  dateTo?: string; // 结束日期（ISO格式，可选）
  status?: string; // 按状态筛选（可选）
}

/**
 * 状态统计项
 */
export interface StatusCount {
  status: string; // 状态显示名称（中文）
  statusKey: string; // 状态原始值（英文）
  count: number; // 该状态的事件数量
  percentage: number; // 占比（0-100）
}

/**
 * 趋势统计点
 */
export interface TrendPoint {
  date: string; // 日期（YYYY-MM-DD 或 YYYY-WW 或 YYYY-MM）
  totalCount: number; // 该时间段新增的总事件数
  completedCount: number; // 该时间段完成的事件数
  completionRate: number; // 完成率（0-100）
}

/**
 * 节点统计指标
 */
export interface NodeMetric {
  vertexId: number; // 节点ID
  vertexName: string; // 节点名称
  count: number; // 该节点的事件数量
  avgDuration: number; // 该节点平均停留时长（秒）
}

/**
 * 处理人统计指标
 */
export interface UserMetric {
  userId: string; // 用户ID
  userName: string; // 用户姓名
  count: number; // 处理的事件数量
  rank: number; // 排名（1开始）
}

/**
 * 组织统计指标
 */
export interface OrgMetric {
  orgValue: string; // 组织字段值
  count: number; // 该组织的事件数量
  avgDuration: number; // 该组织的平均处理时长（秒）
}

/**
 * 事件待处理统计
 */
export interface EventPendingStats {
  eventConfigId: string; // 事件配置ID
  eventName: string; // 事件名称
  pendingCount: number; // 未开始事件数
  processingCount: number; // 处理中事件数
  total: number; // 该事件的总数（未完成的事件）
}

// ===== 统计请求参数类型 =====

/**
 * 获取处理时长统计请求参数
 */
export interface GetDurationStatsParams extends StatsQueryParams {}

/**
 * 获取状态统计请求参数
 */
export interface GetStatusStatsParams extends StatsQueryParams {}

/**
 * 获取趋势统计请求参数
 */
export interface GetTrendStatsParams extends StatsQueryParams {
  groupBy?: "day" | "week" | "month"; // 聚合维度（默认day）
}

/**
 * 获取节点统计请求参数
 */
export interface GetNodeStatsParams extends StatsQueryParams {}

/**
 * 获取处理人统计请求参数
 */
export interface GetUserStatsParams extends StatsQueryParams {
  topN?: number; // Top N数量（默认10）
}

/**
 * 获取组织统计请求参数
 */
export interface GetOrgStatsParams extends StatsQueryParams {}

/**
 * 获取待处理事件统计请求参数
 */
export interface GetPendingStatsParams extends StatsQueryParams {}

// ===== 统计响应类型 =====

/**
 * 处理时长统计响应
 */
export interface GetDurationStatsResponse extends BaseResponse {
  data?: {
    avgDuration: number; // 平均处理时长（秒）
    minDuration: number; // 最短处理时长（秒）
    maxDuration: number; // 最长处理时长（秒）
    totalCount: number; // 统计范围内的总事件数
    completedCount: number; // 已完成事件数
  };
}

/**
 * 状态统计响应
 */
export interface GetStatusStatsResponse extends BaseResponse {
  data?: {
    statusCounts: StatusCount[]; // 各状态数量（按数量降序）
    total: number; // 总事件数量
  };
}

/**
 * 趋势统计响应
 */
export interface GetTrendStatsResponse extends BaseResponse {
  data?: {
    timePoints: TrendPoint[]; // 时间点数据（按时间正序）
  };
}

/**
 * 节点统计响应
 */
export interface GetNodeStatsResponse extends BaseResponse {
  data?: {
    nodeMetrics: NodeMetric[]; // 节点指标列表（按事件数量降序）
  };
}

/**
 * 处理人统计响应
 */
export interface GetUserStatsResponse extends BaseResponse {
  data?: {
    userMetrics: UserMetric[]; // 用户指标列表（按处理数量降序）
    total: number; // 统计范围内的总处理人数
  };
}

/**
 * 组织统计响应
 */
export interface GetOrgStatsResponse extends BaseResponse {
  data?: {
    orgMetrics: OrgMetric[]; // 组织指标列表（按事件数量降序）
  };
}

/**
 * 待处理事件统计响应
 */
export interface GetPendingStatsResponse extends BaseResponse {
  data?: {
    eventStats: EventPendingStats[]; // 按事件配置分组的统计（按事件名称排序）
    totalPending: number; // 所有事件的未开始总数
    totalProcessing: number; // 所有事件的处理中总数
    total: number; // 总事件数（未完成的事件）
  };
}