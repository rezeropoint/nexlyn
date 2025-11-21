/**
 * 工作台业务流程管理 API 服务封装
 */
import { request } from "umi";
import type {
  AbortJourneyParams,
  AbortJourneyResponse,
  GetCurrentProcessingUsersParams,
  GetCurrentProcessingUsersResponse,
  GetFlowDetailParams,
  GetFlowDetailResponse,
  GetFlowJourneyBySNParams,
  GetFlowJourneyBySNResponse,
  GetFlowJourneyDetailParams,
  GetFlowJourneyDetailResponse,
  GetJourneyMomentsParams,
  GetJourneyMomentsResponse,
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

// ===== 流程详情相关API =====

/**
 * 获取流程详情
 * @param params - 请求参数
 * @param params.flowId - 流程ID
 * @param params.journeyId - Journey ID
 * @returns 流程详情数据
 */
export async function getFlowJourneyDetail(
  params: GetFlowJourneyDetailParams
): Promise<GetFlowJourneyDetailResponse> {
  const { flowId, journeyId } = params;
  return request(`${API_PREFIX}/flows/${flowId}/journeys/${journeyId}/detail`, {
    method: "GET",
  });
}

/**
 * 获取审批历史时间线
 * @param params - 请求参数
 * @param params.flowId - 流程ID
 * @param params.journeyId - Journey ID
 * @returns 审批历史记录列表
 */
export async function getJourneyMoments(
  params: GetJourneyMomentsParams
): Promise<GetJourneyMomentsResponse> {
  const { flowId, journeyId } = params;
  return request(`${API_PREFIX}/flows/${flowId}/journeys/${journeyId}/moments`, {
    method: "GET",
  });
}

/**
 * 获取当前处理人列表
 * @param params - 请求参数
 * @param params.flowId - 流程ID
 * @param params.journeyId - Journey ID
 * @returns 当前处理人列表
 */
export async function getCurrentProcessingUsers(
  params: GetCurrentProcessingUsersParams
): Promise<GetCurrentProcessingUsersResponse> {
  const { flowId, journeyId } = params;
  return request(
    `${API_PREFIX}/flows/${flowId}/journeys/${journeyId}/processing-users`,
    {
      method: "GET",
    }
  );
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
 * 根据流程编号查询
 * @param params - 请求参数
 * @param params.flowId - 流程ID
 * @param params.sn - 流程编号
 * @returns 流程记录
 */
export async function getFlowJourneyBySN(
  params: GetFlowJourneyBySNParams
): Promise<GetFlowJourneyBySNResponse> {
  const { flowId, sn } = params;
  return request(`${API_PREFIX}/flows/${flowId}/journeys/search-by-sn`, {
    method: "GET",
    params: { sn },
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

/**
 * 获取流程元数据（字段、节点、边）
 * @param params - 请求参数
 * @param params.flowId - 流程ID
 * @returns 流程元数据
 */
export async function getFlowDetail(
  params: GetFlowDetailParams
): Promise<GetFlowDetailResponse> {
  const { flowId } = params;
  return request(`${API_PREFIX}/flows/${flowId}/detail`, {
    method: "GET",
  });
}
