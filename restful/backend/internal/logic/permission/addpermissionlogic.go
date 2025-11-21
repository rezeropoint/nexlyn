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

type AddPermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 添加权限规则
func NewAddPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddPermissionLogic {
	return &AddPermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddPermissionLogic) AddPermission(req *types.AddPermissionRequest) (resp *types.AddPermissionResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "permission"),
		logx.Field("operation", "add_permission"),
		logx.Field("status", "started"),
		logx.Field("user_id", req.UserKey),
		logx.Field("tenant_id", req.TenantKey),
		logx.Field("resource", req.Resource),
		logx.Field("action", req.Action),
	).Info("开始添加权限规则")

	// 从JWT中获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "permission"),
			logx.Field("operation", "add_permission"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从JWT获取用户信息失败")

		return &types.AddPermissionResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
		}, nil
	}

	// 添加权限规则（使用 CasbinX 的 GrantPermission 方法，自动进行安全检查）
	// 将字符串Action转换为Action常量
	action, err := core.ParseAction(req.Action)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "permission"),
			logx.Field("operation", "add_permission"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("解析Action失败")

		return &types.AddPermissionResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "无效的操作类型",
			},
		}, nil
	}

	resource, err := core.ParseSystemResource(req.Resource)
	if err != nil {
		return &types.AddPermissionResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "无效的资源类型",
			},
		}, nil
	}

	err = l.svcCtx.Casbinx.GrantPermission(jwtUser.UserKey, req.UserKey, req.TenantKey, core.Permission{Resource: resource, Action: action})
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "permission"),
			logx.Field("operation", "add_permission"),
			logx.Field("status", "failed"),
			logx.Field("user_id", req.UserKey),
			logx.Field("tenant_id", req.TenantKey),
			logx.Field("resource", req.Resource),
			logx.Field("action", req.Action),
			logx.Field("error", err.Error()),
		).Error("添加权限规则失败")

		return &types.AddPermissionResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  config.FormatError(config.ErrMsgPermissionCheck, err),
			},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "permission"),
		logx.Field("operation", "add_permission"),
		logx.Field("status", "success"),
		logx.Field("user_id", req.UserKey),
		logx.Field("tenant_id", req.TenantKey),
		logx.Field("resource", req.Resource),
		logx.Field("action", req.Action),
	).Info("添加权限规则成功")

	return &types.AddPermissionResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "添加权限规则成功",
		},
	}, nil
}
