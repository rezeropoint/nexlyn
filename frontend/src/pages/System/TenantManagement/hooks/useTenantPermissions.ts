import type { Tenant } from "@/services/tenant";
import { useAccess, useModel } from "@umijs/max";
import type { TenantPermissionCheck } from "../types";

/**
 * 租户权限检查Hook
 * 封装所有租户相关的权限检查逻辑
 */
export const useTenantPermissions = (): TenantPermissionCheck => {
  const access = useAccess();
  const { initialState } = useModel("@@initialState");
  const currentUser = initialState?.currentUser;

  // 检查是否可以操作指定租户
  const canOperateTenant = (tenant: Tenant): boolean => {
    // 超级管理员可以操作所有租户
    if (access.isSuperAdmin()) {
      return true;
    }

    // 非超级管理员不能删除自己所属的租户
    if (tenant.id === currentUser?.tenantInfo?.tenantId) {
      return false;
    }

    // 其他情况下，需要有相应的权限
    return true;
  };

  return {
    canCreate: access.canCreateTenants(),
    canUpdate: access.canUpdateTenants(),
    canDelete: access.canDeleteTenants(),
    canOperateTenant,
  };
};
