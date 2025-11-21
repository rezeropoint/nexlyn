// 用户管理常量定义

// 用户状态选项
export const USER_STATUS_OPTIONS = [
  { label: "活跃", value: "active" },
  { label: "禁用", value: "inactive" },
];

// 表格默认配置
export const TABLE_CONFIG = {
  DEFAULT_PAGE_SIZE: 10,
  SCROLL_X: 800,
};

// 表单验证规则
export const FORM_RULES = {
  REQUIRED_NAME: { required: true, message: "用户姓名为必填项" },
  REQUIRED_USERNAME: { required: true, message: "用户名为必填项" },
  REQUIRED_EMAIL: { required: true, message: "邮箱为必填项" },
  EMAIL_FORMAT: { type: "email" as const, message: "邮箱格式不正确" },
  REQUIRED_TENANT: { required: true, message: "租户为必填项" },
  REQUIRED_PASSWORD: { required: true, message: "密码为必填项" },
  PASSWORD_LENGTH: { min: 6, message: "密码长度至少6位" },
};

// 用户状态颜色映射
export const USER_STATUS_COLORS = {
  active: "green",
  inactive: "red",
} as const;

// 用户状态标签文本映射
export const USER_STATUS_TEXT = {
  active: "活跃",
  inactive: "禁用",
} as const;
