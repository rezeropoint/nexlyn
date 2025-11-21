import type { ColumnInfo } from "@/pages/EventManagement/types";
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  LoadingOutlined,
  PauseCircleOutlined,
  SwapOutlined,
  UserOutlined,
} from "@ant-design/icons";
import { Empty, Space, Spin, Tag, Timeline, Typography } from "antd";

const { Text } = Typography;
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import React from "react";
import styles from "./EventDetailContent.less";

// 启用 UTC 插件
dayjs.extend(utc);

/**
 * 渲染字段值（支持多种类型）
 */
const renderValue = (value: unknown): string => {
  if (value === null || value === undefined) return "-";
  if (typeof value === "boolean") return value ? "是" : "否";
  if (typeof value === "object") return JSON.stringify(value);
  return String(value);
};

/**
 * 获取状态配置（包含图标）
 */
const getStatusConfig = (status: string) => {
  const statusMap: Record<
    string,
    { color: string; text: string; icon: React.ReactNode }
  > = {
    stashed: {
      color: "default",
      text: "编写中",
      icon: <ExclamationCircleOutlined />,
    },
    processing: {
      color: "processing",
      text: "处理中",
      icon: <LoadingOutlined />,
    },
    approved: {
      color: "success",
      text: "已通过",
      icon: <CheckCircleOutlined />,
    },
    refused: { color: "warning", text: "已回退", icon: <SwapOutlined /> },
    transferred: { color: "cyan", text: "已转交", icon: <SwapOutlined /> },
    skipped: {
      color: "default",
      text: "已跳过",
      icon: <CloseCircleOutlined />,
    },
    cancelled: {
      color: "error",
      text: "已撤销",
      icon: <CloseCircleOutlined />,
    },
    receding: {
      color: "warning",
      text: "回退中",
      icon: <LoadingOutlined />,
    },
    suspended: {
      color: "default",
      text: "已暂停",
      icon: <PauseCircleOutlined />,
    },
    finished: {
      color: "success",
      text: "已完成",
      icon: <CheckCircleOutlined />,
    },
    aborted: { color: "error", text: "已终止", icon: <CloseCircleOutlined /> },
    pending: {
      color: "warning",
      text: "待处理",
      icon: <ClockCircleOutlined />,
    },
  };
  return (
    statusMap[status] || {
      color: "default",
      text: status,
      icon: <ExclamationCircleOutlined />,
    }
  );
};

export interface EventDetailData {
  journeyId: number;
  currentStatus: string;
  initiatorUserId: string;
  initiatorUserName: string;
  initiatedAt: string;
  latestBusinessData: Record<string, unknown>;
  flowHistory: Array<{
    vertexId: number;
    vertexName: string;
    vertexAlias: string;
    userNames: string[];
    createdAt: string;
    updatedAt: string;
  }>;
}

export interface EventDetailContentProps {
  loading: boolean;
  detail: EventDetailData | null;
  columns: ColumnInfo[];
}

/**
 * 事件详情内容组件（纯展示）
 */
const EventDetailContent: React.FC<EventDetailContentProps> = ({
  loading,
  detail,
  columns,
}) => {
  if (loading) {
    return (
      <div
        className={`${styles["event-detail-content"]} ${styles["detail-loading"]}`}
      >
        <Spin size="small" />
      </div>
    );
  }

  if (!detail) {
    return (
      <div
        className={`${styles["event-detail-content"]} ${styles["detail-empty"]}`}
      >
        <Empty description="无详情数据" image={Empty.PRESENTED_IMAGE_SIMPLE} />
      </div>
    );
  }

  const statusConfig = getStatusConfig(detail.currentStatus || "pending");

  // 获取业务字段（排除系统字段）
  const businessFields = columns.filter(
    (col) =>
      ![
        "slp_journey_id",
        "slp_assignment_id",
        "slp_status",
        "vertex_name",
      ].includes(col.field)
  );

  return (
    <div
      className={`${styles["event-detail-content"]} ${styles["item-detail"]}`}
    >
      {/* 基本信息行 */}
      <div className={styles["detail-header"]}>
        {/* 状态独立一行 */}
        <div className={styles["header-status"]}>
          <Tag color={statusConfig.color} icon={statusConfig.icon}>
            {statusConfig.text}
          </Tag>
        </div>
        {/* 发起人和时间在同一行 */}
        <div className={styles["header-info"]}>
          <span className={styles["detail-info"]}>
            <UserOutlined style={{ marginRight: 4 }} />
            {detail.initiatorUserName || detail.initiatorUserId || "-"}
          </span>
          <span className={styles["detail-info"]}>
            <ClockCircleOutlined style={{ marginRight: 4 }} />
            {detail.initiatedAt
              ? dayjs.utc(detail.initiatedAt).format("MM-DD HH:mm")
              : "-"}
          </span>
        </div>
      </div>

      {/* 业务数据 */}
      {businessFields.length > 0 && detail.latestBusinessData && (
        <div className={styles["detail-section"]}>
          <div className={styles["section-title"]}>业务数据</div>
          <div className={styles["detail-fields"]}>
            {businessFields.map((col) => {
              const value = detail.latestBusinessData[col.field];
              if (value === null || value === undefined) return null;

              return (
                <div key={col.field} className={styles["field-row"]}>
                  <span className={styles["field-label"]}>
                    {col.displayName || col.field}:
                  </span>
                  <span
                    className={styles["field-value"]}
                    title={renderValue(value)}
                  >
                    {renderValue(value)}
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* 流程历史（使用 Timeline 组件） */}
      {detail.flowHistory && detail.flowHistory.length > 0 && (
        <div className={styles["detail-section"]}>
          <div className={styles["section-title"]}>流程历史</div>
          <Timeline
            className={styles["timeline-container"]}
            items={detail.flowHistory.map((node, index) => {
              // 判断是否为当前节点（最后一个节点）
              const isCurrentNode = index === detail.flowHistory.length - 1;
              // 判断是否为起点节点（第一个节点）
              const isStartNode = index === 0;
              // 只有当前节点才显示状态
              const nodeStatusConfig = isCurrentNode
                ? getStatusConfig(detail.currentStatus || "pending")
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
                  <div className={styles["timeline-node"]}>
                    <Space
                      direction="vertical"
                      size={4}
                      className={styles["node-content"]}
                    >
                      {/* 第一行：时间（较小字体，次要信息） */}
                      <Text type="secondary" className={styles["node-time"]}>
                        {node.updatedAt
                          ? dayjs.utc(node.updatedAt).format("MM-DD HH:mm")
                          : node.createdAt
                            ? dayjs.utc(node.createdAt).format("MM-DD HH:mm")
                            : "-"}
                      </Text>

                      {/* 第二行：节点名称和状态 */}
                      <Space size={6}>
                        <Text strong className={styles["node-title"]}>
                          {node.vertexName || node.vertexAlias || "节点"}
                        </Text>
                        {isCurrentNode && (
                          <Tag color={nodeStatusConfig.color}>
                            {nodeStatusConfig.text}
                          </Tag>
                        )}
                      </Space>

                      {/* 第三行：处理人 */}
                      <Space size="small" className={styles["node-meta"]}>
                        <UserOutlined className={styles["user-icon"]} />
                        <Text type="secondary" className={styles["user-text"]}>
                          {node.userNames && node.userNames.length > 0
                            ? node.userNames.join("、")
                            : "-"}
                        </Text>
                      </Space>
                    </Space>
                  </div>
                ),
              };
            })}
          />
        </div>
      )}
    </div>
  );
};

export default EventDetailContent;
