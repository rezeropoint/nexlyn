/**
 * 时序数据相关类型
 */

// 时序数据项
export interface TimeSeriesDataItem {
  timestamp: string; // 数据时间戳
  deviceId: string; // 设备唯一标识符
  deviceModel: string; // 设备型号
  deviceCategory: string; // 设备类别
  fieldName: string; // 字段名称
  value: any; // 字段值
}

// 字段值和时间戳
export interface FieldValue {
  value: any; // 字段值
  timestamp: string; // 时间戳
}

// 设备最新值
export interface DeviceLatestValuesItem {
  deviceId: string; // 设备ID
  fields: Record<string, FieldValue>; // 字段名->字段值映射
}

// 字段统计信息
export interface FieldStatItem {
  min: number; // 最小值
  max: number; // 最大值
  avg: number; // 平均值
  sum: number; // 总和
  count: number; // 数据点数量
  lastValue: any; // 最新值
  lastTime: string; // 最新值时间戳
}

// 设备统计数据
export interface DeviceStatisticsItem {
  deviceId: string; // 设备ID
  deviceModel: string; // 设备型号
  fields: Record<string, FieldStatItem>; // 字段名->统计信息映射
}
