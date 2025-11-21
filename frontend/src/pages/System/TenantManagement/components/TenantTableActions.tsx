import type { Tenant } from "@/services/tenant";
import {
  CheckCircleOutlined,
  DeleteOutlined,
  EditOutlined,
  StopOutlined,
} from "@ant-design/icons";
import { Button, Modal, Space } from "antd";
import React from "react";
import { useTenantPermissions } from "../hooks/useTenantPermissions";

interface TenantTableActionsProps {
  record: Tenant;
  onEdit: (record: Tenant) => void;
  onDelete: (record: Tenant) => void;
  onToggleStatus: (record: Tenant) => void;
}

/**
 * 租户表格操作列组件
 * 负责渲染表格中每行的操作按钮
 */
const TenantTableActions: React.FC<TenantTableActionsProps> = ({
  record,
  onEdit,
  onDelete,
  onToggleStatus,
}) => {
  const permissions = useTenantPermissions();

  // 状态切换处理
  const handleToggleStatus = () => {
    const actionText = record.status === "active" ? "禁用" : "启用";
    Modal.confirm({
      title: `确定要${actionText}这个租户吗？`,
      content:
        record.status === "active"
          ? "禁用后该租户将无法正常使用"
          : "启用后该租户可以正常使用",
      okText: "确定",
      cancelText: "取消",
      onOk: () => onToggleStatus(record),
    });
  };

  // 删除确认处理
  const handleDeleteConfirm = () => {
    Modal.confirm({
      title: "确定要删除这个租户吗？",
      content: "删除后将无法恢复，该租户下的所有数据都将被删除",
      okText: "确定",
      cancelText: "取消",
      onOk: () => onDelete(record),
    });
  };

  const canOperate = permissions.canOperateTenant(record);

  return (
    <Space size="small">
      {/* 状态切换按钮 */}
      {permissions.canUpdate && canOperate && (
        <Button
          type="link"
          size="small"
          icon={
            record.status === "active" ? (
              <StopOutlined />
            ) : (
              <CheckCircleOutlined />
            )
          }
          onClick={handleToggleStatus}
          title={record.status === "active" ? "禁用租户" : "启用租户"}
        >
          {record.status === "active" ? "禁用" : "启用"}
        </Button>
      )}

      {/* 删除按钮 */}
      {permissions.canDelete && canOperate && (
        <Button
          type="link"
          size="small"
          danger
          icon={<DeleteOutlined />}
          onClick={handleDeleteConfirm}
        >
          删除
        </Button>
      )}

      {/* 编辑按钮 */}
      {permissions.canUpdate && canOperate && (
        <Button
          type="link"
          size="small"
          icon={<EditOutlined />}
          onClick={() => onEdit(record)}
        >
          编辑
        </Button>
      )}
    </Space>
  );
};

export default TenantTableActions;
