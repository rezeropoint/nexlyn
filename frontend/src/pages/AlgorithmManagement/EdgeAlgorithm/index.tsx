import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExperimentOutlined,
  InfoCircleOutlined,
  NodeIndexOutlined,
  PlayCircleOutlined,
  RightOutlined,
} from "@ant-design/icons";
import { PageContainer } from "@ant-design/pro-components";
import {
  Alert,
  Button,
  Card,
  Col,
  Row,
  Space,
  Statistic,
  Tag,
  Typography,
  theme,
} from "antd";
import classNames from "classnames";
import React from "react";
import { useNavigate } from "umi";
import styles from "./index.less";

const { Paragraph, Text, Title } = Typography;

/**
 * 边缘算法管理介绍页面
 * 展示AI Box设备管理的核心功能模块
 */
const EdgeAlgorithm: React.FC = () => {
  const { token } = theme.useToken();
  const navigate = useNavigate();

  const featureModules = [
    {
      id: "task-management",
      icon: <ExperimentOutlined />,
      title: "任务管理",
      description:
        "查询AI Box算法任务列表、获取算法能力详情、配置和控制算法任务。",
      path: "/algorithm-management/edge-algorithm/task-management",
      status: "已实现",
      features: [
        { id: "query", text: "任务查询" },
        { id: "capability", text: "能力获取" },
        { id: "config", text: "任务配置" },
        { id: "control", text: "任务控制" },
      ],
    },
    {
      id: "media-channel",
      icon: <PlayCircleOutlined />,
      title: "媒体通道",
      description: "管理AI Box的视频流输入，配置RTSP、RTMP等媒体源。",
      path: "/algorithm-management/edge-algorithm/media-channel",
      status: "开发中",
      features: [
        { id: "config", text: "通道配置" },
        { id: "query", text: "通道查询" },
        { id: "delete", text: "通道删除" },
      ],
    },
    {
      id: "system-settings",
      icon: <NodeIndexOutlined />,
      title: "系统设置",
      description: "系统配置、网络管理、参数设置、日志查询等系统级功能。",
      path: "/algorithm-management/edge-algorithm/system-settings",
      status: "开发中",
      features: [
        { id: "network", text: "网络配置" },
        { id: "params", text: "参数管理" },
        { id: "log", text: "日志查询" },
        { id: "face", text: "人脸识别" },
      ],
    },
  ];

  const protocolFeatures = [
    { id: "get-capability", name: "算法能力获取", implemented: true },
    { id: "config-media", name: "媒体通道配置", implemented: false },
    { id: "delete-media", name: "删除媒体通道", implemented: false },
    { id: "query-media", name: "查询媒体通道", implemented: false },
    { id: "config-task", name: "算法任务配置", implemented: false },
    { id: "control-task", name: "算法任务控制", implemented: false },
    { id: "delete-task", name: "算法任务删除", implemented: false },
    { id: "query-task", name: "算法任务查询", implemented: true },
    { id: "preview-task", name: "获取任务预览图", implemented: false },
    { id: "query-network", name: "查询盒子网络配置", implemented: false },
    { id: "system-cmd", name: "通用系统命令", implemented: false },
    { id: "config-param", name: "增改参数配置", implemented: false },
    { id: "delete-param", name: "删除用户参数", implemented: false },
    { id: "query-param", name: "查询参数配置", implemented: false },
    { id: "query-log", name: "查询调试日志", implemented: false },
    { id: "config-template", name: "获取配置模板信息", implemented: false },
    { id: "face-recognition", name: "人脸识别管理", implemented: false },
    { id: "device-network", name: "获取设备网络配置", implemented: false },
  ];

  const implementedCount = protocolFeatures.filter((f) => f.implemented).length;
  const totalCount = protocolFeatures.length;

  const performanceMetrics = [
    {
      id: "completion",
      metric: "协议完成度",
      value: `${implementedCount}/${totalCount}`,
      benchmark: "持续开发中",
    },
    {
      id: "latency",
      metric: "响应延迟",
      value: "< 100ms",
      benchmark: "MQTT实时通信",
    },
    {
      id: "concurrent",
      metric: "并发设备",
      value: "1000+",
      benchmark: "支持大规模接入",
    },
    {
      id: "reliability",
      metric: "可靠性",
      value: "99.9%",
      benchmark: "高可用保障",
    },
  ];

  return (
    <PageContainer>
      <Alert
        message="功能完善中"
        description="AI Box边缘算法管理平台正在开发完善中，当前已实现2/18个MQTT协议接口，敬请期待更多功能上线。"
        type="info"
        icon={<InfoCircleOutlined />}
        showIcon
        className={styles.marginBottom24}
      />

      {/* 顶部Banner */}
      <Card
        className={styles.bannerCard}
        styles={{
          body: {
            backgroundImage: `linear-gradient(75deg, ${token.colorError} 0%, ${token.colorErrorActive} 100%)`,
            color: "white",
          },
        }}
      >
        <div className={styles.bannerContent}>
          <div className={styles.bannerTitle}>边缘算法管理平台</div>
          <Paragraph className={styles.bannerDescription}>
            基于MQTT协议的AI
            Box设备管理平台，提供算法任务、媒体通道、系统配置等全方位管理能力。
            支持边缘侧智能分析，实现本地实时推理和快速响应。
          </Paragraph>
        </div>
      </Card>

      {/* 统计概览 */}
      <Row gutter={[16, 16]} className={styles.marginBottom24}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="功能模块"
              value={featureModules.length}
              prefix={<NodeIndexOutlined />}
              suffix="个"
              valueStyle={{
                fontSize: "24px",
                fontWeight: "bold",
                color: token.colorInfo,
              }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="协议接口"
              value={totalCount}
              prefix={<ExperimentOutlined />}
              suffix="个"
              valueStyle={{
                fontSize: "24px",
                fontWeight: "bold",
                color: token.colorSuccess,
              }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="已实现接口"
              value={implementedCount}
              prefix={<CheckCircleOutlined />}
              suffix={`/ ${totalCount}`}
              valueStyle={{
                fontSize: "24px",
                fontWeight: "bold",
                color: token.colorWarning,
              }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="完成进度"
              value={((implementedCount / totalCount) * 100).toFixed(1)}
              suffix="%"
              valueStyle={{
                fontSize: "24px",
                fontWeight: "bold",
                color: token.colorPrimary,
              }}
            />
          </Card>
        </Col>
      </Row>

      {/* 功能模块 */}
      <Card title="功能模块" className={styles.marginBottom24}>
        <Paragraph
          className={classNames(
            styles.moduleDescription,
            styles.marginBottom16
          )}
        >
          AI
          Box边缘算法管理平台提供完整的边缘侧算法管理能力，从任务配置到系统运维全方位覆盖。
        </Paragraph>
        <Row gutter={[16, 16]}>
          {featureModules.map((module) => (
            <Col xs={24} md={12} lg={8} key={module.id}>
              <Card
                hoverable={module.status === "已实现"}
                className={classNames(
                  styles.moduleCard,
                  module.status === "已实现"
                    ? styles.moduleCardClickable
                    : styles.moduleCardNonClickable
                )}
                onClick={() => {
                  if (module.status === "已实现" && module.path) {
                    navigate(module.path);
                  }
                }}
              >
                <Space direction="vertical" style={{ width: "100%" }}>
                  <div className={styles.moduleHeader}>
                    <Space>
                      {module.icon}
                      <strong>{module.title}</strong>
                    </Space>
                    <Tag
                      color={
                        module.status === "已实现"
                          ? "success"
                          : module.status === "开发中"
                          ? "processing"
                          : "default"
                      }
                    >
                      {module.status}
                    </Tag>
                  </div>
                  <Paragraph className={styles.moduleDescription}>
                    {module.description}
                  </Paragraph>
                  <Space wrap>
                    {module.features.map((feature) => (
                      <Tag key={feature.id}>{feature.text}</Tag>
                    ))}
                  </Space>
                  {module.status === "已实现" && (
                    <Button
                      type="link"
                      icon={<RightOutlined />}
                      className={styles.useButton}
                    >
                      立即使用
                    </Button>
                  )}
                </Space>
              </Card>
            </Col>
          ))}
        </Row>
      </Card>

      {/* 性能指标 */}
      <Card title="性能指标" className={styles.marginBottom24}>
        <Row gutter={[16, 16]}>
          {performanceMetrics.map((metric, index) => {
            // 使用主题色生成渐变
            const gradients = [
              `linear-gradient(135deg, ${token.colorPrimary} 0%, ${token.colorPrimaryActive} 100%)`,
              `linear-gradient(135deg, ${token.colorError} 0%, ${token.colorErrorActive} 100%)`,
              `linear-gradient(135deg, ${token.colorInfo} 0%, ${token.colorInfoActive} 100%)`,
              `linear-gradient(135deg, ${token.colorSuccess} 0%, ${token.colorSuccessActive} 100%)`,
            ];
            return (
              <Col xs={24} sm={12} md={6} key={metric.id}>
                <Card
                  className={styles.metricCard}
                  style={{
                    background: gradients[index % 4],
                  }}
                >
                  <Title level={4} className={styles.metricTitle}>
                    {metric.value}
                  </Title>
                  <div className={styles.metricName}>{metric.metric}</div>
                  <div className={styles.metricBenchmark}>
                    {metric.benchmark}
                  </div>
                </Card>
              </Col>
            );
          })}
        </Row>
      </Card>

      {/* MQTT协议接口列表 */}
      <Card title="MQTT协议接口实现进度" className={styles.marginBottom24}>
        <Paragraph
          className={classNames(
            styles.moduleDescription,
            styles.marginBottom16
          )}
        >
          基于MQTT协议实现与AI Box设备的实时通信，当前已实现{implementedCount}
          个接口，其余功能正在开发中。
        </Paragraph>
        <Row gutter={[16, 16]}>
          {protocolFeatures.map((feature) => (
            <Col xs={24} sm={12} md={8} lg={6} key={feature.id}>
              <Space>
                {feature.implemented ? (
                  <CheckCircleOutlined
                    className={styles.protocolIconImplemented}
                  />
                ) : (
                  <CloseCircleOutlined className={styles.protocolIconPending} />
                )}
                <Text type={feature.implemented ? "success" : "secondary"}>
                  {feature.name}
                </Text>
              </Space>
            </Col>
          ))}
        </Row>
      </Card>

      {/* 核心优势 */}
      <Card title="核心优势">
        <div className={styles.advantageContainer}>
          <Row gutter={[24, 24]}>
            <Col xs={24} md={12}>
              <Title level={5}>💡 边缘计算优势</Title>
              <Paragraph className={styles.advantageText}>
                • <Text strong>本地推理</Text>: 边缘侧实时计算，无需上传云端
                <br />• <Text strong>低延迟</Text>: 毫秒级响应，满足实时性要求
                <br />• <Text strong>隐私保护</Text>: 数据不出边缘，保护用户隐私
                <br />• <Text strong>降低成本</Text>: 减少带宽消耗和云端算力
              </Paragraph>
            </Col>
            <Col xs={24} md={12}>
              <Title level={5}>🚀 平台特色</Title>
              <Paragraph className={styles.advantageText}>
                • <Text strong>MQTT通信</Text>: 轻量级协议，适合边缘设备
                <br />• <Text strong>实时管理</Text>: 远程配置和控制AI Box
                <br />• <Text strong>多设备支持</Text>: 统一管理大规模边缘设备
                <br />• <Text strong>可扩展性</Text>: 灵活的功能模块设计
              </Paragraph>
            </Col>
          </Row>
        </div>
      </Card>
    </PageContainer>
  );
};

export default EdgeAlgorithm;
