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

// 字段类型选项
export const FIELD_TYPE_OPTIONS = [
  { label: "字符串", value: "string" },
  { label: "整数", value: "int" },
  { label: "浮点数", value: "float" },
  { label: "布尔值", value: "bool" },
];

// 每页显示数量
export const DEFAULT_PAGE_SIZE = 10;
export const PAGE_SIZE_OPTIONS = ["10", "20", "50", "100"];
