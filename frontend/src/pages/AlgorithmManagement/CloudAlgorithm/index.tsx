import {
  CloudServerOutlined,
  CodeOutlined,
  DatabaseOutlined,
  DeploymentUnitOutlined,
  ExperimentOutlined,
  InfoCircleOutlined,
  RocketOutlined,
  SyncOutlined,
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
 * 云端算法管理介绍页面
 * 展示云端算法中心的模型训练、管理、部署等功能
 */
const CloudAlgorithm: React.FC = () => {
  const { token } = theme.useToken();

  const features = [
    {
      id: "training",
      title: "模型训练",
      description: "强大的云端算力支持大规模模型训练",
      items: [
        { id: "distributed", text: "分布式训练" },
        { id: "gpu", text: "GPU/TPU加速" },
        { id: "auto-tune", text: "超参数自动调优" },
        { id: "monitor", text: "训练可视化监控" },
      ],
      color: token.colorPrimary,
      icon: (
        <ExperimentOutlined
          style={{ fontSize: 20, color: token.colorPrimary }}
        />
      ),
    },
    {
      id: "repository",
      title: "算法仓库",
      description: "集中管理和版本控制算法模型",
      items: [
        { id: "version", text: "模型版本管理" },
        { id: "asset", text: "算法资产库" },
        { id: "permission", text: "权限控制" },
        { id: "search", text: "快速检索" },
      ],
      color: token.colorSuccess,
      icon: (
        <DatabaseOutlined style={{ fontSize: 20, color: token.colorSuccess }} />
      ),
    },
    {
      id: "deployment",
      title: "智能部署",
      description: "一键部署到边缘或端侧设备",
      items: [
        { id: "package", text: "自动打包" },
        { id: "distribute", text: "批量分发" },
        { id: "canary", text: "灰度发布" },
        { id: "rollback", text: "回滚机制" },
      ],
      color: token.colorWarning,
      icon: (
        <DeploymentUnitOutlined
          style={{ fontSize: 20, color: token.colorWarning }}
        />
      ),
    },
  ];

  const workflowSteps = [
    {
      title: "数据准备",
      description: "上传训练数据集到云端存储",
      icon: <DatabaseOutlined />,
    },
    {
      title: "模型训练",
      description: "GPU/TPU集群分布式训练",
      icon: <ExperimentOutlined />,
    },
    {
      title: "模型验证",
      description: "验证集评估模型性能",
      icon: <CodeOutlined />,
    },
    {
      title: "版本管理",
      description: "模型版本化存储到仓库",
      icon: <SyncOutlined />,
    },
    {
      title: "一键部署",
      description: "部署到云端或边缘设备",
      icon: <DeploymentUnitOutlined />,
    },
  ];

  const performanceMetrics = [
    {
      id: "acceleration",
      metric: "训练加速",
      value: "10倍",
      benchmark: "GPU集群加速",
    },
    {
      id: "concurrent",
      metric: "并发训练",
      value: "100+",
      benchmark: "任务数",
    },
    {
      id: "storage",
      metric: "模型存储",
      value: "无限制",
      benchmark: "弹性扩展",
    },
    {
      id: "efficiency",
      metric: "部署效率",
      value: "< 5分钟",
      benchmark: "快速上线",
    },
  ];

  return (
    <PageContainer>
      <Alert
        message="功能完善中"
        description="云端算法中心正在开发完善中，敬请期待强大的云端训练和部署能力。"
        type="info"
        icon={<InfoCircleOutlined />}
        showIcon
        style={{ marginBottom: 24 }}
      />

      {/* 顶部Banner */}
      <Card className={styles.banner}>
        <div className={styles.bannerContent}>
          <div className={styles.bannerTitle}>云端算法中心</div>
          <Paragraph className={styles.bannerDescription}>
            提供强大的云端算法训练、管理和部署能力，支持分布式训练、模型版本管理、智能部署等全流程功能。
            基于GPU/TPU集群，为AI算法提供海量算力支持。
          </Paragraph>
        </div>
      </Card>

      {/* 统计概览 */}
      <Flex gap={16} wrap="wrap" style={{ marginBottom: 24 }}>
        <div style={{ flex: "1 1 calc(25% - 12px)", minWidth: 200 }}>
          <Card>
            <Statistic
              title="算法模型"
              value={128}
              prefix={
                <CloudServerOutlined style={{ color: token.colorPrimary }} />
              }
              suffix="个"
              valueStyle={{ fontSize: "24px", fontWeight: "bold" }}
            />
          </Card>
        </div>
        <div style={{ flex: "1 1 calc(25% - 12px)", minWidth: 200 }}>
          <Card>
            <Statistic
              title="训练任务"
              value={356}
              prefix={<RocketOutlined style={{ color: token.colorSuccess }} />}
              suffix="次"
              valueStyle={{ fontSize: "24px", fontWeight: "bold" }}
            />
          </Card>
        </div>
        <div style={{ flex: "1 1 calc(25% - 12px)", minWidth: 200 }}>
          <Card>
            <Statistic
              title="部署次数"
              value={892}
              prefix={
                <DeploymentUnitOutlined style={{ color: token.colorWarning }} />
              }
              suffix="次"
              valueStyle={{ fontSize: "24px", fontWeight: "bold" }}
            />
          </Card>
        </div>
        <div style={{ flex: "1 1 calc(25% - 12px)", minWidth: 200 }}>
          <Card>
            <Statistic
              title="训练加速"
              value={10}
              prefix={
                <DatabaseOutlined style={{ color: token.colorInfoText }} />
              }
              suffix="倍"
              valueStyle={{ fontSize: "24px", fontWeight: "bold" }}
            />
          </Card>
        </div>
      </Flex>

      {/* 核心功能 */}
      <Card title="核心功能" style={{ marginBottom: 24 }}>
        <Paragraph
          style={{ color: token.colorTextSecondary, marginBottom: 16 }}
        >
          云端算法中心提供完整的AI算法全生命周期管理，从数据准备到模型部署，一站式解决方案。
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

      {/* 训练部署流程 */}
      <Card title="训练部署流程" style={{ marginBottom: 24 }}>
        <Steps
          current={-1}
          items={workflowSteps.map((step) => ({
            title: step.title,
            description: step.description,
            icon: step.icon,
          }))}
        />
        <div className={styles.workflowTip}>
          <Text strong>工作流程: </Text>
          <Text style={{ color: token.colorTextSecondary }}>
            云端算法中心采用标准化的训练部署流程，从数据准备到模型上线全程自动化。
            支持分布式训练加速、自动化版本管理和一键部署，大幅提升AI算法的开发效率。
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

      {/* 核心优势 */}
      <Card title="核心优势">
        <div className={styles.advantageSection}>
          <Flex gap={24} wrap="wrap">
            <div style={{ flex: "1 1 calc(50% - 12px)", minWidth: 300 }}>
              <Title level={5}>💡 技术优势</Title>
              <Paragraph style={{ color: token.colorTextSecondary }}>
                • <Text strong>海量数据</Text>: 云端存储和处理PB级训练数据
                <br />• <Text strong>强大算力</Text>: 弹性扩展的GPU/TPU集群
                <br />• <Text strong>分布式训练</Text>: 支持大规模模型并行训练
                <br />• <Text strong>自动调优</Text>: 智能超参数搜索和优化
              </Paragraph>
            </div>
            <div style={{ flex: "1 1 calc(50% - 12px)", minWidth: 300 }}>
              <Title level={5}>🚀 业务价值</Title>
              <Paragraph style={{ color: token.colorTextSecondary }}>
                • <Text strong>灵活部署</Text>: 支持公有云、私有云、混合云
                <br />• <Text strong>协同开发</Text>:
                团队协作、实验管理、结果共享
                <br />• <Text strong>快速迭代</Text>: 版本管理和一键回滚
                <br />• <Text strong>降本增效</Text>: 按需付费，降低硬件成本
              </Paragraph>
            </div>
          </Flex>
        </div>
      </Card>
    </PageContainer>
  );
};

export default CloudAlgorithm;
