// 标签作用域配置
export const TAG_SCOPES = {
  user: { text: "用户", color: "green", status: "Processing" },
  tenant: { text: "租户", color: "blue", status: "Success" },
} as const;

// 表格配置
export const TABLE_CONFIG = {
  DEFAULT_PAGE_SIZE: 10,
  SEARCH_LABEL_WIDTH: 120,
  SCROLL_X: 1000,
} as const;

// 表单验证规则
export const FORM_RULES = {
  LABEL: {
    REQUIRED: { required: true, message: "请输入标签名称" },
    MAX_LENGTH: { max: 50, message: "标签名称不能超过50个字符" },
  },
  DESCRIPTION: {
    MAX_LENGTH: { max: 200, message: "描述不能超过200个字符" },
  },
  SCOPE: {
    REQUIRED: { required: true, message: "请选择标签作用域" },
  },
} as const;

// 默认值
export const DEFAULTS = {
  SCOPE: "user" as const,
} as const;
