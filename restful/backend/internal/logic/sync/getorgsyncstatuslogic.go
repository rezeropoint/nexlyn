// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package sync

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrgSyncStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询组织同步状态
func NewGetOrgSyncStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrgSyncStatusLogic {
	return &GetOrgSyncStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrgSyncStatusLogic) GetOrgSyncStatus(req *types.GetOrgSyncStatusRequest) (resp *types.GetOrgSyncStatusResponse, err error) {
	// 检查是否启用Skylark同步
	if !l.svcCtx.Config.SkylarkSyncEnabled || l.svcCtx.AdminSyncClient == nil {
		return &types.GetOrgSyncStatusResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "Skylark同步功能未启用",
			},
		}, nil
	}

	// 从 JWT 获取当前用户信息（租户ID）
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		return &types.GetOrgSyncStatusResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "获取当前用户信息失败",
			},
		}, nil
	}

	// 参数验证
	if len(req.OrgIds) == 0 {
		return &types.GetOrgSyncStatusResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "组织ID列表不能为空",
			},
		}, nil
	}

	// 调用 eventsync gRPC 批量查询组织同步状态
	statusResp, err := l.svcCtx.AdminSyncClient.GetOrgSyncStatus(l.ctx, &pb.GetOrgSyncStatusReq{
		TenantId: currentUser.TenantId,
		OrgIds:   req.OrgIds,
	})

	if err != nil {
		logx.WithContext(l.ctx).Error("调用eventsync查询组织同步状态失败: ", err)
		return &types.GetOrgSyncStatusResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "查询失败",
			},
		}, nil
	}

	if !statusResp.Success {
		return &types.GetOrgSyncStatusResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  statusResp.Message,
			},
		}, nil
	}

	return &types.GetOrgSyncStatusResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "查询成功",
		},
		Data: struct {
			SyncStatus map[string]bool `json:"syncStatus"`
		}{
			SyncStatus: statusResp.SyncStatus,
		},
	}, nil
}
