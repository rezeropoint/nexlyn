import {
  createRole,
  deleteRole,
  getAvailablePermissions,
  getRoleList,
  type PermissionItem,
  type RoleInfo,
  updateRole,
} from "@/services/permission/role";
import {
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  PlusOutlined,
  UserOutlined,
} from "@ant-design/icons";
import {
  type ActionType,
  DrawerForm,
  ModalForm,
  PageContainer,
  type ProColumns,
  ProFormItem,
  ProFormText,
  ProFormTextArea,
  ProTable,
} from "@ant-design/pro-components";
import { useAccess } from "@umijs/max";
import { Alert, Button, message, Popconfirm, Space, Tag } from "antd";
import React, { useCallback, useRef, useState } from "react";
import PermissionSelector from "./components/PermissionSelector";
import styles from "./index.less";

const RolePermissions: React.FC = () => {
  const access = useAccess();
  const actionRef = useRef<ActionType>();
  const [availablePermissions, setAvailablePermissions] = useState<
    PermissionItem[]
  >([]);
  const [selectedRole, setSelectedRole] = useState<RoleInfo | undefined>();

  // 加载可用权限列表
  const loadAvailablePermissions = useCallback(async () => {
    try {
      const res = await getAvailablePermissions();
      if (res.code === 0 && res.data) {
        setAvailablePermissions(res.data);
      }
    } catch (error) {
      console.error("获取可用权限列表失败:", error);
    }
  }, []);

  // 组件挂载时加载权限列表
  React.useEffect(() => {
    loadAvailablePermissions();
  }, [loadAvailablePermissions]);

  // 删除角色
  const handleDeleteRole = async (record: RoleInfo) => {
    try {
      const res = await deleteRole(record.roleKey);
      if (res.code === 0) {
        message.success(res.msg || "删除角色成功");
        actionRef.current?.reload();
      } else {
        message.error(res.msg || "删除角色失败");
      }
    } catch (_error) {
      message.error("删除角色失败");
    }
  };

  // 创建角色
  const handleCreateRole = async (values: any) => {
    try {
      // 将权限分组转换为权限字符串数组
      const permissions = values.permissions || [];

      const res = await createRole({
        roleKey: values.roleKey,
        roleName: values.roleName,
        description: values.description || "",
        permissions: permissions,
      });

      if (res.code === 0) {
        message.success(res.msg || "创建角色成功");
        actionRef.current?.reload();
        return true;
      } else {
        message.error(res.msg || "创建角色失败");
        return false;
      }
    } catch (_error) {
      message.error("创建角色失败");
      return false;
    }
  };

  // 更新角色
  const handleUpdateRole = async (values: any) => {
    if (!selectedRole) return false;

    try {
      const res = await updateRole(selectedRole.roleKey, {
        roleName: values.roleName,
        description: values.description || "",
        permissions: values.permissions || [],
      });

      if (res.code === 0) {
        message.success(res.msg || "更新角色成功");
        actionRef.current?.reload();
        return true;
      } else {
        message.error(res.msg || "更新角色失败");
        return false;
      }
    } catch (_error) {
      message.error("更新角色失败");
      return false;
    }
  };

  // 格式化权限显示
  const _formatPermissions = (permissions: string[]) => {
    return permissions
      .map((perm) => {
        const permItem = availablePermissions.find(
          (item) => `${item.resource}:${item.action}` === perm
        );
        return permItem ? permItem.description : perm;
      })
      .join("、");
  };

  // 表格列定义
  const columns: ProColumns<RoleInfo>[] = [
    {
      title: "角色键",
      dataIndex: "roleKey",
      width: 120,
      ellipsis: true,
      copyable: true,
    },
    {
      title: "角色名称",
      dataIndex: "roleName",
      width: 120,
      ellipsis: true,
    },
    {
      title: "描述",
      dataIndex: "description",
      width: 200,
      ellipsis: true,
    },
    {
      title: "权限数量",
      dataIndex: "permissions",
      width: 100,
      render: (_, record) => (
        <Tag color="blue">{record.permissions.length}个</Tag>
      ),
    },
    {
      title: "操作",
      valueType: "option",
      key: "option",
      fixed: "right",
      width: 180,
      render: (_, record) => {
        const actions = [];

        // 查看/编辑角色按钮
        if (access.canReadPermissions?.()) {
          const isReadOnly = !access.canAddRoles?.();
          actions.push(
            <ModalForm
              key={isReadOnly ? "view" : "edit"}
              title={isReadOnly ? `查看角色 - ${record.roleName}` : "编辑角色"}
              trigger={
                <Button
                  type="link"
                  size="small"
                  icon={isReadOnly ? <EyeOutlined /> : <EditOutlined />}
                >
                  {isReadOnly ? "查看" : "编辑"}
                </Button>
              }
              onFinish={isReadOnly ? undefined : handleUpdateRole}
              width={600}
              modalProps={{
                onCancel: () => setSelectedRole(undefined),
              }}
              onOpenChange={(open) => {
                if (open) {
                  setSelectedRole(record);
                } else {
                  setSelectedRole(undefined);
                }
              }}
              initialValues={{
                roleKey: record.roleKey,
                roleName: record.roleName,
                description: record.description,
                permissions: record.permissions,
              }}
              submitter={isReadOnly ? false : undefined}
            >
              <ProFormText name="roleKey" label="角色键" disabled />
              <ProFormText
                name="roleName"
                label="角色名称"
                disabled={isReadOnly}
                rules={
                  isReadOnly
                    ? []
                    : [{ required: true, message: "请输入角色名称" }]
                }
              />
              <ProFormTextArea
                name="description"
                label="描述"
                disabled={isReadOnly}
              />
              <ProFormItem name="permissions" label="权限">
                <PermissionSelector
                  availablePermissions={availablePermissions}
                  disabled={isReadOnly}
                />
              </ProFormItem>
            </ModalForm>
          );
        }

        // 删除角色按钮
        if (access.canRemoveRoles?.()) {
          actions.push(
            <Popconfirm
              key="delete"
              title="确定要删除这个角色吗？"
              description="删除角色后，所有拥有该角色的用户将失去相关权限"
              onConfirm={() => handleDeleteRole(record)}
              okText="确定"
              cancelText="取消"
            >
              <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                删除
              </Button>
            </Popconfirm>
          );
        }

        return <Space size="small">{actions}</Space>;
      },
    },
  ];

  // 权限检查
  if (!access.canReadPermissions?.()) {
    return (
      <PageContainer>
        <div className={styles.noPermissionContainer}>
          <UserOutlined className={styles.noPermissionIcon} />
          <h2>权限不足</h2>
          <p>您需要权限管理权限才能访问此页面</p>
        </div>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <Alert
        message="🎭 角色权限管理"
        description={
          <div>
            • <strong>统一管理所有角色</strong>
            ：包括系统预定义角色（super_admin、admin、user）和自定义角色
            <br />• <strong>权限模板</strong>
            ：角色定义了一组权限的集合，用户分配角色后自动获得这些权限
            <br />• <strong>角色管理</strong>
            ：创建自定义角色并为其配置相应的权限
            <br />• 角色权限变更会影响所有拥有该角色的用户
          </div>
        }
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />
      <ProTable<RoleInfo>
        headerTitle="角色权限管理"
        actionRef={actionRef}
        rowKey="roleKey"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() =>
          [
            access.canAddRoles?.() && (
              <DrawerForm
                key="add-role"
                title="创建角色"
                trigger={
                  <Button type="primary">
                    <PlusOutlined /> 创建角色
                  </Button>
                }
                onFinish={handleCreateRole}
                width={600}
              >
                <ProFormText
                  name="roleKey"
                  label="角色键"
                  placeholder="请输入角色键（英文）"
                  rules={[
                    { required: true, message: "请输入角色键" },
                    {
                      pattern: /^[a-zA-Z_][a-zA-Z0-9_]*$/,
                      message:
                        "角色键只能包含字母、数字和下划线，且以字母或下划线开头",
                    },
                  ]}
                />
                <ProFormText
                  name="roleName"
                  label="角色名称"
                  placeholder="请输入角色显示名称"
                  rules={[{ required: true, message: "请输入角色名称" }]}
                />
                <ProFormTextArea
                  name="description"
                  label="描述"
                  placeholder="请输入角色描述"
                />
                <ProFormItem name="permissions" label="权限">
                  <PermissionSelector
                    availablePermissions={availablePermissions}
                  />
                </ProFormItem>
              </DrawerForm>
            ),
          ].filter(Boolean)
        }
        request={async (params) => {
          try {
            const res = await getRoleList({
              current: params.current,
              pageSize: params.pageSize,
              roleKey: params.roleKey,
              roleName: params.roleName,
            });

            return {
              data: res.data || [],
              success: res.code === 0,
              total: res.total || 0,
            };
          } catch (_error) {
            message.error("获取角色列表失败");
            return { data: [], success: false, total: 0 };
          }
        }}
        columns={columns}
        scroll={{ x: 1000 }}
      />
    </PageContainer>
  );
};

export default RolePermissions;
