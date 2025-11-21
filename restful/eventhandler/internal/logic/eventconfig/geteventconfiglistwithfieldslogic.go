package eventconfig

import (
	"context"
	"strings"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetEventConfigListWithFieldsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取事件配置列表（包含字段）
func NewGetEventConfigListWithFieldsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEventConfigListWithFieldsLogic {
	return &GetEventConfigListWithFieldsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetEventConfigListWithFieldsLogic) GetEventConfigListWithFields(req *types.GetEventConfigListWithFieldsRequest) (resp *types.GetEventConfigListWithFieldsResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "skylark_event"),
		logx.Field("operation", "get_event_config_list_with_fields"),
		logx.Field("status", "started"),
		logx.Field("current", req.Current),
		logx.Field("page_size", req.PageSize),
		logx.Field("keyword", req.Keyword),
		logx.Field("enabled", req.Enabled),
	).Info("开始获取事件配置列表")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "get_event_config_list_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetEventConfigListWithFieldsResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查事件配置读取权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceEventConfig, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "get_event_config_list_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetEventConfigListWithFieldsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "get_event_config_list_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetEventConfigListWithFieldsResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要事件配置查看权限"},
		}, nil
	}

	// 调用Skylark引擎获取事件配置列表
	eventConfigs, err := l.svcCtx.SkylarkEngine.ListEventWithFields(l.ctx, jwtUser.TenantId, req.Enabled)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_event"),
			logx.Field("operation", "get_event_config_list_with_fields"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("error", err.Error()),
		).Error("Skylark引擎获取事件配置列表失败")

		return &types.GetEventConfigListWithFieldsResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 关键词搜索过滤（匹配事件名称）
	var filteredConfigs []*types.EventConfigWithFields
	if req.Keyword != "" {
		keyword := strings.ToLower(req.Keyword)
		for _, cfg := range eventConfigs {
			if strings.Contains(strings.ToLower(cfg.EventConfig.Name), keyword) {
				converted := svc.ConvertCoreEventConfigWithFieldsToTypes(cfg)
				filteredConfigs = append(filteredConfigs, &converted)
			}
		}
	} else {
		// 无关键词，返回所有配置
		for _, cfg := range eventConfigs {
			converted := svc.ConvertCoreEventConfigWithFieldsToTypes(cfg)
			filteredConfigs = append(filteredConfigs, &converted)
		}
	}

	// 计算总数
	total := int64(len(filteredConfigs))

	// 分页处理
	current := req.Current
	pageSize := req.PageSize
	if current < 1 {
		current = 1
	}
	if pageSize < 1 {
		pageSize = 20 // 默认每页20条
	}

	start := (current - 1) * pageSize
	end := start + pageSize

	// 处理越界情况
	if start >= int64(len(filteredConfigs)) {
		filteredConfigs = []*types.EventConfigWithFields{} // 超出范围返回空列表
	} else {
		if end > int64(len(filteredConfigs)) {
			end = int64(len(filteredConfigs))
		}
		filteredConfigs = filteredConfigs[start:end]
	}

	// 转换为非指针列表
	result := make([]types.EventConfigWithFields, 0, len(filteredConfigs))
	for _, cfg := range filteredConfigs {
		result = append(result, *cfg)
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "skylark_event"),
		logx.Field("operation", "get_event_config_list_with_fields"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", jwtUser.TenantId),
		logx.Field("total", total),
		logx.Field("current", current),
		logx.Field("page_size", pageSize),
	).Info("获取事件配置列表成功")

	return &types.GetEventConfigListWithFieldsResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		PageParams: types.PageParams{
			Current:  current,
			PageSize: pageSize,
			Total:    total,
		},
		Data: struct {
			List []types.EventConfigWithFields `json:"list"`
		}{
			List: result,
		},
	}, nil
}
