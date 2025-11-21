package device

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type BindDeviceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 绑定设备
func NewBindDeviceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindDeviceLogic {
	return &BindDeviceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BindDeviceLogic) BindDevice(req *types.BindDeviceRequest) (resp *types.BindDeviceResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_device"),
		logx.Field("operation", "bind_device"),
		logx.Field("status", "started"),
		logx.Field("device_id", req.DeviceId),
		logx.Field("device_model", req.DeviceModel),
	).Info("开始绑定设备")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device"),
			logx.Field("operation", "bind_device"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.BindDeviceResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查设备绑定权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceIoTDevice, Action: casbinxcore.ActionWrite},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device"),
			logx.Field("operation", "bind_device"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.BindDeviceResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device"),
			logx.Field("operation", "bind_device"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.BindDeviceResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要设备管理权限"},
		}, nil
	}

	// 构建设备绑定元数据（使用用户的默认组织ID）
	metadata := core.DeviceBindingMetadata{
		DeviceID:         req.DeviceId,
		DeviceName:       req.DeviceName,
		DeviceAlias:      req.DeviceAlias,
		DeviceModel:      req.DeviceModel,
		Description:      req.Description,
		Location:         req.Location,
		InstallationDate: req.InstallationDate,
		Status:           req.Status,
		TagIDs:           req.TagIds,
		TenantID:         jwtUser.TenantId,
		OrgID:            jwtUser.PrimaryOrgId, // 使用用户默认组织
		CreatedBy:        jwtUser.UserId,
	}

	// 调用IoT引擎绑定设备
	deviceBindingID, err := l.svcCtx.IoTEngine.BindDevice(l.ctx, metadata)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device"),
			logx.Field("operation", "bind_device"),
			logx.Field("status", "failed"),
			logx.Field("device_id", req.DeviceId),
			logx.Field("error", err.Error()),
		).Error("IoT引擎绑定设备失败")

		return &types.BindDeviceResponse{
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
		logx.Field("module", "iot_device"),
		logx.Field("operation", "bind_device"),
		logx.Field("status", "success"),
		logx.Field("binding_id", deviceBindingID),
		logx.Field("device_id", req.DeviceId),
		logx.Field("device_model", req.DeviceModel),
	).Info("绑定设备成功")

	return &types.BindDeviceResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: struct {
			Id       string `json:"id"`
			DeviceId string `json:"deviceId"`
		}{
			Id:       deviceBindingID,
			DeviceId: req.DeviceId,
		},
	}, nil
}
