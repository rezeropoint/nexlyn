package tenant

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新租户状态
func NewUpdateTenantStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantStatusLogic {
	return &UpdateTenantStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTenantStatusLogic) UpdateTenantStatus(req *types.UpdateTenantStatusRequest) (resp *types.UpdateTenantStatusResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tenant"),
		logx.Field("operation", "update_tenant_status"),
		logx.Field("status", "started"),
		logx.Field("tenant_id", req.Id),
		logx.Field("target_status", req.Status),
	).Info("开始更新租户状态")

	// 获取当前操作者（用于权限验证和审计）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "update_tenant_status"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户失败")
		return &types.UpdateTenantStatusResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 权限验证：检查是否有租户状态管理权限
	hasStatusPermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, "*", core.Permission{Resource: core.ResourceTenant, Action: core.ActionWrite})
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "update_tenant_status"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")
		return &types.UpdateTenantStatusResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}
	if !hasStatusPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "update_tenant_status"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("权限不足：需要租户管理权限")
		return &types.UpdateTenantStatusResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要租户管理权限"},
		}, nil
	}

	// 验证状态值
	switch req.Status {
	case "active", "inactive", "deleted":
		// 有效状态
	default:
		return &types.UpdateTenantStatusResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "状态值无效：必须为active、inactive或deleted"},
		}, nil
	}

	// 检查租户是否存在
	var existingTenant struct {
		Id         string `db:"id"`
		TenantKey  string `db:"tenant_key"`
		CreatedBy  string `db:"created_by"`
		TenantName string `db:"tenant_name"`
	}
	getTenantQuery := `
		SELECT id, tenant_key, COALESCE(created_by, '') as created_by, tenant_name
		FROM system_tenants
		WHERE id = $1`

	err = l.svcCtx.DBConn.QueryRow(&existingTenant, getTenantQuery, req.Id)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "update_tenant_status"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("查询租户信息失败")
		return &types.UpdateTenantStatusResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantQuery, err)},
		}, nil
	}

	// 业务规则检查：超级管理员不能修改自己所属的租户状态
	if existingTenant.TenantKey == jwtUser.TenantKey {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "update_tenant_status"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("user_tenant_key", jwtUser.TenantKey),
			logx.Field("target_tenant_key", existingTenant.TenantKey),
			logx.Field("target_status", req.Status),
		).Error("超级管理员尝试修改自己所属租户的状态")
		return &types.UpdateTenantStatusResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "不能修改自己所属租户的状态"},
		}, nil
	}

	// 更新租户状态
	updateStatusQuery := `
		UPDATE system_tenants
		SET status = $1, updated_by = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3`
	_, err = l.svcCtx.DBConn.Exec(updateStatusQuery, req.Status, jwtUser.UserId, req.Id)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "update_tenant_status"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", req.Id),
			logx.Field("target_status", req.Status),
			logx.Field("error", err.Error()),
		).Error("更新租户状态失败")
		return &types.UpdateTenantStatusResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantUpdate, err)},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tenant"),
		logx.Field("operation", "update_tenant_status"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", req.Id),
		logx.Field("tenant_name", existingTenant.TenantName),
		logx.Field("target_status", req.Status),
		logx.Field("updated_by", jwtUser.UserId),
	).Info("更新租户状态成功")

	return &types.UpdateTenantStatusResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "更新租户状态成功",
		},
	}, nil
}
