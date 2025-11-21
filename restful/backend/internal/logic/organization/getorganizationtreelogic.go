package organization

import (
	"context"
	"database/sql"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrganizationTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取组织树
func NewGetOrganizationTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrganizationTreeLogic {
	return &GetOrganizationTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrganizationTreeLogic) GetOrganizationTree(req *types.GetOrganizationTreeRequest) (resp *types.GetOrganizationTreeResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "get_organization_tree"),
		logx.Field("status", "started"),
		logx.Field("root_id", req.RootId),
	).Info("开始获取组织树")

	// 获取当前用户信息
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization_tree"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从JWT获取用户信息失败")

		return &types.GetOrganizationTreeResponse{
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
			logx.Field("operation", "get_organization_tree"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetOrganizationTreeResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}
	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization_tree"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
		).Error("权限不足")

		return &types.GetOrganizationTreeResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足"},
		}, nil
	}

	// 根据根节点ID构建查询条件
	var baseQuery string
	var args []interface{}

	if req.RootId == "" {
		// 查询该租户下的所有组织（根节点和所有子节点）
		baseQuery = `
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
			WHERE o.tenant_id = (SELECT id FROM system_tenants WHERE tenant_key = $1)
			  AND o.deleted_at IS NULL
			ORDER BY o.path, o.sort_order
		`
		args = []interface{}{currentUser.TenantKey}
	} else {
		// 查询指定根节点及其所有子节点
		baseQuery = `
			WITH root_org AS (
				SELECT path FROM system_organizations
				WHERE id = $2 AND tenant_id = (SELECT id FROM system_tenants WHERE tenant_key = $1)
				  AND deleted_at IS NULL
			)
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
			CROSS JOIN root_org r
			WHERE o.tenant_id = (SELECT id FROM system_tenants WHERE tenant_key = $1)
			  AND o.deleted_at IS NULL
			  AND (o.path = r.path OR o.path LIKE r.path || '%')
			ORDER BY o.path, o.sort_order
		`
		args = []interface{}{currentUser.TenantKey, req.RootId}
	}

	// 执行查询
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
		CreatedAt   string         `db:"created_at"`
		UpdatedAt   string         `db:"updated_at"`
	}

	err = l.svcCtx.DBConn.QueryRowsCtx(l.ctx, &organizations, baseQuery, args...)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization_tree"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("error", err.Error()),
		).Error("查询组织数据失败")

		return &types.GetOrganizationTreeResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgOrganizationQuery, err)},
		}, nil
	}

	// 构建树形结构
	orgMap := make(map[string]*types.OrganizationTree)

	// 先创建所有节点
	for _, org := range organizations {
		var descPtr *string
		if org.Description.Valid {
			v := org.Description.String
			descPtr = &v
		}
		var parentPtr *string
		if org.ParentId.Valid {
			v := org.ParentId.String
			parentPtr = &v
		}
		var managerIdPtr *string
		if org.ManagerId.Valid {
			v := org.ManagerId.String
			managerIdPtr = &v
		}
		var managerNamePtr *string
		if org.ManagerName.Valid {
			v := org.ManagerName.String
			managerNamePtr = &v
		}

		node := &types.OrganizationTree{
			Organization: types.Organization{
				Id:          org.Id,
				Code:        org.Code,
				Name:        org.Name,
				Type:        org.Type,
				Description: descPtr,
				Path:        org.Path,
				ParentId:    parentPtr,
				Level:       org.Level,
				SortOrder:   org.SortOrder,
				Status:      org.Status,
				ManagerId:   managerIdPtr,
				ManagerName: managerNamePtr,
				MemberCount: org.MemberCount,
				CreatedAt:   org.CreatedAt,
				UpdatedAt:   org.UpdatedAt,
			},
			Children: []types.OrganizationTree{},
		}

		orgMap[org.Id] = node
	}

	// 构建父子关系 - 使用递归方式正确构建多层树结构
	var buildChildren func(nodeId string) []types.OrganizationTree
	buildChildren = func(nodeId string) []types.OrganizationTree {
		var children []types.OrganizationTree
		for _, org := range organizations {
			if org.ParentId.Valid && org.ParentId.String == nodeId {
				childNode := orgMap[org.Id]
				// 递归构建子节点的子节点
				childNode.Children = buildChildren(org.Id)
				children = append(children, *childNode)
			}
		}
		return children
	}

	// 找到根节点并构建完整树结构
	var result []types.OrganizationTree
	for _, org := range organizations {
		if !org.ParentId.Valid || org.ParentId.String == "" {
			// 根节点
			rootNode := orgMap[org.Id]
			rootNode.Children = buildChildren(org.Id)
			result = append(result, *rootNode)
		}
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "get_organization_tree"),
		logx.Field("status", "success"),
		logx.Field("user_key", currentUser.UserKey),
		logx.Field("tenant_key", currentUser.TenantKey),
		logx.Field("root_id", req.RootId),
		logx.Field("result_count", len(result)),
	).Info("获取组织树成功")

	return &types.GetOrganizationTreeResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "success"},
		Data:         result,
	}, nil
}
