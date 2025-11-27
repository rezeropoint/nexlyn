package user

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
)

type GetUserOptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取用户选项列表（下拉框用）
func NewGetUserOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserOptionsLogic {
	return &GetUserOptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserOptionsLogic) GetUserOptions(req *types.GetUserOptionsRequest) (resp *types.GetUserOptionsResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "get_user_options"),
		logx.Field("status", "started"),
		logx.Field("tenant_id", req.TenantId),
		logx.Field("keyword", req.Keyword),
		logx.Field("limit", req.Limit),
	).Info("开始获取用户选项列表")

	// 获取当前操作者（用于权限验证和租户隔离）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "get_user_options"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户失败")
		return &types.GetUserOptionsResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 权限验证：检查是否有用户查看权限
	hasReadPermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: core.ResourceUser, Action: core.ActionRead})
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "get_user_options"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")
		return &types.GetUserOptionsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}

	if !hasReadPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "get_user_options"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")
		return &types.GetUserOptionsResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: config.FormatError(config.ErrMsgPermissionDenied, err)},
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

	// 构建查询条件（基于 users(u) 与 tenants(t)）
	var whereClause []string
	var args []interface{}
	argIndex := 1

	// 固定条件：只查询未删除的用户
	whereClause = append(whereClause, "u.status != 'deleted'")

	// 根据搜索关键字添加WHERE子句
	if req.Keyword != "" {
		whereClause = append(whereClause, fmt.Sprintf("(u.user_name ILIKE $%d OR u.name ILIKE $%d OR u.email ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+req.Keyword+"%")
		argIndex++
	}

	// 多租户过滤：
	// - 有跨租户权限：可查看所有租户；若指定 tenantId，则按指定租户过滤
	// - 无跨租户权限：强制限定为其自身租户
	hasCrossTenantPermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: core.ResourceTenant, Action: core.ActionRead})
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "get_user_options"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("跨租户权限检查失败")
		return &types.GetUserOptionsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}

	if hasCrossTenantPermission {
		// 有跨租户权限，可以查看指定租户或所有租户
		if req.TenantId != "" {
			whereClause = append(whereClause, fmt.Sprintf("t.tenant_key = $%d", argIndex))
			args = append(args, req.TenantId)
			argIndex++
		}
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "get_user_options"),
			logx.Field("status", "info"),
			logx.Field("user_key", jwtUser.UserKey),
		).Info("跨租户权限：允许查看指定或所有租户用户")
	} else {
		// 无跨租户权限，只能查看自己租户（使用TenantKey匹配tenant_key字段）
		whereClause = append(whereClause, fmt.Sprintf("t.tenant_key = $%d", argIndex))
		args = append(args, jwtUser.TenantKey)
		argIndex++

		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "get_user_options"),
			logx.Field("status", "info"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("tenant_id", jwtUser.TenantId),
		).Info("普通用户：应用租户隔离过滤")
	}

	// 角色级别过滤：根据当前用户的角色级别过滤可见的用户
	// - super_admin 可以看到所有用户
	// - admin 不能看到 super_admin 用户
	// - user 不能看到 admin 和 super_admin 用户
	// 权限过滤现在通过Casbin处理，移除基于role字段的过滤

	whereSQL := ""
	if len(whereClause) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClause, " AND ")
	}

	// 查询用户选项列表
	query := fmt.Sprintf(`
        SELECT u.id::text, u.user_key, u.user_name, u.name, u.email,
               COALESCE(t.id::text, '') as tenant_id,
               COALESCE(t.tenant_key, '') as tenant_key
        FROM system_users u
        LEFT JOIN system_tenants t ON u.tenant_id = t.id
        %s
        ORDER BY u.name ASC
        LIMIT $%d
	`, whereSQL, argIndex)

	args = append(args, limit)

	var userOptions []types.UserOption
	err = l.svcCtx.DBConn.QueryRowsPartial(&userOptions, query, args...)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "get_user_options"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询用户选项列表失败")
		return &types.GetUserOptionsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgUserOptions, err)},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "get_user_options"),
		logx.Field("status", "success"),
		logx.Field("count", len(userOptions)),
	).Info("获取用户选项列表成功")

	return &types.GetUserOptionsResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "获取用户选项列表成功",
		},
		Data: struct {
			List []types.UserOption `json:"list"`
		}{
			List: userOptions,
		},
	}, nil
}
