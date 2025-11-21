// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package flowmetadata

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetFlowDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取流程元数据（字段、节点、边）
func NewGetFlowDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFlowDetailLogic {
	return &GetFlowDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFlowDetailLogic) GetFlowDetail(req *types.GetFlowDetailRequest) (resp *types.GetFlowDetailResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "flow_metadata"),
		logx.Field("operation", "get_flow_detail"),
		logx.Field("status", "started"),
		logx.Field("flow_id", req.FlowId),
	).Info("开始获取流程元数据")

	// 1. 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "flow_metadata"),
			logx.Field("operation", "get_flow_detail"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetFlowDetailResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 2. 权限验证：检查flow:read权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{
			Resource: auth.ResourceFlow,
			Action:   casbinxcore.ActionRead,
		},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "flow_metadata"),
			logx.Field("operation", "get_flow_detail"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetFlowDetailResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "flow_metadata"),
			logx.Field("operation", "get_flow_detail"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetFlowDetailResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要流程查询权限"},
		}, nil
	}

	// 3. 调用Skylark引擎获取流程元数据
	// SDK会自动处理：
	//   - 根据tenantID获取平台配置
	//   - 查询Flow表获取字段、节点、边信息
	flowDetail, err := l.svcCtx.SkylarkEngine.GetFlowDetail(
		l.ctx,
		jwtUser.TenantId,
		req.FlowId,
	)

	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "flow_metadata"),
			logx.Field("operation", "get_flow_detail"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("flow_id", req.FlowId),
			logx.Field("error", err.Error()),
		).Error("获取流程元数据失败")

		return &types.GetFlowDetailResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 4. 转换结果
	detailData := svc.ConvertCoreFlowDetailToTypes(flowDetail)

	// 5. 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "flow_metadata"),
		logx.Field("operation", "get_flow_detail"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", jwtUser.TenantId),
		logx.Field("flow_id", req.FlowId),
		logx.Field("flow_title", detailData.Title),
		logx.Field("fields_count", len(detailData.Fields)),
		logx.Field("vertices_count", len(detailData.Vertices)),
		logx.Field("edges_count", len(detailData.Edges)),
	).Info("获取流程元数据成功")

	return &types.GetFlowDetailResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: detailData,
	}, nil
}
