import { useAccess } from "@umijs/max";
import type { UseDeviceTagPermissionsReturn } from "../types";

/**
 * 设备标签权限管理Hook
 */
export const useDeviceTagPermissions = (): UseDeviceTagPermissionsReturn => {
  const access = useAccess();

  return {
    canRead: access.hasPermission?.("device_tag", "read") || true, // 临时允许所有用户读取
    canCreate:
      access.hasPermission?.("device_tag", "write") ||
      access.isAdmin?.() ||
      true, // 临时允许
    canUpdate:
      access.hasPermission?.("device_tag", "write") ||
      access.isAdmin?.() ||
      true, // 临时允许
    canDelete:
      access.hasPermission?.("device_tag", "write") ||
      access.isAdmin?.() ||
      true, // 临时允许
  };
};
