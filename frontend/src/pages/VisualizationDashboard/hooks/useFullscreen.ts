import type { MessageInstance } from "antd/es/message/interface";
import { useEffect, useState } from "react";

/**
 * 全屏功能hook
 * @param message message 实例
 */
export const useFullscreen = (message: MessageInstance) => {
  const [isFullscreen, setIsFullscreen] = useState(false);

  // 全屏切换
  const toggleFullscreen = async () => {
    try {
      if (!isFullscreen) {
        // 进入全屏
        const docElement = document.documentElement as HTMLElement & {
          webkitRequestFullscreen?: () => Promise<void>;
          msRequestFullscreen?: () => Promise<void>;
        };
        if (docElement.requestFullscreen) {
          await docElement.requestFullscreen();
        } else if (docElement.webkitRequestFullscreen) {
          await docElement.webkitRequestFullscreen();
        } else if (docElement.msRequestFullscreen) {
          await docElement.msRequestFullscreen();
        }
      } else {
        // 退出全屏
        const doc = document as Document & {
          webkitExitFullscreen?: () => Promise<void>;
          msExitFullscreen?: () => Promise<void>;
        };
        if (doc.exitFullscreen) {
          await doc.exitFullscreen();
        } else if (doc.webkitExitFullscreen) {
          await doc.webkitExitFullscreen();
        } else if (doc.msExitFullscreen) {
          await doc.msExitFullscreen();
        }
      }
    } catch (_error) {
      message.error("全屏切换失败");
    }
  };

  // 监听全屏状态变化
  useEffect(() => {
    const handleFullscreenChange = () => {
      const doc = document as Document & {
        webkitFullscreenElement?: Element;
        msFullscreenElement?: Element;
      };
      const fullscreenElement =
        document.fullscreenElement ||
        doc.webkitFullscreenElement ||
        doc.msFullscreenElement;
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

  return {
    isFullscreen,
    toggleFullscreen,
  };
};
