import {
  getPlatformConfig,
  savePlatformConfig,
  testPlatformConnection,
} from "@/services/eventhandler";
import {
  ApiOutlined,
  CheckCircleOutlined,
  InfoCircleOutlined,
  SaveOutlined,
} from "@ant-design/icons";
import {
  PageContainer,
  ProCard,
  ProForm,
  ProFormDigit,
  ProFormText,
} from "@ant-design/pro-components";
import { Alert, App, Button, Descriptions, Flex, Space, Spin } from "antd";
import React, { useEffect, useRef, useState } from "react";
import type { ProFormInstance } from "@ant-design/pro-components";
import { DEFAULT_VALUES, FORM_RULES, MESSAGE } from "../constants";
import type { PlatformConfig, SavePlatformConfigRequest } from "../types";
import styles from "./index.less";

/**
 * 平台配置页面
 * 管理Skylark数据库连接配置
 */
const PlatformConfigPage: React.FC = () => {
  const { message } = App.useApp();
  const formRef = useRef<ProFormInstance>();
  const [loading, setLoading] = useState(false);
  const [testingConnection, setTestingConnection] = useState(false);
  const [config, setConfig] = useState<PlatformConfig | null>(null);
  const [connectionStatus, setConnectionStatus] = useState<{
    success?: boolean;
    message?: string;
    flowsCount?: number;
  } | null>(null);

  // 加载现有配置
  const loadConfig = async () => {
    setLoading(true);
    try {
      const response = await getPlatformConfig();
      if (response.code === 0 && response.data) {
        setConfig(response.data);
        // 设置表单值（不包含密码和Token）
        formRef.current?.setFieldsValue({
          host: response.data.host,
          port: response.data.port,
          database: response.data.database,
          username: response.data.username,
          namespaceId: response.data.namespaceId,
          apiBaseUrl: response.data.apiBaseUrl,
        });
      }
    } catch (error) {
      console.error("加载配置失败:", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadConfig();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // 保存配置
  const handleSave = async (values: SavePlatformConfigRequest) => {
    try {
      const response = await savePlatformConfig(values);
      if (response.code === 0) {
        message.success(MESSAGE.SAVE_SUCCESS);
        await loadConfig();
        setConnectionStatus(null);
      } else {
        message.error(response.msg || MESSAGE.SAVE_FAILED);
      }
    } catch (error) {
      console.error("保存配置失败:", error);
      message.error(MESSAGE.SAVE_FAILED);
    }
  };

  // 测试连接
  const handleTestConnection = async () => {
    const values = await formRef.current?.validateFields();
    if (!values) return;
    setTestingConnection(true);
    setConnectionStatus(null);

    try {
      const response = await testPlatformConnection(values);
      if (response.code === 0 && response.data) {
        setConnectionStatus(response.data);
        if (response.data.success) {
          message.success(MESSAGE.TEST_CONNECTION_SUCCESS);
        } else {
          message.error(
            response.data.message || MESSAGE.TEST_CONNECTION_FAILED
          );
        }
      } else {
        message.error(response.msg || MESSAGE.TEST_CONNECTION_FAILED);
        setConnectionStatus({
          success: false,
          message: response.msg || "连接失败",
        });
      }
    } catch (error) {
      console.error("测试连接失败:", error);
      message.error(MESSAGE.TEST_CONNECTION_FAILED);
      setConnectionStatus({
        success: false,
        message: "网络错误或服务不可用",
      });
    } finally {
      setTestingConnection(false);
    }
  };

  return (
    <PageContainer>
      <Spin spinning={loading}>
        <ProCard ghost>
          <Flex gap={16} wrap="wrap">
            <div className={styles.configCardLeft}>
              <ProCard
                title="当前配置"
                headerBordered
                className={styles.configCard}
              >
                {config ? (
                  <Descriptions bordered column={1} size="small">
                    <Descriptions.Item label="数据库地址">
                      {config.host}
                    </Descriptions.Item>
                    <Descriptions.Item label="端口">
                      {config.port}
                    </Descriptions.Item>
                    <Descriptions.Item label="数据库名">
                      {config.database}
                    </Descriptions.Item>
                    <Descriptions.Item label="用户名">
                      {config.username}
                    </Descriptions.Item>
                    <Descriptions.Item label="命名空间ID">
                      {config.namespaceId}
                    </Descriptions.Item>
                    <Descriptions.Item label="API基础地址">
                      {config.apiBaseUrl}
                    </Descriptions.Item>
                    <Descriptions.Item label="API Token">
                      已配置
                    </Descriptions.Item>
                    <Descriptions.Item label="更新时间">
                      {config.updatedAt}
                    </Descriptions.Item>
                  </Descriptions>
                ) : (
                  <Alert
                    message="暂无配置"
                    description="尚未配置Skylark平台，请在右侧表单中填写配置信息。"
                    type="info"
                    showIcon
                  />
                )}

                {connectionStatus && (
                  <Alert
                    message={
                      connectionStatus.success ? (
                        <Space>
                          <CheckCircleOutlined />
                          连接成功
                        </Space>
                      ) : (
                        "连接失败"
                      )
                    }
                    description={
                      connectionStatus.success
                        ? `成功连接到Skylark平台，发现 ${
                            connectionStatus.flowsCount || 0
                          } 个流程`
                        : connectionStatus.message || "无法连接到数据库"
                    }
                    type={connectionStatus.success ? "success" : "error"}
                    showIcon
                    className={styles.connectionStatus}
                  />
                )}
              </ProCard>
            </div>
            <div className={styles.configCardRight}>
              <ProCard title="编辑配置" headerBordered>
                <Alert
                  message="配置说明"
                  description="请配置Skylark平台的PostgreSQL数据库连接信息。配置保存后，可以测试连接是否正常。"
                  type="info"
                  showIcon
                  icon={<InfoCircleOutlined />}
                  className={styles.configDescription}
                />
                <ProForm
                  formRef={formRef}
                  onFinish={handleSave}
                  submitter={{
                    render: () => (
                      <Space>
                        <Button
                          type="primary"
                          icon={<SaveOutlined />}
                          onClick={() => formRef.current?.submit()}
                        >
                          保存配置
                        </Button>
                        <Button
                          icon={<ApiOutlined />}
                          onClick={handleTestConnection}
                          loading={testingConnection}
                        >
                          测试连接
                        </Button>
                      </Space>
                    ),
                  }}
                  layout="vertical"
                >
                  <Flex gap={16}>
                    <div className={styles.formFieldEqual}>
                      <ProFormText
                        name="host"
                        label="数据库地址"
                        placeholder="例如: localhost 或 192.168.1.100"
                        rules={FORM_RULES.HOST}
                        tooltip="Skylark平台PostgreSQL数据库的IP地址或域名"
                      />
                    </div>
                    <div className={styles.formFieldEqual}>
                      <ProFormDigit
                        name="port"
                        label="端口"
                        placeholder="默认: 5432"
                        initialValue={DEFAULT_VALUES.PORT}
                        rules={FORM_RULES.PORT}
                        fieldProps={{
                          min: 1,
                          max: 65535,
                          className: styles.fullWidthInput,
                        }}
                      />
                    </div>
                  </Flex>
                  <Flex gap={16}>
                    <div className={styles.formFieldEqual}>
                      <ProFormText
                        name="database"
                        label="数据库名"
                        placeholder="例如: skylark"
                        rules={FORM_RULES.DATABASE}
                        tooltip="要连接的数据库名称"
                      />
                    </div>
                    <div className={styles.formFieldEqual}>
                      <ProFormDigit
                        name="namespaceId"
                        label="命名空间ID"
                        placeholder="默认: 1"
                        initialValue={DEFAULT_VALUES.NAMESPACE_ID}
                        rules={[
                          { required: true, message: "请输入命名空间ID" },
                        ]}
                        tooltip="Skylark平台的namespace_id，用于租户隔离"
                        fieldProps={{
                          min: 1,
                          className: styles.fullWidthInput,
                        }}
                      />
                    </div>
                  </Flex>
                  <Flex gap={16}>
                    <div className={styles.formFieldEqual}>
                      <ProFormText
                        name="username"
                        label="用户名"
                        placeholder="数据库用户名"
                        rules={FORM_RULES.USERNAME}
                      />
                    </div>
                    <div className={styles.formFieldEqual}>
                      <ProFormText.Password
                        name="password"
                        label="密码"
                        placeholder="数据库密码"
                        rules={FORM_RULES.PASSWORD}
                        tooltip="密码将加密存储，查询时不会返回"
                      />
                    </div>
                  </Flex>
                  <Flex gap={16}>
                    <div className={styles.formFieldEqual}>
                      <ProFormText
                        name="apiBaseUrl"
                        label="API基础地址"
                        placeholder="例如: skylark.example.com"
                        rules={FORM_RULES.API_BASE_URL}
                        tooltip="Skylark平台API基础地址（纯域名，不包含http://或https://）"
                      />
                    </div>
                    <div className={styles.formFieldEqual}>
                      <ProFormText.Password
                        name="apiToken"
                        label="API Token"
                        placeholder="API认证令牌"
                        rules={FORM_RULES.API_TOKEN}
                        tooltip="API认证Token将加密存储，查询时不会返回"
                      />
                    </div>
                  </Flex>
                </ProForm>
              </ProCard>
            </div>
          </Flex>
        </ProCard>
      </Spin>
    </PageContainer>
  );
};

export default PlatformConfigPage;
