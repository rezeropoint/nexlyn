// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package graphconfig

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteGraphConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除逻辑图配置
func NewDeleteGraphConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteGraphConfigLogic {
	return &DeleteGraphConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteGraphConfigLogic) DeleteGraphConfig(req *types.DeleteGraphConfigRequest) (resp *types.DeleteGraphConfigResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_graphconfig"),
		logx.Field("operation", "delete_graphconfig"),
		logx.Field("status", "started"),
		logx.Field("id", req.Id),
	).Info("开始删除逻辑图配置")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "delete_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.DeleteGraphConfigResponse{
			BaseResponse: types.BaseResponse{
				Code:    401,
				Message: "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查逻辑图配置删除权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceLynxGraphConfig, Action: casbinxcore.ActionDelete},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "delete_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.DeleteGraphConfigResponse{
			BaseResponse: types.BaseResponse{Code: 500, Message: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "delete_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.DeleteGraphConfigResponse{
			BaseResponse: types.BaseResponse{Code: 403, Message: "权限不足：需要逻辑图配置删除权限"},
		}, nil
	}

	// 调用Manager删除逻辑图配置
	err = l.svcCtx.LynxManager.DeleteGraphConfig(l.ctx, core.GraphKey{ID: req.Id})
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "delete_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("Manager删除逻辑图配置失败")

		return &types.DeleteGraphConfigResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "服务器错误: " + err.Error(),
			},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_graphconfig"),
		logx.Field("operation", "delete_graphconfig"),
		logx.Field("status", "success"),
		logx.Field("id", req.Id),
	).Info("删除逻辑图配置成功")

	return &types.DeleteGraphConfigResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
	}, nil
}
