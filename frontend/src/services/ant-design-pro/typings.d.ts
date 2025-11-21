// @ts-ignore
/* eslint-disable */

// ===== 重构后的类型定义兼容性转接文件 =====
// 这个文件现在作为兼容性转接文件，导入模块化的类型定义
// 确保现有代码可以无缝工作，同时提供更好的类型组织结构

/// <reference path="../../types/api/index.d.ts" />

// 所有类型现在都通过 types/api/ 目录下的模块化文件进行定义
// 这个文件只保留必要的兼容性声明和一些过渡期需要的类型

declare namespace API {
  // ===== 分页请求参数 (补充) =====
  /** 分页请求参数 */
  type PageParamsRequest = {
    current?: number;
    pageSize?: number;
  };
}
