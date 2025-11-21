import type { ProFormInstance } from "@ant-design/pro-components";
import {
  ModalForm,
  ProFormSelect,
  ProFormSwitch,
  ProFormText,
} from "@ant-design/pro-components";
import React, { useEffect, useRef } from "react";
import {
  ORGANIZATION_STATUS_OPTIONS,
  RELATION_TYPE_OPTIONS,
} from "../constants";

interface EditMemberFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: {
    relationType?: string;
    positionTitle?: string;
    isPrimary?: boolean;
    status?: string;
  }) => Promise<boolean>;
  currentMember?: API.UserOrgRelation;
  organization?: API.Organization | API.OrganizationBrief;
}

/**
 * 编辑组织成员表单组件
 */
const EditMemberForm: React.FC<EditMemberFormProps> = ({
  open,
  onOpenChange,
  onFinish,
  currentMember,
  organization,
}) => {
  const formRef = useRef<ProFormInstance>();

  // 使用useEffect同步表单数据
  useEffect(() => {
    if (open && currentMember) {
      formRef.current?.setFieldsValue({
        relationType: currentMember.relationType,
        positionTitle: currentMember.positionTitle,
        isPrimary: currentMember.isPrimary,
        status: currentMember.status,
      });
    }
  }, [open, currentMember]);

  if (!currentMember || !organization) return null;

  return (
    <ModalForm
      title={`编辑成员 ${currentMember.name} 在 ${organization.name} 的信息`}
      open={open}
      formRef={formRef}
      onOpenChange={(visible) => {
        if (!visible) {
          formRef.current?.resetFields();
        }
        onOpenChange(visible);
      }}
      onFinish={async (values) => {
        return await onFinish(values);
      }}
      layout="vertical"
      width={600}
      modalProps={{
        destroyOnHidden: true,
      }}
    >
      {/* 显示用户基本信息 */}
      <ProFormText
        name="userName"
        label="用户名"
        disabled
        initialValue={currentMember.userName}
      />

      <ProFormText
        name="name"
        label="姓名"
        disabled
        initialValue={currentMember.name}
      />

      <ProFormText
        name="email"
        label="邮箱"
        disabled
        initialValue={currentMember.email}
      />

      <ProFormSelect
        name="relationType"
        label="关系类型"
        placeholder="请选择关系类型"
        options={RELATION_TYPE_OPTIONS}
        rules={[{ required: true, message: "请选择关系类型!" }]}
        tooltip="修改关系类型会影响用户在此组织的权限"
      />

      <ProFormText
        name="positionTitle"
        label="职位标题"
        placeholder="请输入用户在此组织的职位标题（可选）"
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

      <ProFormSelect
        name="status"
        label="关系状态"
        placeholder="请选择关系状态"
        options={ORGANIZATION_STATUS_OPTIONS.filter(
          (option) => option.value !== "deleted"
        )}
        rules={[{ required: true, message: "请选择关系状态!" }]}
        tooltip="停用状态的成员无法访问组织相关资源"
      />

      <ProFormText
        name="joinedAt"
        label="加入时间"
        disabled
        initialValue={currentMember.joinedAt}
      />
    </ModalForm>
  );
};

export default EditMemberForm;
