package organization

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserOrganizationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取用户所属组织
func NewGetUserOrganizationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserOrganizationsLogic {
	return &GetUserOrganizationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserOrganizationsLogic) GetUserOrganizations(req *types.GetUserOrganizationsRequest) (resp *types.GetUserOrganizationsResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "get_user_organizations"),
		logx.Field("status", "started"),
		logx.Field("user_id", req.Id),
		logx.Field("is_primary", req.IsPrimary),
	).Info("开始获取用户所属组织")

	// 获取当前用户信息
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_user_organizations"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从 JWT 获取用户信息失败")

		return &types.GetUserOrganizationsResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 权限验证：检查用户读取权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		currentUser.UserKey,
		currentUser.TenantKey,
		core.Permission{Resource: core.ResourceUser, Action: core.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_user_organizations"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetUserOrganizationsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}
	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_user_organizations"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
		).Error("权限不足")

		return &types.GetUserOrganizationsResponse{
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
			logx.Field("operation", "get_user_organizations"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("error", err.Error()),
		).Error("获取租户信息失败")

		return &types.GetUserOrganizationsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantQuery, err)},
		}, nil
	}

	// 验证用户存在且属于当前租户
	var userCount int
	checkUserQuery := `SELECT COUNT(*) FROM system_users WHERE id = $1 AND tenant_id = $2 AND status = 'active'`
	err = l.svcCtx.DBConn.QueryRowCtx(l.ctx, &userCount, checkUserQuery, req.Id, tenantId)
	if err != nil || userCount == 0 {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_user_organizations"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("target_user_id", req.Id),
			logx.Field("error", err),
		).Error("用户不存在或不属于当前租户")

		return &types.GetUserOrganizationsResponse{
			BaseResponse: types.BaseResponse{Code: 404, Msg: config.ErrMsgUserNotFound},
		}, nil
	}

	// 构建查询条件
	whereClause := `WHERE uor.user_id = $1 AND uor.deleted_at IS NULL AND o.tenant_id = $2 AND o.deleted_at IS NULL`
	args := []interface{}{req.Id, tenantId}
	argIndex := 3

	if req.IsPrimary {
		whereClause += ` AND uor.is_primary = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, true)
		argIndex++
	}

	// 查询用户组织关系 - 使用临时结构体处理NULL值
	var tempRelations []struct {
		Id            string         `db:"id"`
		UserId        string         `db:"user_id"`
		UserKey       string         `db:"user_key"`
		UserName      string         `db:"user_name"`
		Name          string         `db:"name"`
		Email         string         `db:"email"`
		OrgId         string         `db:"org_id"`
		OrgCode       string         `db:"org_code"`
		OrgName       string         `db:"org_name"`
		RelationType  string         `db:"relation_type"`
		PositionTitle sql.NullString `db:"position_title"`
		IsPrimary     bool           `db:"is_primary"`
		Status        string         `db:"status"`
		JoinedAt      string         `db:"joined_at"`
	}
	relationsQuery := fmt.Sprintf(`
		SELECT uor.id, uor.user_id, u.user_key, u.user_name, u.name, u.email,
		       uor.org_id, o.code as org_code, o.name as org_name, uor.relation_type,
		       uor.position_title, uor.is_primary, uor.status, uor.created_at as joined_at
		FROM system_user_org_relations uor
		JOIN system_users u ON uor.user_id = u.id
		JOIN system_organizations o ON uor.org_id = o.id
		%s
		ORDER BY uor.is_primary DESC, o.level ASC, o.sort_order ASC
	`, whereClause)

	err = l.svcCtx.DBConn.QueryRowsCtx(l.ctx, &tempRelations, relationsQuery, args...)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_user_organizations"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("target_user_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("查询用户组织关系失败")

		return &types.GetUserOrganizationsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgOrganizationQuery, err)},
		}, nil
	}

	// 转换临时结构体为API类型
	relations := make([]types.UserOrgRelation, len(tempRelations))
	for i, temp := range tempRelations {
		relations[i] = types.UserOrgRelation{
			Id:           temp.Id,
			UserId:       temp.UserId,
			UserKey:      temp.UserKey,
			UserName:     temp.UserName,
			Name:         temp.Name,
			Email:        temp.Email,
			OrgId:        temp.OrgId,
			OrgCode:      temp.OrgCode,
			OrgName:      temp.OrgName,
			RelationType: temp.RelationType,
			PositionTitle: func() string {
				if temp.PositionTitle.Valid {
					return temp.PositionTitle.String
				}
				return ""
			}(),
			IsPrimary: temp.IsPrimary,
			Status:    temp.Status,
			JoinedAt:  temp.JoinedAt,
		}
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "get_user_organizations"),
		logx.Field("status", "success"),
		logx.Field("user_key", currentUser.UserKey),
		logx.Field("tenant_key", currentUser.TenantKey),
		logx.Field("target_user_id", req.Id),
		logx.Field("relation_count", len(relations)),
	).Info("获取用户所属组织成功")

	return &types.GetUserOrganizationsResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "查询成功"},
		Data: struct {
			List []types.UserOrgRelation `json:"list"`
		}{
			List: relations,
		},
	}, nil
}
