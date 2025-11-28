import {
  ApartmentOutlined,
  ApiOutlined,
  BranchesOutlined,
  CalculatorOutlined,
  ClockCircleOutlined,
  ControlOutlined,
  DatabaseOutlined,
  FilterOutlined,
  FunctionOutlined,
  InfoCircleOutlined,
  NodeIndexOutlined,
  SyncOutlined,
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
  Steps,
  Table,
  Tag,
  Typography,
  theme,
} from "antd";
import React from "react";

const { Title, Paragraph, Text } = Typography;

/**
 * 逻辑引擎页面
 * 展示基于LynxGraph的可视化规则引擎核心功能和技术架构
 */
const LogicEngine: React.FC = () => {
  const { token } = theme.useToken();

  const engineStats = [
    {
      title: "活跃逻辑图",
      value: 248,
      prefix: <ApartmentOutlined style={{ color: token.colorPrimary }} />,
      precision: 0,
    },
    {
      title: "运行节点",
      value: 3567,
      prefix: <NodeIndexOutlined style={{ color: token.colorSuccess }} />,
      precision: 0,
    },
    {
      title: "处理事件",
      value: 156789,
      prefix: <ThunderboltOutlined style={{ color: token.colorWarning }} />,
      precision: 0,
    },
    {
      title: "执行路径",
      value: 892,
      prefix: <BranchesOutlined style={{ color: token.colorInfo }} />,
      precision: 0,
    },
  ];

  const logicBlocks = [
    {
      type: "状态机",
      icon: <ControlOutlined style={{ fontSize: 20, color: token.colorPrimary }} />,
      description: "对事件或对象进行状态迁移建模",
      examples: ["目标行为过程建模", "设备异常状态跟踪"],
      color: token.colorPrimary,
    },
    {
      type: "表达式计算器",
      icon: <CalculatorOutlined style={{ fontSize: 20, color: token.colorSuccess }} />,
      description: "通过DSL表达式判断输入是否满足条件",
      examples: ['value > 70 && type == "smoke"', "温度超阈值判断"],
      color: token.colorSuccess,
    },
    {
      type: "聚合器",
      icon: <FunctionOutlined style={{ fontSize: 20, color: token.colorWarning }} />,
      description: "多事件组合触发判断，支持时间窗",
      examples: ["红外+烟感5秒内同时告警", "多传感器数据融合"],
      color: token.colorWarning,
    },
    {
      type: "条件网关",
      icon: <FilterOutlined style={{ fontSize: 20, color: token.purple }} />,
      description: "多分支IF/ELSE判断",
      examples: ["满足规则A走路径X", "业务规则分支路由"],
      color: token.purple,
    },
    {
      type: "计数器",
      icon: <DatabaseOutlined style={{ fontSize: 20, color: token.magenta }} />,
      description: "支持滑动时间窗口内计数",
      examples: ["10秒内告警次数 >= 3", "频次统计分析"],
      color: token.magenta,
    },
    {
      type: "定时器",
      icon: <ClockCircleOutlined style={{ fontSize: 20, color: token.cyan }} />,
      description: "支持延迟、持续时间判断",
      examples: ["滞留超30s触发", "延迟10s再执行"],
      color: token.cyan,
    },
  ];

  const engineFeatures = [
    {
      icon: <NodeIndexOutlined style={{ fontSize: 32, color: token.colorPrimary }} />,
      title: "信息原子设计",
      description:
        '不以"事件"为最小单元，而是采用信息原子模式，支持检测框、轨迹、区域坐标等多种数据类型。',
    },
    {
      icon: <ThunderboltOutlined style={{ fontSize: 32, color: token.colorSuccess }} />,
      title: "路径驱动执行",
      description:
        '采用"路径即逻辑"的流式推理模式，每个路径可独立触发、并行执行，避免整体图绑定。',
    },
    {
      icon: <DatabaseOutlined style={{ fontSize: 32, color: token.colorWarning }} />,
      title: "数据共享机制",
      description:
        "中心事件缓冲区+图级上下文引用，避免数据冗余，支持多逻辑图并发处理。",
    },
    {
      icon: <BranchesOutlined style={{ fontSize: 32, color: token.purple }} />,
      title: "智能事件调度",
      description:
        "内置事件调度器支持多图订阅索引，同一事件可触发多个逻辑图执行环境。",
    },
    {
      icon: <ControlOutlined style={{ fontSize: 32, color: token.magenta }} />,
      title: "状态机制管理",
      description:
        "支持跨事件关联与幂等控制，状态可配置TTL，构建时间窗口逻辑。",
    },
    {
      icon: <SyncOutlined style={{ fontSize: 32, color: token.cyan }} />,
      title: "驻留协程优化",
      description:
        "每条路径对应常驻执行单元，使用独立通道接收事件，避免频繁创建销毁协程。",
    },
  ];

  const architectureSteps = [
    {
      title: "事件输入",
      description: "外部数据进入DataCore统一缓冲",
      icon: <ApiOutlined />,
    },
    {
      title: "事件调度",
      description: "Dispatcher根据订阅索引分发事件",
      icon: <BranchesOutlined />,
    },
    {
      title: "路径匹配",
      description: "入口节点条件匹配，确定执行路径",
      icon: <FilterOutlined />,
    },
    {
      title: "逻辑执行",
      description: "积木节点串行执行，状态联结处理",
      icon: <NodeIndexOutlined />,
    },
    {
      title: "结果输出",
      description: "执行结果输出或触发后续动作",
      icon: <ThunderboltOutlined />,
    },
  ];

  const performanceMetrics = [
    { metric: "事件处理延迟", value: "< 10ms", benchmark: "毫秒级响应" },
    { metric: "并发处理能力", value: "100万 TPS", benchmark: "高并发支持" },
    { metric: "逻辑图规模", value: "1000+ 节点", benchmark: "复杂场景支持" },
    { metric: "状态存储", value: "TB级", benchmark: "海量状态管理" },
  ];

  const applicationScenarios = [
    {
      scenario: "智慧安防",
      description: "人员行为分析、异常检测、联动告警",
      logicExamples: ["人员徘徊检测", "区域入侵告警", "人群聚集预警"],
      complexity: "中等",
    },
    {
      scenario: "工业监控",
      description: "设备状态监控、故障预测、自动控制",
      logicExamples: ["设备振动异常", "温度超限处理", "生产线协调"],
      complexity: "复杂",
    },
    {
      scenario: "智慧交通",
      description: "交通流量分析、违规检测、信号控制",
      logicExamples: ["车辆违停检测", "交通拥堵判断", "信号灯联控"],
      complexity: "高",
    },
    {
      scenario: "环境监测",
      description: "环境数据采集、异常预警、自动调节",
      logicExamples: ["空气质量预警", "噪音超标处理", "能耗优化控制"],
      complexity: "简单",
    },
  ];

  return (
    <PageContainer>
      {/* 功能完善中提示 */}
      <Alert
        message="功能完善中"
        description="LynxGraph逻辑引擎正在开发完善中，敬请期待更多强大功能的上线。"
        type="info"
        icon={<InfoCircleOutlined />}
        showIcon
        style={{ marginBottom: 24 }}
      />

      <Card
        style={{
          borderRadius: 8,
          marginBottom: 24,
        }}
        styles={{
          body: {
            backgroundImage: `linear-gradient(75deg, ${token.colorPrimaryBg} 0%, ${token.colorInfoBg} 100%)`,
            color: token.colorText,
          },
        }}
      >
        <div
          style={{
            backgroundPosition: "100% -30%",
            backgroundRepeat: "no-repeat",
            backgroundSize: "274px auto",
            backgroundImage:
              "url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjc0IiBoZWlnaHQ9IjE4MCIgdmlld0JveD0iMCAwIDI3NCAxODAiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+CjxjaXJjbGUgY3g9IjIyMCIgY3k9IjQwIiByPSIyMCIgZmlsbD0icmdiYSgwLDAsMCwwLjEpIi8+CjxyZWN0IHg9IjE5MCIgeT0iNzAiIHdpZHRoPSI2MCIgaGVpZ2h0PSIzMCIgcng9IjUiIGZpbGw9InJnYmEoMCwwLDAsMC4wNykiLz4KPGNpcmNsZSBjeD0iMjQwIiBjeT0iMTIwIiByPSIyNSIgZmlsbD0icmdiYSgwLDAsMCwwLjA1KSIvPgo8L3N2Zz4=')",
          }}
        >
          <div
            style={{ fontSize: "28px", fontWeight: "bold", marginBottom: 16 }}
          >
            LynxGraph 逻辑引擎
          </div>
          <Paragraph
            style={{
              fontSize: "16px",
              color: token.colorTextSecondary,
              marginBottom: 32,
            }}
          >
            基于信息原子和路径驱动的新一代逻辑引擎，通过积木化编程实现复杂业务逻辑。
            支持强并发、异步容忍、状态编排，为智能系统提供强大的规则引擎能力。
          </Paragraph>
        </div>
      </Card>

      {/* 统计概览 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        {engineStats.map((stat) => (
          <Col xs={24} sm={12} md={6} key={stat.title}>
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

      {/* 逻辑积木原语 */}
      <Card title="逻辑积木原语库" style={{ marginBottom: 24 }}>
        <Paragraph
          style={{ color: token.colorTextSecondary, marginBottom: 16 }}
        >
          LynxGraph提供丰富的逻辑积木原语，每个积木都是独立的逻辑处理单元，通过组合可以构建复杂的业务逻辑。
        </Paragraph>
        <Row gutter={[16, 16]}>
          {logicBlocks.map((block) => (
            <Col xs={24} md={12} lg={8} key={block.type}>
              <Card
                size="small"
                hoverable
                style={{
                  borderLeft: `4px solid ${block.color}`,
                  height: "100%",
                }}
              >
                <div
                  style={{
                    display: "flex",
                    alignItems: "center",
                    marginBottom: 8,
                  }}
                >
                  {block.icon}
                  <Text strong style={{ marginLeft: 8, color: block.color }}>
                    {block.type}
                  </Text>
                </div>
                <Paragraph
                  style={{
                    fontSize: "13px",
                    color: token.colorTextSecondary,
                    marginBottom: 12,
                  }}
                >
                  {block.description}
                </Paragraph>
                <div>
                  {block.examples.map((example) => (
                    <Tag
                      key={example}
                      style={{ fontSize: "11px", marginBottom: 4 }}
                    >
                      {example}
                    </Tag>
                  ))}
                </div>
              </Card>
            </Col>
          ))}
        </Row>
      </Card>

      {/* 核心特性 */}
      <Card title="核心技术特性" style={{ marginBottom: 24 }}>
        <Row gutter={[24, 24]}>
          {engineFeatures.map((feature) => (
            <Col xs={24} md={12} lg={8} key={feature.title}>
              <Card
                hoverable
                style={{
                  height: "100%",
                  borderRadius: 8,
                }}
                styles={{
                  body: {
                    padding: "20px",
                    display: "flex",
                    flexDirection: "column",
                    height: "100%",
                  },
                }}
              >
                <div style={{ textAlign: "center", marginBottom: 16 }}>
                  {feature.icon}
                </div>
                <Title
                  level={5}
                  style={{ textAlign: "center", marginBottom: 12 }}
                >
                  {feature.title}
                </Title>
                <Paragraph
                  style={{
                    color: token.colorTextSecondary,
                    textAlign: "center",
                    flex: 1,
                    marginBottom: 0,
                  }}
                >
                  {feature.description}
                </Paragraph>
              </Card>
            </Col>
          ))}
        </Row>
      </Card>

      {/* 执行架构流程 */}
      <Card title="执行架构流程" style={{ marginBottom: 24 }}>
        <Steps
          current={-1}
          items={architectureSteps.map((step, _index) => ({
            title: step.title,
            description: step.description,
            icon: step.icon,
          }))}
        />
        <div
          style={{
            marginTop: 24,
            padding: "16px",
            backgroundColor: token.colorFillAlter,
            borderRadius: 8,
          }}
        >
          <Text strong>设计哲学: </Text>
          <Text style={{ color: token.colorTextSecondary }}>
            路径驱动，状态联结 -
            采用"路径即逻辑"的流式推理模式，避免传统流程引擎中对整体图的沉重绑定。
            每个路径可独立触发、并行执行，通过状态机制实现跨路径的数据联结。
          </Text>
        </div>
      </Card>

      {/* 性能指标 */}
      <Card title="性能指标" style={{ marginBottom: 24 }}>
        <Row gutter={[16, 16]}>
          {performanceMetrics.map((metric, index) => (
            <Col xs={24} sm={12} md={6} key={metric.metric}>
              <Card
                style={{
                  textAlign: "center",
                  background: `linear-gradient(135deg, ${
                    index % 4 === 0
                      ? token.colorPrimary
                      : index % 4 === 1
                      ? token.magenta
                      : index % 4 === 2
                      ? token.colorInfo
                      : token.colorSuccess
                  } 0%, ${
                    index % 4 === 0
                      ? token.purple
                      : index % 4 === 1
                      ? token.magenta
                      : index % 4 === 2
                      ? token.cyan
                      : token.cyan
                  } 100%)`,
                  color: "white",
                  borderRadius: 8,
                }}
              >
                <Title level={4} style={{ color: "white", marginBottom: 8 }}>
                  {metric.value}
                </Title>
                <div style={{ fontSize: "14px", marginBottom: 4 }}>
                  {metric.metric}
                </div>
                <div style={{ fontSize: "12px", color: token.colorTextTertiary }}>
                  {metric.benchmark}
                </div>
              </Card>
            </Col>
          ))}
        </Row>
      </Card>

      {/* 应用场景 */}
      <Card title="典型应用场景">
        <Table
          dataSource={applicationScenarios}
          pagination={false}
          size="middle"
          columns={[
            {
              title: "应用场景",
              dataIndex: "scenario",
              key: "scenario",
              render: (text) => <Text strong>{text}</Text>,
            },
            {
              title: "场景描述",
              dataIndex: "description",
              key: "description",
              render: (text) => (
                <Text style={{ color: token.colorTextSecondary }}>{text}</Text>
              ),
            },
            {
              title: "逻辑示例",
              dataIndex: "logicExamples",
              key: "logicExamples",
              render: (examples) => (
                <Space wrap>
                  {examples.map((example: string) => (
                    <Tag key={example}>{example}</Tag>
                  ))}
                </Space>
              ),
            },
            {
              title: "复杂度",
              dataIndex: "complexity",
              key: "complexity",
              render: (complexity) => {
                const color =
                  complexity === "简单"
                    ? "green"
                    : complexity === "中等"
                    ? "orange"
                    : complexity === "复杂"
                    ? "red"
                    : "purple";
                return <Tag color={color}>{complexity}</Tag>;
              },
            },
          ]}
        />

        <div
          style={{
            marginTop: 24,
            padding: "16px",
            backgroundColor: token.colorFillAlter,
            borderRadius: 8,
          }}
        >
          <Row gutter={[24, 24]}>
            <Col xs={24} md={12}>
              <Title level={5}>💡 核心优势</Title>
              <Paragraph style={{ color: token.colorTextSecondary }}>
                • <Text strong>强并发</Text>: 多个事件可同时触发不同路径
                <br />• <Text strong>异步容忍</Text>: 事件到达无先后限制
                <br />• <Text strong>状态编排</Text>:
                通过状态实现时间/空间/行为关联
                <br />• <Text strong>热插拔</Text>: 调度层支持随时新增/更新图
              </Paragraph>
            </Col>
            <Col xs={24} md={12}>
              <Title level={5}>🔧 技术特色</Title>
              <Paragraph style={{ color: token.colorTextSecondary }}>
                • <Text strong>信息原子</Text>: 支持轨迹、检测框、原始信号等形式
                <br />• <Text strong>驻留协程</Text>:
                PathWorker常驻执行单元优化性能
                <br />• <Text strong>幂等控制</Text>: 状态标记控制避免重复触发
                <br />• <Text strong>可视化编辑</Text>: 图形化逻辑编排界面
              </Paragraph>
            </Col>
          </Row>
        </div>
      </Card>
    </PageContainer>
  );
};

export default LogicEngine;
