import { useAccess } from "@umijs/max";
import type { UseDevicePermissionsReturn } from "../types";

/**
 * 设备权限管理Hook
 * 集中管理设备相关权限检查
 */
export const useDevicePermissions = (): UseDevicePermissionsReturn => {
  const access = useAccess();

  // 检查是否为超级管理员
  const isSuperAdmin = access.isSuperAdmin?.() || false;

  // 检查是否为租户管理员
  const _isTenantAdmin = access.isTenantAdmin?.() || false;

  return {
    // 读取权限 - 查看设备列表和详情
    canRead: isSuperAdmin || access.hasPermission?.("device", "read") || true, // 临时允许所有用户读取

    // 绑定权限 - 绑定新设备（主要是超级管理员权限）
    canBind:
      isSuperAdmin ||
      access.hasPermission?.("device", "write") ||
      access.isAdmin?.() ||
      true, // 临时允许

    // 解绑权限 - 解绑设备
    canUnbind:
      isSuperAdmin ||
      access.hasPermission?.("device", "write") ||
      access.isAdmin?.() ||
      true, // 临时允许

    // 编辑权限 - 编辑设备别名和标签
    canEdit:
      isSuperAdmin ||
      access.hasPermission?.("device", "write") ||
      access.isAdmin?.() ||
      true, // 临时允许

    // 控制权限 - 设备控制（PTZ、录像等）
    canControl:
      isSuperAdmin ||
      access.hasPermission?.("device", "control") ||
      access.isAdmin?.() ||
      true, // 临时允许

    // 查看未绑定设备权限 - 只有超级管理员可以查看跨租户的未绑定设备
    canViewUnbound: isSuperAdmin,
  };
};
