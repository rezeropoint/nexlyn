package config

import (
	"fmt"
)

// 数据库操作错误
const (
	ErrMsgDatabaseQuery       = "数据库查询失败"
	ErrMsgDatabaseInsert      = "数据库插入失败"
	ErrMsgDatabaseUpdate      = "数据库更新失败"
	ErrMsgDatabaseDelete      = "数据库删除失败"
	ErrMsgDatabaseConnection  = "数据库连接失败"
	ErrMsgDatabaseTransaction = "数据库事务失败"
)

// 用户相关错误
const (
	ErrMsgUserQuery        = "查询用户信息失败"
	ErrMsgUserCreate       = "创建用户失败"
	ErrMsgUserUpdate       = "更新用户信息失败"
	ErrMsgUserDelete       = "删除用户失败"
	ErrMsgUserPasswordHash = "用户密码加密失败"
	ErrMsgUserEmailExists  = "邮箱已存在"
	ErrMsgUserNotFound     = "用户不存在"
	ErrMsgUserValidation   = "用户数据验证失败"
	ErrMsgUserOptions      = "查询用户选项失败"
)

// 租户相关错误
const (
	ErrMsgTenantQuery      = "查询租户信息失败"
	ErrMsgTenantCreate     = "创建租户失败"
	ErrMsgTenantUpdate     = "更新租户信息失败"
	ErrMsgTenantDelete     = "删除租户失败"
	ErrMsgTenantNotFound   = "租户不存在"
	ErrMsgTenantValidation = "租户数据验证失败"
	ErrMsgTenantOptions    = "查询租户选项失败"
	ErrMsgTenantUserCheck  = "检查租户下用户失败"
)

// 权限相关错误
const (
	ErrMsgPermissionCheck  = "权限检查失败"
	ErrMsgPermissionDenied = "权限不足"
	ErrMsgRoleQuery        = "查询角色信息失败"
	ErrMsgRoleCreate       = "创建角色失败"
	ErrMsgRoleUpdate       = "更新角色信息失败"
	ErrMsgRoleDelete       = "删除角色失败"
)

// 标签相关错误
const (
	ErrMsgTagQuery      = "查询标签信息失败"
	ErrMsgTagCreate     = "创建标签失败"
	ErrMsgTagUpdate     = "更新标签信息失败"
	ErrMsgTagDelete     = "删除标签失败"
	ErrMsgTagValidation = "标签数据验证失败"
	ErrMsgTagOptions    = "查询标签选项失败"
)

// 认证相关错误
const (
	ErrMsgAuthLogin      = "用户登录失败"
	ErrMsgAuthToken      = "令牌验证失败"
	ErrMsgAuthInvalid    = "身份验证无效"
	ErrMsgPasswordVerify = "密码验证失败"
)

// 组织相关错误
const (
	ErrMsgOrganizationQuery         = "查询组织信息失败"
	ErrMsgOrganizationCreate        = "创建组织失败"
	ErrMsgOrganizationUpdate        = "更新组织信息失败"
	ErrMsgOrganizationDelete        = "删除组织失败"
	ErrMsgOrganizationNotFound      = "组织不存在或不属于当前租户"
	ErrMsgOrganizationValidation    = "组织数据验证失败"
	ErrMsgOrganizationMemberAdd     = "添加组织成员失败"
	ErrMsgOrganizationMemberRemove  = "移除组织成员失败"
	ErrMsgOrganizationMemberExists  = "用户已在该组织中"
	ErrMsgOrganizationMemberUpdate  = "更新组织成员信息失败"
	ErrMsgOrganizationMove          = "移动组织失败"
	ErrMsgOrganizationCodeExists    = "组织代码已存在"
	ErrMsgOrganizationLevelExceeded = "组织层级不能超过10层"
)

// 格式化错误信息，包含具体错误详情
func FormatError(baseMsg string, err error) string {
	if err != nil {
		return fmt.Sprintf("%s: %v", baseMsg, err)
	}
	return baseMsg
}

// 格式化带参数的错误信息
func FormatErrorf(baseMsg string, args ...interface{}) string {
	return fmt.Sprintf(baseMsg, args...)
}
