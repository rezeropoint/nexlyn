/**
 * 平台管理相关常量定义
 */

// 平台类型选项 (当前仅支持Skylark)
export const PLATFORM_TYPE_OPTIONS = [
  { label: "Skylark流程引擎", value: "skylark" },
];

// 平台类型标签颜色
export const PLATFORM_TYPE_COLORS: Record<string, string> = {
  skylark: "blue",
};

// 平台类型显示名称
export const PLATFORM_TYPE_NAMES: Record<string, string> = {
  skylark: "Skylark流程引擎",
};

// 启用状态选项
export const ENABLED_STATUS_OPTIONS = [
  { label: "全部", value: "" },
  { label: "已启用", value: "true" },
  { label: "已禁用", value: "false" },
];

// HTTP方法选项
export const HTTP_METHOD_OPTIONS = [
  { label: "POST", value: "POST" },
  { label: "PUT", value: "PUT" },
  { label: "PATCH", value: "PATCH" },
  { label: "GET", value: "GET" },
];

// 分发类型选项
export const DISPATCH_TYPE_OPTIONS = [
  { label: "Skylark流程触发", value: "skylark_flows", platformType: "skylark" },
  { label: "Skylark表单提交", value: "skylark_forms", platformType: "skylark" },
  { label: "LynxGraph逻辑引擎", value: "lynxgraph", platformType: null },
  { label: "日志记录", value: "log", platformType: null },
];

// 分发类型显示名称
export const DISPATCH_TYPE_NAMES: Record<string, string> = {
  skylark_flows: "Skylark流程触发",
  skylark_forms: "Skylark表单提交",
  lynxgraph: "LynxGraph逻辑引擎",
  log: "日志记录",
};

// 分发类型标签颜色
export const DISPATCH_TYPE_COLORS: Record<string, string> = {
  skylark_flows: "blue",
  skylark_forms: "cyan",
  lynxgraph: "purple",
  log: "default",
};

// 每页显示数量
export const DEFAULT_PAGE_SIZE = 10;
export const PAGE_SIZE_OPTIONS = ["10", "20", "50", "100"];
