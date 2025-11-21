import type {
  DataProcessingConfig,
  DeviceControlConfig,
  OnlineDetectionConfig,
  SensorTemplate,
  UpdateSensorTemplateRequest,
} from "@/services/iot";
import { getSensorTemplate } from "@/services/iot";
import { DrawerForm, type ProFormInstance } from "@ant-design/pro-components";
import { message, Spin, Tabs } from "antd";
import React, { useEffect, useRef, useState } from "react";
import { useDeviceCategories } from "../../hooks/useDeviceCategories";
import { useStandardFields } from "../../hooks/useStandardFields";
import ControlConfigTab from "../CreateTemplateForm/ControlConfigTab";
import DataProcessingTab from "../CreateTemplateForm/DataProcessingTab";
import OnlineDetectionTab from "../CreateTemplateForm/OnlineDetectionTab";
import EditBasicInfoTab from "./EditBasicInfoTab";

interface EditTemplateFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: UpdateSensorTemplateRequest) => Promise<boolean>;
  currentRow: SensorTemplate | undefined;
}

/**
 * 编辑设备模板表单组件（重构版）
 */
const EditTemplateForm: React.FC<EditTemplateFormProps> = ({
  open,
  onOpenChange,
  onFinish,
  currentRow,
}) => {
  const formRef = useRef<ProFormInstance>();
  const [loading, setLoading] = useState<boolean>(false);

  // 当前选择的设备类别
  const [selectedCategory, setSelectedCategory] = useState<string>("");

  // 使用自定义Hooks加载设备类别和标准字段
  const { deviceCategories, loading: categoriesLoading } =
    useDeviceCategories(open);
  const { standardFields, loading: fieldsLoading } =
    useStandardFields(selectedCategory);

  // 加载模板详情（包含完整配置）
  useEffect(() => {
    if (open && currentRow?.id) {
      setLoading(true);
      getSensorTemplate(currentRow.id)
        .then((response) => {
          if (response.code === 0 && response.data) {
            const detail = response.data;

            // 设置当前选择的类别，触发标准字段加载
            if (detail.category) {
              setSelectedCategory(detail.category);
            }

            // 设置表单初始值（展开配置到各个字段）
            const values: any = {
              // 基础信息
              model: detail.model,
              name: detail.name,
              category: detail.category,
              manufacturer: detail.manufacturer,
              description: detail.description,
              version: detail.version,
              enabled: detail.enabled,
            };

            // 在线检测配置（只在配置有效时才设置表单字段）
            if (detail.onlineConfig?.topicSuffix) {
              values.onlineTopicSuffix = detail.onlineConfig.topicSuffix;
              values.onlineStrategy = detail.onlineConfig.strategy;
              values.timeoutSeconds = detail.onlineConfig.timeoutSeconds;
              values.fieldChecks = detail.onlineConfig.fieldChecks || [];
            }

            // 业务数据处理配置（只在配置有效时才设置表单字段）
            if (detail.businessConfig?.topicSuffix) {
              values.businessTopicSuffix = detail.businessConfig.topicSuffix;
              values.timestampPath = detail.businessConfig.timestampPath;
              values.timestampFormat = detail.businessConfig.timestampFormat;
              values.fieldMappings = detail.businessConfig.fieldMappings || [];
              values.filterRules = detail.businessConfig.filterRules || null;
              values.dispatchConfigs =
                detail.businessConfig.dispatchConfigs || [];
            }

            // 控制配置（只在配置有效时才设置表单字段）
            if (detail.controlConfig?.commandSuffix) {
              values.commandSuffix = detail.controlConfig.commandSuffix;
              values.responseSuffix = detail.controlConfig.responseSuffix;
            }

            // 使用 form 实例设置字段值
            formRef.current?.setFieldsValue(values);
          } else {
            message.error(response.msg || "加载模板详情失败");
          }
        })
        .catch((error) => {
          message.error(`加载模板详情失败: ${error.message}`);
        })
        .finally(() => {
          setLoading(false);
        });
    } else if (!open) {
      // 关闭时清空
      formRef.current?.resetFields();
      setSelectedCategory("");
    }
  }, [open, currentRow]);

  // 处理表单提交
  const handleFinish = async (values: any) => {
    // 构建在线检测配置（可选）
    // category 和 model 来自禁用的表单字段（值来自数据库，不会改变）
    let onlineConfig: OnlineDetectionConfig | undefined;
    if (values.onlineTopicSuffix && values.onlineStrategy) {
      onlineConfig = {
        category: values.category || "",
        model: values.model || "",
        topicSuffix: values.onlineTopicSuffix,
        strategy: values.onlineStrategy,
        timeoutSeconds: values.timeoutSeconds || 300,
        fieldChecks: values.fieldChecks || [],
      };
    }

    // 构建业务数据处理配置（可选）
    // category 和 model 来自禁用的表单字段（值来自数据库，不会改变）
    let businessConfig: DataProcessingConfig | undefined;
    if (values.businessTopicSuffix && values.fieldMappings?.length > 0) {
      businessConfig = {
        category: values.category || "",
        model: values.model || "",
        topicSuffix: values.businessTopicSuffix,
        timestampPath: values.timestampPath,
        timestampFormat: values.timestampFormat || "unix",
        fieldMappings: values.fieldMappings,
        filterRules: values.filterRules || null,
        dispatchConfigs: values.dispatchConfigs || [],
      };
    }

    // 构建控制配置（可选）
    // category 和 model 来自禁用的表单字段（值来自数据库，不会改变）
    let controlConfig: DeviceControlConfig | undefined;
    if (values.commandSuffix && values.responseSuffix) {
      controlConfig = {
        category: values.category || "",
        model: values.model || "",
        commandSuffix: values.commandSuffix,
        responseSuffix: values.responseSuffix,
      };
    }

    // 构建更新请求
    // 注意：model 和 category 必须传递（来自禁用字段），但后端会验证不允许修改
    const requestData: UpdateSensorTemplateRequest = {
      model: values.model,
      name: values.name,
      category: values.category,
      manufacturer: values.manufacturer,
      description: values.description,
      version: values.version,
      enabled: values.enabled,
      onlineConfig,
      businessConfig,
      controlConfig,
    };

    return await onFinish(requestData);
  };

  return (
    <DrawerForm
      formRef={formRef}
      title="编辑设备模板"
      open={open}
      width={800}
      onOpenChange={onOpenChange}
      onFinish={handleFinish}
      drawerProps={{
        destroyOnHidden: true,
      }}
    >
      <Spin spinning={loading}>
        <Tabs
          defaultActiveKey="basic"
          items={[
            {
              key: "basic",
              label: "基础信息",
              forceRender: true, // 强制渲染，确保表单字段被创建
              children: (
                <EditBasicInfoTab
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
      </Spin>
    </DrawerForm>
  );
};

export default EditTemplateForm;
