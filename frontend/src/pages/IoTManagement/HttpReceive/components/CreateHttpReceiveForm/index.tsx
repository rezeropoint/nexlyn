import type { CreateHttpReceiveRequest } from "@/services/iot/api/httpReceive";
import { DrawerForm } from "@ant-design/pro-components";
import { Tabs } from "antd";
import React from "react";
import BasicInfoTab from "./BasicInfoTab";
import DataProcessingTab from "./DataProcessingTab";

interface CreateHttpReceiveFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: CreateHttpReceiveRequest) => Promise<boolean>;
}

/**
 * 创建 HTTP 接收配置表单组件
 *
 * HTTP 接收功能处理任意格式的 JSON 数据，不需要设备类别和标准字段
 */
const CreateHttpReceiveForm: React.FC<CreateHttpReceiveFormProps> = ({
  open,
  onOpenChange,
  onFinish,
}) => {
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
    const requestData: CreateHttpReceiveRequest = {
      name: values.name,
      description: values.description,
      enabled: values.enabled !== false,
      timestampPath: values.timestampPath,
      timestampFormat: values.timestampFormat,
      deviceIdPath: values.deviceIdPath,
      fieldMappings: values.fieldMappings || [],
      dispatchConfigs,
    };

    return await onFinish(requestData);
  };

  return (
    <DrawerForm
      title="创建 HTTP 接收配置"
      open={open}
      width={800}
      onOpenChange={onOpenChange}
      onFinish={handleFinish}
      drawerProps={{
        destroyOnHidden: true,
      }}
      initialValues={{
        enabled: true,
        fieldMappings: [],
        dispatchConfigs: [],
      }}
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

export default CreateHttpReceiveForm;
