package organization

import (
	"context"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"
	pb "github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UpdateOrganizationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// orgInfo 用于存储组织旧值以便同步判断
type orgInfo struct {
	Name        string  `db:"name"`
	Description string  `db:"description"`
	ManagerId   *string `db:"manager_id"`
}

// 更新组织
func NewUpdateOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOrganizationLogic {
	return &UpdateOrganizationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateOrganizationLogic) UpdateOrganization(req *types.UpdateOrganizationRequest) (resp *types.UpdateOrganizationResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "update_organization"),
		logx.Field("status", "started"),
		logx.Field("org_id", req.Id),
	).Info("开始更新组织")

	// 获取当前用户信息
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "update_organization"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从 JWT 获取用户信息失败")

		return &types.UpdateOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 权限验证：检查组织写入权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		currentUser.UserKey,
		currentUser.TenantKey,
		core.Permission{Resource: core.ResourceOrganization, Action: core.ActionWrite},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "update_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.UpdateOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}
	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "update_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
		).Error("权限不足")

		return &types.UpdateOrganizationResponse{
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
			logx.Field("operation", "update_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("error", err.Error()),
		).Error("获取租户信息失败")

		return &types.UpdateOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantQuery, err)},
		}, nil
	}

	// 验证组织存在且属于当前租户，同时获取旧的值用于同步判断
	var oldOrgInfo orgInfo
	checkOrgQuery := `SELECT name, description, manager_id FROM system_organizations WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`
	err = l.svcCtx.DBConn.QueryRowCtx(l.ctx, &oldOrgInfo, checkOrgQuery, req.Id, tenantId)
	if err != nil {
		// 组织不存在
		if err.Error() == "sql: no rows in result set" {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "organization"),
				logx.Field("operation", "update_organization"),
				logx.Field("status", "failed"),
				logx.Field("user_key", currentUser.UserKey),
				logx.Field("tenant_key", currentUser.TenantKey),
				logx.Field("org_id", req.Id),
			).Error("组织不存在或不属于当前租户")

			return &types.UpdateOrganizationResponse{
				BaseResponse: types.BaseResponse{Code: 404, Msg: "组织不存在或不属于当前租户"},
			}, nil
		}

		// 其他数据库错误
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "update_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("org_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("验证组织存在性失败")

		return &types.UpdateOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgOrganizationQuery, err)},
		}, nil
	}

	// 使用事务处理组织更新
	err = l.svcCtx.DBConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 更新组织信息
		updateQuery := `
			UPDATE system_organizations SET
				name = $1, type = $2, description = $3,
				sort_order = $4, manager_id = $5, status = $6, updated_at = CURRENT_TIMESTAMP
			WHERE id = $7 AND tenant_id = $8 AND deleted_at IS NULL
		`

		_, err := session.ExecCtx(ctx, updateQuery,
			req.Name, req.Type, req.Description,
			req.SortOrder, req.ManagerId, req.Status, req.Id, tenantId)
		if err != nil {
			return fmt.Errorf(config.ErrMsgOrganizationUpdate)
		}

		// 同步到 Skylark（事务内，失败则回滚）
		if l.svcCtx.Config.SkylarkSyncEnabled && l.svcCtx.AdminSyncClient != nil {
			syncCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			// 构建更新请求（只同步变化的字段）
			updateReq := &pb.UpdateOrganizationReq{
				TenantId:    tenantId,
				LocalOrgId:  req.Id,
				Name:        req.Name,
				Description: req.Description,
			}

			// 判断哪些字段发生了变化
			if oldOrgInfo.Name != req.Name {
				updateReq.UpdateName = true
			}
			if oldOrgInfo.Description != req.Description {
				updateReq.UpdateDescription = true
			}
			if (oldOrgInfo.ManagerId == nil && req.ManagerId != "") || (oldOrgInfo.ManagerId != nil && *oldOrgInfo.ManagerId != req.ManagerId) {
				updateReq.UpdateManager = true
				updateReq.ManagerId = req.ManagerId
			}

			// 如果有字段变化，调用 Skylark 同步
			if updateReq.UpdateName || updateReq.UpdateDescription || updateReq.UpdateManager {
				syncResp, syncErr := l.svcCtx.AdminSyncClient.UpdateOrganization(syncCtx, updateReq)

				if syncErr != nil {
					logx.WithContext(ctx).WithFields(
						logx.Field("service", l.svcCtx.Config.RestConf.Name),
						logx.Field("pod", l.svcCtx.PodName),
						logx.Field("module", "organization"),
						logx.Field("operation", "sync_update_to_skylark"),
						logx.Field("status", "failed"),
						logx.Field("org_id", req.Id),
						logx.Field("error", syncErr.Error()),
					).Error("同步更新组织到 Skylark 失败，回滚事务")
					return fmt.Errorf("更新组织到 Skylark 失败: %v", syncErr)
				}

				if !syncResp.Success {
					logx.WithContext(ctx).WithFields(
						logx.Field("service", l.svcCtx.Config.RestConf.Name),
						logx.Field("pod", l.svcCtx.PodName),
						logx.Field("module", "organization"),
						logx.Field("operation", "sync_update_to_skylark"),
						logx.Field("status", "failed"),
						logx.Field("org_id", req.Id),
						logx.Field("message", syncResp.Message),
						logx.Field("error_code", syncResp.ErrorCode),
					).Error("同步更新组织到 Skylark 失败，回滚事务")
					return fmt.Errorf("更新组织到 Skylark 失败: %s", syncResp.Message)
				}

				logx.WithContext(ctx).WithFields(
					logx.Field("service", l.svcCtx.Config.RestConf.Name),
					logx.Field("pod", l.svcCtx.PodName),
					logx.Field("module", "organization"),
					logx.Field("operation", "sync_update_to_skylark"),
					logx.Field("status", "success"),
					logx.Field("org_id", req.Id),
				).Info("成功同步更新组织到 Skylark")
			}
		}

		return nil
	})

	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "update_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("org_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("更新组织失败")

		return &types.UpdateOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: err.Error()},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "update_organization"),
		logx.Field("status", "success"),
		logx.Field("user_key", currentUser.UserKey),
		logx.Field("tenant_key", currentUser.TenantKey),
		logx.Field("org_id", req.Id),
	).Info("更新组织成功")

	return &types.UpdateOrganizationResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "更新成功"},
	}, nil
}
