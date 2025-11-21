import type { DeviceCategoryInfo } from "@/services/iot";
import { getDeviceCategories } from "@/services/iot";
import { message } from "antd";
import { useEffect, useState } from "react";

/**
 * 设备类别加载Hook
 * @param open - 是否打开（用于触发加载）
 * @returns {deviceCategories, loading}
 */
export const useDeviceCategories = (open: boolean) => {
  const [deviceCategories, setDeviceCategories] = useState<
    DeviceCategoryInfo[]
  >([]);
  const [loading, setLoading] = useState<boolean>(false);

  useEffect(() => {
    if (open) {
      setLoading(true);
      getDeviceCategories()
        .then((response) => {
          if (response.code === 0 && response.data?.categories) {
            setDeviceCategories(response.data.categories);
          } else {
            message.warning("获取设备类别列表失败，将使用默认选项");
            setDeviceCategories([]);
          }
        })
        .catch((error) => {
          console.error("获取设备类别失败:", error);
          message.warning("获取设备类别列表失败，将使用默认选项");
          setDeviceCategories([]);
        })
        .finally(() => {
          setLoading(false);
        });
    } else {
      // 关闭时重置
      setDeviceCategories([]);
    }
  }, [open]);

  return { deviceCategories, loading };
};
