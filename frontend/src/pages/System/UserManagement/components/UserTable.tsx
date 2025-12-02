import {
  DeleteOutlined,
  LockOutlined,
  PlusOutlined,
  UnlockOutlined,
} from "@ant-design/icons";
import {
  type ActionType,
  type ProColumns,
  ProTable,
} from "@ant-design/pro-components";
import { Button, message, Modal, Space, Tag } from "antd";
import React, { useState, useEffect } from "react";
// API 类型通过全局声明获取
import { getUserList } from "@/services/user";
import { getUserSyncStatus } from "@/services/sync";
import { formatDateTime } from "@/utils/date";
import {
  TABLE_CONFIG,
  USER_STATUS_COLORS,
  USER_STATUS_TEXT,
} from "../constants";
import { useUserPermissions } from "../hooks/useUserPermissions";
import type { UserSelectionState } from "../types";
import styles from "./UserTable.less";
import UserTableActions from "./UserTableActions";

interface UserTableProps {
  actionRef: React.MutableRefObject<ActionType | undefined>;
  selectionState: UserSelectionState;
  onSelectionChange: (state: Partial<UserSelectionState>) => void;
  onCreateUser: () => void;
  onEditUser: (record: API.UserBrief) => void;
  onDeleteUser: (record: API.UserBrief) => void;
  onToggleStatus: (record: API.UserBrief) => void;
  onResetPassword: (record: API.UserBrief) => void;
  onBatchDelete: (users: API.UserBrief[]) => void;
  onBatchToggleStatus: (
    users: API.UserBrief[],
    targetStatus: "active" | "inactive"
  ) => void;
  onSyncUser: (record: API.UserBrief) => Promise<void>;
  onBindUser: (record: API.UserBrief) => void;
  onUnbindUser: (record: API.UserBrief) => Promise<void>;
}

/**
 * 用户表格组件
 */
