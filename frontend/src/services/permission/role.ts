import { request } from "@umijs/max";

// 角色信息接口
export interface RoleInfo {
  roleKey: string;
  roleName: string;
  description: string;
  permissions: string[];
}

// 创建角色请求
export interface CreateRoleRequest {
  roleKey: string;
  roleName: string;
  description?: string;
  permissions: string[];
}

// 更新角色请求
export interface UpdateRoleRequest {
  roleName?: string;
  description?: string;
  permissions?: string[];
}

// 获取角色列表请求
export interface GetRoleListRequest {
  current?: number;
  pageSize?: number;
  roleKey?: string;
  roleName?: string;
}

// 权限项
export interface PermissionItem {
  resource: string;
  action: string;
  description: string;
  category: string;
}

// 为角色添加权限请求
export interface AddRolePermissionRequest {
  roleKey: string;
  resource: string;
  action: string;
  tenantKey: string;
}

// 为角色移除权限请求
export interface RemoveRolePermissionRequest {
  roleKey: string;
  resource: string;
  action: string;
  tenantKey: string;
}

/**
 * 创建角色
 */
export async function createRole(data: CreateRoleRequest) {
  return request<API.BaseResponse>("/api/v1/permission/role", {
    method: "POST",
    data,
  });
}

/**
 * 更新角色
 */
export async function updateRole(roleKey: string, data: UpdateRoleRequest) {
  return request<API.BaseResponse>(`/api/v1/permission/role/${roleKey}`, {
    method: "PUT",
    data,
  });
}

/**
 * 删除角色
 */
export async function deleteRole(roleKey: string) {
  return request<API.BaseResponse>(`/api/v1/permission/role/${roleKey}`, {
    method: "DELETE",
  });
}

/**
 * 获取角色列表
 */
export async function getRoleList(params: GetRoleListRequest = {}) {
  return request<{
    code: number;
    msg: string;
    data: RoleInfo[];
    total: number;
    current: number;
    pageSize: number;
  }>("/api/v1/permission/role/list", {
    method: "GET",
    params,
  });
}

/**
 * 获取角色详情
 */
export async function getRole(roleKey: string) {
  return request<{
    code: number;
    msg: string;
    data: RoleInfo;
  }>(`/api/v1/permission/role/${roleKey}`, {
    method: "GET",
  });
}

/**
 * 为角色添加权限
 */
export async function addRolePermission(data: AddRolePermissionRequest) {
  return request<API.BaseResponse>("/api/v1/permission/role/permission/add", {
    method: "POST",
    data,
  });
}

/**
 * 为角色移除权限
 */
export async function removeRolePermission(data: RemoveRolePermissionRequest) {
  return request<API.BaseResponse>(
    "/api/v1/permission/role/permission/remove",
    {
      method: "DELETE",
      data,
    }
  );
}

/**
 * 获取可用权限列表
 */
export async function getAvailablePermissions() {
  return request<{
    code: number;
    msg: string;
    data: PermissionItem[];
  }>("/api/v1/permission/available", {
    method: "GET",
  });
}
