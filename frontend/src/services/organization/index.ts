// @ts-ignore
/* eslint-disable */
import { request } from "@umijs/max";

const API_PREFIX = "/api/v1";

// ===== 组织管理相关API =====

/** 获取组织树 GET /organization/tree */
export async function getOrganizationTree(
  params: API.GetOrganizationTreeRequest,
  options?: { [key: string]: any }
) {
  return request<API.GetOrganizationTreeResponse>(
    `${API_PREFIX}/organization/tree`,
    {
      method: "GET",
      params: {
        ...params,
      },
      ...(options || {}),
    }
  );
}

/** 获取组织详情 GET /organization/:id */
export async function getOrganization(
  params: API.GetOrganizationRequest,
  options?: { [key: string]: any }
) {
  return request<API.GetOrganizationResponse>(
    `${API_PREFIX}/organization/${params.id}`,
    {
      method: "GET",
      ...(options || {}),
    }
  );
}

/** 获取组织列表 GET /organization/list */
export async function getOrganizationList(
  params: API.GetOrganizationListRequest,
  options?: { [key: string]: any }
) {
  return request<API.GetOrganizationListResponse>(
    `${API_PREFIX}/organization/list`,
    {
      method: "GET",
      params: {
        ...params,
      },
      ...(options || {}),
    }
  );
}

/** 获取组织选项列表（下拉框用） GET /organization/options */
export async function getOrganizationOptions(
  params?: API.GetOrganizationOptionsRequest,
  options?: { [key: string]: any }
) {
  return request<API.GetOrganizationOptionsResponse>(
    `${API_PREFIX}/organization/options`,
    {
      method: "GET",
      params: {
        ...params,
      },
      ...(options || {}),
    }
  );
}

/** 创建组织 POST /organization */
export async function createOrganization(
  body: API.CreateOrganizationRequest,
  options?: { [key: string]: any }
) {
  return request<API.CreateOrganizationResponse>(`${API_PREFIX}/organization`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    data: body,
    ...(options || {}),
  });
}

/** 更新组织信息 PUT /organization/:id */
export async function updateOrganization(
  params: { id: string },
  body: Omit<API.UpdateOrganizationRequest, "id">,
  options?: { [key: string]: any }
) {
  return request<API.UpdateOrganizationResponse>(
    `${API_PREFIX}/organization/${params.id}`,
    {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      data: body,
      ...(options || {}),
    }
  );
}

/** 删除组织 DELETE /organization/:id */
export async function deleteOrganization(
  params: { id: string },
  body?: { forceDelete?: boolean },
  options?: { [key: string]: any }
) {
  return request<API.DeleteOrganizationResponse>(
    `${API_PREFIX}/organization/${params.id}`,
    {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
      },
      data: body,
      ...(options || {}),
    }
  );
}

/** 移动组织 POST /organization/:id/move */
export async function moveOrganization(
  params: { id: string },
  body: Omit<API.MoveOrganizationRequest, "id">,
  options?: { [key: string]: any }
) {
  return request<API.MoveOrganizationResponse>(
    `${API_PREFIX}/organization/${params.id}/move`,
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

// ===== 组织成员管理相关API =====

/** 获取组织成员 GET /organization/:id/members */
export async function getOrganizationMembers(
  params: API.GetOrganizationMembersRequest,
  options?: { [key: string]: any }
) {
  const { id, ...queryParams } = params;
  return request<API.GetOrganizationMembersResponse>(
    `${API_PREFIX}/organization/${id}/members`,
    {
      method: "GET",
      params: {
        ...queryParams,
      },
      ...(options || {}),
    }
  );
}

/** 添加组织成员 POST /organization/:id/members */
export async function addOrganizationMember(
  params: { id: string },
  body: Omit<API.AddOrganizationMemberRequest, "id">,
  options?: { [key: string]: any }
) {
  return request<API.AddOrganizationMemberResponse>(
    `${API_PREFIX}/organization/${params.id}/members`,
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

/** 移除组织成员 DELETE /organization/:orgId/members/:userId */
export async function removeOrganizationMember(
  params: API.RemoveOrganizationMemberRequest,
  options?: { [key: string]: any }
) {
  return request<API.RemoveOrganizationMemberResponse>(
    `${API_PREFIX}/organization/${params.orgId}/members/${params.userId}`,
    {
      method: "DELETE",
      ...(options || {}),
    }
  );
}

/** 更新成员关系 PUT /organization/:orgId/members/:userId */
export async function updateOrganizationMember(
  params: { orgId: string; userId: string },
  body: Omit<API.UpdateOrganizationMemberRequest, "orgId" | "userId">,
  options?: { [key: string]: any }
) {
  return request<API.UpdateOrganizationMemberResponse>(
    `${API_PREFIX}/organization/${params.orgId}/members/${params.userId}`,
    {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      data: body,
      ...(options || {}),
    }
  );
}

/** 获取用户所属组织 GET /user/:id/organizations */
export async function getUserOrganizations(
  params: API.GetUserOrganizationsRequest,
  options?: { [key: string]: any }
) {
  const { id, ...queryParams } = params;
  return request<API.GetUserOrganizationsResponse>(
    `${API_PREFIX}/user/${id}/organizations`,
    {
      method: "GET",
      params: {
        ...queryParams,
      },
      ...(options || {}),
    }
  );
}
