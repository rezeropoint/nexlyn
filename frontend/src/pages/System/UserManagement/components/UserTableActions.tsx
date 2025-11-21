import {
  DeleteOutlined,
  DisconnectOutlined,
  EditOutlined,
  KeyOutlined,
  LinkOutlined,
  LockOutlined,
  MoreOutlined,
  SyncOutlined,
  UnlockOutlined,
} from "@ant-design/icons";
import { Button, Dropdown, Modal, Space } from "antd";
import React from "react";
// API 类型通过全局声明获取
import { useUserPermissions } from "../hooks/useUserPermissions";

interface UserTableActionsProps {
  record: API.UserBrief;
  onEdit: (record: API.UserBrief) => void;
  onDelete: (record: API.UserBrief) => void;
  onToggleStatus: (record: API.UserBrief) => void;
  onResetPassword: (record: API.UserBrief) => void;
  onSyncUser: (record: API.UserBrief) => Promise<void>;
  onBindUser: (record: API.UserBrief) => void;
  onUnbindUser: (record: API.UserBrief) => Promise<void>;
  syncStatus?: boolean;
}

/**
 * 用户表格操作列组件
 * 负责渲染表格中每行的操作按钮
 */
const UserTableActions: React.FC<UserTableActionsProps> = ({
  record,
  onEdit,
  onDelete,
  onToggleStatus,
  onResetPassword,
  onSyncUser,
  onBindUser,
  onUnbindUser,
  syncStatus,
}) => {
  const permissions = useUserPermissions();

  // 同步处理
  const handleSync = async () => {
    await onSyncUser(record);
  };

  // 绑定处理
  const handleBind = () => {
    onBindUser(record);
  };

  // 解绑处理
  const handleUnbind = async () => {
    await onUnbindUser(record);
  };

  // 构建下拉菜单项
  const menuItems = [
    {
      key: "edit",
      label: "编辑信息",
      icon: <EditOutlined />,
      onClick: () => onEdit(record),
    },
    {
      key: "resetPassword",
      label: "重置密码",
      icon: <KeyOutlined />,
      onClick: () => onResetPassword(record),
    },
    // 未同步时显示同步和绑定选项
    ...(!syncStatus
      ? [
          {
            key: "sync",
            label: "同步到Skylark",
            icon: <SyncOutlined />,
            onClick: handleSync,
          },
          {
            key: "bind",
            label: "绑定到Skylark",
            icon: <LinkOutlined />,
            onClick: handleBind,
          },
        ]
      : []),
    // 已同步时只显示解绑选项
    ...(syncStatus
      ? [
          {
            key: "unbind",
            label: "解除Skylark绑定",
            icon: <DisconnectOutlined />,
            onClick: handleUnbind,
          },
        ]
      : []),
    {
      key: "delete",
      label: "删除用户",
      icon: <DeleteOutlined />,
      danger: true,
      onClick: () => {
        Modal.confirm({
          title: "确定要删除这个用户吗？",
          content: "删除后将无法恢复",
          okText: "确定",
          cancelText: "取消",
          onOk: () => onDelete(record),
        });
      },
    },
  ];

  // 删除确认处理
  const handleDeleteConfirm = () => {
    Modal.confirm({
      title: "确定要删除这个用户吗？",
      content: "删除后将无法恢复",
      okText: "确定",
      cancelText: "取消",
      onOk: () => onDelete(record),
    });
  };

  return (
    <Space size="small">
      {/* 状态切换按钮 */}
      {permissions.canUpdate && permissions.canOperateUser(record) && (
        <Button
          type="link"
          size="small"
          icon={
            record.status === "active" ? <LockOutlined /> : <UnlockOutlined />
          }
          onClick={() => onToggleStatus(record)}
          title={record.status === "active" ? "禁用用户" : "启用用户"}
        >
          {record.status === "active" ? "禁用" : "启用"}
        </Button>
      )}

      {/* 删除按钮 */}
      {permissions.canDelete && permissions.canOperateUser(record) && (
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

      {/* 更多操作下拉菜单 */}
      {permissions.canUpdate && permissions.canOperateUser(record) && (
        <Dropdown
          menu={{
            items: menuItems
              .filter((item) => {
                // 根据权限过滤菜单项
                if (item.key === "delete") return false; // 删除按钮已经独立，不在更多菜单中显示
                if (item.key === "edit" && !permissions.canOperateUser(record))
                  return false;
                if (
                  item.key === "resetPassword" &&
                  !permissions.canOperateUser(record)
                )
                  return false;
                return true;
              })
              .map((item) => ({
                ...item,
                onClick: undefined, // 移除onClick，在menu的onClick中统一处理
              })),
            onClick: ({ key }) => {
              const item = menuItems.find((i) => i.key === key);
              if (item && "onClick" in item && item.onClick) {
                item.onClick();
              }
            },
          }}
          trigger={["click"]}
        >
          <Button type="link" size="small" icon={<MoreOutlined />}>
            更多
          </Button>
        </Dropdown>
      )}
    </Space>
  );
};

export default UserTableActions;
