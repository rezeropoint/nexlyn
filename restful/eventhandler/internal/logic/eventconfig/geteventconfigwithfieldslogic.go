package eventconfig

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetEventConfigWithFieldsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取事件配置详情（包含字段）
func NewGetEventConfigWithFieldsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEventConfigWithFieldsLogic {
	return &GetEventConfigWithFieldsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetEventConfigWithFieldsLogic) GetEventConfigWithFields(req *types.GetEventConfigWithFieldsRequest) (resp *types.GetEventConfigWithFieldsResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "skylark_event"),
		logx.Field("operation", "get_event_config_with_fields"),
		logx.Field("status", "started"),
		logx.Field("event_id", req.Id),
	).Info("开始获取事件配置详情")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "get_event_config_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetEventConfigWithFieldsResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查事件配置读取权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceEventConfig, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "get_event_config_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetEventConfigWithFieldsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "get_event_config_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetEventConfigWithFieldsResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要事件配置查看权限"},
		}, nil
	}

	// 调用Skylark引擎获取事件配置
	eventCfg, err := l.svcCtx.SkylarkEngine.GetEventWithFields(l.ctx, req.Id, jwtUser.TenantId)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "get_event_config_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("event_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("Skylark引擎获取事件配置失败")

		return &types.GetEventConfigWithFieldsResponse{
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
		logx.Field("operation", "get_event_config_with_fields"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", jwtUser.TenantId),
		logx.Field("event_id", req.Id),
		logx.Field("event_name", eventCfg.EventConfig.Name),
		logx.Field("fields_count", len(eventCfg.Fields)),
	).Info("获取事件配置详情成功")

	return &types.GetEventConfigWithFieldsResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: svc.ConvertCoreEventConfigWithFieldsToTypes(eventCfg),
	}, nil
}
