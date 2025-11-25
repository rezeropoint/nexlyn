/**
 * 工作台业务流程管理 API 服务封装
 */
import { request } from "umi";
import type {
  AbortJourneyParams,
  AbortJourneyResponse,
  GetJourneyFullDetailParams,
  GetJourneyFullDetailResponse,
  GetProposedJourneysParams,
  GetProposedJourneysResponse,
  GetUserAssignmentsParams,
  GetUserAssignmentsResponse,
  SearchJourneysParams,
  SearchJourneysResponse,
  UpdateFlowJourneyStatusParams,
  UpdateFlowJourneyStatusResponse,
} from "./types";

// 导出统计接口
export * from "./stats";

const API_PREFIX = "/api/v1";

// ===== 用户任务相关API =====

/**
 * 获取用户任务列表
 * @param params - 查询参数
 * @param params.category - 类别：processed（待办）/proposed（我发起的）/cc（抄送）
 * @param params.page - 页码
 * @param params.pageSize - 每页数量
 * @returns 任务列表和总数
 */
export async function getUserAssignments(
  params: GetUserAssignmentsParams
): Promise<GetUserAssignmentsResponse> {
  return request(`${API_PREFIX}/my-assignments`, {
    method: "GET",
    params,
  });
}

/**
 * 获取用户发起的流程列表
 * @param params - 查询参数
 * @param params.page - 页码
 * @param params.pageSize - 每页数量
 * @returns 流程列表和总数
 */
export async function getProposedJourneys(
  params: GetProposedJourneysParams
): Promise<GetProposedJourneysResponse> {
  return request(`${API_PREFIX}/my-proposed-journeys`, {
    method: "GET",
    params,
  });
}

// ===== 流程Journey操作相关API =====

/**
 * 终止流程
 * @param params - 请求参数
 * @param params.flowId - 流程ID
 * @param params.journeyId - Journey ID
 * @returns 操作结果
 */
export async function abortJourney(
  params: AbortJourneyParams
): Promise<AbortJourneyResponse> {
  const { flowId, journeyId } = params;
  return request(`${API_PREFIX}/flows/${flowId}/journeys/${journeyId}`, {
    method: "DELETE",
  });
}

/**
 * 获取流程完整详情（一站式接口，包含基础信息、审批历史、待处理节点）
 * @param params - 请求参数
 * @param params.flowId - 流程ID
 * @param params.journeyId - Journey ID
 * @returns 流程完整详情
 */
export async function getJourneyFullDetail(
  params: GetJourneyFullDetailParams
): Promise<GetJourneyFullDetailResponse> {
  const { flowId, journeyId } = params;
  return request(`${API_PREFIX}/flows/${flowId}/journeys/${journeyId}/full-detail`, {
    method: "GET",
  });
}

/**
 * 高级搜索流程记录
 * @param params - 搜索条件
 * @param params.flowId - 流程ID
 * @param params.status - 流程状态（可选）
 * @param params.keyword - 关键词（可选）
 * @param params.initiatorId - 发起人ID（可选）
 * @param params.createdFrom - 创建时间起始（可选）
 * @param params.createdTo - 创建时间结束（可选）
 * @param params.page - 页码
 * @param params.pageSize - 每页数量
 * @returns 流程列表和总数
 */
export async function searchJourneys(
  params: SearchJourneysParams
): Promise<SearchJourneysResponse> {
  const { flowId, ...searchParams } = params;
  return request(`${API_PREFIX}/flows/${flowId}/journeys/search`, {
    method: "POST",
    data: searchParams,
  });
}

/**
 * 更新流程Journey状态（审批操作）
 * @param params - 请求参数
 * @param params.flowId - 流程ID
 * @param params.journeyId - Journey ID
 * @param params.assignmentId - Assignment ID
 * @param params.operation - 操作类型：approve/refuse/transfer/cancel
 * @param params.comment - 处理意见（可选）
 * @param params.carbonCopyUserIds - 抄送用户UUID列表（可选）
 * @param params.data - 字段数据更新（可选）
 * @returns 操作结果
 */
export async function updateFlowJourneyStatus(
  params: UpdateFlowJourneyStatusParams
): Promise<UpdateFlowJourneyStatusResponse> {
  const { flowId, journeyId, ...actionData } = params;
  return request(`${API_PREFIX}/flows/${flowId}/journeys/${journeyId}/actions`, {
    method: "POST",
    data: actionData,
  });
}

// ===== 流程元数据相关API =====
// 注意：getFlowDetail 已被移除，流程元数据现在包含在 getJourneyFullDetail 响应中
