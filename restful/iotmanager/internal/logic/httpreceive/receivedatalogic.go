// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package httpreceive

import (
	"context"

	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReceiveDataLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 接收 HTTP 数据
func NewReceiveDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReceiveDataLogic {
	return &ReceiveDataLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ReceiveData 处理接收到的HTTP数据
// 注意：此方法不需要JWT认证，由外部设备直接调用
// configId: 配置ID（从URL路径参数获取）
// data: 请求体解析后的JSON数据（由Handler层解析并传入）
func (l *ReceiveDataLogic) ReceiveData(configId string, data map[string]any) (resp *types.ReceiveDataResponse, err error) {
	// 记录数据接收
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "http_receive"),
		logx.Field("operation", "receive_data"),
		logx.Field("config_id", configId),
	).Info("接收到HTTP数据")

	// 调用IoT引擎处理数据
	err = l.svcCtx.IoTEngine.ProcessHttpReceiveData(l.ctx, configId, data)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "http_receive"),
			logx.Field("operation", "receive_data"),
			logx.Field("config_id", configId),
			logx.Field("error", err.Error()),
		).Error("处理HTTP数据失败")

		return &types.ReceiveDataResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	return &types.ReceiveDataResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
	}, nil
}
