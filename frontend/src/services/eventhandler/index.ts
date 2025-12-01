/**
 * EventHandler API 服务封装
 */
import type {
  BaseResponse,
  CreateEventConfigRequest,
  CreateOrgMappingRequest,
  EventConfigWithFields,
  EventDetailResponse,
  FieldMetadata,
  FlowInfo,
  ListEventConfigsRequest,
  ListOrgMappingsRequest,
  OrgMapping,
  PlatformConfig,
  QueryEventDataRequest,
  QueryEventDataResponse,
  SavePlatformConfigRequest,
  TestConnectionRequest,
  UpdateEventConfigRequest,
  UpdateOrgMappingRequest,
} from "@/pages/EventManagement/types";
import { request } from "umi";
import type {
  GetDurationStatsRequest,
  GetDurationStatsResponse,
  GetNodeStatsRequest,
  GetNodeStatsResponse,
  GetOrgStatsRequest,
  GetOrgStatsResponse,
  GetPendingStatsRequest,
  GetPendingStatsResponse,
  GetStatusStatsRequest,
  GetStatusStatsResponse,
  GetTrendStatsRequest,
  GetTrendStatsResponse,
  GetUserStatsRequest,
  GetUserStatsResponse,
} from "./types";

const API_PREFIX = "/api/v1";

// ===== 平台配置相关API =====

/**
 * 获取Skylark平台配置
 */
export async function getPlatformConfig(): Promise<
  BaseResponse & { data?: PlatformConfig }
> {
  return request(`${API_PREFIX}/skylark-platform/config`, {
    method: "GET",
  });
}

/**
 * 保存（创建或更新）平台配置
 */
export async function savePlatformConfig(
  data: SavePlatformConfigRequest
): Promise<BaseResponse & { data?: { id: string } }> {
  return request(`${API_PREFIX}/skylark-platform/config`, {
    method: "POST",
    data,
  });
}

/**
 * 测试平台连接
 */
export async function testPlatformConnection(
  data: TestConnectionRequest
): Promise<
  BaseResponse & {
    data?: { success: boolean; flowsCount?: number; message?: string };
  }
> {
  return request(`${API_PREFIX}/skylark-platform/test-connection`, {
    method: "POST",
    data,
  });
}

/**
 * 获取流程列表
 * @param configuredOnly 是否只返回已配置事件的流程（默认false返回所有流程）
 */
export async function getFlowList(
  configuredOnly = false
): Promise<BaseResponse & { data?: { list: FlowInfo[] } }> {
  return request(`${API_PREFIX}/skylark-platform/flows`, {
    method: "GET",
    params: { configuredOnly },
  });
}

/**
 * 获取流程字段列表
 */
export async function getFlowFields(
  flowId: number
): Promise<
  BaseResponse & {
    data?: { systemFields: FieldMetadata[]; businessFields: FieldMetadata[] };
  }
> {
  return request(`${API_PREFIX}/skylark-platform/flows/${flowId}/fields`, {
    method: "GET",
  });
}

// ===== 事件配置管理API =====

/**
 * 创建事件配置（包含字段）
 */
export async function createEventConfig(
  data: CreateEventConfigRequest
): Promise<BaseResponse & { data?: { id: string; name: string } }> {
  return request(`${API_PREFIX}/event-configs`, {
    method: "POST",
    data,
  });
}

/**
 * 更新事件配置（包含字段）
 */
export async function updateEventConfig(
  id: string,
  data: UpdateEventConfigRequest
): Promise<BaseResponse> {
  return request(`${API_PREFIX}/event-configs/${id}`, {
    method: "PUT",
    data,
  });
}

/**
 * 获取事件配置详情（包含字段）
 */
export async function getEventConfig(
  id: string
): Promise<BaseResponse & { data?: EventConfigWithFields }> {
  return request(`${API_PREFIX}/event-configs/${id}`, {
    method: "GET",
  });
}

/**
 * 获取事件配置列表（包含字段）
 */
export async function getEventConfigList(
  params: ListEventConfigsRequest
): Promise<
  BaseResponse & {
    data?: { list: EventConfigWithFields[] };
    total?: number;
    page?: number;
  }
