// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package sync

import (
	"context"

	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncHealthCheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// eventsync服务健康检查
func NewSyncHealthCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncHealthCheckLogic {
	return &SyncHealthCheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncHealthCheckLogic) SyncHealthCheck() (resp *types.SyncHealthCheckResponse, err error) {
	// 检查是否启用Skylark同步
	if !l.svcCtx.Config.SkylarkSyncEnabled || l.svcCtx.AdminSyncClient == nil {
		return &types.SyncHealthCheckResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "Skylark同步功能未启用",
			},
			Data: struct {
				Healthy bool `json:"healthy"`
			}{
				Healthy: false,
			},
		}, nil
	}

	// 调用 eventsync gRPC 健康检查
	healthResp, err := l.svcCtx.AdminSyncClient.HealthCheck(l.ctx, &pb.HealthCheckReq{})
	if err != nil {
		logx.WithContext(l.ctx).Error("调用eventsync健康检查失败: ", err)
		return &types.SyncHealthCheckResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "健康检查失败",
			},
			Data: struct {
				Healthy bool `json:"healthy"`
			}{
				Healthy: false,
			},
		}, nil
	}

	return &types.SyncHealthCheckResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "健康检查成功",
		},
		Data: struct {
			Healthy bool `json:"healthy"`
		}{
			Healthy: healthResp.Healthy,
		},
	}, nil
}
