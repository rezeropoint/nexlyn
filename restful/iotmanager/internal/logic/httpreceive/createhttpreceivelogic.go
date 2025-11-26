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

type CreateHttpReceiveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建 HTTP 接收配置
func NewCreateHttpReceiveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateHttpReceiveLogic {
	return &CreateHttpReceiveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateHttpReceiveLogic) CreateHttpReceive(req *types.CreateHttpReceiveRequest) (resp *types.CreateHttpReceiveResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_http_receive"),
		logx.Field("operation", "create"),
		logx.Field("status", "started"),
		logx.Field("name", req.Name),
	).Info("开始创建HTTP接收配置")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_http_receive"),
			logx.Field("operation", "create"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.CreateHttpReceiveResponse{
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
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_http_receive"),
			logx.Field("operation", "create"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.CreateHttpReceiveResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_http_receive"),
			logx.Field("operation", "create"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.CreateHttpReceiveResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要HTTP接收配置管理权限"},
		}, nil
	}

	// 转换API请求为Core类型
	metadata, config := svc.ConvertHttpReceiveRequestToCore(req, jwtUser.TenantId)
	metadata.CreatedBy = jwtUser.UserId

	// 调用IoT引擎创建配置
	configID, err := l.svcCtx.IoTEngine.CreateHttpReceiveConfig(l.ctx, metadata, config)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_http_receive"),
			logx.Field("operation", "create"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("IoT引擎创建HTTP接收配置失败")

		return &types.CreateHttpReceiveResponse{
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
		logx.Field("module", "iot_http_receive"),
		logx.Field("operation", "create"),
		logx.Field("status", "success"),
		logx.Field("config_id", configID),
		logx.Field("name", req.Name),
	).Info("创建HTTP接收配置成功")

	return &types.CreateHttpReceiveResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		}{
			Id:   configID,
			Name: req.Name,
		},
	}, nil
}
