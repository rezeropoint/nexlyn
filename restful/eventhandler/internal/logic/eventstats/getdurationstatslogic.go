package eventstats

import (
	"context"
	"time"

	"github.com/rezeropoint/go-skylark/v2/core"
	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDurationStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取处理时长统计
func NewGetDurationStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDurationStatsLogic {
	return &GetDurationStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDurationStatsLogic) GetDurationStats(req *types.GetDurationStatsRequest) (resp *types.GetDurationStatsResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "event_stats"),
		logx.Field("operation", "get_duration_stats"),
		logx.Field("status", "started"),
		logx.Field("event_ids", req.EventIds),
		logx.Field("org_id", req.OrgId),
	).Info("开始获取处理时长统计")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_stats"),
			logx.Field("operation", "get_duration_stats"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetDurationStatsResponse{
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
			logx.Field("module", "event_stats"),
			logx.Field("operation", "get_duration_stats"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetDurationStatsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_stats"),
			logx.Field("operation", "get_duration_stats"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetDurationStatsResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要事件数据查询权限"},
		}, nil
	}

	// 验证组织访问权限（检查请求的orgId是否在用户权限范围内）
	userOrgID := jwtUser.PrimaryOrgId
	if userOrgID == "" {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_stats"),
			logx.Field("operation", "get_duration_stats"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户组织信息缺失")

		return &types.GetDurationStatsResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "用户组织信息缺失"},
		}, nil
	}

	// 验证用户是否有权访问请求的组织
	hasOrgAccess, err := auth.ValidateOrgAccess(l.ctx, l.svcCtx.DBConn, req.OrgId, userOrgID, jwtUser.TenantId)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_stats"),
			logx.Field("operation", "get_duration_stats"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("验证组织访问权限失败")

		return &types.GetDurationStatsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "验证组织权限失败: " + err.Error()},
		}, nil
	}

	if !hasOrgAccess {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_stats"),
			logx.Field("operation", "get_duration_stats"),
			logx.Field("status", "failed"),
			logx.Field("user_org_id", userOrgID),
			logx.Field("requested_org_id", req.OrgId),
		).Error("用户无权访问请求的组织")

		return &types.GetDurationStatsResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：无权访问指定组织的数据"},
		}, nil
	}

	// 查询用户组织权限范围（包括所有子组织）
	userOrgIDs, err := auth.GetAllChildOrgIds(l.ctx, l.svcCtx.DBConn, req.OrgId, jwtUser.TenantId)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_stats"),
			logx.Field("operation", "get_duration_stats"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询组织权限范围失败")

		return &types.GetDurationStatsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "查询组织失败: " + err.Error()},
		}, nil
	}

	// 解析时间参数
	var dateFrom, dateTo *time.Time
	if req.DateFrom != "" {
		t, err := time.Parse(time.RFC3339, req.DateFrom)
		if err != nil {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "event_stats"),
				logx.Field("operation", "get_duration_stats"),
				logx.Field("status", "failed"),
				logx.Field("date_from", req.DateFrom),
				logx.Field("error", err.Error()),
			).Error("开始日期格式错误")

			return &types.GetDurationStatsResponse{
				BaseResponse: types.BaseResponse{Code: 400, Msg: "开始日期格式错误，需要ISO 8601格式"},
			}, nil
		}
		dateFrom = &t
	}

	if req.DateTo != "" {
		t, err := time.Parse(time.RFC3339, req.DateTo)
		if err != nil {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "event_stats"),
				logx.Field("operation", "get_duration_stats"),
				logx.Field("status", "failed"),
				logx.Field("date_to", req.DateTo),
				logx.Field("error", err.Error()),
			).Error("结束日期格式错误")

			return &types.GetDurationStatsResponse{
				BaseResponse: types.BaseResponse{Code: 400, Msg: "结束日期格式错误，需要ISO 8601格式"},
			}, nil
		}
		dateTo = &t
	}

	// 构建统计查询请求
	statsReq := &core.StatsCriteria{
		EventConfigIDs: req.EventIds,
		TenantID:       jwtUser.TenantId,
		UserOrgIDs:     userOrgIDs,
		DateFrom:       dateFrom,
		DateTo:         dateTo,
		Status:         req.Status,
	}

	// 调用Skylark引擎查询处理时长统计
	statsResp, err := l.svcCtx.SkylarkEngine.GetDurationStats(l.ctx, statsReq)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_stats"),
			logx.Field("operation", "get_duration_stats"),
			logx.Field("status", "failed"),
			logx.Field("event_ids", req.EventIds),
			logx.Field("error", err.Error()),
		).Error("Skylark引擎查询处理时长统计失败")

		return &types.GetDurationStatsResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 转换响应格式
	response := svc.ConvertCoreDurationStatsToTypes(statsResp)

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "event_stats"),
		logx.Field("operation", "get_duration_stats"),
		logx.Field("status", "success"),
		logx.Field("event_ids", req.EventIds),
		logx.Field("total_count", statsResp.TotalCount),
		logx.Field("completed_count", statsResp.CompletedCount),
	).Info("获取处理时长统计成功")

	return &response, nil
}
