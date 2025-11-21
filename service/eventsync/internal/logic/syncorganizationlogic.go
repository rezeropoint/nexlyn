package logic

import (
	"context"

	"github.com/rezeropoint/nexlyn/service/eventsync/internal/svc"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncOrganizationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncOrganizationLogic {
	return &SyncOrganizationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 同步单个组织到Skylark（自动判断根组织或子组织）
func (l *SyncOrganizationLogic) SyncOrganization(in *pb.SyncOrganizationReq) (*pb.SyncOrganizationResp, error) {
	// 1. 参数验证
	if in.TenantId == "" || in.LocalOrgId == "" || in.Name == "" || in.FounderId == "" {
		l.Errorf("参数验证失败: TenantId=%s, LocalOrgId=%s, Name=%s, FounderId=%s",
			in.TenantId, in.LocalOrgId, in.Name, in.FounderId)
		return &pb.SyncOrganizationResp{
			Success:   false,
			Message:   "租户ID、组织ID、组织名称和创建人ID不能为空",
			ErrorCode: "INVALID_PARAMS",
		}, nil
	}

	var err error

	// 2. 判断是根组织还是子组织
	if in.ParentLocalOrgId == "" {
		// 创建根组织
		l.Infof("创建根组织: TenantId=%s, LocalOrgId=%s, Name=%s", in.TenantId, in.LocalOrgId, in.Name)
		err = l.svcCtx.AdminEngine.CreateOrganization(
			l.ctx,
			in.TenantId,
			in.LocalOrgId,
			in.Name,
			in.Description,
			in.FounderId,
		)
	} else {
		// 创建子组织
		l.Infof("创建子组织: TenantId=%s, LocalOrgId=%s, ParentLocalOrgId=%s, Name=%s",
			in.TenantId, in.LocalOrgId, in.ParentLocalOrgId, in.Name)
		err = l.svcCtx.AdminEngine.CreateSubOrganization(
			l.ctx,
			in.TenantId,
			in.LocalOrgId,
			in.ParentLocalOrgId,
			in.Name,
			in.Description,
			in.FounderId,
		)
	}

	if err != nil {
		l.Errorf("创建Skylark组织失败: %v (TenantId=%s, LocalOrgId=%s)", err, in.TenantId, in.LocalOrgId)
		return &pb.SyncOrganizationResp{
			Success:   false,
			Message:   err.Error(),
			ErrorCode: "CREATE_FAILED",
		}, nil
	}

	l.Infof("成功同步组织到Skylark: LocalOrgId=%s", in.LocalOrgId)

	return &pb.SyncOrganizationResp{
		Success:     true,
		Message:     "组织同步成功",
		RemoteOrgId: in.LocalOrgId, // v2.1.2+使用LocalOrgId作为唯一标识
	}, nil
}
