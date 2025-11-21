import type { DeviceTag, NexlynDevice } from "@/services/video";
import {
  DrawerForm,
  type ProFormInstance,
  ProFormSelect,
} from "@ant-design/pro-components";
import React, { useEffect, useRef } from "react";
import type { DeviceTagsFormValues } from "../types";

interface EditDeviceTagsFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: DeviceTagsFormValues) => Promise<boolean>;
  currentDevice?: NexlynDevice;
  availableTags: DeviceTag[];
}

/**
 * 编辑设备标签表单组件
 */
const EditDeviceTagsForm: React.FC<EditDeviceTagsFormProps> = ({
  open,
  onOpenChange,
  onFinish,
  currentDevice,
  availableTags,
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
    <DrawerForm<DeviceTagsFormValues>
      title={`编辑设备标签 - ${
        currentDevice?.device_alias ||
        currentDevice?.name ||
        currentDevice?.device_id
      }`}
      open={open}
      onOpenChange={handleOpenChange}
      onFinish={onFinish}
      formRef={formRef}
      width={600}
      drawerProps={{ destroyOnHidden: true }}
    >
      <ProFormSelect
        name="tags"
        label="设备标签"
        placeholder="请选择设备标签"
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

export default EditDeviceTagsForm;
