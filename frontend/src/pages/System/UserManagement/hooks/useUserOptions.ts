import { getTagOptions } from "@/services/tags";
import { getTenantList } from "@/services/tenant";
import { useCallback, useState } from "react";
import type { UseUserOptionsReturn } from "../types";

/**
 * 用户选项数据Hook
 * 负责标签和租户下拉选项的加载和管理
 */
export const useUserOptions = (): UseUserOptionsReturn => {
  const [tagOptions, setTagOptions] = useState<API.TagOption[]>([]);
  const [tenantOptions, setTenantOptions] = useState<
    { label: string; value: string }[]
  >([]);

  // 加载标签选项
  const loadTagOptions = useCallback(async () => {
    try {
      const res = await getTagOptions({ scope: "user", limit: 100 });
      if (res.code === 0 && res.data?.list) {
        setTagOptions(res.data.list);
      }
    } catch (error) {
      console.error("加载标签选项失败:", error);
    }
  }, []);

  // 加载租户选项
  const loadTenantOptions = useCallback(async () => {
    try {
      const res = await getTenantList({ pageSize: 100, status: "active" });
      if (res.code === 0 && res.data?.list) {
        setTenantOptions(
          res.data.list.map((tenant) => ({
            label: tenant.tenantName,
            value: tenant.tenantKey,
          }))
        );
      }
    } catch (error) {
      console.error("加载租户选项失败:", error);
    }
  }, []);

  return {
    tagOptions,
    tenantOptions,
    loadTagOptions,
    loadTenantOptions,
  };
};
