// 通用基础类型定义

declare namespace API {
  /** 基础响应 */
  type BaseResponse = {
    code: number;
    msg?: string;
  };

  /** 分页参数 */
  type PageParams = {
    current?: number;
    pageSize?: number;
    total?: number;
  };

  /** 分页请求参数 */
  type PageParamsRequest = {
    current?: number;
    pageSize?: number;
  };

  /** 错误响应 */
  type ErrorResponse = {
    /** 业务约定的错误码 */
    errorCode: string;
    /** 业务上的错误信息 */
    errorMessage?: string;
    /** 业务上的请求是否成功 */
    success?: boolean;
  };

  /** 用户地理位置 */
  type UserLocation = {
    label?: string;
    key?: string;
  };

  /** 用户地理信息 */
  type UserGeographic = {
    province?: UserLocation;
    city?: UserLocation;
  };

  /** 用户标签 */
  type UserTag = {
    id?: string;
    label?: string;
  };
}
