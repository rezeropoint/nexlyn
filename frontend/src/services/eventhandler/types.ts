/**
 * EventHandler 事件统计分析类型定义
 */

// ===== 通用响应类型 =====
export interface BaseResponse {
  code: number;
  msg: string;
}

// ===== 通用统计请求参数 =====
export interface StatsQueryParams {
  eventIds: string[]; // 事件配置ID列表（必填，支持单个或多个ID）
  orgId: string; // 组织ID（必填，前端指定，后端验证权限）
  dateFrom?: string; // 开始日期（ISO格式，可选）
  dateTo?: string; // 结束日期（ISO格式，可选）
  status?: string; // 按状态筛选（可选，如"completed"）
}

// ===== 1. 处理时长统计 =====
export interface GetDurationStatsRequest extends StatsQueryParams {}

export interface DurationStatsData {
  avgDuration: number; // 平均处理时长（秒）
  minDuration: number; // 最短处理时长（秒）
  maxDuration: number; // 最长处理时长（秒）
  totalCount: number; // 统计范围内的总事件数
  completedCount: number; // 已完成事件数
}

export interface GetDurationStatsResponse extends BaseResponse {
  data?: DurationStatsData;
}

// ===== 2. 状态统计 =====
export interface GetStatusStatsRequest extends StatsQueryParams {}

export interface StatusCount {
  status: string; // 状态显示名称（中文）
  statusKey: string; // 状态原始值（英文）
  count: number; // 该状态的事件数量
  percentage: number; // 占比（0-100）
}

export interface StatusStatsData {
  statusCounts: StatusCount[]; // 各状态数量（按数量降序）
  total: number; // 总事件数量
}

export interface GetStatusStatsResponse extends BaseResponse {
  data?: StatusStatsData;
}

// ===== 3. 趋势统计 =====
export interface GetTrendStatsRequest extends StatsQueryParams {
  groupBy?: "day" | "week" | "month"; // 聚合维度（默认day）
}

export interface TrendPoint {
  date: string; // 日期（YYYY-MM-DD 或 YYYY-WW 或 YYYY-MM）
  totalCount: number; // 该时间段新增的总事件数
  completedCount: number; // 该时间段完成的事件数
  completionRate: number; // 完成率（0-100）
}

export interface TrendStatsData {
  timePoints: TrendPoint[]; // 时间点数据（按时间正序）
}

export interface GetTrendStatsResponse extends BaseResponse {
  data?: TrendStatsData;
}

// ===== 4. 节点统计 =====
export interface GetNodeStatsRequest extends StatsQueryParams {}

export interface NodeMetric {
  vertexId: number; // 节点ID
  vertexName: string; // 节点名称
  count: number; // 该节点的事件数量
  avgDuration: number; // 该节点平均停留时长（秒）
}

export interface NodeStatsData {
  nodeMetrics: NodeMetric[]; // 节点指标列表（按事件数量降序）
}

export interface GetNodeStatsResponse extends BaseResponse {
  data?: NodeStatsData;
}

// ===== 5. 处理人统计 =====
export interface GetUserStatsRequest extends StatsQueryParams {
  topN?: number; // Top N数量（默认10）
}

export interface UserMetric {
  userId: string; // 用户ID
  userName: string; // 用户姓名
  count: number; // 处理的事件数量
  rank: number; // 排名（1开始）
}

export interface UserStatsData {
  userMetrics: UserMetric[]; // 用户指标列表（按处理数量降序）
  total: number; // 统计范围内的总处理人数
}

export interface GetUserStatsResponse extends BaseResponse {
  data?: UserStatsData;
}

// ===== 6. 组织统计 =====
export interface GetOrgStatsRequest extends StatsQueryParams {}

export interface OrgMetric {
  orgValue: string; // 组织字段值
  count: number; // 该组织的事件数量
  avgDuration: number; // 该组织的平均处理时长（秒）
}

export interface OrgStatsData {
  orgMetrics: OrgMetric[]; // 组织指标列表（按事件数量降序）
}

export interface GetOrgStatsResponse extends BaseResponse {
  data?: OrgStatsData;
}

// ===== 7. 待处理事件统计 =====
export interface GetPendingStatsRequest extends StatsQueryParams {}

// 单个事件的待处理统计
export interface EventPendingStats {
  eventConfigId: string; // 事件配置ID
  eventName: string; // 事件名称
  pendingCount: number; // 该事件的未开始数量
  processingCount: number; // 该事件的处理中数量
  total: number; // 该事件的总未完成数量
}

export interface PendingStatsData {
  eventStats: EventPendingStats[]; // 各事件的待处理统计数组
  totalPending: number; // 所有事件的未开始事件总数
  totalProcessing: number; // 所有事件的处理中事件总数
  total: number; // 所有事件的未完成事件总数
}

export interface GetPendingStatsResponse extends BaseResponse {
  data?: PendingStatsData;
}
