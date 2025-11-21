// 告警统计相关类型定义

export type TimeRange = "today" | "week" | "month";
export type AlarmLevel = "严重" | "重要" | "一般" | "轻微";

// 告警记录
export interface AlarmRecord {
  id: string;
  deviceId: string;
  deviceName: string;
  alarmType: string;
  level: AlarmLevel;
  message: string;
  location: string;
  timestamp: string;
}

// 告警趋势
export interface AlarmTrend {
  date: string;
  count: number;
}

// 告警统计
export interface AlarmStats {
  total: number;
  serious: number;
  important: number;
  normal: number;
  minor: number;
}

// 组件Props
export interface AlarmStatisticsProps {
  className?: string;
  orgId?: string; // 组织ID（从父组件传入）
}
