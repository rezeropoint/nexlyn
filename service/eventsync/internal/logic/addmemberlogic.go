package logic

import (
	"context"

	"github.com/rezeropoint/nexlyn/service/eventsync/internal/svc"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddMemberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddMemberLogic {
	return &AddMemberLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 添加成员到组织
func (l *AddMemberLogic) AddMember(in *pb.AddMemberReq) (*pb.AddMemberResp, error) {
	// 1. 参数验证
	if in.TenantId == "" || in.LocalOrgId == "" || in.LocalMemberId == "" {
		l.Errorf("参数验证失败: TenantId=%s, LocalOrgId=%s, LocalMemberId=%s",
			in.TenantId, in.LocalOrgId, in.LocalMemberId)
		return &pb.AddMemberResp{
			Success:   false,
			Message:   "租户ID、组织ID和成员ID不能为空",
			ErrorCode: "INVALID_PARAMS",
		}, nil
	}

	// 2. 调用AdminEngine添加成员
	l.Infof("添加成员到Skylark组织: LocalOrgId=%s, LocalMemberId=%s", in.LocalOrgId, in.LocalMemberId)
	err := l.svcCtx.AdminEngine.AddMember(l.ctx, in.TenantId, in.LocalOrgId, in.LocalMemberId)
	if err != nil {
		l.Errorf("添加成员到Skylark组织失败: %v (LocalOrgId=%s, LocalMemberId=%s)",
			err, in.LocalOrgId, in.LocalMemberId)
		return &pb.AddMemberResp{
			Success:   false,
			Message:   err.Error(),
			ErrorCode: "ADD_MEMBER_FAILED",
		}, nil
	}

	l.Infof("成功添加成员到Skylark组织: LocalOrgId=%s, LocalMemberId=%s", in.LocalOrgId, in.LocalMemberId)

	return &pb.AddMemberResp{
		Success: true,
		Message: "成员添加成功",
	}, nil
}
