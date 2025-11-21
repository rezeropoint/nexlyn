// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package flowjourney

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/rezeropoint/go-skylark/v2/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFlowJourneyStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新流程Journey状态（审批操作：通过/回退/转交/撤销）
func NewUpdateFlowJourneyStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFlowJourneyStatusLogic {
	return &UpdateFlowJourneyStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFlowJourneyStatusLogic) UpdateFlowJourneyStatus(req *types.UpdateFlowJourneyStatusRequest) (resp *types.UpdateFlowJourneyStatusResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "flow_journey"),
		logx.Field("operation", "update_journey_status"),
		logx.Field("status", "started"),
		logx.Field("journey_id", req.JourneyId),
		logx.Field("assignment_id", req.AssignmentId),
		logx.Field("operation_type", req.Operation),
	).Info("开始处理流程审批操作")

	// 1. 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "flow_journey"),
			logx.Field("operation", "update_journey_status"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.UpdateFlowJourneyStatusResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 2. 权限验证：检查流程审批权限（需要 flow_journey:write 权限）
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{
			Resource: auth.ResourceFlowJourney,
			Action:   casbinxcore.ActionWrite,
		},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "flow_journey"),
			logx.Field("operation", "update_journey_status"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.UpdateFlowJourneyStatusResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "flow_journey"),
			logx.Field("operation", "update_journey_status"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("reason", "permission_denied"),
			logx.Field("required_permission", "flow_journey:write"),
		).Error("权限不足: 缺少 flow_journey:write 权限")

		return &types.UpdateFlowJourneyStatusResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足: 缺少 flow_journey:write 权限"},
		}, nil
	}

	// 3. 构建更新选项
	options := core.UpdateJourneyStatusOptions{
		Comment:           req.Comment,
		CarbonCopyUserIDs: req.CarbonCopyUserIds,
		Data:              svc.ConvertTypesDataToCoreData(req.Data),
	}

	// 4. 调用Skylark引擎执行审批操作
	// SDK会自动处理：
	//   - localUserID到remoteUserID的映射
	//   - 根据tenantID获取平台配置
	err = l.svcCtx.SkylarkEngine.UpdateFlowJourneyStatus(
		l.ctx,
		jwtUser.TenantId,
		req.FlowId, // 流程ID（由前端提供）
		req.JourneyId,
		req.AssignmentId,
		jwtUser.UserId, // localUserID - SDK会自动转换为remoteUserID
		core.JourneyOperation(req.Operation),
		options,
	)

	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "flow_journey"),
			logx.Field("operation", "update_journey_status"),
			logx.Field("status", "failed"),
			logx.Field("journey_id", req.JourneyId),
			logx.Field("assignment_id", req.AssignmentId),
			logx.Field("error", err.Error()),
		).Error("执行流程审批操作失败")

		return &types.UpdateFlowJourneyStatusResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 5. 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "flow_journey"),
		logx.Field("operation", "update_journey_status"),
		logx.Field("status", "success"),
		logx.Field("journey_id", req.JourneyId),
		logx.Field("assignment_id", req.AssignmentId),
		logx.Field("operation_type", req.Operation),
		logx.Field("local_user_id", jwtUser.UserId),
	).Info("流程审批操作成功")

	return &types.UpdateFlowJourneyStatusResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
	}, nil
}
