// 租户管理常量定义

// 租户状态选项
export const TENANT_STATUS_OPTIONS = [
  { label: "活跃", value: "active" },
  { label: "禁用", value: "inactive" },
  { label: "过期", value: "expired" },
];

// 表格默认配置
export const TABLE_CONFIG = {
  DEFAULT_PAGE_SIZE: 10,
  SCROLL_X: 1200,
};

// 表单验证规则
export const FORM_RULES = {
  REQUIRED_TENANT_KEY: { required: true, message: "租户标识为必填项" },
  REQUIRED_TENANT_NAME: { required: true, message: "租户名称为必填项" },
  REQUIRED_CONTACT_EMAIL: { required: true, message: "联系邮箱为必填项" },
  EMAIL_FORMAT: { type: "email" as const, message: "邮箱格式不正确" },
  REQUIRED_MAX_USERS: { required: true, message: "最大用户数为必填项" },
  MIN_MAX_USERS: {
    type: "number" as const,
    min: 1,
    message: "最大用户数至少为1",
  },
};

// 租户状态颜色映射
export const TENANT_STATUS_COLORS = {
  active: "green",
  inactive: "red",
  expired: "orange",
} as const;

// 租户状态标签文本映射
export const TENANT_STATUS_TEXT = {
  active: "活跃",
  inactive: "禁用",
  expired: "过期",
} as const;

// 默认值
export const DEFAULT_VALUES = {
  MAX_USERS: 10,
} as const;
