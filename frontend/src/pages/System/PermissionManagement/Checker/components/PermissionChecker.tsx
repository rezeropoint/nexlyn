import { checkPermission } from "@/services/permission";
import { getTenantOptions } from "@/services/tenant";
import { getUserOptions } from "@/services/user";
import { CheckCircleOutlined, CloseCircleOutlined } from "@ant-design/icons";
import {
  ProForm,
  type ProFormInstance,
  ProFormSelect,
} from "@ant-design/pro-components";
import { useAccess, useModel } from "@umijs/max";
import {
  Alert,
  Button,
  Card,
  Col,
  Row,
  Select,
  Space,
  Tag,
  Typography,
} from "antd";
import React, { useEffect, useRef, useState } from "react";
import styles from "./PermissionChecker.less";

const { Title, Text } = Typography;

interface PermissionCheckResult {
  hasPermission: boolean;
  userKey: string;
  tenantKey: string;
  resource: string;
  action: string;
}

/**
 * 权限检查工具组件
 */
const PermissionChecker: React.FC = () => {
  const access = useAccess();
  const { initialState } = useModel("@@initialState");
  const formRef = useRef<ProFormInstance>();
  const [loading, setLoading] = useState(false);
  const [checkResult, setCheckResult] = useState<PermissionCheckResult | null>(
    null
  );
  const [selectedResource, setSelectedResource] = useState<string>("");
  const [currentUserTenant, setCurrentUserTenant] = useState<{
    tenantKey: string;
    tenantName: string;
  } | null>(null);
  const [canSelectTenant, setCanSelectTenant] = useState<boolean>(false);

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
    // 只有超级管理员可以选择租户；其他用户只能检查自己租户的权限
    const hasMultiTenantAccess = access.isSuperAdmin?.() || false;
    setCanSelectTenant(hasMultiTenantAccess);
  }, [access]);

  // 设置默认租户值
  useEffect(() => {
    if (!canSelectTenant && currentUserTenant && formRef.current) {
      formRef.current.setFieldsValue({
        tenantKey: currentUserTenant.tenantKey,
      });
    }
  }, [canSelectTenant, currentUserTenant]);

  // 处理权限检查
  const handleCheck = async (values: any) => {
    setLoading(true);
    try {
      // 如果没有选择租户（普通管理员），使用当前用户的租户
      const tenantKey = values.tenantKey || currentUserTenant?.tenantKey;

      if (!tenantKey) {
        console.error("无法获取租户信息");
        setCheckResult(null);
        return;
      }

      const response = await checkPermission({
        userKey: values.userKey,
        tenantKey: tenantKey,
        resource: values.resource,
        action: values.action,
      });

      if (response.code === 0) {
        setCheckResult({
          hasPermission: response.hasPermission || false,
          userKey: values.userKey,
          tenantKey: tenantKey,
          resource: values.resource,
          action: values.action,
        });
      } else {
        setCheckResult(null);
      }
    } catch (error) {
      console.error("权限检查失败:", error);
      setCheckResult(null);
    } finally {
      setLoading(false);
    }
  };

  // 重置表单和结果
  const handleReset = () => {
    formRef.current?.resetFields();
    setCheckResult(null);
    setSelectedResource("");
  };

  // 获取操作选项（基于资源类型）
  const getActionOptions = (resource: string) => {
    const baseActions = [
      { label: "读取 (read)", value: "read" },
      { label: "写入 (write)", value: "write" },
      { label: "删除 (delete)", value: "delete" },
      { label: "管理 (admin)", value: "admin" },
    ];

    // 根据资源类型过滤可用操作
    if (resource.includes(":")) {
      const resourceType = resource.split(":")[0];
      switch (resourceType) {
        case "user":
          return baseActions;
        case "tenant":
          return baseActions;
        case "graph":
          return baseActions;
        case "infoatom":
          return baseActions.filter((action) => action.value !== "admin");
        default:
          return baseActions;
      }
    }

    return baseActions;
  };

  return (
    <div>
      <Card
        title="权限检查"
        extra={<Button onClick={handleReset}>重置</Button>}
      >
        <ProForm
          formRef={formRef}
          layout="vertical"
          onFinish={handleCheck}
          submitter={{
            render: () => (
              <Space>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={loading}
                  icon={<CheckCircleOutlined />}
                >
                  检查权限
                </Button>
              </Space>
            ),
          }}
        >
          <Row gutter={16}>
            <Col span={12}>
              <ProFormSelect
                name="userKey"
                label="选择用户"
                placeholder="请选择要检查的用户"
                request={async () => {
                  try {
                    const response = await getUserOptions({});
                    if (response.code === 0 && response.data?.list) {
                      return response.data.list.map((user: any) => ({
                        label: `${user.userName} (${user.name})`,
                        value: user.userKey,
                      }));
                    }
                    return [];
                  } catch (error) {
                    console.error("获取用户选项失败:", error);
                    return [];
                  }
                }}
                rules={[{ required: true, message: "请选择用户" }]}
                showSearch
              />
            </Col>
            <Col span={12}>
              {canSelectTenant ? (
                <ProFormSelect
                  name="tenantKey"
                  label="选择租户"
                  placeholder="请选择租户"
                  request={async () => {
                    try {
                      const response = await getTenantOptions({});
                      if (response.code === 0 && response.data?.list) {
                        return response.data.list.map((tenant: any) => ({
                          label: tenant.tenantName,
                          value: tenant.tenantKey,
                        }));
                      }
                      return [];
                    } catch (error) {
                      console.error("获取租户选项失败:", error);
                      return [];
                    }
                  }}
                  rules={[{ required: true, message: "请选择租户" }]}
                  showSearch
                />
              ) : (
                <ProForm.Item
                  name="tenantKey"
                  label="租户"
                  rules={[{ required: true, message: "请选择租户" }]}
                >
                  <Select
                    placeholder="当前租户"
                    disabled
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
                </ProForm.Item>
              )}
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <ProForm.Item
                name="resource"
                label="资源"
                rules={[{ required: true, message: "请选择或输入资源" }]}
              >
                <Select
                  placeholder="请选择或输入资源"
                  showSearch
                  allowClear
                  onChange={setSelectedResource}
                  filterOption={(input, option) =>
                    (option?.label ?? "")
                      .toLowerCase()
                      .includes(input.toLowerCase())
                  }
                  options={[
                    { label: "用户管理 (user)", value: "user" },
                    { label: "用户读取 (user:read)", value: "user:read" },
                    { label: "用户写入 (user:write)", value: "user:write" },
                    { label: "租户管理 (tenant)", value: "tenant" },
                    { label: "租户读取 (tenant:read)", value: "tenant:read" },
                    { label: "租户写入 (tenant:write)", value: "tenant:write" },
                    { label: "图配置管理 (graph)", value: "graph" },
                    { label: "图配置读取 (graph:read)", value: "graph:read" },
                    { label: "图配置写入 (graph:write)", value: "graph:write" },
                    { label: "信息原子管理 (infoatom)", value: "infoatom" },
                    { label: "标签管理 (tags)", value: "tags" },
                    { label: "权限管理 (permission)", value: "permission" },
                  ]}
                />
              </ProForm.Item>
            </Col>
            <Col span={12}>
              <ProForm.Item
                name="action"
                label="操作"
                rules={[{ required: true, message: "请选择操作" }]}
              >
                <Select
                  placeholder="请选择操作"
                  options={getActionOptions(selectedResource)}
                />
              </ProForm.Item>
            </Col>
          </Row>
        </ProForm>

        {/* 检查结果 */}
        {checkResult && (
          <Card title="检查结果" className={styles.resultCard} type="inner">
            <Space direction="vertical" className={styles.resultDetails}>
              <Alert
                message={
                  <Space>
                    {checkResult.hasPermission ? (
                      <CheckCircleOutlined className={styles.successIcon} />
                    ) : (
                      <CloseCircleOutlined className={styles.errorIcon} />
                    )}
                    <Text strong>
                      {checkResult.hasPermission ? "有权限" : "无权限"}
                    </Text>
                  </Space>
                }
                type={checkResult.hasPermission ? "success" : "error"}
                showIcon
              />

              <div>
                <Title level={5}>检查详情</Title>
                <Space direction="vertical">
                  <div>
                    <Text strong>用户Key：</Text>
                    <Tag color="blue">{checkResult.userKey}</Tag>
                  </div>
                  <div>
                    <Text strong>租户Key：</Text>
                    <Tag color="green">{checkResult.tenantKey}</Tag>
                  </div>
                  <div>
                    <Text strong>资源：</Text>
                    <Tag color="orange">{checkResult.resource}</Tag>
                  </div>
                  <div>
                    <Text strong>操作：</Text>
                    <Tag color="purple">{checkResult.action}</Tag>
                  </div>
                </Space>
              </div>
            </Space>
          </Card>
        )}
      </Card>
    </div>
  );
};

export default PermissionChecker;
