package template

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSensorTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除设备模板
func NewDeleteSensorTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSensorTemplateLogic {
	return &DeleteSensorTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteSensorTemplateLogic) DeleteSensorTemplate(req *types.DeleteSensorTemplateRequest) (resp *types.DeleteSensorTemplateResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_template"),
		logx.Field("operation", "delete_template"),
		logx.Field("status", "started"),
		logx.Field("template_id", req.Id),
	).Info("开始删除设备模板")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_template"),
			logx.Field("operation", "delete_template"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.DeleteSensorTemplateResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查模板删除权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceIoTTemplate, Action: casbinxcore.ActionWrite},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_template"),
			logx.Field("operation", "delete_template"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.DeleteSensorTemplateResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_template"),
			logx.Field("operation", "delete_template"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.DeleteSensorTemplateResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要模板管理权限"},
		}, nil
	}

	// 调用IoT引擎删除模板
	err = l.svcCtx.IoTEngine.DeleteTemplate(l.ctx, req.Id)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_template"),
			logx.Field("operation", "delete_template"),
			logx.Field("status", "failed"),
			logx.Field("template_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("IoT引擎删除模板失败")

		return &types.DeleteSensorTemplateResponse{
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
		logx.Field("module", "iot_template"),
		logx.Field("operation", "delete_template"),
		logx.Field("status", "success"),
		logx.Field("template_id", req.Id),
	).Info("删除设备模板成功")

	return &types.DeleteSensorTemplateResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
	}, nil
}
