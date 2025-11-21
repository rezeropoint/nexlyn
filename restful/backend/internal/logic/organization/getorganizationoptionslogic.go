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

type GetOrganizationOptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取组织下拉选项
func NewGetOrganizationOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrganizationOptionsLogic {
	return &GetOrganizationOptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrganizationOptionsLogic) GetOrganizationOptions(req *types.GetOrganizationOptionsRequest) (resp *types.GetOrganizationOptionsResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "get_organization_options"),
		logx.Field("status", "started"),
		logx.Field("tenant_id", req.TenantId),
		logx.Field("keyword", req.Keyword),
		logx.Field("parent_id", req.ParentId),
		logx.Field("org_type", req.Type),
		logx.Field("limit", req.Limit),
	).Info("开始获取组织下拉选项")

	// 获取当前用户信息
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization_options"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从 JWT 获取用户信息失败")

		return &types.GetOrganizationOptionsResponse{
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
			logx.Field("operation", "get_organization_options"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetOrganizationOptionsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}
	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization_options"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
		).Error("权限不足")

		return &types.GetOrganizationOptionsResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足"},
		}, nil
	}

	// 确定目标租户ID
	targetTenantId := currentUser.TenantId
	if req.TenantId != "" && req.TenantId != currentUser.TenantId {
		// TODO: 检查是否为超级管理员权限
		// 这里可以添加超级管理员权限检查
		targetTenantId = req.TenantId
	}

	// 构建查询条件
	whereClause := `WHERE tenant_id = $1 AND status = 'active' AND deleted_at IS NULL`
	args := []interface{}{targetTenantId}
	argIndex := 2

	if req.Keyword != "" {
		whereClause += fmt.Sprintf(` AND (name ILIKE $%d OR code ILIKE $%d)`, argIndex, argIndex)
		args = append(args, "%"+req.Keyword+"%")
		argIndex++
	}

	if req.ParentId != "" {
		whereClause += fmt.Sprintf(` AND parent_id = $%d`, argIndex)
		args = append(args, req.ParentId)
		argIndex++
	}

	if req.Type != "" {
		whereClause += fmt.Sprintf(` AND type = $%d`, argIndex)
		args = append(args, req.Type)
		argIndex++
	}

	// 设置默认限制数量
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// 查询组织选项
	var options []types.OrganizationOption
	optionsQuery := fmt.Sprintf(`
		SELECT id, code, name, path, level, status
		FROM system_organizations
		%s
		ORDER BY level ASC, sort_order ASC, name ASC
		LIMIT $%d
	`, whereClause, argIndex)
	args = append(args, limit)

	err = l.svcCtx.DBConn.QueryRowsCtx(l.ctx, &options, optionsQuery, args...)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "get_organization_options"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("error", err.Error()),
		).Error("查询组织选项失败")

		return &types.GetOrganizationOptionsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgOrganizationQuery, err)},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "get_organization_options"),
		logx.Field("status", "success"),
		logx.Field("user_key", currentUser.UserKey),
		logx.Field("tenant_key", currentUser.TenantKey),
		logx.Field("option_count", len(options)),
	).Info("获取组织下拉选项成功")

	return &types.GetOrganizationOptionsResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "查询成功"},
		Data: struct {
			List []types.OrganizationOption `json:"list"`
		}{
			List: options,
		},
	}, nil
}
