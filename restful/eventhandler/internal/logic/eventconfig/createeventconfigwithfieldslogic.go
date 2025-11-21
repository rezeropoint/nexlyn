package eventconfig

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	"github.com/rezeropoint/go-skylark/v2/core"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateEventConfigWithFieldsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建事件配置（包含字段，事务保证）
func NewCreateEventConfigWithFieldsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateEventConfigWithFieldsLogic {
	return &CreateEventConfigWithFieldsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateEventConfigWithFieldsLogic) CreateEventConfigWithFields(req *types.CreateEventConfigWithFieldsRequest) (resp *types.CreateEventConfigWithFieldsResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "skylark_event"),
		logx.Field("operation", "create_event_config_with_fields"),
		logx.Field("status", "started"),
		logx.Field("event_name", req.Name),
		logx.Field("flow_id", req.FlowId),
	).Info("开始创建事件配置")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "create_event_config_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.CreateEventConfigWithFieldsResponse{
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
			logx.Field("operation", "create_event_config_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.CreateEventConfigWithFieldsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "create_event_config_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.CreateEventConfigWithFieldsResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要事件配置写入权限"},
		}, nil
	}

	// 验证字段配置
	if len(req.Fields) == 0 {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "create_event_config_with_fields"),
			logx.Field("status", "failed"),
		).Error("字段配置不能为空")

		return &types.CreateEventConfigWithFieldsResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "字段配置不能为空"},
		}, nil
	}

	for i, field := range req.Fields {
		if field.FieldName == "" {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "skylark_event"),
				logx.Field("operation", "create_event_config_with_fields"),
				logx.Field("status", "failed"),
				logx.Field("field_index", i),
			).Error("字段名不能为空")

			return &types.CreateEventConfigWithFieldsResponse{
				BaseResponse: types.BaseResponse{Code: 400, Msg: "字段名不能为空"},
			}, nil
		}

		if field.DisplayName == "" {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "skylark_event"),
				logx.Field("operation", "create_event_config_with_fields"),
				logx.Field("status", "failed"),
				logx.Field("field_index", i),
				logx.Field("field_name", field.FieldName),
			).Error("字段显示名称不能为空")

			return &types.CreateEventConfigWithFieldsResponse{
				BaseResponse: types.BaseResponse{Code: 400, Msg: "字段显示名称不能为空"},
			}, nil
		}

		if !core.IsValidFieldType(field.FieldType) {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "skylark_event"),
				logx.Field("operation", "create_event_config_with_fields"),
				logx.Field("status", "failed"),
				logx.Field("field_index", i),
				logx.Field("field_name", field.FieldName),
				logx.Field("field_type", field.FieldType),
			).Error("字段类型无效")

			return &types.CreateEventConfigWithFieldsResponse{
				BaseResponse: types.BaseResponse{Code: 400, Msg: "字段类型无效: " + field.FieldType},
			}, nil
		}
	}

	// 转换为Core请求
	coreReq := svc.ConvertCreateEventRequestToCore(req, jwtUser.TenantId, jwtUser.UserId)

	// 调用Skylark引擎创建事件配置
	eventConfigID, err := l.svcCtx.SkylarkEngine.CreateEventWithFields(l.ctx, coreReq)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "create_event_config_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("event_name", req.Name),
			logx.Field("error", err.Error()),
		).Error("Skylark引擎创建事件配置失败")

		return &types.CreateEventConfigWithFieldsResponse{
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
		logx.Field("operation", "create_event_config_with_fields"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", jwtUser.TenantId),
		logx.Field("event_id", eventConfigID),
		logx.Field("event_name", req.Name),
		logx.Field("fields_count", len(req.Fields)),
	).Info("创建事件配置成功")

	return &types.CreateEventConfigWithFieldsResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		}{
			Id:   eventConfigID,
			Name: req.Name,
		},
	}, nil
}
