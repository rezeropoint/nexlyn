import { LoadingOutlined, PlayCircleOutlined } from "@ant-design/icons";
import { Alert, message, Spin } from "antd";
import dayjs from "dayjs";
import React, { useCallback, useEffect, useRef, useState } from "react";
import type { ChannelInfo, RecordingItem } from "../index";
import styles from "./RecordingPlayer.less";

// 声明 Jessibuca 全局类型
declare global {
  interface Window {
    Jessibuca: any;
  }
}

interface RecordingPlayerProps {
  selectedChannel: ChannelInfo | null;
  playingRecording: RecordingItem | null;
  customStartTime?: number; // 自定义开始时间（毫秒时间戳）
}

const RecordingPlayer: React.FC<RecordingPlayerProps> = ({
  selectedChannel,
  playingRecording,
  customStartTime,
}) => {
  const playerRef = useRef<any>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 生成播放URL
  const generatePlayUrl = useCallback(
    (recording: RecordingItem, channel: ChannelInfo, customStart?: number) => {
      // 使用自定义开始时间（如果提供），否则使用录像文件的开始时间
      const actualStartTime = customStart
        ? Math.floor(customStart / 1000)
        : Math.floor(dayjs(recording.startTime).valueOf() / 1000);
      const endTime = Math.floor(dayjs(recording.endTime).valueOf() / 1000);
      const timestamp = Date.now();

      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      const host = window.location.host; // 使用当前页面的host，本地开发时会通过代理转发
      return `${protocol}//${host}/flv/gb_${timestamp}/${channel.deviceId}/${channel.channelId}?start=${actualStartTime}&end=${endTime}`;
    },
    []
  );

  const containerId = React.useMemo(() => {
    if (!playingRecording || !selectedChannel) return "";
    return `recording-player-${selectedChannel.deviceId}-${
      selectedChannel.channelId
    }-${playingRecording.startTime.replace(
      /[^a-zA-Z0-9_-]/g,
      "_"
    )}-${Date.now()}`;
  }, [playingRecording, selectedChannel]);

  // 初始化 Jessibuca 播放器
  const initPlayer = useCallback(async () => {
    if (!playingRecording || !selectedChannel) return;

    let mounted = true;

    try {
      setIsLoading(true);
      setError(null);

      // 确保 Jessibuca 已加载
      if (typeof window.Jessibuca === "undefined") {
        await new Promise((resolve, reject) => {
          const script = document.createElement("script");
          script.src = "/jessibuca/jessibuca.js";
          script.onload = () => {
            setTimeout(() => {
              if (typeof window.Jessibuca !== "undefined") {
                resolve(true);
              } else {
                reject(new Error("Jessibuca 未正确初始化"));
              }
            }, 500);
          };
          script.onerror = reject;
          document.head.appendChild(script);
        });
      }

      if (!mounted) return;

      const container = containerRef.current;
      if (!container) {
        setError("播放器容器未找到");
        setIsLoading(false);
        return;
      }

      // 清空容器
      container.innerHTML = "";

      // 创建播放器容器
      const playerContainer = document.createElement("div");
      playerContainer.id = containerId;
      playerContainer.style.width = "100%";
      playerContainer.style.height = "100%";
      container.appendChild(playerContainer);

      // 使用绝对路径引用decoder文件
      const decoderPath = new URL(
        "/jessibuca/decoder.js",
        window.location.origin
      ).toString();

      const jessibuca = new window.Jessibuca({
        container: playerContainer,
        decoder: decoderPath,
        videoBuffer: 0.2,
        isResize: true, // 启用自适应大小
        useMSE: false, // 与实时播放保持一致
        useWCS: false,
        hasAudio: true,
        autoWasm: true,
        forceNoOffscreen: true,
        isFlv: true, // 添加FLV格式支持
        debug: false, // 关闭调试日志
        timeout: 15000,
        heartTimeout: 30000,
        loadingText: "正在加载录像...",
        showBandwidth: false,
        operateBtns: {
          fullscreen: true,
          screenshot: true,
          play: true,
          audio: true,
        },
      });

      if (!mounted) {
        jessibuca.destroy();
        return;
      }

      playerRef.current = jessibuca;

      // 绑定事件
      jessibuca.on("load", () => {
        if (mounted) {
          setIsLoading(false);
          setError(null);
        }
      });

      jessibuca.on("play", () => {
        if (mounted) {
          setIsLoading(false);
          setError(null);
          message.success("开始播放录像");
        }
      });

      jessibuca.on("error", (error: any) => {
        console.error("录像播放器错误:", error);
        if (mounted) {
          setError("播放失败");
          setIsLoading(false);
        }
      });

      jessibuca.on("timeout", () => {
        console.warn("录像连接超时");
        if (mounted) {
          setError("连接超时");
          setIsLoading(false);
        }
      });

      jessibuca.on("ended", () => {
        if (mounted) {
          message.info("录像播放完成");
        }
      });

      // 开始播放
      const playUrl = generatePlayUrl(
        playingRecording,
        selectedChannel,
        customStartTime
      );
      console.log("录像播放URL:", playUrl);
      jessibuca.play(playUrl);
    } catch (error) {
      console.error("录像播放器初始化失败:", error);
      if (mounted) {
        setError("初始化失败");
        setIsLoading(false);
      }
    }

    return () => {
      mounted = false;
    };
  }, [
    playingRecording,
    selectedChannel,
    customStartTime,
    generatePlayUrl,
    containerId,
  ]);

  // 当播放录像变化时初始化播放器
  useEffect(() => {
    let mounted = true;

    if (playingRecording && selectedChannel) {
      // 延迟初始化确保 DOM 已渲染
      setTimeout(() => {
        if (mounted) {
          initPlayer();
        }
      }, 100);
    }

    return () => {
      mounted = false;
      // 组件卸载时清理播放器
      if (playerRef.current?.destroy) {
        try {
          playerRef.current.destroy();
        } catch (error) {
          console.warn("播放器销毁失败:", error);
        }
        playerRef.current = null;
      }
    };
  }, [playingRecording, selectedChannel, customStartTime, initPlayer]);

  // 超时处理
  useEffect(() => {
    if (!isLoading || error) {
      return;
    }

    const timer = setTimeout(() => {
      if (isLoading) {
        setIsLoading(false);
        setError("加载超时");
      }
    }, 15000);

    return () => {
      clearTimeout(timer);
    };
  }, [isLoading, error]);

  return (
    <div className={styles.playerContainer}>
      {/* 录像信息 */}
      {playingRecording && (
        <div className={styles.recordingInfo}>
          <div>
            <strong>设备:</strong> {playingRecording.name}
          </div>
          {customStartTime ? (
            <div>
              <div>
                <strong>播放时间:</strong>{" "}
                {dayjs(customStartTime).format("YYYY-MM-DD HH:mm:ss")} -{" "}
                {dayjs(playingRecording.endTime).format("HH:mm:ss")}
              </div>
              <div>
                <strong>录像时间:</strong>{" "}
                {dayjs(playingRecording.startTime).format(
                  "YYYY-MM-DD HH:mm:ss"
                )}{" "}
                - {dayjs(playingRecording.endTime).format("HH:mm:ss")}
              </div>
            </div>
          ) : (
            <div>
              <strong>时间:</strong>{" "}
              {dayjs(playingRecording.startTime).format("YYYY-MM-DD HH:mm:ss")}{" "}
              - {dayjs(playingRecording.endTime).format("HH:mm:ss")}
            </div>
          )}
        </div>
      )}

      {/* 错误提示 */}
      {error && (
        <Alert
          message="播放错误"
          description={error}
          type="error"
          closable
          onClose={() => setError(null)}
          className={styles.alertSpacing}
        />
      )}

      {/* 视频播放区域 */}
      <div className={styles.videoArea}>
        <div ref={containerRef} className={styles.playerWrapper} />

        {/* 当没有录像时显示提示 */}
        {!playingRecording && (
          <div className={styles.emptyPlaceholder}>
            <PlayCircleOutlined className={styles.emptyIcon} />
            <div>请选择要播放的录像</div>
          </div>
        )}

        {isLoading && (
          <div className={styles.loadingOverlay}>
            <Spin
              indicator={
                <LoadingOutlined className={styles.loadingIcon} spin />
              }
            />
            <div className={styles.loadingText}>正在加载录像...</div>
          </div>
        )}

        {error && (
          <div className={styles.errorOverlay}>
            <div className={styles.errorTitle}>录像加载失败</div>
            <div className={styles.errorDetail}>{error}</div>
          </div>
        )}
      </div>
    </div>
  );
};

export default RecordingPlayer;
