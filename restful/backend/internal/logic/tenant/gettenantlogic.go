package tenant

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取租户详情
func NewGetTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantLogic {
	return &GetTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantLogic) GetTenant(req *types.GetTenantRequest) (resp *types.GetTenantResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tenant"),
		logx.Field("operation", "get_tenant"),
		logx.Field("status", "started"),
		logx.Field("tenant_key", req.Id),
	).Info("开始获取租户详情")

	// 获取当前操作者（用于权限验证）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "get_tenant"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户失败")
		return &types.GetTenantResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 权限验证：检查是否有租户查看权限（使用Casbinx）
	hasReadPermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: core.ResourceTenant, Action: core.ActionRead})
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "get_tenant"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")
		return &types.GetTenantResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}

	if !hasReadPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "get_tenant"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")
		return &types.GetTenantResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要租户查看权限"},
		}, nil
	}

	// 查询租户基本信息
	var tenantData struct {
		Id           string `db:"id"`
		TenantKey    string `db:"tenant_key"`
		TenantName   string `db:"tenant_name"`
		Description  string `db:"description"`
		ContactEmail string `db:"contact_email"`
		Status       string `db:"status"`
		MaxUsers     int    `db:"max_users"`
		ExpiresAt    string `db:"expires_at"`
		CreatedAt    string `db:"created_at"`
		UpdatedAt    string `db:"updated_at"`
		CreatedBy    string `db:"created_by"`
		UpdatedBy    string `db:"updated_by"`
	}

	getTenantQuery := `
		SELECT
			id::text as id, tenant_key, tenant_name, COALESCE(description, '') as description, COALESCE(contact_email, '') as contact_email,
			status, COALESCE(max_users, 0) as max_users,
			COALESCE(expires_at::text, '') as expires_at,
			COALESCE(created_at::text, '') as created_at, COALESCE(updated_at::text, '') as updated_at,
			COALESCE(created_by, '') as created_by, COALESCE(updated_by, '') as updated_by
		FROM system_tenants
		WHERE id = $1 AND status != 'deleted'`

	err = l.svcCtx.DBConn.QueryRow(&tenantData, getTenantQuery, req.Id)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "get_tenant"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("查询租户信息失败")
		return &types.GetTenantResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantQuery, err)},
		}, nil
	}

	// 映射到返回结构
	var tenant types.Tenant
	tenant.Id = tenantData.Id
	tenant.TenantKey = tenantData.TenantKey
	tenant.TenantName = tenantData.TenantName
	tenant.Description = tenantData.Description
	tenant.ContactEmail = tenantData.ContactEmail
	tenant.MaxUsers = tenantData.MaxUsers
	tenant.CreatedAt = tenantData.CreatedAt
	tenant.UpdatedAt = tenantData.UpdatedAt
	tenant.ExpiresAt = tenantData.ExpiresAt
	tenant.CreatedBy = tenantData.CreatedBy
	tenant.UpdatedBy = tenantData.UpdatedBy

	// 直接使用字符串状态
	tenant.Status = tenantData.Status

	// 查询租户标签
	var tags []string
	tagQuery := `
		SELECT td.label
		FROM system_tenant_tag_relations ttr
		JOIN system_tag_definitions td ON ttr.tag_id = td.id
		WHERE ttr.tenant_id = $1
		AND td.scope = 'tenant'
	`
	err = l.svcCtx.DBConn.QueryRowsPartial(&tags, tagQuery, tenantData.Id)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "get_tenant"),
			logx.Field("status", "warning"),
			logx.Field("tenant_id", tenantData.Id),
			logx.Field("error", err.Error()),
		).Info("查询租户标签失败")
		// 不影响主要功能，标签查询失败不返回错误
	}

	tenant.Tags = tags

	// 租户隔离权限控制：检查是否可以访问目标租户
	if tenantData.TenantKey != jwtUser.TenantKey {
		canAccessTenant, err := l.svcCtx.Casbinx.CanAccessTenant(jwtUser.UserKey, tenantData.TenantKey)
		if err != nil {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "tenant"),
				logx.Field("operation", "get_tenant"),
				logx.Field("status", "failed"),
				logx.Field("user_key", jwtUser.UserKey),
				logx.Field("target_tenant_key", tenantData.TenantKey),
				logx.Field("error", err.Error()),
			).Error("检查租户访问权限失败")
			return &types.GetTenantResponse{
				BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
			}, nil
		}

		if !canAccessTenant {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "tenant"),
				logx.Field("operation", "get_tenant"),
				logx.Field("status", "failed"),
				logx.Field("user_key", jwtUser.UserKey),
				logx.Field("user_tenant_id", jwtUser.TenantKey),
				logx.Field("target_tenant_key", tenantData.TenantKey),
			).Error("拒绝跨租户访问")
			return &types.GetTenantResponse{
				BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：无法访问其他租户信息"},
			}, nil
		}
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tenant"),
		logx.Field("operation", "get_tenant"),
		logx.Field("status", "success"),
		logx.Field("tenant_key", tenantData.TenantKey),
		logx.Field("tenant_name", tenant.TenantName),
	).Info("获取租户详情成功")

	return &types.GetTenantResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "获取租户详情成功",
		},
		Data: tenant,
	}, nil
}
