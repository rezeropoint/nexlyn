package device

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeviceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取设备详情
func NewGetDeviceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeviceLogic {
	return &GetDeviceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDeviceLogic) GetDevice(req *types.GetDeviceRequest) (resp *types.GetDeviceResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_device"),
		logx.Field("operation", "get_device"),
		logx.Field("status", "started"),
		logx.Field("device_id", req.DeviceId),
	).Info("开始获取设备详情")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device"),
			logx.Field("operation", "get_device"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetDeviceResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查设备查看权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceIoTDevice, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device"),
			logx.Field("operation", "get_device"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetDeviceResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device"),
			logx.Field("operation", "get_device"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetDeviceResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要设备查看权限"},
		}, nil
	}

	// 查询用户组织权限范围（包括所有子组织）
	orgIds, err := auth.GetAllChildOrgIds(l.ctx, l.svcCtx.DBConn, req.OrgId, jwtUser.TenantId)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device"),
			logx.Field("operation", "get_device"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询组织权限范围失败")

		return &types.GetDeviceResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "查询组织失败: " + err.Error()},
		}, nil
	}

	// 调用IoT引擎获取设备详情
	device, err := l.svcCtx.IoTEngine.GetDevice(l.ctx, req.DeviceId, jwtUser.TenantId, orgIds)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device"),
			logx.Field("operation", "get_device"),
			logx.Field("status", "failed"),
			logx.Field("device_id", req.DeviceId),
			logx.Field("error", err.Error()),
		).Error("IoT引擎获取设备详情失败")

		return &types.GetDeviceResponse{
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
		logx.Field("operation", "get_device"),
		logx.Field("status", "success"),
		logx.Field("device_id", req.DeviceId),
		logx.Field("device_model", device.DeviceModel),
	).Info("获取设备详情成功")

	return &types.GetDeviceResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: svc.ConvertCoreDeviceBindingToTypes(device),
	}, nil
}
