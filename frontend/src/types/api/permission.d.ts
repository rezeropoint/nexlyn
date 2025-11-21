// 权限管理相关类型定义

declare namespace API {
  /** 权限资源 */
  type PermissionResource = string; // e.g., "user:read", "graph:write"

  /** 权限动作 */
  type PermissionAction = string; // e.g., "read", "write", "delete", "admin"

  /** 角色类型 */
  type RoleKey = "super_admin" | "admin" | "user";

  /** 权限检查请求 */
  type CheckPermissionRequest = {
    userKey: string;
    tenantKey: string;
    resource: PermissionResource;
    action: PermissionAction;
  };

  /** 权限检查响应 */
  type CheckPermissionResponse = BaseResponse & {
    hasPermission: boolean;
  };

  /** 添加权限请求 */
  type AddPermissionRequest = {
    userKey: string;
    tenantKey: string;
    resource: PermissionResource;
    action: PermissionAction;
  };

  /** 移除权限请求 */
  type RemovePermissionRequest = {
    userKey: string;
    tenantKey: string;
    resource: PermissionResource;
    action: PermissionAction;
  };

  /** 添加角色请求 */
  type AddRoleRequest = {
    userKey: string;
    roleKey: RoleKey;
    tenantKey: string;
  };

  /** 移除角色请求 */
  type RemoveRoleRequest = {
    userKey: string;
    roleKey: RoleKey;
    tenantKey: string;
  };

  /** 权限规则信息 (与后端API完全匹配) */
  type PermissionRule = {
    ptype: "p" | "g"; // p: permission rule, g: role rule
    userKey: string;
    tenantKey: string;
    resource: PermissionResource;
    action: PermissionAction;
    roleKey?: string; // 角色键 (用于角色关联规则)
  };

  /** 获取权限规则请求 */
  type GetPermissionRulesRequest = PageParams & {
    userKey?: string;
    tenantKey?: string;
    resource?: string;
    action?: string;
    ptype?: "p" | "g" | ""; // 规则类型筛选
  };

  /** 获取权限规则响应 */
  type GetPermissionRulesResponse = BaseResponse &
    PageParams & {
      data: PermissionRule[];
    };

  /** 获取用户权限响应 */
  type GetUserPermissionsResponse = BaseResponse & {
    permissions: string[];
    roles: string[];
  };

  /** 用户权限信息 */
  type UserPermissions = {
    permissions: PermissionResource[];
    roles: RoleKey[];
    isActive: boolean; // 用户是否活跃
  };

  /** 权限状态 */
  type PermissionState = {
    loading: boolean;
    error?: string;
    lastUpdated?: number;
  };

  /** 权限上下文 */
  type PermissionContext = {
    userKey: string;
    tenantKey: string;
    roles: RoleKey[];
    permissions: PermissionResource[];
    isSuperAdmin: boolean;
    isTenantAdmin: boolean;
    isActive: boolean;
  };

  // ===== 角色权限管理相关类型 =====

  /** 角色信息 */
  type RoleInfo = {
    roleKey: string;
    roleName: string;
    description: string;
    permissions: string[];
  };

  /** 创建角色请求 */
  type CreateRoleRequest = {
    roleKey: string;
    roleName: string;
    description?: string;
    permissions: string[];
  };

  /** 更新角色请求 */
  type UpdateRoleRequest = {
    roleKey: string;
    roleName?: string;
    description?: string;
    permissions?: string[];
  };

  /** 删除角色请求 */
  type DeleteRoleRequest = {
    roleKey: string;
  };

  /** 获取角色列表请求 */
  type GetRoleListRequest = {
    roleKey?: string;
    roleName?: string;
  } & PageParamsRequest;

  /** 获取角色列表响应 */
  type GetRoleListResponse = BaseResponse & {
    data: RoleInfo[];
  } & PageParams;

  /** 获取角色详情请求 */
  type GetRoleRequest = {
    roleKey: string;
  };

  /** 获取角色详情响应 */
  type GetRoleResponse = BaseResponse & {
    data: RoleInfo;
  };

  /** 为角色添加权限请求 */
  type AddRolePermissionRequest = {
    roleKey: string;
    resource: string;
    action: string;
    tenantKey: string;
  };

  /** 为角色移除权限请求 */
  type RemoveRolePermissionRequest = {
    roleKey: string;
    resource: string;
    action: string;
    tenantKey: string;
  };

  /** 权限项 */
  type PermissionItem = {
    resource: string;
    action: string;
    description: string;
    category: string;
  };

  /** 获取可用权限列表响应 */
  type GetAvailablePermissionsResponse = BaseResponse & {
    data: PermissionItem[];
  };
}
