import type { ProFormInstance } from '@ant-design/pro-components';
import { ModalForm, ProFormDigit, ProFormText } from '@ant-design/pro-components';
import { Alert, App } from 'antd';
import React, { useRef } from 'react';
import { bindUser } from '@/services/sync';

interface BindUserFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  currentUser?: API.User | API.UserBrief;
  onSuccess: () => void;
}

/**
 * 用户绑定表单组件
 * 将本地用户绑定到已存在的 Skylark 用户
 */
const BindUserForm: React.FC<BindUserFormProps> = ({
  open,
  onOpenChange,
  currentUser,
  onSuccess,
}) => {
  const formRef = useRef<ProFormInstance>();
  const { message } = App.useApp();

  const handleFinish = async (values: { remoteUserId: number }) => {
    if (!currentUser) {
      message.error('用户信息不存在');
      return false;
    }

    try {
      const response = await bindUser({ id: currentUser.id }, { remoteUserId: values.remoteUserId });
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
        description="绑定后将在同步时关联到指定的 Skylark 用户，而非创建新用户。请确保远程用户 ID 存在且未被其他本地用户绑定。"
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      <ProFormText
        label="用户姓名"
        fieldProps={{
          value: currentUser?.name || '-',
        }}
        disabled
      />

      <ProFormText
        label="用户名"
        fieldProps={{
          value: currentUser?.userName || '-',
        }}
        disabled
      />

      <ProFormDigit
        name="remoteUserId"
        label="远程用户 ID"
        placeholder="请输入 Skylark 系统中的用户 ID"
        rules={[
          { required: true, message: '请输入远程用户 ID' },
          { type: 'number', min: 1, message: '用户 ID 必须大于 0' },
        ]}
        fieldProps={{
          precision: 0,
          style: { width: '100%' },
        }}
      />
    </ModalForm>
  );
};

export default BindUserForm;
