import { getUserOptions } from "@/services/user";
import type { ProFormInstance } from "@ant-design/pro-components";
import {
  ModalForm,
  ProFormSelect,
  ProFormSwitch,
  ProFormText,
} from "@ant-design/pro-components";
import React, { useRef } from "react";
import {
  DEFAULT_VALUES,
  FORM_RULES,
  RELATION_TYPE_OPTIONS,
} from "../constants";

interface AddMemberFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: {
    userId: string;
    relationType?: string;
    positionTitle?: string;
    isPrimary?: boolean;
  }) => Promise<boolean>;
  organization?: API.Organization | API.OrganizationBrief;
}

/**
 * 添加组织成员表单组件
 */
const AddMemberForm: React.FC<AddMemberFormProps> = ({
  open,
  onOpenChange,
  onFinish,
  organization,
}) => {
  const formRef = useRef<ProFormInstance>();

  // 定义初始值常量
  const INITIAL_VALUES = {
    relationType: DEFAULT_VALUES.RELATION_TYPE,
    isPrimary: DEFAULT_VALUES.IS_PRIMARY,
  };

  if (!organization) return null;

  return (
    <ModalForm
      title={`添加成员到 ${organization.name}`}
      open={open}
      formRef={formRef}
      onOpenChange={(visible) => {
        if (!visible) {
          formRef.current?.resetFields();
        }
        onOpenChange(visible);
      }}
      onFinish={async (values: Record<string, any>) => {
        if (!values.userId) {
          throw new Error("请选择用户");
        }
        const memberData = {
          userId: values.userId,
          relationType: values.relationType,
          positionTitle: values.positionTitle,
          isPrimary: values.isPrimary,
        };
        return await onFinish(memberData);
      }}
      initialValues={INITIAL_VALUES}
      layout="vertical"
      width={600}
      modalProps={{
        destroyOnHidden: true,
      }}
    >
      <ProFormSelect
        name="userId"
        label="选择用户"
        placeholder="请搜索并选择要添加的用户"
        rules={[{ required: true, message: "请选择用户!" }]}
        showSearch
        request={async (params) => {
          try {
            const response = await getUserOptions({
              keyword: params.keyWords,
              limit: 50,
            });
            if (response.code === 0 && response.data?.list) {
              return response.data.list.map((user) => ({
                label: `${user.name} (${user.userName}) - ${user.email}`,
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
        fieldProps={{
          filterOption: false,
        }}
      />

      <ProFormSelect
        name="relationType"
        label="关系类型"
        placeholder="请选择关系类型"
        options={RELATION_TYPE_OPTIONS}
        rules={[{ required: true, message: "请选择关系类型!" }]}
      />

      <ProFormText
        name="positionTitle"
        label="职位标题"
        placeholder="请输入用户在此组织的职位标题（可选）"
        rules={[FORM_RULES.POSITION_LENGTH]}
      />

      <ProFormSwitch
        name="isPrimary"
        label="设为主组织"
        tooltip="如果设置为主组织，该用户的其他主组织关系将被自动取消"
        fieldProps={{
          checkedChildren: "是",
          unCheckedChildren: "否",
        }}
      />
    </ModalForm>
  );
};

export default AddMemberForm;
