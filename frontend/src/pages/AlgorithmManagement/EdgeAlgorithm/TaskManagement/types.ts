/**
 * 边缘算法管理页面专用类型定义
 */

import type { DeviceBindingSummary } from "@/services/iot";

/**
 * AI Box 设备选择器选项
 */
export interface AIBoxDeviceOption {
  value: string; // 设备ID
  label: string; // 设备显示名称（包含在线状态）
  deviceId: string; // 设备ID
  deviceName: string; // 设备名称
  deviceAlias: string; // 设备别名
  isOnline: boolean; // 在线状态
  location?: string; // 安装位置
}

/**
 * 任务状态Badge类型映射
 */
export const TASK_STATUS_BADGE_MAP: Record<
  string,
  {
    status: "success" | "processing" | "warning" | "error" | "default";
    text: string;
  }
> = {
  success: { status: "success", text: "运行中" },
  normal: { status: "processing", text: "正常" },
  warning: { status: "warning", text: "告警" },
  danger: { status: "error", text: "异常" },
};

/**
 * 转换设备绑定信息为选择器选项
 */
export function convertDeviceToOption(
  device: DeviceBindingSummary
): AIBoxDeviceOption {
  const onlineStatus = device.isOnline ? "🟢 在线" : "⚫ 离线";
  const displayName = device.deviceAlias || device.deviceName;

  return {
    value: device.deviceId,
    label: `${displayName} (${onlineStatus})`,
    deviceId: device.deviceId,
    deviceName: device.deviceName,
    deviceAlias: device.deviceAlias,
    isOnline: device.isOnline,
    location: device.location,
  };
}
