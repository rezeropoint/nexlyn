/**
 * IoT服务统一导出
 *
 * 目录结构：
 * - constants.ts: API前缀等常量
 * - types/: 类型定义（按模块拆分）
 * - api/: API函数（按模块拆分）
 */

// 导出常量
export { API_PREFIX } from "./constants";

// 导出所有类型
export * from "./types";

// 导出所有API函数
export * from "./api";
