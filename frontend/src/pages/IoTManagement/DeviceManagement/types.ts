/**
 * 设备管理类型定义
 */

import type {
  DeviceBinding,
  DeviceBindingSummary,
  UnboundDevice,
} from "@/services/iot";

// 模态框状态
export interface ModalState {
  bind: boolean;
  edit: boolean;
  detail: boolean;
}

// 选择状态
export interface SelectionState {
  selectedRowKeys: React.Key[];
  selectedRows: DeviceBindingSummary[];
}

// 导出类型以供组件使用
export type { DeviceBinding, DeviceBindingSummary, UnboundDevice };
