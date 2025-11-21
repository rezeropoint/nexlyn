package orgmapping

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteOrgMappingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除组织映射
func NewDeleteOrgMappingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteOrgMappingLogic {
	return &DeleteOrgMappingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteOrgMappingLogic) DeleteOrgMapping(req *types.DeleteOrgMappingRequest) (resp *types.DeleteOrgMappingResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "org_mapping"),
		logx.Field("operation", "delete_org_mapping"),
		logx.Field("status", "started"),
		logx.Field("mapping_id", req.Id),
	).Info("开始删除组织映射")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "delete_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.DeleteOrgMappingResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查组织映射删除权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceOrgMapping, Action: casbinxcore.ActionDelete},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "delete_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.DeleteOrgMappingResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "delete_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.DeleteOrgMappingResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要组织映射删除权限"},
		}, nil
	}

	// 先获取映射，验证租户权限
	existingMapping, err := l.svcCtx.SkylarkEngine.GetOrgMapping(l.ctx, req.Id)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "delete_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("mapping_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("获取组织映射失败")

		return &types.DeleteOrgMappingResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 租户隔离验证：防止跨租户删除
	if existingMapping.TenantID != jwtUser.TenantId {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "delete_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("mapping_id", req.Id),
			logx.Field("user_tenant_id", jwtUser.TenantId),
			logx.Field("mapping_tenant_id", existingMapping.TenantID),
		).Error("租户隔离验证失败")

		return &types.DeleteOrgMappingResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：无权删除该组织映射"},
		}, nil
	}

	// 调用Skylark引擎删除组织映射
	err = l.svcCtx.SkylarkEngine.DeleteOrgMapping(l.ctx, req.Id)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "delete_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("mapping_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("Skylark引擎删除组织映射失败")

		return &types.DeleteOrgMappingResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "org_mapping"),
		logx.Field("operation", "delete_org_mapping"),
		logx.Field("status", "success"),
		logx.Field("mapping_id", req.Id),
		logx.Field("tenant_id", jwtUser.TenantId),
	).Info("删除组织映射成功")

	return &types.DeleteOrgMappingResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
	}, nil
}