const UserTable: React.FC<UserTableProps> = ({
  actionRef,
  selectionState,
  onSelectionChange,
  onCreateUser,
  onEditUser,
  onDeleteUser,
  onToggleStatus,
  onResetPassword,
  onBatchDelete,
  onBatchToggleStatus,
  onSyncUser,
  onBindUser,
  onUnbindUser,
}) => {
  const permissions = useUserPermissions();
  const { selectedRowKeys, selectedRows } = selectionState;

  // 同步状态管理
  const [syncStatusMap, setSyncStatusMap] = useState<Record<string, boolean>>(
    {},
  );
  const [currentUserIds, setCurrentUserIds] = useState<string[]>([]);

  // 查询同步状态
  useEffect(() => {
    if (currentUserIds.length === 0) return;

    const fetchSyncStatus = async () => {
      try {
        const response = await getUserSyncStatus({ userIds: currentUserIds });
        if (response.code === 0 && response.data?.syncStatus) {
          setSyncStatusMap(response.data.syncStatus);
        }
      } catch (error) {
        // 静默失败，不影响主要功能
        console.error("查询同步状态失败:", error);
      }
    };

    fetchSyncStatus();
  }, [currentUserIds]);

  // 表格列定义
  const columns: ProColumns<API.UserBrief>[] = [
    {
      title: "用户名",
      dataIndex: "userName",
      width: 120,
      ellipsis: true,
      copyable: true,
      fixed: "left",
    },
    {
      title: "姓名",
      dataIndex: "name",
      width: 100,
      ellipsis: true,
    },
    {
      title: "邮箱",
      dataIndex: "email",
      width: 180,
      ellipsis: true,
      copyable: true,
    },
    {
      title: "租户",
      dataIndex: ["tenantInfo", "tenantName"],
      width: 120,
      ellipsis: true,
      render: (_, record) => record.tenantInfo?.tenantName || "未知",
    },
    {
      title: "标签",
      dataIndex: "tags",
      width: 120,
      render: (_, record: API.UserBrief) => {
        const tags = record.tags;
        if (!tags || tags.length === 0) return "-";
        return (
          <div>
            {tags.slice(0, 2).map((tag: any) => (
              <Tag key={tag.id} className={styles.tagItem}>
                {tag.label}
              </Tag>
            ))}
            {tags.length > 2 && <Tag color="default">+{tags.length - 2}</Tag>}
          </div>
        );
      },
    },
    {
      title: "创建时间",
      dataIndex: "createdAt",
      width: 160,
      valueType: "dateTime",
      search: false,
      render: (_, record: API.UserBrief) => formatDateTime(record.createdAt),
    },
    {
      title: "更新时间",
      dataIndex: "updatedAt",
      width: 160,
      valueType: "dateTime",
      search: false,
      render: (_, record: API.UserBrief) => formatDateTime(record.updatedAt),
    },
    {
      title: "状态",
      dataIndex: "status",
      width: 80,
      render: (status) => (
        <Tag
          color={USER_STATUS_COLORS[status as keyof typeof USER_STATUS_COLORS]}
        >
          {USER_STATUS_TEXT[status as keyof typeof USER_STATUS_TEXT]}
        </Tag>
      ),
    },
    {
      title: "同步状态",
      dataIndex: "syncStatus",
      width: 150,
      search: false,
      render: (_, record) => {
        const isSynced = syncStatusMap[record.id];
        if (isSynced === undefined) {
          return <Tag color="default">未查询</Tag>;
        }
        if (isSynced) {
          return <Tag color="success">已同步</Tag>;
        }
        return <Tag color="default">未同步</Tag>;
      },
    },
    {
      title: "操作",
      valueType: "option",
      key: "option",
      width: 260,
      fixed: "right",
      render: (_, record) => (
        <UserTableActions
          record={record}
          onEdit={onEditUser}
          onDelete={onDeleteUser}
          onToggleStatus={onToggleStatus}
          onResetPassword={onResetPassword}
          onSyncUser={onSyncUser}
          onBindUser={onBindUser}
          onUnbindUser={onUnbindUser}
          syncStatus={syncStatusMap[record.id]}
        />
      ),
    },
  ];

  // 工具栏渲染
  const toolBarRender = () => {
    const hasSelected = selectedRowKeys.length > 0;
    const canBatchDelete =
      hasSelected &&
      selectedRows.every(
        (user) => permissions.canDelete && permissions.canOperateUser(user)
      );
    const canBatchUpdate =
      hasSelected &&
      selectedRows.every(
        (user) => permissions.canUpdate && permissions.canOperateUser(user)
      );

    return [
      // 多选操作按钮
      hasSelected && (
        <Space key="batch-actions">
          {canBatchUpdate && (
            <Button
              onClick={() => {
                Modal.confirm({
                  title: `确定要禁用选中的 ${selectedRows.length} 个用户吗？`,
                  content: "禁用后用户将无法登录",
                  okText: "确定",
                  cancelText: "取消",
                  onOk: () => onBatchToggleStatus(selectedRows, "inactive"),
                });
              }}
              icon={<LockOutlined />}
            >
              批量禁用
            </Button>
          )}
          {canBatchUpdate && (
            <Button
              onClick={() => {
                Modal.confirm({
                  title: `确定要启用选中的 ${selectedRows.length} 个用户吗？`,
                  content: "启用后用户可以正常登录",
                  okText: "确定",
                  cancelText: "取消",
                  onOk: () => onBatchToggleStatus(selectedRows, "active"),
                });
              }}
              icon={<UnlockOutlined />}
            >
              批量启用
            </Button>
          )}
          {canBatchDelete && (
            <Button
              danger
              onClick={() => {
                Modal.confirm({
                  title: `确定要删除选中的 ${selectedRows.length} 个用户吗？`,
                  content: "删除后将无法恢复",
                  okText: "确定",
                  cancelText: "取消",
                  onOk: () => onBatchDelete(selectedRows),
                });
              }}
              icon={<DeleteOutlined />}
            >
              批量删除
            </Button>
          )}
        </Space>
      ),
      // 新建用户按钮
      permissions.canCreate && (
        <Button type="primary" key="primary" onClick={onCreateUser}>
          <PlusOutlined /> 新建用户
        </Button>
      ),
    ].filter(Boolean);
  };

  return (
    <ProTable<API.UserBrief, API.GetUserListRequest>
      headerTitle="用户管理"
      actionRef={actionRef}
      rowKey="id"
      search={{
        labelWidth: 120,
      }}
      scroll={{ x: TABLE_CONFIG.SCROLL_X }}
      options={{
        reload: true,
        setting: true,
        density: true,
      }}
      pagination={{
        defaultPageSize: TABLE_CONFIG.DEFAULT_PAGE_SIZE,
        showQuickJumper: true,
        showSizeChanger: true,
      }}
      toolBarRender={toolBarRender}
      request={async (params, _sort, _filter) => {
        try {
          const response = await getUserList({
            current: params.current,
            pageSize: params.pageSize,
            name: params.name,
            email: params.email,
            userName: params.userName,
            status: params.status,
            tenantId: params.tenantId,
            tagIds: params.tagIds,
          });

          const data = response.data?.list || [];
          const total = response.total || 0;
          const success = response.code === 0;

          // 设置当前页用户ID，触发同步状态查询
          if (success && data.length > 0) {
            const userIds = data.map((user: API.UserBrief) => user.id);
            setCurrentUserIds(userIds);
          }

          return {
            data: data,
            success: success,
            total: total,
          };
        } catch (_error) {
          message.error("获取用户列表失败");
          return {
            data: [],
            success: false,
            total: 0,
          };
        }
      }}
      columns={columns}
      rowSelection={{
        selectedRowKeys,
        onChange: (selectedRowKeys, selectedRows) => {
          onSelectionChange({ selectedRowKeys, selectedRows });
        },
      }}
    />
  );
};

export default UserTable;
