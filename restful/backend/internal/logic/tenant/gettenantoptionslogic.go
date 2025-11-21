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

	"github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantOptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取租户选项列表（下拉框用）
func NewGetTenantOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantOptionsLogic {
	return &GetTenantOptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantOptionsLogic) GetTenantOptions(req *types.GetTenantOptionsRequest) (resp *types.GetTenantOptionsResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tenant"),
		logx.Field("operation", "get_tenant_options"),
		logx.Field("status", "started"),
		logx.Field("keyword", req.Keyword),
		logx.Field("limit", req.Limit),
	).Info("开始获取租户选项列表")

	// 获取当前操作者（用于权限验证）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "get_tenant_options"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户失败")
		return &types.GetTenantOptionsResponse{
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
			logx.Field("operation", "get_tenant_options"),
			logx.Field("status", "failed"),
			logx.Field("user_id", jwtUser.UserId),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")
		return &types.GetTenantOptionsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}

	if !hasReadPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "get_tenant_options"),
			logx.Field("status", "failed"),
			logx.Field("user_id", jwtUser.UserId),
		).Error("用户权限不足")
		return &types.GetTenantOptionsResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要租户查看权限"},
		}, nil
	}

	// 设置默认限制数量
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // 限制最大数量
	}

	// 构建查询条件
	var whereClause []string
	var args []interface{}
	argIndex := 1

	// 固定条件：只查询未删除的租户
	whereClause = append(whereClause, "status != 'deleted'")

	// 根据搜索关键字添加WHERE子句
	if req.Keyword != "" {
		whereClause = append(whereClause, fmt.Sprintf("tenant_name ILIKE $%d", argIndex))
		args = append(args, "%"+req.Keyword+"%")
		argIndex++
	}

	// 状态过滤
	if req.Status != "" {
		// 验证状态值的有效性
		validStatuses := map[string]bool{
			"active":   true,
			"inactive": true,
			"deleted":  true,
		}

		if !validStatuses[req.Status] {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "tenant"),
				logx.Field("operation", "get_tenant_options"),
				logx.Field("status", "warning"),
				logx.Field("status_filter", req.Status),
			).Info("无效的状态过滤值")
		} else {
			whereClause = append(whereClause, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, req.Status)
			argIndex++
		}
	}

	// 租户隔离权限控制：获取用户可访问的租户列表
	userTenants, err := l.svcCtx.Casbinx.GetUserTenants(jwtUser.UserKey)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "get_tenant_options"),
			logx.Field("status", "failed"),
			logx.Field("user_id", jwtUser.UserId),
			logx.Field("error", err.Error()),
		).Error("获取用户可访问租户列表失败")
		return &types.GetTenantOptionsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantOptions, err)},
		}, nil
	}

	// 如果有租户限制，则添加过滤条件
	if len(userTenants) > 0 {
		whereClause = append(whereClause, fmt.Sprintf("tenant_key = ANY($%d::text[])", argIndex))
		args = append(args, pq.Array(userTenants))
		argIndex++

		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "get_tenant_options"),
			logx.Field("status", "info"),
			logx.Field("user_id", jwtUser.UserId),
			logx.Field("accessible_tenants", len(userTenants)),
		).Info("应用租户隔离过滤")
	}

	whereSQL := ""
	if len(whereClause) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClause, " AND ")
	}

	// 查询租户选项列表
	query := fmt.Sprintf(`
		SELECT 
			tenant_key,
			id::text as tenant_id,
			tenant_name,
			COALESCE(status, 'active') as status
		FROM system_tenants
		%s
		ORDER BY tenant_name ASC
		LIMIT $%d
	`, whereSQL, argIndex)

	args = append(args, limit)

	var tenantRows []struct {
		TenantKey  string `db:"tenant_key"`
		TenantId   string `db:"tenant_id"`
		TenantName string `db:"tenant_name"`
		Status     string `db:"status"`
	}
	err = l.svcCtx.DBConn.QueryRowsPartial(&tenantRows, query, args...)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "get_tenant_options"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询租户选项列表失败")
		return &types.GetTenantOptionsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantOptions, err)},
		}, nil
	}

	// 转换为 TenantOption 类型
	var tenantOptions []types.TenantOption
	for _, row := range tenantRows {
		tenantOptions = append(tenantOptions, types.TenantOption{
			TenantKey:  row.TenantKey,
			TenantId:   row.TenantId,
			TenantName: row.TenantName,
			Status:     row.Status,
		})
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tenant"),
		logx.Field("operation", "get_tenant_options"),
		logx.Field("status", "success"),
		logx.Field("count", len(tenantOptions)),
	).Info("获取租户选项列表成功")

	return &types.GetTenantOptionsResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "获取租户选项列表成功",
		},
		Data: struct {
			List []types.TenantOption `json:"list"`
		}{
			List: tenantOptions,
		},
	}, nil
}
