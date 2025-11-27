import { getOrganizationOptions } from "@/services/organization";
import { getUserOptions } from "@/services/user";
import type { ProFormInstance } from "@ant-design/pro-components";
import {
  DrawerForm,
  ProFormDigit,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
} from "@ant-design/pro-components";
import { useModel } from "@umijs/max";
import React, { useRef } from "react";
import {
  DEFAULT_VALUES,
  FORM_RULES,
  ORGANIZATION_FORM_STATUS_OPTIONS,
  ORGANIZATION_TYPE_OPTIONS,
} from "../constants";

interface CreateOrganizationFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: API.CreateOrganizationRequest) => Promise<boolean>;
  parentId?: string; // 父组织ID，用于指定在某个组织下创建子组织
}

/**
 * 创建组织表单组件
 */
const CreateOrganizationForm: React.FC<CreateOrganizationFormProps> = ({
  open,
  onOpenChange,
  onFinish,
  parentId,
}) => {
  const formRef = useRef<ProFormInstance>();
  const { initialState } = useModel("@@initialState");

  // 定义初始值常量
  const INITIAL_VALUES = {
    type: DEFAULT_VALUES.ORG_TYPE,
    status: DEFAULT_VALUES.ORG_STATUS,
    sortOrder: DEFAULT_VALUES.SORT_ORDER,
  };

  return (
    <DrawerForm
      title="新建组织"
      width={600}
      open={open}
      formRef={formRef}
      onOpenChange={(visible) => {
        if (!visible) {
          formRef.current?.resetFields();
        }
        onOpenChange(visible);
      }}
      onFinish={async (values: any) => {
        const formData = {
          ...values,
          parentId: parentId || values.parentId,
        } as API.CreateOrganizationRequest;
        return await onFinish(formData);
      }}
      initialValues={INITIAL_VALUES}
      layout="vertical"
      drawerProps={{
        destroyOnHidden: true,
      }}
    >
      <ProFormText
        name="code"
        label="组织代码"
        placeholder="请输入组织代码"
        rules={[FORM_RULES.REQUIRED_ORG_CODE, FORM_RULES.ORG_CODE_PATTERN]}
        extra="组织代码用于系统内部标识，只能包含字母、数字、下划线和连字符"
      />

      <ProFormText
        name="name"
        label="组织名称"
        placeholder="请输入组织名称"
        rules={[FORM_RULES.REQUIRED_ORG_NAME, FORM_RULES.ORG_NAME_LENGTH]}
      />

      <ProFormSelect
        name="type"
        label="组织类型"
        placeholder="请选择组织类型"
        options={ORGANIZATION_TYPE_OPTIONS}
        rules={[{ required: true, message: "请选择组织类型!" }]}
      />

      {/* 只有在未指定parentId时才显示父组织选择 */}
      {!parentId && (
        <ProFormSelect
          name="parentId"
          label="父组织"
          placeholder="请选择父组织（留空则创建为根组织）"
          request={async () => {
            try {
              const response = await getOrganizationOptions({});
              if (response.code === 0 && response.data?.list) {
                return response.data.list.map((org) => ({
                  label: `${org.name} (${org.code})`,
                  value: org.id,
                }));
              }
              return [];
            } catch (error) {
              console.error("获取组织选项失败:", error);
              return [];
            }
          }}
        />
      )}

      <ProFormSelect
        name="managerId"
        label="负责人"
        placeholder="请选择负责人（可选）"
        showSearch
        request={async (params) => {
          try {
            const response = await getUserOptions({
              keyword: params.keyWords,
              limit: 50,
              tenantId: initialState?.currentUser?.tenantInfo?.tenantId,
            });
            if (response.code === 0 && response.data?.list) {
              return response.data.list.map((user) => ({
                label: `${user.name} (${user.userName})`,
                value: user.id,
              }));
            }
            return [];
          } catch (error) {
            console.error("获取用户选项失败:", error);
            return [];
          }
        }}
        debounceTime={300}
      />

      <ProFormDigit
        name="sortOrder"
        label="排序"
        placeholder="请输入排序值"
        fieldProps={{
          precision: 0,
          min: 0,
          max: 999999,
        }}
        extra="数值越小排序越靠前"
      />

      <ProFormSelect
        name="status"
        label="状态"
        placeholder="请选择状态"
        options={ORGANIZATION_FORM_STATUS_OPTIONS}
        rules={[{ required: true, message: "请选择状态!" }]}
      />

      <ProFormTextArea
        name="description"
        label="描述"
        placeholder="请输入组织描述"
        rules={[FORM_RULES.DESCRIPTION_LENGTH]}
        fieldProps={{
          maxLength: 500,
          rows: 4,
          showCount: true,
        }}
      />
    </DrawerForm>
  );
};

export default CreateOrganizationForm;
