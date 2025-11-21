import {
  CloudDownloadOutlined,
  CompressOutlined,
  InfoCircleOutlined,
  MobileOutlined,
  SafetyCertificateOutlined,
  SettingOutlined,
  SyncOutlined,
  ThunderboltOutlined,
  WifiOutlined,
} from "@ant-design/icons";
import { PageContainer } from "@ant-design/pro-components";
import {
  Alert,
  Card,
  Flex,
  Space,
  Statistic,
  Steps,
  Typography,
  theme,
} from "antd";
import React from "react";
import styles from "./index.less";

const { Paragraph, Title, Text } = Typography;

/**
 * 端侧算法管理介绍页面
 * 展示摄像头内置算法的管理功能
 */
const DeviceAlgorithm: React.FC = () => {
  const { token } = theme.useToken();

  const features = [
    {
      id: "builtin",
      title: "内置算法",
      description: "摄像头内置AI芯片运行的算法",
      items: [
        { id: "face", text: "人脸识别" },
        { id: "object", text: "目标检测" },
        { id: "behavior", text: "行为分析" },
        { id: "plate", text: "车牌识别" },
      ],
      color: token.colorPrimary,
      icon: (
        <CompressOutlined style={{ fontSize: 20, color: token.colorPrimary }} />
      ),
    },
    {
      id: "inference",
      title: "本地推理",
      description: "摄像头端侧实时计算，无需回传",
      items: [
        { id: "response", text: "毫秒级响应" },
        { id: "network", text: "减少网络传输" },
        { id: "privacy", text: "保护隐私数据" },
        { id: "bandwidth", text: "降低带宽成本" },
      ],
      color: token.colorSuccess,
      icon: (
        <WifiOutlined style={{ fontSize: 20, color: token.colorSuccess }} />
      ),
    },
    {
      id: "management",
      title: "算法管理",
      description: "远程管理摄像头内置算法",
      items: [
        { id: "deploy", text: "算法下发部署" },
        { id: "config", text: "参数远程配置" },
        { id: "monitor", text: "性能监控" },
        { id: "version", text: "版本管理" },
      ],
      color: token.colorWarning,
      icon: (
        <SyncOutlined style={{ fontSize: 20, color: token.colorWarning }} />
      ),
    },
  ];

  const workflowSteps = [
    {
      title: "算法开发",
      description: "基于需求开发AI算法",
      icon: <CompressOutlined />,
    },
    {
      title: "模型优化",
      description: "适配摄像头芯片",
      icon: <SyncOutlined />,
    },
    {
      title: "算法下发",
      description: "远程部署到摄像头",
      icon: <MobileOutlined />,
    },
    {
      title: "参数配置",
      description: "调整算法参数",
      icon: <SettingOutlined />,
    },
    {
      title: "性能监控",
      description: "监控算法运行状态",
      icon: <ThunderboltOutlined />,
    },
  ];

  const performanceMetrics = [
    { id: "speed", metric: "推理速度", value: "< 50ms", benchmark: "实时响应" },
    {
      id: "accuracy",
      metric: "识别准确率",
      value: "> 95%",
      benchmark: "高精度",
    },
    {
      id: "concurrent",
      metric: "并发能力",
      value: "16路",
      benchmark: "多路视频",
    },
    { id: "offline", metric: "离线运行", value: "100%", benchmark: "无需回传" },
  ];

  return (
    <PageContainer>
      <Alert
        message="功能完善中"
        description="摄像头内置算法管理平台正在开发完善中，敬请期待智能摄像头算法下发和管理能力。"
        type="info"
        icon={<InfoCircleOutlined />}
        showIcon
        style={{ marginBottom: 24 }}
      />

      {/* 顶部Banner */}
      <Card className={styles.banner}>
        <div className={styles.bannerContent}>
          <div className={styles.bannerTitle}>端侧算法管理</div>
          <Paragraph className={styles.bannerDescription}>
            为智能摄像头提供内置算法管理能力，支持算法远程下发、参数配置、性能监控等功能。
            通过摄像头内置AI芯片实现本地实时推理，降低网络带宽消耗，保护数据隐私。
          </Paragraph>
        </div>
      </Card>

      {/* 统计概览 */}
      <Flex gap={16} wrap="wrap" style={{ marginBottom: 24 }}>
        <div style={{ flex: "1 1 calc(25% - 12px)", minWidth: 200 }}>
          <Card>
            <Statistic
              title="支持设备类型"
              value={3}
              prefix={<MobileOutlined style={{ color: token.colorPrimary }} />}
              suffix="类"
              valueStyle={{ fontSize: "24px", fontWeight: "bold" }}
            />
          </Card>
        </div>
        <div style={{ flex: "1 1 calc(25% - 12px)", minWidth: 200 }}>
          <Card>
            <Statistic
              title="模型压缩率"
              value={80}
              prefix={
                <CloudDownloadOutlined style={{ color: token.colorSuccess }} />
              }
              suffix="%"
              valueStyle={{ fontSize: "24px", fontWeight: "bold" }}
            />
          </Card>
        </div>
        <div style={{ flex: "1 1 calc(25% - 12px)", minWidth: 200 }}>
          <Card>
            <Statistic
              title="推理速度"
              value={50}
              prefix={
                <ThunderboltOutlined style={{ color: token.colorWarning }} />
              }
              suffix="ms"
              valueStyle={{ fontSize: "24px", fontWeight: "bold" }}
            />
          </Card>
        </div>
        <div style={{ flex: "1 1 calc(25% - 12px)", minWidth: 200 }}>
          <Card>
            <Statistic
              title="功耗降低"
              value={60}
              prefix={
                <SafetyCertificateOutlined
                  style={{ color: token.colorInfoText }}
                />
              }
              suffix="%"
              valueStyle={{ fontSize: "24px", fontWeight: "bold" }}
            />
          </Card>
        </div>
      </Flex>

      {/* 核心特性 */}
      <Card title="核心特性" style={{ marginBottom: 24 }}>
        <Paragraph
          style={{ color: token.colorTextSecondary, marginBottom: 16 }}
        >
          端侧算法管理针对资源受限的终端设备，提供轻量化模型部署和高效推理能力。
        </Paragraph>
        <Flex gap={16} wrap="wrap">
          {features.map((feature) => (
            <div
              key={feature.id}
              style={{ flex: "1 1 calc(33.33% - 11px)", minWidth: 280 }}
            >
              <Card
                size="small"
                hoverable
                className={styles.featureCard}
                style={{ borderLeft: `4px solid ${feature.color}` }}
              >
                <div className={styles.featureHeader}>
                  {feature.icon}
                  <Text
                    strong
                    className={styles.featureIcon}
                    style={{ color: feature.color }}
                  >
                    {feature.title}
                  </Text>
                </div>
                <Paragraph className={styles.featureDescription}>
                  {feature.description}
                </Paragraph>
                <Space wrap>
                  {feature.items.map((item) => (
                    <span key={item.id} className={styles.featureItem}>
                      • {item.text}
                    </span>
                  ))}
                </Space>
              </Card>
            </div>
          ))}
        </Flex>
      </Card>

      {/* 部署流程 */}
      <Card title="端侧部署流程" style={{ marginBottom: 24 }}>
        <Steps
          current={-1}
          items={workflowSteps.map((step) => ({
            title: step.title,
            description: step.description,
            icon: step.icon,
          }))}
        />
        <div className={styles.workflowTip}>
          <Text strong>部署策略: </Text>
          <Text style={{ color: token.colorTextSecondary }}>
            端侧算法部署强调轻量化和高效性，通过模型压缩、格式转换、芯片适配等步骤，
            确保AI算法能在资源受限的终端设备上流畅运行，实现边缘智能。
          </Text>
        </div>
      </Card>

      {/* 性能指标 */}
      <Card title="性能指标" style={{ marginBottom: 24 }}>
        <Flex gap={16} wrap="wrap">
          {performanceMetrics.map((metric, index) => (
            <div
              key={metric.id}
              style={{ flex: "1 1 calc(25% - 12px)", minWidth: 200 }}
            >
              <Card
                className={`${styles.metricCard} ${
                  styles[`metricCard${index % 4}`]
                }`}
              >
                <Title level={4} className={styles.metricValue}>
                  {metric.value}
                </Title>
                <div className={styles.metricLabel}>{metric.metric}</div>
                <div className={styles.metricBenchmark}>{metric.benchmark}</div>
              </Card>
            </div>
          ))}
        </Flex>
      </Card>

      {/* 应用场景 */}
      <Card title="典型应用场景">
        <div className={styles.scenarioSection}>
          <Flex gap={24} wrap="wrap">
            <div style={{ flex: "1 1 calc(33.33% - 16px)", minWidth: 280 }}>
              <Title level={5}>🔒 智慧安防</Title>
              <Paragraph style={{ color: token.colorTextSecondary }}>
                • <Text strong>周界防护</Text>: 入侵检测、徘徊检测、翻越围栏
                <br />• <Text strong>人脸识别</Text>:
                黑名单布控、VIP识别、访客管理
                <br />• <Text strong>行为分析</Text>:
                打架斗殴、摔倒检测、聚众检测
                <br />• <Text strong>区域管控</Text>:
                区域入侵、禁入区域、人员计数
              </Paragraph>
            </div>
            <div style={{ flex: "1 1 calc(33.33% - 16px)", minWidth: 280 }}>
              <Title level={5}>🏢 智慧园区</Title>
              <Paragraph style={{ color: token.colorTextSecondary }}>
                • <Text strong>车辆管理</Text>: 车牌识别、违停检测、车位占用
                <br />• <Text strong>人员管理</Text>:
                员工考勤、访客识别、人员轨迹
                <br />• <Text strong>环境监测</Text>:
                烟火检测、垃圾识别、积水检测
                <br />• <Text strong>运营分析</Text>:
                人流统计、热力图分析、停留时长
              </Paragraph>
            </div>
            <div style={{ flex: "1 1 calc(33.33% - 16px)", minWidth: 280 }}>
              <Title level={5}>🚗 智慧交通</Title>
              <Paragraph style={{ color: token.colorTextSecondary }}>
                • <Text strong>交通违章</Text>: 闯红灯、压线、逆行、违停
                <br />• <Text strong>车流监控</Text>:
                车流统计、拥堵检测、车速测算
                <br />• <Text strong>事件检测</Text>: 事故检测、抛洒物、行人横穿
                <br />• <Text strong>车辆识别</Text>:
                车牌识别、车型识别、车辆特征
              </Paragraph>
            </div>
          </Flex>
        </div>
      </Card>
    </PageContainer>
  );
};

export default DeviceAlgorithm;
