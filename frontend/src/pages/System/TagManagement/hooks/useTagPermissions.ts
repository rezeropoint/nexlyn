import { useAccess } from "@umijs/max";
import type { UseTagPermissionsReturn } from "../types";

/**
 * 标签权限管理Hook
 * 区分用户标签权限和租户标签权限
 */
export const useTagPermissions = (): UseTagPermissionsReturn => {
  const access = useAccess();

  return {
    // 综合标签权限（用户标签或租户标签）
    canRead: access.canReadTags?.() || false,
    canCreate: access.canCreateTags?.() || false,
    canUpdate: access.canUpdateTags?.() || false,
    canDelete: access.canDeleteTags?.() || false,

    // 用户标签权限（租户管理员及以上）
    canManageUserTags: access.canManageUserTags?.() || false,
    canReadUserTags: access.canReadUserTags?.() || false,
    canUpdateUserTags: access.canUpdateUserTags?.() || false,
    canDeleteUserTags: access.canDeleteUserTags?.() || false,

    // 租户标签权限（仅超级管理员）
    canManageTenantTags: access.canManageTenantTags?.() || false,
    canReadTenantTags: access.canReadTenantTags?.() || false,
    canUpdateTenantTags: access.canUpdateTenantTags?.() || false,
    canDeleteTenantTags: access.canDeleteTenantTags?.() || false,
  };
};
