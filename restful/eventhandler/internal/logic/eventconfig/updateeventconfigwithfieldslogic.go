package eventconfig

import (
	"context"

	"github.com/rezeropoint/go-skylark/v2/core"
	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateEventConfigWithFieldsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新事件配置（包含字段，完整替换策略）
func NewUpdateEventConfigWithFieldsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateEventConfigWithFieldsLogic {
	return &UpdateEventConfigWithFieldsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateEventConfigWithFieldsLogic) UpdateEventConfigWithFields(req *types.UpdateEventConfigWithFieldsRequest) (resp *types.UpdateEventConfigWithFieldsResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "skylark_event"),
		logx.Field("operation", "update_event_config_with_fields"),
		logx.Field("status", "started"),
		logx.Field("event_id", req.Id),
	).Info("开始更新事件配置")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "update_event_config_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.UpdateEventConfigWithFieldsResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查事件配置写入权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceEventConfig, Action: casbinxcore.ActionWrite},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "update_event_config_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.UpdateEventConfigWithFieldsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "update_event_config_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.UpdateEventConfigWithFieldsResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要事件配置写入权限"},
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
			logx.Field("operation", "update_event_config_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("event_id", req.Id),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("error", err.Error()),
		).Error("查询事件配置失败")

		return &types.UpdateEventConfigWithFieldsResponse{
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
			logx.Field("operation", "update_event_config_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("event_id", req.Id),
			logx.Field("user_tenant_id", jwtUser.TenantId),
			logx.Field("event_tenant_id", existingCfg.EventConfig.TenantID),
		).Error("租户权限不匹配")

		return &types.UpdateEventConfigWithFieldsResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "无权操作该事件配置"},
		}, nil
	}

	// 验证字段配置（如果提供）
	if len(req.Fields) > 0 {
		for i, field := range req.Fields {
			if field.FieldName == "" {
				logx.WithContext(l.ctx).WithFields(
					logx.Field("service", l.svcCtx.Config.RestConf.Name),
					logx.Field("pod", l.svcCtx.PodName),
					logx.Field("module", "skylark_event"),
					logx.Field("operation", "update_event_config_with_fields"),
					logx.Field("status", "failed"),
					logx.Field("field_index", i),
				).Error("字段名不能为空")

				return &types.UpdateEventConfigWithFieldsResponse{
					BaseResponse: types.BaseResponse{Code: 400, Msg: "字段名不能为空"},
				}, nil
			}

			if field.DisplayName == "" {
				logx.WithContext(l.ctx).WithFields(
					logx.Field("service", l.svcCtx.Config.RestConf.Name),
					logx.Field("pod", l.svcCtx.PodName),
					logx.Field("module", "skylark_event"),
					logx.Field("operation", "update_event_config_with_fields"),
					logx.Field("status", "failed"),
					logx.Field("field_index", i),
					logx.Field("field_name", field.FieldName),
				).Error("字段显示名称不能为空")

				return &types.UpdateEventConfigWithFieldsResponse{
					BaseResponse: types.BaseResponse{Code: 400, Msg: "字段显示名称不能为空"},
				}, nil
			}

			if !core.IsValidFieldType(field.FieldType) {
				logx.WithContext(l.ctx).WithFields(
					logx.Field("service", l.svcCtx.Config.RestConf.Name),
					logx.Field("pod", l.svcCtx.PodName),
					logx.Field("module", "skylark_event"),
					logx.Field("operation", "update_event_config_with_fields"),
					logx.Field("status", "failed"),
					logx.Field("field_index", i),
					logx.Field("field_name", field.FieldName),
					logx.Field("field_type", field.FieldType),
				).Error("字段类型无效")

				return &types.UpdateEventConfigWithFieldsResponse{
					BaseResponse: types.BaseResponse{Code: 400, Msg: "字段类型无效: " + field.FieldType},
				}, nil
			}
		}
	}

	// 转换为Core请求（传入现有配置以提取FlowID和FlowTitle）
	coreReq := svc.ConvertUpdateEventRequestToCore(req, req.Id, jwtUser.TenantId, jwtUser.UserId, existingCfg)

	// 调用Skylark引擎更新事件配置
	err = l.svcCtx.SkylarkEngine.UpdateEventWithFields(l.ctx, coreReq)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "update_event_config_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("event_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("Skylark引擎更新事件配置失败")

		return &types.UpdateEventConfigWithFieldsResponse{
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
		logx.Field("operation", "update_event_config_with_fields"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", jwtUser.TenantId),
		logx.Field("event_id", req.Id),
		logx.Field("fields_count", len(req.Fields)),
	).Info("更新事件配置成功")

	return &types.UpdateEventConfigWithFieldsResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
	}, nil
}
