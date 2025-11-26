/**
 * HTTP 接收配置管理 API
 */
import { request } from "@umijs/max";
import { API_PREFIX } from "../constants";
import type { BaseResponse, PageParams } from "../types/common";
import type {
  HttpReceiveDetail,
  HttpReceiveDispatchConfig,
  HttpReceiveFieldMapping,
  HttpReceiveMetadata,
} from "../types";

// 获取配置列表请求
export interface GetHttpReceiveListRequest {
  page?: number;
  pageSize?: number;
  keyword?: string; // 关键词搜索（名称/描述）
}

// 获取配置列表响应
export interface GetHttpReceiveListResponse extends BaseResponse, PageParams {
  data?: {
    list: HttpReceiveMetadata[];
  };
}

export async function getHttpReceiveList(
  params?: GetHttpReceiveListRequest,
  options?: { [key: string]: any }
) {
  return request<GetHttpReceiveListResponse>(
    `${API_PREFIX}/http-receive/list`,
    {
      method: "GET",
      params,
      ...(options || {}),
    }
  );
}

// 获取配置详情响应
export interface GetHttpReceiveResponse extends BaseResponse {
  data?: HttpReceiveDetail;
}

export async function getHttpReceive(
  id: string,
  options?: { [key: string]: any }
) {
  return request<GetHttpReceiveResponse>(`${API_PREFIX}/http-receive/${id}`, {
    method: "GET",
    ...(options || {}),
  });
}

// 创建配置请求
export interface CreateHttpReceiveRequest {
  name: string; // 配置名称（必填）
  description?: string; // 配置描述
  enabled?: boolean; // 是否启用（默认true）
  timestampPath?: string; // 时间戳字段路径
  timestampFormat?: string; // 时间戳格式：unix/unix_ms/unix_nano/iso8601/rfc3339
  deviceIdPath?: string; // 设备ID字段路径
  fieldMappings: HttpReceiveFieldMapping[]; // 字段映射配置（必填，使用自定义字段名）
  dispatchConfigs?: HttpReceiveDispatchConfig[]; // 分发配置
}

// 创建配置响应
export interface CreateHttpReceiveResponse extends BaseResponse {
  data?: {
    id: string; // 配置ID
    name: string; // 配置名称
  };
}

export async function createHttpReceive(
  data: CreateHttpReceiveRequest,
  options?: { [key: string]: any }
) {
  return request<CreateHttpReceiveResponse>(`${API_PREFIX}/http-receive`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    data,
    ...(options || {}),
  });
}

// 更新配置请求
export interface UpdateHttpReceiveRequest {
  name?: string; // 配置名称
  description?: string; // 配置描述
  enabled?: boolean; // 是否启用
  timestampPath?: string; // 时间戳字段路径
  timestampFormat?: string; // 时间戳格式
  deviceIdPath?: string; // 设备ID字段路径
  fieldMappings?: HttpReceiveFieldMapping[]; // 字段映射配置（使用自定义字段名）
  dispatchConfigs?: HttpReceiveDispatchConfig[]; // 分发配置
}

export async function updateHttpReceive(
  id: string,
  data: UpdateHttpReceiveRequest,
  options?: { [key: string]: any }
) {
  return request<BaseResponse>(`${API_PREFIX}/http-receive/${id}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    data,
    ...(options || {}),
  });
}

// 删除配置
export async function deleteHttpReceive(
  id: string,
  options?: { [key: string]: any }
) {
  return request<BaseResponse>(`${API_PREFIX}/http-receive/${id}`, {
    method: "DELETE",
    ...(options || {}),
  });
}
