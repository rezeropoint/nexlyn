package logic

import (
	"context"

	"github.com/rezeropoint/nexlyn/service/eventsync/internal/svc"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteOrganizationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteOrganizationLogic {
	return &DeleteOrganizationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除组织同步
func (l *DeleteOrganizationLogic) DeleteOrganization(in *pb.DeleteOrganizationReq) (*pb.DeleteOrganizationResp, error) {
	// 1. 参数验证
	if in.TenantId == "" || in.LocalOrgId == "" {
		l.Errorf("参数验证失败: TenantId=%s, LocalOrgId=%s", in.TenantId, in.LocalOrgId)
		return &pb.DeleteOrganizationResp{
			Success:   false,
			Message:   "租户ID和组织ID不能为空",
			ErrorCode: "INVALID_PARAMS",
		}, nil
	}

	// 2. 调用 AdminEngine 删除组织
	err := l.svcCtx.AdminEngine.DeleteOrganization(
		l.ctx,
		in.TenantId,
		in.LocalOrgId,
	)
	if err != nil {
		l.Errorf("删除Skylark组织失败: %v (TenantId=%s, LocalOrgId=%s)", err, in.TenantId, in.LocalOrgId)
		return &pb.DeleteOrganizationResp{
			Success:   false,
			Message:   err.Error(),
			ErrorCode: "DELETE_FAILED",
		}, nil
	}

	l.Infof("成功删除Skylark组织: TenantId=%s, LocalOrgId=%s", in.TenantId, in.LocalOrgId)

	return &pb.DeleteOrganizationResp{
		Success: true,
		Message: "组织删除成功",
	}, nil
}
