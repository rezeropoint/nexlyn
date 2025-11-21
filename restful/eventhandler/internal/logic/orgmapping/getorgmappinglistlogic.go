package orgmapping

import (
	"context"
	"strings"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrgMappingListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取组织映射列表（租户的所有映射）
func NewGetOrgMappingListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrgMappingListLogic {
	return &GetOrgMappingListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrgMappingListLogic) GetOrgMappingList(req *types.GetOrgMappingListRequest) (resp *types.GetOrgMappingListResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "org_mapping"),
		logx.Field("operation", "get_org_mapping_list"),
		logx.Field("status", "started"),
		logx.Field("current", req.Current),
		logx.Field("page_size", req.PageSize),
		logx.Field("keyword", req.Keyword),
	).Info("开始获取组织映射列表")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "get_org_mapping_list"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetOrgMappingListResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查组织映射读取权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceOrgMapping, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "get_org_mapping_list"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetOrgMappingListResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "get_org_mapping_list"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetOrgMappingListResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要组织映射查看权限"},
		}, nil
	}

	// 调用Skylark引擎获取租户的所有组织映射
	mappings, err := l.svcCtx.SkylarkEngine.ListOrgMappings(l.ctx, jwtUser.TenantId)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "get_org_mapping_list"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("error", err.Error()),
		).Error("Skylark引擎获取组织映射列表失败")

		return &types.GetOrgMappingListResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 关键词搜索过滤（匹配 RemoteOrgValue 或 LocalOrgName）
	var filteredMappings []*types.OrgMapping
	if req.Keyword != "" {
		keyword := strings.ToLower(req.Keyword)
		for _, mapping := range mappings {
			if strings.Contains(strings.ToLower(mapping.RemoteOrgValue), keyword) ||
				strings.Contains(strings.ToLower(mapping.LocalOrgName), keyword) {
				converted := svc.ConvertCoreOrgMappingToTypes(mapping)
				filteredMappings = append(filteredMappings, &converted)
			}
		}
	} else {
		// 无关键词，返回所有映射
		for _, mapping := range mappings {
			converted := svc.ConvertCoreOrgMappingToTypes(mapping)
			filteredMappings = append(filteredMappings, &converted)
		}
	}

	// 计算总数
	total := int64(len(filteredMappings))

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
	if start >= int64(len(filteredMappings)) {
		filteredMappings = []*types.OrgMapping{} // 超出范围返回空列表
	} else {
		if end > int64(len(filteredMappings)) {
			end = int64(len(filteredMappings))
		}
		filteredMappings = filteredMappings[start:end]
	}

	// 转换为非指针列表
	result := make([]types.OrgMapping, 0, len(filteredMappings))
	for _, mapping := range filteredMappings {
		result = append(result, *mapping)
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "org_mapping"),
		logx.Field("operation", "get_org_mapping_list"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", jwtUser.TenantId),
		logx.Field("total", total),
		logx.Field("current", current),
		logx.Field("page_size", pageSize),
	).Info("获取组织映射列表成功")

	return &types.GetOrgMappingListResponse{
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
			List []types.OrgMapping `json:"list"`
		}{
			List: result,
		},
	}, nil
}
