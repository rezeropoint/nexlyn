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

type DeleteHttpReceiveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除 HTTP 接收配置
func NewDeleteHttpReceiveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteHttpReceiveLogic {
	return &DeleteHttpReceiveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteHttpReceiveLogic) DeleteHttpReceive(req *types.DeleteHttpReceiveRequest) (resp *types.DeleteHttpReceiveResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_http_receive"),
		logx.Field("operation", "delete"),
		logx.Field("status", "started"),
		logx.Field("config_id", req.Id),
	).Info("开始删除HTTP接收配置")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("module", "iot_http_receive"),
			logx.Field("operation", "delete"),
			logx.Field("config_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.DeleteHttpReceiveResponse{
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
		casbinxcore.Permission{Resource: auth.ResourceIoTHttpReceive, Action: casbinxcore.ActionDelete},
	)
	if err != nil {
		return &types.DeleteHttpReceiveResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		return &types.DeleteHttpReceiveResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足"},
		}, nil
	}

	// 调用IoT引擎删除配置
	err = l.svcCtx.IoTEngine.DeleteHttpReceiveConfig(l.ctx, req.Id, jwtUser.TenantId)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("module", "iot_http_receive"),
			logx.Field("operation", "delete"),
			logx.Field("config_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("删除HTTP接收配置失败")

		return &types.DeleteHttpReceiveResponse{
			BaseResponse: types.BaseResponse{Code: code, Msg: msg},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_http_receive"),
		logx.Field("operation", "delete"),
		logx.Field("status", "success"),
		logx.Field("config_id", req.Id),
	).Info("删除HTTP接收配置成功")

	return &types.DeleteHttpReceiveResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
	}, nil
}
