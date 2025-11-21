package platform

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePlatformLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新平台配置
func NewUpdatePlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePlatformLogic {
	return &UpdatePlatformLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePlatformLogic) UpdatePlatform(req *types.UpdatePlatformRequest) (resp *types.UpdatePlatformResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_platform"),
		logx.Field("operation", "update_platform"),
		logx.Field("status", "started"),
		logx.Field("platform_id", req.Id),
		logx.Field("platform_type", req.Type),
	).Info("开始更新平台配置")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_platform"),
			logx.Field("operation", "update_platform"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.UpdatePlatformResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查平台配置更新权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceIoTPlatform, Action: casbinxcore.ActionWrite},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_platform"),
			logx.Field("operation", "update_platform"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.UpdatePlatformResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_platform"),
			logx.Field("operation", "update_platform"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.UpdatePlatformResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要平台配置管理权限"},
		}, nil
	}

	// 转换API请求为Core类型
	metadata, config := svc.ConvertPlatformUpdateRequestToCore(req, jwtUser.TenantId)
	metadata.UpdatedBy = jwtUser.UserId

	platformType := core.PlatformType(req.Type)

	// 调用IoT引擎更新平台配置
	err = l.svcCtx.IoTEngine.UpdatePlatform(l.ctx, platformType, req.Id, jwtUser.TenantId, *metadata, config)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_platform"),
			logx.Field("operation", "update_platform"),
			logx.Field("status", "failed"),
			logx.Field("platform_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("IoT引擎更新平台配置失败")

		return &types.UpdatePlatformResponse{
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
		logx.Field("module", "iot_platform"),
		logx.Field("operation", "update_platform"),
		logx.Field("status", "success"),
		logx.Field("platform_id", req.Id),
		logx.Field("platform_type", req.Type),
	).Info("更新平台配置成功")

	return &types.UpdatePlatformResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
	}, nil
}
