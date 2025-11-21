package permission

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckPermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 检查用户权限
func NewCheckPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckPermissionLogic {
	return &CheckPermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckPermissionLogic) CheckPermission(req *types.CheckPermissionRequest) (resp *types.CheckPermissionResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "permission"),
		logx.Field("operation", "check_permission"),
		logx.Field("status", "started"),
		logx.Field("user_id", req.UserKey),
		logx.Field("tenant_id", req.TenantKey),
		logx.Field("resource", req.Resource),
		logx.Field("action", req.Action),
	).Info("开始检查用户权限")

	// 从JWT中获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "permission"),
			logx.Field("operation", "check_permission"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从JWT获取用户信息失败")

		return &types.CheckPermissionResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
			HasPermission: false,
		}, nil
	}

	// 检查权限 - 将字符串Action转换为Action常量
	action, err := core.ParseAction(req.Action)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "permission"),
			logx.Field("operation", "check_permission"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("解析Action失败")

		return &types.CheckPermissionResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "无效的操作类型",
			},
			HasPermission: false,
		}, nil
	}

	resource, err := core.ParseSystemResource(req.Resource)
	if err != nil {
		return &types.CheckPermissionResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "无效的资源类型",
			},
		}, nil
	}

	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(req.UserKey, req.TenantKey, core.Permission{Resource: resource, Action: action})
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "permission"),
			logx.Field("operation", "check_permission"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("检查权限失败")

		return &types.CheckPermissionResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  config.FormatError(config.ErrMsgPermissionCheck, err),
			},
			HasPermission: false,
		}, nil
	}

	// 记录结果
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "permission"),
		logx.Field("operation", "check_permission"),
		logx.Field("status", "success"),
		logx.Field("user_id", req.UserKey),
		logx.Field("tenant_id", req.TenantKey),
		logx.Field("resource", req.Resource),
		logx.Field("action", req.Action),
		logx.Field("has_permission", hasPermission),
		logx.Field("operator_user_id", jwtUser.UserId),
	).Info("检查用户权限完成")

	return &types.CheckPermissionResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "检查权限完成",
		},
		HasPermission: hasPermission,
	}, nil
}
