package logic

import (
	"context"

	"github.com/rezeropoint/nexlyn/service/eventsync/internal/svc"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBindUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindUserLogic {
	return &BindUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 绑定已存在的远程用户
func (l *BindUserLogic) BindUser(in *pb.BindUserReq) (*pb.BindUserResp, error) {
	// 1. 参数验证
	if in.TenantId == "" || in.LocalUserId == "" || in.RemoteUserId <= 0 {
		l.Errorf("参数验证失败: TenantId=%s, LocalUserId=%s, RemoteUserId=%d",
			in.TenantId, in.LocalUserId, in.RemoteUserId)
		return &pb.BindUserResp{
			Success:   false,
			Message:   "租户ID、本地用户ID不能为空，远程用户ID必须大于0",
			ErrorCode: "INVALID_PARAMS",
		}, nil
	}

	// 2. 调用 AdminEngine 绑定用户
	l.Infof("绑定用户: TenantId=%s, LocalUserId=%s, RemoteUserId=%d", in.TenantId, in.LocalUserId, in.RemoteUserId)
	err := l.svcCtx.AdminEngine.BindUser(
		l.ctx,
		in.TenantId,
		in.LocalUserId,
		int(in.RemoteUserId),
	)

	if err != nil {
		l.Errorf("绑定用户失败: %v (TenantId=%s, LocalUserId=%s, RemoteUserId=%d)",
			err, in.TenantId, in.LocalUserId, in.RemoteUserId)

		// 判断错误类型
		errMsg := err.Error()
		errorCode := "OPERATION_FAILED"
		if svc.Contains(errMsg, "not found") || svc.Contains(errMsg, "不存在") {
			errorCode = "NOT_FOUND"
		} else if svc.Contains(errMsg, "already exists") || svc.Contains(errMsg, "已存在") {
			errorCode = "MAPPING_EXISTS"
		}

		return &pb.BindUserResp{
			Success:   false,
			Message:   errMsg,
			ErrorCode: errorCode,
		}, nil
	}

	l.Infof("成功绑定用户: LocalUserId=%s -> RemoteUserId=%d", in.LocalUserId, in.RemoteUserId)

	return &pb.BindUserResp{
		Success: true,
		Message: "用户绑定成功",
	}, nil
}
