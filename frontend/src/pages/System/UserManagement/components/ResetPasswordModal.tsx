import { ModalForm, ProFormText } from "@ant-design/pro-components";
import React from "react";
import { FORM_RULES } from "../constants";

interface ResetPasswordModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: { password: string }) => Promise<boolean>;
}

/**
 * 重置密码模态框组件
 */
const ResetPasswordModal: React.FC<ResetPasswordModalProps> = ({
  open,
  onOpenChange,
  onFinish,
}) => {
  return (
    <ModalForm
      title="重置密码"
      open={open}
      onOpenChange={onOpenChange}
      onFinish={onFinish}
      modalProps={{
        destroyOnHidden: true,
      }}
    >
      <ProFormText.Password
        name="password"
        label="新密码"
        placeholder="请输入新密码"
        rules={[FORM_RULES.REQUIRED_PASSWORD, FORM_RULES.PASSWORD_LENGTH]}
      />
    </ModalForm>
  );
};

export default ResetPasswordModal;
