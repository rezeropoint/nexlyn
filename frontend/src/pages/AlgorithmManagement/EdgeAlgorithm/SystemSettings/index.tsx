import {
  FileTextOutlined,
  InfoCircleOutlined,
  SettingOutlined,
  UserOutlined,
  WifiOutlined,
} from "@ant-design/icons";
import { PageContainer } from "@ant-design/pro-components";
import { Alert, Card, Col, List, Row, Space } from "antd";
import React from "react";

/**
 * 系统设置页面
 *
 * 功能：
 * - 查询盒子网络配置（2.1.10）
 * - 获取设备网络配置（2.1.18）
 * - 通用系统命令（2.1.11）
 * - 增改参数配置（2.1.12）
 * - 删除用户参数（2.1.13）
 * - 查询参数配置（2.1.14）
 * - 查询调试日志（2.1.15）
 * - 获取当前配置的模板信息（2.1.16）
 * - 人脸识别统一管理接口（2.1.17）
 */
const SystemSettings: React.FC = () => {
  const featureCategories = [
    {
      id: "network",
      icon: <WifiOutlined />,
      title: "网络配置",
      items: ["查询盒子网络配置", "获取设备网络配置"],
    },
    {
      id: "params",
      icon: <SettingOutlined />,
      title: "参数管理",
      items: ["增改参数配置", "删除用户参数", "查询参数配置"],
    },
    {
      id: "ops",
      icon: <FileTextOutlined />,
      title: "系统运维",
      items: ["通用系统命令", "查询调试日志", "获取配置模板信息"],
    },
    {
      id: "ai",
      icon: <UserOutlined />,
      title: "智能功能",
      items: ["人脸识别统一管理"],
    },
  ];

  return (
    <PageContainer
      header={{
        title: "系统设置",
        subTitle: "AI Box系统配置、参数管理和日志查询",
      }}
    >
      <Alert
        message="功能开发中"
        description="系统配置、参数管理、日志查询等功能正在开发中，敬请期待。"
        type="info"
        icon={<InfoCircleOutlined />}
        showIcon
        style={{ marginBottom: 24 }}
      />

      <Row gutter={[16, 16]}>
        {featureCategories.map((category) => (
          <Col xs={24} md={12} key={category.id}>
            <Card
              title={
                <Space>
                  {category.icon}
                  <span>{category.title}</span>
                </Space>
              }
              style={{ height: "100%" }}
            >
              <List
                size="small"
                dataSource={category.items}
                renderItem={(item) => <List.Item>{item}</List.Item>}
              />
            </Card>
          </Col>
        ))}
      </Row>
    </PageContainer>
  );
};

export default SystemSettings;
