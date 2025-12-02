import { getOrganizationList } from "@/services/organization";
import { getOrgSyncStatus } from "@/services/sync";
import {
  DeleteOutlined,
  DisconnectOutlined,
  EditOutlined,
  ExclamationCircleOutlined,
  LinkOutlined,
  MoreOutlined,
  PlusOutlined,
  SwapOutlined,
  SyncOutlined,
  TeamOutlined,
} from "@ant-design/icons";
import {
  type ActionType,
  type ProColumns,
  ProTable,
} from "@ant-design/pro-components";
import { useAccess } from "@umijs/max";
import type { MenuProps } from "antd";
import { Button, Dropdown, Input, message, Modal, Space, Tag } from "antd";
import React, { useState, useEffect } from "react";
import {
  ORGANIZATION_STATUS_COLORS,
  ORGANIZATION_STATUS_OPTIONS,
  ORGANIZATION_STATUS_TEXT,
  ORGANIZATION_TYPE_OPTIONS,
  TABLE_CONFIG,
} from "../constants";
import type { OrganizationSelectionState } from "../types";
import styles from "./OrganizationTable.less";

const { confirm } = Modal;

interface OrganizationTableProps {
  actionRef: React.MutableRefObject<ActionType | undefined>;
  selectionState: OrganizationSelectionState;
  onSelectionChange: (state: Partial<OrganizationSelectionState>) => void;
  onCreateOrganization: (parentId?: string) => void;
  onEditOrganization: (record: API.OrganizationBrief) => void;
  onDeleteOrganization: (record: API.OrganizationBrief) => void;
  onViewMembers: (record: API.OrganizationBrief) => void;
  onMoveOrganization: (record: API.OrganizationBrief) => void;
  onBatchDelete: (organizations: API.OrganizationBrief[]) => void;
  onBatchToggleStatus: (
    organizations: API.OrganizationBrief[],
    targetStatus: "active" | "inactive"
  ) => void;
  onSyncOrganization: (record: API.OrganizationBrief) => Promise<void>;
  onBindOrganization: (record: API.OrganizationBrief) => void;
  onUnbindOrganization: (record: API.OrganizationBrief) => Promise<void>;
  selectedOrgId?: string; // 选中的组织ID，用于过滤子组织
}

/**
 * 组织列表表格组件
 */
