import { getEventDetail } from "@/services/eventhandler";
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  UserOutlined,
} from "@ant-design/icons";
import {
  Alert,
  Badge,
  Card,
  Descriptions,
  Drawer,
  Empty,
  Space,
  Spin,
  Tag,
  Timeline,
} from "antd";
import dayjs from "dayjs";
import React, { useEffect, useState } from "react";
import { DATE_FORMAT, FLOW_STATUS_MAP } from "../constants";
import type { FlowNode } from "../types";
import styles from "./EventDetailDrawer.less";

interface EventDetailDrawerProps {
  open: boolean;
  onClose: () => void;
  eventConfigId: string;
  journeyId: number;
}

/**
 * 事件详情抽屉
 * 展示事件的完整信息和流转历史
 */
const EventDetailDrawer: React.FC<EventDetailDrawerProps> = ({
  open,
  onClose,
  eventConfigId,
  journeyId,
}) => {
  const [loading, setLoading] = useState(false);
  const [eventDetail, setEventDetail] = useState<any>(null);

  // 加载事件详情
  const loadEventDetail = async () => {
    setLoading(true);
    try {
      const response = await getEventDetail(eventConfigId, journeyId);
      if (response.code === 0 && response.data) {
        setEventDetail(response.data);
      }
    } catch (error) {
      console.error("加载事件详情失败:", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (open && eventConfigId && journeyId) {
      loadEventDetail();
    }
  }, [open, eventConfigId, journeyId]);

  // 格式化时间
  const formatDateTime = (dateStr: string) => {
    return dateStr ? dayjs(dateStr).format(DATE_FORMAT.DATETIME) : "-";
  };

  // 渲染流程状态
  const renderStatus = (status: string) => {
    const config = FLOW_STATUS_MAP[status as keyof typeof FLOW_STATUS_MAP];
    return config ? (
      <Badge status={config.status as any} text={config.text} />
    ) : (
      <Tag>{status}</Tag>
    );
  };

  // 渲染业务数据
  const renderBusinessData = (data: Record<string, any>) => {
    const entries = Object.entries(data).filter(
      ([key]) => !key.startsWith("slp_")
    );

    if (entries.length === 0) {
      return <Empty description="暂无业务数据" />;
    }

    return (
      <Descriptions bordered size="small" column={1}>
        {entries.map(([key, value]) => (
          <Descriptions.Item key={key} label={key}>
            {value !== null && value !== undefined ? String(value) : "-"}
          </Descriptions.Item>
        ))}
      </Descriptions>
    );
  };

  // 渲染流转历史时间线
  const renderFlowHistory = (history: FlowNode[]) => {
    if (!history || history.length === 0) {
      return <Empty description="暂无流转历史" />;
    }

    return (
      <Timeline
        mode="left"
        items={history.map((node, index) => {
          const isFirst = index === 0;
          const isLast = index === history.length - 1;

          return {
            key: node.vertexId,
            color: isLast ? "green" : "blue",
            dot: isFirst ? (
              <ClockCircleOutlined />
            ) : isLast ? (
              <CheckCircleOutlined />
            ) : undefined,
            children: (
              <>
                <div className={styles.timelineNode}>
                  <Tag color="blue">{node.vertexName || node.vertexAlias}</Tag>
                  {/* 只在最后一个节点显示状态 */}
                  {isLast &&
                    eventDetail?.currentStatus &&
                    renderStatus(eventDetail.currentStatus)}
                </div>
                <div className={styles.timelineUser}>
                  <Space>
                    <UserOutlined />
                    {/* 支持多人处理：用"、"连接所有用户名 */}
                    {node.userNames && node.userNames.length > 0
                      ? node.userNames.join("、")
                      : "未知用户"}
                  </Space>
                </div>
                <div className={styles.timelineTime}>
                  创建：{formatDateTime(node.createdAt)}
                  {node.updatedAt !== node.createdAt && (
                    <>
                      <br />
                      更新：{formatDateTime(node.updatedAt)}
                    </>
                  )}
                </div>
              </>
            ),
          };
        })}
      />
    );
  };

  return (
    <Drawer
      title={`事件详情 - Journey #${journeyId}`}
      placement="right"
      onClose={onClose}
      open={open}
      width={800}
      destroyOnHidden
    >
      <Spin spinning={loading}>
        {eventDetail ? (
          <div className={styles.drawerContent}>
            {/* 概要信息 */}
            <Alert
              message="事件概要"
              description={
                <Descriptions size="small" column={2}>
                  <Descriptions.Item label="Journey ID">
                    {eventDetail.journeyId}
                  </Descriptions.Item>
                  <Descriptions.Item label="当前状态">
                    {renderStatus(eventDetail.currentStatus)}
                  </Descriptions.Item>
                  <Descriptions.Item label="发起人">
                    {eventDetail.initiatorUserName ||
                      eventDetail.initiatorUserId ||
                      "-"}
                  </Descriptions.Item>
                  <Descriptions.Item label="发起时间">
                    {formatDateTime(eventDetail.initiatedAt)}
                  </Descriptions.Item>
                </Descriptions>
              }
              type="info"
              className={styles.sectionMargin}
            />

            {/* 业务数据 */}
            <Card
              title="最新业务数据"
              size="small"
              className={styles.sectionMargin}
            >
              {renderBusinessData(eventDetail.latestBusinessData || {})}
            </Card>

            {/* 流转历史 */}
            <Card title="流程流转历史" size="small">
              {renderFlowHistory(eventDetail.flowHistory || [])}
            </Card>
          </div>
        ) : (
          !loading && <Empty description="暂无数据" />
        )}
      </Spin>
    </Drawer>
  );
};

export default EventDetailDrawer;
