import {
  CloudServerOutlined,
  ControlOutlined,
  DeploymentUnitOutlined,
  ExperimentOutlined,
  InfoCircleOutlined,
  NodeIndexOutlined,
  RobotOutlined,
  ThunderboltOutlined,
} from "@ant-design/icons";
import { PageContainer } from "@ant-design/pro-components";
import {
  Alert,
  Card,
  Col,
  Row,
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
 * 算法管理主页面
 * 展示云边端一体化AI算法管理平台的核心功能
 */
const AlgorithmManagement: React.FC = () => {
  const { token } = theme.useToken();

  // 使用主题颜色
  const themeColors = {
    primary: token.colorPrimary,
    success: token.colorSuccess,
    warning: token.colorWarning,
    purple: token.purple,
    magenta: token.magenta,
    cyan: token.cyan,
  };

  const algorithmStats = [
    {
      id: "edge",
      title: "边缘算法",
      value: 42,
      prefix: <NodeIndexOutlined className="icon-primary" />,
      precision: 0,
    },
    {
      id: "cloud",
      title: "云端算法",
      value: 128,
      prefix: <CloudServerOutlined className="icon-success" />,
      precision: 0,
    },
    {
      id: "tasks",
      title: "算法任务",
      value: 356,
      prefix: <ExperimentOutlined className="icon-warning" />,
      precision: 0,
    },
    {
      id: "models",
      title: "AI模型",
      value: 89,
      prefix: <RobotOutlined className="icon-purple" />,
      precision: 0,
    },
  ];

  const coreCapabilities = [
    {
      id: "edge-management",
      icon: <NodeIndexOutlined className="icon-primary" style={{ fontSize: 32 }} />,
      title: "边缘算法管理",
      description:
        "支持AI Box边缘设备算法任务管理、能力查询和配置优化，实现边缘侧智能分析。",
      features: [
        { id: "schedule", text: "任务调度" },
        { id: "query", text: "能力查询" },
        { id: "config", text: "参数配置" },
        { id: "monitor", text: "性能监控" },
      ],
    },
    {
      id: "cloud-center",
      icon: <CloudServerOutlined className="icon-success" style={{ fontSize: 32 }} />,
      title: "云端算法中心",
      description: "集中管理云端算法库，支持算法版本控制、模型训练和部署分发。",
      features: [
        { id: "repository", text: "算法仓库" },
        { id: "version", text: "版本管理" },
        { id: "training", text: "模型训练" },
        { id: "deploy", text: "一键部署" },
      ],
    },
    {
      id: "model-management",
      icon: <RobotOutlined className="icon-warning" style={{ fontSize: 32 }} />,
      title: "AI模型管理",
      description: "统一管理深度学习模型，支持TensorFlow、PyTorch等主流框架。",
      features: [
        { id: "library", text: "模型库" },
        { id: "convert", text: "格式转换" },
        { id: "optimize", text: "性能优化" },
        { id: "quantize", text: "量化压缩" },
      ],
    },
    {
      id: "inference-acceleration",
      icon: <ThunderboltOutlined className="icon-purple" style={{ fontSize: 32 }} />,
      title: "算法推理加速",
      description: "基于GPU/NPU的高性能推理引擎，支持并发推理和资源调度。",
      features: [
        { id: "gpu", text: "GPU加速" },
        { id: "npu", text: "NPU支持" },
        { id: "batch", text: "批处理" },
        { id: "balance", text: "负载均衡" },
      ],
    },
    {
      id: "scheduling-center",
      icon: <ControlOutlined className="icon-magenta" style={{ fontSize: 32 }} />,
      title: "算法调度中心",
      description: "智能调度算法任务，支持优先级管理、资源分配和任务编排。",
      features: [
        { id: "queue", text: "任务队列" },
        { id: "priority", text: "优先级" },
        { id: "allocation", text: "资源分配" },
        { id: "orchestration", text: "依赖编排" },
      ],
    },
    {
      id: "cloud-edge-collaboration",
      icon: <DeploymentUnitOutlined className="icon-cyan" style={{ fontSize: 32 }} />,
      title: "云边协同",
      description: "云边端一体化架构，支持算法云端训练、边缘部署和端侧推理。",
      features: [
        { id: "cloud-train", text: "云端训练" },
        { id: "edge-deploy", text: "边缘部署" },
        { id: "device-inference", text: "端侧推理" },
        { id: "optimize", text: "协同优化" },
      ],
    },
  ];

  const algorithmLifecycle = [
    {
      id: "development",
      title: "算法开发",
      description: "模型设计、训练和验证",
      color: themeColors.primary,
    },
    {
      id: "deployment",
      title: "算法部署",
      description: "模型打包和分发部署",
      color: themeColors.success,
    },
    {
      id: "scheduling",
      title: "任务调度",
      description: "智能调度和资源分配",
      color: themeColors.warning,
    },
    {
      id: "execution",
      title: "推理执行",
      description: "实时推理和结果输出",
      color: themeColors.purple,
    },
    {
      id: "monitoring",
      title: "性能监控",
      description: "监控分析和性能优化",
      color: themeColors.magenta,
    },
    {
      id: "iteration",
      title: "模型迭代",
      description: "反馈优化和版本更新",
      color: themeColors.cyan,
    },
  ];

  // 渐变背景生成函数（使用主题色）
  const getGradientBackground = (index: number) => {
    const gradients = [
      `linear-gradient(135deg, ${themeColors.purple} 0%, ${token.colorPrimaryActive} 100%)`,
      `linear-gradient(135deg, ${themeColors.magenta} 0%, ${token.colorErrorActive} 100%)`,
      `linear-gradient(135deg, ${themeColors.cyan} 0%, ${token.colorInfoActive} 100%)`,
      `linear-gradient(135deg, ${token.colorSuccess} 0%, ${themeColors.cyan} 100%)`,
    ];
    return gradients[index % 4];
  };

  const applicationScenarios = [
    {
      id: "security",
      scenario: "智能安防",
      description: "人脸识别、行为分析、异常检测",
      algorithms: "目标检测、人脸识别、姿态估计",
      deployment: "边缘+云端",
    },
    {
      id: "quality-inspection",
      scenario: "工业质检",
      description: "缺陷检测、产品分类、尺寸测量",
      algorithms: "图像分类、目标检测、语义分割",
      deployment: "边缘侧",
    },
    {
      id: "traffic",
      scenario: "智慧交通",
      description: "车辆识别、违章检测、流量分析",
      algorithms: "目标跟踪、车牌识别、车型分类",
      deployment: "边缘+云端",
    },
    {
      id: "retail",
      scenario: "智慧零售",
      description: "客流统计、商品识别、行为分析",
      algorithms: "人数统计、商品检测、动作识别",
      deployment: "边缘+云端",
    },
  ];

  return (
    <PageContainer>
      {/* 功能完善中提示 */}
      <Alert
        message="功能完善中"
        description="云边端一体化AI算法管理平台正在开发完善中，当前已支持边缘算法管理功能，更多功能敬请期待。"
        type="info"
        icon={<InfoCircleOutlined />}
        showIcon
        className={styles.alertContainer}
      />

      <Card
        className={styles.heroCard}
        styles={{
          body: {
            backgroundImage: `linear-gradient(75deg, ${themeColors.purple} 0%, ${token.colorPrimaryActive} 100%)`,
            color: "white",
          },
        }}
      >
        <div
          className={styles.heroContent}
          style={{
            backgroundImage:
              "url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjc0IiBoZWlnaHQ9IjE4MCIgdmlld0JveD0iMCAwIDI3NCAxODAiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+CjxyZWN0IHg9IjIwMCIgeT0iNDAiIHdpZHRoPSI0MCIgaGVpZ2h0PSIyMCIgZmlsbD0icmdiYSgyNTUsMjU1LDI1NSwwLjEpIi8+CjxyZWN0IHg9IjE4MCIgeT0iODAiIHdpZHRoPSIzMCIgaGVpZ2h0PSIxNSIgZmlsbD0icmdiYSgyNTUsMjU1LDI1NSwwLjA3KSIvPgo8Y2lyY2xlIGN4PSIyMzAiIGN5PSIxMDAiIHI9IjE1IiBmaWxsPSJyZ2JhKDI1NSwyNTUsMjU1LDAuMDUpIi8+Cjwvc3ZnPg==')",
          }}
        >
          <div className={styles.heroTitle}>云边端一体化AI算法管理平台</div>
          <Paragraph className={styles.heroDescription}>
            打造云边端协同的AI算法管理平台，支持算法全生命周期管理、智能任务调度和高性能推理引擎。
            为智能安防、工业质检、智慧交通等场景提供完整的AI算法解决方案。
          </Paragraph>
        </div>
      </Card>

      {/* 统计概览 */}
      <Row gutter={[16, 16]} className={styles.statsRow}>
        {algorithmStats.map((stat) => (
          <Col xs={24} sm={12} md={6} key={stat.id}>
            <Card>
              <Statistic
                title={stat.title}
                value={stat.value}
                precision={stat.precision}
                prefix={stat.prefix}
                valueStyle={{ fontSize: "24px", fontWeight: "bold" }}
              />
            </Card>
          </Col>
        ))}
      </Row>

      {/* 核心能力 */}
      <Card title="核心技术能力" className={styles.capabilitiesCard}>
        <Row gutter={[24, 24]}>
          {coreCapabilities.map((capability) => (
            <Col xs={24} md={12} lg={8} key={capability.id}>
              <Card hoverable className={styles.capabilityCard}>
                <div className={styles.capabilityIconContainer}>
                  {capability.icon}
                </div>
                <Title level={5} className={styles.capabilityTitle}>
                  {capability.title}
                </Title>
                <Paragraph className={styles.capabilityDescription}>
                  {capability.description}
                </Paragraph>
                <div className={styles.capabilityFeatures}>
                  <Space wrap>
                    {capability.features.map((feature) => (
                      <Tag key={feature.id} className={styles.featureTag}>
                        {feature.text}
                      </Tag>
                    ))}
                  </Space>
                </div>
              </Card>
            </Col>
          ))}
        </Row>
      </Card>

      {/* 算法生命周期 */}
      <Card title="算法全生命周期管理" className={styles.lifecycleCard}>
        <Timeline
          items={algorithmLifecycle.map((stage, _index) => ({
            color: stage.color,
            children: (
              <div className={styles.lifecycleStage}>
                <Title
                  level={5}
                  className={styles.stageTitle}
                  style={{ color: stage.color }}
                >
                  {stage.title}
                </Title>
                <Paragraph className={styles.stageDescription}>
                  {stage.description}
                </Paragraph>
              </div>
            ),
          }))}
        />
      </Card>

      {/* 应用场景 */}
      <Card title="应用场景" className={styles.scenariosCard}>
        <Row gutter={[16, 16]}>
          {applicationScenarios.map((app, index) => (
            <Col xs={24} md={12} key={app.id}>
              <Card
                className={styles.scenarioCard}
                style={{ background: getGradientBackground(index) }}
              >
                <Title level={4} className={styles.scenarioTitle}>
                  {app.scenario}
                </Title>
                <Paragraph className={styles.scenarioDescription}>
                  {app.description}
                </Paragraph>
                <div className={styles.scenarioInfoRow}>
                  <Text strong className={styles.infoLabel}>
                    算法类型:{" "}
                  </Text>
                  <Text className={styles.infoValue}>{app.algorithms}</Text>
                </div>
                <div className={styles.scenarioInfoRow}>
                  <Text strong className={styles.infoLabel}>
                    部署方式:{" "}
                  </Text>
                  <Text className={styles.infoValue}>{app.deployment}</Text>
                </div>
              </Card>
            </Col>
          ))}
        </Row>
      </Card>

      {/* 技术架构优势 */}
      <Card title="技术架构优势">
        <Row gutter={[24, 24]}>
          <Col xs={24} md={12}>
            <Title level={5} className={styles.architectureSectionTitle}>
              🌐 云边端协同架构
            </Title>
            <Paragraph className={styles.architectureDescription}>
              采用云边端一体化架构，云端负责模型训练和管理，边缘侧实现本地推理，
              端侧提供轻量化部署，实现智能协同和资源优化。
            </Paragraph>

            <Title level={5} className={styles.architectureSectionTitle}>
              ⚡ 高性能推理引擎
            </Title>
            <Paragraph className={styles.architectureDescription}>
              支持GPU/NPU硬件加速，提供批处理、并发推理和模型优化能力。
              集成主流深度学习框架，支持多种模型格式转换和部署。
            </Paragraph>
          </Col>
          <Col xs={24} md={12}>
            <Title level={5} className={styles.architectureSectionTitle}>
              🔒 安全可信保障
            </Title>
            <Paragraph className={styles.architectureDescription}>
              提供算法加密、权限管理和审计日志功能，保障算法和数据安全。
              支持隐私计算和联邦学习，满足数据安全和隐私保护要求。
            </Paragraph>

            <Title level={5} className={styles.architectureSectionTitle}>
              📊 智能监控运维
            </Title>
            <Paragraph className={styles.architectureDescription}>
              实时监控算法性能和资源使用情况，提供智能调优建议。
              全链路追踪和可视化分析，支持算法质量评估和持续优化。
            </Paragraph>
          </Col>
        </Row>
      </Card>
    </PageContainer>
  );
};

export default AlgorithmManagement;
