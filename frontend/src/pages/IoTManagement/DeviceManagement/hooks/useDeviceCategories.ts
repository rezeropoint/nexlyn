import type { DeviceCategoryInfo } from "@/services/iot";
import { getDeviceCategories } from "@/services/iot";
import { useEffect, useState } from "react";

/**
 * 设备类别选项Hook
 * @returns {categoryOptions, loading}
 */
export const useDeviceCategories = () => {
  const [categoryOptions, setCategoryOptions] = useState<
    { label: string; value: string }[]
  >([]);
  const [loading, setLoading] = useState<boolean>(false);

  useEffect(() => {
    setLoading(true);
    getDeviceCategories()
      .then((response) => {
        if (response.code === 0 && response.data?.categories) {
          // 提取设备类别选项
          const categories = response.data.categories.map(
            (category: DeviceCategoryInfo) => ({
              label: category.name,
              value: category.code,
            })
          );
          setCategoryOptions(categories);
        } else {
          setCategoryOptions([]);
        }
      })
      .catch((error) => {
        console.error("获取设备类别列表失败:", error);
        setCategoryOptions([]);
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  return { categoryOptions, loading };
};
