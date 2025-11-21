package organization

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrganizationListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取组织列表
func NewGetOrganizationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrganizationListLogic {
	return &GetOrganizationListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrganizationListLogic) GetOrganizationList(req *types.GetOrganizationListRequest) (resp *types.GetOrganizationListResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "get_organization_list"),
		logx.Field("status", "started"),
		logx.Field("current", req.Current),
		logx.Field("page_size", req.PageSize),
	).Info("开始获取组织列表")

	// 获取当前用户信息
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization_list"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从 JWT 获取用户信息失败")

		return &types.GetOrganizationListResponse{
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
			logx.Field("operation", "get_organization_list"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetOrganizationListResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}
	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization_list"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
		).Error("权限不足")

		return &types.GetOrganizationListResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足"},
		}, nil
	}

	// 处理分页参数
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	offset := (req.Current - 1) * req.PageSize

	// 构建基础查询条件
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	// 租户隔离
	whereConditions = append(whereConditions, "o.tenant_id = (SELECT id FROM system_tenants WHERE tenant_key = $"+strconv.Itoa(argIndex)+")")
	args = append(args, currentUser.TenantKey)
	argIndex++

	// 软删除过滤
	whereConditions = append(whereConditions, "o.deleted_at IS NULL")

	// 根据请求参数构建过滤条件
	if req.Code != "" {
		whereConditions = append(whereConditions, "o.code ILIKE $"+strconv.Itoa(argIndex))
		args = append(args, "%"+req.Code+"%")
		argIndex++
	}
	if req.Name != "" {
		whereConditions = append(whereConditions, "o.name ILIKE $"+strconv.Itoa(argIndex))
		args = append(args, "%"+req.Name+"%")
		argIndex++
	}
	if req.Type != "" {
		whereConditions = append(whereConditions, "o.type = $"+strconv.Itoa(argIndex))
		args = append(args, req.Type)
		argIndex++
	}
	if req.Status != "" {
		whereConditions = append(whereConditions, "o.status = $"+strconv.Itoa(argIndex))
		args = append(args, req.Status)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 查询总数
	var total int64
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM system_organizations o
		WHERE %s
	`, whereClause)

	err = l.svcCtx.DBConn.QueryRowCtx(l.ctx, &total, countQuery, args...)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization_list"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("查询组织总数失败")

		return &types.GetOrganizationListResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgOrganizationQuery, err)},
		}, nil
	}

	// 查询列表数据
	listQuery := fmt.Sprintf(`
		SELECT
			o.id, o.code, o.name, o.type, o.description,
			o.parent_id, o.path, o.level, o.sort_order, o.status,
			o.manager_id, u.name as manager_name,
			COALESCE(member_count.count, 0) as member_count,
			o.created_at, o.updated_at
		FROM system_organizations o
		LEFT JOIN system_users u ON o.manager_id = u.id
		LEFT JOIN (
			SELECT org_id, COUNT(*) as count
			FROM system_user_org_relations
			WHERE deleted_at IS NULL
			GROUP BY org_id
		) member_count ON o.id = member_count.org_id
		WHERE %s
		ORDER BY o.path, o.sort_order
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, req.PageSize, offset)

	var organizations []struct {
		Id          string         `db:"id"`
		Code        string         `db:"code"`
		Name        string         `db:"name"`
		Type        string         `db:"type"`
		Description sql.NullString `db:"description"`
		ParentId    sql.NullString `db:"parent_id"`
		Path        string         `db:"path"`
		Level       int            `db:"level"`
		SortOrder   int            `db:"sort_order"`
		Status      string         `db:"status"`
		ManagerId   sql.NullString `db:"manager_id"`
		ManagerName sql.NullString `db:"manager_name"`
		MemberCount int            `db:"member_count"`
		CreatedAt   sql.NullTime   `db:"created_at"`
		UpdatedAt   sql.NullTime   `db:"updated_at"`
	}

	err = l.svcCtx.DBConn.QueryRowsCtx(l.ctx, &organizations, listQuery, args...)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization_list"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("查询组织列表失败")

		return &types.GetOrganizationListResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgOrganizationQuery, err)},
		}, nil
	}

	// 转换为响应格式
	var orgList []types.OrganizationBrief
	for _, org := range organizations {
		orgBrief := types.OrganizationBrief{
			Id:          org.Id,
			Code:        org.Code,
			Name:        org.Name,
			Type:        org.Type,
			Path:        org.Path,
			Level:       org.Level,
			Status:      org.Status,
			MemberCount: org.MemberCount,
		}

		if org.ParentId.Valid {
			pid := org.ParentId.String
			orgBrief.ParentId = &pid
		}
		if org.ManagerName.Valid {
			mn := org.ManagerName.String
			orgBrief.ManagerName = &mn
		}

		orgList = append(orgList, orgBrief)
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "get_organization_list"),
		logx.Field("status", "success"),
		logx.Field("user_key", currentUser.UserKey),
		logx.Field("tenant_key", currentUser.TenantKey),
		logx.Field("total", total),
		logx.Field("returned", len(orgList)),
	).Info("获取组织列表成功")

	return &types.GetOrganizationListResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "success"},
		PageParams: types.PageParams{
			Current:  req.Current,
			PageSize: req.PageSize,
			Total:    total,
		},
		Data: types.OrganizationListData{
			List: orgList,
		},
	}, nil
}
