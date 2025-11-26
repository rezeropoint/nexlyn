// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package httpreceive

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateHttpReceiveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新 HTTP 接收配置
func NewUpdateHttpReceiveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateHttpReceiveLogic {
	return &UpdateHttpReceiveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateHttpReceiveLogic) UpdateHttpReceive(req *types.UpdateHttpReceiveRequest) (resp *types.UpdateHttpReceiveResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_http_receive"),
		logx.Field("operation", "update"),
		logx.Field("status", "started"),
		logx.Field("config_id", req.Id),
	).Info("开始更新HTTP接收配置")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("module", "iot_http_receive"),
			logx.Field("operation", "update"),
			logx.Field("config_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.UpdateHttpReceiveResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceIoTHttpReceive, Action: casbinxcore.ActionWrite},
	)
	if err != nil {
		return &types.UpdateHttpReceiveResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		return &types.UpdateHttpReceiveResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足"},
		}, nil
	}

	// 转换为Core类型
	metadata, config := svc.ConvertHttpReceiveUpdateRequestToCore(req, jwtUser.TenantId)

	// 调用IoT引擎更新配置
	err = l.svcCtx.IoTEngine.UpdateHttpReceiveConfig(l.ctx, req.Id, jwtUser.TenantId, metadata, config)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("module", "iot_http_receive"),
			logx.Field("operation", "update"),
			logx.Field("config_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("更新HTTP接收配置失败")

		return &types.UpdateHttpReceiveResponse{
			BaseResponse: types.BaseResponse{Code: code, Msg: msg},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_http_receive"),
		logx.Field("operation", "update"),
		logx.Field("status", "success"),
		logx.Field("config_id", req.Id),
	).Info("更新HTTP接收配置成功")

	return &types.UpdateHttpReceiveResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
	}, nil
}
