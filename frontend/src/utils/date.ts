import dayjs from "dayjs";

/**
 * 统一的日期时间格式化工具
 */

// 统一的日期时间格式
export const DATE_TIME_FORMAT = "YYYY-MM-DD HH:mm:ss";
export const DATE_FORMAT = "YYYY-MM-DD";

/**
 * 格式化日期时间，自动识别输入格式
 * @param dateValue - 日期值，可以是 Unix 时间戳、ISO 字符串或 Date 对象
 * @param format - 输出格式，默认为 'YYYY-MM-DD HH:mm:ss'
 * @returns 格式化后的日期字符串，如果输入无效则返回 '-'
 */
export const formatDateTime = (
  dateValue: any,
  format: string = DATE_TIME_FORMAT
): string => {
  if (!dateValue) {
    return "-";
  }

  let dayjsObj: dayjs.Dayjs;

  try {
    if (typeof dateValue === "number") {
      // Unix 时间戳处理
      // 如果是10位数字（秒），需要转换为毫秒
      const timestamp = dateValue < 10000000000 ? dateValue * 1000 : dateValue;
      dayjsObj = dayjs(timestamp);
    } else if (typeof dateValue === "string") {
      // ISO 字符串或其他字符串格式
      dayjsObj = dayjs(dateValue);
    } else if (dateValue instanceof Date) {
      // Date 对象
      dayjsObj = dayjs(dateValue);
    } else {
      // 其他类型，尝试直接解析
      dayjsObj = dayjs(dateValue);
    }

    if (!dayjsObj.isValid()) {
      return "-";
    }

    return dayjsObj.format(format);
  } catch (error) {
    console.warn("Date formatting error:", error, "Input:", dateValue);
    return "-";
  }
};

/**
 * 格式化为日期（不包含时间）
 * @param dateValue - 日期值
 * @returns 格式化后的日期字符串
 */
export const formatDate = (dateValue: any): string => {
  return formatDateTime(dateValue, DATE_FORMAT);
};

/**
 * ProTable 日期时间列的通用配置
 * 用于在 ProTable 列定义中统一日期时间格式
 */
export const createDateTimeColumn = (
  title: string,
  dataIndex: string,
  options: {
    width?: number;
    search?: boolean | { transform?: (value: any) => any };
    sorter?: boolean;
    format?: string;
  } = {}
) => ({
  title,
  dataIndex,
  width: options.width || 160,
  valueType: "dateTime" as const,
  search: options.search ?? false,
  sorter: options.sorter || false,
  render: (_: any, record: any) =>
    formatDateTime(record[dataIndex], options.format),
  // ProTable 的 valueType: 'dateTime' 默认配置
  fieldProps: {
    format: options.format || DATE_TIME_FORMAT,
  },
});

/**
 * 检查日期值是否有效
 * @param dateValue - 日期值
 * @returns 是否为有效日期
 */
export const isValidDate = (dateValue: any): boolean => {
  if (!dateValue) return false;

  try {
    let dayjsObj: dayjs.Dayjs;

    if (typeof dateValue === "number") {
      const timestamp = dateValue < 10000000000 ? dateValue * 1000 : dateValue;
      dayjsObj = dayjs(timestamp);
    } else {
      dayjsObj = dayjs(dateValue);
    }

    return dayjsObj.isValid();
  } catch {
    return false;
  }
};

export default {
  formatDateTime,
  formatDate,
  createDateTimeColumn,
  isValidDate,
  DATE_TIME_FORMAT,
  DATE_FORMAT,
};
