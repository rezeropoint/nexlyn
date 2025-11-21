/**
 * 设备管理常量定义
 */

// 表格配置
export const TABLE_CONFIG = {
  DEFAULT_PAGE_SIZE: 10,
  SCROLL_X: 1840,
};

// 设备状态映射
export const DEVICE_STATUS_MAP = {
  active: { text: "正常", color: "success" },
  inactive: { text: "停用", color: "default" },
  maintenance: { text: "维护中", color: "warning" },
  error: { text: "故障", color: "error" },
  decommissioned: { text: "已退役", color: "default" },
};

// 设备状态选项
export const DEVICE_STATUS_OPTIONS = Object.entries(DEVICE_STATUS_MAP).map(
  ([value, { text }]) => ({
    label: text,
    value,
  })
);

// 在线状态选项
export const ONLINE_STATUS_OPTIONS = [
  { label: "在线", value: "true" },
  { label: "离线", value: "false" },
];
