import type { StandardFieldInfo } from "@/services/iot";
import { getStandardFields } from "@/services/iot";
import { message } from "antd";
import { useEffect, useState } from "react";

/**
 * 标准字段加载Hook
 * @param category - 设备类别
 * @returns {standardFields, loading}
 */
export const useStandardFields = (category: string) => {
  const [standardFields, setStandardFields] = useState<StandardFieldInfo[]>([]);
  const [loading, setLoading] = useState<boolean>(false);

  useEffect(() => {
    if (category) {
      setLoading(true);
      getStandardFields(category)
        .then((response) => {
          if (response.code === 0 && response.data?.fields) {
            setStandardFields(response.data.fields);
          } else {
            message.warning("获取标准字段列表失败");
            setStandardFields([]);
          }
        })
        .catch((error) => {
          console.error("获取标准字段失败:", error);
          message.warning("获取标准字段列表失败");
          setStandardFields([]);
        })
        .finally(() => {
          setLoading(false);
        });
    } else {
      setStandardFields([]);
    }
  }, [category]);

  return { standardFields, loading };
};
