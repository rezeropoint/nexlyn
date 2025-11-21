package organization

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrganizationMembersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取组织成员
func NewGetOrganizationMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrganizationMembersLogic {
	return &GetOrganizationMembersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrganizationMembersLogic) GetOrganizationMembers(req *types.GetOrganizationMembersRequest) (resp *types.GetOrganizationMembersResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "get_organization_members"),
		logx.Field("status", "started"),
		logx.Field("org_id", req.Id),
		logx.Field("current", req.Current),
		logx.Field("page_size", req.PageSize),
		logx.Field("keyword", req.Keyword),
		logx.Field("relation_type", req.RelationType),
		logx.Field("is_primary", req.IsPrimary),
	).Info("开始获取组织成员列表")

	// 获取当前用户信息
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization_members"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从 JWT 获取用户信息失败")

		return &types.GetOrganizationMembersResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 权限验证：检查组织读取权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		currentUser.UserKey,
		currentUser.TenantKey,
		core.Permission{Resource: core.ResourceOrganization, Action: core.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization_members"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetOrganizationMembersResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}
	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization_members"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
		).Error("权限不足")

		return &types.GetOrganizationMembersResponse{
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
			logx.Field("operation", "get_organization_members"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("error", err.Error()),
		).Error("获取租户信息失败")

		return &types.GetOrganizationMembersResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantQuery, err)},
		}, nil
	}

	// 验证组织存在且属于当前租户
	var orgCount int
	checkOrgQuery := `SELECT COUNT(*) FROM system_organizations WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`
	err = l.svcCtx.DBConn.QueryRowCtx(l.ctx, &orgCount, checkOrgQuery, req.Id, tenantId)
	if err != nil || orgCount == 0 {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization_members"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("org_id", req.Id),
		).Error("组织不存在或不属于当前租户")

		return &types.GetOrganizationMembersResponse{
			BaseResponse: types.BaseResponse{Code: 404, Msg: config.ErrMsgOrganizationNotFound},
		}, nil
	}

	// 构建查询条件
	whereClause := `WHERE uor.org_id = $1 AND uor.deleted_at IS NULL AND u.tenant_id = $2 AND u.status = 'active'`
	args := []interface{}{req.Id, tenantId}
	argIndex := 3

	if req.Keyword != "" {
		whereClause += fmt.Sprintf(` AND (u.user_name ILIKE $%d OR u.name ILIKE $%d OR u.email ILIKE $%d)`, argIndex, argIndex, argIndex)
		args = append(args, "%"+req.Keyword+"%")
		argIndex++
	}

	if req.RelationType != "" {
		whereClause += fmt.Sprintf(` AND uor.relation_type = $%d`, argIndex)
		args = append(args, req.RelationType)
		argIndex++
	}

	if req.IsPrimary {
		whereClause += fmt.Sprintf(` AND uor.is_primary = $%d`, argIndex)
		args = append(args, true)
		argIndex++
	}

	// 计算分页参数
	current := req.Current
	if current <= 0 {
		current = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (current - 1) * pageSize

	// 查询总数
	var total int64
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM system_user_org_relations uor
		JOIN system_users u ON uor.user_id = u.id
		%s
	`, whereClause)

	err = l.svcCtx.DBConn.QueryRowCtx(l.ctx, &total, countQuery, args...)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization_members"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("org_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("查询组织成员总数失败")

		return &types.GetOrganizationMembersResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgOrganizationQuery, err)},
		}, nil
	}

	// 查询成员列表
	var memberData []struct {
		Id            string `db:"id"`
		UserId        string `db:"user_id"`
		UserKey       string `db:"user_key"`
		UserName      string `db:"user_name"`
		Name          string `db:"name"`
		Email         string `db:"email"`
		OrgId         string `db:"org_id"`
		OrgCode       string `db:"org_code"`
		OrgName       string `db:"org_name"`
		RelationType  string `db:"relation_type"`
		PositionTitle string `db:"position_title"`
		IsPrimary     bool   `db:"is_primary"`
		Status        string `db:"status"`
		JoinedAt      string `db:"joined_at"`
	}
	membersQuery := fmt.Sprintf(`
		SELECT uor.id, uor.user_id, u.user_key, u.user_name, u.name, u.email,
		       uor.org_id, '' as org_code, '' as org_name, uor.relation_type, 
		       uor.position_title, uor.is_primary, uor.status, uor.created_at as joined_at
		FROM system_user_org_relations uor
		JOIN system_users u ON uor.user_id = u.id
		%s
		ORDER BY uor.is_primary DESC, uor.created_at ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	err = l.svcCtx.DBConn.QueryRowsCtx(l.ctx, &memberData, membersQuery, args...)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization_members"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("org_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("查询组织成员列表失败")

		return &types.GetOrganizationMembersResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgOrganizationQuery, err)},
		}, nil
	}

	// 转换为响应数据格式
	var members []types.UserOrgRelation
	for _, member := range memberData {
		members = append(members, types.UserOrgRelation{
			Id:            member.Id,
			UserId:        member.UserId,
			UserKey:       member.UserKey,
			UserName:      member.UserName,
			Name:          member.Name,
			Email:         member.Email,
			OrgId:         member.OrgId,
			OrgCode:       member.OrgCode,
			OrgName:       member.OrgName,
			RelationType:  member.RelationType,
			PositionTitle: member.PositionTitle,
			IsPrimary:     member.IsPrimary,
			Status:        member.Status,
			JoinedAt:      member.JoinedAt,
		})
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "get_organization_members"),
		logx.Field("status", "success"),
		logx.Field("user_key", currentUser.UserKey),
		logx.Field("tenant_key", currentUser.TenantKey),
		logx.Field("org_id", req.Id),
		logx.Field("total", total),
		logx.Field("member_count", len(members)),
	).Info("获取组织成员列表成功")

	return &types.GetOrganizationMembersResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "查询成功"},
		PageParams: types.PageParams{
			Current:  current,
			PageSize: pageSize,
			Total:    total,
		},
		Data: types.OrganizationMembersData{
			List: members,
		},
	}, nil
}
