// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

const API_PREFIX = '/api/v1';

/** 同步单个用户到Skylark POST /sync/user/:id */
export async function syncUser(params: { id: string }, options?: { [key: string]: any }) {
  return request<API.SyncUserResponse>(`${API_PREFIX}/sync/user/${params.id}`, {
    method: 'POST',
    ...(options || {}),
  });
}

/** 查询用户同步状态 GET /sync/users/status */
export async function getUserSyncStatus(
  params: { userIds: string[] },
  options?: { [key: string]: any },
) {
  return request<API.GetUserSyncStatusResponse>(`${API_PREFIX}/sync/users/status`, {
    method: 'GET',
    params: {
      userIds: params.userIds,
    },
    ...(options || {}),
  });
}

/** 同步单个组织到Skylark POST /sync/organization/:id */
export async function syncOrganization(
  params: { id: string },
  options?: { [key: string]: any },
) {
  return request<API.SyncOrganizationResponse>(`${API_PREFIX}/sync/organization/${params.id}`, {
    method: 'POST',
    ...(options || {}),
  });
}

/** 查询组织同步状态 GET /sync/organizations/status */
export async function getOrgSyncStatus(
  params: { orgIds: string[] },
  options?: { [key: string]: any },
) {
  return request<API.GetOrgSyncStatusResponse>(`${API_PREFIX}/sync/organizations/status`, {
    method: 'GET',
    params: {
      orgIds: params.orgIds,
    },
    ...(options || {}),
  });
}

/** eventsync服务健康检查 GET /sync/health */
export async function syncHealthCheck(options?: { [key: string]: any }) {
  return request<API.SyncHealthCheckResponse>(`${API_PREFIX}/sync/health`, {
    method: 'GET',
    ...(options || {}),
  });
}

/** 绑定用户到已存在的Skylark用户 POST /sync/user/:id/bind */
export async function bindUser(
  params: { id: string },
  body: API.BindUserRequest,
  options?: { [key: string]: any },
) {
  return request<API.BindUserResponse>(`${API_PREFIX}/sync/user/${params.id}/bind`, {
    method: 'POST',
    data: body,
    ...(options || {}),
  });
}

/** 解绑用户的Skylark映射 DELETE /sync/user/:id/bind */
export async function unbindUser(params: { id: string }, options?: { [key: string]: any }) {
  return request<API.UnbindUserResponse>(`${API_PREFIX}/sync/user/${params.id}/bind`, {
    method: 'DELETE',
    ...(options || {}),
  });
}

/** 绑定组织到已存在的Skylark组织 POST /sync/organization/:id/bind */
export async function bindOrganization(
  params: { id: string },
  body: API.BindOrganizationRequest,
  options?: { [key: string]: any },
) {
  return request<API.BindOrganizationResponse>(
    `${API_PREFIX}/sync/organization/${params.id}/bind`,
    {
      method: 'POST',
      data: body,
      ...(options || {}),
    },
  );
}

/** 解绑组织的Skylark映射 DELETE /sync/organization/:id/bind */
export async function unbindOrganization(
  params: { id: string },
  options?: { [key: string]: any },
) {
  return request<API.UnbindOrganizationResponse>(
    `${API_PREFIX}/sync/organization/${params.id}/bind`,
    {
      method: 'DELETE',
      ...(options || {}),
    },
  );
}
