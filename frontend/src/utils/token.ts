/**
 * JWT Token 工具函数
 */

const TOKEN_KEY = "token";

/**
 * 解析 JWT Token 的 payload 部分
 * JWT 格式: header.payload.signature
 */
function parseJwtPayload(token: string): Record<string, unknown> | null {
  try {
    const parts = token.split(".");
    if (parts.length !== 3) {
      return null;
    }
    // Base64Url 解码 payload
    const payload = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const decoded = atob(payload);
    return JSON.parse(decoded);
  } catch {
    return null;
  }
}

/**
 * 获取 Token 过期时间戳（秒）
 * @returns 过期时间戳，如果无法解析则返回 null
 */
export function getTokenExpireTime(): number | null {
  const token = localStorage.getItem(TOKEN_KEY);
  if (!token) {
    return null;
  }

  const payload = parseJwtPayload(token);
  if (!payload || typeof payload.exp !== "number") {
    return null;
  }

  return payload.exp;
}

/**
 * 检查 Token 是否已过期
 * @param bufferSeconds 提前多少秒判定为过期（默认 60 秒）
 * @returns true 表示已过期或即将过期
 */
export function isTokenExpired(bufferSeconds = 60): boolean {
  const expireTime = getTokenExpireTime();
  if (expireTime === null) {
    return true; // 无法解析视为过期
  }

  const now = Math.floor(Date.now() / 1000);
  return now >= expireTime - bufferSeconds;
}

/**
 * 获取 Token 剩余有效时间（秒）
 * @returns 剩余秒数，如果已过期返回 0，无法解析返回 null
 */
export function getTokenRemainingTime(): number | null {
  const expireTime = getTokenExpireTime();
  if (expireTime === null) {
    return null;
  }

  const now = Math.floor(Date.now() / 1000);
  const remaining = expireTime - now;
  return remaining > 0 ? remaining : 0;
}

/**
 * 清除 Token
 */
export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

/**
 * 获取 Token
 */
export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

/**
 * 设置 Token
 */
export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}
