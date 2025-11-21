import { getRoleList } from "@/services/permission/role";
import type { CreateTenantRequest } from "@/services/tenant";
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
import React, { useCallback, useEffect, useRef, useState } from "react";
import { DEFAULT_VALUES, FORM_RULES } from "../constants";
import { useTenantOptions } from "../hooks/useTenantOptions";

interface CreateTenantFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: CreateTenantRequest) => Promise<boolean>;
}

/**
 * 创建租户表单组件
 */
const CreateTenantForm: React.FC<CreateTenantFormProps> = ({
  open,
  onOpenChange,
  onFinish,
}) => {
  const formRef = useRef<ProFormInstance>();
  const { tagOptions, loadTagOptions } = useTenantOptions();

  // 初始状态值
  const INITIAL_ROLE_OPTIONS: { label: string; value: string }[] = [];
  const [roleOptions, setRoleOptions] = useState(INITIAL_ROLE_OPTIONS);
  const [roleLoading, setRoleLoading] = useState(false);

  // 加载角色选项
  const loadRoleOptions = useCallback(async () => {
    try {
      setRoleLoading(true);
      const response = await getRoleList({ pageSize: 100 }); // 获取所有角色
      if (response.code === 0 && response.data) {
        const options = response.data.map((role) => ({
          label: `${role.roleName} (${role.roleKey})`,
          value: role.roleKey,
        }));
        setRoleOptions(options);
      }
    } catch (error) {
      console.error("加载角色选项失败:", error);
    } finally {
      setRoleLoading(false);
    }
  }, []);

  // 当表单打开时加载角色选项
  useEffect(() => {
    if (open && roleOptions.length === 0) {
      loadRoleOptions();
    }
  }, [open, roleOptions.length, loadRoleOptions]);

  return (
    <DrawerForm
      formRef={formRef}
      title="新建租户"
      width={600}
      open={open}
      onOpenChange={(visible) => {
        if (!visible) {
          // 关闭时重置（顺序：formRef → state → callback）
          formRef.current?.resetFields();
          setRoleOptions(INITIAL_ROLE_OPTIONS);
          setRoleLoading(false);
        }
        onOpenChange(visible);
      }}
      onFinish={onFinish}
      drawerProps={{
        destroyOnHidden: true,
      }}
    >
      <ProFormText
        name="tenantKey"
        label="租户标识"
        placeholder="请输入租户标识"
        rules={[FORM_RULES.REQUIRED_TENANT_KEY]}
        extra="租户的业务标识符，创建后不可修改"
      />

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

      <ProFormSelect
        name="adminRoleKey"
        label="管理员角色"
        placeholder="请选择租户管理员角色"
        options={roleOptions}
        rules={[{ required: true, message: "请选择租户管理员角色" }]}
        fieldProps={{
          loading: roleLoading,
          onDropdownVisibleChange: (open) => {
            if (open && roleOptions.length === 0) {
              loadRoleOptions();
            }
          },
          showSearch: true,
          filterOption: (input, option) =>
            (option?.label ?? "").toLowerCase().includes(input.toLowerCase()),
        }}
        extra="选择为该租户管理员分配的角色，角色必须包含必要的用户管理权限且不包含跨租户权限"
      />

      <ProFormDigit
        name="maxUsers"
        label="最大用户数"
        placeholder="请输入最大用户数"
        rules={[FORM_RULES.REQUIRED_MAX_USERS, FORM_RULES.MIN_MAX_USERS]}
        initialValue={DEFAULT_VALUES.MAX_USERS}
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

export default CreateTenantForm;
