package platform

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetFlowListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取远程Skylark流程列表
func NewGetFlowListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFlowListLogic {
	return &GetFlowListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFlowListLogic) GetFlowList(req *types.GetFlowListRequest) (resp *types.GetFlowListResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "skylark_platform"),
		logx.Field("operation", "get_flow_list"),
		logx.Field("status", "started"),
	).Info("开始获取Skylark流程列表")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_platform"),
			logx.Field("operation", "get_flow_list"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetFlowListResponse{
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
			logx.Field("operation", "get_flow_list"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetFlowListResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_platform"),
			logx.Field("operation", "get_flow_list"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetFlowListResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要Skylark平台配置查看权限"},
		}, nil
	}

	// 调用Skylark引擎获取流程列表
	flows, err := l.svcCtx.SkylarkEngine.GetFlowList(l.ctx, jwtUser.TenantId)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_platform"),
			logx.Field("operation", "get_flow_list"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("error", err.Error()),
		).Error("Skylark引擎获取流程列表失败")

		return &types.GetFlowListResponse{
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
		logx.Field("module", "skylark_platform"),
		logx.Field("operation", "get_flow_list"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", jwtUser.TenantId),
		logx.Field("flows_count", len(flows)),
	).Info("获取Skylark流程列表成功")

	return &types.GetFlowListResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: struct {
			List []types.FlowInfo `json:"list"`
		}{
			List: svc.ConvertCoreFlowInfoListToTypes(flows),
		},
	}, nil
}
