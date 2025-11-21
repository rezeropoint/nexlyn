package user

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCurrentUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取当前用户信息
func NewGetCurrentUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrentUserLogic {
	return &GetCurrentUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCurrentUserLogic) GetCurrentUser() (resp *types.GetCurrentUserResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "get_current_user"),
		logx.Field("status", "started"),
	).Info("开始获取当前用户信息")

	// 从JWT中获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "get_current_user"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从JWT获取用户信息失败")

		return &types.GetCurrentUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
		}, nil
	}

	// 查询用户详细信息
	var user struct {
		TenantKey   string `db:"tenant_key"`
		Id          string `db:"id"`
		UserKey     string `db:"user_key"`
		UserName    string `db:"user_name"`
		Name        string `db:"name"`
		Email       string `db:"email"`
		Phone       string `db:"phone"`
		Avatar      string `db:"avatar"`
		Signature   string `db:"signature"`
		Title       string `db:"title"`
		Country     string `db:"country"`
		Address     string `db:"address"`
		Status      string `db:"status"`
		CreatedAt   string `db:"created_at"`
		UpdatedAt   string `db:"updated_at"`
		LastLoginAt string `db:"last_login_at"`
		// 租户信息
		TenantIdStr string `db:"tenant_id_str"`
		TenantName  string `db:"tenant_name"`
	}

	query := `
        SELECT 
            u.id, u.user_key, u.user_name, u.name, u.email, t.tenant_key,
            COALESCE(u.phone, '') as phone, 
            COALESCE(u.avatar, '') as avatar, 
            COALESCE(u.signature, '') as signature, 
            COALESCE(u.title, '') as title, 
            COALESCE(u.country, '') as country, 
            COALESCE(u.address, '') as address, 
            u.status,
            COALESCE(TO_CHAR(u.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at,
            COALESCE(TO_CHAR(u.updated_at, 'YYYY-MM-DD HH24:MI:SS'), '') as updated_at,
            COALESCE(TO_CHAR(u.last_login_at, 'YYYY-MM-DD HH24:MI:SS'), '') as last_login_at,
            COALESCE(u.tenant_id::text, '') as tenant_id_str,
            COALESCE(t.tenant_name, '') as tenant_name
        FROM system_users u
        LEFT JOIN system_tenants t ON u.tenant_id = t.id
        WHERE u.user_key = $1 AND u.status != 'deleted'
    `

	err = l.svcCtx.DBConn.QueryRow(&user, query, jwtUser.UserKey)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "get_current_user"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserId),
			logx.Field("error", err.Error()),
		).Error("查询用户信息失败")

		return &types.GetCurrentUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 404,
				Msg:  "用户不存在",
			},
		}, nil
	}

	// 查询用户标签（通过关系表）
	var tags []types.UserTag
	// 直接使用查询到的用户UUID作为内部ID
	userInternalID := user.Id
	if userInternalID != "" {
		tagQuery := `
            SELECT td.id AS id, td.label AS label
            FROM user_tag_relations utr
            JOIN system_tag_definitions td ON utr.tag_id = td.id
            WHERE utr.user_id = $1 AND td.scope = 'user'
        `
		tagErr := l.svcCtx.DBConn.QueryRowsPartial(&tags, tagQuery, userInternalID)
		_ = tagErr // 忽略错误，标签不是必需的
	}

	// 查询用户地理信息
	var geographic types.UserGeographic
	geoQuery := `
		SELECT COALESCE(province_key, '') as province_key,
		       COALESCE(province_label, '') as province_label,
		       COALESCE(city_key, '') as city_key,
		       COALESCE(city_label, '') as city_label
		FROM user_geographic
		WHERE user_id = $1
	`

	var geoData struct {
		ProvinceKey   string `db:"province_key"`
		ProvinceLabel string `db:"province_label"`
		CityKey       string `db:"city_key"`
		CityLabel     string `db:"city_label"`
	}
	err = l.svcCtx.DBConn.QueryRow(&geoData, geoQuery, user.Id)

	if err == nil {
		geographic = types.UserGeographic{
			Province: types.UserLocation{
				Key:   geoData.ProvinceKey,
				Label: geoData.ProvinceLabel,
			},
			City: types.UserLocation{
				Key:   geoData.CityKey,
				Label: geoData.CityLabel,
			},
		}
	}

	// 查询用户关联的组织及其所有下属组织ID列表
	// 使用路径编码模式：先获取用户直接关联的组织，再通过路径查询其下属组织
	// 主组织及其下属组织排在最前面
	var organizationIds []string

	// 获取用户的租户ID用于租户隔离
	var userTenantId string
	tenantQuery := `SELECT tenant_id::text FROM system_users WHERE id = $1`
	err = l.svcCtx.DBConn.QueryRow(&userTenantId, tenantQuery, user.Id)
	if err != nil || userTenantId == "" {
		// 如果获取租户ID失败，继续使用空的organizationIds
		logx.WithContext(l.ctx).Error("获取用户租户ID失败，组织列表将为空")
		userTenantId = ""
	}

	if userTenantId != "" {
		orgQuery := `
			WITH user_orgs AS (
				-- 获取用户直接关联的组织，包含是否为主组织的信息
				SELECT o.id, o.path, o.tenant_id, uor.is_primary
				FROM system_user_org_relations uor
				JOIN system_organizations o ON uor.org_id = o.id
				WHERE uor.user_id = $1
					AND uor.status = 'active'
					AND (uor.end_date IS NULL OR uor.end_date >= CURRENT_DATE)
					AND o.status = 'active'
					AND o.deleted_at IS NULL
					AND o.tenant_id = $2  -- 租户隔离
			),
			primary_org AS (
				-- 获取主组织的路径
				SELECT path FROM user_orgs WHERE is_primary = true LIMIT 1
			)
			-- 查询用户组织及其所有下属组织（通过路径匹配）
			SELECT DISTINCT o.id,
				-- 判断该组织是否属于主组织树（用于排序）
				CASE
					WHEN EXISTS (SELECT 1 FROM primary_org po WHERE o.path = po.path OR o.path LIKE po.path || '%')
					THEN 0  -- 主组织及其下属组织排在前面
					ELSE 1  -- 其他组织排在后面
				END as sort_order,
				o.level,
				o.sort_order as org_sort_order
			FROM system_organizations o
			CROSS JOIN user_orgs uo
			WHERE o.status = 'active'
				AND o.deleted_at IS NULL
				AND o.tenant_id = $2  -- 租户隔离：只查询同租户的组织
				AND (o.path = uo.path OR o.path LIKE uo.path || '%')
			ORDER BY sort_order, level, org_sort_order, id
		`

		var orgIds []struct {
			Id           string `db:"id"`
			SortOrder    int    `db:"sort_order"`
			Level        int    `db:"level"`
			OrgSortOrder int    `db:"org_sort_order"`
		}

		err = l.svcCtx.DBConn.QueryRowsPartial(&orgIds, orgQuery, user.Id, userTenantId)
		if err == nil {
			for _, org := range orgIds {
				organizationIds = append(organizationIds, org.Id)
			}
		}
	}

	// 构建响应数据
	currentUser := types.User{
		Id:        user.Id,
		UserKey:   user.UserKey,
		UserName:  user.UserName,
		Email:     user.Email,
		Phone:     user.Phone,
		Name:      user.Name,
		Avatar:    user.Avatar,
		Signature: user.Signature,
		Title:     user.Title,
		TenantInfo: types.TenantInfo{
			TenantKey:  user.TenantKey,
			TenantId:   user.TenantIdStr,
			TenantName: user.TenantName,
		},
		OrganizationIds: organizationIds,
		Tags:            tags,
		Geographic:      geographic,
		Country:         user.Country,
		Address:         user.Address,
		Status:          user.Status,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
		LastLoginAt:     user.LastLoginAt,
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "get_current_user"),
		logx.Field("status", "success"),
		logx.Field("user_key", jwtUser.UserKey),
	).Info("获取当前用户信息成功")

	return &types.GetCurrentUserResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "获取用户信息成功",
		},
		Data: currentUser,
	}, nil
}
