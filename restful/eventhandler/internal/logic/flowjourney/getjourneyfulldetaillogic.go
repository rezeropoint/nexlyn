// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package flowjourney

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	"github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetJourneyFullDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取流程完整详情（一站式接口）
func NewGetJourneyFullDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetJourneyFullDetailLogic {
	return &GetJourneyFullDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetJourneyFullDetailLogic) GetJourneyFullDetail(req *types.GetJourneyFullDetailRequest) (resp *types.GetJourneyFullDetailResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "flow_journey"),
		logx.Field("operation", "get_journey_full_detail"),
		logx.Field("status", "started"),
		logx.Field("flow_id", req.FlowId),
		logx.Field("journey_id", req.JourneyId),
	).Info("开始获取流程完整详情")

	// 1. 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "flow_journey"),
			logx.Field("operation", "get_journey_full_detail"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetJourneyFullDetailResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 2. 权限验证（需要 flow:read 权限）
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		core.Permission{Resource: auth.ResourceFlow, Action: core.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "flow_journey"),
			logx.Field("operation", "get_journey_full_detail"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("user_id", jwtUser.UserId),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetJourneyFullDetailResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "权限检查失败: " + err.Error(),
			},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "flow_journey"),
			logx.Field("operation", "get_journey_full_detail"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("user_id", jwtUser.UserId),
			logx.Field("reason", "permission_denied"),
			logx.Field("required_permission", "flow:read"),
		).Error("权限不足: 缺少 flow:read 权限")

		return &types.GetJourneyFullDetailResponse{
			BaseResponse: types.BaseResponse{
				Code: 403,
				Msg:  "权限不足: 缺少 flow:read 权限",
			},
		}, nil
	}

	// 3. 调用Skylark引擎获取流程完整详情
	fullDetail, err := l.svcCtx.SkylarkEngine.GetJourneyFullDetail(
		l.ctx,
		jwtUser.TenantId,
		req.FlowId,
		req.JourneyId,
	)

	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "flow_journey"),
			logx.Field("operation", "get_journey_full_detail"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("flow_id", req.FlowId),
			logx.Field("journey_id", req.JourneyId),
			logx.Field("error", err.Error()),
		).Error("获取流程完整详情失败")

		return &types.GetJourneyFullDetailResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 4. 转换结果
	data := svc.ConvertCoreJourneyFullDetailToTypes(fullDetail)

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "flow_journey"),
		logx.Field("operation", "get_journey_full_detail"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", jwtUser.TenantId),
		logx.Field("flow_id", req.FlowId),
		logx.Field("journey_id", req.JourneyId),
		logx.Field("journey_status", fullDetail.BasicInfo.Status),
		logx.Field("history_count", len(fullDetail.History)),
		logx.Field("pending_nodes_count", len(fullDetail.PendingNodes)),
	).Info("获取流程完整详情成功")

	// 5. 返回响应
	return &types.GetJourneyFullDetailResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: data,
	}, nil
}
