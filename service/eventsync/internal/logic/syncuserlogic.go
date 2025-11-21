package logic

import (
	"context"

	"github.com/rezeropoint/nexlyn/service/eventsync/internal/svc"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncUserLogic {
	return &SyncUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 同步单个用户到Skylark
func (l *SyncUserLogic) SyncUser(in *pb.SyncUserReq) (*pb.SyncUserResp, error) {
	// 1. 参数验证
	if in.TenantId == "" || in.LocalUserId == "" || in.Name == "" {
		l.Errorf("参数验证失败: TenantId=%s, LocalUserId=%s, Name=%s", in.TenantId, in.LocalUserId, in.Name)
		return &pb.SyncUserResp{
			Success:   false,
			Message:   "租户ID、用户ID和姓名不能为空",
			ErrorCode: "INVALID_PARAMS",
		}, nil
	}

	// 2. 调用 AdminEngine 创建用户（v2.1.3+参数为string类型）
	err := l.svcCtx.AdminEngine.CreateUser(
		l.ctx,
		in.TenantId,
		in.LocalUserId,
		in.Name,
		in.Identifier,
		in.Phone,
		in.Openid,
	)
	if err != nil {
		l.Errorf("创建Skylark用户失败: %v (TenantId=%s, LocalUserId=%s)", err, in.TenantId, in.LocalUserId)
		return &pb.SyncUserResp{
			Success:   false,
			Message:   err.Error(),
			ErrorCode: "CREATE_FAILED",
		}, nil
	}

	l.Infof("成功同步用户到Skylark: LocalUserId=%s", in.LocalUserId)

	return &pb.SyncUserResp{
		Success:      true,
		Message:      "用户同步成功",
		RemoteUserId: in.LocalUserId, // v2.1.2+使用LocalUserId作为唯一标识
	}, nil
}
