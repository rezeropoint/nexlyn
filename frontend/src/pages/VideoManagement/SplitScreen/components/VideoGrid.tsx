import { CloseOutlined, PlayCircleOutlined } from "@ant-design/icons";
import { Button, Flex, message } from "antd";
import classNames from "classnames";
import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { GridLayout } from "../index";
import type { ChannelPlayInfo } from "../types";
import SimplifiedVideoPlayer from "./SimplifiedVideoPlayer";
import styles from "./VideoGrid.less";

interface VideoGridProps {
  layout: GridLayout;
  playingChannels: Map<number, ChannelPlayInfo>;
  onChannelPlay: (gridIndex: number, channelInfo: ChannelPlayInfo) => void;
  onStopPlay: (gridIndex: number) => void;
}

interface VideoGridCellProps {
  index: number;
  channelInfo?: ChannelPlayInfo;
  onStopPlay: (index: number) => void;
  onChannelDrop: (index: number, channelInfo: ChannelPlayInfo) => void;
  cellHeight: number;
  cellWidth: number;
}

// 单个网格单元组件
const VideoGridCell: React.FC<VideoGridCellProps> = React.memo(
  ({
    index,
    channelInfo,
    onStopPlay,
    onChannelDrop,
    cellHeight,
    cellWidth,
  }) => {
    const [isDragOver, setIsDragOver] = useState(false);

    // 处理停止播放
    const handleStop = useCallback(() => {
      onStopPlay(index);
    }, [index, onStopPlay]);

    // 处理拖拽事件
    const handleDragOver = useCallback((e: React.DragEvent) => {
      e.preventDefault();
      e.dataTransfer.dropEffect = "copy";
      setIsDragOver(true);
    }, []);

    const handleDragLeave = useCallback((e: React.DragEvent) => {
      e.preventDefault();
      setIsDragOver(false);
    }, []);

    const handleDrop = useCallback(
      (e: React.DragEvent) => {
        e.preventDefault();
        setIsDragOver(false);

        try {
          const channelData = JSON.parse(e.dataTransfer.getData("text/plain"));
          if (channelData.deviceId && channelData.channelId) {
            onChannelDrop(index, channelData);
            message.success(
              `已将 ${
                channelData.channelName || channelData.channelId
              } 拖拽到网格 ${index + 1}`
            );
          }
        } catch (error) {
          console.error("拖拽数据解析失败:", error);
          message.error("拖拽失败");
        }
      },
      [index, onChannelDrop]
    );

    if (channelInfo) {
      return (
        <div
          className={styles.playingCell}
          style={{ height: cellHeight, width: cellWidth }}
        >
          <div className={styles.videoHeader}>
            <div
              className={styles.videoTitle}
              title={channelInfo.channelName || channelInfo.channelId}
            >
              {channelInfo.channelName || channelInfo.channelId}
            </div>
            <Button
              type="text"
              size="small"
              icon={<CloseOutlined />}
              onClick={handleStop}
              className={styles.closeButton}
            />
          </div>

          <div className={styles.playerContainer}>
            <SimplifiedVideoPlayer
              deviceId={channelInfo.deviceId}
              channelId={channelInfo.channelId}
            />
          </div>
        </div>
      );
    }

    // 空白网格
    return (
      <div
        className={classNames(styles.emptyCell, {
          [styles.dragOver]: isDragOver,
        })}
        style={{ height: cellHeight, width: cellWidth }}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
      >
        <PlayCircleOutlined className={styles.playIcon} />
        <span className={styles.gridLabel}>网格 {index + 1}</span>
        <span className={styles.hintText}>
          {isDragOver ? "松开鼠标放置视频" : "拖拽设备通道到此处"}
        </span>
      </div>
    );
  }
);

VideoGridCell.displayName = "VideoGridCell";

