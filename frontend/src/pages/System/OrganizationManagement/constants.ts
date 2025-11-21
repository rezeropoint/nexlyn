// 组织管理相关常量定义

// 组织类型选项
export const ORGANIZATION_TYPE_OPTIONS = [
  { label: "公司", value: "company" },
  { label: "部门", value: "department" },
  { label: "团队", value: "team" },
  { label: "项目组", value: "project" },
  { label: "其他", value: "other" },
];

// 组织状态选项（包含已删除状态，用于查询和显示）
export const ORGANIZATION_STATUS_OPTIONS = [
  { label: "正常", value: "active" },
  { label: "停用", value: "inactive" },
  { label: "已删除", value: "deleted" },
];

// 新建/编辑表单状态选项（不包含已删除状态）
export const ORGANIZATION_FORM_STATUS_OPTIONS = [
  { label: "正常", value: "active" },
  { label: "停用", value: "inactive" },
];

// 组织状态颜色映射
export const ORGANIZATION_STATUS_COLORS = {
  active: "success",
  inactive: "warning",
  deleted: "error",
} as const;

// 组织状态文本映射
export const ORGANIZATION_STATUS_TEXT = {
  active: "正常",
  inactive: "停用",
  deleted: "已删除",
} as const;

// 关系类型选项
export const RELATION_TYPE_OPTIONS = [
  { label: "负责人", value: "manager" },
  { label: "成员", value: "member" },
];

// 关系类型颜色映射
export const RELATION_TYPE_COLORS = {
  manager: "red",
  member: "blue",
} as const;

// 关系类型文本映射
export const RELATION_TYPE_TEXT = {
  manager: "负责人",
  member: "成员",
} as const;

// 表单验证规则
export const FORM_RULES = {
  REQUIRED_ORG_CODE: { required: true, message: "请输入组织代码!" },
  REQUIRED_ORG_NAME: { required: true, message: "请输入组织名称!" },
  ORG_CODE_PATTERN: {
    pattern: /^[a-zA-Z0-9_-]{2,32}$/,
    message: "组织代码只能包含字母、数字、下划线和连字符，长度2-32位!",
  },
  ORG_NAME_LENGTH: {
    min: 2,
    max: 100,
    message: "组织名称长度应在2-100字符之间!",
  },
  DESCRIPTION_LENGTH: {
    max: 500,
    message: "描述长度不能超过500字符!",
  },
  POSITION_LENGTH: {
    max: 100,
    message: "职位标题长度不能超过100字符!",
  },
} as const;

// 表格配置
export const TABLE_CONFIG = {
  PAGE_SIZE: 10,
  PAGE_SIZE_OPTIONS: ["10", "20", "50", "100"] as const,
  SCROLL_X: 1200,
} as const;

// 树形组件配置
export const TREE_CONFIG = {
  HEIGHT: 600,
  VIRTUAL_HEIGHT: 400,
} as const;

// 默认值
export const DEFAULT_VALUES = {
  ORG_TYPE: "department",
  ORG_STATUS: "active",
  RELATION_TYPE: "member",
  SORT_ORDER: 0,
  IS_PRIMARY: false,
} as const;
