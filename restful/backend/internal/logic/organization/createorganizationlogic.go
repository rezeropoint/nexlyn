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

type CreateOrganizationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建组织
func NewCreateOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrganizationLogic {
	return &CreateOrganizationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrganizationLogic) CreateOrganization(req *types.CreateOrganizationRequest) (resp *types.CreateOrganizationResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "create_organization"),
		logx.Field("status", "started"),
		logx.Field("org_code", req.Code),
		logx.Field("org_name", req.Name),
		logx.Field("org_type", req.Type),
	).Info("开始创建组织")

	// 获取当前用户信息
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "create_organization"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从 JWT 获取用户信息失败")

		return &types.CreateOrganizationResponse{
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
			logx.Field("operation", "create_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.CreateOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}
	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "create_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
		).Error("权限不足")

		return &types.CreateOrganizationResponse{
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
			logx.Field("operation", "create_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("error", err.Error()),
		).Error("获取租户信息失败")

		return &types.CreateOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantQuery, err)},
		}, nil
	}

	// 使用事务处理组织创建
	var orgId, orgPath string
	var skylarkRemoteOrgID string

	err = l.svcCtx.DBConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 检查组织代码在租户内的唯一性
		var existingCount int
		checkCodeQuery := `SELECT COUNT(*) FROM system_organizations WHERE code = $1 AND tenant_id = $2 AND deleted_at IS NULL`
		err := session.QueryRowCtx(ctx, &existingCount, checkCodeQuery, req.Code, tenantId)
		if err != nil {
			return fmt.Errorf(config.ErrMsgDatabaseQuery)
		}
		if existingCount > 0 {
			return fmt.Errorf(config.ErrMsgOrganizationCodeExists)
		}

		// 构建父组织信息和路径
		var parentPath string
		var level int = 1

		if req.ParentId != "" {
			// 验证父组织存在且属于同一租户
			var parentInfo struct {
				Path  string `db:"path"`
				Level int    `db:"level"`
			}
			getParentQuery := `SELECT path, level FROM system_organizations WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`
			err := session.QueryRowCtx(ctx, &parentInfo, getParentQuery, req.ParentId, tenantId)
			if err != nil {
				return fmt.Errorf(config.ErrMsgOrganizationNotFound)
			}

			parentPath = parentInfo.Path
			level = parentInfo.Level + 1

			// 检查层级限制（不超过10层）
			if level > 10 {
				return fmt.Errorf(config.ErrMsgOrganizationLevelExceeded)
			}
		}

		// 插入组织基本信息
		insertQuery := `
			INSERT INTO system_organizations (
				code, name, type, description,
				parent_id, level, sort_order, status, tenant_id,
				manager_id
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, 'active', $8, $9
			) RETURNING id
		`

		// 处理可空的 UUID 字段：空字符串转为 NULL
		var parentId, managerId interface{}
		if req.ParentId != "" {
			parentId = req.ParentId
		}
		if req.ManagerId != "" {
			managerId = req.ManagerId
		}

		err = session.QueryRowCtx(ctx, &orgId, insertQuery,
			req.Code, req.Name, req.Type, req.Description,
			parentId, level, req.SortOrder, tenantId, managerId)
		if err != nil {
			return fmt.Errorf(config.ErrMsgDatabaseInsert)
		}

		// 构建路径编码：parent_path + "/" + org_id + "/"
		if parentPath == "" {
			orgPath = "/" + orgId + "/"
		} else {
			// Ensure parentPath does not end with a slash before appending
			if len(parentPath) > 0 && parentPath[len(parentPath)-1] == '/' {
				orgPath = parentPath + orgId + "/"
			} else {
				orgPath = parentPath + "/" + orgId + "/"
			}
		}

		// 更新path字段
		updatePathQuery := `UPDATE system_organizations SET path = $1 WHERE id = $2`
		_, err = session.ExecCtx(ctx, updatePathQuery, orgPath, orgId)
		if err != nil {
			return fmt.Errorf(config.ErrMsgDatabaseUpdate)
		}

		// ========================================
		// 两阶段提交：同步到 Skylark（在事务内）
		// ========================================
		// 如果启用了 Skylark 同步，同步组织到 Skylark
		if l.svcCtx.Config.SkylarkSyncEnabled && l.svcCtx.AdminSyncClient != nil {
			// 确定创建人ID
			founderID := req.ManagerId
			if founderID == "" {
				founderID = currentUser.UserId // 使用当前用户作为创建人
			}

			syncCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			syncResp, syncErr := l.svcCtx.AdminSyncClient.SyncOrganization(syncCtx, &pb.SyncOrganizationReq{
				TenantId:         currentUser.TenantId,
				LocalOrgId:       orgId,
				ParentLocalOrgId: req.ParentId, // 空字符串表示根组织
				Name:             req.Name,
				Description:      req.Description,
				FounderId:        founderID,
			})

			if syncErr != nil {
				logx.WithContext(ctx).WithFields(
					logx.Field("service", l.svcCtx.Config.RestConf.Name),
					logx.Field("pod", l.svcCtx.PodName),
					logx.Field("module", "organization"),
					logx.Field("operation", "sync_to_skylark"),
					logx.Field("status", "failed"),
					logx.Field("org_id", orgId),
					logx.Field("error", syncErr.Error()),
				).Error("Skylark 组织同步失败，事务将回滚")

				return fmt.Errorf("组织同步到 Skylark 失败: %v", syncErr)
			}

			if !syncResp.Success {
				logx.WithContext(ctx).WithFields(
					logx.Field("service", l.svcCtx.Config.RestConf.Name),
					logx.Field("pod", l.svcCtx.PodName),
					logx.Field("module", "organization"),
					logx.Field("operation", "sync_to_skylark"),
					logx.Field("status", "failed"),
					logx.Field("org_id", orgId),
					logx.Field("message", syncResp.Message),
					logx.Field("error_code", syncResp.ErrorCode),
				).Error("Skylark 返回同步失败，事务将回滚")

				return fmt.Errorf("组织同步到 Skylark 失败: %s", syncResp.Message)
			}

			skylarkRemoteOrgID = syncResp.RemoteOrgId

			logx.WithContext(ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "organization"),
				logx.Field("operation", "sync_to_skylark"),
				logx.Field("status", "success"),
				logx.Field("org_id", orgId),
				logx.Field("remote_org_id", skylarkRemoteOrgID),
			).Info("组织已成功同步到 Skylark")
		}

		return nil
	})

	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "create_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("org_code", req.Code),
			logx.Field("error", err.Error()),
		).Error("创建组织失败")

		return &types.CreateOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: err.Error()},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "create_organization"),
		logx.Field("status", "success"),
		logx.Field("user_key", currentUser.UserKey),
		logx.Field("tenant_key", currentUser.TenantKey),
		logx.Field("org_id", orgId),
		logx.Field("org_code", req.Code),
		logx.Field("org_name", req.Name),
		logx.Field("path", orgPath),
	).Info("创建组织成功")

	return &types.CreateOrganizationResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "创建成功"},
		Data: struct {
			Id   string `json:"id"`   // 新创建的组织ID
			Code string `json:"code"` // 组织代码
			Path string `json:"path"` // 路径编码
		}{
			Id:   orgId,
			Code: req.Code,
			Path: orgPath,
		},
	}, nil
}
