import type { ProFormInstance } from "@ant-design/pro-components";
import {
  DrawerForm,
  ProFormText,
  ProFormTextArea,
} from "@ant-design/pro-components";
import React, { useEffect, useRef } from "react";
import { FORM_RULES } from "../constants";

interface EditTagFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: API.UpdateTagRequest) => Promise<boolean>;
  currentRow?: API.TagDefinition;
}

/**
 * 编辑标签表单组件
 */
const EditTagForm: React.FC<EditTagFormProps> = ({
  open,
  onOpenChange,
  onFinish,
  currentRow,
}) => {
  const formRef = useRef<ProFormInstance>();

  // 使用useEffect同步表单数据
  useEffect(() => {
    if (open && currentRow) {
      formRef.current?.setFieldsValue(currentRow);
    }
  }, [open, currentRow]);

  return (
    <DrawerForm<API.UpdateTagRequest>
      title="编辑标签"
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

export default EditTagForm;
