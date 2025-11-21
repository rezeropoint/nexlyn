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

type QueryEventDataLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询事件数据列表（动态字段，Journey聚合，返回最新Assignment）
func NewQueryEventDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryEventDataLogic {
	return &QueryEventDataLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryEventDataLogic) QueryEventData(req *types.QueryEventDataRequest) (resp *types.QueryEventDataResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "event_data"),
		logx.Field("operation", "query_event_data"),
		logx.Field("status", "started"),
		logx.Field("event_id", req.EventId),
		logx.Field("page", req.Page),
		logx.Field("page_size", req.PageSize),
	).Info("开始查询事件数据列表")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_data"),
			logx.Field("operation", "query_event_data"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.QueryEventDataResponse{
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
			logx.Field("operation", "query_event_data"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.QueryEventDataResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_data"),
			logx.Field("operation", "query_event_data"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.QueryEventDataResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要事件数据查询权限"},
		}, nil
	}

	// 获取用户所属组织ID
	userOrgID := jwtUser.PrimaryOrgId
	if userOrgID == "" {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_data"),
			logx.Field("operation", "query_event_data"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户组织信息缺失")

		return &types.QueryEventDataResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "用户组织信息缺失"},
		}, nil
	}

	// 验证用户是否有权访问指定组织
	hasAccess, err := auth.ValidateOrgAccess(l.ctx, l.svcCtx.DBConn, req.OrgId, userOrgID, jwtUser.TenantId)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_data"),
			logx.Field("operation", "query_event_data"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("验证组织访问权限失败")

		return &types.QueryEventDataResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "验证组织权限失败: " + err.Error()},
		}, nil
	}

	if !hasAccess {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_data"),
			logx.Field("operation", "query_event_data"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("target_org_id", req.OrgId),
			logx.Field("user_org_id", userOrgID),
		).Error("用户无权访问指定组织")

		return &types.QueryEventDataResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "无权访问指定组织"},
		}, nil
	}

	// 查询指定组织及其所有子组织
	userOrgIDs, err := auth.GetAllChildOrgIds(l.ctx, l.svcCtx.DBConn, req.OrgId, jwtUser.TenantId)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_data"),
			logx.Field("operation", "query_event_data"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询组织失败")

		return &types.QueryEventDataResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "查询组织失败: " + err.Error()},
		}, nil
	}

	// 构建查询请求（只传EventConfigID，引擎内部加载配置）
	queryReq := &core.QueryRequest{
		EventConfigID: req.EventId,
		TenantID:      jwtUser.TenantId,
		UserOrgIDs:    userOrgIDs,
		Page:          req.Page,
		PageSize:      req.PageSize,
		Keyword:       req.Keyword,
		SortField:     req.SortField,
		SortOrder:     req.SortOrder,
		Status:        req.Status, // 状态筛选（可选，支持多个状态）
	}

	// 设置默认值
	if queryReq.Page <= 0 {
		queryReq.Page = 1
	}
	if queryReq.PageSize <= 0 {
		queryReq.PageSize = 20
	}
	if queryReq.PageSize > 100 {
		queryReq.PageSize = 100 // 限制最大每页数量
	}

	// 调用Skylark引擎查询事件数据
	queryResp, err := l.svcCtx.SkylarkEngine.QueryEventData(l.ctx, queryReq)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "event_data"),
			logx.Field("operation", "query_event_data"),
			logx.Field("status", "failed"),
			logx.Field("event_id", req.EventId),
			logx.Field("error", err.Error()),
		).Error("Skylark引擎查询事件数据失败")

		return &types.QueryEventDataResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 转换列信息
	columns := make([]types.ColumnInfo, 0, len(queryResp.Columns))
	for _, col := range queryResp.Columns {
		columns = append(columns, types.ColumnInfo{
			Field:       col.Field,
			DisplayName: col.DisplayName,
			Type:        col.Type,
		})
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "event_data"),
		logx.Field("operation", "query_event_data"),
		logx.Field("status", "success"),
		logx.Field("event_id", req.EventId),
		logx.Field("total", queryResp.Total),
		logx.Field("records", len(queryResp.Records)),
	).Info("查询事件数据成功")

	return &types.QueryEventDataResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: struct {
			Columns []types.ColumnInfo       `json:"columns"`
			Records []map[string]interface{} `json:"records"`
			Total   int64                    `json:"total"`
		}{
			Columns: columns,
			Records: queryResp.Records, // 直接使用引擎返回的动态数据
			Total:   queryResp.Total,
		},
	}, nil
}
