import { getRoleList } from "@/services/permission/role";
import { useCallback, useState } from "react";

interface RoleOption {
  label: string;
  value: string;
}

export interface UseRoleOptionsReturn {
  roleOptions: RoleOption[];
  loading: boolean;
  loadRoleOptions: () => Promise<void>;
  refreshRoleOptions: () => Promise<void>;
}

/**
 * 角色选项数据Hook
 * 用于统一管理角色选项的获取和状态
 */
export const useRoleOptions = (): UseRoleOptionsReturn => {
  const [roleOptions, setRoleOptions] = useState<RoleOption[]>([]);
  const [loading, setLoading] = useState(false);

  // 加载角色选项
  const loadRoleOptions = useCallback(async () => {
    if (loading) return;

    setLoading(true);
    try {
      const res = await getRoleList({ pageSize: 100 });
      if (res.code === 0 && res.data) {
        setRoleOptions(
          res.data.map((role) => ({
            label: role.roleName,
            value: role.roleKey,
          }))
        );
      }
    } catch (error) {
      console.error("加载角色选项失败:", error);
    } finally {
      setLoading(false);
    }
  }, [loading]);

  // 刷新角色选项（强制重新加载）
  const refreshRoleOptions = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getRoleList({ pageSize: 100 });
      if (res.code === 0 && res.data) {
        setRoleOptions(
          res.data.map((role) => ({
            label: role.roleName,
            value: role.roleKey,
          }))
        );
      }
    } catch (error) {
      console.error("刷新角色选项失败:", error);
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    roleOptions,
    loading,
    loadRoleOptions,
    refreshRoleOptions,
  };
};