const OrganizationTable: React.FC<OrganizationTableProps> = ({
  actionRef,
  selectionState,
  onSelectionChange,
  onCreateOrganization,
  onEditOrganization,
  onDeleteOrganization,
  onViewMembers,
  onMoveOrganization,
  onBatchDelete,
  onBatchToggleStatus,
  onSyncOrganization,
  onBindOrganization,
  onUnbindOrganization,
  selectedOrgId,
}) => {
  const access = useAccess();
  const { selectedRowKeys, selectedRows } = selectionState;

  // 同步状态管理
  const [syncStatusMap, setSyncStatusMap] = useState<Record<string, boolean>>(
    {},
  );
  const [currentOrgIds, setCurrentOrgIds] = useState<string[]>([]);

  // 查询同步状态
  useEffect(() => {
    if (currentOrgIds.length === 0) return;

    const fetchSyncStatus = async () => {
      try {
        const response = await getOrgSyncStatus({ orgIds: currentOrgIds });
        if (response.code === 0 && response.data?.syncStatus) {
          setSyncStatusMap(response.data.syncStatus);
        }
      } catch (error) {
        // 静默失败，不影响主要功能
        console.error("查询同步状态失败:", error);
      }
    };

    fetchSyncStatus();
  }, [currentOrgIds]);

  // 单个删除确认
  const handleDeleteClick = (record: API.OrganizationBrief) => {
    confirm({
      title: "确认删除",
      icon: <ExclamationCircleOutlined />,
      content: (
        <div>
          <p>
            确定要删除组织 <strong>{record.name}</strong> 吗？
          </p>
          <p className={styles.deleteHint}>此操作不可恢复，请谨慎操作。</p>
        </div>
      ),
      okText: "确认删除",
      okType: "danger",
      cancelText: "取消",
      onOk: () => onDeleteOrganization(record),
    });
  };

  // 批量删除确认
  const handleBatchDeleteClick = () => {
    if (selectedRows.length === 0) {
      message.warning("请先选择要删除的组织");
      return;
    }

    confirm({
      title: "确认批量删除",
      icon: <ExclamationCircleOutlined />,
      content: (
        <div>
          <p>
            确定要删除选中的 <strong>{selectedRows.length}</strong> 个组织吗？
          </p>
          <p className={styles.deleteWarning}>
            注意：删除组织可能会同时删除其子组织。
          </p>
          <p className={styles.deleteHint}>此操作不可恢复，请谨慎操作。</p>
        </div>
      ),
      okText: "确认删除",
      okType: "danger",
      cancelText: "取消",
      onOk: () => onBatchDelete(selectedRows),
    });
  };

  // 批量状态切换确认
  const handleBatchToggleStatusClick = (
    targetStatus: "active" | "inactive"
  ) => {
    if (selectedRows.length === 0) {
      message.warning("请先选择要操作的组织");
      return;
    }

    const actionText = targetStatus === "active" ? "启用" : "停用";
    confirm({
      title: `确认批量${actionText}`,
      icon: <ExclamationCircleOutlined />,
      content: `确定要${actionText}选中的 ${selectedRows.length} 个组织吗？`,
      okText: `确认${actionText}`,
      cancelText: "取消",
      onOk: () => onBatchToggleStatus(selectedRows, targetStatus),
    });
  };

  // 表格列定义
  const columns: ProColumns<API.OrganizationBrief>[] = [
    {
      title: "搜索",
      key: "keyword",
      hideInTable: true,
      dataIndex: "keyword",
      renderFormItem: () => {
        return <Input.Search placeholder="搜索组织名称或代码..." allowClear />;
      },
    },
    {
      title: "组织代码",
      dataIndex: "code",
      width: 140,
      ellipsis: true,
      copyable: true,
      fixed: "left",
      search: false,
    },
    {
      title: "组织名称",
      dataIndex: "name",
      ellipsis: true,
      fixed: "left",
      search: false,
      minWidth: 150,
    },
    {
      title: "类型",
      dataIndex: "type",
      width: 80,
      valueType: "select",
      valueEnum: ORGANIZATION_TYPE_OPTIONS.reduce((acc, item) => {
        acc[item.value] = { text: item.label };
        return acc;
      }, {} as Record<string, { text: string }>),
    },
    {
      title: "层级",
      dataIndex: "level",
      width: 60,
      render: (level) => `L${level}`,
      sorter: true,
      align: "center",
    },
    {
      title: "负责人",
      dataIndex: "managerName",
      width: 100,
      ellipsis: true,
      render: (managerName) =>
        managerName || <span className={styles.unsetManager}>未设置</span>,
    },
    {
      title: "成员",
      dataIndex: "memberCount",
      width: 60,
      render: (memberCount) => `${memberCount}`,
      sorter: true,
      align: "center",
    },
    {
      title: "状态",
      dataIndex: "status",
      width: 70,
      valueType: "select",
      valueEnum: ORGANIZATION_STATUS_OPTIONS.reduce((acc, item) => {
        acc[item.value] = {
          text: item.label,
          status:
            item.value === "active"
              ? "Success"
              : item.value === "inactive"
              ? "Warning"
              : "Error",
        };
        return acc;
      }, {} as Record<string, { text: string; status: string }>),
      render: (_, record) => (
        <Tag
          color={
            ORGANIZATION_STATUS_COLORS[
              record.status as keyof typeof ORGANIZATION_STATUS_COLORS
            ]
          }
          className={styles.statusTag}
        >
          {
            ORGANIZATION_STATUS_TEXT[
              record.status as keyof typeof ORGANIZATION_STATUS_TEXT
            ]
          }
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
      width: 140,
      fixed: "right",
      render: (_, record) => {
        const actions = [];

        // 成员管理 - 高频操作，保留在外面
        actions.push(
          <Button
            key="members"
            type="link"
            size="small"
            icon={<TeamOutlined />}
            onClick={() => onViewMembers(record)}
          >
            成员
          </Button>
        );

        // 低频操作放入更多菜单
        const moreItems: MenuProps["items"] = [];

        // 编辑组织
        if (access.canUpdateOrganizations()) {
          moreItems.push({
            key: "edit",
            icon: <EditOutlined />,
            label: "编辑组织",
            onClick: () => onEditOrganization(record),
          });
        }

        // 添加子组织
        if (access.canCreateOrganizations()) {
          moreItems.push({
            key: "addChild",
            icon: <PlusOutlined />,
            label: "添加子组织",
            onClick: () => onCreateOrganization(record.id),
          });
        }

        // 移动组织
        if (access.canMoveOrganizations()) {
          moreItems.push({
            key: "move",
            icon: <SwapOutlined />,
            label: "移动组织",
            onClick: () => onMoveOrganization(record),
          });
        }

        // Skylark 同步和绑定选项
        const syncStatus = syncStatusMap[record.id];

        // 未同步时显示同步和绑定选项
        if (!syncStatus) {
          moreItems.push({
            key: "sync",
            icon: <SyncOutlined />,
            label: "同步到Skylark",
            onClick: () => onSyncOrganization(record),
          });
          moreItems.push({
            key: "bind",
            icon: <LinkOutlined />,
            label: "绑定到Skylark",
            onClick: () => onBindOrganization(record),
          });
        }

        // 已同步时只显示解绑选项
        if (syncStatus) {
          moreItems.push({
            key: "unbind",
            icon: <DisconnectOutlined />,
            label: "解除Skylark绑定",
            onClick: () => onUnbindOrganization(record),
          });
        }

        // 分隔线
        if (access.canDeleteOrganizations() && moreItems.length > 0) {
          moreItems.push({
            type: "divider",
          });
        }

        // 删除组织 - 危险操作
        if (access.canDeleteOrganizations()) {
          moreItems.push({
            key: "delete",
            icon: <DeleteOutlined />,
            label: "删除组织",
            danger: true,
            onClick: () => handleDeleteClick(record),
          });
        }

        // 如果有更多操作，添加下拉菜单
        if (moreItems.length > 0) {
          actions.push(
            <Dropdown
              key="more"
              menu={{ items: moreItems }}
              trigger={["click"]}
            >
              <Button type="link" size="small" icon={<MoreOutlined />}>
                更多
              </Button>
            </Dropdown>
          );
        }

        return <Space size={0}>{actions}</Space>;
      },
    },
  ];

  // 工具栏渲染
  const toolBarRender = () => {
    const hasSelected = selectedRowKeys.length > 0;
    const canBatchDelete = hasSelected && access.canDeleteOrganizations();
    const canBatchUpdate = hasSelected && access.canUpdateOrganizations();

    // 基础工具栏按钮
    const baseActions = [];

    // 新建组织按钮
    if (access.canCreateOrganizations()) {
      baseActions.push(
        <Button
          key="create"
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => onCreateOrganization()}
        >
          新建组织
        </Button>
      );
    }

    // 批量操作按钮（仅在有选中项时显示）
    const batchActions = hasSelected
      ? [
          <Space key="batch-actions">
            {canBatchUpdate && (
              <>
                <Button
                  onClick={() => handleBatchToggleStatusClick("active")}
                  disabled={
                    !selectedRows.some((org) => org.status !== "active")
                  }
                >
                  批量启用
                </Button>
                <Button
                  onClick={() => handleBatchToggleStatusClick("inactive")}
                  disabled={
                    !selectedRows.some((org) => org.status !== "inactive")
                  }
                >
                  批量停用
                </Button>
              </>
            )}
            {canBatchDelete && (
              <Button danger onClick={handleBatchDeleteClick}>
                批量删除
              </Button>
            )}
          </Space>,
        ]
      : [];

    return [...baseActions, ...batchActions];
  };

  return (
    <>
      {selectedOrgId && (
        <div className={styles.selectedOrgHint}>
          <span>当前显示的是选中组织的子组织列表</span>
          <Button
            type="link"
            size="small"
            className={styles.clearFilterButton}
            onClick={() => {
              // 触发表格刷新，清空过滤
              actionRef.current?.reset?.();
            }}
          >
            显示所有组织
          </Button>
        </div>
      )}
      <ProTable<API.OrganizationBrief>
        actionRef={actionRef}
        columns={columns}
        request={async (params, sort) => {
          try {
            // 处理关键字搜索，如果有keyword则用于name和code的搜索
            const searchParams = params.keyword
              ? {
                  name: params.keyword,
                  code: params.keyword,
                }
              : {
                  code: params.code,
                  name: params.name,
                };

            const response = await getOrganizationList({
              current: params.current || 1,
              pageSize: params.pageSize || TABLE_CONFIG.PAGE_SIZE,
              ...searchParams,
              type: params.type,
              status: params.status,
              parentId: selectedOrgId || params.parentId, // 优先使用选中的组织ID作为父组织过滤
              level: params.level,
              ...sort,
            });

            if (response.code === 0) {
              const data = response.data?.list || [];

              // 设置当前页组织ID，触发同步状态查询
              if (data.length > 0) {
                const orgIds = data.map((org: API.OrganizationBrief) => org.id);
                setCurrentOrgIds(orgIds);
              }

              return {
                data: data,
                success: true,
                total: response.total || 0,
              };
            } else {
              message.error(response.msg || "获取组织列表失败");
              return {
                data: [],
                success: false,
                total: 0,
              };
            }
          } catch (error) {
            console.error("获取组织列表失败:", error);
            message.error("获取组织列表失败");
            return {
              data: [],
              success: false,
              total: 0,
            };
          }
        }}
        rowKey="id"
        search={{
          labelWidth: "auto",
          defaultCollapsed: true,
        }}
        pagination={{
          defaultPageSize: TABLE_CONFIG.PAGE_SIZE,
          showSizeChanger: true,
          pageSizeOptions: [...TABLE_CONFIG.PAGE_SIZE_OPTIONS],
        }}
        scroll={{ x: "max-content" }}
        toolBarRender={toolBarRender}
        rowSelection={
          access.canDeleteOrganizations() || access.canUpdateOrganizations()
            ? {
                selectedRowKeys,
                onChange: (keys, rows) => {
                  onSelectionChange({
                    selectedRowKeys: keys,
                    selectedRows: rows,
                  });
                },
              }
            : false
        }
        tableAlertRender={({ selectedRowKeys }) => (
          <Space size={24}>
            <span>
              已选择 <strong>{selectedRowKeys.length}</strong> 项
            </span>
          </Space>
        )}
      />
    </>
  );
};

export default OrganizationTable;
