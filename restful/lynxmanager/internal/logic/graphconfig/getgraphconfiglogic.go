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

type GetGraphConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取逻辑图配置详情
func NewGetGraphConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGraphConfigLogic {
	return &GetGraphConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGraphConfigLogic) GetGraphConfig(req *types.GetGraphConfigRequest) (resp *types.GetGraphConfigResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_graphconfig"),
		logx.Field("operation", "get_graphconfig"),
		logx.Field("status", "started"),
		logx.Field("id", req.Id),
	).Info("开始获取逻辑图配置详情")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "get_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetGraphConfigResponse{
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
			logx.Field("operation", "get_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetGraphConfigResponse{
			BaseResponse: types.BaseResponse{Code: 500, Message: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "get_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetGraphConfigResponse{
			BaseResponse: types.BaseResponse{Code: 403, Message: "权限不足：需要逻辑图配置查看权限"},
		}, nil
	}

	// 调用Manager获取逻辑图配置
	graphConfig, err := l.svcCtx.LynxManager.GetGraphConfig(l.ctx, core.GraphKey{ID: req.Id})
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "get_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("Manager获取逻辑图配置失败")

		return &types.GetGraphConfigResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "服务器错误: " + err.Error(),
			},
		}, nil
	}

	// 验证组织权限：检查逻辑图所属组织是否在用户可见范围内
	orgIds, err := auth.GetAllChildOrgIds(l.ctx, l.svcCtx.DBConn, jwtUser.PrimaryOrgId, jwtUser.TenantId)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "get_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询用户可见组织列表失败")

		return &types.GetGraphConfigResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "服务器错误: 查询组织权限失败",
			},
		}, nil
	}

	// 检查逻辑图的org_id是否在用户的可见组织列表中
	hasOrgPermission := false
	for _, orgId := range orgIds {
		if orgId == graphConfig.OrgID {
			hasOrgPermission = true
			break
		}
	}

	if !hasOrgPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "get_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("graph_org_id", graphConfig.OrgID),
			logx.Field("user_org_id", jwtUser.PrimaryOrgId),
		).Error("用户无权访问该组织的逻辑图")

		return &types.GetGraphConfigResponse{
			BaseResponse: types.BaseResponse{
				Code:    403,
				Message: "权限不足：无法访问其他组织的逻辑图",
			},
		}, nil
	}

	// Phase 2 重构：转换为API类型（包含完整详情和标签名称）
	data := l.svcCtx.ConvertCoreGraphConfigToDetail(l.ctx, graphConfig)

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_graphconfig"),
		logx.Field("operation", "get_graphconfig"),
		logx.Field("status", "success"),
		logx.Field("id", req.Id),
	).Info("获取逻辑图配置详情成功")

	return &types.GetGraphConfigResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Data: data,
	}, nil
}
