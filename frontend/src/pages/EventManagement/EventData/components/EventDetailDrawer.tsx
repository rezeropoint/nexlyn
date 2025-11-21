import type {
  ColumnInfo,
  EventDetailResponse,
  FlowNode,
} from "@/pages/EventManagement/types";
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  SyncOutlined,
  UserOutlined,
} from "@ant-design/icons";
import {
  Card,
  Descriptions,
  Drawer,
  Empty,
  Space,
  Spin,
  Tag,
  Timeline,
  Typography,
} from "antd";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import React from "react";
import "./EventDataTable.less";

// 启用 UTC 插件
dayjs.extend(utc);

const { Text, Title } = Typography;

/**
 * 渲染字段值（支持多种类型）
 */
const renderValue = (value: any) => {
  if (value === null || value === undefined) {
    return <Text type="secondary">-</Text>;
  }
  if (typeof value === "boolean") {
    return (
      <Tag color={value ? "success" : "default"}>{value ? "是" : "否"}</Tag>
    );
  }
  if (typeof value === "object") {
    return <Text code>{JSON.stringify(value)}</Text>;
  }
  return String(value);
};

/**
 * 获取流程状态的显示配置
 * 状态映射来自后端 pkg/skylarkq/internal/query/internal.go
 */
const getStatusConfig = (status: string) => {
  const statusMap: Record<string, { color: string; text: string; icon: any }> =
    {
      stashed: {
        color: "default",
        text: "编写中",
        icon: <ClockCircleOutlined />,
      },
      processing: {
        color: "processing",
        text: "处理中",
        icon: <SyncOutlined spin />,
      },
      approved: {
        color: "success",
        text: "已通过",
        icon: <CheckCircleOutlined />,
      },
      refused: {
        color: "warning",
        text: "已回退",
        icon: <CloseCircleOutlined />,
      },
      transferred: { color: "cyan", text: "已转交", icon: <SyncOutlined /> },
      skipped: {
        color: "default",
        text: "已跳过",
        icon: <ClockCircleOutlined />,
      },
      cancelled: {
        color: "error",
        text: "已撤销",
        icon: <CloseCircleOutlined />,
      },
      receding: { color: "warning", text: "回退中", icon: <SyncOutlined /> },
      suspended: {
        color: "default",
        text: "已暂停",
        icon: <ClockCircleOutlined />,
      },
      finished: {
        color: "success",
        text: "已完成",
        icon: <CheckCircleOutlined />,
      },
      aborted: {
        color: "error",
        text: "已终止",
        icon: <CloseCircleOutlined />,
      },
    };
  return (
    statusMap[status] || {
      color: "default",
      text: status,
      icon: <ClockCircleOutlined />,
    }
  );
};

interface EventDetailDrawerProps {
  open: boolean;
  onClose: () => void;
  detail: EventDetailResponse["data"] | null;
  loading: boolean;
  columns?: ColumnInfo[]; // 表格列配置，用于显示字段名（后端已过滤数据）
}

/**
 * 计算抽屉的动态宽度
 * 根据业务数据的字段数量和内容长度智能调整
 * 策略：默认使用舒适宽度，只在内容过长时才动态扩宽
 */
const calculateDrawerWidth = (
  detail: EventDetailResponse["data"] | null
): number | string => {
  const DEFAULT_WIDTH = 600; // 舒适的默认宽度
  const MIN_WIDTH = 600; // 最小宽度
  const MAX_WIDTH = 1400; // 最大宽度
  const EXPAND_THRESHOLD = 50; // 内容长度超过这个值才考虑扩宽

  if (!detail || !detail.latestBusinessData) {
    return DEFAULT_WIDTH;
  }

  const businessData = detail.latestBusinessData;

  // 计算最长字段值的长度
  let maxValueLength = 0;

  Object.values(businessData).forEach((value) => {
    const valueStr = value !== null && value !== undefined ? String(value) : "";
    maxValueLength = Math.max(maxValueLength, valueStr.length);
  });

  // 如果内容不长，直接返回默认宽度
  if (maxValueLength <= EXPAND_THRESHOLD) {
    return DEFAULT_WIDTH;
  }

  // 内容过长时，根据长度动态扩宽
  // 标签固定180px + 左右padding约80px + 内容区域
  // 中文字符约14px，英文约8px，这里取平均11px
  const estimatedContentWidth = maxValueLength * 11;
  const estimatedTotalWidth = 260 + estimatedContentWidth;

  // 限制在最小和最大宽度之间
  const finalWidth = Math.max(
    MIN_WIDTH,
    Math.min(MAX_WIDTH, estimatedTotalWidth)
  );

  return finalWidth;
};

