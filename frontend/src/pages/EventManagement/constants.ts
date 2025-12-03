/**
 * 事件管理相关常量定义
 */

// ===== 表格配置 =====
export const TABLE_CONFIG = {
  DEFAULT_PAGE_SIZE: 20,
  PAGE_SIZE_OPTIONS: ["10", "20", "50", "100"],
};

// ===== 字段类型映射 =====
export const FIELD_TYPE_OPTIONS = [
  { label: "文本", value: "string" },
  { label: "数字", value: "number" },
  { label: "日期", value: "date" },
  { label: "日期时间", value: "datetime" },
  { label: "布尔值", value: "boolean" },
  { label: "图片(Base64)", value: "imageBase64" },
];

export const FIELD_TYPE_MAP: Record<string, { text: string; color: string }> = {
  string: { text: "文本", color: "blue" },
  number: { text: "数字", color: "green" },
  date: { text: "日期", color: "orange" },
  datetime: { text: "日期时间", color: "purple" },
  boolean: { text: "布尔值", color: "cyan" },
  imageBase64: { text: "图片(Base64)", color: "magenta" },
};

// ===== 流程状态映射 =====
export const FLOW_STATUS_MAP: Record<
  string,
  { text: string; color: string; status: string }
> = {
  processing: { text: "处理中", color: "blue", status: "Processing" },
  completed: { text: "已完成", color: "green", status: "Success" },
  rejected: { text: "已拒绝", color: "red", status: "Error" },
  pending: { text: "待处理", color: "gold", status: "Warning" },
};

export const FLOW_STATUS_OPTIONS = Object.entries(FLOW_STATUS_MAP).map(
  ([key, { text }]) => ({
    label: text,
    value: key,
  })
);

// ===== 启用状态选项 =====
export const ENABLED_OPTIONS = [
  { label: "启用", value: true },
  { label: "禁用", value: false },
];

// ===== API 端点配置 =====
export const API_PREFIX = "/api/v1";

export const API_ENDPOINTS = {
  // 平台配置
  PLATFORM_CONFIG: `${API_PREFIX}/skylark-platform/config`,
  TEST_CONNECTION: `${API_PREFIX}/skylark-platform/test-connection`,
  FLOW_LIST: `${API_PREFIX}/skylark-platform/flows`,
  FLOW_FIELDS: (flowId: number) =>
    `${API_PREFIX}/skylark-platform/flows/${flowId}/fields`,

  // 事件配置
  EVENT_CONFIGS: `${API_PREFIX}/event-configs`,
  EVENT_CONFIG_LIST: `${API_PREFIX}/event-configs/list`,
  EVENT_CONFIG_DETAIL: (id: string) => `${API_PREFIX}/event-configs/${id}`,

  // 组织映射
  ORG_MAPPINGS: `${API_PREFIX}/org-mappings`,
  ORG_MAPPING_LIST: `${API_PREFIX}/org-mappings/list`,
  ORG_MAPPING_BATCH: `${API_PREFIX}/org-mappings/batch`,
  ORG_MAPPING_DETAIL: (id: string) => `${API_PREFIX}/org-mappings/${id}`,

  // 事件数据查询
  EVENT_DATA_QUERY: (eventId: string) =>
    `${API_PREFIX}/events/${eventId}/data/query`,
  EVENT_DETAIL: (eventId: string, journeyId: number) =>
    `${API_PREFIX}/events/${eventId}/data/${journeyId}`,
};

// ===== 表单验证规则 =====
export const FORM_RULES = {
  REQUIRED: { required: true, message: "此字段为必填项" },
  HOST: [
    { required: true, message: "请输入数据库地址" },
    { pattern: /^[a-zA-Z0-9.-]+$/, message: "请输入有效的主机地址" },
  ],
  PORT: [
    { required: true, message: "请输入端口号" },
    { type: "number", min: 1, max: 65535, message: "端口号必须在1-65535之间" },
  ],
  DATABASE: [
    { required: true, message: "请输入数据库名称" },
    { pattern: /^[a-zA-Z_][a-zA-Z0-9_]*$/, message: "数据库名称格式不正确" },
  ],
  USERNAME: [
    { required: true, message: "请输入用户名" },
    { min: 1, max: 100, message: "用户名长度应在1-100个字符之间" },
  ],
  PASSWORD: [
    { required: true, message: "请输入密码" },
    { min: 1, message: "密码不能为空" },
  ],
  API_BASE_URL: [
    { required: true, message: "请输入Skylark API基础地址" },
    {
      pattern: /^[a-zA-Z0-9][-a-zA-Z0-9.]*[a-zA-Z0-9]$/,
      message: "请输入有效的域名（例如：skylark.example.com）"
    },
  ],
  API_TOKEN: [
    { required: true, message: "请输入API认证Token" },
    { min: 1, message: "Token不能为空" },
  ],
  EVENT_CONFIG_NAME: [
    { required: true, message: "请输入配置名称" },
    { min: 1, max: 100, message: "配置名称长度应在1-100个字符之间" },
  ],
};

// ===== 时间格式化配置 =====
export const DATE_FORMAT = {
  DATE: "YYYY-MM-DD",
  DATETIME: "YYYY-MM-DD HH:mm:ss",
  TIME: "HH:mm:ss",
};

// ===== ProTable valueType 映射 =====
export const getProTableValueType = (fieldType: string): string => {
  switch (fieldType) {
    case "date":
      return "date";
    case "datetime":
      return "dateTime";
    case "number":
      return "digit";
    case "boolean":
      return "select";
    default:
      return "text";
  }
};

// ===== 消息提示配置 =====
export const MESSAGE = {
  // 成功消息
  SAVE_SUCCESS: "保存成功",
  DELETE_SUCCESS: "删除成功",
  CREATE_SUCCESS: "创建成功",
  UPDATE_SUCCESS: "更新成功",
  TEST_CONNECTION_SUCCESS: "连接测试成功",
  BATCH_IMPORT_SUCCESS: "批量导入成功",

  // 错误消息
  SAVE_FAILED: "保存失败",
  DELETE_FAILED: "删除失败",
  CREATE_FAILED: "创建失败",
  UPDATE_FAILED: "更新失败",
  FETCH_FAILED: "获取数据失败",
  TEST_CONNECTION_FAILED: "连接测试失败",
  BATCH_IMPORT_FAILED: "批量导入失败",

  // 确认消息
  DELETE_CONFIRM: "确定要删除此项吗？",
  BATCH_DELETE_CONFIRM: "确定要删除选中的项吗？",
  UNSAVED_CHANGES_CONFIRM: "您有未保存的更改，确定要离开吗？",
};

// ===== 默认配置值 =====
export const DEFAULT_VALUES = {
  PORT: 5432,
  NAMESPACE_ID: 1,
  FIELD_DISPLAY_ORDER: 0,
  PAGE_SIZE: 20,
};
