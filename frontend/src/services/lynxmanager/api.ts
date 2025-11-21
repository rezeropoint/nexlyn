/**
 * LynxGraph Manager API 调用函数
 * @description 信息原子类型、逻辑图配置等API接口
 */
import { request } from '@umijs/max';
import type {
  BatchQueryTagsRequest,
  BatchQueryTagsResponse,
  CreateGraphConfigRequest,
  CreateGraphConfigResponse,
  CreateInfoAtomTypeRequest,
  CreateInfoAtomTypeResponse,
  CreateTagRequest,
  CreateTagResponse,
  DeleteGraphConfigRequest,
  DeleteGraphConfigResponse,
  DeleteInfoAtomTypeRequest,
  DeleteInfoAtomTypeResponse,
  DeleteTagRequest,
  DeleteTagResponse,
  GetBlockSpecsResponse,
  GetGraphConfigListRequest,
  GetGraphConfigListResponse,
  GetGraphConfigRequest,
  GetGraphConfigResponse,
  GetInfoAtomTypeListRequest,
  GetInfoAtomTypeListResponse,
  GetInfoAtomTypeRequest,
  GetInfoAtomTypeResponse,
  GetTagResponse,
  ListTagsRequest,
  ListTagsResponse,
  UpdateGraphConfigRequest,
  UpdateGraphConfigResponse,
  UpdateInfoAtomTypeRequest,
  UpdateInfoAtomTypeResponse,
  UpdateTagRequest,
  UpdateTagResponse,
} from './types';

// API 前缀
const API_PREFIX = '/api/lynxmanager';

// ========== 信息原子类型相关接口 ==========

/**
 * 创建信息原子类型
 * @param data 创建请求参数
 * @param options 额外选项
 */
export async function createInfoAtomType(
  data: CreateInfoAtomTypeRequest,
  options?: { [key: string]: any },
) {
  return request<CreateInfoAtomTypeResponse>(`${API_PREFIX}/infoatom-type/create`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data,
    ...(options || {}),
  });
}

/**
 * 查询信息原子类型列表
 * @param params 查询参数
 * @param options 额外选项
 */
export async function listInfoAtomTypes(
  params?: GetInfoAtomTypeListRequest,
  options?: { [key: string]: any },
) {
  return request<GetInfoAtomTypeListResponse>(`${API_PREFIX}/infoatom-type/list`, {
    method: 'GET',
    params,
    ...(options || {}),
  });
}

/**
 * 获取信息原子类型详情
 * @param id 信息原子类型ID
 * @param options 额外选项
 */
export async function getInfoAtomType(id: string, options?: { [key: string]: any }) {
  return request<GetInfoAtomTypeResponse>(`${API_PREFIX}/infoatom-type/get/${id}`, {
    method: 'GET',
    ...(options || {}),
  });
}

/**
 * 更新信息原子类型
 * @param id 信息原子类型ID
 * @param data 更新请求参数
 * @param options 额外选项
 */
