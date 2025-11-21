package tenant

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DeleteTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除租户
func NewDeleteTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTenantLogic {
	return &DeleteTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTenantLogic) DeleteTenant(req *types.DeleteTenantRequest) (resp *types.DeleteTenantResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tenant"),
		logx.Field("operation", "delete_tenant"),
		logx.Field("status", "started"),
		logx.Field("tenant_id", req.Id),
	).Info("开始删除租户")

	// 获取当前操作者（用于权限验证和审计）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "delete_tenant"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户失败")
		return &types.DeleteTenantResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 权限验证：检查是否有租户删除权限（使用Casbinx）
	hasDeletePermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: core.ResourceTenant, Action: core.ActionDelete})
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "delete_tenant"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")
		return &types.DeleteTenantResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}

	if !hasDeletePermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "delete_tenant"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")
		return &types.DeleteTenantResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要租户删除权限"},
		}, nil
	}

	// 检查租户是否存在
	var existingTenant struct {
		Id         string `db:"id"`
		TenantKey  string `db:"tenant_key"`
		CreatedBy  string `db:"created_by"`
		TenantName string `db:"tenant_name"`
		Status     string `db:"status"`
	}
	getTenantQuery := `
		SELECT id, tenant_key, COALESCE(created_by, '') as created_by, tenant_name, status
		FROM system_tenants
		WHERE id = $1`

	err = l.svcCtx.DBConn.QueryRow(&existingTenant, getTenantQuery, req.Id)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "delete_tenant"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("查询租户信息失败")
		return &types.DeleteTenantResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantQuery, err)},
		}, nil
	}

	// 检查租户是否已经被删除
	if existingTenant.Status == "deleted" {
		return &types.DeleteTenantResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "租户已被删除"},
		}, nil
	}

	// 业务规则检查：不能删除自己所属的租户
	if existingTenant.TenantKey == jwtUser.TenantKey {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "delete_tenant"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("user_tenant_id", jwtUser.TenantKey),
			logx.Field("target_tenant_key", existingTenant.TenantKey),
		).Error("尝试删除自己的租户")
		return &types.DeleteTenantResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "不能删除自己所属的租户"},
		}, nil
	}

	// 检查跨租户操作权限：使用Casbinx检查是否可以访问目标租户
	if existingTenant.TenantKey != jwtUser.TenantKey {
		canAccessTenant, err := l.svcCtx.Casbinx.CanAccessTenant(jwtUser.UserKey, existingTenant.TenantKey)
		if err != nil {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "tenant"),
				logx.Field("operation", "delete_tenant"),
				logx.Field("status", "failed"),
				logx.Field("user_key", jwtUser.UserKey),
				logx.Field("target_tenant_key", existingTenant.TenantKey),
				logx.Field("error", err.Error()),
			).Error("检查租户访问权限失败")
			return &types.DeleteTenantResponse{
				BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
			}, nil
		}

		if !canAccessTenant {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "tenant"),
				logx.Field("operation", "delete_tenant"),
				logx.Field("status", "failed"),
				logx.Field("user_key", jwtUser.UserKey),
				logx.Field("user_tenant_id", jwtUser.TenantKey),
				logx.Field("target_tenant_key", existingTenant.TenantKey),
			).Error("跨租户权限不足")
			return &types.DeleteTenantResponse{
				BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：无法操作其他租户"},
			}, nil
		}

		// 进一步检查是否有对目标租户的删除权限
		hasDeletePermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, existingTenant.TenantKey, core.Permission{Resource: core.ResourceTenant, Action: core.ActionDelete})
		if err != nil {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "tenant"),
				logx.Field("operation", "delete_tenant"),
				logx.Field("status", "failed"),
				logx.Field("user_key", jwtUser.UserKey),
				logx.Field("target_tenant_key", existingTenant.TenantKey),
				logx.Field("error", err.Error()),
			).Error("检查租户删除权限失败")
			return &types.DeleteTenantResponse{
				BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
			}, nil
		}

		if !hasDeletePermission {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "tenant"),
				logx.Field("operation", "delete_tenant"),
				logx.Field("status", "failed"),
				logx.Field("user_key", jwtUser.UserKey),
				logx.Field("user_tenant_id", jwtUser.TenantKey),
				logx.Field("target_tenant_key", existingTenant.TenantKey),
			).Error("缺少租户删除权限")
			return &types.DeleteTenantResponse{
				BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：缺少租户删除权限"},
			}, nil
		}
	}

	// 使用sqlx进行数据库操作

	// 检查租户下是否有活跃用户
	var usersCountResult struct {
		Count int `db:"count"`
	}
	checkUsersQuery := "SELECT COUNT(*) as count FROM system_users WHERE tenant_id = $1 AND status = 'active'"
	err = l.svcCtx.DBConn.QueryRow(&usersCountResult, checkUsersQuery, existingTenant.Id)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "delete_tenant"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("检查租户下用户失败")
		return &types.DeleteTenantResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantUserCheck, err)},
		}, nil
	}

	if usersCountResult.Count > 0 {
		return &types.DeleteTenantResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "无法删除租户：租户下存在活跃用户"},
		}, nil
	}

	// 使用事务处理整个租户删除过程（标签删除 + 租户软删除 + 用户状态更新）
	err = l.svcCtx.DBConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 删除租户标签关联
		deleteTagQuery := "DELETE FROM system_tenant_tag_relations WHERE tenant_id = $1"
		_, err := session.ExecCtx(ctx, deleteTagQuery, existingTenant.Id)
		if err != nil {
			return fmt.Errorf("删除租户标签关联失败: %w", err)
		}

		// 软删除租户下的所有用户（将状态设置为deleted）
		deleteUsersQuery := `
			UPDATE system_users 
			SET status = 'deleted', updated_at = CURRENT_TIMESTAMP
			WHERE tenant_id = $1 AND status != 'deleted'`
		_, err = session.ExecCtx(ctx, deleteUsersQuery, existingTenant.Id)
		if err != nil {
			return fmt.Errorf("删除租户下用户失败: %w", err)
		}

		// 软删除租户（更新状态为删除）
		updateTenantQuery := `
			UPDATE system_tenants
			SET status = 'deleted', updated_by = $1, updated_at = CURRENT_TIMESTAMP
			WHERE id = $2`
		_, err = session.ExecCtx(ctx, updateTenantQuery, jwtUser.UserId, req.Id)
		if err != nil {
			return fmt.Errorf("软删除租户失败: %w", err)
		}

		return nil
	})

	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "delete_tenant"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("租户删除事务失败")
		return &types.DeleteTenantResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: fmt.Sprintf("删除租户失败: %v", err)},
		}, nil
	}

	// 清理相关的Casbin权限规则
	if l.svcCtx.Casbinx != nil {
		// 获取租户下的所有用户
		tenantUsers, err := l.svcCtx.Casbinx.GetUserTenants(existingTenant.TenantKey)
		if err != nil {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "tenant"),
				logx.Field("operation", "delete_tenant"),
				logx.Field("status", "warning"),
				logx.Field("tenant_id", req.Id),
				logx.Field("error", err.Error()),
			).Error("获取租户用户列表失败")
		} else {
			// 清理所有与该租户相关的用户权限
			for _, userKey := range tenantUsers {
				// 撤销用户对该租户的所有权限
				permissions, _ := l.svcCtx.Casbinx.GetDirectPermissionsSecure(jwtUser.UserKey, userKey, existingTenant.TenantKey)
				for _, perm := range permissions {
					_ = l.svcCtx.Casbinx.RevokePermission(jwtUser.UserKey, userKey, existingTenant.TenantKey, perm)
				}

				// 移除用户在该租户下的所有角色
				roles, _ := l.svcCtx.Casbinx.GetUserRoles(userKey, existingTenant.TenantKey)
				for _, roleKey := range roles {
					_ = l.svcCtx.Casbinx.RemoveRole(jwtUser.UserKey, userKey, roleKey, existingTenant.TenantKey)
				}
			}
		}
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tenant"),
		logx.Field("operation", "delete_tenant"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", req.Id),
		logx.Field("tenant_name", existingTenant.TenantName),
		logx.Field("deleted_by", jwtUser.UserId),
	).Info("删除租户成功")

	return &types.DeleteTenantResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "删除租户成功",
		},
	}, nil
}
