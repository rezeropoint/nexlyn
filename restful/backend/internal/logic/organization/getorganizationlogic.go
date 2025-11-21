package organization

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrganizationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取组织详情
func NewGetOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrganizationLogic {
	return &GetOrganizationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrganizationLogic) GetOrganization(req *types.GetOrganizationRequest) (resp *types.GetOrganizationResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "get_organization"),
		logx.Field("status", "started"),
		logx.Field("org_id", req.Id),
	).Info("开始获取组织详情")

	// 获取当前用户信息
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从 JWT 获取用户信息失败")

		return &types.GetOrganizationResponse{
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
			logx.Field("operation", "get_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}
	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
		).Error("权限不足")

		return &types.GetOrganizationResponse{
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
			logx.Field("operation", "get_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("error", err.Error()),
		).Error("获取租户信息失败")

		return &types.GetOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantQuery, err)},
		}, nil
	}

	// 查询组织信息
	var orgData struct {
		Id          string  `db:"id"`
		Code        string  `db:"code"`
		Name        string  `db:"name"`
		Type        string  `db:"type"`
		Description *string `db:"description"`
		Path        string  `db:"path"`
		ParentId    *string `db:"parent_id"`
		Level       int     `db:"level"`
		SortOrder   int     `db:"sort_order"`
		ManagerId   *string `db:"manager_id"`
		ManagerName *string `db:"manager_name"`
		MemberCount int     `db:"member_count"`
		Status      string  `db:"status"`
		TenantId    string  `db:"tenant_id"`
		CreatedBy   *string `db:"created_by"`
		CreatedAt   string  `db:"created_at"`
		UpdatedBy   *string `db:"updated_by"`
		UpdatedAt   string  `db:"updated_at"`
	}
	orgQuery := `
		SELECT id, code, name, type, description, path, parent_id,
		       level, sort_order, manager_id,
		       (SELECT name FROM system_users WHERE id = system_organizations.manager_id) as manager_name,
		       (SELECT COUNT(*) FROM system_user_org_relations WHERE org_id = system_organizations.id AND deleted_at IS NULL) as member_count,
		       status, tenant_id, created_by, created_at, updated_by, updated_at
		FROM system_organizations
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`

	err = l.svcCtx.DBConn.QueryRowCtx(l.ctx, &orgData, orgQuery, req.Id, tenantId)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("org_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("查询组织详情失败")

		return &types.GetOrganizationResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgOrganizationQuery, err)},
		}, nil
	}

	// 转换为响应数据格式
	org := types.Organization{
		Id:          orgData.Id,
		Code:        orgData.Code,
		Name:        orgData.Name,
		Type:        orgData.Type,
		Description: orgData.Description,
		Path:        orgData.Path,
		ParentId:    orgData.ParentId,
		Level:       orgData.Level,
		SortOrder:   orgData.SortOrder,
		ManagerId:   orgData.ManagerId,
		ManagerName: orgData.ManagerName,
		MemberCount: orgData.MemberCount,
		Status:      orgData.Status,
		TenantId:    orgData.TenantId,
		CreatedBy:   orgData.CreatedBy,
		UpdatedBy:   orgData.UpdatedBy,
		CreatedAt:   orgData.CreatedAt,
		UpdatedAt:   orgData.UpdatedAt,
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "get_organization"),
		logx.Field("status", "success"),
		logx.Field("user_key", currentUser.UserKey),
		logx.Field("tenant_key", currentUser.TenantKey),
		logx.Field("org_id", req.Id),
		logx.Field("org_code", org.Code),
		logx.Field("org_name", org.Name),
	).Info("获取组织详情成功")

	return &types.GetOrganizationResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "查询成功"},
		Data:         org,
	}, nil
}
