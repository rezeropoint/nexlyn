package organization

import (
	"context"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DeleteOrganizationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除组织
func NewDeleteOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteOrganizationLogic {
	return &DeleteOrganizationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteOrganizationLogic) DeleteOrganization(req *types.DeleteOrganizationRequest) (resp *types.DeleteOrganizationResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "delete_organization"),
		logx.Field("status", "started"),
		logx.Field("org_id", req.Id),
		logx.Field("force_delete", req.ForceDelete),
	).Info("开始删除组织")

	// 获取当前用户信息
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "delete_organization"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从 JWT 获取用户信息失败")

		return &types.DeleteOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 权限验证：检查组织删除权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		currentUser.UserKey,
		currentUser.TenantKey,
		core.Permission{Resource: core.ResourceOrganization, Action: core.ActionDelete},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "delete_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.DeleteOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}
	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "delete_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
		).Error("权限不足")

		return &types.DeleteOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足"},
		}, nil
	}

	// 获取租户ID
	var tenantId string
	getTenantQuery := `SELECT id FROM system_tenants WHERE tenant_key = $1 AND status = 'active'`
	err = l.svcCtx.DBConn.QueryRowCtx(l.ctx, &tenantId, getTenantQuery, currentUser.TenantKey)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "delete_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("error", err.Error()),
		).Error("获取租户信息失败")

		return &types.DeleteOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantQuery, err)},
		}, nil
	}

	// 使用事务处理删除操作
	var deletedCount int

	err = l.svcCtx.DBConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 验证组织存在且属于当前租户
		var orgInfo struct {
			Path string `db:"path"`
		}
		checkOrgQuery := `SELECT path FROM system_organizations WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`
		err := session.QueryRowCtx(ctx, &orgInfo, checkOrgQuery, req.Id, tenantId)
		if err != nil {
			return fmt.Errorf(config.ErrMsgOrganizationNotFound)
		}

		// 检查是否有子组织
		var childCount int
		checkChildQuery := `SELECT COUNT(*) FROM system_organizations WHERE path LIKE $1 AND id != $2 AND tenant_id = $3 AND deleted_at IS NULL`
		err = session.QueryRowCtx(ctx, &childCount, checkChildQuery, orgInfo.Path+"%", req.Id, tenantId)
		if err != nil {
			return fmt.Errorf(config.ErrMsgDatabaseQuery)
		}

		if childCount > 0 && !req.ForceDelete {
			return fmt.Errorf("存在子组织，请先删除子组织或使用强制删除")
		}

		// 根据是否强制删除，决定删除范围
		var deleteQuery string
		var deleteArgs []interface{}

		if req.ForceDelete {
			// 级联删除：删除当前组织及所有子组织
			deleteQuery = `
				UPDATE system_organizations 
				SET deleted_at = CURRENT_TIMESTAMP 
				WHERE (path = $1 OR path LIKE $2) AND tenant_id = $3 AND deleted_at IS NULL
			`
			deleteArgs = []interface{}{orgInfo.Path, orgInfo.Path + "%", tenantId}
		} else {
			// 只删除当前组织
			deleteQuery = `
				UPDATE system_organizations 
				SET deleted_at = CURRENT_TIMESTAMP 
				WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
			`
			deleteArgs = []interface{}{req.Id, tenantId}
		}

		result, err := session.ExecCtx(ctx, deleteQuery, deleteArgs...)
		if err != nil {
			return fmt.Errorf(config.ErrMsgDatabaseDelete)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf(config.ErrMsgDatabaseQuery)
		}
		deletedCount = int(affected)

		// 软删除相关的用户组织关系
		if req.ForceDelete {
			_, err = session.ExecCtx(ctx, `
				UPDATE system_user_org_relations 
				SET deleted_at = CURRENT_TIMESTAMP 
				WHERE org_id IN (
					SELECT id FROM system_organizations 
					WHERE (path = $1 OR path LIKE $2) AND tenant_id = $3
				) AND deleted_at IS NULL
			`, orgInfo.Path, orgInfo.Path+"%", tenantId)
		} else {
			_, err = session.ExecCtx(ctx, `
				UPDATE system_user_org_relations 
				SET deleted_at = CURRENT_TIMESTAMP 
				WHERE org_id = $1 AND deleted_at IS NULL
			`, req.Id)
		}

		if err != nil {
			return fmt.Errorf(config.ErrMsgDatabaseDelete)
		}

		// ========================================
		// 事务内同步删除 Skylark 组织
		// ========================================
		// 如果启用了 Skylark 同步，在事务内删除远程组织
		if l.svcCtx.Config.SkylarkSyncEnabled && l.svcCtx.AdminSyncClient != nil {
			syncCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			syncResp, syncErr := l.svcCtx.AdminSyncClient.DeleteOrganization(syncCtx, &pb.DeleteOrganizationReq{
				TenantId:   currentUser.TenantId,
				LocalOrgId: req.Id,
			})

			if syncErr != nil {
				logx.WithContext(ctx).WithFields(
					logx.Field("service", l.svcCtx.Config.RestConf.Name),
					logx.Field("pod", l.svcCtx.PodName),
					logx.Field("module", "organization"),
					logx.Field("operation", "delete_from_skylark"),
					logx.Field("status", "failed"),
					logx.Field("org_id", req.Id),
					logx.Field("error", syncErr.Error()),
				).Error("Skylark 组织删除失败，事务将回滚")

				return fmt.Errorf("删除 Skylark 组织失败: %v", syncErr)
			}

			if !syncResp.Success {
				logx.WithContext(ctx).WithFields(
					logx.Field("service", l.svcCtx.Config.RestConf.Name),
					logx.Field("pod", l.svcCtx.PodName),
					logx.Field("module", "organization"),
					logx.Field("operation", "delete_from_skylark"),
					logx.Field("status", "failed"),
					logx.Field("org_id", req.Id),
					logx.Field("message", syncResp.Message),
					logx.Field("error_code", syncResp.ErrorCode),
				).Error("Skylark 返回删除失败，事务将回滚")

				return fmt.Errorf("删除 Skylark 组织失败: %s", syncResp.Message)
			}

			logx.WithContext(ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "organization"),
				logx.Field("operation", "delete_from_skylark"),
				logx.Field("status", "success"),
				logx.Field("org_id", req.Id),
			).Info("组织已成功从 Skylark 删除")
		}

		return nil
	})

	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "delete_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("org_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("删除组织失败（包含 Skylark 删除失败时的自动回滚）")

		return &types.DeleteOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: err.Error()},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "delete_organization"),
		logx.Field("status", "success"),
		logx.Field("user_key", currentUser.UserKey),
		logx.Field("tenant_key", currentUser.TenantKey),
		logx.Field("org_id", req.Id),
		logx.Field("force_delete", req.ForceDelete),
		logx.Field("deleted_count", deletedCount),
	).Info("删除组织成功")

	return &types.DeleteOrganizationResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "删除成功"},
		Data: struct {
			DeletedCount int `json:"deletedCount"` // 删除的组织数量（包含子组织）
		}{
			DeletedCount: deletedCount,
		},
	}, nil
}
