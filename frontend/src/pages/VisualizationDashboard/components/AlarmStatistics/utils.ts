import {
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  WarningOutlined,
} from "@ant-design/icons";
import dayjs from "dayjs";
import React from "react";

/**
 * 获取告警图标
 * @param level 告警级别
 */
export const getAlarmIcon = (level: string): React.ReactElement => {
  switch (level) {
    case "严重":
      return React.createElement(ExclamationCircleOutlined, {
        style: { color: "#ff4d4f" },
      });
    case "重要":
      return React.createElement(WarningOutlined, {
        style: { color: "#fa8c16" },
      });
    default:
      return React.createElement(ClockCircleOutlined, {
        style: { color: "#fadb14" },
      });
  }
};

/**
 * 格式化时间显示
 * @param timestamp 时间戳
 */
export const formatTimeDisplay = (timestamp: string): string => {
  const time = dayjs(timestamp);
  if (time.isSame(dayjs(), "day")) {
    return time.format("HH:mm");
  }
  return time.format("MM-DD HH:mm");
};
