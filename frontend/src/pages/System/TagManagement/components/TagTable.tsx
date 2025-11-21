import { getTagList } from "@/services/tags";
import { formatDateTime } from "@/utils/date";
import {
  DeleteOutlined,
  EditOutlined,
  InfoCircleOutlined,
  PlusOutlined,
} from "@ant-design/icons";
import {
  type ActionType,
  type ProColumns,
  ProTable,
} from "@ant-design/pro-components";
import { Alert, Button, message, Modal, Popconfirm, Tag, Tooltip } from "antd";
import React from "react";
import { TABLE_CONFIG, TAG_SCOPES } from "../constants";
import { useTagPermissions } from "../hooks/useTagPermissions";
import type { TagSelectionState } from "../types";
import styles from "./TagTable.less";

interface TagTableProps {
  actionRef: React.MutableRefObject<ActionType | undefined>;
  selectionState: TagSelectionState;
  onSelectionChange: (state: Partial<TagSelectionState>) => void;
  onCreateTag: () => void;
  onEditTag: (record: API.TagDefinition) => void;
  onDeleteTag: (record: API.TagDefinition) => void;
  onBatchDelete: (tags: API.TagDefinition[]) => void;
  defaultScope: string;
}

/**
 * 标签表格组件
 */
const TagTable: React.FC<TagTableProps> = ({
  actionRef,
  selectionState,
  onSelectionChange,
  onCreateTag,
  onEditTag,
  onDeleteTag,
  onBatchDelete,
  defaultScope,
}) => {
  const permissions = useTagPermissions();
  const { selectedRowKeys, selectedRows } = selectionState;

  // 表格列定义
  const columns: ProColumns<API.TagDefinition>[] = [
    {
      title: "标签ID",
      dataIndex: "id",
      ellipsis: true,
      copyable: true,
      hideInSearch: true,
    },
    {
      title: "标签名称",
      dataIndex: "label",
      ellipsis: true,
      render: (_, record) => (
        <Tag color="blue" className={styles.tagLabel}>
          {record.label}
        </Tag>
      ),
    },
    {
      title: "作用域",
      dataIndex: "scope",
      valueType: "select",
      valueEnum: {
        user: { text: TAG_SCOPES.user.text, status: TAG_SCOPES.user.status },
        tenant: {
          text: TAG_SCOPES.tenant.text,
          status: TAG_SCOPES.tenant.status,
        },
      },
      fieldProps: {
        placeholder: "请选择作用域",
      },
      render: (_, record) => {
        const scopeConfig =
          TAG_SCOPES[record.scope as keyof typeof TAG_SCOPES] ||
          TAG_SCOPES.user;
        const isTenantTag = record.scope === "tenant";

        return (
          <div className={styles.scopeContainer}>
            <Tag color={scopeConfig.color}>{scopeConfig.text}</Tag>
            {isTenantTag && !permissions.canManageTenantTags && (
              <Tooltip title="租户标签应用需要超级管理员权限">
                <InfoCircleOutlined className={styles.warningIcon} />
              </Tooltip>
            )}
          </div>
        );
      },
    },
    {
      title: "描述",
      dataIndex: "description",
      ellipsis: true,
      hideInSearch: true,
    },
    {
      title: "创建时间",
      dataIndex: "createdAt",
      width: 160,
      valueType: "dateTime",
      hideInSearch: true,
      sorter: true,
      render: (_, record) => formatDateTime(record.createdAt),
    },
    {
      title: "更新时间",
      dataIndex: "updatedAt",
      width: 160,
      valueType: "dateTime",
      hideInSearch: true,
      sorter: true,
      render: (_, record) => formatDateTime(record.updatedAt),
    },
    {
      title: "操作",
      valueType: "option",
      key: "option",
      fixed: "right",
      width: 150,
      render: (_, record) => {
        const isTenantTag = record.scope === "tenant";
        const isUserTag = record.scope === "user";

        // 基于标签作用域检查权限
        const canManageThisTag = isTenantTag
          ? permissions.canManageTenantTags
          : permissions.canManageUserTags;

        const showPermissionWarning = !canManageThisTag;

        const getPermissionTooltip = () => {
          if (isTenantTag && !permissions.canManageTenantTags) {
            return "租户标签管理需要超级管理员权限";
          }
          if (isUserTag && !permissions.canManageUserTags) {
            return "用户标签管理需要租户管理员权限";
          }
          return "";
        };

        const editButton = permissions.canUpdate && (
          <Tooltip
            key="edit-tooltip"
            title={showPermissionWarning ? getPermissionTooltip() : "编辑标签"}
          >
            <Button
              key="edit"
              type="link"
              size="small"
              icon={<EditOutlined />}
              onClick={() => onEditTag(record)}
              disabled={showPermissionWarning}
              className={showPermissionWarning ? styles.disabledButton : ""}
            >
              编辑
            </Button>
          </Tooltip>
        );

        const deleteButton = permissions.canDelete && (
          <Tooltip
            key="delete-tooltip"
            title={showPermissionWarning ? getPermissionTooltip() : "删除标签"}
          >
            <Popconfirm
              key="delete"
              title="确定要删除这个标签吗？"
              description="删除标签可能会影响相关的数据，请谨慎操作"
              onConfirm={() => onDeleteTag(record)}
              okText="确定"
              cancelText="取消"
              disabled={showPermissionWarning}
            >
              <Button
                type="link"
                size="small"
                danger
                icon={<DeleteOutlined />}
                disabled={showPermissionWarning}
                className={showPermissionWarning ? styles.disabledButton : ""}
              >
                删除
              </Button>
            </Popconfirm>
          </Tooltip>
        );

        return [editButton, deleteButton].filter(Boolean);
      },
    },
  ];

  // 权限检查
  if (!permissions.canRead) {
    return (
      <div className={styles.noPermissionContainer}>
        <h2>权限不足</h2>
        <p>您需要标签管理权限才能访问此页面</p>
      </div>
    );
  }

  // 权限提示组件
  const renderPermissionInfo = () => {
    const hasUserTagPermission = permissions.canManageUserTags;
    const hasTenantTagPermission = permissions.canManageTenantTags;
    const hasAllPermissions = hasUserTagPermission && hasTenantTagPermission;

    return (
      <Alert
        message="权限说明"
        description={
          <div>
            <p>
              • 用户标签管理：创建、编辑、删除用户标签（租户管理员及以上权限）
            </p>
            <p>• 租户标签管理：创建、编辑、删除租户标签（仅超级管理员权限）</p>
            {!hasAllPermissions && (
              <p className={styles.permissionWarning}>
                <InfoCircleOutlined />
                {!hasTenantTagPermission && "租户标签管理需要超级管理员权限"}
                {!hasUserTagPermission && !hasTenantTagPermission && " · "}
                {!hasUserTagPermission && "用户标签管理需要租户管理员权限"}
              </p>
            )}
          </div>
        }
        type={hasAllPermissions ? "success" : "warning"}
        showIcon
        className={styles.permissionAlert}
        closable
      />
    );
  };

  return (
    <>
      {renderPermissionInfo()}
      <ProTable<API.TagDefinition>
        headerTitle="标签管理"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: TABLE_CONFIG.SEARCH_LABEL_WIDTH,
          defaultCollapsed: false,
        }}
        form={{
          initialValues: {
            scope: defaultScope,
          },
        }}
        toolBarRender={() => {
          const hasSelected = selectedRowKeys.length > 0;
          const canBatchDelete = hasSelected && permissions.canDelete;

          return [
            // 多选操作按钮
            hasSelected && canBatchDelete && (
              <Button
                key="batch-delete"
                danger
                onClick={() => {
                  Modal.confirm({
                    title: `确定要删除选中的 ${selectedRows.length} 个标签吗？`,
                    content: "删除标签可能会影响相关的数据，请谨慎操作",
                    okText: "确定",
                    cancelText: "取消",
                    onOk: () => onBatchDelete(selectedRows),
                  });
                }}
                icon={<DeleteOutlined />}
              >
                批量删除
              </Button>
            ),
            // 创建标签按钮
            permissions.canCreate && (
              <Button
                key="create-tag"
                type="primary"
                onClick={onCreateTag}
                icon={<PlusOutlined />}
              >
                创建标签
              </Button>
            ),
          ].filter(Boolean);
        }}
        request={async (params) => {
          try {
            // 如果没有传入作用域，使用默认值
            const scope = params.scope || defaultScope;

            const res = await getTagList({
              current: params.current,
              pageSize: params.pageSize,
              scope: scope,
              keyword: params.label || params.keyword,
            });

            // 如果后端返回错误，显示具体错误信息
            if (res.code !== 0) {
              message.error(res.msg || "获取标签列表失败");
              return { data: [], success: false, total: 0 };
            }

            return {
              data: res.data?.list || [],
              success: res.code === 0,
              total: res.pageParams?.total || res.total || 0,
            };
          } catch (error) {
            console.error("获取标签列表失败:", error);
            // 如果是网络错误或其他异常，统一显示通用错误信息
            message.error("获取标签列表失败");
            return { data: [], success: false, total: 0 };
          }
        }}
        columns={columns}
        scroll={{ x: TABLE_CONFIG.SCROLL_X }}
        rowSelection={{
          selectedRowKeys,
          onChange: (selectedRowKeys, selectedRows) => {
            onSelectionChange({ selectedRowKeys, selectedRows });
          },
          onSelect: (record, selected, _selectedRows) => {
            console.log("选中/取消选中标签:", record.label, selected);
          },
          onSelectAll: (selected, selectedRows, _changeRows) => {
            console.log("全选/取消全选:", selected, selectedRows.length);
          },
        }}
      />
    </>
  );
};

export default TagTable;
