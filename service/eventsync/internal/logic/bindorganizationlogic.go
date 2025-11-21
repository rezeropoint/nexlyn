package logic

import (
	"context"

	"github.com/rezeropoint/nexlyn/service/eventsync/internal/svc"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindOrganizationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBindOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindOrganizationLogic {
	return &BindOrganizationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 绑定已存在的远程组织
func (l *BindOrganizationLogic) BindOrganization(in *pb.BindOrganizationReq) (*pb.BindOrganizationResp, error) {
	// 1. 参数验证
	if in.TenantId == "" || in.LocalOrgId == "" || in.RemoteOrgId <= 0 {
		l.Errorf("参数验证失败: TenantId=%s, LocalOrgId=%s, RemoteOrgId=%d",
			in.TenantId, in.LocalOrgId, in.RemoteOrgId)
		return &pb.BindOrganizationResp{
			Success:   false,
			Message:   "租户ID、本地组织ID不能为空，远程组织ID必须大于0",
			ErrorCode: "INVALID_PARAMS",
		}, nil
	}

	// 2. 调用 AdminEngine 绑定组织
	l.Infof("绑定组织: TenantId=%s, LocalOrgId=%s, RemoteOrgId=%d", in.TenantId, in.LocalOrgId, in.RemoteOrgId)
	err := l.svcCtx.AdminEngine.BindOrganization(
		l.ctx,
		in.TenantId,
		in.LocalOrgId,
		int(in.RemoteOrgId),
	)

	if err != nil {
		l.Errorf("绑定组织失败: %v (TenantId=%s, LocalOrgId=%s, RemoteOrgId=%d)",
			err, in.TenantId, in.LocalOrgId, in.RemoteOrgId)

		// 判断错误类型
		errMsg := err.Error()
		errorCode := "OPERATION_FAILED"
		if svc.Contains(errMsg, "not found") || svc.Contains(errMsg, "不存在") {
			errorCode = "NOT_FOUND"
		} else if svc.Contains(errMsg, "already exists") || svc.Contains(errMsg, "已存在") {
			errorCode = "MAPPING_EXISTS"
		}

		return &pb.BindOrganizationResp{
			Success:   false,
			Message:   errMsg,
			ErrorCode: errorCode,
		}, nil
	}

	l.Infof("成功绑定组织: LocalOrgId=%s -> RemoteOrgId=%d", in.LocalOrgId, in.RemoteOrgId)

	return &pb.BindOrganizationResp{
		Success: true,
		Message: "组织绑定成功",
	}, nil
}
