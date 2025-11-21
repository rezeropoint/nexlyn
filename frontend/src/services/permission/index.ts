// @ts-ignore
/* eslint-disable */
import { request } from "@umijs/max";

const API_PREFIX = "/api/v1";

/** 检查用户权限 POST /permission/check */
export async function checkPermission(
  body: API.CheckPermissionRequest,
  options?: { [key: string]: any }
) {
  return request<API.CheckPermissionResponse>(
    `${API_PREFIX}/permission/check`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      data: body,
      ...(options || {}),
    }
  );
}

/** 获取当前用户权限 GET /permission/current */
export async function getCurrentUserPermissions(options?: {
  [key: string]: any;
}) {
  return request<API.GetUserPermissionsResponse>(
    `${API_PREFIX}/permission/current`,
    {
      method: "GET",
      ...(options || {}),
    }
  );
}

/** 获取权限规则列表 GET /permission/rules */
export async function getPermissionRules(
  params: API.GetPermissionRulesRequest,
  options?: { [key: string]: any }
) {
  return request<API.GetPermissionRulesResponse>(
    `${API_PREFIX}/permission/rules`,
    {
      method: "GET",
      params: {
        ...params,
      },
      ...(options || {}),
    }
  );
}

/** 添加权限规则 POST /permission/add */
export async function addPermission(
  body: API.AddPermissionRequest,
  options?: { [key: string]: any }
) {
  return request<API.BaseResponse>(`${API_PREFIX}/permission/add`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    data: body,
    ...(options || {}),
  });
}

/** 移除权限规则 DELETE /permission/remove */
export async function removePermission(
  body: API.RemovePermissionRequest,
  options?: { [key: string]: any }
) {
  return request<API.BaseResponse>(`${API_PREFIX}/permission/remove`, {
    method: "DELETE",
    headers: {
      "Content-Type": "application/json",
    },
    data: body,
    ...(options || {}),
  });
}

/** 为用户添加角色 POST /permission/role/add */
export async function addRoleForUser(
  body: API.AddRoleRequest,
  options?: { [key: string]: any }
) {
  return request<API.BaseResponse>(`${API_PREFIX}/permission/role/add`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    data: body,
    ...(options || {}),
  });
}

/** 为用户移除角色 DELETE /permission/role/remove */
export async function removeRoleForUser(
  body: API.RemoveRoleRequest,
  options?: { [key: string]: any }
) {
  return request<API.BaseResponse>(`${API_PREFIX}/permission/role/remove`, {
    method: "DELETE",
    headers: {
      "Content-Type": "application/json",
    },
    data: body,
    ...(options || {}),
  });
}

/** 获取可用权限列表 GET /permission/available */
export async function getAvailablePermissions(options?: {
  [key: string]: any;
}) {
  return request<API.GetAvailablePermissionsResponse>(
    `${API_PREFIX}/permission/available`,
    {
      method: "GET",
      ...(options || {}),
    }
  );
}
