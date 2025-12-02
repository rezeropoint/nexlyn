/**
 * HTTP 接收配置相关常量定义
 */

/**
 * JSON 字段路径（sourcePath）使用说明
 *
 * 支持的路径格式：
 * 1. 简单字段：          "temperature"
 * 2. 嵌套对象：          "data.temperature"
 * 3. 数组索引：          "Result.Tags[0]"
 * 4. 混合使用：          "data.items[2].name"
 * 5. 多维数组：          "matrix[0][1]"
 *
 * 示例数据：
 * {
 *   "Result": {
 *     "Tags": ["温度", "湿度", "压力"],
 *     "Items": [
 *       { "id": 1, "name": "sensor1", "value": 25.5 },
 *       { "id": 2, "name": "sensor2", "value": 60.0 }
 *     ]
 *   }
 * }
 *
 * 字段映射配置示例：
 * - 提取第一个标签：       sourcePath = "Result.Tags[0]"         → "温度"
 * - 提取第一个传感器值：   sourcePath = "Result.Items[0].value"  → 25.5
 * - 提取第二个传感器名称： sourcePath = "Result.Items[1].name"   → "sensor2"
 */

// 分发类型选项
export const DISPATCH_TYPE_OPTIONS = [
  {
    label: "Skylark 流程触发",
    value: "skylark_flows",
    platformType: "skylark",
  },
  {
    label: "Skylark 表单提交",
    value: "skylark_forms",
    platformType: "skylark",
  },
  { label: "LynxGraph 逻辑引擎", value: "lynxgraph", platformType: null },
  { label: "日志记录", value: "log", platformType: null },
];

// 分发类型显示名称
export const DISPATCH_TYPE_NAMES: Record<string, string> = {
  skylark_flows: "Skylark 流程触发",
  skylark_forms: "Skylark 表单提交",
  lynxgraph: "LynxGraph 逻辑引擎",
  log: "日志记录",
};

// 分发类型标签颜色
export const DISPATCH_TYPE_COLORS: Record<string, string> = {
  skylark_flows: "blue",
  skylark_forms: "cyan",
  lynxgraph: "purple",
  log: "default",
};

// 时间戳格式选项
export const TIMESTAMP_FORMAT_OPTIONS = [
  { label: "Unix 时间戳 (秒) - 如: 1732608000", value: "unix" },
  { label: "Unix 时间戳 (毫秒) - 如: 1732608000000", value: "unix_ms" },
  { label: "Unix 时间戳 (纳秒) - 如: 1732608000000000000", value: "unix_nano" },
  { label: "ISO 8601 - 如: 2024-11-26T12:00:00Z", value: "iso8601" },
  { label: "RFC 3339 - 如: 2024-11-26T12:00:00+08:00", value: "rfc3339" },
];

// 时间戳格式显示名称
export const TIMESTAMP_FORMAT_NAMES: Record<string, string> = {
  unix: "Unix 时间戳 (秒)",
  unix_ms: "Unix 时间戳 (毫秒)",
  unix_nano: "Unix 时间戳 (纳秒)",
  iso8601: "ISO 8601",
  rfc3339: "RFC 3339",
};

// 所有字段类型选项
export const ALL_FIELD_TYPE_OPTIONS = [
  { label: "字符串", value: "string" },
  { label: "图片URL", value: "imageURL" },
  { label: "图片Base64", value: "imageBase64" },
  { label: "整数", value: "int" },
  { label: "浮点数", value: "float" },
  { label: "布尔值", value: "bool" },
];

// Skylark 支持的字段类型（go-skylark FieldType）
export const SKYLARK_FIELD_TYPES = ["string", "imageURL", "imageBase64"];

// LynxGraph 支持的字段类型（lynxgraph/core FieldType）
export const LYNXGRAPH_FIELD_TYPES = ["string", "int", "float", "bool"];

// 各分发类型支持的字段类型映射
export const DISPATCH_SUPPORTED_FIELD_TYPES: Record<string, string[]> = {
  skylark_flows: SKYLARK_FIELD_TYPES,
  skylark_forms: SKYLARK_FIELD_TYPES,
  lynxgraph: LYNXGRAPH_FIELD_TYPES,
  log: ALL_FIELD_TYPE_OPTIONS.map((opt) => opt.value), // log 支持所有类型
};

// 根据分发配置获取支持的字段类型（取交集）
export const getSupportedFieldTypes = (
  dispatchConfigs?: Array<{ type?: string }>
): string[] => {
  if (!dispatchConfigs || dispatchConfigs.length === 0) {
    // 无分发配置时，支持所有类型
    return ALL_FIELD_TYPE_OPTIONS.map((opt) => opt.value);
  }

  // 计算所有分发配置支持类型的交集
  let result: string[] | null = null;
  for (const dc of dispatchConfigs) {
    if (!dc.type) continue;
    const supported =
      DISPATCH_SUPPORTED_FIELD_TYPES[dc.type] ||
      ALL_FIELD_TYPE_OPTIONS.map((opt) => opt.value);
    if (result === null) {
      result = [...supported];
    } else {
      // 取交集
      result = result.filter((x) => supported.includes(x));
    }
  }

  return result ?? ALL_FIELD_TYPE_OPTIONS.map((opt) => opt.value);
};

// 根据支持的类型过滤字段类型选项
export const getFieldTypeOptions = (supportedTypes: string[]) => {
  return ALL_FIELD_TYPE_OPTIONS.filter((opt) =>
    supportedTypes.includes(opt.value)
  );
};

// 向后兼容：默认字段类型选项（所有类型）
export const FIELD_TYPE_OPTIONS = ALL_FIELD_TYPE_OPTIONS;

// 每页显示数量
export const DEFAULT_PAGE_SIZE = 10;
export const PAGE_SIZE_OPTIONS = ["10", "20", "50", "100"];
