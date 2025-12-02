import type { RequestOptions } from "@@/plugin-request/request";
import type { RequestConfig } from "@umijs/max";
import { message, notification } from "antd";
import { clearToken, isTokenExpired } from "@/utils/token";

// 错误处理方案： 错误类型
enum ErrorShowType {
  SILENT = 0,
  WARN_MESSAGE = 1,
  ERROR_MESSAGE = 2,
  NOTIFICATION = 3,
  REDIRECT = 9,
}
// 与后端约定的响应数据格式
interface ResponseStructure {
  success: boolean;
  data: any;
  errorCode?: number;
  errorMessage?: string;
  showType?: ErrorShowType;
}

/**
 * @name 错误处理
 * pro 自带的错误处理， 可以在这里做自己的改动
 * @doc https://umijs.org/docs/max/request#配置
 */
export const errorConfig: RequestConfig = {
  // 错误处理： umi@3 的错误处理方案。
  errorConfig: {
    // 错误抛出
    errorThrower: (res) => {
      const { success, data, errorCode, errorMessage, showType } =
        res as unknown as ResponseStructure;
      if (!success) {
        const error: any = new Error(errorMessage);
        error.name = "BizError";
        error.info = { errorCode, errorMessage, showType, data };
        throw error; // 抛出自制的错误
      }
    },
    // 错误接收及处理
    errorHandler: (error: any, opts: any) => {
      if (opts?.skipErrorHandler) throw error;
      // 我们的 errorThrower 抛出的错误。
      if (error.name === "BizError") {
        const errorInfo: ResponseStructure | undefined = error.info;
        if (errorInfo) {
          const { errorMessage, errorCode } = errorInfo;
          switch (errorInfo.showType) {
            case ErrorShowType.SILENT:
              // do nothing
              break;
            case ErrorShowType.WARN_MESSAGE:
              message.warning(errorMessage);
              break;
            case ErrorShowType.ERROR_MESSAGE:
              message.error(errorMessage);
              break;
            case ErrorShowType.NOTIFICATION:
              notification.open({
                description: errorMessage,
                message: errorCode,
              });
              break;
            case ErrorShowType.REDIRECT:
              // TODO: redirect
              break;
            default:
              message.error(errorMessage);
          }
        }
      } else if (error.response) {
        // Axios 的错误
        // 请求成功发出且服务器也响应了状态码，但状态代码超出了 2xx 的范围
        const { status, data } = error.response;

        // 处理 401 未授权错误
        if (status === 401) {
          const reason = data?.reason;
          const errorMsg = data?.message || "用户认证失败";

          // 根据 reason 区分不同的 401 错误类型
          if (reason === "TOKEN_EXPIRED") {
            message.error("登录已过期，请重新登录");
          } else if (reason === "TOKEN_MALFORMED") {
            message.error("无效的登录凭证");
          } else {
            message.error(errorMsg);
          }

          // 清除本地 token
          localStorage.removeItem("token");

          // 跳转到登录页
          if (
            typeof window !== "undefined" &&
            window.location.pathname !== "/user/login"
          ) {
            setTimeout(() => {
              window.location.href = "/user/login";
            }, 1000);
          }
          return;
        }

        message.error(`Response status:${status}`);
      } else if (error.request) {
        // 请求已经成功发起，但没有收到响应
        // \`error.request\` 在浏览器中是 XMLHttpRequest 的实例，
        // 而在node.js中是 http.ClientRequest 的实例
        message.error("None response! Please retry.");
      } else {
        // 发送请求时出了点问题
        message.error("Request error, please retry.");
      }
    },
  },

  // 请求拦截器
  requestInterceptors: [
    (config: RequestOptions) => {
      // 拦截请求配置，进行个性化处理。
      const token = localStorage.getItem("token");
      if (token) {
        // 请求前预检：检查 Token 是否即将过期（提前 30 秒）
        if (isTokenExpired(30)) {
          clearToken();
          message.error("登录已过期，请重新登录");
          // 跳转到登录页
          if (
            typeof window !== "undefined" &&
            window.location.pathname !== "/user/login"
          ) {
            setTimeout(() => {
              window.location.href = "/user/login";
            }, 1000);
          }
          // 抛出错误阻止请求发送
          throw new Error("Token expired");
        }

        const headers = { ...config.headers, Authorization: `Bearer ${token}` };
        return { ...config, headers };
      }
      return config;
    },
  ],

  // 响应拦截器
  responseInterceptors: [
    (response) => {
      // 拦截响应数据，进行个性化处理
      const { data } = response;

      // 处理后端统一返回格式：{ code, msg, data } 或 { code, message, data }
      if (data && typeof data === "object" && "code" in data) {
        const apiResponse = data as {
          code: number;
          msg?: string;
          message?: string;
          [key: string]: any;
        };
        // 只处理认证相关错误，其他错误交给业务代码处理
        if (apiResponse.code === 401) {
          const errorMsg =
            apiResponse.msg || apiResponse.message || "用户认证失败";
          message.error(errorMsg);
          // 认证失败时跳转到登录页
          if (
            typeof window !== "undefined" &&
            window.location.pathname !== "/user/login"
          ) {
            setTimeout(() => {
              window.location.href = "/user/login";
            }, 1000);
          }
        }
        // 其他业务错误（403、400、500等）不在这里处理，交给业务代码处理
      }
      // 兼容旧格式
      else if (
        data &&
        typeof data === "object" &&
        "success" in data &&
        (data as { success: boolean }).success === false
      ) {
        message.error("请求失败！");
      }

      return response;
    },
  ],
};