> {
  return request(`${API_PREFIX}/event-configs/list`, {
    method: "GET",
    params,
  });
}

/**
 * 删除事件配置
 */
export async function deleteEventConfig(id: string): Promise<BaseResponse> {
  return request(`${API_PREFIX}/event-configs/${id}`, {
    method: "DELETE",
  });
}

// ===== 组织映射管理API =====

/**
 * 创建组织映射
 */
export async function createOrgMapping(
  data: CreateOrgMappingRequest
): Promise<BaseResponse & { data?: { id: string } }> {
  return request(`${API_PREFIX}/org-mappings`, {
    method: "POST",
    data,
  });
}

/**
 * 获取组织映射列表
 */
export async function getOrgMappingList(
  params: ListOrgMappingsRequest
): Promise<
  BaseResponse & {
    data?: { list: OrgMapping[] };
    total?: number;
    page?: number;
  }
> {
  return request(`${API_PREFIX}/org-mappings/list`, {
    method: "GET",
    params,
  });
}

/**
 * 获取单个组织映射
 */
export async function getOrgMapping(
  id: string
): Promise<BaseResponse & { data?: OrgMapping }> {
  return request(`${API_PREFIX}/org-mappings/${id}`, {
    method: "GET",
  });
}

/**
 * 更新组织映射
 */
export async function updateOrgMapping(
  id: string,
  data: UpdateOrgMappingRequest
): Promise<BaseResponse> {
  return request(`${API_PREFIX}/org-mappings/${id}`, {
    method: "PUT",
    data,
  });
}

/**
 * 删除组织映射
 */
export async function deleteOrgMapping(id: string): Promise<BaseResponse> {
  return request(`${API_PREFIX}/org-mappings/${id}`, {
    method: "DELETE",
  });
}

// ===== 事件数据查询API =====

/**
 * 查询事件数据列表
 */
export async function queryEventData(
  eventId: string,
  params: QueryEventDataRequest
): Promise<QueryEventDataResponse> {
  return request(`${API_PREFIX}/events/${eventId}/data/query`, {
    method: "POST",
    data: params,
  });
}

/**
 * 获取事件详情
 */
export async function getEventDetail(
  eventId: string,
  journeyId: number
): Promise<EventDetailResponse> {
  return request(`${API_PREFIX}/events/${eventId}/data/${journeyId}`, {
    method: "GET",
  });
}

// ===== 事件统计分析API =====

/**
 * 获取处理时长统计
 */
export async function getDurationStats(
  params: GetDurationStatsRequest
): Promise<GetDurationStatsResponse> {
  return request(`${API_PREFIX}/events/stats/duration`, {
    method: "GET",
    params,
  });
}

/**
 * 获取状态统计
 */
export async function getStatusStats(
  params: GetStatusStatsRequest
): Promise<GetStatusStatsResponse> {
  return request(`${API_PREFIX}/events/stats/status`, {
    method: "GET",
    params,
  });
}

/**
 * 获取趋势统计
 */
export async function getTrendStats(
  params: GetTrendStatsRequest
): Promise<GetTrendStatsResponse> {
  return request(`${API_PREFIX}/events/stats/trend`, {
    method: "GET",
    params,
  });
}

/**
 * 获取节点统计
 */
export async function getNodeStats(
  params: GetNodeStatsRequest
): Promise<GetNodeStatsResponse> {
  return request(`${API_PREFIX}/events/stats/node`, {
    method: "GET",
    params,
  });
}

/**
 * 获取处理人统计
 */
export async function getUserStats(
  params: GetUserStatsRequest
): Promise<GetUserStatsResponse> {
  return request(`${API_PREFIX}/events/stats/user`, {
    method: "GET",
    params,
  });
}

/**
 * 获取组织统计
 */
export async function getOrgStats(
  params: GetOrgStatsRequest
): Promise<GetOrgStatsResponse> {
  return request(`${API_PREFIX}/events/stats/org`, {
    method: "GET",
    params,
  });
}

/**
 * 获取待处理事件统计
 */
export async function getPendingStats(
  params: GetPendingStatsRequest
): Promise<GetPendingStatsResponse> {
  return request(`${API_PREFIX}/events/stats/pending`, {
    method: "GET",
    params,
  });
}
