import {
  ApiOutlined,
  CloudServerOutlined,
  InfoCircleOutlined,
  VideoCameraOutlined,
} from "@ant-design/icons";
import { PageContainer } from "@ant-design/pro-components";
import { Alert, Card, Col, List, Row, Space, Typography, theme } from "antd";
import React from "react";

const { Title, Paragraph, Text } = Typography;

/**
 * 媒体通道管理页面
 *
 * 功能：
 * - 媒体通道配置（2.1.2）
 * - 删除媒体通道（2.1.3）
 * - 查询媒体通道（2.1.4）
 */
const MediaChannel: React.FC = () => {
  const { token } = theme.useToken();

  const supportedProtocols = [
    {
      id: "rtsp",
      icon: <VideoCameraOutlined />,
      name: "RTSP",
      description: "实时流协议，广泛应用于网络摄像头",
      example: "rtsp://username:password@ip:port/stream",
    },
    {
      id: "rtmp",
      icon: <CloudServerOutlined />,
      name: "RTMP",
      description: "实时消息传输协议，适用于流媒体推送",
      example: "rtmp://server/live/stream",
    },
    {
      id: "gb28181",
      icon: <ApiOutlined />,
      name: "GB28181",
      description: "国标协议，用于安防监控系统",
      example: "通过国标设备ID接入",
    },
  ];

  const features = [
    {
      id: "config",
      title: "媒体通道配置",
      items: [
        "支持RTSP、RTMP、GB28181等主流协议",
        "灵活的通道参数配置",
        "支持认证和加密连接",
      ],
    },
    {
      id: "query",
      title: "查询媒体通道",
      items: ["实时显示通道状态", "查看通道详细信息", "监控流量和性能指标"],
    },
    {
      id: "delete",
      title: "删除媒体通道",
      items: ["支持单个/批量删除", "删除前二次确认", "自动释放资源"],
    },
  ];

  return (
    <PageContainer
      header={{
        title: "媒体通道",
        subTitle: "管理AI Box设备的视频流接入和配置",
      }}
    >
      <Alert
        message="功能开发中"
        description="媒体通道配置、查询和删除功能正在开发中，敬请期待。"
        type="info"
        icon={<InfoCircleOutlined />}
        showIcon
        style={{ marginBottom: 24 }}
      />

      {/* 支持的协议 */}
      <Card title="支持的协议" style={{ marginBottom: 24 }}>
        <Row gutter={[16, 16]}>
          {supportedProtocols.map((protocol) => (
            <Col xs={24} md={8} key={protocol.id}>
              <Card>
                <Space direction="vertical" style={{ width: "100%" }}>
                  <div>
                    <Space>
                      {protocol.icon}
                      <Title level={5} style={{ margin: 0 }}>
                        {protocol.name}
                      </Title>
                    </Space>
                  </div>
                  <Paragraph
                    style={{ color: token.colorTextSecondary, marginBottom: 8 }}
                  >
                    {protocol.description}
                  </Paragraph>
                  <Text code style={{ fontSize: "12px" }}>
                    {protocol.example}
                  </Text>
                </Space>
              </Card>
            </Col>
          ))}
        </Row>
      </Card>

      {/* 核心功能 */}
      <Card title="即将支持的功能">
        <Row gutter={[16, 16]}>
          {features.map((feature) => (
            <Col xs={24} md={8} key={feature.id}>
              <Card size="small" title={feature.title}>
                <List
                  size="small"
                  dataSource={feature.items}
                  renderItem={(item) => <List.Item>{item}</List.Item>}
                />
              </Card>
            </Col>
          ))}
        </Row>
      </Card>
    </PageContainer>
  );
};

export default MediaChannel;
