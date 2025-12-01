// @ts-ignore
/* eslint-disable */
import { request } from "@umijs/max";
import { getCurrentUserPermissions } from "../permission";

const API_PREFIX = "/api/v1";

/** 用户登录 POST /auth/login */
export async function login(
  body: API.LoginParams,
  options?: { [key: string]: any }
) {
  return request<API.LoginResponse>(`${API_PREFIX}/auth/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    data: body, // 字段名已经与后端一致，直接发送
    ...(options || {}),
  });
}

/** 获取当前用户信息 GET /user/current */
export async function getCurrentUser(options?: { [key: string]: any }) {
  return request<API.GetCurrentUserResponse>(`${API_PREFIX}/user/current`, {
    method: "GET",
    ...(options || {}),
  });
}

/** 根据ID获取用户详情 GET /user/:id */
export async function getUserById(
  params: { id: string },
  options?: { [key: string]: any }
) {
  return request<API.GetUserByIdResponse>(`${API_PREFIX}/user/${params.id}`, {
    method: "GET",
    ...(options || {}),
  });
}

/** 获取用户列表 GET /user/list */
export async function getUserList(
  params: API.GetUserListRequest,
  options?: { [key: string]: any }
) {
  const url = `${API_PREFIX}/user/list`;
  const result = request<API.GetUserListResponse>(url, {
    method: "GET",
    params: {
      ...params,
    },
    ...(options || {}),
  });
  return result;
}

/** 按用户ID列表批量获取用户信息 GET /user/list?ids=... */
export async function getUsersByIds(
  ids: string[],
  options?: { [key: string]: any }
) {
  if (!ids || ids.length === 0) {
    return { code: 0, data: { list: [] } };
  }
  return getUserList({ ids }, options);
}

/** 获取用户选项列表（下拉框用） GET /user/options */
export async function getUserOptions(
  params?: {
    tenantId?: string;
    keyword?: string;
    limit?: number;
  },
  options?: { [key: string]: any }
) {
  return request<{
    code: number;
    msg?: string;
    data?: {
      list: {
        id: string; // 用户UUID
        userKey: string;
        userName: string; // 与后端保持一致
        name: string;
        email: string;
        tenantId: string;
      }[];
    };
  }>(`${API_PREFIX}/user/options`, {
    method: "GET",
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** 创建新用户 POST /user */
export async function createUser(
  body: API.CreateUserRequest,
  options?: { [key: string]: any }
) {
  return request<API.CreateUserResponse>(`${API_PREFIX}/user`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    data: body,
    ...(options || {}),
  });
}

/** 更新用户信息 PUT /user/:id */
export async function updateUser(
  params: { id: string },
  body: API.UpdateUserRequest,
  options?: { [key: string]: any }
) {
  return request<API.UpdateUserResponse>(`${API_PREFIX}/user/${params.id}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    data: body,
    ...(options || {}),
  });
}

/** 删除用户 DELETE /user/:id */
export async function deleteUser(
  params: { id: string },
  options?: { [key: string]: any }
) {
  return request<API.DeleteUserResponse>(`${API_PREFIX}/user/${params.id}`, {
    method: "DELETE",
    ...(options || {}),
  });
}

