// @ts-ignore
/* eslint-disable */
import { request } from "@umijs/max";

/** 获取当前的用户 GET /api/currentUser - 保留用于兼容性 */
export async function currentUser(options?: { [key: string]: any }) {
  return request<{
    data: API.CurrentUser;
  }>("/api/currentUser", {
    method: "GET",
    ...(options || {}),
  });
}

// ===== 以下模板代码已清理，如果项目中仍有使用请迁移到对应的业务服务 =====

// 登录功能已迁移到 services/user/index.ts
// 通知功能如果需要可在相应业务模块中实现
// 规则管理功能为Ant Design Pro模板示例，已移除