export async function updateInfoAtomType(
  id: string,
  data: Omit<UpdateInfoAtomTypeRequest, 'id'>,
  options?: { [key: string]: any },
) {
  return request<UpdateInfoAtomTypeResponse>(`${API_PREFIX}/infoatom-type/update/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    data,
    ...(options || {}),
  });
}

/**
 * 删除信息原子类型
 * @param id 信息原子类型ID
 * @param options 额外选项
 */
export async function deleteInfoAtomType(id: string, options?: { [key: string]: any }) {
  return request<DeleteInfoAtomTypeResponse>(`${API_PREFIX}/infoatom-type/delete/${id}`, {
    method: 'DELETE',
    ...(options || {}),
  });
}

// ========== 逻辑图配置相关接口 ==========

/**
 * 创建逻辑图配置
 * @param data 创建请求参数
 * @param options 额外选项
 */
export async function createGraphConfig(
  data: CreateGraphConfigRequest,
  options?: { [key: string]: any },
) {
  return request<CreateGraphConfigResponse>(`${API_PREFIX}/graph/create`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data,
    ...(options || {}),
  });
}

/**
 * 查询逻辑图配置列表
 * @param params 查询参数
 * @param options 额外选项
 */
export async function listGraphConfigs(
  params?: GetGraphConfigListRequest,
  options?: { [key: string]: any },
) {
  return request<GetGraphConfigListResponse>(`${API_PREFIX}/graph/list`, {
    method: 'GET',
    params,
    ...(options || {}),
  });
}

/**
 * 获取逻辑图配置详情
 * @param id 逻辑图ID
 * @param options 额外选项
 */
export async function getGraphConfig(id: string, options?: { [key: string]: any }) {
  return request<GetGraphConfigResponse>(`${API_PREFIX}/graph/get/${id}`, {
    method: 'GET',
    ...(options || {}),
  });
}

/**
 * 更新逻辑图配置
 * @param id 逻辑图ID
 * @param data 更新请求参数
 * @param options 额外选项
 */
export async function updateGraphConfig(
  id: string,
  data: Omit<UpdateGraphConfigRequest, 'id'>,
  options?: { [key: string]: any },
) {
  return request<UpdateGraphConfigResponse>(`${API_PREFIX}/graph/update/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    data,
    ...(options || {}),
  });
}

/**
 * 删除逻辑图配置
 * @param id 逻辑图ID
 * @param options 额外选项
 */
export async function deleteGraphConfig(id: string, options?: { [key: string]: any }) {
  return request<DeleteGraphConfigResponse>(`${API_PREFIX}/graph/delete/${id}`, {
    method: 'DELETE',
    ...(options || {}),
  });
}

// ========== 逻辑积木规格相关接口 ==========

/**
 * 获取逻辑积木规格列表
 * @param options 额外选项
 */
export async function getBlockSpecs(options?: { [key: string]: any }) {
  return request<GetBlockSpecsResponse>(`${API_PREFIX}/block/specs`, {
    method: 'GET',
    ...(options || {}),
  });
}

// ========== 标签管理相关接口 ==========

/**
 * 创建标签
 * @param data 创建请求参数
 * @param options 额外选项
 */
export async function createTag(
  data: CreateTagRequest,
  options?: { [key: string]: any },
) {
  return request<CreateTagResponse>(`${API_PREFIX}/tags`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data,
    ...(options || {}),
  });
}

/**
 * 查询标签列表
 * @param params 查询参数
 * @param options 额外选项
 */
export async function listTags(
  params?: ListTagsRequest,
  options?: { [key: string]: any },
) {
  return request<ListTagsResponse>(`${API_PREFIX}/tags`, {
    method: 'GET',
    params,
    ...(options || {}),
  });
}

/**
 * 获取标签详情
 * @param id 标签ID
 * @param options 额外选项
 */
export async function getTag(id: string, options?: { [key: string]: any }) {
  return request<GetTagResponse>(`${API_PREFIX}/tags/${id}`, {
    method: 'GET',
    ...(options || {}),
  });
}

/**
 * 更新标签
 * @param id 标签ID
 * @param data 更新请求参数
 * @param options 额外选项
 */
export async function updateTag(
  id: string,
  data: Omit<UpdateTagRequest, 'id'>,
  options?: { [key: string]: any },
) {
  return request<UpdateTagResponse>(`${API_PREFIX}/tags/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    data,
    ...(options || {}),
  });
}

/**
 * 删除标签
 * @param id 标签ID
 * @param options 额外选项
 */
export async function deleteTag(id: string, options?: { [key: string]: any }) {
  return request<DeleteTagResponse>(`${API_PREFIX}/tags/${id}`, {
    method: 'DELETE',
    ...(options || {}),
  });
}

/**
 * 批量查询标签
 * @param data 批量查询请求参数
 * @param options 额外选项
 */
export async function batchQueryTags(
  data: BatchQueryTagsRequest,
  options?: { [key: string]: any },
) {
  return request<BatchQueryTagsResponse>(`${API_PREFIX}/tags/batch-query`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data,
    ...(options || {}),
  });
}
