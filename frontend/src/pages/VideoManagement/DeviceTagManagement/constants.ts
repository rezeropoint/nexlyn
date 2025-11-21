// 表格配置
export const DEVICE_TAG_TABLE_CONFIG = {
  DEFAULT_PAGE_SIZE: 20,
  SEARCH_LABEL_WIDTH: 120,
  SCROLL_X: 1000,
} as const;

// 表单验证规则
export const DEVICE_TAG_FORM_RULES = {
  TAG_NAME: {
    REQUIRED: { required: true, message: "请输入标签名称" },
    MAX_LENGTH: { max: 20, message: "标签名称不能超过20个字符" },
  },
  DESCRIPTION: {
    MAX_LENGTH: { max: 200, message: "描述不能超过200个字符" },
  },
} as const;
