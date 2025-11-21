// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package userassignments

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProposedJourneysLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取用户发起的流程
func NewGetProposedJourneysLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProposedJourneysLogic {
	return &GetProposedJourneysLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProposedJourneysLogic) GetProposedJourneys(req *types.GetProposedJourneysRequest) (resp *types.GetProposedJourneysResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user_assignments"),
		logx.Field("operation", "get_proposed_journeys"),
		logx.Field("status", "started"),
		logx.Field("flow_id", req.FlowId),
		logx.Field("page", req.Page),
		logx.Field("page_size", req.PageSize),
	).Info("开始获取用户发起的流程列表")

	// 1. 从JWT获取用户信息（无需额外权限验证，用户只能查自己发起的流程）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user_assignments"),
			logx.Field("operation", "get_proposed_journeys"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetProposedJourneysResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 2. 调用Skylark引擎获取用户发起的流程列表（按创建时间倒序）
	// SDK会自动处理：
	//   - localUserID到remoteUserID的映射
	//   - 根据tenantID获取平台配置
	//   - 发起人信息的自动补充
	journeys, total, err := l.svcCtx.SkylarkEngine.GetProposedJourneys(
		l.ctx,
		jwtUser.TenantId,
		req.FlowId,     // 流程ID
		jwtUser.UserId, // localUserID - SDK会自动转换为remoteUserID
		req.Page,
		req.PageSize,
	)

	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user_assignments"),
			logx.Field("operation", "get_proposed_journeys"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("user_id", jwtUser.UserId),
			logx.Field("flow_id", req.FlowId),
			logx.Field("error", err.Error()),
		).Error("获取用户发起的流程列表失败")

		return &types.GetProposedJourneysResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 3. 转换结果
	journeyList := svc.ConvertCoreJourneyListToTypes(journeys)

	// 4. 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user_assignments"),
		logx.Field("operation", "get_proposed_journeys"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", jwtUser.TenantId),
		logx.Field("user_id", jwtUser.UserId),
		logx.Field("flow_id", req.FlowId),
		logx.Field("count", len(journeyList)),
		logx.Field("total", total),
	).Info("获取用户发起的流程列表成功")

	return &types.GetProposedJourneysResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: types.JourneyListData{
			List:  journeyList,
			Total: total,
		},
	}, nil
}
