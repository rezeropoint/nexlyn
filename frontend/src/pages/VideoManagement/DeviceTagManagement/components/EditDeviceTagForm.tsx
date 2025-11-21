import type { DeviceTag } from "@/services/video";
import {
  DrawerForm,
  ProFormText,
  ProFormTextArea,
  type ProFormInstance,
} from "@ant-design/pro-components";
import React, { useEffect, useRef } from "react";

// 表单验证规则
const DEVICE_TAG_FORM_RULES = {
  TAG_NAME: {
    REQUIRED: { required: true, message: "请输入标签名称" },
    MAX_LENGTH: { max: 20, message: "标签名称不能超过20个字符" },
  },
  DESCRIPTION: {
    MAX_LENGTH: { max: 200, message: "描述不能超过200个字符" },
  },
} as const;

interface EditDeviceTagFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: any) => Promise<boolean>;
  currentRow?: DeviceTag;
}

/**
 * 编辑设备标签表单组件
 */
const EditDeviceTagForm: React.FC<EditDeviceTagFormProps> = ({
  open,
  onOpenChange,
  onFinish,
  currentRow,
}) => {
  const formRef = useRef<ProFormInstance>();

  // 使用 useEffect 同步数据到表单
  useEffect(() => {
    if (currentRow && open) {
      formRef.current?.setFieldsValue({
        tagName: currentRow.tagName,
        description: currentRow.description,
      });
    }
  }, [currentRow, open]); // 依赖项不包含 formRef

  return (
    <DrawerForm
      formRef={formRef}
      title="编辑设备标签"
      open={open}
      onOpenChange={(visible) => {
        if (!visible) {
          // 关闭时重置
          formRef.current?.resetFields();
        }
        onOpenChange(visible);
      }}
      onFinish={onFinish}
      width={600}
      drawerProps={{
        destroyOnHidden: true,
      }}
    >
      <ProFormText
        name="tagName"
        label="标签名称"
        placeholder="请输入标签名称"
        rules={[
          DEVICE_TAG_FORM_RULES.TAG_NAME.REQUIRED,
          DEVICE_TAG_FORM_RULES.TAG_NAME.MAX_LENGTH,
        ]}
      />
      <ProFormTextArea
        name="description"
        label="描述"
        placeholder="请输入标签描述（可选）"
        rows={3}
        rules={[DEVICE_TAG_FORM_RULES.DESCRIPTION.MAX_LENGTH]}
      />
    </DrawerForm>
  );
};

export default EditDeviceTagForm;
