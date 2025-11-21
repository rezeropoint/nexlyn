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

type AbortJourneyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 终止流程
func NewAbortJourneyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AbortJourneyLogic {
	return &AbortJourneyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AbortJourneyLogic) AbortJourney(req *types.AbortJourneyRequest) (resp *types.AbortJourneyResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "flow_journey"),
		logx.Field("operation", "abort_journey"),
		logx.Field("status", "started"),
		logx.Field("flow_id", req.FlowId),
		logx.Field("journey_id", req.JourneyId),
	).Info("开始终止流程")

	// 1. 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "flow_journey"),
			logx.Field("operation", "abort_journey"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.AbortJourneyResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 2. 权限验证（需要 flow:delete 权限 - 终止流程需要删除权限）
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		core.Permission{Resource: auth.ResourceFlow, Action: core.ActionDelete},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "flow_journey"),
			logx.Field("operation", "abort_journey"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("user_id", jwtUser.UserId),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.AbortJourneyResponse{
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
			logx.Field("operation", "abort_journey"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("user_id", jwtUser.UserId),
			logx.Field("reason", "permission_denied"),
			logx.Field("required_permission", "flow:delete"),
		).Error("权限不足: 缺少 flow:delete 权限")

		return &types.AbortJourneyResponse{
			BaseResponse: types.BaseResponse{
				Code: 403,
				Msg:  "权限不足: 缺少 flow:delete 权限",
			},
		}, nil
	}

	// 3. 调用Skylark引擎终止流程
	err = l.svcCtx.SkylarkEngine.AbortJourney(
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
			logx.Field("operation", "abort_journey"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("flow_id", req.FlowId),
			logx.Field("journey_id", req.JourneyId),
			logx.Field("error", err.Error()),
		).Error("终止流程失败")

		return &types.AbortJourneyResponse{
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
		logx.Field("module", "flow_journey"),
		logx.Field("operation", "abort_journey"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", jwtUser.TenantId),
		logx.Field("flow_id", req.FlowId),
		logx.Field("journey_id", req.JourneyId),
		logx.Field("operator", jwtUser.UserId),
	).Info("终止流程成功")

	// 4. 返回响应
	return &types.AbortJourneyResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
	}, nil
}
