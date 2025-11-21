/**
 * 公共类型定义
 */

// 标准响应格式
export interface BaseResponse {
  code: number;
  msg?: string;
  message?: string; // 兼容字段
}

// 分页参数
export interface PageParams {
  page?: number;
  pageSize?: number;
  total?: number;
  pages?: number;
}
