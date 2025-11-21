// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package graphconfig

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetGraphConfigListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询逻辑图配置列表
func NewGetGraphConfigListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGraphConfigListLogic {
	return &GetGraphConfigListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGraphConfigListLogic) GetGraphConfigList(req *types.GetGraphConfigListRequest) (resp *types.GetGraphConfigListResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_graphconfig"),
		logx.Field("operation", "list_graphconfigs"),
		logx.Field("status", "started"),
		logx.Field("page", req.Page),
		logx.Field("page_size", req.PageSize),
		logx.Field("tenant_id", req.TenantId),
	).Info("开始获取逻辑图配置列表")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "list_graphconfigs"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetGraphConfigListResponse{
			BaseResponse: types.BaseResponse{
				Code:    401,
				Message: "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查逻辑图配置查看权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceLynxGraphConfig, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "list_graphconfigs"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetGraphConfigListResponse{
			BaseResponse: types.BaseResponse{Code: 500, Message: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "list_graphconfigs"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetGraphConfigListResponse{
			BaseResponse: types.BaseResponse{Code: 403, Message: "权限不足：需要逻辑图配置查看权限"},
		}, nil
	}

	// 规范化分页参数
	page, pageSize, _ := svc.NormalizePagination(req.Page, req.PageSize)

	// 查询用户可见的组织ID列表（本组织+所有子组织）
	orgIds, err := auth.GetAllChildOrgIds(l.ctx, l.svcCtx.DBConn, jwtUser.PrimaryOrgId, jwtUser.TenantId)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "list_graphconfigs"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询用户可见组织列表失败")

		return &types.GetGraphConfigListResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "服务器错误: 查询组织权限失败",
			},
		}, nil
	}

	// 构建查询条件（传递完整的可见组织ID列表给Manager层）
	query := core.GraphList{
		Page:     int64(page),
		PageSize: int64(pageSize),
		TenantId: req.TenantId,
		OrgIDs:   orgIds, // 传递完整的组织ID列表（包含所有子组织）
		Name:     req.Name,
		Tags:     req.Tags,
		Enable:   req.IsEnabled,
	}

	// 调用Manager查询逻辑图配置列表
	graphConfigs, total, _, _, err := l.svcCtx.LynxManager.GetGraphConfigList(l.ctx, query)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "list_graphconfigs"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("Manager查询逻辑图配置列表失败")

		return &types.GetGraphConfigListResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "服务器错误: " + err.Error(),
			},
		}, nil
	}

	// Phase 2 重构：转换为API类型（只返回元数据，包含标签名称）
	list := l.svcCtx.ConvertCoreGraphConfigsToMetadata(l.ctx, graphConfigs)

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_graphconfig"),
		logx.Field("operation", "list_graphconfigs"),
		logx.Field("status", "success"),
		logx.Field("total", total),
		logx.Field("page", page),
		logx.Field("page_size", pageSize),
	).Info("获取逻辑图配置列表成功")

	return &types.GetGraphConfigListResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
		PageParams: types.PageParams{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
		Data: struct {
			List []types.GraphConfigMetadata `json:"list"`
		}{
			List: list,
		},
	}, nil
}
