/**
 * 设备类别和标准字段相关类型
 */

// 设备类别信息
export interface DeviceCategoryInfo {
  code: string; // 类别代码（如：temperature_sensor）
  name: string; // 中文名称（如：温度传感器）
  nameEn: string; // 英文名称（如：Temperature Sensor）
  description: string; // 描述
  icon: string; // 图标名称（用于前端展示）
}

// 标准字段信息
export interface StandardFieldInfo {
  name: string; // 字段名（如：temperature）
  displayName: string; // 显示名称（如：温度）
  fieldType: string; // 字段类型（string/number/boolean等）
  unit?: string; // 单位（如：℃, %, ppm）
  description?: string; // 字段描述
  required: boolean; // 是否必填
  minValue?: number; // 最小值（数值类型）
  maxValue?: number; // 最大值（数值类型）
  enumValues?: string[]; // 枚举值（枚举类型）
}
