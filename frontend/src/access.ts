/**
 * @see https://umijs.org/docs/max/access#access
 * 基于Casbin的细粒度权限控制系统
 *
 * CasbinX v0.7.11+ 安全模型：
 * - 所有权限操作都会在后端验证操作者权限（从JWT token获取operatorKey）
 * - 权限分为两级：普通权限（可变更）和系统权限（不可变更）
 * - 包含系统权限的角色受到分配/移除保护，除租户初始化场景外
 */

// ===== 系统权限常量定义 =====
// 这些权限被标识为系统权限，在后端完全不可变更
const SYSTEM_PERMISSIONS = [
  "tenant:write",
  "tenant:delete",
  "system:write",
  "system:read",
  "user:write",
  "user:read",
  "permission:write",
  "permission:read",
  "role:write",
  "role:read",
  "role:delete",
];

export default function access(
  initialState: { currentUser?: API.CurrentUser } | undefined
) {
  const { currentUser } = initialState ?? {};

  // ===== 核心权限检查函数 =====

  /**
   * 检查用户是否有特定资源的特定操作权限
   * @param resource 资源名称，如 "user", "graph", "infoatom"
   * @param action 操作名称，如 "read", "write", "delete", "admin"
   * @returns 是否有权限
   */
  const hasPermission = (
    resource: string,
    action: string = "read"
  ): boolean => {
    if (!currentUser?.isActive) {
      return false;
    }

    // 超级管理员拥有所有权限 - 兼容roles数组和role字段
    const isSuperAdmin =
      currentUser.roles?.includes("super_admin") ||
      currentUser.role === "super_admin";
    if (isSuperAdmin) {
      return true;
    }

    // 检查具体权限 - 支持 "*" 通配符权限匹配
    const requiredPermission = `${resource}:${action}`;
    const wildcardPermission = `${resource}:*`;

    const hasSpecificPermission =
      currentUser.permissions?.includes(requiredPermission) || false;
    const hasWildcardPermission =
      currentUser.permissions?.includes(wildcardPermission) || false;

    return hasSpecificPermission || hasWildcardPermission;
  };

  /**
   * 检查用户是否有资源的任何权限
   * @param resource 资源名称
   * @returns 是否有任何权限
   */
  const hasAnyPermission = (resource: string): boolean => {
    if (!currentUser?.isActive) {
      return false;
    }

    // 超级管理员拥有所有权限 - 兼容roles数组和role字段
    const isSuperAdmin =
      currentUser.roles?.includes("super_admin") ||
      currentUser.role === "super_admin";
    if (isSuperAdmin) {
      return true;
    }

    return (
      currentUser.permissions?.some((permission) =>
        permission.startsWith(`${resource}:`)
      ) || false
    );
  };

  /**
   * 检查指定权限是否为系统权限（不可变更）
   * @param resource 资源名称
   * @param action 操作名称
   * @returns 是否为系统权限
   */
  const isSystemPermission = (resource: string, action: string): boolean => {
    const permissionKey = `${resource}:${action}`;
    return SYSTEM_PERMISSIONS.includes(permissionKey);
  };

  // ===== 用户状态检查 =====

  /** 检查是否为超级管理员 */
  const isSuperAdmin = (): boolean => {
    return (
      currentUser?.roles?.includes("super_admin") ||
      currentUser?.role === "super_admin" ||
      false
    );
  };

  /** 检查是否为租户管理员 */
  const isTenantAdmin = (): boolean => {
    return (
      currentUser?.roles?.includes("admin") ||
      currentUser?.role === "admin" ||
      false
    );
  };

  /** 检查是否为管理员（租户管理员或更高） */
  const isAdmin = (): boolean => {
    return (
      currentUser?.roles?.some(
        (role) => role === "admin" || role === "super_admin"
      ) ?? false
    );
  };

  /** 检查是否为活跃用户 */
  const isActive = (): boolean => {
    return currentUser?.isActive ?? false;
  };

  /**
   * 检查是否可以分配角色给用户
   * 需要用户管理权限 (user:write)，后端还会检查角色是否包含系统权限
   */
  const canAssignRole = (): boolean => {
    return hasPermission("user", "write");
  };

  /**
   * 检查是否可以移除用户的角色
   * 需要用户管理权限 (user:write)，后端还会检查角色是否包含系统权限
   */
  const canRemoveRole = (): boolean => {
    return hasPermission("user", "write");
  };

  /** 检查当前用户是否可以操作目标用户（级别检查） */
  const canOperateUser = (targetUser: API.User | API.UserBrief): boolean => {
    if (!currentUser?.isActive) return false;

    // 超级管理员可以操作所有用户
    if (isSuperAdmin()) return true;

    // 用户可以操作自己（修改自己的信息）
    let targetUserKey: string;
    if ("userKey" in targetUser) {
      targetUserKey = targetUser.userKey;
    } else {
      targetUserKey = (targetUser as any).user_key;
    }
    if (currentUser.userKey === targetUserKey) {
      return true;
    }

    // 获取目标用户的角色
    // 如果是User类型且有roles数组，使用roles；否则使用role字段
    let targetRoles: string[] = [];
    if ("roles" in targetUser && targetUser.roles) {
      targetRoles = targetUser.roles;
    } else if (targetUser.role) {
      targetRoles = [targetUser.role as string];
    }

    // 普通管理员不能操作超级管理员
    if (isTenantAdmin() && targetRoles.includes("super_admin")) {
      return false;
    }

    // 普通用户不能操作管理员级别的用户
    // 检查当前用户的所有角色，判断是否为普通用户（没有admin或super_admin角色）
    const currentUserRoles =
      currentUser.roles || (currentUser.role ? [currentUser.role] : []);
    const isCurrentUserOnlyUser = !currentUserRoles.some(
      (role) => role === "admin" || role === "super_admin"
    );
    if (
      isCurrentUserOnlyUser &&
      (targetRoles.includes("admin") || targetRoles.includes("super_admin"))
    ) {
      return false;
    }

    return true;
  };

  // ===== 资源特定权限检查 =====

  // 用户管理
  const canManageUsers = () => hasAnyPermission("user");
  const canCreateUsers = () => hasPermission("user", "write");
  const canUpdateUsers = () => hasPermission("user", "write");
  const canDeleteUsers = () => hasPermission("user", "delete");
  const canReadUsers = () => hasPermission("user", "read");

  // 租户管理（仅超级管理员）
  const canManageTenants = () => hasAnyPermission("tenant");
  const canCreateTenants = () => hasPermission("tenant", "write");
  const canUpdateTenants = () => hasPermission("tenant", "write");
  const canDeleteTenants = () => hasPermission("tenant", "delete");
  const canReadTenants = () => hasPermission("tenant", "read");

  // 租户标签应用管理（仅超级管理员）
  const canManageTenantTags = () => hasAnyPermission("tag_tenant");
  const canReadTenantTags = () => hasPermission("tag_tenant", "read");
  const canUpdateTenantTags = () => hasPermission("tag_tenant", "write");
  const canDeleteTenantTags = () => hasPermission("tag_tenant", "delete");

  // 用户标签应用管理（租户管理员及以上）
  const canManageUserTags = () => hasAnyPermission("tag_user");
  const canReadUserTags = () => hasPermission("tag_user", "read");
  const canUpdateUserTags = () => hasPermission("tag_user", "write");
  const canDeleteUserTags = () => hasPermission("tag_user", "delete");

  // 标签管理 - 基于作用域的权限检查
  const canManageTags = () =>
    hasAnyPermission("tag_user") || hasAnyPermission("tag_tenant");
  const canCreateTags = () =>
    hasPermission("tag_user", "write") || hasPermission("tag_tenant", "write");
  const canUpdateTags = () =>
    hasPermission("tag_user", "write") || hasPermission("tag_tenant", "write");
  const canDeleteTags = () =>
    hasPermission("tag_user", "delete") ||
    hasPermission("tag_tenant", "delete");
  const canReadTags = () =>
    hasPermission("tag_user", "read") || hasPermission("tag_tenant", "read");

  // 权限管理
  const canManagePermissions = () => hasAnyPermission("permission");
  const canReadPermissions = () => hasPermission("permission", "read");
  const canAddPermissions = () => hasPermission("permission", "write");
  const canRemovePermissions = () => hasPermission("permission", "delete");
  const canAddRoles = () => hasPermission("user", "write"); // 角色分配需要用户管理权限
  const canRemoveRoles = () => hasPermission("user", "write"); // 角色移除需要用户管理权限
  const canCheckPermissions = () => hasPermission("permission", "read");

  // 角色权限管理 - 修正为精确的权限检查
  const canManageRolePermissions = () => hasAnyPermission("role");
  const canCreateRoles = () => hasPermission("role", "write");
  const canUpdateRoles = () => hasPermission("role", "write");
  const canDeleteRoles = () => hasPermission("role", "delete");
  const canConfigureRolePermissions = () => hasPermission("role", "write");

  // 设备管理权限
  const canBindDevices = () => {
    // 超级管理员和租户管理员可以绑定设备
    return (
      isSuperAdmin() || isTenantAdmin() || hasPermission("device", "write")
    );
  };
  const canManageDeviceTags = () => {
    // 超级管理员和租户管理员可以管理设备标签
    return isSuperAdmin() || isTenantAdmin() || hasPermission("tag", "write");
  };

  // GB28181流媒体管理权限
  const canAccessMediaManagement = () => {
    return hasPermission("gb28181_stream", "read");
  };

  // ===== 物联管理权限 =====
  const canAccessIoTDevice = () => hasPermission("iot_device", "read");
  const canAccessIoTTemplate = () => hasPermission("iot_template", "read");
  const canAccessIoTTag = () => hasPermission("iot_tag", "read");
  const canAccessIoTPlatform = () => hasPermission("iot_platform", "read");
  const canAccessIoTHttpReceive = () => hasPermission("iot_http_receive", "read");

  // ===== 视频管理权限 =====
  const canAccessGB28181Device = () => hasPermission("gb28181_device", "read");
  const canAccessGB28181SplitScreen = () =>
    hasPermission("gb28181_stream", "read");
  const canAccessGB28181Recording = () =>
    hasPermission("gb28181_stream", "read");

  // ===== 逻辑引擎权限 =====
  const canAccessLynxOverview = () => hasPermission("lynx_graphconfig", "read");
  const canAccessLynxTag = () => hasPermission("lynx_tag", "read");
  const canAccessLynxInfoAtomType = () =>
    hasPermission("lynx_infoatomtype", "read");
  const canAccessLynxGraphConfig = () =>
    hasPermission("lynx_graphconfig", "read");

  // ===== 事件管理权限 =====
  const canAccessEventData = () => hasPermission("event_data", "read");
  const canAccessEventConfig = () => hasPermission("event_config", "read");
  const canAccessOrgMapping = () => hasPermission("org_mapping", "read");
  const canAccessSkylarkPlatform = () =>
    hasPermission("skylark_platform", "read");

  // ===== 模块级权限检查（用于一级菜单）=====
  // 物联管理模块：任一子页面有权限则显示
  const canAccessIoTModule = () =>
    hasAnyPermission("iot_device") ||
    hasAnyPermission("iot_template") ||
    hasAnyPermission("iot_tag") ||
    hasAnyPermission("iot_platform") ||
    hasAnyPermission("iot_http_receive");

  // 视频管理模块
  const canAccessVideoModule = () =>
    hasAnyPermission("gb28181_device") ||
    hasAnyPermission("gb28181_stream") ||
    hasAnyPermission("gb28181_tag");

  // 逻辑引擎模块
  const canAccessLynxModule = () =>
    hasAnyPermission("lynx_graphconfig") ||
    hasAnyPermission("lynx_tag") ||
    hasAnyPermission("lynx_infoatomtype");

  // 事件管理模块
  const canAccessEventModule = () =>
    hasAnyPermission("event_data") ||
    hasAnyPermission("event_config") ||
    hasAnyPermission("org_mapping") ||
    hasAnyPermission("skylark_platform");

  // 系统管理模块
  const canAccessSystemModule = () =>
    hasAnyPermission("user") ||
    hasAnyPermission("permission") ||
    hasAnyPermission("tenant") ||
    hasAnyPermission("tag_user") ||
    hasAnyPermission("tag_tenant") ||
    hasAnyPermission("organization");

  // 组织管理权限
  const canManageOrganizations = () => hasAnyPermission("organization");
  const canCreateOrganizations = () => hasPermission("organization", "write");
  const canUpdateOrganizations = () => hasPermission("organization", "write");
  const canDeleteOrganizations = () => hasPermission("organization", "delete");
  const canReadOrganizations = () => hasPermission("organization", "read");
  const canMoveOrganizations = () => hasPermission("organization", "write");
  const canManageOrganizationMembers = () =>
    hasPermission("organization", "write");

  // ===== 返回权限检查对象 =====

  return {
    // 核心权限检查
    hasPermission,
    hasAnyPermission,
    isSystemPermission,

    // 用户状态
    isSuperAdmin,
    isTenantAdmin,
    isAdmin,
    isActive,
    canOperateUser,
    canAssignRole,
    canRemoveRole,

    // 资源特定权限
    canManageUsers,
    canCreateUsers,
    canUpdateUsers,
    canDeleteUsers,
    canReadUsers,

    canManageTenants,
    canCreateTenants,
    canUpdateTenants,
    canDeleteTenants,
    canReadTenants,

    // 租户标签权限
    canManageTenantTags,
    canReadTenantTags,
    canUpdateTenantTags,
    canDeleteTenantTags,

    // 用户标签权限
    canManageUserTags,
    canReadUserTags,
    canUpdateUserTags,
    canDeleteUserTags,

    canManageTags,
    canCreateTags,
    canUpdateTags,
    canDeleteTags,
    canReadTags,

    canManagePermissions,
    canReadPermissions,
    canAddPermissions,
    canRemovePermissions,
    canAddRoles,
    canRemoveRoles,
    canCheckPermissions,

    // 角色权限管理
    canManageRolePermissions,
    canCreateRoles,
    canUpdateRoles,
    canDeleteRoles,
    canConfigureRolePermissions,

    // 设备管理权限（用于路由访问控制）
    canBindDevices,
    canManageDeviceTags,

    // GB28181流媒体管理权限
    canAccessMediaManagement,

    // 物联管理权限
    canAccessIoTDevice,
    canAccessIoTTemplate,
    canAccessIoTTag,
    canAccessIoTPlatform,
    canAccessIoTHttpReceive,

    // 视频管理权限（补充）
    canAccessGB28181Device,
    canAccessGB28181SplitScreen,
    canAccessGB28181Recording,

    // 逻辑引擎权限
    canAccessLynxOverview,
    canAccessLynxTag,
    canAccessLynxInfoAtomType,
    canAccessLynxGraphConfig,

    // 事件管理权限
    canAccessEventData,
    canAccessEventConfig,
    canAccessOrgMapping,
    canAccessSkylarkPlatform,

    // 模块级权限（用于一级菜单）
    canAccessIoTModule,
    canAccessVideoModule,
    canAccessLynxModule,
    canAccessEventModule,
    canAccessSystemModule,

    // 组织管理权限
    canManageOrganizations,
    canCreateOrganizations,
    canUpdateOrganizations,
    canDeleteOrganizations,
    canReadOrganizations,
    canMoveOrganizations,
    canManageOrganizationMembers,

    // 兼容性方法
    canAdmin: isAdmin(),
  };
}
