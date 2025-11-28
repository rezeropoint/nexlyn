import {
  AlertOutlined,
  ApiOutlined,
  CloudSyncOutlined,
  DatabaseOutlined,
  InfoCircleOutlined,
  MonitorOutlined,
  SecurityScanOutlined,
  SettingOutlined,
  WifiOutlined,
} from "@ant-design/icons";
import { PageContainer } from "@ant-design/pro-components";
import {
  Alert,
  Card,
  Flex,
  Progress,
  Space,
  Statistic,
  Tag,
  Timeline,
  Typography,
  theme,
} from "antd";
import React from "react";
import styles from "./index.less";

const { Title, Paragraph, Text } = Typography;

/**
 * 物联管理页面
 * 展示物联网设备全生命周期管理平台的核心功能和技术特性
 */
const IoTManagement: React.FC = () => {
  const { token } = theme.useToken();

  const deviceStats = [
    {
      title: "在线设备",
      value: 8924,
      prefix: <WifiOutlined className="icon-success" />,
      precision: 0,
    },
    {
      title: "数据采集点",
      value: 45678,
      prefix: <DatabaseOutlined className="icon-info" />,
      precision: 0,
    },
    {
      title: "告警事件",
      value: 156,
      prefix: <AlertOutlined className="icon-warning" />,
      precision: 0,
    },
    {
      title: "协议类型",
      value: 12,
      prefix: <ApiOutlined className="icon-primary" />,
      precision: 0,
    },
  ];

  const protocolSupport = [
    { name: "MQTT", status: "已支持", color: token.colorSuccess, usage: 85 },
    { name: "CoAP", status: "已支持", color: token.colorSuccess, usage: 72 },
    {
      name: "HTTP/HTTPS",
      status: "已支持",
      color: token.colorSuccess,
      usage: 95,
    },
    {
      name: "WebSocket",
      status: "已支持",
      color: token.colorSuccess,
      usage: 68,
    },
    { name: "Modbus", status: "已支持", color: token.colorSuccess, usage: 78 },
    { name: "OPC UA", status: "已支持", color: token.colorSuccess, usage: 45 },
    { name: "LoRaWAN", status: "已支持", color: token.colorSuccess, usage: 35 },
    { name: "NB-IoT", status: "开发中", color: token.colorWarning, usage: 20 },
  ];

  const coreCapabilities = [
    {
      icon: <WifiOutlined className="icon-info" style={{ fontSize: 32 }} />,
      title: "多协议设备接入",
      description:
        "支持MQTT、CoAP、HTTP、Modbus等主流物联网协议，提供统一的设备接入网关。",
      features: ["协议转换", "数据解析", "设备认证", "负载均衡"],
    },
    {
      icon: <DatabaseOutlined className="icon-success" style={{ fontSize: 32 }} />,
      title: "海量数据处理",
      description:
        "支持千万级设备并发接入，PB级数据存储，毫秒级实时数据处理能力。",
      features: ["时序数据库", "流式计算", "数据清洗", "智能压缩"],
    },
    {
      icon: <SecurityScanOutlined className="icon-warning" style={{ fontSize: 32 }} />,
      title: "设备安全管控",
      description: "提供设备身份认证、通信加密、权限管理等全方位安全保障机制。",
      features: ["身份认证", "TLS加密", "访问控制", "安全审计"],
    },
    {
      icon: <CloudSyncOutlined className="icon-primary" style={{ fontSize: 32 }} />,
      title: "远程运维管理",
      description:
        "支持设备远程配置、固件升级、故障诊断等运维操作，降低现场维护成本。",
      features: ["OTA升级", "远程诊断", "配置下发", "日志采集"],
    },
    {
      icon: <MonitorOutlined className="icon-error" style={{ fontSize: 32 }} />,
      title: "智能监控告警",
      description: "基于AI算法的设备状态监控，异常检测和预测性维护能力。",
      features: ["状态监控", "异常检测", "预测维护", "智能告警"],
    },
    {
      icon: <SettingOutlined className="icon-info-text" style={{ fontSize: 32 }} />,
      title: "边缘计算支持",
      description: "支持边缘设备本地计算，减少数据传输延迟，提升响应速度。",
      features: ["边缘网关", "本地计算", "数据预处理", "断网续传"],
    },
  ];

  const deviceLifecycle = [
    {
      title: "设备注册",
      description: "设备首次接入，完成身份验证和基础配置",
      color: token.colorInfo,
    },
    {
      title: "数据采集",
      description: "持续采集设备传感器数据和状态信息",
      color: token.colorSuccess,
    },
    {
      title: "数据处理",
      description: "对采集数据进行清洗、聚合和分析处理",
      color: token.colorWarning,
    },
    {
      title: "业务应用",
      description: "将处理后的数据用于业务决策和控制",
      color: token.colorPrimary,
    },
    {
      title: "运维管理",
      description: "设备状态监控、维护和升级管理",
      color: token.colorError,
    },
    {
      title: "设备退役",
      description: "设备生命周期结束，数据归档和安全清理",
      color: token.colorInfoText,
    },
  ];

  const industryApplications = [
    {
      industry: "智慧城市",
      description: "城市基础设施监控、环境监测、交通管理",
      devices: "路灯、传感器、监控设备",
      scale: "10万+ 设备",
    },
    {
      industry: "工业制造",
      description: "生产设备监控、质量检测、预测性维护",
      devices: "PLC、传感器、机器人",
      scale: "5万+ 设备",
    },
    {
      industry: "智慧农业",
      description: "环境监测、自动灌溉、病虫害预警",
      devices: "土壤传感器、气象站、无人机",
      scale: "2万+ 设备",
    },
    {
      industry: "智慧园区",
      description: "能耗管理、安防监控、环境调节",
      devices: "智能表计、门禁、空调",
      scale: "8万+ 设备",
    },
  ];

  return (
    <PageContainer>
      {/* 功能完善中提示 */}
      <Alert
        message="功能完善中"
        description="物联网设备管理平台正在开发完善中，敬请期待更多强大功能的上线。"
        type="info"
        icon={<InfoCircleOutlined />}
        showIcon
        className={styles.alertTip}
      />

      <Card className={styles.headerCard}>
        <div className={styles.headerContent}>
          <div className={styles.headerTitle}>物联网设备管理平台</div>
          <Paragraph className={styles.headerParagraph}>
            全方位的物联网设备生命周期管理平台，支持多协议设备接入、海量数据处理、智能监控告警。
            为智慧城市、工业制造、智慧农业等领域提供完整的IoT解决方案。
          </Paragraph>
        </div>
      </Card>

      {/* 统计概览 */}
      <Flex gap={16} wrap="wrap" className={styles.statsRow}>
        {deviceStats.map((stat) => (
          <div
            key={stat.title}
            style={{ flex: "0 0 calc(25% - 12px)", minWidth: "200px" }}
          >
            <Card className={styles.statisticCard}>
              <Statistic
                title={stat.title}
                value={stat.value}
                precision={stat.precision}
                prefix={stat.prefix}
              />
            </Card>
          </div>
        ))}
      </Flex>

      {/* 协议支持情况 */}
      <Card title="协议支持情况" className={styles.contentCard}>
        <Flex gap={16} wrap="wrap">
          {protocolSupport.map((protocol) => (
            <div
              key={protocol.name}
              style={{ flex: "0 0 calc(25% - 12px)", minWidth: "200px" }}
            >
              <Card size="small" className={styles.protocolCard}>
                <div className={styles.protocolName}>
                  <Text strong>{protocol.name}</Text>
                </div>
                <div className={styles.protocolTag}>
                  <Tag color={protocol.color}>{protocol.status}</Tag>
                </div>
                <Progress
                  percent={protocol.usage}
                  size="small"
                  strokeColor={protocol.color}
                  format={(percent) => `${percent}%`}
                />
              </Card>
            </div>
          ))}
        </Flex>
      </Card>

      {/* 核心能力 */}
      <Card title="核心技术能力" className={styles.sectionCard}>
        <Flex gap={24} wrap="wrap">
          {coreCapabilities.map((capability) => (
            <div
              key={capability.title}
              style={{ flex: "0 0 calc(33.33% - 16px)", minWidth: "280px" }}
            >
              <Card hoverable className={styles.capabilityCard}>
                <div className={styles.capabilityIcon}>{capability.icon}</div>
                <Title level={5} className={styles.capabilityTitle}>
                  {capability.title}
                </Title>
                <Paragraph className={styles.capabilityDescription}>
                  {capability.description}
                </Paragraph>
                <div className={styles.capabilityFeatures}>
                  <Space wrap>
                    {capability.features.map((feature) => (
                      <Tag key={feature} className={styles.featureTag}>
                        {feature}
                      </Tag>
                    ))}
                  </Space>
                </div>
              </Card>
            </div>
          ))}
        </Flex>
      </Card>

      {/* 设备生命周期 */}
      <Card title="设备全生命周期管理" className={styles.sectionCard}>
        <Timeline
          mode="alternate"
          items={deviceLifecycle.map((stage, _index) => ({
            color: stage.color,
            children: (
              <div className={styles.lifecycleStage}>
                <Title
                  level={5}
                  style={{ color: stage.color }}
                  className={styles.stageTitle}
                >
                  {stage.title}
                </Title>
                <Paragraph className={styles.lifecycleDescription}>
                  {stage.description}
                </Paragraph>
              </div>
            ),
          }))}
        />
      </Card>

      {/* 行业应用案例 */}
      <Card title="行业应用案例" className={styles.sectionCard}>
        <Flex gap={16} wrap="wrap">
          {industryApplications.map((app) => {
            // 根据行业名称确定颜色索引
            const colorIndex =
              app.industry === "智慧城市"
                ? 0
                : app.industry === "工业制造"
                ? 1
                : app.industry === "智慧农业"
                ? 2
                : 3;
            const gradientColors = [
              { start: token.colorPrimary, end: token.colorPrimaryActive }, // 主题色系
              { start: token.colorError, end: token.colorErrorActive }, // 错误色系
              { start: token.colorInfo, end: token.colorInfoActive }, // 信息色系
              { start: token.colorSuccess, end: token.colorSuccessActive }, // 成功色系
            ];
            const colors = gradientColors[colorIndex];

            return (
              <div
                key={app.industry}
                style={{ flex: "0 0 calc(50% - 8px)", minWidth: "300px" }}
              >
                <Card
                  className={styles.industryCard}
                  style={{
                    background: `linear-gradient(135deg, ${colors.start} 0%, ${colors.end} 100%)`,
                  }}
                >
                  <Title level={4} className={styles.industryTitle}>
                    {app.industry}
                  </Title>
                  <Paragraph className={styles.industryDescription}>
                    {app.description}
                  </Paragraph>
                  <div className={styles.industryInfo}>
                    <Text strong className={styles.industryLabel}>
                      设备类型:{" "}
                    </Text>
                    <Text className={styles.industryValue}>{app.devices}</Text>
                  </div>
                  <div>
                    <Text strong className={styles.industryLabel}>
                      部署规模:{" "}
                    </Text>
                    <Text className={styles.industryValue}>{app.scale}</Text>
                  </div>
                </Card>
              </div>
            );
          })}
        </Flex>
      </Card>

      {/* 技术架构优势 */}
      <Card title="技术架构优势">
        <Flex gap={24} wrap="wrap">
          <div style={{ flex: "0 0 calc(50% - 12px)", minWidth: "300px" }}>
            <Title level={5} className={styles.architectureTitle}>
              🌐 分布式架构设计
            </Title>
            <Paragraph className={styles.architectureParagraph}>
              采用微服务架构，支持水平扩展，单集群可支持千万级设备并发接入。
              提供多区域部署能力，满足数据本地化和低延迟要求。
            </Paragraph>

            <Title level={5} className={styles.architectureTitle}>
              ⚡ 高性能数据处理
            </Title>
            <Paragraph className={styles.architectureParagraph}>
              基于流式计算引擎，支持实时数据处理和复杂事件处理。
              集成时序数据库，提供高效的时间序列数据存储和查询能力。
            </Paragraph>
          </div>
          <div style={{ flex: "0 0 calc(50% - 12px)", minWidth: "300px" }}>
            <Title level={5} className={styles.architectureTitle}>
              🔒 企业级安全保障
            </Title>
            <Paragraph className={styles.architectureParagraph}>
              提供设备级安全认证、端到端数据加密、细粒度权限控制。
              支持国密算法和行业安全标准，满足关键基础设施安全要求。
            </Paragraph>

            <Title level={5} className={styles.architectureTitle}>
              📊 智能运维监控
            </Title>
            <Paragraph className={styles.architectureParagraph}>
              基于AI的设备健康状态评估，提供预测性维护建议。
              全链路监控体系，支持设备性能分析和故障快速定位。
            </Paragraph>
          </div>
        </Flex>
      </Card>
    </PageContainer>
  );
};

export default IoTManagement;
