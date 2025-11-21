package eventdata

import (
	"context"

	"github.com/rezeropoint/go-skylark/v2/core"
	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetEventDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取事件详情（Journey的所有Assignment，用于详情页展示流转历史）
func NewGetEventDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEventDetailLogic {
	return &GetEventDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetEventDetailLogic) GetEventDetail(req *types.GetEventDetailRequest) (resp *types.GetEventDetailResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "event_data"),
		logx.Field("operation", "get_event_detail"),
		logx.Field("status", "started"),
		logx.Field("event_id", req.EventId),
		logx.Field("journey_id", req.JourneyId),
	).Info("开始获取事件详情")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_data"),
			logx.Field("operation", "get_event_detail"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetEventDetailResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查事件数据查询权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceEventData, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_data"),
			logx.Field("operation", "get_event_detail"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetEventDetailResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_data"),
			logx.Field("operation", "get_event_detail"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetEventDetailResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要事件数据查询权限"},
		}, nil
	}

	// 获取用户所属组织列表（包括所有子组织）
	// 从JWT中获取用户的主要组织ID
	userOrgID := jwtUser.PrimaryOrgId
	if userOrgID == "" {
		// 如果JWT中没有组织ID，尝试从请求头获取
		userOrgID = l.ctx.Value("orgID").(string)
		if userOrgID == "" {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "event_data"),
				logx.Field("operation", "get_event_detail"),
				logx.Field("status", "failed"),
				logx.Field("user_key", jwtUser.UserKey),
			).Error("用户组织信息缺失")

			return &types.GetEventDetailResponse{
				BaseResponse: types.BaseResponse{Code: 400, Msg: "用户组织信息缺失"},
			}, nil
		}
	}

	// 查询用户组织权限范围（包括所有子组织）
	userOrgIDs, err := auth.GetAllChildOrgIds(l.ctx, l.svcCtx.DBConn, userOrgID, jwtUser.TenantId)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_data"),
			logx.Field("operation", "get_event_detail"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询组织权限范围失败")

		return &types.GetEventDetailResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "查询组织失败: " + err.Error()},
		}, nil
	}

	// 构建详情查询请求
	detailReq := &core.DetailRequest{
		EventConfigID: req.EventId,
		JourneyID:     req.JourneyId,
		TenantID:      jwtUser.TenantId,
		UserOrgIDs:    userOrgIDs,
	}

	// 调用Skylark引擎查询事件详情（引擎自动完成用户名转换）
	detailResp, err := l.svcCtx.SkylarkEngine.GetEventDetail(l.ctx, detailReq)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_data"),
			logx.Field("operation", "get_event_detail"),
			logx.Field("status", "failed"),
			logx.Field("event_id", req.EventId),
			logx.Field("journey_id", req.JourneyId),
			logx.Field("error", err.Error()),
		).Error("Skylark引擎查询事件详情失败")

		return &types.GetEventDetailResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 转换响应格式
	response := svc.ConvertCoreDetailResponseToTypes(detailResp)

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "event_data"),
		logx.Field("operation", "get_event_detail"),
		logx.Field("status", "success"),
		logx.Field("event_id", req.EventId),
		logx.Field("journey_id", req.JourneyId),
		logx.Field("assignment_count", len(detailResp.FlowHistory)),
	).Info("获取事件详情成功")

	return &response, nil
}
