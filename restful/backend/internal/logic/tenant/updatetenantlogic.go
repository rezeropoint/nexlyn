package tenant

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

type UpdateTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新租户信息
func NewUpdateTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantLogic {
	return &UpdateTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTenantLogic) UpdateTenant(req *types.UpdateTenantRequest) (resp *types.UpdateTenantResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tenant"),
		logx.Field("operation", "update_tenant"),
		logx.Field("status", "started"),
		logx.Field("tenant_id", req.Id),
	).Info("开始更新租户信息")

	// 获取当前操作者（用于权限验证和审计）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "update_tenant"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户失败")
		return &types.UpdateTenantResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 权限验证：检查是否有租户更新权限（使用Casbinx）
	hasWritePermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: core.ResourceTenant, Action: core.ActionWrite})
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "update_tenant"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")
		return &types.UpdateTenantResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}

	if !hasWritePermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "update_tenant"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")
		return &types.UpdateTenantResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要租户更新权限"},
		}, nil
	}

	// 检查租户是否存在
	var existingTenant struct {
		Id           string `db:"id"`
		TenantKey    string `db:"tenant_key"`
		CreatedBy    string `db:"created_by"`
		TenantName   string `db:"tenant_name"`
		Description  string `db:"description"`
		ContactEmail string `db:"contact_email"`
	}
	getTenantQuery := `
		SELECT id, tenant_key, COALESCE(created_by, '') as created_by, tenant_name, COALESCE(description, '') as description, COALESCE(contact_email, '') as contact_email
		FROM system_tenants
		WHERE id = $1 AND status != 'deleted'`

	err = l.svcCtx.DBConn.QueryRow(&existingTenant, getTenantQuery, req.Id)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "update_tenant"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("查询租户信息失败")
		return &types.UpdateTenantResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantQuery, err)},
		}, nil
	}

	// 业务规则检查：不能操作自己的租户（使用Casbinx检查全局管理权限）
	hasGlobalTenantPermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, "*", core.Permission{Resource: core.ResourceTenant, Action: core.ActionWrite})
	if err != nil {
		return nil, err
	}
	if !hasGlobalTenantPermission && existingTenant.TenantKey == jwtUser.TenantKey {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "update_tenant"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("user_tenant_id", jwtUser.TenantKey),
			logx.Field("target_tenant_key", existingTenant.TenantKey),
		).Error("尝试更新自己的租户")
		return &types.UpdateTenantResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "不能修改自己所属的租户"},
		}, nil
	}

	// 检查跨租户操作权限
	if existingTenant.TenantKey != jwtUser.TenantKey {
		// 检查是否可以访问目标租户
		canAccessTenant, err := l.svcCtx.Casbinx.CanAccessTenant(jwtUser.UserKey, existingTenant.TenantKey)
		if err != nil {
			return &types.UpdateTenantResponse{
				BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
			}, nil
		}
		if !canAccessTenant {
			return &types.UpdateTenantResponse{
				BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：无法操作其他租户"},
			}, nil
		}
	}

	// 使用sqlx进行数据库操作

	// 构建动态更新查询
	var setClauses []string
	var args []interface{}
	argIndex := 1

	if req.TenantName != "" {
		setClauses = append(setClauses, fmt.Sprintf("tenant_name = $%d", argIndex))
		args = append(args, req.TenantName)
		argIndex++
	}

	if req.Description != "" {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, req.Description)
		argIndex++
	}

	if req.ContactEmail != "" {
		setClauses = append(setClauses, fmt.Sprintf("contact_email = $%d", argIndex))
		args = append(args, req.ContactEmail)
		argIndex++
	}

	if req.MaxUsers > 0 {
		setClauses = append(setClauses, fmt.Sprintf("max_users = $%d", argIndex))
		args = append(args, req.MaxUsers)
		argIndex++
	}

	if req.ExpiresAt != "" {
		setClauses = append(setClauses, fmt.Sprintf("expires_at = $%d", argIndex))
		args = append(args, req.ExpiresAt)
		argIndex++
	}

	// 使用事务处理整个租户更新过程（租户更新 + 标签更新）
	err = l.svcCtx.DBConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 如果有字段需要更新
		if len(setClauses) > 0 {
			setClauses = append(setClauses, fmt.Sprintf("updated_by = $%d", argIndex))
			args = append(args, jwtUser.UserId)
			argIndex++

			setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")

			args = append(args, req.Id)

			updateQuery := fmt.Sprintf(`
				UPDATE system_tenants
				SET %s
				WHERE id = $%d AND status != 'deleted'
			`, strings.Join(setClauses, ", "), argIndex)

			_, err := session.ExecCtx(ctx, updateQuery, args...)
			if err != nil {
				return fmt.Errorf("更新租户基本信息失败: %w", err)
			}
		}

		// 处理标签更新（如果提供了标签）
		if req.Tags != nil && len(req.Tags) >= 0 {
			// 先删除现有的标签关联
			deleteTagQuery := "DELETE FROM system_tenant_tag_relations WHERE tenant_id = $1"
			_, err := session.ExecCtx(ctx, deleteTagQuery, existingTenant.Id)
			if err != nil {
				return fmt.Errorf("删除现有标签关联失败: %w", err)
			}

			// 添加新的标签关联
			for _, tagLabel := range req.Tags {
				// 查找或创建标签
				var tagId string
				findTagQuery := "SELECT id FROM system_tag_definitions WHERE scope = $1 AND label = $2"
				err := session.QueryRowCtx(ctx, &tagId, findTagQuery, "tenant", tagLabel)
				if err != nil {
					// 标签不存在，创建新标签
					createTagQuery := "INSERT INTO system_tag_definitions (scope, label) VALUES ($1, $2) RETURNING id"
					err = session.QueryRowCtx(ctx, &tagId, createTagQuery, "tenant", tagLabel)
					if err != nil {
						return fmt.Errorf("创建标签 %s 失败: %w", tagLabel, err)
					}
				}

				// 创建租户-标签关联
				insertTagRelationQuery := "INSERT INTO system_tenant_tag_relations (tenant_id, tag_id) VALUES ($1, $2)"
				_, err = session.ExecCtx(ctx, insertTagRelationQuery, existingTenant.Id, tagId)
				if err != nil {
					return fmt.Errorf("创建租户标签关联 %s 失败: %w", tagLabel, err)
				}
			}
		}

		return nil
	})

	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "update_tenant"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("租户更新事务失败")
		return &types.UpdateTenantResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: fmt.Sprintf("更新租户信息失败: %v", err)},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tenant"),
		logx.Field("operation", "update_tenant"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", req.Id),
		logx.Field("updated_by", jwtUser.UserId),
	).Info("更新租户信息成功")

	return &types.UpdateTenantResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "更新租户信息成功",
		},
	}, nil
}
