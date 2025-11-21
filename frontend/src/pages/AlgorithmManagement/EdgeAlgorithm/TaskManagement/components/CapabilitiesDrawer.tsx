import type {
  AIBoxAbility,
  AIBoxCapabilities,
  AIBoxParameter,
} from "@/services/iot";
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExperimentOutlined,
  SettingOutlined,
  WarningOutlined,
} from "@ant-design/icons";
import {
  Alert,
  Badge,
  Collapse,
  Descriptions,
  Drawer,
  Empty,
  Space,
  Spin,
  Tag,
  Typography,
} from "antd";
import React from "react";
import styles from "./CapabilitiesDrawer.less";

const { Panel } = Collapse;
const { Text, Title } = Typography;

interface CapabilitiesDrawerProps {
  open: boolean; // 抽屉是否打开
  deviceId?: string; // 设备ID
  capabilities: AIBoxCapabilities | null; // 算法能力数据
  loading?: boolean; // 加载状态
  onClose: () => void; // 关闭回调
}

/**
 * AI Box 算法能力抽屉
 *
 * 功能：
 * - 展示AI Box设备支持的所有算法能力详情
 * - 使用折叠面板展示每个算法
 * - 展示算法参数、属性、报警规则等详细信息
 * - 显示算法授权状态
 */
