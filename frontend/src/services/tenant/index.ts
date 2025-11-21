// @ts-ignore
/* eslint-disable */
import { request } from "@umijs/max";

const API_PREFIX = "/api/v1";

// 导出API类型以保持向后兼容
export type Tenant = API.Tenant;
export type TenantListItem = API.TenantListItem;
export type GetTenantListRequest = API.GetTenantListRequest;
export type GetTenantListResponse = API.GetTenantListResponse;
export type CreateTenantRequest = API.CreateTenantRequest;
export type CreateTenantResponse = API.CreateTenantResponse;
export type UpdateTenantRequest = API.UpdateTenantRequest;
export type UpdateTenantResponse = API.UpdateTenantResponse;
export type UpdateTenantStatusRequest = API.UpdateTenantStatusRequest;
export type UpdateTenantStatusResponse = API.UpdateTenantStatusResponse;
export type GetTenantResponse = API.GetTenantResponse;
export type DeleteTenantResponse = API.DeleteTenantResponse;

/** 获取租户列表 GET /api/v1/tenant/list */
export async function getTenantList(params?: GetTenantListRequest) {
  return request<GetTenantListResponse>(`${API_PREFIX}/tenant/list`, {
    method: "GET",
    params,
  });
}

/** 获取租户选项列表（下拉框用） GET /api/v1/tenant/options */
export async function getTenantOptions(params?: {
  keyword?: string;
  limit?: number;
  status?: string;
}) {
  return request<{
    code: number;
    msg?: string;
    data?: {
      list: API.TenantOption[];
    };
  }>(`${API_PREFIX}/tenant/options`, {
    method: "GET",
    params,
  });
}

/** 创建租户 POST /api/v1/tenant */
export async function createTenant(body: CreateTenantRequest) {
  return request<CreateTenantResponse>(`${API_PREFIX}/tenant`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    data: body,
  });
}

/** 获取租户详情 GET /api/v1/tenant/:id */
export async function getTenant(params: { id: string }) {
  return request<GetTenantResponse>(`${API_PREFIX}/tenant/${params.id}`, {
    method: "GET",
  });
}

/** 更新租户信息 PUT /api/v1/tenant/:id */
export async function updateTenant(
  params: { id: string },
  body: UpdateTenantRequest
) {
  return request<UpdateTenantResponse>(`${API_PREFIX}/tenant/${params.id}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    data: body,
  });
}

/** 更新租户状态 POST /api/v1/tenant/:id/status */
export async function updateTenantStatus(
  params: { id: string },
  body: UpdateTenantStatusRequest
) {
  return request<UpdateTenantStatusResponse>(
    `${API_PREFIX}/tenant/${params.id}/status`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      data: body,
    }
  );
}

/** 删除租户 DELETE /api/v1/tenant/:id */
export async function deleteTenant(params: { id: string }) {
  return request<DeleteTenantResponse>(`${API_PREFIX}/tenant/${params.id}`, {
    method: "DELETE",
  });
}
