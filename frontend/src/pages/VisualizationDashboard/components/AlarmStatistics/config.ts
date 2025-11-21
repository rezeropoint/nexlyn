import type { AlarmLevel } from "./types";

// 告警阈值配置
export const ALARM_THRESHOLDS = {
  temperature: { serious: 35, important: 30, normal: 28 },
  humidity: { serious: 90, important: 80, normal: 70 },
  smoke: { serious: 200, important: 100, normal: 50 },
};

/**
 * 判断告警级别
 * @param fieldName 字段名称
 * @param value 字段值
 * @returns 告警级别，如果不超过阈值则返回null
 */
export const getAlarmLevel = (
  fieldName: string,
  value: number
): AlarmLevel | null => {
  const thresholds =
    ALARM_THRESHOLDS[fieldName as keyof typeof ALARM_THRESHOLDS];
  if (!thresholds) return null;

  if (value >= thresholds.serious) return "严重";
  if (value >= thresholds.important) return "重要";
  if (value >= thresholds.normal) return "一般";
  return null;
};
