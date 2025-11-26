import type { HttpReceiveDetail } from "@/services/iot";
import type { UpdateHttpReceiveRequest } from "@/services/iot/api/httpReceive";
import { DrawerForm } from "@ant-design/pro-components";
import { Tabs } from "antd";
import React, { useMemo } from "react";
import BasicInfoTab from "../CreateHttpReceiveForm/BasicInfoTab";
import DataProcessingTab from "../CreateHttpReceiveForm/DataProcessingTab";

interface EditHttpReceiveFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: UpdateHttpReceiveRequest) => Promise<boolean>;
  initialData?: HttpReceiveDetail;
}

/**
 * 编辑 HTTP 接收配置表单组件
 *
 * HTTP 接收功能处理任意格式的 JSON 数据，不需要设备类别和标准字段
 */
const EditHttpReceiveForm: React.FC<EditHttpReceiveFormProps> = ({
  open,
  onOpenChange,
  onFinish,
  initialData,
}) => {
  // 计算初始值
  const initialValues = useMemo(() => {
    if (!initialData) {
      return {
        enabled: true,
        fieldMappings: [],
        dispatchConfigs: [],
      };
    }

    // 将 dispatchConfigs 中的 extraParams 转为字符串
    const dispatchConfigs = (initialData.dispatchConfigs || []).map(
      (config) => ({
        ...config,
        extraParams: config.extraParams
          ? JSON.stringify(config.extraParams, null, 2)
          : undefined,
      })
    );

    return {
      name: initialData.name,
      description: initialData.description,
      enabled: initialData.enabled,
      timestampPath: initialData.timestampPath,
      timestampFormat: initialData.timestampFormat,
      deviceIdPath: initialData.deviceIdPath,
      fieldMappings: initialData.fieldMappings || [],
      dispatchConfigs,
    };
  }, [initialData]);

  // 处理表单提交
  const handleFinish = async (values: any) => {
    // 处理扩展参数（将字符串转为对象）
    const dispatchConfigs = (values.dispatchConfigs || []).map(
      (config: any) => {
        let extraParams: Record<string, any> | undefined;
        if (config.extraParams) {
          try {
            extraParams = JSON.parse(config.extraParams);
          } catch {
            extraParams = undefined;
          }
        }
        return {
          ...config,
          extraParams,
        };
      }
    );

    // 构建请求数据
    const requestData: UpdateHttpReceiveRequest = {
      name: values.name,
      description: values.description,
      enabled: values.enabled,
      timestampPath: values.timestampPath,
      timestampFormat: values.timestampFormat,
      deviceIdPath: values.deviceIdPath,
      fieldMappings: values.fieldMappings,
      dispatchConfigs,
    };

    return await onFinish(requestData);
  };

  return (
    <DrawerForm
      title="编辑 HTTP 接收配置"
      open={open}
      width={800}
      onOpenChange={onOpenChange}
      onFinish={handleFinish}
      drawerProps={{
        destroyOnHidden: true,
      }}
      initialValues={initialValues}
      key={initialData?.id}
    >
      <Tabs
        defaultActiveKey="basic"
        items={[
          {
            key: "basic",
            label: "基础信息",
            forceRender: true,
            children: <BasicInfoTab />,
          },
          {
            key: "dataProcessing",
            label: "数据处理与分发",
            forceRender: true,
            children: <DataProcessingTab />,
          },
        ]}
      />
    </DrawerForm>
  );
};

export default EditHttpReceiveForm;
