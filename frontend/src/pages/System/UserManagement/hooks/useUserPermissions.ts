import { useAccess } from "@umijs/max";
import type { UserPermissionCheck } from "../types";

/**
 * 用户权限检查Hook
 * 封装所有用户相关的权限检查逻辑
 */
export const useUserPermissions = (): UserPermissionCheck => {
  const access = useAccess();

  return {
    canCreate: access.canCreateUsers(),
    canUpdate: access.canUpdateUsers(),
    canDelete: access.canDeleteUsers(),
    canOperateUser: access.canOperateUser,
  };
};
