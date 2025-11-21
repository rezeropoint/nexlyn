import { useRuleOptions } from "@/pages/System/PermissionManagement/Rules/hooks/useRuleOptions";
import { PlusOutlined } from "@ant-design/icons";
import type { ProFormInstance } from "@ant-design/pro-components";
import {
  DrawerForm,
  ProFormSelect,
  ProFormText,
} from "@ant-design/pro-components";
import { Button } from "antd";
import React, { useRef, useState } from "react";

interface AddRuleFormProps {
  onFinish: (values: API.AddPermissionRequest) => Promise<boolean>;
}

const AddRuleForm: React.FC<AddRuleFormProps> = ({ onFinish }) => {
  const formRef = useRef<ProFormInstance>();
  const [open, setOpen] = useState(false);
  const { userOptions, tenantOptions, loadUserOptions, loadTenantOptions } =
    useRuleOptions();

  // 增强的提交处理，验证用户选择的有效性
  const handleFinish = async (values: API.AddPermissionRequest) => {
    // 验证userKey是否在选项中存在（防止手动输入或其他方式绕过）
    const validUser = userOptions.find(
      (user: any) => user.userKey === values.userKey
    );
    if (!validUser) {
      console.error("Invalid user selection:", values.userKey);
      throw new Error("无效的用户选择，请从用户列表中选择");
    }

    // 验证tenantKey是否在选项中存在
    const validTenant = tenantOptions.find(
      (tenant: any) => tenant.tenantKey === values.tenantKey
    );
    if (!validTenant) {
      console.error("Invalid tenant selection:", values.tenantKey);
      throw new Error("无效的租户选择，请从租户列表中选择");
    }

    return await onFinish(values);
  };

  return (
    <DrawerForm<API.AddPermissionRequest>
      title="为用户添加直接权限"
      open={open}
      formRef={formRef}
      trigger={
        <Button type="primary" onClick={() => setOpen(true)}>
          <PlusOutlined /> 添加用户权限
        </Button>
      }
      onOpenChange={(visible) => {
        if (!visible) {
          formRef.current?.resetFields();
        }
        setOpen(visible);
      }}
      onFinish={handleFinish}
      width={600}
      drawerProps={{
        destroyOnHidden: true,
      }}
    >
      <ProFormSelect
        name="userKey"
        label="用户"
        placeholder="请选择用户"
        rules={[{ required: true, message: "请选择用户" }]}
        showSearch
        options={userOptions.map((user: any) => ({
          label: `${user.name} (${user.userName}) - ${user.email}`,
          value: user.userKey,
        }))}
        fieldProps={{
          mode: undefined, // 确保不是多选模式
          allowClear: true,
          filterOption: (input: string, option: any) =>
            (option?.label ?? "").toLowerCase().includes(input.toLowerCase()),
          onDropdownVisibleChange: (open: boolean) => {
            if (open) {
              loadUserOptions();
            }
          },
          onSearch: (value: string) => {
            loadUserOptions(value);
          },
          // 严格限制：只能选择选项，不能输入任意值
          showArrow: true,
          // 验证选中的值必须在选项中存在
          onSelect: (value: any, _option: any) => {
            // 确保选择的值在用户选项中存在
            const validUser = userOptions.find(
              (user: any) => user.userKey === value
            );
            if (!validUser) {
              console.warn("Invalid user selection:", value);
            }
          },
        }}
        tooltip="只能选择系统中已存在的用户，不能手动输入用户ID"
      />

      <ProFormSelect
        name="tenantKey"
        label="租户"
        placeholder="请选择租户"
        rules={[{ required: true, message: "请选择租户" }]}
        showSearch
        options={tenantOptions.map((tenant: any) => ({
          label: `${tenant.tenantName} (${tenant.tenantKey})`,
          value: tenant.tenantKey,
        }))}
        fieldProps={{
          filterOption: (input: string, option: any) =>
            (option?.label ?? "").toLowerCase().includes(input.toLowerCase()),
          onDropdownVisibleChange: (open: boolean) => {
            if (open) {
              loadTenantOptions();
            }
          },
          onSearch: (value: string) => {
            loadTenantOptions(value);
          },
        }}
      />

      <ProFormText
        name="resource"
        label="资源"
        placeholder="如: user, graph, infoatom"
        rules={[{ required: true, message: "请输入资源名称" }]}
        tooltip="资源类型，如：user（用户）、graph（图）、infoatom（信息原子）等"
      />

      <ProFormSelect
        name="action"
        label="操作"
        placeholder="选择操作类型"
        options={[
          { label: "读取", value: "read" },
          { label: "写入", value: "write" },
          { label: "删除", value: "delete" },
          { label: "管理", value: "admin" },
          { label: "跨租户", value: "cross_tenant" },
        ]}
        rules={[{ required: true, message: "请选择操作类型" }]}
      />
    </DrawerForm>
  );
};

export default AddRuleForm;
