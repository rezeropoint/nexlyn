package logic

import (
	"context"

	"github.com/rezeropoint/nexlyn/service/eventsync/internal/svc"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/rezeropoint/go-skylark/v2/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateOrganizationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOrganizationLogic {
	return &UpdateOrganizationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新组织信息（支持设置管理员）
func (l *UpdateOrganizationLogic) UpdateOrganization(in *pb.UpdateOrganizationReq) (*pb.UpdateOrganizationResp, error) {
	// 1. 参数验证
	if in.TenantId == "" || in.LocalOrgId == "" {
		l.Errorf("参数验证失败: TenantId=%s, LocalOrgId=%s", in.TenantId, in.LocalOrgId)
		return &pb.UpdateOrganizationResp{
			Success:   false,
			Message:   "租户ID和组织ID不能为空",
			ErrorCode: "INVALID_PARAMS",
		}, nil
	}

	// 检查是否至少有一个字段需要更新
	if !in.UpdateName && !in.UpdateDescription && !in.UpdateManager {
		l.Errorf("至少需要更新一个字段: TenantId=%s, LocalOrgId=%s", in.TenantId, in.LocalOrgId)
		return &pb.UpdateOrganizationResp{
			Success:   false,
			Message:   "至少需要更新一个字段",
			ErrorCode: "INVALID_PARAMS",
		}, nil
	}

	// 2. 构造SDK请求（空字符串表示不修改）
	req := &core.UpdateOrganizationRequest{
		TenantID:   in.TenantId,
		LocalOrgID: in.LocalOrgId,
		Name:       "", // 默认不修改
		ManagerID:  "", // 默认不修改
	}

	if in.UpdateName {
		req.Name = in.Name
		l.Infof("更新组织名称: LocalOrgId=%s, Name=%s", in.LocalOrgId, in.Name)
	}

	if in.UpdateManager {
		req.ManagerID = in.ManagerId
		l.Infof("更新组织管理员: LocalOrgId=%s, ManagerID=%s", in.LocalOrgId, in.ManagerId)
	}

	// 注意：v2.3.1 不支持更新 Description，忽略该字段
	if in.UpdateDescription {
		l.Infof("警告: UpdateOrganization 不支持更新 Description 字段（LocalOrgId=%s）", in.LocalOrgId)
	}

	// 3. 调用AdminEngine
	err := l.svcCtx.AdminEngine.UpdateOrganization(l.ctx, req)
	if err != nil {
		l.Errorf("更新Skylark组织失败: %v (TenantId=%s, LocalOrgId=%s)", err, in.TenantId, in.LocalOrgId)
		return &pb.UpdateOrganizationResp{
			Success:   false,
			Message:   err.Error(),
			ErrorCode: "UPDATE_FAILED",
		}, nil
	}

	l.Infof("成功更新Skylark组织: LocalOrgId=%s", in.LocalOrgId)

	return &pb.UpdateOrganizationResp{
		Success: true,
		Message: "组织更新成功",
	}, nil
}
