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

type UpdateOrganizationMemberLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新成员关系
func NewUpdateOrganizationMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOrganizationMemberLogic {
	return &UpdateOrganizationMemberLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateOrganizationMemberLogic) UpdateOrganizationMember(req *types.UpdateOrganizationMemberRequest) (resp *types.UpdateOrganizationMemberResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "update_organization_member"),
		logx.Field("status", "started"),
		logx.Field("org_id", req.OrgId),
		logx.Field("user_id", req.UserId),
		logx.Field("relation_type", req.RelationType),
		logx.Field("position_title", req.PositionTitle),
		logx.Field("is_primary", req.IsPrimary),
		logx.Field("status", req.Status),
	).Info("开始更新组织成员信息")

	// 获取当前用户信息
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "update_organization_member"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从 JWT 获取用户信息失败")

		return &types.UpdateOrganizationMemberResponse{
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
			logx.Field("operation", "update_organization_member"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.UpdateOrganizationMemberResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}
	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "update_organization_member"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
		).Error("权限不足")

		return &types.UpdateOrganizationMemberResponse{
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
			logx.Field("operation", "update_organization_member"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("error", err.Error()),
		).Error("获取租户信息失败")

		return &types.UpdateOrganizationMemberResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantQuery, err)},
		}, nil
	}

	// 使用事务处理成员更新
	err = l.svcCtx.DBConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 验证组织存在且属于当前租户
		var orgCount int
		checkOrgQuery := `SELECT COUNT(*) FROM system_organizations WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`
		err := session.QueryRowCtx(ctx, &orgCount, checkOrgQuery, req.OrgId, tenantId)
		if err != nil || orgCount == 0 {
			return fmt.Errorf(config.ErrMsgOrganizationNotFound)
		}

		// 验证用户存在且属于当前租户
		var userCount int
		checkUserQuery := `SELECT COUNT(*) FROM system_users WHERE id = $1 AND tenant_id = $2 AND status = 'active'`
		err = session.QueryRowCtx(ctx, &userCount, checkUserQuery, req.UserId, tenantId)
		if err != nil || userCount == 0 {
			return fmt.Errorf(config.ErrMsgUserNotFound)
		}

		// 验证成员关系存在
		var relationCount int
		checkRelationQuery := `SELECT COUNT(*) FROM system_user_org_relations WHERE user_id = $1 AND org_id = $2 AND deleted_at IS NULL`
		err = session.QueryRowCtx(ctx, &relationCount, checkRelationQuery, req.UserId, req.OrgId)
		if err != nil || relationCount == 0 {
			return fmt.Errorf("用户组织关系不存在")
		}

		// 构建更新字段
		updateFields := []string{}
		args := []interface{}{}
		argIndex := 1

		if req.RelationType != "" {
			updateFields = append(updateFields, fmt.Sprintf("relation_type = $%d", argIndex))
			args = append(args, req.RelationType)
			argIndex++
		}

		if req.PositionTitle != "" {
			updateFields = append(updateFields, fmt.Sprintf("position_title = $%d", argIndex))
			args = append(args, req.PositionTitle)
			argIndex++
		}

		if req.Status != "" {
			updateFields = append(updateFields, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, req.Status)
			argIndex++
		}

		// 如果设置为主要关系，需先更新其他关系为非主要
		if req.IsPrimary {
			_, err = session.ExecCtx(ctx, `
				UPDATE system_user_org_relations 
				SET is_primary = false 
				WHERE user_id = $1 AND deleted_at IS NULL
			`, req.UserId)
			if err != nil {
				return fmt.Errorf(config.ErrMsgOrganizationMemberUpdate)
			}

			updateFields = append(updateFields, fmt.Sprintf("is_primary = $%d", argIndex))
			args = append(args, true)
			argIndex++
		}

		// 添加更新时间和更新者
		updateFields = append(updateFields, fmt.Sprintf("updated_at = NOW(), updated_by = $%d", argIndex))
		args = append(args, currentUser.UserId)
		argIndex++

		// 如果有字段需要更新
		if len(updateFields) > 1 { // > 1 因为至少有 updated_at
			// 添加WHERE条件参数
			args = append(args, req.UserId, req.OrgId)

			updateQuery := fmt.Sprintf(`
				UPDATE system_user_org_relations 
				SET %s
				WHERE user_id = $%d AND org_id = $%d AND deleted_at IS NULL
			`, strings.Join(updateFields, ", "), argIndex, argIndex+1)

			_, err = session.ExecCtx(ctx, updateQuery, args...)
			if err != nil {
				return fmt.Errorf(config.ErrMsgOrganizationMemberUpdate)
			}
		}

		return nil
	})

	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "update_organization_member"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("org_id", req.OrgId),
			logx.Field("user_id", req.UserId),
			logx.Field("error", err.Error()),
		).Error("更新组织成员信息失败")

		return &types.UpdateOrganizationMemberResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: err.Error()},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "update_organization_member"),
		logx.Field("status", "success"),
		logx.Field("user_key", currentUser.UserKey),
		logx.Field("tenant_key", currentUser.TenantKey),
		logx.Field("org_id", req.OrgId),
		logx.Field("user_id", req.UserId),
	).Info("更新组织成员信息成功")

	return &types.UpdateOrganizationMemberResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "更新成功"},
	}, nil
}
