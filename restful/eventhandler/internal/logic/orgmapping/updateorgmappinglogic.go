package orgmapping

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateOrgMappingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新组织映射
func NewUpdateOrgMappingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOrgMappingLogic {
	return &UpdateOrgMappingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateOrgMappingLogic) UpdateOrgMapping(req *types.UpdateOrgMappingRequest) (resp *types.UpdateOrgMappingResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "org_mapping"),
		logx.Field("operation", "update_org_mapping"),
		logx.Field("status", "started"),
		logx.Field("mapping_id", req.Id),
	).Info("开始更新组织映射")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "update_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.UpdateOrgMappingResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查组织映射写入权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceOrgMapping, Action: casbinxcore.ActionWrite},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "update_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.UpdateOrgMappingResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "update_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.UpdateOrgMappingResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要组织映射写入权限"},
		}, nil
	}

	// 先获取现有映射，验证租户权限
	existingMapping, err := l.svcCtx.SkylarkEngine.GetOrgMapping(l.ctx, req.Id)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "update_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("mapping_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("获取现有组织映射失败")

		return &types.UpdateOrgMappingResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 租户隔离验证：防止跨租户更新
	if existingMapping.TenantID != jwtUser.TenantId {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "update_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("mapping_id", req.Id),
			logx.Field("user_tenant_id", jwtUser.TenantId),
			logx.Field("mapping_tenant_id", existingMapping.TenantID),
		).Error("租户隔离验证失败")

		return &types.UpdateOrgMappingResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：无权更新该组织映射"},
		}, nil
	}

	// 转换更新请求为Core模型
	updatedMapping := svc.ConvertUpdateRequestToCoreOrgMapping(req, existingMapping)

	// 调用Skylark引擎更新组织映射
	err = l.svcCtx.SkylarkEngine.UpdateOrgMapping(l.ctx, updatedMapping)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "update_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("mapping_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("Skylark引擎更新组织映射失败")

		return &types.UpdateOrgMappingResponse{
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
		logx.Field("operation", "update_org_mapping"),
		logx.Field("status", "success"),
		logx.Field("mapping_id", req.Id),
		logx.Field("tenant_id", jwtUser.TenantId),
	).Info("更新组织映射成功")

	return &types.UpdateOrgMappingResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
	}, nil
}
