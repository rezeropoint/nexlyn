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

type GetHttpReceiveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取 HTTP 接收配置详情
func NewGetHttpReceiveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHttpReceiveLogic {
	return &GetHttpReceiveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHttpReceiveLogic) GetHttpReceive(req *types.GetHttpReceiveRequest) (resp *types.GetHttpReceiveResponse, err error) {
	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("module", "iot_http_receive"),
			logx.Field("operation", "get"),
			logx.Field("config_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetHttpReceiveResponse{
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
		casbinxcore.Permission{Resource: auth.ResourceIoTHttpReceive, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		return &types.GetHttpReceiveResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		return &types.GetHttpReceiveResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足"},
		}, nil
	}

	// 调用IoT引擎获取配置详情
	httpReceive, err := l.svcCtx.IoTEngine.GetHttpReceiveConfig(l.ctx, req.Id, jwtUser.TenantId)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("module", "iot_http_receive"),
			logx.Field("operation", "get"),
			logx.Field("config_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("获取HTTP接收配置详情失败")

		return &types.GetHttpReceiveResponse{
			BaseResponse: types.BaseResponse{Code: code, Msg: msg},
		}, nil
	}

	// 转换为API类型
	detail := svc.ConvertCoreHttpReceiveToTypes(httpReceive)

	return &types.GetHttpReceiveResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: detail,
	}, nil
}
