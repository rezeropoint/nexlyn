package organization

import (
	"context"
	"fmt"
	"strings"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type MoveOrganizationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 移动组织
func NewMoveOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MoveOrganizationLogic {
	return &MoveOrganizationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MoveOrganizationLogic) MoveOrganization(req *types.MoveOrganizationRequest) (resp *types.MoveOrganizationResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "move_organization"),
		logx.Field("status", "started"),
		logx.Field("org_id", req.Id),
		logx.Field("new_parent_id", req.NewParentId),
	).Info("开始移动组织")

	// 获取当前用户信息
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "move_organization"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从JWT获取用户信息失败")

		return &types.MoveOrganizationResponse{
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
			logx.Field("operation", "move_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.MoveOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}
	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "move_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
		).Error("权限不足")

		return &types.MoveOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足"},
		}, nil
	}

	// 获取源组织信息
	var sourceOrg struct {
		Id       string  `db:"id"`
		Path     string  `db:"path"`
		Level    int     `db:"level"`
		ParentId *string `db:"parent_id"`
	}

	sourceQuery := `
		SELECT id, path, level, parent_id 
		FROM system_organizations 
		WHERE id = $1 AND tenant_id = (SELECT id FROM system_tenants WHERE tenant_key = $2) 
		  AND deleted_at IS NULL
	`
	err = l.svcCtx.DBConn.QueryRowCtx(l.ctx, &sourceOrg, sourceQuery, req.Id, currentUser.TenantKey)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "move_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("org_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("源组织不存在或查询失败")

		return &types.MoveOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 404, Msg: "组织不存在"},
		}, nil
	}

	// 如果新父组织ID与当前父组织ID相同，无需移动
	if (req.NewParentId == "" && sourceOrg.ParentId == nil) ||
		(req.NewParentId != "" && sourceOrg.ParentId != nil && req.NewParentId == *sourceOrg.ParentId) {
		return &types.MoveOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 0, Msg: "组织位置未变更"},
			Data: struct {
				NewPath string `json:"newPath"`
			}{
				NewPath: sourceOrg.Path,
			},
		}, nil
	}

	// 验证新父组织（如果指定）
	var newParentPath string
	var newParentLevel int
	var newParentId interface{}

	if req.NewParentId != "" {
		var parentOrg struct {
			Id    string `db:"id"`
			Path  string `db:"path"`
			Level int    `db:"level"`
		}

		parentQuery := `
			SELECT id, path, level 
			FROM system_organizations 
			WHERE id = $1 AND tenant_id = (SELECT id FROM system_tenants WHERE tenant_key = $2) 
			  AND deleted_at IS NULL
		`
		err = l.svcCtx.DBConn.QueryRowCtx(l.ctx, &parentOrg, parentQuery, req.NewParentId, currentUser.TenantKey)
		if err != nil {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "organization"),
				logx.Field("operation", "move_organization"),
				logx.Field("status", "failed"),
				logx.Field("user_key", currentUser.UserKey),
				logx.Field("tenant_key", currentUser.TenantKey),
				logx.Field("org_id", req.Id),
				logx.Field("new_parent_id", req.NewParentId),
				logx.Field("error", err.Error()),
			).Error("新父组织不存在或查询失败")

			return &types.MoveOrganizationResponse{
				BaseResponse: types.BaseResponse{Code: 404, Msg: "目标父组织不存在"},
			}, nil
		}

		// 防止循环引用：不能将组织移动到其自身或子组织下
		if parentOrg.Path == sourceOrg.Path || strings.HasPrefix(parentOrg.Path, sourceOrg.Path) {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "organization"),
				logx.Field("operation", "move_organization"),
				logx.Field("status", "failed"),
				logx.Field("user_key", currentUser.UserKey),
				logx.Field("tenant_key", currentUser.TenantKey),
				logx.Field("org_id", req.Id),
				logx.Field("new_parent_id", req.NewParentId),
				logx.Field("source_path", sourceOrg.Path),
				logx.Field("parent_path", parentOrg.Path),
			).Error("不能将组织移动到其自身或子组织下")

			return &types.MoveOrganizationResponse{
				BaseResponse: types.BaseResponse{Code: 400, Msg: "不能将组织移动到其自身或子组织下"},
			}, nil
		}

		// 检查层级深度限制
		if parentOrg.Level >= 10 {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "organization"),
				logx.Field("operation", "move_organization"),
				logx.Field("status", "failed"),
				logx.Field("user_key", currentUser.UserKey),
				logx.Field("tenant_key", currentUser.TenantKey),
				logx.Field("org_id", req.Id),
				logx.Field("new_parent_id", req.NewParentId),
				logx.Field("parent_level", parentOrg.Level),
			).Error("移动后层级深度将超过限制")

			return &types.MoveOrganizationResponse{
				BaseResponse: types.BaseResponse{Code: 400, Msg: "移动后层级深度将超过限制（最大10层）"},
			}, nil
		}

		newParentPath = parentOrg.Path
		newParentLevel = parentOrg.Level
		newParentId = req.NewParentId
	} else {
		// 移动到根级
		newParentPath = "/"
		newParentLevel = 0
		newParentId = nil
	}

	// 计算新路径和层级差异
	newPath := newParentPath + req.Id + "/"
	if newParentPath == "/" {
		newPath = "/" + req.Id + "/"
	}
	levelDiff := (newParentLevel + 1) - sourceOrg.Level

	// 使用事务执行移动操作
	err = l.svcCtx.DBConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 使用递归CTE更新组织树的路径和层级
		updateQuery := `
			WITH RECURSIVE org_tree AS (
				-- 查找要移动的组织及其所有子组织
				SELECT id, path, level
				FROM system_organizations
				WHERE path LIKE $1 || '%'
				  AND tenant_id = (SELECT id FROM system_tenants WHERE tenant_key = $2)
				  AND deleted_at IS NULL
			)
			UPDATE system_organizations 
			SET 
				path = $3 || SUBSTRING(path FROM LENGTH($1) + 1),
				level = level + $4,
				parent_id = CASE WHEN id = $5 THEN $6 ELSE parent_id END,
				sort_order = CASE WHEN id = $5 AND $7 > 0 THEN $7 ELSE sort_order END,
				updated_at = CURRENT_TIMESTAMP
			WHERE id IN (SELECT id FROM org_tree)
		`

		_, err := session.ExecCtx(ctx, updateQuery,
			sourceOrg.Path, currentUser.TenantKey, newPath, levelDiff,
			req.Id, newParentId, req.NewSortOrder)
		if err != nil {
			return fmt.Errorf("更新组织路径失败: %v", err)
		}

		return nil
	})

	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "move_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("org_id", req.Id),
			logx.Field("new_parent_id", req.NewParentId),
			logx.Field("error", err.Error()),
		).Error("移动组织失败")

		return &types.MoveOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgDatabaseUpdate, err)},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "move_organization"),
		logx.Field("status", "success"),
		logx.Field("user_key", currentUser.UserKey),
		logx.Field("tenant_key", currentUser.TenantKey),
		logx.Field("org_id", req.Id),
		logx.Field("new_parent_id", req.NewParentId),
		logx.Field("old_path", sourceOrg.Path),
		logx.Field("new_path", newPath),
		logx.Field("level_diff", levelDiff),
	).Info("移动组织成功")

	return &types.MoveOrganizationResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "移动成功"},
		Data: struct {
			NewPath string `json:"newPath"`
		}{
			NewPath: newPath,
		},
	}, nil
}
