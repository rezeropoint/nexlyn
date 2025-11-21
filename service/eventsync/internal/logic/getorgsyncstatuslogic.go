package logic

import (
	"context"

	"github.com/rezeropoint/nexlyn/service/eventsync/internal/svc"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrgSyncStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrgSyncStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrgSyncStatusLogic {
	return &GetOrgSyncStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询组织同步状态
func (l *GetOrgSyncStatusLogic) GetOrgSyncStatus(in *pb.GetOrgSyncStatusReq) (*pb.GetOrgSyncStatusResp, error) {
	// 1. 参数验证
	if in.TenantId == "" {
		l.Errorf("参数验证失败: TenantId为空")
		return &pb.GetOrgSyncStatusResp{
			Success: false,
			Message: "租户ID不能为空",
		}, nil
	}

	if len(in.OrgIds) == 0 {
		l.Errorf("参数验证失败: 组织ID列表为空")
		return &pb.GetOrgSyncStatusResp{
			Success: false,
			Message: "组织ID列表不能为空",
		}, nil
	}

	// 2. 循环调用 AdminEngine 查询每个组织的同步状态
	syncStatus := make(map[string]bool, len(in.OrgIds))
	for _, orgID := range in.OrgIds {
		isSynced, err := l.svcCtx.AdminEngine.GetOrgSyncStatus(l.ctx, in.TenantId, orgID)
		if err != nil {
			// 数据库错误，记录日志但继续处理其他组织
			l.Errorf("查询组织同步状态失败: orgId=%s, error=%v", orgID, err)
			syncStatus[orgID] = false
		} else {
			syncStatus[orgID] = isSynced
		}
	}

	l.Infof("查询组织同步状态完成: OrgCount=%d", len(in.OrgIds))

	return &pb.GetOrgSyncStatusResp{
		Success:    true,
		Message:    "查询成功",
		SyncStatus: syncStatus,
	}, nil
}
