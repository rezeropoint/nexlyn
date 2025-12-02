import {
  AppstoreOutlined,
  BlockOutlined,
  BorderOutlined,
  ExpandOutlined,
  FullscreenExitOutlined,
  FullscreenOutlined,
} from "@ant-design/icons";
import { PageContainer, ProCard } from "@ant-design/pro-components";
import { useModel } from "@umijs/max";
import { Button, FloatButton, message, Radio, Space, Tooltip } from "antd";
import React, { useCallback, useEffect, useRef, useState } from "react";
import DeviceTree from "./components/DeviceTree";
import VideoGrid from "./components/VideoGrid";
import styles from "./index.less";
import type { ChannelPlayInfo } from "./types";

export type GridLayout = 1 | 4 | 9 | 16;

const SplitScreenMonitor: React.FC = () => {
  const { initialState } = useModel("@@initialState");
  const [gridLayout, setGridLayout] = useState<GridLayout>(4);
  const [playingChannels, setPlayingChannels] = useState<
    Map<number, ChannelPlayInfo>
  >(new Map());
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [organizationId, setOrganizationId] = useState<string>();
  const videoGridRef = useRef<HTMLDivElement>(null);

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

  // 处理通道播放 - 使用 useCallback + 函数式更新避免依赖 playingChannels
  const handleChannelPlay = useCallback(
    (gridIndex: number, channelInfo: ChannelPlayInfo) => {
      setPlayingChannels((prev) => {
        const newMap = new Map(prev);
        newMap.set(gridIndex, channelInfo);
        return newMap;
      });
    },
    []
  );

  // 停止播放 - 使用 useCallback + 函数式更新避免依赖 playingChannels
  const handleStopPlay = useCallback((gridIndex: number) => {
    setPlayingChannels((prev) => {
      const newMap = new Map(prev);
      newMap.delete(gridIndex);
      return newMap;
    });
  }, []);

  // 清空所有播放
  const handleClearAll = () => {
    setPlayingChannels(new Map());
  };

  // 全屏切换
  const handleFullscreenToggle = useCallback(async () => {
    if (!videoGridRef.current) {
      message.error("视频区域未找到");
      return;
    }

    try {
      if (!isFullscreen) {
        // 进入全屏
        if (videoGridRef.current.requestFullscreen) {
          await videoGridRef.current.requestFullscreen();
        } else if ((videoGridRef.current as any).webkitRequestFullscreen) {
          await (videoGridRef.current as any).webkitRequestFullscreen();
        } else if ((videoGridRef.current as any).msRequestFullscreen) {
          await (videoGridRef.current as any).msRequestFullscreen();
        }
      } else {
        // 退出全屏
        if (document.exitFullscreen) {
          await document.exitFullscreen();
        } else if ((document as any).webkitExitFullscreen) {
          await (document as any).webkitExitFullscreen();
        } else if ((document as any).msExitFullscreen) {
          await (document as any).msExitFullscreen();
        }
      }
    } catch (error) {
      console.error("全屏切换失败:", error);
      message.error("全屏切换失败");
    }
  }, [isFullscreen]);

  // 监听全屏状态变化
  React.useEffect(() => {
    const handleFullscreenChange = () => {
      const fullscreenElement =
        document.fullscreenElement ||
        (document as any).webkitFullscreenElement ||
        (document as any).msFullscreenElement;
      setIsFullscreen(!!fullscreenElement);
    };

    document.addEventListener("fullscreenchange", handleFullscreenChange);
    document.addEventListener("webkitfullscreenchange", handleFullscreenChange);
    document.addEventListener("msfullscreenchange", handleFullscreenChange);

    return () => {
      document.removeEventListener("fullscreenchange", handleFullscreenChange);
      document.removeEventListener(
        "webkitfullscreenchange",
        handleFullscreenChange
      );
      document.removeEventListener(
        "msfullscreenchange",
        handleFullscreenChange
      );
    };
  }, []);

  // 布局切换选项
  const layoutOptions = [
    { label: "单屏", value: 1, icon: <BorderOutlined /> },
    { label: "四分屏", value: 4, icon: <BlockOutlined /> },
    { label: "九分屏", value: 9, icon: <AppstoreOutlined /> },
    { label: "十六分屏", value: 16, icon: <ExpandOutlined /> },
  ];

  return (
    <PageContainer
      header={{
        title: "分屏监控",
        subTitle: "多路视频同时监控",
        extra: [
          <Space key="extra">
            <Radio.Group
              value={gridLayout}
              onChange={(e) => setGridLayout(e.target.value)}
              optionType="button"
              buttonStyle="solid"
            >
              {layoutOptions.map((option) => (
                <Radio.Button key={option.value} value={option.value}>
                  <Tooltip title={option.label}>
                    <Space>
                      {option.icon}
                      {option.label}
                    </Space>
                  </Tooltip>
                </Radio.Button>
              ))}
            </Radio.Group>
            <Button
              icon={
                isFullscreen ? (
                  <FullscreenExitOutlined />
                ) : (
                  <FullscreenOutlined />
                )
              }
              onClick={handleFullscreenToggle}
              type="default"
            >
              {isFullscreen ? "退出全屏" : "全屏"}
            </Button>
            <Button
              type="primary"
              danger
              onClick={handleClearAll}
              disabled={playingChannels.size === 0}
            >
              清空所有
            </Button>
          </Space>,
        ],
      }}
    >
      <div style={{ height: "calc(100vh - 112px)" }}>
        {/* 左右分栏布局 */}
        <ProCard split="vertical" gutter={16}>
          {/* 左侧设备树 */}
          <ProCard
            colSpan="400px"
            title="设备通道"
            className={styles.deviceTreeCard}
          >
            <DeviceTree
              onChannelPlay={handleChannelPlay}
              playingChannels={playingChannels}
              maxGrids={gridLayout}
              organizationId={organizationId}
              onOrganizationChange={setOrganizationId}
            />
          </ProCard>

          {/* 右侧分屏区域 */}
          <ProCard className={styles.videoGridCard} bordered={false}>
            <div ref={videoGridRef} className={styles.videoGridContainer}>
              <VideoGrid
                layout={gridLayout}
                playingChannels={playingChannels}
                onChannelPlay={handleChannelPlay}
                onStopPlay={handleStopPlay}
              />
            </div>
          </ProCard>
        </ProCard>
      </div>

      {/* 浮动按钮组 - 全屏时的快速操作 */}
      {isFullscreen && (
        <FloatButton.Group
          trigger="click"
          type="primary"
          icon={<AppstoreOutlined />}
          tooltip="快速操作"
        >
          <FloatButton
            icon={<BorderOutlined />}
            tooltip="单屏"
            onClick={() => setGridLayout(1)}
          />
          <FloatButton
            icon={<BlockOutlined />}
            tooltip="四分屏"
            onClick={() => setGridLayout(4)}
          />
          <FloatButton
            icon={<AppstoreOutlined />}
            tooltip="九分屏"
            onClick={() => setGridLayout(9)}
          />
          <FloatButton
            icon={<ExpandOutlined />}
            tooltip="十六分屏"
            onClick={() => setGridLayout(16)}
          />
          <FloatButton
            icon={<FullscreenExitOutlined />}
            tooltip="退出全屏"
            onClick={handleFullscreenToggle}
          />
        </FloatButton.Group>
      )}
    </PageContainer>
  );
};

export default SplitScreenMonitor;