/**
 * 事件详情抽屉组件
 * 根据 pkg/skylarkq/DEVELOPMENT.md 规范实现
 */
const EventDetailDrawer: React.FC<EventDetailDrawerProps> = ({
  open,
  onClose,
  detail,
  loading,
  columns,
}) => {
  // 动态计算宽度
  const drawerWidth = React.useMemo(
    () => calculateDrawerWidth(detail),
    [detail]
  );

  if (loading) {
    return (
      <Drawer
        title="事件详情"
        width={drawerWidth}
        open={open}
        onClose={onClose}
        className="event-detail-drawer"
      >
        <div className="loading-container">
          <Spin size="large" tip="加载中...">
            <div />
          </Spin>
        </div>
      </Drawer>
    );
  }

  if (!detail) {
    return (
      <Drawer
        title="事件详情"
        width={drawerWidth}
        open={open}
        onClose={onClose}
        className="event-detail-drawer"
      >
        <Empty description="暂无数据" />
      </Drawer>
    );
  }

  const statusConfig = getStatusConfig(detail.currentStatus || "pending");

  return (
    <Drawer
      title="事件详情"
      width={drawerWidth}
      open={open}
      onClose={onClose}
      className="event-detail-drawer"
    >
      <div className="drawer-content">
        {/* 概要信息卡片 */}
        <Card size="small" className="summary-card">
          <Space direction="vertical" size="middle" className="summary-content">
            <Space size="large">
              <Space>
                <Text strong>事件ID:</Text>
                <Text copyable>{detail.journeyId}</Text>
              </Space>
              <Space>
                <Text strong>当前状态:</Text>
                <Tag color={statusConfig.color} icon={statusConfig.icon}>
                  {statusConfig.text}
                </Tag>
              </Space>
            </Space>

            <Space size="large">
              <Space>
                <UserOutlined />
                <Text strong>发起人:</Text>
                <Text>
                  {detail.initiatorUserName || detail.initiatorUserId || "-"}
                </Text>
              </Space>
              <Space>
                <ClockCircleOutlined />
                <Text strong>发起时间:</Text>
                <Text>
                  {detail.initiatedAt
                    ? dayjs
                        .utc(detail.initiatedAt)
                        .format("YYYY-MM-DD HH:mm:ss")
                    : "-"}
                </Text>
              </Space>
            </Space>
          </Space>
        </Card>

        {/* 最新业务数据 */}
        <Card
          title={
            <Title level={5} className="card-title">
              最新业务数据
            </Title>
          }
          size="small"
          className="card-spacing"
        >
          {(() => {
            if (
              !detail.latestBusinessData ||
              Object.keys(detail.latestBusinessData).length === 0
            ) {
              return (
                <Empty
                  description="无业务数据"
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                />
              );
            }

            // 后端已经过滤了可见字段，前端只需要按 columns 配置显示字段名和排序
            if (columns && columns.length > 0) {
              // 过滤出业务字段（排除系统字段）
              const businessColumns = columns.filter(
                (col) =>
                  ![
                    "slp_journey_id",
                    "slp_assignment_id",
                    "slp_status",
                    "vertex_name",
                  ].includes(col.field)
              );

              // 如果没有业务列配置，但有业务数据，说明配置有问题，直接显示原始字段
              if (businessColumns.length === 0) {
                const entries = Object.entries(detail.latestBusinessData);
                return (
                  <Descriptions
                    bordered
                    column={1}
                    size="small"
                    styles={{ label: { width: "180px", fontWeight: 500 } }}
                  >
                    {entries.map(([key, value]) => (
                      <Descriptions.Item label={key} key={key}>
                        {renderValue(value)}
                      </Descriptions.Item>
                    ))}
                  </Descriptions>
                );
              }

              // 按照 columns 配置的顺序显示（使用 displayName）
              return (
                <Descriptions
                  bordered
                  column={1}
                  size="small"
                  styles={{ label: { width: "180px", fontWeight: 500 } }}
                >
                  {businessColumns
                    .map((col) => {
                      // 如果该字段在业务数据中存在，则显示
                      if (col.field in detail.latestBusinessData) {
                        return (
                          <Descriptions.Item
                            label={col.displayName || col.field}
                            key={col.field}
                          >
                            {renderValue(detail.latestBusinessData[col.field])}
                          </Descriptions.Item>
                        );
                      }
                      return null;
                    })
                    .filter(Boolean)}
                </Descriptions>
              );
            }

            // 兜底：没有 columns 配置时，直接显示所有业务数据
            const entries = Object.entries(detail.latestBusinessData);
            return (
              <Descriptions
                bordered
                column={1}
                size="small"
                styles={{ label: { width: "180px", fontWeight: 500 } }}
              >
                {entries.map(([key, value]) => (
                  <Descriptions.Item label={key} key={key}>
                    {renderValue(value)}
                  </Descriptions.Item>
                ))}
              </Descriptions>
            );
          })()}
        </Card>

        {/* 流程流转历史 */}
        <Card
          title={
            <Title level={5} className="card-title">
              流程流转历史
            </Title>
          }
          size="small"
        >
          {detail.flowHistory && detail.flowHistory.length > 0 ? (
            <Timeline
              className="timeline-container"
              items={detail.flowHistory.map((node: FlowNode, index: number) => {
                // 判断是否为当前节点（最后一个节点）
                const isCurrentNode = index === detail.flowHistory.length - 1;
                // 判断是否为起点节点（第一个节点）
                const isStartNode = index === 0;
                // 只有当前节点才显示状态
                const nodeStatusConfig = isCurrentNode
                  ? getStatusConfig(detail.currentStatus)
                  : { color: "default", text: "", icon: null };

                // 节点颜色：起点用 blue，当前节点用状态色，中间节点用 default
                const nodeColor = isCurrentNode
                  ? nodeStatusConfig.color
                  : isStartNode
                    ? "blue"
                    : "default";

                return {
                  color: nodeColor,
                  dot: isCurrentNode ? nodeStatusConfig.icon : undefined,
                  children: (
                    <div className="timeline-node">
                      <Space
                        direction="vertical"
                        size={6}
                        className="node-content"
                      >
                        {/* 第一行：时间（较小字体，次要信息） */}
                        <Text type="secondary" className="node-time">
                          {node.createdAt
                            ? dayjs
                                .utc(node.createdAt)
                                .format("YYYY-MM-DD HH:mm:ss")
                            : "-"}
                        </Text>

                        {/* 第二行：节点名称和状态 */}
                        <Space>
                          <Text strong className="node-title">
                            {node.vertexName || node.vertexAlias || "节点"}
                          </Text>
                          {isCurrentNode && (
                            <Tag color={nodeStatusConfig.color}>
                              {nodeStatusConfig.text}
                            </Tag>
                          )}
                        </Space>

                        {/* 第三行：处理人和更新时间 */}
                        <Space size="middle" wrap className="node-meta">
                          <Space size="small">
                            <UserOutlined className="user-icon" />
                            <Text type="secondary" className="user-text">
                              {node.userNames && node.userNames.length > 0
                                ? node.userNames.join("、")
                                : "-"}
                            </Text>
                          </Space>
                          {node.updatedAt &&
                            node.updatedAt !== node.createdAt && (
                              <Text type="secondary" className="update-time">
                                (更新于{" "}
                                {dayjs
                                  .utc(node.updatedAt)
                                  .format("MM-DD HH:mm:ss")}
                                )
                              </Text>
                            )}
                        </Space>
                      </Space>
                    </div>
                  ),
                };
              })}
            />
          ) : (
            <Empty
              description="暂无流转记录"
              image={Empty.PRESENTED_IMAGE_SIMPLE}
            />
          )}
        </Card>
      </div>
    </Drawer>
  );
};

export default EventDetailDrawer;
