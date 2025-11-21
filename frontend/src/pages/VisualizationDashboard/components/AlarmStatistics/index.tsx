import { useApp } from "@/utils/appContext";
import { Line, Pie } from "@ant-design/charts";
import {
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  WarningOutlined,
} from "@ant-design/icons";
import { useModel } from "@umijs/max";
import {
  Badge,
  Card,
  Col,
  Empty,
  List,
  Radio,
  Row,
  Spin,
  Statistic,
  Tag,
  theme,
} from "antd";
import dayjs from "dayjs";
import React, { useMemo, useState } from "react";
import { getAlarmLevelColor } from "../../utils";
import { getLineConfig, getPieConfig } from "./chartConfigs";
import { useAlarmData } from "./hooks/useAlarmData";
import "./index.less";
import type { AlarmStatisticsProps, AlarmStats, TimeRange } from "./types";
import { formatTimeDisplay, getAlarmIcon } from "./utils";

const AlarmStatistics: React.FC<AlarmStatisticsProps> = ({
  className,
  orgId,
}) => {
  const { message } = useApp();
  const { token } = theme.useToken();
  const { initialState } = useModel("@@initialState");
  const [timeRange, setTimeRange] = useState<TimeRange>("today");

  // 判断是否为暗色主题
  const isDark = useMemo(
    () =>
      initialState?.themeMode === "dark" ||
      initialState?.themeMode === "dark-compact",
    [initialState?.themeMode]
  );

  // 使用自定义hook加载数据
  const { loading, alarmRecords, hasError } = useAlarmData(
    orgId,
    timeRange,
    message
  );

  // 告警统计
  const alarmStats: AlarmStats = useMemo(() => {
    const stats: AlarmStats = {
      total: alarmRecords.length,
      serious: 0,
      important: 0,
      normal: 0,
      minor: 0,
    };

    alarmRecords.forEach((alarm) => {
      switch (alarm.level) {
        case "严重":
          stats.serious += 1;
          break;
        case "重要":
          stats.important += 1;
          break;
        case "一般":
          stats.normal += 1;
          break;
        case "轻微":
          stats.minor += 1;
          break;
      }
    });

    return stats;
  }, [alarmRecords]);

  // 趋势数据
  const trendData = useMemo(() => {
    if (alarmRecords.length === 0) {
      return [];
    }

    const trendMap = new Map<string, number>();
    const format = timeRange === "today" ? "HH:00" : "MM-DD";

    alarmRecords.forEach((alarm) => {
      const timeKey = dayjs(alarm.timestamp).format(format);
      trendMap.set(timeKey, (trendMap.get(timeKey) || 0) + 1);
    });

    return Array.from(trendMap.entries())
      .map(([date, count]) => ({
        time: date,
        count,
      }))
      .sort((a, b) => a.time.localeCompare(b.time));
  }, [alarmRecords, timeRange]);

  // 告警级别分布数据
  const levelDistribution = useMemo(() => {
    if (alarmStats.total === 0) return [];

    return [
      { level: "严重", count: alarmStats.serious },
      { level: "重要", count: alarmStats.important },
      { level: "一般", count: alarmStats.normal },
      { level: "轻微", count: alarmStats.minor },
    ].filter((item) => item.count > 0);
  }, [alarmStats]);

  // 告警级别配置（使用token）
  const levelConfigs = useMemo(
    () => [
      { level: "严重", count: alarmStats.serious, color: token.colorError },
      { level: "重要", count: alarmStats.important, color: token.colorWarning },
      {
        level: "一般",
        count: alarmStats.normal,
        color: token.colorWarningOutline,
      },
      { level: "轻微", count: alarmStats.minor, color: token.colorSuccess },
    ],
    [alarmStats, token]
  );

  // 图表配置
  const lineConfig = useMemo(
    () => getLineConfig(trendData, token, isDark),
    [trendData, token, isDark]
  );
  const pieConfig = useMemo(
    () => getPieConfig(levelDistribution, alarmStats, token, isDark),
    [levelDistribution, alarmStats, token, isDark]
  );

  // 显示空数据状态
  if (!loading && (hasError || alarmRecords.length === 0)) {
    return (
      <div className={`alarm-statistics ${className || ""}`}>
        <Card className="empty-card" size="small">
          <Empty
            description={hasError ? "未能获取到数据" : "暂无告警数据"}
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          />
        </Card>
      </div>
    );
  }

  return (
    <Spin spinning={loading}>
      <div className={`alarm-statistics ${className || ""}`}>
        {/* 时间范围选择 */}
        <Card className="range-card" size="small">
          <div className="range-selector">
            <Radio.Group
              value={timeRange}
              onChange={(e) => setTimeRange(e.target.value)}
              buttonStyle="solid"
              size="small"
            >
              <Radio.Button value="today">今日</Radio.Button>
              <Radio.Button value="week">本周</Radio.Button>
              <Radio.Button value="month">本月</Radio.Button>
            </Radio.Group>
          </div>
        </Card>

        {/* 告警概览 */}
        <Card className="overview-card" title="告警概览" size="small">
          <Row gutter={[12, 12]}>
            <Col span={12}>
              <Statistic
                title="总告警数"
                value={alarmStats.total}
                valueStyle={{ color: token.colorWarning, fontSize: "20px" }}
                prefix={<WarningOutlined />}
                className="total-alarm-stat"
              />
            </Col>
            <Col span={12}>
              <Statistic
                title="严重告警"
                value={alarmStats.serious}
                valueStyle={{ color: token.colorError, fontSize: "20px" }}
                prefix={<ExclamationCircleOutlined />}
                className="serious-alarm-stat"
              />
            </Col>
          </Row>

          <div className="level-stats">
            <Row gutter={[6, 8]}>
              {levelConfigs.map((item) => (
                <Col key={item.level} span={6}>
                  <div className="level-item">
                    <Badge count={item.count} color={item.color} size="small" />
                    <span className="level-label">{item.level}</span>
                  </div>
                </Col>
              ))}
            </Row>
          </div>
        </Card>

        {/* 告警趋势 */}
        <Card
          className="chart-card"
          title={`告警趋势 (${timeRange === "today" ? "小时" : "日期"})`}
          size="small"
        >
          <Line {...lineConfig} height={150} />
        </Card>

        {/* 告警级别分布 */}
        <Card className="chart-card" title="告警级别分布" size="small">
          <Pie {...pieConfig} height={180} />
        </Card>

        {/* 最新告警列表 */}
        <Card className="alarm-list-card" title="最新告警" size="small">
          <List
            className="alarm-list"
            size="small"
            dataSource={alarmRecords.slice(0, 10)}
            renderItem={(item) => (
              <List.Item className="alarm-item">
                <div className="alarm-content">
                  <div className="alarm-header">
                    <div className="alarm-icon">{getAlarmIcon(item.level)}</div>
                    <div className="alarm-info">
                      <div className="alarm-title">
                        {item.deviceName}
                        <Tag
                          color={getAlarmLevelColor(item.level)}
                          className="level-tag"
                        >
                          {item.level}
                        </Tag>
                      </div>
                      <div className="alarm-message">{item.message}</div>
                      <div className="alarm-meta">
                        <span className="alarm-type">{item.alarmType}</span>
                        <span className="alarm-location">
                          📍 {item.location}
                        </span>
                        <span className="alarm-time">
                          <ClockCircleOutlined className="time-icon" />
                          {formatTimeDisplay(item.timestamp)}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </List.Item>
            )}
          />
        </Card>
      </div>
    </Spin>
  );
};

export default AlarmStatistics;
