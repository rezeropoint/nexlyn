/**
 * 平台配置管理API
 */
import { request } from "@umijs/max";
import { API_PREFIX } from "../constants";
import type { BaseResponse, PageParams } from "../types/common";
import type { PlatformDetail, PlatformMetadata } from "../types/platform";

// 获取平台配置列表请求
export interface GetPlatformListRequest {
  page?: number;
  pageSize?: number;
  type?: string; // 平台类型筛选
  keyword?: string; // 关键词搜索（ID或名称）
}

// 获取平台配置列表响应
export interface GetPlatformListResponse extends BaseResponse, PageParams {
  data?: {
    list: PlatformMetadata[];
  };
}

export async function getPlatformList(
  params?: GetPlatformListRequest,
  options?: { [key: string]: any }
) {
  return request<GetPlatformListResponse>(`${API_PREFIX}/platforms/list`, {
    method: "GET",
    params,
    ...(options || {}),
  });
}

// 获取平台配置详情
export interface GetPlatformResponse extends BaseResponse {
  data?: PlatformDetail;
}

export async function getPlatform(
  id: string,
  options?: { [key: string]: any }
) {
  return request<GetPlatformResponse>(`${API_PREFIX}/platforms/${id}`, {
    method: "GET",
    ...(options || {}),
  });
}

// 创建平台配置请求
export interface CreatePlatformRequest {
  // 基础字段
  id: string; // 配置ID（必填，租户内唯一）
  type: string; // 平台类型：skylark/webhook/kafka（必填）
  name: string; // 配置名称（必填）
  description?: string; // 配置描述
  enabled?: boolean; // 是否启用（默认true）

  // Skylark平台配置（type=skylark时必填）
  skylarkDomain?: string; // Skylark域名
  skylarkAuthHeader?: string; // Skylark认证头
  skylarkUserID?: number; // Skylark用户ID

  // Webhook平台配置（type=webhook时必填）
  webhookUrl?: string; // Webhook URL
  webhookHeaders?: Record<string, string>; // 自定义请求头
  webhookMethod?: string; // HTTP方法

  // Kafka平台配置（type=kafka时必填）
  kafkaBrokers?: string[]; // Kafka Brokers
  kafkaTopic?: string; // Kafka Topic
}

// 创建平台配置响应
export interface CreatePlatformResponse extends BaseResponse {
  data?: {
    type: string; // 平台类型
    id: string; // 平台配置ID
    name: string; // 配置名称
  };
}

export async function createPlatform(
  data: CreatePlatformRequest,
  options?: { [key: string]: any }
) {
  return request<CreatePlatformResponse>(`${API_PREFIX}/platforms`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    data,
    ...(options || {}),
  });
}

// 更新平台配置请求
export interface UpdatePlatformRequest {
  // 基础字段
  type: string; // 平台类型：skylark/webhook/kafka（必填，不可修改）
  name?: string; // 配置名称
  description?: string; // 配置描述
  enabled?: boolean; // 是否启用

  // Skylark平台配置（type=skylark时可选更新）
  skylarkDomain?: string; // Skylark域名
  skylarkAuthHeader?: string; // Skylark认证头
  skylarkUserID?: number; // Skylark用户ID

  // Webhook平台配置（type=webhook时可选更新）
  webhookUrl?: string; // Webhook URL
  webhookHeaders?: Record<string, string>; // 自定义请求头
  webhookMethod?: string; // HTTP方法

  // Kafka平台配置（type=kafka时可选更新）
  kafkaBrokers?: string[]; // Kafka Brokers
  kafkaTopic?: string; // Kafka Topic
}

export async function updatePlatform(
  id: string,
  data: UpdatePlatformRequest,
  options?: { [key: string]: any }
) {
  return request<BaseResponse>(`${API_PREFIX}/platforms/${id}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    data,
    ...(options || {}),
  });
}

// 删除平台配置
export async function deletePlatform(
  id: string,
  options?: { [key: string]: any }
) {
  return request<BaseResponse>(`${API_PREFIX}/platforms/${id}`, {
    method: "DELETE",
    ...(options || {}),
  });
}
