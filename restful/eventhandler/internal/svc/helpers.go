package svc

import (
	"fmt"

	"github.com/rezeropoint/go-skylark/v2/core"
)

// HandleSkylarkError 处理Skylark引擎错误并转换为HTTP响应
func HandleSkylarkError(err error) (code int64, msg string) {
	switch err {
	// ========== 配置相关 ==========
	case core.ErrEventConfigNotFound:
		return 404, "事件配置不存在"
	case core.ErrInvalidEventConfig:
		return 400, "事件配置无效"
	case core.ErrRemoteTableNotFound:
		return 404, "远程表不存在"
	case core.ErrPlatformConfigNotFound:
		return 404, "平台配置不存在"
	case core.ErrInvalidPlatformConfig:
		return 400, "平台配置无效"

	// ========== 字段相关 ==========
	case core.ErrInvalidFieldName:
		return 400, "字段名无效"
	case core.ErrNoVisibleFields:
		return 400, "没有可见字段"
	case core.ErrInvalidFieldType:
		return 400, "字段类型无效"

	// ========== 权限相关 ==========
	case core.ErrNoOrgAccess:
		return 403, "没有组织访问权限"
	case core.ErrAccessDenied:
		return 403, "访问被拒绝"
	case core.ErrOrgFieldNotConfigured:
		return 400, "组织字段未配置"

	// ========== 查询相关 ==========
	case core.ErrInvalidQueryParam:
		return 400, "查询参数无效"
	case core.ErrQueryTimeout:
		return 408, "查询超时"
	case core.ErrDatabaseConnection:
		return 500, "数据库连接失败"
	case core.ErrJourneyNotFound:
		return 404, "Journey不存在"
	case core.ErrInvalidPageParam:
		return 400, "分页参数无效"

	// ========== Flow相关 ==========
	case core.ErrFlowNotFound:
		return 404, "流程不存在"
	case core.ErrInvalidFlowID:
		return 400, "流程ID无效"

	// ========== 组织映射相关 ==========
	case core.ErrOrgMappingNotFound:
		return 404, "组织映射不存在"
	case core.ErrOrgMappingExists:
		return 400, "组织映射已存在"
	case core.ErrInvalidLocalOrgID:
		return 400, "本地组织ID无效"
	case core.ErrOrgMappingInUse:
		return 400, "组织映射正在使用中"

	// ========== 用户映射相关 ==========
	case core.ErrUserMappingNotFound:
		return 400, "用户映射不存在：请先绑定 Skylark 账号"
	case core.ErrUserNotFound:
		return 404, "用户不存在"
	case core.ErrUserMappingExists:
		return 400, "用户映射已存在"
	case core.ErrUserCreateFailed:
		return 500, "创建用户失败"

	// ========== Skylark API 相关 ==========
	case core.ErrSkylarkAPIUnauthorized:
		return 401, "Skylark API 认证失败"
	case core.ErrSkylarkAPINotFound:
		return 404, "Skylark API 资源不存在"
	case core.ErrSkylarkAPIBadRequest:
		return 400, "Skylark API 请求参数错误"
	case core.ErrSkylarkAPIServerError:
		return 500, "Skylark API 服务器错误"

	default:
		return 500, fmt.Sprintf("服务器错误: %v", err)
	}
}
