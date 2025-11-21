import { ReloadOutlined, WifiOutlined } from "@ant-design/icons";
import { Line } from "@ant-design/plots";
import { Alert, Button, Card, Space, Statistic, Typography, theme } from "antd";
import dayjs from "dayjs";
import React, { useEffect, useState } from "react";
import type { SystemNetworkInfo } from "../hooks/useSystemSSE";
import styles from "./NetworkMonitor.less";

const { Text } = Typography;

interface NetworkDataPoint {
  time: string;
  type: "上行" | "下行";
  value: number;
  interface: string;
}

interface NetworkMonitorProps {
  networkData?: SystemNetworkInfo[];
  isConnected: boolean;
  error?: string | null;
  onReconnect?: () => void;
}

const NetworkMonitor: React.FC<NetworkMonitorProps> = ({
  networkData = [],
  isConnected,
  error,
  onReconnect,
}) => {
  const { token } = theme.useToken();
  const [historicalData, setHistoricalData] = useState<NetworkDataPoint[]>([]);
  const [currentStats, setCurrentStats] = useState<{
    totalReceive: number;
    totalSent: number;
    totalReceiveSpeed: number;
    totalSentSpeed: number;
  }>({
    totalReceive: 0,
    totalSent: 0,
    totalReceiveSpeed: 0,
    totalSentSpeed: 0,
  });

  // 格式化网络速度
  const formatSpeed = (bytes: number) => {
    if (bytes < 1024) return `${bytes.toFixed(2)} B/s`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB/s`;
    if (bytes < 1024 * 1024 * 1024)
      return `${(bytes / (1024 * 1024)).toFixed(2)} MB/s`;
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB/s`;
  };

  // 格式化总流量
  const formatBytes = (bytes: number) => {
    if (bytes < 1024) return `${bytes.toFixed(2)} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`;
    if (bytes < 1024 * 1024 * 1024)
      return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
  };

  // 更新历史数据和当前统计
  useEffect(() => {
    if (!networkData || networkData.length === 0) return;

    const now = dayjs().format("HH:mm:ss");
    const newPoints: NetworkDataPoint[] = [];

    let totalReceive = 0;
    let totalSent = 0;
    let totalReceiveSpeed = 0;
    let totalSentSpeed = 0;

    // 过滤掉 lo（本地回环）接口，只显示实际网络接口
    const activeInterfaces = networkData.filter(
      (net) =>
        net.name && net.name !== "lo" && (net.receiveSpeed || net.sentSpeed)
    );

    activeInterfaces.forEach((net) => {
      totalReceive += net.receive || 0;
      totalSent += net.sent || 0;
      totalReceiveSpeed += net.receiveSpeed || 0;
      totalSentSpeed += net.sentSpeed || 0;

      // 只记录有速度数据的网络接口
      if (
        net.receiveSpeed &&
        typeof net.receiveSpeed === "number" &&
        !Number.isNaN(net.receiveSpeed)
      ) {
        newPoints.push({
          time: now,
          type: "下行",
          value: net.receiveSpeed / 1024, // 转换为 KB/s
          interface: net.name,
        });
      }

      if (
        net.sentSpeed &&
        typeof net.sentSpeed === "number" &&
        !Number.isNaN(net.sentSpeed)
      ) {
        newPoints.push({
          time: now,
          type: "上行",
          value: net.sentSpeed / 1024, // 转换为 KB/s
          interface: net.name,
        });
      }
    });

    setCurrentStats({
      totalReceive,
      totalSent,
      totalReceiveSpeed,
      totalSentSpeed,
    });

    // 更新历史数据，保留最近60个时间点的数据
    setHistoricalData((prev) => {
      const updated = [...prev, ...newPoints];
      // 按时间分组，保留最近60个时间点
      const timeGroups = updated.reduce((acc, point) => {
        if (!acc[point.time]) acc[point.time] = [];
        acc[point.time].push(point);
        return acc;
      }, {} as Record<string, NetworkDataPoint[]>);

      const sortedTimes = Object.keys(timeGroups).sort();
      const recentTimes = sortedTimes.slice(-60); // 保留最近60个时间点

      return recentTimes.flatMap((time) => timeGroups[time]);
    });
  }, [networkData]);

  const lineConfig = {
    data: historicalData,
    xField: "time",
    yField: "value",
    seriesField: "type",
    smooth: true,
    animation: {
      appear: {
        animation: "path-in",
        duration: 1000,
      },
    },
    color: [token.colorPrimary, token.colorSuccess],
    point: {
      size: 3,
      shape: "circle",
    },
    line: {
      size: 2,
    },
    xAxis: {
      type: "cat",
      tickCount: 6,
      label: {
        formatter: (text: string) => {
          // 只显示每10个数据点的标签
          const index = historicalData.findIndex((d) => d.time === text);
          return index % 10 === 0 ? text : "";
        },
      },
    },
    yAxis: {
      label: {
        formatter: (value: string) => `${value} KB/s`,
      },
    },
    legend: {
      position: "top" as const,
    },
    tooltip: {
      formatter: (datum: NetworkDataPoint) => {
        const value = datum.value;
        const formattedValue =
          typeof value === "number" && !Number.isNaN(value)
            ? `${value.toFixed(2)} KB/s`
            : "0.00 KB/s";
        return {
          name: datum.type,
          value: formattedValue,
        };
      },
    },
  };

  return (
    <Card
      title={
        <Space>
          <WifiOutlined />
          <span>网络监控</span>
          {!isConnected && (
            <Button
              type="link"
              size="small"
              icon={<ReloadOutlined />}
              onClick={onReconnect}
            >
              重连
            </Button>
          )}
        </Space>
      }
      size="small"
    >
      {error && (
        <Alert
          message={error}
          type="warning"
          showIcon
          closable
          className={styles.alertSpacing}
        />
      )}

      <Space
        direction="vertical"
        className={styles.spaceContainer}
        size="middle"
      >
        <div className={styles.statisticsContainer}>
          <Statistic
            title="实时下行速度(被推流)"
            value={formatSpeed(currentStats.totalReceiveSpeed)}
            valueStyle={{ color: token.colorPrimary }}
            className={styles.statisticValuePrimary}
          />
          <Statistic
            title="实时上行速度(播放)"
            value={formatSpeed(currentStats.totalSentSpeed)}
            valueStyle={{ color: token.colorSuccess }}
            className={styles.statisticValueSuccess}
          />
          <Statistic
            title="总下行流量"
            value={formatBytes(currentStats.totalReceive)}
            className={styles.statisticValueNormal}
          />
          <Statistic
            title="总上行流量"
            value={formatBytes(currentStats.totalSent)}
            className={styles.statisticValueNormal}
          />
        </div>

        <div className={styles.chartContainer}>
          {historicalData.length > 0 ? (
            <Line {...lineConfig} />
          ) : (
            <div className={styles.emptyDataContainer}>
              <Text type="secondary">等待网络数据...</Text>
            </div>
          )}
        </div>
      </Space>
    </Card>
  );
};

export default NetworkMonitor;
