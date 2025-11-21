import { getTagOptions } from "@/services/tags";
import { useCallback, useState } from "react";
import type { UseTenantOptionsReturn } from "../types";

/**
 * 租户选项数据Hook
 * 负责标签选项的加载和管理
 */
export const useTenantOptions = (): UseTenantOptionsReturn => {
  const [tagOptions, setTagOptions] = useState<
    { label: string; value: string }[]
  >([]);

  // 加载标签选项
  const loadTagOptions = useCallback(async () => {
    try {
      const res = await getTagOptions({ scope: "tenant", limit: 100 });
      if (res.code === 0 && res.data?.list) {
        const options = res.data.list.map((item) => ({
          label: item.label,
          value: item.label, // 使用 label 作为值
        }));
        setTagOptions(options);
      }
    } catch (error) {
      console.error("加载标签选项失败:", error);
    }
  }, []);

  return {
    tagOptions,
    loadTagOptions,
  };
};
