import type { Tenant, UpdateTenantStatusRequest } from "@/services/tenant";
import { ModalForm, ProFormSelect } from "@ant-design/pro-components";
import React from "react";
import { TENANT_STATUS_OPTIONS } from "../constants";

interface StatusManageModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: UpdateTenantStatusRequest) => Promise<boolean>;
  currentRow?: Tenant;
}

/**
 * 租户状态管理模态框组件
 */
const StatusManageModal: React.FC<StatusManageModalProps> = ({
  open,
  onOpenChange,
  onFinish,
  currentRow,
}) => {
  return (
    <ModalForm
      title="租户状态管理"
      open={open}
      onOpenChange={onOpenChange}
      onFinish={onFinish}
      initialValues={{
        status: currentRow?.status,
      }}
      modalProps={{
        destroyOnHidden: true,
      }}
    >
      <ProFormSelect
        name="status"
        label="租户状态"
        placeholder="请选择租户状态"
        options={TENANT_STATUS_OPTIONS}
      />
    </ModalForm>
  );
};

export default StatusManageModal;
