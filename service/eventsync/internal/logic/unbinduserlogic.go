package logic

import (
	"context"

	"github.com/rezeropoint/nexlyn/service/eventsync/internal/svc"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnbindUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnbindUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnbindUserLogic {
	return &UnbindUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 解绑用户映射
func (l *UnbindUserLogic) UnbindUser(in *pb.UnbindUserReq) (*pb.UnbindUserResp, error) {
	// 1. 参数验证
	if in.TenantId == "" || in.LocalUserId == "" {
		l.Errorf("参数验证失败: TenantId=%s, LocalUserId=%s", in.TenantId, in.LocalUserId)
		return &pb.UnbindUserResp{
			Success:   false,
			Message:   "租户ID、本地用户ID不能为空",
			ErrorCode: "INVALID_PARAMS",
		}, nil
	}

	// 2. 调用 AdminEngine 解绑用户
	l.Infof("解绑用户: TenantId=%s, LocalUserId=%s", in.TenantId, in.LocalUserId)
	err := l.svcCtx.AdminEngine.UnbindUser(
		l.ctx,
		in.TenantId,
		in.LocalUserId,
	)

	if err != nil {
		l.Errorf("解绑用户失败: %v (TenantId=%s, LocalUserId=%s)", err, in.TenantId, in.LocalUserId)

		// 判断错误类型
		errMsg := err.Error()
		errorCode := "OPERATION_FAILED"
		if svc.Contains(errMsg, "not found") || svc.Contains(errMsg, "不存在") {
			errorCode = "MAPPING_NOT_FOUND"
		}

		return &pb.UnbindUserResp{
			Success:   false,
			Message:   errMsg,
			ErrorCode: errorCode,
		}, nil
	}

	l.Infof("成功解绑用户: LocalUserId=%s", in.LocalUserId)

	return &pb.UnbindUserResp{
		Success: true,
		Message: "用户解绑成功",
	}, nil
}
