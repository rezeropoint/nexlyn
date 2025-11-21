import type { DeviceTag, NexlynDevice } from "@/services/video";
import {
  DrawerForm,
  type ProFormInstance,
  ProFormSelect,
  ProFormText,
} from "@ant-design/pro-components";
import React, { useEffect, useRef } from "react";

// 统一的设备编辑表单值
export interface DeviceEditFormValues {
  deviceAlias: string;
  tags?: string[];
}

interface EditDeviceFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: DeviceEditFormValues) => Promise<boolean>;
  currentDevice?: NexlynDevice;
  availableTags: DeviceTag[];
  loading?: boolean;
}

/**
 * 统一的设备编辑表单组件
 * 合并了设备别名和标签编辑功能
 */
const EditDeviceForm: React.FC<EditDeviceFormProps> = ({
  open,
  onOpenChange,
  onFinish,
  currentDevice,
  availableTags,
  loading = false,
}) => {
  const formRef = useRef<ProFormInstance>();

  // 将可用标签转换为选项格式
  const tagOptions = availableTags.map((tag) => ({
    label: tag.tagName,
    value: tag.tagName,
  }));

  // 同步表单数据
  useEffect(() => {
    if (open && currentDevice && formRef.current) {
      formRef.current.setFieldsValue({
        deviceAlias: currentDevice.deviceAlias || "",
        tags: currentDevice.tags || [],
      });
    }
  }, [open, currentDevice]);

  // 处理抽屉关闭
  const handleOpenChange = (visible: boolean) => {
    if (!visible) {
      formRef.current?.resetFields();
    }
    onOpenChange(visible);
  };

  return (
    <DrawerForm<DeviceEditFormValues>
      title={`编辑设备 - ${currentDevice?.name || currentDevice?.deviceId}`}
      open={open}
      onOpenChange={handleOpenChange}
      onFinish={onFinish}
      formRef={formRef}
      width={600}
      loading={loading}
      drawerProps={{ destroyOnHidden: true }}
    >
      <ProFormText
        name="deviceAlias"
        label="设备别名"
        placeholder="请输入设备别名"
        rules={[
          { required: true, message: "请输入设备别名" },
          { max: 50, message: "设备别名不能超过50个字符" },
        ]}
        extra="设备别名用于更好地识别设备，建议使用有意义的名称"
      />

      <ProFormSelect
        name="tags"
        label="设备标签"
        placeholder="请选择设备标签（可选）"
        mode="tags"
        options={tagOptions}
        fieldProps={{
          maxTagCount: 10,
          maxTagTextLength: 20,
          showSearch: true,
          filterOption: (input, option) =>
            (option?.label ?? "").toLowerCase().includes(input.toLowerCase()),
        }}
        rules={[
          {
            validator: async (_, value) => {
              if (value && value.length > 10) {
                throw new Error("最多只能选择10个标签");
              }
              if (value?.some((tag: string) => tag.length > 20)) {
                throw new Error("标签名称不能超过20个字符");
              }
              return Promise.resolve();
            },
          },
        ]}
        extra="可以选择现有标签，也可以输入新标签名称。最多10个标签，每个标签最长20字符"
      />
    </DrawerForm>
  );
};

export default EditDeviceForm;
