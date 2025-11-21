// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package flowjourney

import (
	"context"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/rezeropoint/go-skylark/v2/core"
	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchJourneysLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 高级搜索流程记录
func NewSearchJourneysLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchJourneysLogic {
	return &SearchJourneysLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchJourneysLogic) SearchJourneys(req *types.SearchJourneysRequest) (resp *types.SearchJourneysResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "flow_journey"),
		logx.Field("operation", "search_journeys"),
		logx.Field("status", "started"),
		logx.Field("flow_id", req.FlowId),
		logx.Field("page", req.Page),
		logx.Field("page_size", req.PageSize),
	).Info("开始搜索流程记录")

	// 1. 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "flow_journey"),
			logx.Field("operation", "search_journeys"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.SearchJourneysResponse{
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
		casbinxcore.Permission{Resource: auth.ResourceFlow, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "flow_journey"),
			logx.Field("operation", "search_journeys"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("user_id", jwtUser.UserId),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.SearchJourneysResponse{
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
			logx.Field("operation", "search_journeys"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("user_id", jwtUser.UserId),
			logx.Field("reason", "permission_denied"),
			logx.Field("required_permission", "flow:read"),
		).Error("权限不足: 缺少 flow:read 权限")

		return &types.SearchJourneysResponse{
			BaseResponse: types.BaseResponse{
				Code: 403,
				Msg:  "权限不足: 缺少 flow:read 权限",
			},
		}, nil
	}

	// 3. 构建搜索请求（SDK已改为string类型，直接赋值即可）
	searchReq := &core.JourneySearchRequest{
		FlowID:      req.FlowId,
		Status:      req.Status,      // SDK会忽略空字符串
		Keyword:     req.Keyword,     // SDK会忽略空字符串
		InitiatorID: req.InitiatorId, // 本地用户UUID，SDK会自动转换为远程用户ID
		CreatedFrom: req.CreatedFrom, // SDK会忽略空字符串
		CreatedTo:   req.CreatedTo,   // SDK会忽略空字符串
		Page:        req.Page,
		PageSize:    req.PageSize,
	}

	// 4. 调用Skylark引擎搜索流程
	// SDK会自动将本地用户ID转换为远程用户ID
	journeys, total, err := l.svcCtx.SkylarkEngine.SearchJourneys(
		l.ctx,
		jwtUser.TenantId,
		searchReq,
	)

	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "flow_journey"),
			logx.Field("operation", "search_journeys"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("flow_id", req.FlowId),
			logx.Field("error", err.Error()),
		).Error("搜索流程记录失败")

		return &types.SearchJourneysResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 5. 转换结果
	data := svc.ConvertCoreJourneyListToTypes(journeys)

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "flow_journey"),
		logx.Field("operation", "search_journeys"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", jwtUser.TenantId),
		logx.Field("flow_id", req.FlowId),
		logx.Field("total", total),
	).Info("搜索流程记录成功")

	// 6. 返回响应
	return &types.SearchJourneysResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: struct {
			List  []types.Journey `json:"list"`
			Total int             `json:"total"`
		}{
			List:  data,
			Total: total,
		},
	}, nil
}
