import type { ProFormInstance } from '@ant-design/pro-components';
import { ModalForm, ProFormDigit, ProFormText } from '@ant-design/pro-components';
import { Alert, App } from 'antd';
import React, { useRef } from 'react';
import { bindOrganization } from '@/services/sync';

interface BindOrganizationFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  currentOrganization?: API.Organization | API.OrganizationBrief;
  onSuccess: () => void;
}

/**
 * 组织绑定表单组件
 * 将本地组织绑定到已存在的 Skylark 组织
 */
const BindOrganizationForm: React.FC<BindOrganizationFormProps> = ({
  open,
  onOpenChange,
  currentOrganization,
  onSuccess,
}) => {
  const formRef = useRef<ProFormInstance>();
  const { message } = App.useApp();

  const handleFinish = async (values: { remoteOrgId: number }) => {
    if (!currentOrganization) {
      message.error('组织信息不存在');
      return false;
    }

    try {
      const response = await bindOrganization(
        { id: currentOrganization.id },
        { remoteOrgId: values.remoteOrgId },
      );
      if (response.code === 0) {
        message.success('绑定成功');
        onSuccess();
        return true;
      } else {
        message.error(response.msg || '绑定失败');
        return false;
      }
    } catch (_error) {
      message.error('绑定失败，请稍后重试');
      return false;
    }
  };

  return (
    <ModalForm
      title="绑定到 Skylark 系统"
      width={500}
      open={open}
      formRef={formRef}
      onOpenChange={(visible) => {
        if (!visible) {
          formRef.current?.resetFields();
        }
        onOpenChange(visible);
      }}
      onFinish={handleFinish}
      modalProps={{
        destroyOnHidden: true,
      }}
    >
      <Alert
        message="提示"
        description="绑定后将在同步时关联到指定的 Skylark 组织，而非创建新组织。请确保远程组织 ID 存在且未被其他本地组织绑定。"
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      <ProFormText
        label="组织名称"
        fieldProps={{
          value: currentOrganization?.name || '-',
        }}
        disabled
      />

      <ProFormText
        label="组织编码"
        fieldProps={{
          value: currentOrganization?.code || '-',
        }}
        disabled
      />

      <ProFormDigit
        name="remoteOrgId"
        label="远程组织 ID"
        placeholder="请输入 Skylark 系统中的组织 ID"
        rules={[
          { required: true, message: '请输入远程组织 ID' },
          { type: 'number', min: 1, message: '组织 ID 必须大于 0' },
        ]}
        fieldProps={{
          precision: 0,
          style: { width: '100%' },
        }}
      />
    </ModalForm>
  );
};

export default BindOrganizationForm;
