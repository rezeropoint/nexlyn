package plugin_nexlyn

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Auth JWT认证配置
type Auth struct {
	AccessSecret string `yaml:"AccessSecret"` // JWT签名密钥
	AccessExpire int64  `yaml:"AccessExpire"` // Token过期时间（秒）
}

// JWTUser JWT中存储的用户摘要信息（用于后端接口权限判断）
type JWTUser struct {
	// 用户身份
	UserKey  string `json:"userKey"`  // 用户业务标识符（用于CasbinX权限）
	UserId   string `json:"userId"`   // 用户ID（向后兼容）
	Username string `json:"username"` // 用户名

	// 多租户信息
	TenantKey  string `json:"tenantKey"`  // 租户标识符（用于CasbinX权限域）
	TenantId   string `json:"tenantId"`   // 租户ID（向后兼容）
	TenantName string `json:"tenantName"` // 租户名称

	// 组织信息
	PrimaryOrgId   string `json:"primaryOrgId"`   // 主组织UUID（对应organizations.id，用于数据库关联）
	PrimaryOrgName string `json:"primaryOrgName"` // 主组织名称（对应organizations.name）

	// 权限信息（兼容旧版本）
	Role string `json:"role"` // 用户角色 (super_admin, admin, user)

	// 状态
	Status string `json:"status"` // 用户状态：active-活跃，inactive-禁用，deleted-删除
}

// GetUserFromJWT 从JWT上下文中获取用户信息
func GetUserFromJWT(ctx context.Context) (*JWTUser, error) {
	// 从上下文中获取JWT payload
	payload := ctx.Value("payload")
	if payload == nil {
		return nil, fmt.Errorf("未找到JWT payload")
	}

	// 确保payload是字符串类型（JWTAuthMiddleware已经处理了类型转换）
	payloadStr, ok := payload.(string)
	if !ok {
		return nil, fmt.Errorf("JWT payload格式错误: 期望string类型，实际为%T", payload)
	}

	// 解析用户信息
	var user JWTUser
	err := json.Unmarshal([]byte(payloadStr), &user)
	if err != nil {
		return nil, fmt.Errorf("解析用户信息失败: %v", err)
	}

	return &user, nil
}

// GetUserKey 获取用户Key（优先使用userKey，向后兼容userId）
func (u *JWTUser) GetUserKey() string {
	if u.UserKey != "" {
		return u.UserKey
	}
	return u.UserId
}

// GetTenantKey 获取租户Key（优先使用tenantKey，向后兼容tenantId）
func (u *JWTUser) GetTenantKey() string {
	if u.TenantKey != "" {
		return u.TenantKey
	}
	return u.TenantId
}

// IsActive 检查用户是否活跃
func (u *JWTUser) IsActive() bool {
	return u.Status == "active"
}

// IsSameTenant 检查是否属于同一租户（支持新旧字段）
func (u *JWTUser) IsSameTenant(tenantKey string) bool {
	return u.GetTenantKey() == tenantKey
}

// sendUnauthorizedResponse 发送认证失败响应，根据错误类型返回不同的reason
func sendUnauthorizedResponse(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)

	response := map[string]any{
		"code": 401,
	}

	switch {
	case err == nil:
		response["message"] = "未授权访问"
		response["reason"] = "UNAUTHORIZED"
	case errors.Is(err, jwt.ErrTokenExpired):
		response["message"] = "登录已过期，请重新登录"
		response["reason"] = "TOKEN_EXPIRED"
	case errors.Is(err, jwt.ErrTokenMalformed):
		response["message"] = "无效的令牌格式"
		response["reason"] = "TOKEN_MALFORMED"
	case errors.Is(err, jwt.ErrTokenNotValidYet):
		response["message"] = "令牌尚未生效"
		response["reason"] = "TOKEN_NOT_VALID_YET"
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		response["message"] = "令牌签名无效"
		response["reason"] = "TOKEN_SIGNATURE_INVALID"
	default:
		response["message"] = "未授权访问"
		response["reason"] = "UNAUTHORIZED"
	}

	json.NewEncoder(w).Encode(response)
}

// JWTAuthMiddleware JWT认证中间件
func (n *NexlynPlugin) JWTAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从请求头获取Authorization token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendUnauthorizedResponse(w, nil)
			return
		}

		// 检查Bearer前缀
		tokenString := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:]
		} else {
			sendUnauthorizedResponse(w, jwt.ErrTokenMalformed)
			return
		}

		// 解析JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 验证签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(n.Auth.AccessSecret), nil
		})

		if err != nil {
			sendUnauthorizedResponse(w, err)
			return
		}

		if !token.Valid {
			sendUnauthorizedResponse(w, nil)
			return
		}

		// 获取claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			sendUnauthorizedResponse(w, jwt.ErrTokenMalformed)
			return
		}

		// 获取payload
		payload, exists := claims["payload"]
		if !exists {
			sendUnauthorizedResponse(w, jwt.ErrTokenMalformed)
			return
		}

		// 确保payload是字符串类型
		payloadStr, ok := payload.(string)
		if !ok {
			sendUnauthorizedResponse(w, jwt.ErrTokenMalformed)
			return
		}

		// 将payload添加到请求上下文中
		ctx := context.WithValue(r.Context(), "payload", payloadStr)
		r = r.WithContext(ctx)

		// 继续处理请求
		next.ServeHTTP(w, r)
	}
}
