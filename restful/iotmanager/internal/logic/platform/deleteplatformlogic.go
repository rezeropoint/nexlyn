package platform

import (
	"context"
	"database/sql"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeletePlatformLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除平台配置
func NewDeletePlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePlatformLogic {
	return &DeletePlatformLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeletePlatformLogic) DeletePlatform(req *types.DeletePlatformRequest) (resp *types.DeletePlatformResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_platform"),
		logx.Field("operation", "delete_platform"),
		logx.Field("status", "started"),
		logx.Field("platform_id", req.Id),
	).Info("开始删除平台配置")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_platform"),
			logx.Field("operation", "delete_platform"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.DeletePlatformResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查平台配置删除权限
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
			logx.Field("operation", "delete_platform"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.DeletePlatformResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_platform"),
			logx.Field("operation", "delete_platform"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.DeletePlatformResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要平台配置管理权限"},
		}, nil
	}

	// 首先从数据库查询平台类型（用于构建Etcd key）
	var platformTypeStr string
	query := `SELECT type FROM iot_platform_configs WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`
	err = l.svcCtx.DBConn.QueryRowCtx(l.ctx, &platformTypeStr, query, req.Id, jwtUser.TenantId)
	if err != nil {
		if err == sql.ErrNoRows {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "iot_platform"),
				logx.Field("operation", "delete_platform"),
				logx.Field("status", "failed"),
				logx.Field("platform_id", req.Id),
			).Error("平台配置不存在")

			return &types.DeletePlatformResponse{
				BaseResponse: types.BaseResponse{Code: 404, Msg: "平台配置不存在"},
			}, nil
		}

		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_platform"),
			logx.Field("operation", "delete_platform"),
			logx.Field("status", "failed"),
			logx.Field("platform_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("数据库查询失败")

		return &types.DeletePlatformResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "数据库查询失败"},
		}, nil
	}

	platformType := core.PlatformType(platformTypeStr)

	// 调用IoT引擎删除平台配置
	err = l.svcCtx.IoTEngine.DeletePlatform(l.ctx, platformType, req.Id, jwtUser.TenantId)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_platform"),
			logx.Field("operation", "delete_platform"),
			logx.Field("status", "failed"),
			logx.Field("platform_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("IoT引擎删除平台配置失败")

		return &types.DeletePlatformResponse{
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
		logx.Field("operation", "delete_platform"),
		logx.Field("status", "success"),
		logx.Field("platform_id", req.Id),
		logx.Field("platform_type", platformType),
	).Info("删除平台配置成功")

	return &types.DeletePlatformResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
	}, nil
}