const VideoGrid: React.FC<VideoGridProps> = ({
  layout,
  playingChannels,
  onChannelPlay,
  onStopPlay,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [containerHeight, setContainerHeight] = useState<number>(0);
  const [containerWidth, setContainerWidth] = useState<number>(0);
  const recalcHeight = useCallback(() => {
    const el = containerRef.current;
    if (!el) return;
    const rect = el.getBoundingClientRect();
    // 以视口为基准的目标高度，避免因内容高度变化导致的正反馈增长
    let target = Math.floor(window.innerHeight - rect.top - 8); // 预留一点间距
    if (document.fullscreenElement) {
      // 全屏下直接使用视口高度
      target = Math.floor(window.innerHeight);
    }
    const normalized = Math.max(0, target);
    setContainerHeight((prev) => (prev !== normalized ? normalized : prev));
    const width = Math.max(0, Math.floor(el.clientWidth));
    setContainerWidth((prev) => (prev !== width ? width : prev));
  }, []);

  useEffect(() => {
    const containerElement = containerRef.current;
    if (!containerElement) return;

    recalcHeight();

    let resizeObserver: ResizeObserver | null = null;
    if (typeof ResizeObserver !== "undefined") {
      resizeObserver = new ResizeObserver(() => {
        recalcHeight();
      });
      resizeObserver.observe(containerElement);
    }

    const handleFullscreenChange = () => {
      // 延迟到下一帧，确保布局回流完成
      setTimeout(() => {
        recalcHeight();
      }, 0);
    };

    // 无论是否有 ResizeObserver，仍然监听窗口尺寸与全屏变化
    window.addEventListener("resize", recalcHeight);
    document.addEventListener("fullscreenchange", handleFullscreenChange);
    document.addEventListener(
      "webkitfullscreenchange",
      handleFullscreenChange as any
    );
    document.addEventListener(
      "msfullscreenchange",
      handleFullscreenChange as any
    );

    return () => {
      if (resizeObserver) {
        resizeObserver.disconnect();
      }
      window.removeEventListener("resize", recalcHeight);
      document.removeEventListener("fullscreenchange", handleFullscreenChange);
      document.removeEventListener(
        "webkitfullscreenchange",
        handleFullscreenChange as any
      );
      document.removeEventListener(
        "msfullscreenchange",
        handleFullscreenChange as any
      );
    };
  }, [recalcHeight]);
  // 计算网格配置
  const gridConfig = useMemo(() => {
    switch (layout) {
      case 1:
        return { rows: 1, cols: 1 };
      case 4:
        return { rows: 2, cols: 2 };
      case 9:
        return { rows: 3, cols: 3 };
      case 16:
        return { rows: 4, cols: 4 };
      default:
        return { rows: 2, cols: 2 };
    }
  }, [layout]);

  // 计算网格尺寸
  const { cellHeight, cellWidth } = useMemo(() => {
    const verticalGap = 8;
    const horizontalGap = 8;
    const aspectRatio = 16 / 9;

    const availableHeight = Math.max(containerHeight, 0);
    const availableWidth = Math.max(containerWidth, 0);
    if (!availableHeight || !availableWidth) {
      return { cellHeight: 320, cellWidth: 320 * aspectRatio };
    }

    const heightByContainer =
      (availableHeight - (gridConfig.rows - 1) * verticalGap) / gridConfig.rows;
    const widthOfCol =
      (availableWidth - (gridConfig.cols - 1) * horizontalGap) /
      gridConfig.cols;
    const usableWidthInCol = widthOfCol - horizontalGap;

    const isFullscreen = !!(
      document.fullscreenElement ||
      (document as any).webkitFullscreenElement ||
      (document as any).msFullscreenElement
    );
    if (isFullscreen) {
      return { cellHeight: heightByContainer, cellWidth: usableWidthInCol };
    }

    // 非全屏，保持16:9
    const W = usableWidthInCol;
    const H = heightByContainer;
    const heightFromWidth = W / aspectRatio;

    if (heightFromWidth <= H) {
      // 宽度是限制因素
      return { cellHeight: heightFromWidth, cellWidth: W };
    } else {
      // 高度是限制因素
      const widthFromHeight = H * aspectRatio;
      return { cellHeight: H, cellWidth: widthFromHeight };
    }
  }, [containerHeight, containerWidth, gridConfig.rows, gridConfig.cols]);

  // 处理拖拽到网格
  const handleChannelDrop = useCallback(
    (gridIndex: number, channelInfo: ChannelPlayInfo) => {
      onChannelPlay(gridIndex, channelInfo);
    },
    [onChannelPlay]
  );

  // 生成网格
  const renderGrid = useMemo(() => {
    const cells = [];
    for (let i = 0; i < layout; i++) {
      const channelInfo = playingChannels.get(i);
      cells.push(
        <div key={i} className={styles.cellWrapper}>
          <VideoGridCell
            index={i}
            channelInfo={channelInfo}
            onStopPlay={onStopPlay}
            onChannelDrop={handleChannelDrop}
            cellHeight={cellHeight}
            cellWidth={cellWidth}
          />
        </div>
      );
    }
    return cells;
  }, [
    layout,
    playingChannels,
    onStopPlay,
    handleChannelDrop,
    cellHeight,
    cellWidth,
  ]);

  return (
    <div ref={containerRef} className={styles.container}>
      <Flex wrap="wrap" gap={8}>
        {renderGrid}
      </Flex>
    </div>
  );
};

export default VideoGrid;
