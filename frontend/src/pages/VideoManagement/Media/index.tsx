import VideoPlayer from "@/components/VideoPlayer";
import type { StreamItem } from "@/services/video";
import type { ActionType } from "@ant-design/pro-components";
import { PageContainer } from "@ant-design/pro-components";
import { Badge, Button, Space, Tag, Tooltip } from "antd";
import React, { useEffect, useRef, useState } from "react";
import MediaTable, { type ListStats } from "./components/MediaTable";
import NetworkMonitor from "./components/NetworkMonitor";
import PlayUrlModal from "./components/PlayUrlModal";
import DetailDrawer from "./components/StreamDetailDrawer/DetailDrawer";
import useDetailPolling from "./hooks/useDetailPolling";
import useMediaSSE from "./hooks/useMediaSSE";
import useSystemSSE from "./hooks/useSystemSSE";

const MediaManagement: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [autoRefresh, setAutoRefresh] = useState(true);
  const summary = useMediaSSE();
  const { systemData, isConnected, error, reconnect } = useSystemSSE();
  const [playUrlModalVisible, setPlayUrlModalVisible] = useState(false);
  const [selectedPath, setSelectedPath] = useState<string>("");
  const [listStats, setListStats] = useState<ListStats>({});
  const [videoPlayerVisible, setVideoPlayerVisible] = useState(false);
  const [playingChannel, setPlayingChannel] = useState<{
    deviceId: string;
    channelId: string;
    channelName?: string;
  } | null>(null);

  // 详情抽屉
  const [detailVisible, setDetailVisible] = useState(false);
  const [detailPath, setDetailPath] = useState<string>("");
  const {
    info: detailInfo,
    subscribers: detailSubscribers,
    metrics: metricsHistory,
  } = useDetailPolling(detailVisible, detailPath);

  // 列表定时轮询刷新
  useEffect(() => {
    if (!autoRefresh) return;
    const timer = setInterval(() => {
      actionRef.current?.reload();
    }, 3000);
    return () => clearInterval(timer);
  }, [autoRefresh]);

  return (
    <PageContainer
      header={{
        title: "媒体管理",
        subTitle: "流媒体状态与实时信息",
        extra: [
          <Space key="extra">
            <Tooltip title={autoRefresh ? "自动刷新中" : "已暂停自动刷新"}>
              <Badge
                status={autoRefresh ? "processing" : "default"}
                text={autoRefresh ? "自动刷新" : "手动刷新"}
              />
            </Tooltip>
            <Tag color="blue">流: {listStats.streams ?? "-"}</Tag>
            <Tag color="purple">订阅: {listStats.subscribers ?? "-"}</Tag>
            <Tag color="geekblue">拉流: {summary?.pullCount ?? "-"}</Tag>
            <Button onClick={() => setAutoRefresh((v) => !v)}>
              {autoRefresh ? "暂停自动刷新" : "恢复自动刷新"}
            </Button>
          </Space>,
        ],
      }}
    >
      {/* 网络监控 */}
      <div style={{ marginBottom: 24 }}>
        <NetworkMonitor
          networkData={systemData?.netWork}
          isConnected={isConnected}
          error={error}
          onReconnect={reconnect}
        />
      </div>

      <MediaTable
        actionRef={actionRef as any}
        onPreview={(deviceId, channelId, channelName) => {
          setPlayingChannel({ deviceId, channelId, channelName });
          setVideoPlayerVisible(true);
        }}
        onShowUrl={(p) => {
          setSelectedPath(p);
          setPlayUrlModalVisible(true);
        }}
        onShowDetail={(record: StreamItem) => {
          setDetailPath(record.path);
          setDetailVisible(true);
        }}
        onStatsChange={(s) => setListStats(s)}
      />

      <PlayUrlModal
        open={playUrlModalVisible}
        path={selectedPath}
        onClose={() => setPlayUrlModalVisible(false)}
      />

      <DetailDrawer
        open={detailVisible}
        path={detailPath}
        info={detailInfo}
        subscribers={detailSubscribers}
        metrics={metricsHistory}
        autoRefresh={autoRefresh}
        onClose={() => setDetailVisible(false)}
      />

      {playingChannel && (
        <VideoPlayer
          deviceId={playingChannel.deviceId}
          channelId={playingChannel.channelId}
          channelName={playingChannel.channelName}
          visible={videoPlayerVisible}
          onClose={() => {
            setVideoPlayerVisible(false);
            setPlayingChannel(null);
          }}
          mode="overlay"
        />
      )}
    </PageContainer>
  );
};

export default MediaManagement;