const CapabilitiesDrawer: React.FC<CapabilitiesDrawerProps> = ({
  open,
  deviceId: _deviceId,
  capabilities,
  loading = false,
  onClose,
}) => {
  /**
   * 渲染算法参数
   */
  const renderParameter = (param: AIBoxParameter) => {
    return (
      <Descriptions
        key={param.key}
        size="small"
        column={2}
        bordered
        className={styles.parameterCard}
      >
        <Descriptions.Item label="参数名称" span={2}>
          <Space>
            <Text strong>{param.name}</Text>
            {param.required && <Tag color="red">必填</Tag>}
          </Space>
        </Descriptions.Item>
        <Descriptions.Item label="参数标识">{param.key}</Descriptions.Item>
        <Descriptions.Item label="数据类型">
          <Tag color="blue">{param.class}</Tag>
        </Descriptions.Item>
        {param.min !== undefined && (
          <Descriptions.Item label="最小值">{param.min}</Descriptions.Item>
        )}
        {param.max !== undefined && (
          <Descriptions.Item label="最大值">{param.max}</Descriptions.Item>
        )}
        <Descriptions.Item label="默认值">{param.default}</Descriptions.Item>
        <Descriptions.Item label="当前值">
          {param.value || "-"}
        </Descriptions.Item>
        {param.options && param.options.length > 0 && (
          <Descriptions.Item label="可选项" span={2}>
            <Space wrap>
              {param.options.map((opt) => (
                <Tag key={opt.key} color={opt.enable ? "green" : "default"}>
                  {opt.name} ({opt.value})
                </Tag>
              ))}
            </Space>
          </Descriptions.Item>
        )}
      </Descriptions>
    );
  };

  /**
   * 渲染单个算法能力
   */
  const renderAbility = (ability: AIBoxAbility, index: number) => {
    const panelHeader = (
      <div className={styles.panelHeader}>
        {/* 左侧：算法名称 */}
        <div className={styles.algorithmName}>
          <ExperimentOutlined />
          <Text
            strong
            className={styles.algorithmNameText}
            ellipsis={{ tooltip: ability.name }}
          >
            {ability.name}
          </Text>
        </div>

        {/* 中间：算法类型标签（固定位置） */}
        <div className={styles.algorithmType}>
          {ability.sub ? (
            <Tag color="purple">子算法</Tag>
          ) : (
            <Tag color="blue">主算法</Tag>
          )}
        </div>

        {/* 右侧：授权状态和编号（固定宽度） */}
        <div className={styles.statusArea}>
          <div className={styles.statusBadge}>
            {ability.permitted ? (
              <Badge status="success" text="已授权" />
            ) : (
              <Badge status="error" text="未授权" />
            )}
          </div>
          <div className={styles.codeLabel}>
            <Text type="secondary" className={styles.codeLabelText}>
              主编号: <Text code>{ability.code}</Text>
            </Text>
          </div>
          <div className={styles.codeLabel}>
            <Text type="secondary" className={styles.codeLabelText}>
              子编号: <Text code>{ability.item}</Text>
            </Text>
          </div>
        </div>
      </div>
    );

    return (
      <Panel header={panelHeader} key={index}>
        {/* 算法描述 */}
        {ability.desc && (
          <Alert
            message={ability.desc}
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}

        {/* 算法属性 */}
        <Title level={5}>
          <SettingOutlined /> 算法属性
        </Title>
        <Descriptions
          size="small"
          column={1}
          bordered
          style={{ marginBottom: 16 }}
        >
          <Descriptions.Item label="辅助线配置">
            {ability.attribute.lineRequired ? (
              <Space>
                <CheckCircleOutlined className={styles.onlineIcon} />
                <Text>必须配置</Text>
                {ability.attribute.lineDesc && (
                  <Text type="secondary">（{ability.attribute.lineDesc}）</Text>
                )}
              </Space>
            ) : (
              <Space>
                <CloseCircleOutlined className={styles.offlineIcon} />
                <Text type="secondary">非必须</Text>
              </Space>
            )}
          </Descriptions.Item>
          <Descriptions.Item label="区域配置">
            {ability.attribute.zoneRequired ? (
              <Space>
                <CheckCircleOutlined className={styles.onlineIcon} />
                <Text>必须配置</Text>
                {ability.attribute.zoneDesc && (
                  <Text type="secondary">（{ability.attribute.zoneDesc}）</Text>
                )}
              </Space>
            ) : (
              <Space>
                <CloseCircleOutlined className={styles.offlineIcon} />
                <Text type="secondary">非必须</Text>
              </Space>
            )}
          </Descriptions.Item>
        </Descriptions>

        {/* 算法参数 */}
        {ability.parameters && ability.parameters.length > 0 && (
          <>
            <Title level={5}>
              <SettingOutlined /> 算法参数 ({ability.parameters.length})
            </Title>
            {ability.parameters.map(renderParameter)}
          </>
        )}

        {/* 报警规则 */}
        {ability.policy && ability.policy.length > 0 && (
          <>
            <Title level={5} className={styles.titleMargin}>
              <WarningOutlined /> 报警规则
            </Title>
            <Space wrap>
              {ability.policy.map((p) => (
                <Tag
                  key={`${p.name}-${p.property}`}
                  color="orange"
                  icon={<WarningOutlined />}
                >
                  {p.name} ({p.property})
                </Tag>
              ))}
            </Space>
          </>
        )}
      </Panel>
    );
  };

  /**
   * 渲染抽屉内容
   */
  const renderContent = () => {
    if (loading) {
      return (
        <div className={styles.loadingContainer}>
          <Spin size="large" tip="加载算法能力中...">
            <div />
          </Spin>
        </div>
      );
    }

    if (!capabilities) {
      return (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="暂无算法能力数据"
        />
      );
    }

    return (
      <>
        {/* 设备信息 */}
        <Descriptions
          title="设备信息"
          size="default"
          column={2}
          bordered
          className={styles.parameterCard}
        >
          <Descriptions.Item label="设备ID" span={2}>
            <Text code copyable>
              {capabilities.board_id}
            </Text>
          </Descriptions.Item>
          <Descriptions.Item label="算法数量" span={2}>
            <Space>
              <Text strong className={styles.algorithmCount}>
                {capabilities.abilities?.length || 0}
              </Text>
              <Text type="secondary">个算法</Text>
            </Space>
          </Descriptions.Item>
        </Descriptions>

        {/* 算法能力列表 */}
        {capabilities.abilities && capabilities.abilities.length > 0 ? (
          <Collapse
            accordion
            defaultActiveKey={[0]}
            style={{ marginBottom: 16 }}
          >
            {capabilities.abilities.map(renderAbility)}
          </Collapse>
        ) : (
          <Empty description="该设备暂无可用算法" />
        )}
      </>
    );
  };

  return (
    <Drawer
      title="AI Box 算法能力详情"
      placement="right"
      size="large"
      open={open}
      onClose={onClose}
      destroyOnHidden
    >
      {renderContent()}
    </Drawer>
  );
};

export default CapabilitiesDrawer;
