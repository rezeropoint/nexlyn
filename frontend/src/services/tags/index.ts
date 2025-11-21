// @ts-ignore
/* eslint-disable */
import { request } from "@umijs/max";

const API_PREFIX = "/api/v1";

/** 获取标签选项 GET /tags/options */
export async function getTagOptions(
  params: API.GetTagOptionsRequest,
  options?: { [key: string]: any }
) {
  return request<API.GetTagOptionsResponse>(`${API_PREFIX}/tags/options`, {
    method: "GET",
    params,
    ...(options || {}),
  });
}

/** 获取标签列表 GET /tags/list */
export async function getTagList(
  params: API.GetTagListRequest,
  options?: { [key: string]: any }
) {
  return request<API.GetTagListResponse>(`${API_PREFIX}/tags/list`, {
    method: "GET",
    params,
    ...(options || {}),
  });
}

/** 获取标签详情 GET /tags/:id */
export async function getTag(
  params: { id: string },
  options?: { [key: string]: any }
) {
  return request<API.GetTagResponse>(`${API_PREFIX}/tags/${params.id}`, {
    method: "GET",
    ...(options || {}),
  });
}

/** 创建标签 POST /tags */
export async function createTag(
  body: API.CreateTagRequest,
  options?: { [key: string]: any }
) {
  return request<API.CreateTagResponse>(`${API_PREFIX}/tags`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    data: body,
    ...(options || {}),
  });
}

/** 更新标签 PUT /tags/:id */
export async function updateTag(
  params: { id: string },
  body: API.UpdateTagRequest,
  options?: { [key: string]: any }
) {
  return request<API.UpdateTagResponse>(`${API_PREFIX}/tags/${params.id}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    data: body,
    ...(options || {}),
  });
}

/** 删除标签 DELETE /tags/:id */
export async function deleteTag(
  params: { id: string },
  options?: { [key: string]: any }
) {
  return request<API.DeleteTagResponse>(`${API_PREFIX}/tags/${params.id}`, {
    method: "DELETE",
    ...(options || {}),
  });
}
