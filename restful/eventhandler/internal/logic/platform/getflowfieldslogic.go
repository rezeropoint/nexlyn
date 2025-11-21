package platform

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetFlowFieldsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取远程Skylark流程字段列表
func NewGetFlowFieldsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFlowFieldsLogic {
	return &GetFlowFieldsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFlowFieldsLogic) GetFlowFields(req *types.GetFlowFieldsRequest) (resp *types.GetFlowFieldsResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "skylark_platform"),
		logx.Field("operation", "get_flow_fields"),
		logx.Field("status", "started"),
		logx.Field("flow_id", req.FlowId),
	).Info("开始获取Skylark流程字段列表")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_platform"),
			logx.Field("operation", "get_flow_fields"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetFlowFieldsResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查平台配置读取权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceSkylarkPlatform, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_platform"),
			logx.Field("operation", "get_flow_fields"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetFlowFieldsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_platform"),
			logx.Field("operation", "get_flow_fields"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetFlowFieldsResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要Skylark平台配置查看权限"},
		}, nil
	}

	// 调用Skylark引擎获取流程字段列表
	fields, err := l.svcCtx.SkylarkEngine.GetFlowFields(l.ctx, jwtUser.TenantId, req.FlowId)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_platform"),
			logx.Field("operation", "get_flow_fields"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("flow_id", req.FlowId),
			logx.Field("error", err.Error()),
		).Error("Skylark引擎获取流程字段列表失败")

		return &types.GetFlowFieldsResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 分离系统字段和业务字段
	systemFields, businessFields := svc.ConvertCoreFieldMetadataListToTypes(fields)

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "skylark_platform"),
		logx.Field("operation", "get_flow_fields"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", jwtUser.TenantId),
		logx.Field("flow_id", req.FlowId),
		logx.Field("system_fields_count", len(systemFields)),
		logx.Field("business_fields_count", len(businessFields)),
	).Info("获取Skylark流程字段列表成功")

	return &types.GetFlowFieldsResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: struct {
			SystemFields   []types.FieldMetadata `json:"systemFields"`
			BusinessFields []types.FieldMetadata `json:"businessFields"`
		}{
			SystemFields:   systemFields,
			BusinessFields: businessFields,
		},
	}, nil
}
