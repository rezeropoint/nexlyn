/**
 * 设备标签管理API
 */
import { request } from "@umijs/max";
import { API_PREFIX } from "../constants";
import type { BaseResponse, PageParams } from "../types/common";
import type { DeviceTagSummary } from "../types/device";
import type { DeviceTag } from "../types/tag";

// 创建标签请求
export interface CreateTagRequest {
  name: string; // 标签名称(必填)
  color?: string; // 标签颜色
  icon?: string; // 标签图标
  description?: string; // 标签描述
}

// 创建标签响应
export interface CreateTagResponse extends BaseResponse {
  data?: {
    id: string; // 标签ID
    name: string; // 标签名称
  };
}

export async function createTag(
  data: CreateTagRequest,
  options?: { [key: string]: any }
) {
  return request<CreateTagResponse>(`${API_PREFIX}/device-tags`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    data,
    ...(options || {}),
  });
}

// 获取标签详情
export interface GetTagResponse extends BaseResponse {
  data?: DeviceTag;
}

export async function getTag(tagId: string, options?: { [key: string]: any }) {
  return request<GetTagResponse>(`${API_PREFIX}/device-tags/${tagId}`, {
    method: "GET",
    ...(options || {}),
  });
}

// 获取标签列表请求
export interface ListTagsRequest {
  page?: number;
  pageSize?: number;
  name?: string; // 名称模糊查询
  keyword?: string; // 关键词搜索(名称、描述)
}

// 获取标签列表响应
export interface ListTagsResponse extends BaseResponse, PageParams {
  data?: {
    list: DeviceTagSummary[];
  };
}

export async function listTags(
  params?: ListTagsRequest,
  options?: { [key: string]: any }
) {
  return request<ListTagsResponse>(`${API_PREFIX}/device-tags/list`, {
    method: "GET",
    params,
    ...(options || {}),
  });
}

// 更新标签请求
export interface UpdateTagRequest {
  name?: string; // 标签名称
  color?: string; // 标签颜色
  icon?: string; // 标签图标
  description?: string; // 标签描述
}

export async function updateTag(
  tagId: string,
  data: UpdateTagRequest,
  options?: { [key: string]: any }
) {
  return request<BaseResponse>(`${API_PREFIX}/device-tags/${tagId}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    data,
    ...(options || {}),
  });
}

// 删除标签
export async function deleteTag(
  tagId: string,
  options?: { [key: string]: any }
) {
  return request<BaseResponse>(`${API_PREFIX}/device-tags/${tagId}`, {
    method: "DELETE",
    ...(options || {}),
  });
}
