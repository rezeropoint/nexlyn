package eventconfig

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteEventConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除事件配置（软删除）
func NewDeleteEventConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteEventConfigLogic {
	return &DeleteEventConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteEventConfigLogic) DeleteEventConfig(req *types.DeleteEventConfigRequest) (resp *types.DeleteEventConfigResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "skylark_event"),
		logx.Field("operation", "delete_event_config"),
		logx.Field("status", "started"),
		logx.Field("event_id", req.Id),
	).Info("开始删除事件配置")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "delete_event_config"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.DeleteEventConfigResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查事件配置删除权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceEventConfig, Action: casbinxcore.ActionDelete},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "delete_event_config"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.DeleteEventConfigResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "delete_event_config"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.DeleteEventConfigResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要事件配置删除权限"},
		}, nil
	}

	// 先查询事件配置以验证租户权限
	existingCfg, err := l.svcCtx.SkylarkEngine.GetEventWithFields(l.ctx, req.Id, jwtUser.TenantId)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "delete_event_config"),
			logx.Field("status", "failed"),
			logx.Field("event_id", req.Id),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("error", err.Error()),
		).Error("查询事件配置失败")

		return &types.DeleteEventConfigResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 验证租户权限
	if existingCfg.EventConfig.TenantID != jwtUser.TenantId {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "delete_event_config"),
			logx.Field("status", "failed"),
			logx.Field("event_id", req.Id),
			logx.Field("user_tenant_id", jwtUser.TenantId),
			logx.Field("event_tenant_id", existingCfg.EventConfig.TenantID),
		).Error("租户权限不匹配")

		return &types.DeleteEventConfigResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "无权操作该事件配置"},
		}, nil
	}

	// 调用Skylark引擎删除事件配置（软删除）
	err = l.svcCtx.SkylarkEngine.DeleteEvent(l.ctx, req.Id, jwtUser.TenantId)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "delete_event_config"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("event_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("Skylark引擎删除事件配置失败")

		return &types.DeleteEventConfigResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "skylark_event"),
		logx.Field("operation", "delete_event_config"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", jwtUser.TenantId),
		logx.Field("event_id", req.Id),
		logx.Field("event_name", existingCfg.EventConfig.Name),
	).Info("删除事件配置成功")

	return &types.DeleteEventConfigResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
	}, nil
}
