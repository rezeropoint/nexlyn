package platform

import (
	"context"

	"github.com/rezeropoint/go-skylark/v2/core"
	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type SavePlatformConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建/更新Skylark平台配置
func NewSavePlatformConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SavePlatformConfigLogic {
	return &SavePlatformConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SavePlatformConfigLogic) SavePlatformConfig(req *types.SavePlatformConfigRequest) (resp *types.SavePlatformConfigResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "skylark_platform"),
		logx.Field("operation", "save_platform_config"),
		logx.Field("status", "started"),
	).Info("开始保存Skylark平台配置")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_platform"),
			logx.Field("operation", "save_platform_config"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.SavePlatformConfigResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查平台配置写入权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceSkylarkPlatform, Action: casbinxcore.ActionWrite},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_platform"),
			logx.Field("operation", "save_platform_config"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.SavePlatformConfigResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_platform"),
			logx.Field("operation", "save_platform_config"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.SavePlatformConfigResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要Skylark平台配置写入权限"},
		}, nil
	}

	// 转换为Core平台配置
	platformCfg := svc.ConvertAPIRequestToCorePlatformConfig(req, jwtUser.TenantId, jwtUser.UserId)

	// 检查平台配置是否已存在
	existingCfg, err := l.svcCtx.SkylarkEngine.GetPlatformConfig(l.ctx, jwtUser.TenantId)

	var platformID string
	isUpdate := false

	if err == nil && existingCfg != nil {
		// 平台配置已存在，执行更新
		isUpdate = true
		platformCfg.ID = existingCfg.ID
		platformID = existingCfg.ID

		err = l.svcCtx.SkylarkEngine.UpdatePlatformConfig(l.ctx, platformCfg)
		if err != nil {
			code, msg := svc.HandleSkylarkError(err)
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "skylark_platform"),
				logx.Field("operation", "save_platform_config"),
				logx.Field("status", "failed"),
				logx.Field("tenant_id", jwtUser.TenantId),
				logx.Field("is_update", true),
				logx.Field("error", err.Error()),
			).Error("Skylark引擎更新平台配置失败")

			return &types.SavePlatformConfigResponse{
				BaseResponse: types.BaseResponse{
					Code: code,
					Msg:  msg,
				},
			}, nil
		}
	} else if err == core.ErrPlatformConfigNotFound {
		// 平台配置不存在，执行创建
		isUpdate = false
		_, err = l.svcCtx.SkylarkEngine.CreatePlatformConfig(l.ctx, platformCfg)
		if err != nil {
			code, msg := svc.HandleSkylarkError(err)
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "skylark_platform"),
				logx.Field("operation", "save_platform_config"),
				logx.Field("status", "failed"),
				logx.Field("tenant_id", jwtUser.TenantId),
				logx.Field("is_update", false),
				logx.Field("error", err.Error()),
			).Error("Skylark引擎创建平台配置失败")

			return &types.SavePlatformConfigResponse{
				BaseResponse: types.BaseResponse{
					Code: code,
					Msg:  msg,
				},
			}, nil
		}

		// 重新获取平台配置以获取ID
		createdCfg, err := l.svcCtx.SkylarkEngine.GetPlatformConfig(l.ctx, jwtUser.TenantId)
		if err == nil && createdCfg != nil {
			platformID = createdCfg.ID
		}
	} else {
		// 其他错误
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_platform"),
			logx.Field("operation", "save_platform_config"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("error", err.Error()),
		).Error("检查平台配置是否存在时失败")

		return &types.SavePlatformConfigResponse{
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
		logx.Field("module", "skylark_platform"),
		logx.Field("operation", "save_platform_config"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", jwtUser.TenantId),
		logx.Field("is_update", isUpdate),
		logx.Field("platform_id", platformID),
	).Info("保存Skylark平台配置成功")

	return &types.SavePlatformConfigResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: struct {
			Id string `json:"id"`
		}{
			Id: platformID,
		},
	}, nil
}
