package logic

import (
	"context"

	"github.com/rezeropoint/nexlyn/service/eventsync/internal/svc"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveMemberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveMemberLogic {
	return &RemoveMemberLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 从组织移除成员
func (l *RemoveMemberLogic) RemoveMember(in *pb.RemoveMemberReq) (*pb.RemoveMemberResp, error) {
	// 1. 参数验证
	if in.TenantId == "" || in.LocalOrgId == "" || in.LocalMemberId == "" {
		l.Errorf("参数验证失败: TenantId=%s, LocalOrgId=%s, LocalMemberId=%s",
			in.TenantId, in.LocalOrgId, in.LocalMemberId)
		return &pb.RemoveMemberResp{
			Success:   false,
			Message:   "租户ID、组织ID和成员ID不能为空",
			ErrorCode: "INVALID_PARAMS",
		}, nil
	}

	// 2. 调用AdminEngine移除成员
	l.Infof("从Skylark组织移除成员: LocalOrgId=%s, LocalMemberId=%s", in.LocalOrgId, in.LocalMemberId)
	err := l.svcCtx.AdminEngine.RemoveMember(l.ctx, in.TenantId, in.LocalOrgId, in.LocalMemberId)
	if err != nil {
		l.Errorf("从Skylark组织移除成员失败: %v (LocalOrgId=%s, LocalMemberId=%s)",
			err, in.LocalOrgId, in.LocalMemberId)
		return &pb.RemoveMemberResp{
			Success:   false,
			Message:   err.Error(),
			ErrorCode: "REMOVE_MEMBER_FAILED",
		}, nil
	}

	l.Infof("成功从Skylark组织移除成员: LocalOrgId=%s, LocalMemberId=%s", in.LocalOrgId, in.LocalMemberId)

	return &pb.RemoveMemberResp{
		Success: true,
		Message: "成员移除成功",
	}, nil
}
