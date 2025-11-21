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
import React, { useEffect, useRef } from "react";
import {
  FORM_RULES,
  ORGANIZATION_FORM_STATUS_OPTIONS,
  ORGANIZATION_TYPE_OPTIONS,
} from "../constants";

interface EditOrganizationFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: API.UpdateOrganizationRequest) => Promise<boolean>;
  currentRow?: API.Organization;
}

/**
 * 编辑组织表单组件
 */
const EditOrganizationForm: React.FC<EditOrganizationFormProps> = ({
  open,
  onOpenChange,
  onFinish,
  currentRow,
}) => {
  const formRef = useRef<ProFormInstance>();
  const { initialState } = useModel("@@initialState");

  // 使用useEffect同步表单数据
  useEffect(() => {
    if (open && currentRow) {
      formRef.current?.setFieldsValue({
        name: currentRow.name,
        type: currentRow.type,
        description: currentRow.description,
        managerId: currentRow.managerId,
        sortOrder: currentRow.sortOrder,
        status: currentRow.status,
      });
    }
  }, [open, currentRow]);

  if (!currentRow) return null;

  return (
    <DrawerForm
      title="编辑组织"
      width={600}
      open={open}
      formRef={formRef}
      onOpenChange={(visible) => {
        if (!visible) {
          formRef.current?.resetFields();
        }
        onOpenChange(visible);
      }}
      onFinish={async (values) => {
        const formData: API.UpdateOrganizationRequest = {
          id: currentRow.id,
          ...values,
        };
        return await onFinish(formData);
      }}
      layout="vertical"
      drawerProps={{
        destroyOnHidden: true,
      }}
    >
      <ProFormText
        name="code"
        label="组织代码"
        disabled
        initialValue={currentRow.code}
        extra="组织代码创建后不可修改"
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

      <ProFormText
        name="path"
        label="组织路径"
        disabled
        initialValue={currentRow.path}
        extra="组织在层级结构中的路径，系统自动生成"
      />

      <ProFormText
        name="level"
        label="组织层级"
        disabled
        initialValue={`第 ${currentRow.level} 层`}
      />

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
                value: user.userKey,
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

      <ProFormText
        name="memberCount"
        label="成员数量"
        disabled
        initialValue={`${currentRow.memberCount} 人`}
      />

      <ProFormText
        name="createdAt"
        label="创建时间"
        disabled
        initialValue={currentRow.createdAt}
      />

      <ProFormText
        name="updatedAt"
        label="更新时间"
        disabled
        initialValue={currentRow.updatedAt}
      />
    </DrawerForm>
  );
};

export default EditOrganizationForm;
