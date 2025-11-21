package permission

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCurrentUserPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取当前用户权限
func NewGetCurrentUserPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrentUserPermissionsLogic {
	return &GetCurrentUserPermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCurrentUserPermissionsLogic) GetCurrentUserPermissions() (resp *types.GetUserPermissionsResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "permission"),
		logx.Field("operation", "get_current_user_permissions"),
		logx.Field("status", "started"),
	).Info("开始获取当前用户权限")

	// 从JWT中获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "permission"),
			logx.Field("operation", "get_current_user_permissions"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从JWT获取用户信息失败")

		return &types.GetUserPermissionsResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
		}, nil
	}

	// 调试信息：打印JWT用户信息
	logx.WithContext(l.ctx).WithFields(
		logx.Field("user_key", jwtUser.UserKey),
		logx.Field("tenant_id", jwtUser.TenantId),
		logx.Field("tenant_key", jwtUser.TenantKey),
		logx.Field("username", jwtUser.Username),
	).Info("JWT用户信息")

	// 获取用户有效权限（包含通过角色继承的权限）
	permissions, err := l.svcCtx.Casbinx.GetEffectivePermissionsSecure(jwtUser.UserKey, jwtUser.UserKey, jwtUser.TenantKey)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "permission"),
			logx.Field("operation", "get_current_user_permissions"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("tenant_key", jwtUser.TenantKey),
			logx.Field("error", err.Error()),
		).Error("获取用户权限失败")

		return &types.GetUserPermissionsResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "获取用户权限失败",
			},
		}, nil
	}

	// 获取用户角色列表
	roles, err := l.svcCtx.Casbinx.GetUserRoles(jwtUser.UserKey, jwtUser.TenantKey)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "permission"),
			logx.Field("operation", "get_current_user_permissions"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("获取用户角色失败")

		return &types.GetUserPermissionsResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "获取用户角色失败",
			},
		}, nil
	}

	// 转换权限格式以匹配响应结构
	permissionStrings := make([]string, len(permissions))
	for i, perm := range permissions {
		permissionStrings[i] = fmt.Sprintf("%s:%s", perm.Resource, perm.Action)
	}

	// 调试信息：打印获取到的权限和角色
	logx.WithContext(l.ctx).WithFields(
		logx.Field("permissions", permissionStrings),
		logx.Field("roles", roles),
		logx.Field("permissions_count", len(permissionStrings)),
		logx.Field("roles_count", len(roles)),
	).Info("获取到的权限和角色信息")

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "permission"),
		logx.Field("operation", "get_current_user_permissions"),
		logx.Field("status", "success"),
		logx.Field("user_key", jwtUser.UserKey),
		logx.Field("permissions_count", len(permissionStrings)),
		logx.Field("roles_count", len(roles)),
	).Info("获取当前用户权限成功")

	return &types.GetUserPermissionsResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "获取用户权限成功",
		},
		Permissions: permissionStrings,
		Roles:       roles,
	}, nil
}
