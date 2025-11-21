package logic

import (
	"context"

	"github.com/rezeropoint/nexlyn/service/eventsync/internal/svc"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserSyncStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserSyncStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSyncStatusLogic {
	return &GetUserSyncStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询用户同步状态
func (l *GetUserSyncStatusLogic) GetUserSyncStatus(in *pb.GetUserSyncStatusReq) (*pb.GetUserSyncStatusResp, error) {
	// 1. 参数验证
	if in.TenantId == "" {
		l.Errorf("参数验证失败: TenantId为空")
		return &pb.GetUserSyncStatusResp{
			Success: false,
			Message: "租户ID不能为空",
		}, nil
	}

	if len(in.UserIds) == 0 {
		l.Errorf("参数验证失败: 用户ID列表为空")
		return &pb.GetUserSyncStatusResp{
			Success: false,
			Message: "用户ID列表不能为空",
		}, nil
	}

	// 2. 循环调用 AdminEngine 查询每个用户的同步状态
	syncStatus := make(map[string]bool, len(in.UserIds))
	for _, userID := range in.UserIds {
		isSynced, err := l.svcCtx.AdminEngine.GetUserSyncStatus(l.ctx, in.TenantId, userID)
		if err != nil {
			// 数据库错误，记录日志但继续处理其他用户
			l.Errorf("查询用户同步状态失败: userId=%s, error=%v", userID, err)
			syncStatus[userID] = false
		} else {
			syncStatus[userID] = isSynced
		}
	}

	l.Infof("查询用户同步状态完成: UserCount=%d", len(in.UserIds))

	return &pb.GetUserSyncStatusResp{
		Success:    true,
		Message:    "查询成功",
		SyncStatus: syncStatus,
	}, nil
}
