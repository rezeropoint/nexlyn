import {
  createEventConfig,
  getFlowFields,
  getFlowList,
  updateEventConfig,
} from "@/services/eventhandler";
import type { ProFormInstance } from "@ant-design/pro-components";
import {
  DrawerForm,
  ProFormSelect,
  ProFormSwitch,
  ProFormText,
  ProFormTextArea,
} from "@ant-design/pro-components";
import { Alert, message, Spin } from "antd";
import React, { useEffect, useRef, useState } from "react";
import { MESSAGE } from "../constants";
import type {
  EventConfigWithFields,
  FieldConfigInput,
  FieldMetadata,
  FlowInfo,
} from "../types";
import styles from "./EventConfigForm.less";
import FieldConfigTable from "./FieldConfigTable";

interface EventConfigFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  currentConfig?: EventConfigWithFields;
  onSuccess: () => void;
}

/**
 * 事件配置表单
 */
const EventConfigForm: React.FC<EventConfigFormProps> = ({
  open,
  onOpenChange,
  currentConfig,
  onSuccess,
}) => {
  const formRef = useRef<ProFormInstance>();
  const [loading, setLoading] = useState(false);
  const [flows, setFlows] = useState<FlowInfo[]>([]);
  const [flowFields, setFlowFields] = useState<{
    systemFields: FieldMetadata[];
    businessFields: FieldMetadata[];
  }>({ systemFields: [], businessFields: [] });
  const [fields, setFields] = useState<FieldConfigInput[]>([]);
  const isEdit = !!currentConfig;

  // 定义初始状态常量
  const INITIAL_FLOW_FIELDS = { systemFields: [], businessFields: [] };
  const INITIAL_FIELDS: FieldConfigInput[] = [];

  // 加载流程列表
  const loadFlows = async () => {
    try {
      const response = await getFlowList();
      if (response.code === 0 && response.data) {
        setFlows(response.data.list || []);
      }
    } catch (error) {
      console.error("获取流程列表失败:", error);
      message.error("获取流程列表失败");
    }
  };

  // 加载流程字段
  const loadFlowFields = async (flowId: number) => {
    setLoading(true);
    try {
      const response = await getFlowFields(flowId);
      if (response.code === 0 && response.data) {
        // 确保字段数组不为null
        setFlowFields({
          systemFields: response.data.systemFields || [],
          businessFields: response.data.businessFields || [],
        });

        // 如果是新建，自动生成默认字段配置
        if (!isEdit) {
          const businessFields = response.data.businessFields || [];
          const defaultFields: FieldConfigInput[] = businessFields.map(
            (field, index) => ({
              fieldName: field.fieldName,
              displayName: field.fieldName,
              fieldType: mapDataTypeToFieldType(field.dataType),
              isVisible: true,
              displayOrder: index,
              isSearchable:
                field.fieldName.includes("name") ||
                field.fieldName.includes("title"),
            })
          );
          setFields(defaultFields);
        }
      }
    } catch (error) {
      console.error("获取流程字段失败:", error);
      message.error("获取流程字段失败");
    } finally {
      setLoading(false);
    }
  };

  // 数据类型映射
  const mapDataTypeToFieldType = (dataType: string): string => {
    if (dataType.includes("int") || dataType.includes("numeric"))
      return "number";
    if (dataType.includes("bool")) return "boolean";
    if (dataType.includes("date")) return "datetime";
    if (dataType.includes("timestamp")) return "datetime";
    return "string";
  };

  useEffect(() => {
    if (open) {
      loadFlows();

      if (currentConfig) {
        // 编辑模式：设置表单值
        formRef.current?.setFieldsValue({
          name: currentConfig.name,
          flowId: currentConfig.flowId,
          orgFieldName: currentConfig.orgFieldName,
          description: currentConfig.description,
          enabled: currentConfig.enabled,
        });
        setFields(
          currentConfig.fields.map((f) => ({
            fieldName: f.fieldName,
            displayName: f.displayName,
            fieldType: f.fieldType,
            isVisible: f.isVisible,
            displayOrder: f.displayOrder,
            isSearchable: f.isSearchable,
          }))
        );

        // 加载对应流程的字段
        if (currentConfig.flowId) {
          loadFlowFields(currentConfig.flowId);
        }
      } else {
        // 新建模式：重置表单
        formRef.current?.resetFields();
        formRef.current?.setFieldsValue({ enabled: true });
        setFields([]);
      }
    }
  }, [open, currentConfig]);

  // 处理流程选择变化
  const handleFlowChange = (flowId: number) => {
    loadFlowFields(flowId);

    // 更新流程标题
    const selectedFlow = flows.find((f) => f.id === flowId);
    if (selectedFlow) {
      formRef.current?.setFieldValue("flowTitle", selectedFlow.title);
    }
  };

  // 提交表单
  const handleSubmit = async (values: any) => {
    try {
      const data = {
        ...values,
        fields,
      };

      const response = isEdit
        ? await updateEventConfig(currentConfig.id, data)
        : await createEventConfig(data);

      if (response.code === 0) {
        message.success(
          isEdit ? MESSAGE.UPDATE_SUCCESS : MESSAGE.CREATE_SUCCESS
        );
        onSuccess();
        return true;
      } else {
        message.error(
          response.msg ||
            (isEdit ? MESSAGE.UPDATE_FAILED : MESSAGE.CREATE_FAILED)
        );
        return false;
      }
    } catch (error) {
      console.error("保存事件配置失败:", error);
      message.error(isEdit ? MESSAGE.UPDATE_FAILED : MESSAGE.CREATE_FAILED);
      return false;
    }
  };

  return (
    <DrawerForm
      title={isEdit ? "编辑事件配置" : "新建事件配置"}
      open={open}
      formRef={formRef}
      onOpenChange={(visible) => {
        if (!visible) {
          formRef.current?.resetFields();
          setFlowFields(INITIAL_FLOW_FIELDS);
          setFields(INITIAL_FIELDS);
        }
        onOpenChange(visible);
      }}
      onFinish={handleSubmit}
      drawerProps={{
        destroyOnHidden: true,
      }}
      width={800}
      layout="horizontal"
      labelCol={{ span: 4 }}
      wrapperCol={{ span: 20 }}
    >
      <Spin spinning={loading}>
        <ProFormText
          name="name"
          label="事件名称"
          placeholder="请输入事件显示名称"
          rules={[{ required: true, message: "请输入事件名称" }]}
          tooltip="用于页面展示的自定义事件名称"
        />

        <ProFormSelect
          name="flowId"
          label="关联流程"
          placeholder="请选择Skylark流程"
          rules={[{ required: true, message: "请选择流程" }]}
          disabled={isEdit}
          options={flows.map((flow) => ({
            label: `${flow.title} (ID: ${flow.id})`,
            value: flow.id,
          }))}
          fieldProps={{
            onChange: handleFlowChange,
          }}
          tooltip={isEdit ? "流程ID创建后不可修改" : "选择要查询的Skylark流程"}
        />

        {flowFields.businessFields.length > 0 && (
          <ProFormSelect
            name="orgFieldName"
            label="组织字段"
            placeholder="选择用于组织权限过滤的字段（可选）"
            options={flowFields.businessFields.map((field) => ({
              label: field.fieldName,
              value: field.fieldName,
            }))}
            tooltip="选择包含组织信息的字段，用于数据权限控制"
          />
        )}

        <ProFormTextArea
          name="description"
          label="描述"
          placeholder="请输入事件配置描述"
          fieldProps={{
            rows: 2,
          }}
        />

        <ProFormSwitch
          name="enabled"
          label="启用状态"
          checkedChildren="启用"
          unCheckedChildren="禁用"
          initialValue={true}
        />

        {flowFields.businessFields.length > 0 ? (
          <>
            <Alert
              message="字段配置"
              description="请配置要在事件列表中显示的字段。可以调整显示名称、顺序和搜索设置。"
              type="info"
              className={styles.fieldConfigAlert}
            />
            <FieldConfigTable
              fields={fields}
              onChange={setFields}
              availableFields={flowFields.businessFields}
            />
          </>
        ) : (
          formRef.current?.getFieldValue("flowId") && (
            <Alert
              message="请先选择流程"
              description="选择流程后将自动加载可用字段"
              type="warning"
            />
          )
        )}
      </Spin>
    </DrawerForm>
  );
};

export default EventConfigForm;
