package permission

import (
	"context"
	"strings"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveRoleForUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 为用户移除角色
func NewRemoveRoleForUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveRoleForUserLogic {
	return &RemoveRoleForUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveRoleForUserLogic) RemoveRoleForUser(req *types.RemoveRoleRequest) (resp *types.RemoveRoleForUserResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "permission"),
		logx.Field("operation", "remove_role_for_user"),
		logx.Field("status", "started"),
		logx.Field("user_id", req.UserKey),
		logx.Field("role_key", req.RoleKey),
		logx.Field("tenant_id", req.TenantKey),
	).Info("开始为用户移除角色")

	// 从JWT中获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "permission"),
			logx.Field("operation", "remove_role_for_user"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从JWT获取用户信息失败")

		return &types.RemoveRoleForUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
		}, nil
	}

	// 为用户移除角色（使用 CasbinX 的 RemoveRole 方法，自动进行安全检查）
	err = l.svcCtx.Casbinx.RemoveRole(jwtUser.UserKey, req.UserKey, req.RoleKey, req.TenantKey)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "permission"),
			logx.Field("operation", "remove_role_for_user"),
			logx.Field("status", "failed"),
			logx.Field("user_id", req.UserKey),
			logx.Field("role_key", req.RoleKey),
			logx.Field("tenant_id", req.TenantKey),
			logx.Field("error", err.Error()),
		).Error("为用户移除角色失败")

		// 直接返回具体错误信息，权限相关错误返回403，其他错误返回500
		errorMsg := err.Error()
		var statusCode int64
		statusCode = 500
		if strings.Contains(errorMsg, "没有") && strings.Contains(errorMsg, "权限") {
			statusCode = 403
		}

		return &types.RemoveRoleForUserResponse{
			BaseResponse: types.BaseResponse{
				Code: statusCode,
				Msg:  errorMsg,
			},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "permission"),
		logx.Field("operation", "remove_role_for_user"),
		logx.Field("status", "success"),
		logx.Field("user_id", req.UserKey),
		logx.Field("role_key", req.RoleKey),
		logx.Field("tenant_id", req.TenantKey),
	).Info("为用户移除角色成功")

	return &types.RemoveRoleForUserResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "为用户移除角色成功",
		},
	}, nil
}
