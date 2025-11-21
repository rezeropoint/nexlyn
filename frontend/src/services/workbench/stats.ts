/**
 * 工作台统计 API 服务封装
 */
import { request } from "umi";
import type {
  GetDurationStatsParams,
  GetDurationStatsResponse,
  GetNodeStatsParams,
  GetNodeStatsResponse,
  GetOrgStatsParams,
  GetOrgStatsResponse,
  GetPendingStatsParams,
  GetPendingStatsResponse,
  GetStatusStatsParams,
  GetStatusStatsResponse,
  GetTrendStatsParams,
  GetTrendStatsResponse,
  GetUserStatsParams,
  GetUserStatsResponse,
} from "./types";

const API_PREFIX = "/api/v1";

// ===== 统计相关API =====

/**
 * 获取待处理事件统计
 * @param params - 查询参数
 * @returns 待处理事件统计数据
 */
export async function getPendingStats(
  params: GetPendingStatsParams
): Promise<GetPendingStatsResponse> {
  return request(`${API_PREFIX}/events/stats/pending`, {
    method: "GET",
    params: {
      ...params,
      eventIds: params.eventIds?.join(","), // 数组转字符串
    },
  });
}

/**
 * 获取状态统计
 * @param params - 查询参数
 * @returns 状态统计数据
 */
export async function getStatusStats(
  params: GetStatusStatsParams
): Promise<GetStatusStatsResponse> {
  return request(`${API_PREFIX}/events/stats/status`, {
    method: "GET",
    params: {
      ...params,
      eventIds: params.eventIds?.join(","), // 数组转字符串
    },
  });
}

/**
 * 获取趋势统计
 * @param params - 查询参数
 * @returns 趋势统计数据
 */
export async function getTrendStats(
  params: GetTrendStatsParams
): Promise<GetTrendStatsResponse> {
  return request(`${API_PREFIX}/events/stats/trend`, {
    method: "GET",
    params: {
      ...params,
      eventIds: params.eventIds?.join(","), // 数组转字符串
    },
  });
}

/**
 * 获取处理时长统计
 * @param params - 查询参数
 * @returns 处理时长统计数据
 */
export async function getDurationStats(
  params: GetDurationStatsParams
): Promise<GetDurationStatsResponse> {
  return request(`${API_PREFIX}/events/stats/duration`, {
    method: "GET",
    params: {
      ...params,
      eventIds: params.eventIds?.join(","), // 数组转字符串
    },
  });
}

/**
 * 获取处理人统计
 * @param params - 查询参数
 * @returns 处理人统计数据
 */
export async function getUserStats(
  params: GetUserStatsParams
): Promise<GetUserStatsResponse> {
  return request(`${API_PREFIX}/events/stats/user`, {
    method: "GET",
    params: {
      ...params,
      eventIds: params.eventIds?.join(","), // 数组转字符串
    },
  });
}

/**
 * 获取组织统计
 * @param params - 查询参数
 * @returns 组织统计数据
 */
export async function getOrgStats(
  params: GetOrgStatsParams
): Promise<GetOrgStatsResponse> {
  return request(`${API_PREFIX}/events/stats/org`, {
    method: "GET",
    params: {
      ...params,
      eventIds: params.eventIds?.join(","), // 数组转字符串
    },
  });
}

/**
 * 获取节点统计
 * @param params - 查询参数
 * @returns 节点统计数据
 */
export async function getNodeStats(
  params: GetNodeStatsParams
): Promise<GetNodeStatsResponse> {
  return request(`${API_PREFIX}/events/stats/node`, {
    method: "GET",
    params: {
      ...params,
      eventIds: params.eventIds?.join(","), // 数组转字符串
    },
  });
}

