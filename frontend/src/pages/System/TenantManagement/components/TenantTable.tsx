import type { Tenant } from "@/services/tenant";
import { getTenantList } from "@/services/tenant";
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
import dayjs from "dayjs";
import React from "react";
import {
  TABLE_CONFIG,
  TENANT_STATUS_COLORS,
  TENANT_STATUS_TEXT,
} from "../constants";
import { useTenantPermissions } from "../hooks/useTenantPermissions";
import type { TenantSelectionState } from "../types";
import styles from "./TenantTable.less";
import TenantTableActions from "./TenantTableActions";

interface TenantTableProps {
  actionRef: React.MutableRefObject<ActionType | undefined>;
  selectionState: TenantSelectionState;
  onSelectionChange: (state: Partial<TenantSelectionState>) => void;
  onCreateTenant: () => void;
  onEditTenant: (record: Tenant) => void;
  onDeleteTenant: (record: Tenant) => void;
  onToggleStatus: (record: Tenant) => void;
  onBatchDelete: (tenants: Tenant[]) => void;
  onBatchToggleStatus: (
    tenants: Tenant[],
    targetStatus: "active" | "inactive"
  ) => void;
}

/**
 * 租户表格组件
 */
const TenantTable: React.FC<TenantTableProps> = ({
  actionRef,
  selectionState,
  onSelectionChange,
  onCreateTenant,
  onEditTenant,
  onDeleteTenant,
  onToggleStatus,
  onBatchDelete,
  onBatchToggleStatus,
}) => {
  const permissions = useTenantPermissions();
  const { selectedRowKeys, selectedRows } = selectionState;

  // 表格列定义
  const columns: ProColumns<Tenant>[] = [
    {
      title: "租户ID",
      dataIndex: "id",
      width: 150,
      ellipsis: true,
      copyable: true,
      fixed: "left",
      hideInTable: true,
    },
    {
      title: "租户Key",
      dataIndex: "tenantKey",
      width: 150,
      ellipsis: true,
      copyable: true,
    },
    {
      title: "租户名称",
      dataIndex: "tenantName",
      width: 150,
      ellipsis: true,
    },
    {
      title: "描述",
      dataIndex: "description",
      width: 200,
      ellipsis: true,
    },
    {
      title: "联系邮箱",
      dataIndex: "contactEmail",
      width: 180,
      ellipsis: true,
      copyable: true,
    },
    {
      title: "标签",
      dataIndex: "tags",
      width: 150,
      render: (_, record) => {
        const tags = record.tags;
        if (!tags || tags.length === 0) return "-";
        return (
          <div>
            {tags.slice(0, 2).map((tag: string) => (
              <Tag key={tag} className={styles.tagItem}>
                {tag}
              </Tag>
            ))}
            {tags.length > 2 && <Tag color="default">+{tags.length - 2}</Tag>}
          </div>
        );
      },
    },
    {
      title: "最大用户数",
      dataIndex: "maxUsers",
      width: 120,
      align: "center",
    },
    {
      title: "创建时间",
      dataIndex: "createdAt",
      width: 150,
      render: (_, record) =>
        record.createdAt
          ? dayjs(record.createdAt).format("YYYY-MM-DD HH:mm:ss")
          : "-",
    },
    {
      title: "过期时间",
      dataIndex: "expiresAt",
      width: 150,
      render: (_, record) =>
        record.expiresAt
          ? dayjs(record.expiresAt).format("YYYY-MM-DD HH:mm:ss")
          : "永不过期",
    },
    {
      title: "状态",
      dataIndex: "status",
      width: 100,
      render: (_, record) => (
        <Tag
          color={
            TENANT_STATUS_COLORS[
              record.status as keyof typeof TENANT_STATUS_COLORS
            ]
          }
        >
          {TENANT_STATUS_TEXT[
            record.status as keyof typeof TENANT_STATUS_TEXT
          ] || record.status}
        </Tag>
      ),
    },
    {
      title: "操作",
      valueType: "option",
      key: "option",
      width: 220,
      fixed: "right",
      render: (_, record) => (
        <TenantTableActions
          record={record}
          onEdit={onEditTenant}
          onDelete={onDeleteTenant}
          onToggleStatus={onToggleStatus}
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
        (tenant) =>
          permissions.canDelete && permissions.canOperateTenant(tenant)
      );
    const canBatchUpdate =
      hasSelected &&
      selectedRows.every(
        (tenant) =>
          permissions.canUpdate && permissions.canOperateTenant(tenant)
      );

    return [
      // 多选操作按钮
      hasSelected && (
        <Space key="batch-actions">
          {canBatchUpdate && (
            <Button
              onClick={() => {
                Modal.confirm({
                  title: `确定要禁用选中的 ${selectedRows.length} 个租户吗？`,
                  content: "禁用后这些租户将无法正常使用",
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
                  title: `确定要启用选中的 ${selectedRows.length} 个租户吗？`,
                  content: "启用后这些租户可以正常使用",
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
                  title: `确定要删除选中的 ${selectedRows.length} 个租户吗？`,
                  content: "删除后将无法恢复，租户下的所有数据都将被删除",
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
      // 新建租户按钮
      permissions.canCreate && (
        <Button type="primary" key="primary" onClick={onCreateTenant}>
          <PlusOutlined /> 新建租户
        </Button>
      ),
    ].filter(Boolean);
  };

  return (
    <ProTable<Tenant>
      headerTitle="租户管理"
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
      request={async (params) => {
        try {
          const response = await getTenantList({
            current: params.current,
            pageSize: params.pageSize,
            tenantName: params.tenantName,
            status: params.status,
          });

          console.log("租户列表数据示例:", response.data?.list?.[0]);

          const data = response.data?.list || [];
          const total = response.total || 0;
          const success = response.code === 0;

          return {
            data: data,
            success: success,
            total: total,
          };
        } catch (error) {
          console.error("获取租户列表失败:", error);
          message.error("获取租户列表失败");
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
        onSelect: (record, selected, _selectedRows) => {
          console.log("选中/取消选中租户:", record.tenantName, selected);
        },
        onSelectAll: (selected, selectedRows, _changeRows) => {
          console.log("全选/取消全选:", selected, selectedRows.length);
        },
      }}
    />
  );
};

export default TenantTable;
