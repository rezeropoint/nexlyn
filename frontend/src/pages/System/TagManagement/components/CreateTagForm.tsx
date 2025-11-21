import type { ProFormInstance } from "@ant-design/pro-components";
import {
  DrawerForm,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
} from "@ant-design/pro-components";
import React, { useRef } from "react";
import { FORM_RULES } from "../constants";

interface CreateTagFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: API.CreateTagRequest) => Promise<boolean>;
}

/**
 * 创建标签表单组件
 */
const CreateTagForm: React.FC<CreateTagFormProps> = ({
  open,
  onOpenChange,
  onFinish,
}) => {
  const formRef = useRef<ProFormInstance>();

  return (
    <DrawerForm<API.CreateTagRequest>
      title="创建标签"
      open={open}
      formRef={formRef}
      onOpenChange={(visible) => {
        if (!visible) {
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
      <ProFormSelect
        name="scope"
        label="作用域"
        placeholder="请选择标签作用域"
        options={[
          { label: "用户标签", value: "user" },
          { label: "租户标签", value: "tenant" },
        ]}
        rules={[FORM_RULES.SCOPE.REQUIRED]}
      />
      <ProFormText
        name="label"
        label="标签名称"
        placeholder="请输入标签名称"
        rules={[FORM_RULES.LABEL.REQUIRED, FORM_RULES.LABEL.MAX_LENGTH]}
      />
      <ProFormTextArea
        name="description"
        label="描述"
        placeholder="请输入标签描述（可选）"
        rows={3}
        rules={[FORM_RULES.DESCRIPTION.MAX_LENGTH]}
      />
    </DrawerForm>
  );
};

export default CreateTagForm;
