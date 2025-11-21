import {
  CameraOutlined,
  CloseOutlined,
  ControlOutlined,
  FullscreenExitOutlined,
  FullscreenOutlined,
  LoadingOutlined,
  MutedOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  SoundOutlined,
} from "@ant-design/icons";
import { message, Slider, Spin, Tooltip } from "antd";
import React, { useEffect, useRef, useState } from "react";
import PTZControl from "./PTZControl";
import "./index.less";

// 声明 Jessibuca 全局类型
declare global {
  interface Window {
    Jessibuca: any;
  }
}

interface VideoPlayerProps {
  deviceId: string;
  channelId: string;
  channelName?: string;
  visible: boolean;
  onClose: () => void;
  mode?: "inline" | "overlay";
  // 可选：直接传入完整的 streamPath（deviceId/channelId）以避免解析差异
  streamPath?: string;
}

const VideoPlayer: React.FC<VideoPlayerProps> = ({
  deviceId,
  channelId,
  channelName,
  visible,
  onClose,
  mode = "overlay",
  streamPath,
}) => {
  const playerRef = useRef<any>(null);
  const [playing, setPlaying] = useState(true);
  const [volume, setVolume] = useState(50);
  const [isMuted, setIsMuted] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isClosing, setIsClosing] = useState(false);
  const [showPTZControl, setShowPTZControl] = useState(false);

  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  // 优先使用传入的 streamPath，否则拼接 deviceId/channelId
  const resolvedStreamPath = React.useMemo(() => {
    return streamPath && streamPath.length > 0
      ? streamPath
      : `${deviceId}/${channelId}`;
  }, [streamPath, deviceId, channelId]);

  const videoUrl = `${protocol}//${window.location.host}/flv/${resolvedStreamPath}.flv`;
  const safeContainerId = React.useMemo(
    () =>
      `jessibuca-container-${resolvedStreamPath.replace(
        /[^a-zA-Z0-9_-]/g,
        "_"
      )}`,
    [resolvedStreamPath]
  );

  // 使用绝对路径引用 public 目录下的解码器文件 - 避免每次渲染都创建新的 URL
  const decoderPath = React.useMemo(
    () => new URL("/jessibuca/decoder.js", window.location.origin).toString(),
    []
  );

  // 稳定的播放器配置，避免重复渲染
  const _playerConfig = React.useMemo(
    () => ({
      videoBuffer: 0.2,
      useMSE: false,
      useWCS: false,
      hasAudio: true,
      autoWasm: true,
      forceNoOffscreen: true,
      isNotMute: true,
      loadingText: "正在加载视频...",
      debug: false, // 关闭调试信息
      timeout: 10000, // 10秒超时
    }),
    []
  );

  useEffect(() => {
    if (visible) {
      setIsLoading(true);
      setError(null);
      setPlaying(true);

      // 初始化原生 Jessibuca 播放器
      const initPlayer = async () => {
        try {
          // 动态加载本地 Jessibuca 库
          if (typeof window.Jessibuca === "undefined") {
            try {
              // 加载本地 Jessibuca 库
              await new Promise((resolve, reject) => {
                const script = document.createElement("script");
                script.src = "/jessibuca/jessibuca.js";
                script.onload = () => {
                  // 等待一下确保 window.Jessibuca 被正确设置
                  setTimeout(() => {
                    if (typeof window.Jessibuca !== "undefined") {
                      resolve(true);
                    } else {
                      console.error(
                        "❌ 本地版本加载后 window.Jessibuca 仍未定义"
                      );
                      reject(new Error("Jessibuca 未正确初始化"));
                    }
                  }, 500);
                };
                script.onerror = (error) => {
                  console.error("❌ 本地 jessibuca.js 加载失败:", error);
                  reject(error);
                };
                document.head.appendChild(script);
              });
            } catch (loadError) {
              console.error("❌ 加载本地 Jessibuca 库失败:", loadError);
              setError("播放器库加载失败");
              setIsLoading(false);
              return;
            }
          } else {
          }

          // 使用 streamPath 作为容器 ID 的一部分，避免不同页面/模式下的冲突
          const container = document.getElementById(safeContainerId);

          if (!container) {
            console.error("❌ 播放器容器未找到:", safeContainerId);
            setError("播放器容器未找到");
            setIsLoading(false);
            return;
          }

          // 检查容器尺寸
          const rect = container.getBoundingClientRect();

          if (rect.width === 0 || rect.height === 0) {
            console.warn("⚠️ 容器尺寸为0，可能导致视频无法显示");
          }

          const jessibuca = new window.Jessibuca({
            container: container,
            decoder: decoderPath,
            videoBuffer: 0.2,
            isResize: true,
            useMSE: false, // 改为 false，与 playerConfig 保持一致
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

          playerRef.current = jessibuca;

          // 绑定事件
          jessibuca.on("load", () => {
            setIsLoading(false);
          });

          jessibuca.on("play", () => {
            setIsLoading(false);
            setError(null);
            setPlaying(true);
          });

          jessibuca.on("pause", () => {
            setPlaying(false);
          });

          jessibuca.on("error", (error: any) => {
            console.error("❌ 播放器错误:", error);
            setError(`播放错误: ${error.message || error}`);
            setIsLoading(false);
          });

          jessibuca.on("timeout", () => {
            console.warn("⏰ 连接超时");
            setError("连接超时，请检查网络");
            setIsLoading(false);
          });

          // 开始播放
          jessibuca.play(videoUrl);
        } catch (error) {
          console.error("❌ 播放器初始化失败:", error);
          setError("播放器初始化失败");
          setIsLoading(false);
        }
      };

      // 延迟初始化确保 DOM 已渲染
      setTimeout(initPlayer, 100);
    }

    return () => {
      // 清理函数，组件卸载或 visible 变为 false 时执行
      if (playerRef.current?.destroy) {
        try {
          playerRef.current.destroy();
        } catch (error) {
          console.warn("播放器销毁失败:", error);
        }
        playerRef.current = null;
      }
    };
  }, [visible, videoUrl, decoderPath, safeContainerId]);

  // 当外部将 visible 从 true 切换为 false 时，做关闭过渡动画
  const prevVisibleRef = useRef<boolean>(visible);
  useEffect(() => {
    if (prevVisibleRef.current && !visible) {
      setIsClosing(true);
      const timer = setTimeout(() => setIsClosing(false), 200);
      prevVisibleRef.current = visible;
      return () => {
        clearTimeout(timer);
      };
    }
    prevVisibleRef.current = visible;
    return undefined;
  }, [visible]);

  // 超时处理：如果 15 秒后仍在加载，显示错误
  useEffect(() => {
    if (!isLoading) {
      return;
    }

    const timer = setTimeout(() => {
      if (isLoading) {
        console.warn("播放器加载超时 - 15秒内未收到 onLoad 事件");
        setIsLoading(false);
        setError("加载超时，请重试");
      }
    }, 15000);

    return () => {
      clearTimeout(timer);
    };
  }, [isLoading, error, playing, videoUrl, decoderPath]);

  const handlePlayPause = () => {
    const newPlaying = !playing;
    setPlaying(newPlaying);

    // 通过 ref 直接控制播放器
    if (playerRef.current) {
      if (newPlaying) {
        playerRef.current.play?.();
      } else {
        playerRef.current.pause?.();
      }
    }
  };

  const handleVolumeChange = (value: number) => {
    setVolume(value);

    // 如果当前是静音状态且移动了滑块，自动解除静音
    if (isMuted && value > 0) {
      setIsMuted(false);
      // 解除播放器的静音状态
      if (playerRef.current?.cancelMute) {
        playerRef.current.cancelMute();
      }
    }

    // 非静音状态时设置播放器音量
    if (!isMuted || (isMuted && value > 0)) {
      if (playerRef.current?.setVolume) {
        playerRef.current.setVolume(value / 100);
      }
    }
  };

  const handleMuteToggle = () => {
    const newMuted = !isMuted;
    setIsMuted(newMuted);

    // 直接设置播放器静音状态
    if (playerRef.current) {
      if (newMuted) {
        playerRef.current.mute?.();
      } else {
        playerRef.current.cancelMute?.();
        // 解除静音时恢复之前保存的音量
        if (playerRef.current.setVolume) {
          playerRef.current.setVolume(volume / 100);
        }
      }
    }
  };

  const handleScreenshot = () => {
    if (playerRef.current?.screenshot) {
      try {
        playerRef.current.screenshot(
          `${channelName || deviceId}_${Date.now()}.png`,
          "image/png"
        );
        message.success("截图成功！");
      } catch (error) {
        console.error("Screenshot error:", error);
        message.error("截图失败");
      }
    } else {
      message.error("截图功能不可用");
    }
  };

  const handleFullscreenToggle = () => {
    const newFullscreen = !isFullscreen;
    setIsFullscreen(newFullscreen);

    // 直接控制播放器全屏
    if (playerRef.current) {
      if (newFullscreen) {
        playerRef.current.setFullscreen?.(true);
      } else {
        playerRef.current.setFullscreen?.(false);
      }
    }
  };

  if (!visible && !isClosing) {
    return null;
  }

  const renderPlayer = () => (
    <div
      className={`video-player-container ${mode} ${
        isFullscreen ? "fullscreen-mode" : ""
      }`}
    >
      <div className="player-header">
        <span className="channel-name">{channelName || "实时视频"}</span>
        <CloseOutlined
          className="close-icon"
          onClick={() => {
            // 播放器内部先做关闭动画，再通知外部真正关闭
            setIsClosing(true);
            setTimeout(() => {
              onClose();
            }, 200);
          }}
        />
      </div>
      <div className="player-body">
        {error ? (
          <div className="player-error">
            <p>视频加载失败</p>
            <span>{error}</span>
          </div>
        ) : (
          <>
            <div
              id={safeContainerId}
              style={{
                width: "100%",
                height: "100%",
                background: "#000",
                position: "relative",
                zIndex: 1,
              }}
            />
            {isLoading && (
              <div className="player-loading" style={{ zIndex: 2 }}>
                <Spin
                  indicator={<LoadingOutlined style={{ fontSize: 32 }} spin />}
                />
                <p>加载中...</p>
              </div>
            )}
          </>
        )}
      </div>
      <div className="player-controls">
        <div className="control-left">
          <Tooltip title={playing ? "暂停" : "播放"}>
            <div onClick={handlePlayPause} className="control-icon">
              {playing ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
            </div>
          </Tooltip>
        </div>
        <div className="control-right">
          <Tooltip title="云台控制">
            <div
              onClick={() => setShowPTZControl(!showPTZControl)}
              className={`control-icon ${showPTZControl ? "active" : ""}`}
            >
              <ControlOutlined />
            </div>
          </Tooltip>
          <Tooltip title="截图">
            <div onClick={handleScreenshot} className="control-icon">
              <CameraOutlined />
            </div>
          </Tooltip>
          <Tooltip title={isMuted ? "取消静音" : "静音"}>
            <div
              onClick={handleMuteToggle}
              className={`control-icon ${isMuted ? "muted" : ""}`}
            >
              {isMuted ? <MutedOutlined /> : <SoundOutlined />}
            </div>
          </Tooltip>
          <div className="volume-slider-container">
            <Slider
              min={0}
              max={100}
              value={volume}
              onChange={handleVolumeChange}
              tooltip={{ formatter: (value) => `${value}%` }}
            />
          </div>
          <Tooltip title={isFullscreen ? "退出全屏" : "全屏"}>
            <div onClick={handleFullscreenToggle} className="control-icon">
              {isFullscreen ? (
                <FullscreenExitOutlined />
              ) : (
                <FullscreenOutlined />
              )}
            </div>
          </Tooltip>
        </div>
      </div>

      {/* 云台控制组件 */}
      {showPTZControl && (
        <div className="ptz-control-wrapper">
          <PTZControl
            deviceId={deviceId}
            channelId={channelId}
            visible={showPTZControl}
            onClose={() => setShowPTZControl(false)}
          />
        </div>
      )}
    </div>
  );

  if (mode === "overlay") {
    return (
      <div className={`video-player-overlay${isClosing ? " closing" : ""}`}>
        {renderPlayer()}
      </div>
    );
  }

  return renderPlayer();
};

export default VideoPlayer;
