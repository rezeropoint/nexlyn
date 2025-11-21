import type {
  CreateSensorTemplateRequest,
  DataProcessingConfig,
  DeviceControlConfig,
  OnlineDetectionConfig,
} from "@/services/iot";
import { DrawerForm } from "@ant-design/pro-components";
import { Tabs } from "antd";
import React, { useState } from "react";
import { useDeviceCategories } from "../../hooks/useDeviceCategories";
import { useStandardFields } from "../../hooks/useStandardFields";
import BasicInfoTab from "./BasicInfoTab";
import ControlConfigTab from "./ControlConfigTab";
import DataProcessingTab from "./DataProcessingTab";
import OnlineDetectionTab from "./OnlineDetectionTab";

interface CreateTemplateFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: CreateSensorTemplateRequest) => Promise<boolean>;
}

/**
 * 创建设备模板表单组件
 */
const CreateTemplateForm: React.FC<CreateTemplateFormProps> = ({
  open,
  onOpenChange,
  onFinish,
}) => {
  // 当前选择的设备类别
  const [selectedCategory, setSelectedCategory] = useState<string>("");

  // 使用自定义Hooks加载设备类别和标准字段
  const { deviceCategories, loading: categoriesLoading } =
    useDeviceCategories(open);
  const { standardFields, loading: fieldsLoading } =
    useStandardFields(selectedCategory);

  // 处理表单提交
  const handleFinish = async (values: any) => {
    // 构建在线检测配置（可选）
    let onlineConfig: OnlineDetectionConfig | undefined;
    const hasOnlineConfig = values.onlineTopicSuffix && values.onlineStrategy;
    if (hasOnlineConfig) {
      onlineConfig = {
        category: values.category || "",
        model: values.model,
        topicSuffix: values.onlineTopicSuffix,
        strategy: values.onlineStrategy,
        timeoutSeconds: values.timeoutSeconds || 300,
        fieldChecks: values.fieldChecks || [],
      };
    }

    // 构建业务数据处理配置（可选）
    let businessConfig: DataProcessingConfig | undefined;
    const hasBusinessConfig =
      values.businessTopicSuffix && values.fieldMappings?.length > 0;
    if (hasBusinessConfig) {
      businessConfig = {
        category: values.category || "",
        model: values.model,
        topicSuffix: values.businessTopicSuffix,
        timestampPath: values.timestampPath,
        timestampFormat: values.timestampFormat || "unix",
        fieldMappings: values.fieldMappings,
        filterRules: values.filterRules || null,
      };
    }

    // 构建控制配置（可选）
    let controlConfig: DeviceControlConfig | undefined;
    const hasControlConfig = values.commandSuffix && values.responseSuffix;
    if (hasControlConfig) {
      controlConfig = {
        category: values.category || "",
        model: values.model,
        commandSuffix: values.commandSuffix,
        responseSuffix: values.responseSuffix,
      };
    }

    // 构建请求数据
    const requestData: CreateSensorTemplateRequest = {
      model: values.model,
      name: values.name,
      category: values.category,
      manufacturer: values.manufacturer,
      description: values.description,
      version: values.version || "1.0.0",
      enabled: values.enabled !== false,
      onlineConfig,
      businessConfig,
      controlConfig,
    };

    return await onFinish(requestData);
  };

  return (
    <DrawerForm
      title="创建设备模板"
      open={open}
      width={800}
      onOpenChange={onOpenChange}
      onFinish={handleFinish}
      drawerProps={{
        destroyOnHidden: true,
      }}
      initialValues={{
        enabled: true,
        version: "1.0.0",
        // 在线检测默认值
        onlineStrategy: "any_data",
        timeoutSeconds: 300,
        fieldChecks: [],
        // 业务数据处理默认值
        timestampFormat: "unix",
        fieldMappings: [],
        filterRules: null,
        // 控制配置默认值
        commandSuffix: "control",
        responseSuffix: "control_response",
      }}
    >
      <Tabs
        defaultActiveKey="basic"
        items={[
          {
            key: "basic",
            label: "基础信息",
            forceRender: true, // 强制渲染，确保表单字段被创建
            children: (
              <BasicInfoTab
                deviceCategories={deviceCategories}
                categoriesLoading={categoriesLoading}
                onCategoryChange={setSelectedCategory}
              />
            ),
          },
          {
            key: "online",
            label: "在线检测配置",
            forceRender: true, // 强制渲染，确保表单字段被创建
            children: <OnlineDetectionTab />,
          },
          {
            key: "business",
            label: "业务数据处理",
            forceRender: true, // 强制渲染，确保表单字段被创建
            children: (
              <DataProcessingTab
                deviceCategories={deviceCategories}
                standardFields={standardFields}
                fieldsLoading={fieldsLoading}
              />
            ),
          },
          {
            key: "control",
            label: "设备控制配置",
            forceRender: true, // 强制渲染，确保表单字段被创建
            children: <ControlConfigTab />,
          },
        ]}
      />
    </DrawerForm>
  );
};

export default CreateTemplateForm;
