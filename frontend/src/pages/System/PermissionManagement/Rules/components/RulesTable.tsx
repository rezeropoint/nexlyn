import { getPermissionRules } from "@/services/permission";
import { CheckCircleOutlined, DeleteOutlined } from "@ant-design/icons";
import {
  type ActionType,
  type ProColumns,
  ProTable,
} from "@ant-design/pro-components";
import { useAccess } from "@umijs/max";
import { Button, message, Popconfirm, Space, Tag } from "antd";
import React from "react";
import AddRuleForm from "./AddRuleForm";
import styles from "./RulesTable.less";

interface RulesTableProps {
  actionRef?: React.MutableRefObject<ActionType | undefined>;
  onCheckPermission: (record: API.PermissionRule) => void;
  onDeleteRule: (record: API.PermissionRule) => void;
  onAddRule: (values: API.AddPermissionRequest) => Promise<boolean>;
}

const RulesTable: React.FC<RulesTableProps> = ({
  actionRef,
  onCheckPermission,
  onDeleteRule,
  onAddRule,
}) => {
  const access = useAccess();

  // 表格列定义
  const columns: ProColumns<API.PermissionRule>[] = [
    {
      title: "用户Key",
      dataIndex: "userKey",
      ellipsis: true,
      copyable: true,
    },
    {
      title: "租户Key",
      dataIndex: "tenantKey",
      ellipsis: true,
      copyable: true,
    },
    {
      title: "资源",
      dataIndex: "resource",
      ellipsis: true,
      render: (_, record) => <Tag color="blue">{record.resource}</Tag>,
    },
    {
      title: "操作",
      dataIndex: "action",
      ellipsis: true,
      render: (_, record) => <Tag color="green">{record.action}</Tag>,
    },
    {
      title: "操作",
      valueType: "option",
      key: "option",
      fixed: "right",
      width: 180,
      render: (_, record) => {
        const actions = [];

        // 检查权限按钮
        actions.push(
          <Button
            key="check"
            type="link"
            size="small"
            icon={<CheckCircleOutlined />}
            onClick={() => onCheckPermission(record)}
          >
            检查
          </Button>
        );

        // 权限检查：是否可以删除这条规则
        const isSuperAdmin = access.isSuperAdmin?.();
        const canRemoveBasicPermission = access.canRemovePermissions?.();

        // 检查是否为超级管理员相关的权限规则
        const isProtectedRule =
          record.userKey === "super_admin" || // 超级管理员角色的权限
          record.userKey?.includes("super-admin") || // 超级管理员用户的权限
          (record.resource === "permission" && record.action === "admin") || // 权限管理权限
          record.roleKey === "super_admin"; // 超级管理员角色分配

        const canDeleteThisRule =
          canRemoveBasicPermission && (isSuperAdmin || !isProtectedRule);

        // 删除按钮
        if (canDeleteThisRule) {
          actions.push(
            <Popconfirm
              key="delete"
              title="确定要删除这条权限规则吗？"
              description="此操作不可逆，请谨慎操作"
              onConfirm={() => onDeleteRule(record)}
              okText="确定"
              cancelText="取消"
            >
              <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                删除
              </Button>
            </Popconfirm>
          );
        } else {
          // 如果有基础删除权限但不能删除这条特定规则，显示禁用的删除按钮
          const tooltipMessage = canRemoveBasicPermission
            ? "无法删除超级管理员相关权限规则"
            : "权限不足";

          actions.push(
            <Button
              key="delete-disabled"
              type="link"
              size="small"
              disabled
              icon={<DeleteOutlined />}
              title={tooltipMessage}
              className={styles.disabledDeleteButton}
            >
              删除
            </Button>
          );
        }

        return <Space size="small">{actions}</Space>;
      },
    },
  ];

  return (
    <ProTable<API.PermissionRule>
      headerTitle="用户直接权限规则"
      actionRef={actionRef}
      rowKey="userKey"
      search={{
        labelWidth: 120,
      }}
      toolBarRender={() =>
        [
          access.canAddPermissions?.() && (
            <AddRuleForm key="add-rule" onFinish={onAddRule} />
          ),
        ].filter(Boolean)
      }
      request={async (params) => {
        try {
          const res = await getPermissionRules({
            current: params.current,
            pageSize: params.pageSize,
            userKey: params.userKey,
            tenantKey: params.tenantKey,
            resource: params.resource,
            action: params.action,
            ptype: "p", // 只获取权限规则
          });

          return {
            data: res.data || [],
            success: res.code === 0,
            total: res.total || 0,
          };
        } catch (_error) {
          message.error("获取权限规则列表失败");
          return { data: [], success: false, total: 0 };
        }
      }}
      columns={columns}
      scroll={{ x: 800 }}
    />
  );
};

export default RulesTable;
