import { useRoleOptions } from "@/hooks/useRoleOptions";
import {
  addRoleForUser,
  getPermissionRules,
  removeRoleForUser,
} from "@/services/permission";
import { getTenantOptions } from "@/services/tenant";
import { getUserOptions } from "@/services/user";
import { DeleteOutlined, PlusOutlined, UserOutlined } from "@ant-design/icons";
import {
  type ActionType,
  DrawerForm,
  PageContainer,
  type ProColumns,
  ProFormSelect,
  ProTable,
} from "@ant-design/pro-components";
import { useAccess, useModel } from "@umijs/max";
import { Alert, Button, message, Popconfirm, Space, Tag } from "antd";
import React, { useCallback, useEffect, useRef, useState } from "react";
import styles from "./index.less";

const UserRoles: React.FC = () => {
  const access = useAccess();
  const { initialState } = useModel("@@initialState");
  const actionRef = useRef<ActionType>();
  const [userOptions, setUserOptions] = useState<any[]>([]);
  const [tenantOptions, setTenantOptions] = useState<any[]>([]);
  const [currentUserTenant, setCurrentUserTenant] = useState<{
    tenantKey: string;
    tenantName: string;
  } | null>(null);
  const [canSelectTenant, setCanSelectTenant] = useState<boolean>(false);

  // 使用共享的角色选项Hook
  const { roleOptions, loadRoleOptions } = useRoleOptions();

  // 获取当前用户租户信息
  useEffect(() => {
    if (initialState?.currentUser?.tenantInfo) {
      const tenant = initialState.currentUser.tenantInfo;
      setCurrentUserTenant({
        tenantKey: tenant.tenantKey,
        tenantName: tenant.tenantName,
      });
    }
  }, [initialState]);

  // 检查是否有跨租户权限
  useEffect(() => {
    // 只有超级管理员可以选择租户；其他用户只能操作自己的租户
    const hasMultiTenantAccess = access.isSuperAdmin?.() || false;
    setCanSelectTenant(hasMultiTenantAccess);
  }, [access]);

  // 加载用户选项
  const loadUserOptions = useCallback(async (keyword?: string) => {
    try {
      const res = await getUserOptions({ keyword, limit: 50 });
      if (res.code === 0 && res.data?.list) {
        setUserOptions(res.data.list);
      }
    } catch (error) {
      console.error("获取用户选项失败:", error);
    }
  }, []);

  // 加载租户选项
  const loadTenantOptions = useCallback(async (keyword?: string) => {
    try {
      const res = await getTenantOptions({
        keyword,
        limit: 50,
        status: "active",
      });
      if (res.code === 0 && res.data?.list) {
        setTenantOptions(res.data.list);
      }
    } catch (error) {
      console.error("获取租户选项失败:", error);
    }
  }, []);

  // 删除用户角色
  const handleRemoveRole = async (record: API.PermissionRule) => {
    if (record.ptype === "g" && record.roleKey) {
      try {
        const res = await removeRoleForUser({
          userKey: record.userKey,
          roleKey: record.roleKey as API.RoleKey,
          tenantKey: record.tenantKey,
        });

        if (res.code === 0) {
          message.success(res.msg || "移除角色成功");
          actionRef.current?.reload();
        }
        // 错误信息已由响应拦截器处理，不需要重复显示
      } catch (error) {
        // 网络错误等异常情况才显示通用错误
        console.error("移除角色失败:", error);
      }
    }
  };

  // 添加用户角色
  const handleAddRole = async (values: API.AddRoleRequest) => {
    try {
      // 如果没有选择租户（普通管理员），使用当前用户的租户
      const tenantKey = values.tenantKey || currentUserTenant?.tenantKey;

      if (!tenantKey) {
        message.error("无法获取租户信息，请刷新页面后重试");
        return false;
      }

      const submitValues = {
        ...values,
        tenantKey,
      };

      const res = await addRoleForUser(submitValues);
      if (res.code === 0) {
        message.success(res.msg || "添加角色成功");
        actionRef.current?.reload();
        return true;
      }
      // 错误信息已由响应拦截器处理，不需要重复显示
      return false;
    } catch (error) {
      // 网络错误等异常情况才显示通用错误
      console.error("添加角色失败:", error);
      return false;
    }
  };

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
      title: "角色",
      dataIndex: "roleKey",
      ellipsis: true,
      render: (_, record) => {
        if (!record.roleKey) return "-";

        const roleConfig = {
          super_admin: { color: "red", text: "超级管理员" },
          admin: { color: "orange", text: "租户管理员" },
          user: { color: "blue", text: "普通用户" },
        };

        const config = roleConfig[
          record.roleKey as keyof typeof roleConfig
        ] || {
          color: "default",
          text: record.roleKey,
        };

        return <Tag color={config.color}>{config.text}</Tag>;
      },
    },
    {
      title: "操作",
      valueType: "option",
      key: "option",
      fixed: "right",
      width: 120,
      render: (_, record) => {
        const actions = [];

        // 删除角色按钮
        if (access.canRemoveRoles?.()) {
          actions.push(
            <Popconfirm
              key="delete"
              title="确定要移除这个角色吗？"
              description="移除角色后用户可能失去相关权限"
              onConfirm={() => handleRemoveRole(record)}
              okText="确定"
              cancelText="取消"
            >
              <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                移除
              </Button>
            </Popconfirm>
          );
        } else {
          // 占位按钮
          actions.push(
            <Button
              key="delete-placeholder"
              type="link"
              size="small"
              disabled
              className={styles.disabledDefaultTag}
            >
              --
            </Button>
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
        message="👥 用户角色分配"
        description={
          <>
            <div>
              • <strong>角色分配</strong>：为用户分配系统角色或自定义角色
            </div>
            <div>
              • <strong>权限继承</strong>：用户获得所分配角色的所有权限
            </div>
            <div>
              • <strong>多角色支持</strong>
              ：一个用户可以拥有多个角色，权限自动合并
            </div>
            <div>• 角色本身的权限配置请使用"角色权限管理"页面</div>
          </>
        }
        type="warning"
        showIcon
        style={{ marginBottom: 16 }}
      />
      <ProTable<API.PermissionRule>
        headerTitle="用户角色管理"
        actionRef={actionRef}
        rowKey={(record) =>
          `${record.userKey || ""}-${record.roleKey || ""}-${
            record.tenantKey || ""
          }`
        }
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => {
          const actions = [];

          if (access.canAddRoles?.()) {
            actions.push(
              <DrawerForm<API.AddRoleRequest>
                key="add-role"
                title="为用户添加角色"
                trigger={
                  <Button type="primary">
                    <PlusOutlined /> 添加用户角色
                  </Button>
                }
                onFinish={handleAddRole}
                width={600}
              >
                <ProFormSelect
                  name="userKey"
                  label="用户"
                  placeholder="请选择用户"
                  rules={[{ required: true, message: "请选择用户" }]}
                  showSearch
                  options={userOptions.map((user: any) => ({
                    label: `${user.name} (${user.userName}) - ${user.email}`,
                    value: user.userKey,
                  }))}
                  fieldProps={{
                    filterOption: (input: any, option: any) =>
                      (option?.label ?? "")
                        .toLowerCase()
                        .includes(input.toLowerCase()),
                    onDropdownVisibleChange: (open) => {
                      if (open) {
                        loadUserOptions();
                      }
                    },
                    onSearch: (value) => {
                      loadUserOptions(value);
                    },
                  }}
                />
                <ProFormSelect
                  name="roleKey"
                  label="角色"
                  placeholder="选择角色类型"
                  options={roleOptions}
                  rules={[{ required: true, message: "请选择角色类型" }]}
                  fieldProps={{
                    onDropdownVisibleChange: (open) => {
                      if (open) {
                        loadRoleOptions();
                      }
                    },
                  }}
                />
                {canSelectTenant ? (
                  <ProFormSelect
                    name="tenantKey"
                    label="租户"
                    placeholder="请选择租户"
                    rules={[{ required: true, message: "请选择租户" }]}
                    showSearch
                    options={tenantOptions.map((tenant: any) => ({
                      label: `${tenant.tenantName} (${tenant.tenantKey})`,
                      value: tenant.tenantKey,
                    }))}
                    fieldProps={{
                      filterOption: (input: any, option: any) =>
                        (option?.label ?? "")
                          .toLowerCase()
                          .includes(input.toLowerCase()),
                      onDropdownVisibleChange: (open) => {
                        if (open) {
                          loadTenantOptions();
                        }
                      },
                      onSearch: (value) => {
                        loadTenantOptions(value);
                      },
                    }}
                  />
                ) : (
                  <ProFormSelect
                    name="tenantKey"
                    label="租户"
                    disabled
                    rules={[{ required: true, message: "请选择租户" }]}
                    initialValue={currentUserTenant?.tenantKey}
                    options={
                      currentUserTenant
                        ? [
                            {
                              label: `${currentUserTenant.tenantName} (${currentUserTenant.tenantKey})`,
                              value: currentUserTenant.tenantKey,
                            },
                          ]
                        : []
                    }
                  />
                )}
              </DrawerForm>
            );
          }

          return actions;
        }}
        request={async (params) => {
          try {
            const res = await getPermissionRules({
              current: params.current,
              pageSize: params.pageSize,
              userKey: params.userKey,
              tenantKey: params.tenantKey,
              ptype: "g", // 只获取角色规则
            });

            return {
              data: res.data || [],
              success: res.code === 0,
              total: res.total || 0,
            };
          } catch (_error) {
            message.error("获取用户角色列表失败");
            return { data: [], success: false, total: 0 };
          }
        }}
        columns={columns}
        scroll={{ x: 600 }}
      />
    </PageContainer>
  );
};

export default UserRoles;
