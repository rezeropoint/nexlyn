import type { ProFormInstance } from "@ant-design/pro-components";
import {
  DrawerForm,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
} from "@ant-design/pro-components";
import React, { useRef } from "react";
import { FORM_RULES } from "../constants";
// API 类型通过全局声明获取
import { useUserOptions } from "../hooks/useUserOptions";

interface CreateUserFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: API.CreateUserRequest) => Promise<boolean>;
  currentUser?: API.User;
}

/**
 * 创建用户表单组件
 */
const CreateUserForm: React.FC<CreateUserFormProps> = ({
  open,
  onOpenChange,
  onFinish,
  currentUser,
}) => {
  const formRef = useRef<ProFormInstance>();
  const { tagOptions, tenantOptions, loadTagOptions, loadTenantOptions } =
    useUserOptions();

  return (
    <DrawerForm
      title="新建用户"
      width={600}
      open={open}
      formRef={formRef}
      onOpenChange={(visible) => {
        if (!visible) {
          formRef.current?.resetFields();
        }
        onOpenChange(visible);
      }}
      onFinish={onFinish}
      drawerProps={{
        destroyOnHidden: true,
      }}
    >
      <ProFormText
        name="name"
        label="用户姓名"
        placeholder="请输入用户姓名"
        rules={[FORM_RULES.REQUIRED_NAME]}
      />

      <ProFormText
        name="userName"
        label="用户名"
        placeholder="请输入用户名"
        rules={[FORM_RULES.REQUIRED_USERNAME]}
      />

      <ProFormText
        name="email"
        label="邮箱"
        placeholder="请输入邮箱"
        rules={[FORM_RULES.REQUIRED_EMAIL, FORM_RULES.EMAIL_FORMAT]}
      />

      <ProFormText name="phone" label="电话" placeholder="请输入电话号码" />

      <ProFormText name="title" label="职位" placeholder="请输入职位" />

      {/* 租户选择框 - 只有超级管理员可以选择，其他用户固定为当前租户 */}
      {currentUser?.roles?.includes("super_admin") ? (
        <ProFormSelect
          name="tenantId"
          label="所属租户"
          placeholder="请选择租户"
          rules={[FORM_RULES.REQUIRED_TENANT]}
          options={tenantOptions}
          fieldProps={{
            onDropdownVisibleChange: (open) => {
              if (open && tenantOptions.length === 0) {
                loadTenantOptions();
              }
            },
          }}
        />
      ) : (
        <ProFormText
          name="tenantId"
          label="所属租户"
          disabled
          initialValue={currentUser?.tenantInfo?.tenantId || ""}
          fieldProps={{
            value: currentUser?.tenantInfo?.tenantName || "当前租户",
          }}
        />
      )}

      <ProFormText name="country" label="国家" placeholder="请输入国家" />

      <ProFormText name="address" label="地址" placeholder="请输入地址" />

      <ProFormSelect
        name="tagIds"
        label="用户标签"
        mode="multiple"
        placeholder="请选择用户标签"
        options={tagOptions.map((tag) => ({ label: tag.label, value: tag.id }))}
        fieldProps={{
          onDropdownVisibleChange: (open) => {
            if (open && tagOptions.length === 0) {
              loadTagOptions();
            }
          },
        }}
      />

      <ProFormText
        name={["geographic", "province", "label"]}
        label="省份"
        placeholder="请输入省份"
      />

      <ProFormText
        name={["geographic", "city", "label"]}
        label="城市"
        placeholder="请输入城市"
      />

      <ProFormTextArea
        name="signature"
        label="个人签名"
        placeholder="请输入个人签名"
      />
    </DrawerForm>
  );
};

export default CreateUserForm;
