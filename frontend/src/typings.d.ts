declare module "slash2";
declare module "*.css";
declare module "*.less";
declare module "*.scss";
declare module "*.sass";
declare module "*.svg";
declare module "*.png";
declare module "*.jpg";
declare module "*.jpeg";
declare module "*.gif";
declare module "*.bmp";
declare module "*.tiff";
declare module "omit.js";
declare module "numeral";
declare module "mockjs";
declare module "react-fittext";

declare const REACT_APP_ENV: "test" | "dev" | "pre" | false;

// Skylark Sync 相关类型定义
declare namespace API {
  // 同步单个用户响应
  type SyncUserResponse = {
    code: number;
    msg?: string;
    data?: {
      remoteUserId: string;
    };
  };

  // 查询用户同步状态响应
  type GetUserSyncStatusResponse = {
    code: number;
    msg?: string;
    data?: {
      syncStatus: Record<string, boolean>; // key: 用户ID, value: 是否已同步
    };
  };

  // 同步单个组织响应
  type SyncOrganizationResponse = {
    code: number;
    msg?: string;
    data?: {
      remoteOrgId: string;
    };
  };

  // 查询组织同步状态响应
  type GetOrgSyncStatusResponse = {
    code: number;
    msg?: string;
    data?: {
      syncStatus: Record<string, boolean>; // key: 组织ID, value: 是否已同步
    };
  };

  // 健康检查响应
  type SyncHealthCheckResponse = {
    code: number;
    msg?: string;
    data?: {
      healthy: boolean;
    };
  };

  // 用户绑定请求
  type BindUserRequest = {
    remoteUserId: number;
  };

  // 用户绑定响应
  type BindUserResponse = {
    code: number;
    msg?: string;
  };

  // 用户解绑响应
  type UnbindUserResponse = {
    code: number;
    msg?: string;
  };

  // 组织绑定请求
  type BindOrganizationRequest = {
    remoteOrgId: number;
  };

  // 组织绑定响应
  type BindOrganizationResponse = {
    code: number;
    msg?: string;
  };

  // 组织解绑响应
  type UnbindOrganizationResponse = {
    code: number;
    msg?: string;
  };
}
