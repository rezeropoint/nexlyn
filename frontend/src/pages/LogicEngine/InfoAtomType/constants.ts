/**
 * 信息原子类型管理 - 常量定义
 */

import type { FieldType } from '@/services/lynxmanager';

/** 表格配置 */
export const TABLE_CONFIG = {
  DEFAULT_PAGE_SIZE: 10, // 默认分页大小
  SCROLL_X: 1200, // 表格横向滚动宽度
};

/** 字段类型选项 */
export const FIELD_TYPE_OPTIONS: Array<{ label: string; value: FieldType }> = [
  { label: '字符串 (String)', value: 'string' as FieldType },
  { label: '整数 (Int)', value: 'int' as FieldType },
  { label: '浮点数 (Float)', value: 'float' as FieldType },
  { label: '布尔值 (Bool)', value: 'bool' as FieldType },
];

/** 字段类型标签颜色映射 */
export const FIELD_TYPE_COLORS: Record<string, string> = {
  string: 'blue',
  int: 'green',
  float: 'orange',
  bool: 'purple',
};

/** 默认版本号 */
export const DEFAULT_VERSION = 'v1';
