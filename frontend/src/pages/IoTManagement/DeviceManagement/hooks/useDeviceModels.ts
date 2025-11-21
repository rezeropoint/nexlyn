import { getSensorTemplateList } from "@/services/iot";
import { useEffect, useState } from "react";

/**
 * 设备型号选项Hook
 * @returns {modelOptions, categoryOptions, loading}
 */
export const useDeviceModels = () => {
  const [modelOptions, setModelOptions] = useState<
    { label: string; value: string }[]
  >([]);
  const [loading, setLoading] = useState<boolean>(false);

  useEffect(() => {
    setLoading(true);
    getSensorTemplateList({ enabled: "true" })
      .then((response) => {
        if (response.code === 0 && response.data?.list) {
          // 提取设备型号选项
          const models = response.data.list.map((template) => ({
            label: `${template.name} (${template.model})`,
            value: template.model,
          }));
          setModelOptions(models);
        } else {
          setModelOptions([]);
        }
      })
      .catch((error) => {
        console.error("获取设备型号列表失败:", error);
        setModelOptions([]);
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  return { modelOptions, loading };
};
