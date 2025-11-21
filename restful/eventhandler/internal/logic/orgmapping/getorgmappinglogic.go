package orgmapping

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrgMappingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取单个组织映射
func NewGetOrgMappingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrgMappingLogic {
	return &GetOrgMappingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrgMappingLogic) GetOrgMapping(req *types.GetOrgMappingRequest) (resp *types.GetOrgMappingResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "org_mapping"),
		logx.Field("operation", "get_org_mapping"),
		logx.Field("status", "started"),
		logx.Field("mapping_id", req.Id),
	).Info("开始获取组织映射详情")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "get_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetOrgMappingResponse{
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
			logx.Field("operation", "get_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetOrgMappingResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "get_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetOrgMappingResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要组织映射查看权限"},
		}, nil
	}

	// 调用Skylark引擎获取组织映射
	mapping, err := l.svcCtx.SkylarkEngine.GetOrgMapping(l.ctx, req.Id)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "get_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("mapping_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("Skylark引擎获取组织映射失败")

		return &types.GetOrgMappingResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 租户隔离验证：防止跨租户访问
	if mapping.TenantID != jwtUser.TenantId {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "get_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("mapping_id", req.Id),
			logx.Field("user_tenant_id", jwtUser.TenantId),
			logx.Field("mapping_tenant_id", mapping.TenantID),
		).Error("租户隔离验证失败")

		return &types.GetOrgMappingResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：无权访问该组织映射"},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "org_mapping"),
		logx.Field("operation", "get_org_mapping"),
		logx.Field("status", "success"),
		logx.Field("mapping_id", req.Id),
		logx.Field("remote_org_value", mapping.RemoteOrgValue),
	).Info("获取组织映射详情成功")

	return &types.GetOrgMappingResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: svc.ConvertCoreOrgMappingToTypes(mapping),
	}, nil
}
