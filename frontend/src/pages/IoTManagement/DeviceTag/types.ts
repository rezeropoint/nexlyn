/**
 * 设备标签管理类型定义
 */

import type { DeviceTag, DeviceTagSummary } from "@/services/iot";

// 模态框状态
export interface ModalState {
  create: boolean;
  edit: boolean;
}

// 选择状态
export interface SelectionState {
  selectedRowKeys: React.Key[];
  selectedRows: DeviceTagSummary[];
}

// 导出类型以供组件使用
export type { DeviceTag, DeviceTagSummary };
