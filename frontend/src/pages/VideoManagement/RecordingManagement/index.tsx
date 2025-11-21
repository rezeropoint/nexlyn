import { PageContainer, ProCard } from "@ant-design/pro-components";
import { useModel } from "@umijs/max";
import { message } from "antd";
import React, { useCallback, useEffect, useState } from "react";
import DeviceTree from "./components/DeviceTree";
import RecordingList from "./components/RecordingList";
import RecordingPlayer from "./components/RecordingPlayer";
import styles from "./index.less";

export interface RecordingItem {
  deviceId: string;
  name: string;
  filePath: string;
  address: string;
  startTime: string;
  endTime: string;
  secrecy: number;
  type: string;
  recorderId: string;
}

export interface ChannelInfo {
  deviceId: string;
  channelId: string;
  deviceName?: string;
  channelName?: string;
}

const RecordingManagement: React.FC = () => {
  const { initialState } = useModel("@@initialState");
  const [selectedChannel, setSelectedChannel] = useState<ChannelInfo | null>(
    null
  );
  const [recordings, setRecordings] = useState<RecordingItem[]>([]);
  const [playingRecording, setPlayingRecording] =
    useState<RecordingItem | null>(null);
  const [customStartTime, setCustomStartTime] = useState<number | undefined>(
    undefined
  );
  const [organizationId, setOrganizationId] = useState<string>();
  const [loadingRecordings, setLoadingRecordings] = useState(false);

  // 获取用户默认组织（使用 organizationIds 的第一个值）
  useEffect(() => {
    try {
      const orgIds = initialState?.currentUser?.organizationIds;
      if (orgIds && orgIds.length > 0) {
        setOrganizationId((prev) => prev ?? orgIds[0]);
      }
    } catch (error) {
      console.error("获取默认组织失败:", error);
    }
  }, [initialState?.currentUser]);

  // 处理通道选择
  const handleChannelSelect = useCallback((channelInfo: ChannelInfo) => {
    setSelectedChannel(channelInfo);
    setPlayingRecording(null); // 清空当前播放
    setRecordings([]); // 清空录像列表
    message.info(
      `已选择通道: ${channelInfo.channelName || channelInfo.channelId}`
    );
  }, []);

  // 处理录像数据更新
  const handleRecordingsUpdate = useCallback(
    (recordingData: RecordingItem[]) => {
      setRecordings(recordingData);
    },
    []
  );

  // 处理录像播放
  const handlePlayRecording = useCallback(
    (recording: RecordingItem, customStart?: number) => {
      if (!selectedChannel) {
        message.error("请先选择通道");
        return;
      }
      setPlayingRecording(recording);
      setCustomStartTime(customStart);

      if (customStart) {
        const startTime = new Date(customStart).toLocaleTimeString();
        message.success(
          `开始播放录像: ${recording.name}，从 ${startTime} 开始`
        );
      } else {
        message.success(`开始播放录像: ${recording.name}`);
      }
    },
    [selectedChannel]
  );

  return (
    <PageContainer
      header={{
        title: "录像管理",
        subTitle: "设备录像查询与回放",
      }}
    >
      <div
        style={{
          display: "flex",
          gap: "16px",
          width: "100%",
          maxWidth: "100vw",
        }}
      >
        {/* 左侧设备树 */}
        <div style={{ width: "350px", flexShrink: 0 }}>
          <ProCard title="设备通道" className={styles.deviceTreeCard}>
            <DeviceTree
              onChannelSelect={handleChannelSelect}
              selectedChannel={selectedChannel}
              organizationId={organizationId}
              onOrganizationChange={setOrganizationId}
            />
          </ProCard>
        </div>

        {/* 右侧录像区域 */}
        <div
          style={{
            flex: 1,
            display: "flex",
            flexDirection: "column",
            gap: "16px",
            minWidth: 0,
          }}
        >
          {/* 录像播放器 */}
          <ProCard className={styles.videoPlayerCard}>
            <RecordingPlayer
              selectedChannel={selectedChannel}
              playingRecording={playingRecording}
              customStartTime={customStartTime}
            />
          </ProCard>

          {/* 录像列表 */}
          <ProCard title="录像列表">
            <RecordingList
              selectedChannel={selectedChannel}
              recordings={recordings}
              onRecordingsUpdate={handleRecordingsUpdate}
              onPlayRecording={handlePlayRecording}
              playingRecording={playingRecording}
              loading={loadingRecordings}
              onLoadingChange={setLoadingRecordings}
            />
          </ProCard>
        </div>
      </div>
    </PageContainer>
  );
};

export default RecordingManagement;
