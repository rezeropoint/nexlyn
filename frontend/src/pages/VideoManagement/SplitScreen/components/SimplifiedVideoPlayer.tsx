import { LoadingOutlined } from "@ant-design/icons";
import { Spin } from "antd";
import React, { useEffect, useRef, useState } from "react";
import styles from "./SimplifiedVideoPlayer.less";

// 声明 Jessibuca 全局类型
declare global {
  interface Window {
    Jessibuca: any;
  }
}

interface SimplifiedVideoPlayerProps {
  deviceId: string;
  channelId: string;
  streamPath?: string;
}

const SimplifiedVideoPlayer: React.FC<SimplifiedVideoPlayerProps> = ({
  deviceId,
  channelId,
  streamPath,
}) => {
  const playerRef = useRef<any>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const resolvedStreamPath = React.useMemo(() => {
    return streamPath && streamPath.length > 0
      ? streamPath
      : `${deviceId}/${channelId}`;
  }, [streamPath, deviceId, channelId]);

  const videoUrl = `${protocol}//${window.location.host}/flv/${resolvedStreamPath}.flv`;
  const containerId = React.useMemo(
    () =>
      `simplified-player-${resolvedStreamPath.replace(
        /[^a-zA-Z0-9_-]/g,
        "_"
      )}-${Date.now()}`,
    [resolvedStreamPath]
  );

  useEffect(() => {
    let mounted = true;

    const initPlayer = async () => {
      if (!mounted) return;

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

        const jessibuca = new window.Jessibuca({
          container: playerContainer,
          decoder: new URL(
            "/jessibuca/decoder.js",
            window.location.origin
          ).toString(),
          videoBuffer: 0.2,
          isResize: true,
          useMSE: false,
          useWCS: false,
          hasAudio: true,
          autoWasm: true,
          forceNoOffscreen: true,
          debug: true,
          timeout: 10000,
          heartTimeout: 30000,
          loadingText: "正在加载视频...",
          showBandwidth: false,
          operateBtns: {
            fullscreen: false,
            screenshot: false,
            play: false,
            audio: false,
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
          }
        });

        jessibuca.on("error", (error: any) => {
          console.error("播放器错误:", error);
          if (mounted) {
            setError("播放失败");
            setIsLoading(false);
          }
        });

        jessibuca.on("timeout", () => {
          console.warn("连接超时");
          if (mounted) {
            setError("连接超时");
            setIsLoading(false);
          }
        });

        // 开始播放
        jessibuca.play(videoUrl);
      } catch (error) {
        console.error("播放器初始化失败:", error);
        if (mounted) {
          setError("初始化失败");
          setIsLoading(false);
        }
      }
    };

    // 延迟初始化确保 DOM 已渲染
    setTimeout(initPlayer, 100);

    return () => {
      mounted = false;
      if (playerRef.current?.destroy) {
        try {
          playerRef.current.destroy();
        } catch (error) {
          console.warn("播放器销毁失败:", error);
        }
        playerRef.current = null;
      }
    };
  }, [videoUrl, containerId]);

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

  if (error) {
    return (
      <div className={styles.errorContainer}>
        <div className={styles.errorTitle}>视频加载失败</div>
        <div className={styles.errorDetail}>{error}</div>
      </div>
    );
  }

  return (
    <div className={styles.playerContainer}>
      <div ref={containerRef} className={styles.playerWrapper} />
      {isLoading && (
        <div className={styles.loadingOverlay}>
          <Spin
            indicator={<LoadingOutlined className={styles.loadingIcon} spin />}
          />
          <div className={styles.loadingText}>加载中...</div>
        </div>
      )}
    </div>
  );
};

export default SimplifiedVideoPlayer;
