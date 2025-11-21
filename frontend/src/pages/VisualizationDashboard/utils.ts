// 设备和告警相关的工具函数

// 获取设备状态颜色
export const getDeviceStatusColor = (status: string): string => {
  switch (status) {
    case "在线":
      return "#52c41a";
    case "离线":
      return "#d9d9d9";
    case "告警":
      return "#faad14";
    case "故障":
      return "#ff4d4f";
    default:
      return "#d9d9d9";
  }
};

// 获取告警级别颜色
export const getAlarmLevelColor = (level: string): string => {
  switch (level) {
    case "严重":
      return "#ff4d4f";
    case "重要":
      return "#fa8c16";
    case "一般":
      return "#fadb14";
    case "轻微":
      return "#52c41a";
    default:
      return "#d9d9d9";
  }
};

// 获取设备类型图标
export const getDeviceTypeIcon = (type: string): string => {
  switch (type) {
    case "监控设备":
      return "📹";
    case "门禁设备":
      return "🚪";
    case "环境传感器":
    case "temperature_sensor": // 支持后端类型
      return "🌡️";
    case "烟感探测器":
      return "💨";
    case "红外探测器":
      return "🔴";
    case "网络设备":
      return "🌐";
    default:
      return "📟";
  }
};
