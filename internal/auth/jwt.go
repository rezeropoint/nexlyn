package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

func GetJwtToken(secretKey string, iat, seconds int64, payload string) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["payload"] = payload
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}

// JWTUser JWT中存储的用户摘要信息（用于后端接口权限判断）
type JWTUser struct {
	// 用户身份
	UserKey  string `json:"userKey"`  // 用户业务标识符（对应system_users.user_key，用于业务逻辑和Casbin）
	UserId   string `json:"userId"`   // 用户UUID（对应users.id，用于数据库关联）
	Username string `json:"userName"` // 用户名（对应users.user_name）

	// 多租户信息
	TenantKey  string `json:"tenantKey"`  // 租户标识符（对应system_tenants.tenant_key，用于Casbin权限域）
	TenantId   string `json:"tenantId"`   // 租户UUID（对应tenants.id，用于数据库关联）
	TenantName string `json:"tenantName"` // 租户名称

	// 组织信息
	PrimaryOrgId   string `json:"primaryOrgId"`   // 主组织UUID（对应organizations.id，用于数据库关联）
	PrimaryOrgName string `json:"primaryOrgName"` // 主组织名称（对应organizations.name）

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

	payloadStr, ok := payload.(string)
	if !ok {
		return nil, fmt.Errorf("JWT payload格式错误")
	}

	// 解析用户信息
	var user JWTUser
	err := json.Unmarshal([]byte(payloadStr), &user)
	if err != nil {
		return nil, fmt.Errorf("解析用户信息失败: %v", err)
	}

	return &user, nil
}

// CreateJWTUser 创建JWT用户摘要信息
func CreateJWTUser(userKey, userId, userName string,
	tenantKey, tenantId, tenantName string,
	primaryOrgId, primaryOrgName string,
	status string) *JWTUser {
	return &JWTUser{
		UserKey:  userKey,
		UserId:   userId,
		Username: userName,

		TenantKey:  tenantKey,
		TenantId:   tenantId,
		TenantName: tenantName,

		PrimaryOrgId:   primaryOrgId,
		PrimaryOrgName: primaryOrgName,

		Status: status,
	}
}

// IsActive 检查用户是否活跃
func (u *JWTUser) IsActive() bool {
	return u.Status == "active"
}

// IsSameTenant 检查是否属于同一租户（传入tenant_key进行比较）
func (u *JWTUser) IsSameTenant(tenantKey string) bool {
	return u.TenantKey == tenantKey
}
