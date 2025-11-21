package orgmapping

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrgMappingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建组织映射（租户级别）
func NewCreateOrgMappingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrgMappingLogic {
	return &CreateOrgMappingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrgMappingLogic) CreateOrgMapping(req *types.CreateOrgMappingRequest) (resp *types.CreateOrgMappingResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "org_mapping"),
		logx.Field("operation", "create_org_mapping"),
		logx.Field("status", "started"),
	).Info("开始创建组织映射")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "create_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.CreateOrgMappingResponse{
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
			logx.Field("operation", "create_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.CreateOrgMappingResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "create_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.CreateOrgMappingResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要组织映射写入权限"},
		}, nil
	}

	// 转换请求为Core模型
	coreMapping := svc.ConvertCreateRequestToCoreOrgMapping(req, jwtUser.TenantId)

	// 调用Skylark引擎创建组织映射
	mappingID, err := l.svcCtx.SkylarkEngine.CreateOrgMapping(l.ctx, coreMapping)
	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "org_mapping"),
			logx.Field("operation", "create_org_mapping"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("remote_org_value", req.RemoteOrgValue),
			logx.Field("error", err.Error()),
		).Error("Skylark引擎创建组织映射失败")

		return &types.CreateOrgMappingResponse{
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
		logx.Field("operation", "create_org_mapping"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", jwtUser.TenantId),
		logx.Field("mapping_id", mappingID),
		logx.Field("remote_org_value", req.RemoteOrgValue),
	).Info("创建组织映射成功")

	return &types.CreateOrgMappingResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: struct {
			Id string `json:"id"`
		}{
			Id: mappingID,
		},
	}, nil
}