/** 重置用户密码(仅限管理员使用,无需验证旧密码) POST /user/:id/reset-password */
export async function resetUserPassword(
  params: { id: string },
  body: API.ResetPasswordRequest,
  options?: { [key: string]: any }
) {
  return request<API.ResetPasswordResponse>(
    `${API_PREFIX}/user/${params.id}/reset-password`,
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

/** 更新用户状态 POST /user/:id/status */
export async function updateUserStatus(
  params: { id: string },
  body: API.UpdateUserStatusRequest,
  options?: { [key: string]: any }
) {
  return request<API.UpdateUserStatusResponse>(
    `${API_PREFIX}/user/${params.id}/status`,
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

/** 获取当前用户信息（包含完整的权限数据） */
export async function getCurrentUserWithPermissions(options?: {
  [key: string]: any;
}) {
  try {
    // 并行获取用户信息和权限信息
    const [userResponse, permissionsResponse] = await Promise.all([
      getCurrentUser(options),
      getCurrentUserPermissions(options),
    ]);

    // 如果用户信息获取成功，合并权限信息
    if (
      userResponse.code === 0 &&
      userResponse.data &&
      permissionsResponse.code === 0
    ) {
      const user = userResponse.data;
      const permissions = permissionsResponse;

      // 构建完整的用户对象
      const completeUser: API.User = {
        ...user,
        // 确保类型正确
        role: user.role as API.RoleKey,
        permissions: permissions.permissions || [],
        roles: (permissions.roles || []).map(
          (role: string) => role as API.RoleKey
        ),
        isActive: user.status === "active",
        // 组织架构信息
        organizationIds: user.organizationIds || [],
        // 其他字段保持原样
        tags: user.tags || [],
        geographic: user.geographic || {},
        phone: user.phone || "",
        title: user.title || "",
        avatar: user.avatar || "",
        signature: user.signature || "",
        country: user.country || "",
        address: user.address || "",
        createdAt: user.createdAt || "",
        updatedAt: user.updatedAt || "",
        lastLoginAt: user.lastLoginAt || "",
      };

      return {
        ...userResponse,
        data: completeUser,
      };
    }

    // 如果权限获取失败，构建基本用户对象
    if (userResponse.code === 0 && userResponse.data) {
      const user = userResponse.data;
      const basicUser: API.User = {
        ...user,
        role: user.role as API.RoleKey,
        permissions: [],
        roles: [user.role as API.RoleKey],
        isActive: user.status === "active",
        // 组织架构信息
        organizationIds: user.organizationIds || [],
        tags: user.tags || [],
        geographic: user.geographic || {},
        phone: user.phone || "",
        title: user.title || "",
        avatar: user.avatar || "",
        signature: user.signature || "",
        country: user.country || "",
        address: user.address || "",
        createdAt: user.createdAt || "",
        updatedAt: user.updatedAt || "",
        lastLoginAt: user.lastLoginAt || "",
      };

      return {
        ...userResponse,
        data: basicUser,
      };
    }

    // 如果用户信息获取失败，返回错误响应
    return userResponse;
  } catch (error) {
    // 权限获取失败时，回退到基本用户信息
    try {
      const userResponse = await getCurrentUser(options);
      if (userResponse.code === 0 && userResponse.data) {
        const user = userResponse.data;
        const basicUser: API.User = {
          ...user,
          role: user.role as API.RoleKey,
          permissions: [],
          roles: [user.role as API.RoleKey],
          isActive: user.status === "active",
          tags: user.tags || [],
          geographic: user.geographic || {},
          phone: user.phone || "",
          title: user.title || "",
          avatar: user.avatar || "",
          signature: user.signature || "",
          country: user.country || "",
          address: user.address || "",
          createdAt: user.createdAt || "",
          updatedAt: user.updatedAt || "",
          lastLoginAt: user.lastLoginAt || "",
        };

        return {
          ...userResponse,
          data: basicUser,
        };
      } else {
        // 用户信息获取失败，返回原始错误响应
        return userResponse;
      }
    } catch (fallbackError) {
      console.error("获取用户信息失败:", fallbackError);
      // 所有获取都失败，返回错误响应
      return {
        code: 500,
        msg: "获取用户信息失败",
      };
    }
  }
}

/** 上传用户头像 POST /user/avatar/upload */
export async function uploadAvatar(
  file: File,
  options?: { [key: string]: any }
) {
  const formData = new FormData();
  formData.append('file', file);
  return request<API.UploadAvatarResponse>(`${API_PREFIX}/user/avatar/upload`, {
    method: 'POST',
    data: formData,
    ...(options || {}),
  });
}

/** 修改当前用户密码(需验证旧密码) POST /user/password/change */
export async function changePassword(
  body: API.ChangePasswordRequest,
  options?: { [key: string]: any }
) {
  return request<API.ChangePasswordResponse>(`${API_PREFIX}/user/password/change`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 更新当前用户信息(用户修改自己的资料,无需user:write权限) POST /user/current/profile */
export async function updateCurrentUser(
  body: API.UpdateCurrentUserRequest,
  options?: { [key: string]: any }
) {
  return request<API.UpdateCurrentUserResponse>(`${API_PREFIX}/user/current/profile`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}
