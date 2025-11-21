package logic

import (
	"context"

	"github.com/rezeropoint/nexlyn/service/eventsync/internal/svc"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnbindOrganizationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnbindOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnbindOrganizationLogic {
	return &UnbindOrganizationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 解绑组织映射
func (l *UnbindOrganizationLogic) UnbindOrganization(in *pb.UnbindOrganizationReq) (*pb.UnbindOrganizationResp, error) {
	// 1. 参数验证
	if in.TenantId == "" || in.LocalOrgId == "" {
		l.Errorf("参数验证失败: TenantId=%s, LocalOrgId=%s", in.TenantId, in.LocalOrgId)
		return &pb.UnbindOrganizationResp{
			Success:   false,
			Message:   "租户ID、本地组织ID不能为空",
			ErrorCode: "INVALID_PARAMS",
		}, nil
	}

	// 2. 调用 AdminEngine 解绑组织
	l.Infof("解绑组织: TenantId=%s, LocalOrgId=%s", in.TenantId, in.LocalOrgId)
	err := l.svcCtx.AdminEngine.UnbindOrganization(
		l.ctx,
		in.TenantId,
		in.LocalOrgId,
	)

	if err != nil {
		l.Errorf("解绑组织失败: %v (TenantId=%s, LocalOrgId=%s)", err, in.TenantId, in.LocalOrgId)

		// 判断错误类型
		errMsg := err.Error()
		errorCode := "OPERATION_FAILED"
		if svc.Contains(errMsg, "not found") || svc.Contains(errMsg, "不存在") {
			errorCode = "MAPPING_NOT_FOUND"
		}

		return &pb.UnbindOrganizationResp{
			Success:   false,
			Message:   errMsg,
			ErrorCode: errorCode,
		}, nil
	}

	l.Infof("成功解绑组织: LocalOrgId=%s", in.LocalOrgId)

	return &pb.UnbindOrganizationResp{
		Success: true,
		Message: "组织解绑成功",
	}, nil
}
