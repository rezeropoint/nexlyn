package device

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListUnboundDevicesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取未绑定设备列表
func NewListUnboundDevicesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUnboundDevicesLogic {
	return &ListUnboundDevicesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUnboundDevicesLogic) ListUnboundDevices(req *types.ListUnboundDevicesRequest) (resp *types.ListUnboundDevicesResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_device"),
		logx.Field("operation", "list_unbound_devices"),
		logx.Field("status", "started"),
	).Info("开始获取未绑定设备列表")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device"),
			logx.Field("operation", "list_unbound_devices"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.ListUnboundDevicesResponse{
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
			logx.Field("operation", "list_unbound_devices"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.ListUnboundDevicesResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device"),
			logx.Field("operation", "list_unbound_devices"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.ListUnboundDevicesResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要设备查看权限"},
		}, nil
	}

	// 调用IoT引擎获取未绑定设备列表（租户级查询，不限制组织）
	devices, err := l.svcCtx.IoTEngine.ListUnboundDevices(l.ctx, jwtUser.TenantId)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device"),
			logx.Field("operation", "list_unbound_devices"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("IoT引擎获取未绑定设备列表失败")

		return &types.ListUnboundDevicesResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 转换为API类型
	list := svc.ConvertCoreUnboundDevicesToTypes(devices)

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_device"),
		logx.Field("operation", "list_unbound_devices"),
		logx.Field("status", "success"),
		logx.Field("count", len(list)),
	).Info("获取未绑定设备列表成功")

	return &types.ListUnboundDevicesResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: struct {
			List []types.UnboundDevice `json:"list"`
		}{
			List: list,
		},
	}, nil
}
