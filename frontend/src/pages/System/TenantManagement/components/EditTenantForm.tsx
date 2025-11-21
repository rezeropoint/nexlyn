import type { Tenant, UpdateTenantRequest } from "@/services/tenant";
import {
  DrawerForm,
  ProFormDatePicker,
  ProFormDigit,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
  type ProFormInstance,
} from "@ant-design/pro-components";
import dayjs from "dayjs";
import React, { useEffect, useRef } from "react";
import { FORM_RULES } from "../constants";
import { useTenantOptions } from "../hooks/useTenantOptions";

interface EditTenantFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: UpdateTenantRequest) => Promise<boolean>;
  currentRow?: Tenant;
}

/**
 * 编辑租户表单组件
 */
const EditTenantForm: React.FC<EditTenantFormProps> = ({
  open,
  onOpenChange,
  onFinish,
  currentRow,
}) => {
  const formRef = useRef<ProFormInstance>();
  const { tagOptions, loadTagOptions } = useTenantOptions();

  // 使用 useEffect 同步数据到表单
  useEffect(() => {
    if (currentRow && open) {
      formRef.current?.setFieldsValue({
        id: currentRow.id,
        tenantName: currentRow.tenantName,
        description: currentRow.description,
        contactEmail: currentRow.contactEmail,
        tags: currentRow.tags,
        maxUsers: currentRow.maxUsers,
        expiresAt: currentRow.expiresAt ? dayjs(currentRow.expiresAt) : undefined,
      });
    }
  }, [currentRow, open]); // 依赖项不包含 formRef

  return (
    <DrawerForm
      formRef={formRef}
      title="编辑租户"
      width={600}
      open={open}
      onOpenChange={(visible) => {
        if (!visible) {
          // 关闭时重置
          formRef.current?.resetFields();
        }
        onOpenChange(visible);
      }}
      onFinish={onFinish}
      drawerProps={{
        destroyOnHidden: true,
      }}
    >
      <ProFormText name="id" label="租户ID" disabled extra="租户ID不可修改" />

      <ProFormText
        name="tenantName"
        label="租户名称"
        placeholder="请输入租户名称"
        rules={[FORM_RULES.REQUIRED_TENANT_NAME]}
      />

      <ProFormTextArea
        name="description"
        label="租户描述"
        placeholder="请输入租户描述"
        fieldProps={{
          maxLength: 500,
          rows: 4,
          showCount: true,
        }}
      />

      <ProFormText
        name="contactEmail"
        label="联系邮箱"
        placeholder="请输入联系邮箱"
        rules={[FORM_RULES.REQUIRED_CONTACT_EMAIL, FORM_RULES.EMAIL_FORMAT]}
      />

      <ProFormSelect
        name="tags"
        label="租户标签"
        mode="tags"
        placeholder="请选择或输入标签"
        options={tagOptions}
        fieldProps={{
          onDropdownVisibleChange: (open) => {
            if (open && tagOptions.length === 0) {
              loadTagOptions();
            }
          },
        }}
        extra="可以选择已有标签或输入新标签"
      />

      <ProFormDigit
        name="maxUsers"
        label="最大用户数"
        placeholder="请输入最大用户数"
        rules={[FORM_RULES.REQUIRED_MAX_USERS, FORM_RULES.MIN_MAX_USERS]}
        fieldProps={{
          min: 1,
          precision: 0,
        }}
        extra="该租户允许创建的最大用户数量"
      />

      <ProFormDatePicker
        name="expiresAt"
        label="过期时间"
        placeholder="请选择过期时间"
        fieldProps={{
          format: "YYYY-MM-DD HH:mm:ss",
          showTime: true,
          disabledDate: (current: any) =>
            current && current < dayjs().startOf("day"),
        }}
        extra="留空表示永不过期"
      />
    </DrawerForm>
  );
};

export default EditTenantForm;
