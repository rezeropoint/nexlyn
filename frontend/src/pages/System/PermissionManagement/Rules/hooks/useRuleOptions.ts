import { getTenantOptions } from "@/services/tenant";
import { getUserOptions } from "@/services/user";
import { useCallback, useState } from "react";

interface UserOption {
  userKey: string;
  userName: string;
  name: string;
  email: string;
  tenantId: string;
}

interface TenantOption {
  tenantKey: string;
  tenantId: string;
  tenantName: string;
  status: string;
}

export const useRuleOptions = () => {
  const [userOptions, setUserOptions] = useState<UserOption[]>([]);
  const [tenantOptions, setTenantOptions] = useState<TenantOption[]>([]);
  const [loading, setLoading] = useState({
    users: false,
    tenants: false,
  });

  // 加载用户选项
  const loadUserOptions = useCallback(async (keyword?: string) => {
    try {
      setLoading((prev) => ({ ...prev, users: true }));

      const res = await getUserOptions({
        keyword,
        limit: 50,
      });

      if (res.code === 0 && res.data?.list) {
        setUserOptions(res.data.list);
      } else {
        console.error("获取用户选项失败:", res.msg);
        setUserOptions([]);
      }
    } catch (error) {
      console.error("获取用户选项失败:", error);
      setUserOptions([]);
    } finally {
      setLoading((prev) => ({ ...prev, users: false }));
    }
  }, []);

  // 加载租户选项
  const loadTenantOptions = useCallback(async (keyword?: string) => {
    try {
      setLoading((prev) => ({ ...prev, tenants: true }));

      const res = await getTenantOptions({
        keyword,
        limit: 50,
        status: "active", // 只获取活跃的租户
      });

      if (res.code === 0 && res.data?.list) {
        setTenantOptions(res.data.list);
      } else {
        console.error("获取租户选项失败:", res.msg);
        setTenantOptions([]);
      }
    } catch (error) {
      console.error("获取租户选项失败:", error);
      setTenantOptions([]);
    } finally {
      setLoading((prev) => ({ ...prev, tenants: false }));
    }
  }, []);

  return {
    userOptions,
    tenantOptions,
    loading,
    loadUserOptions,
    loadTenantOptions,
  };
};
