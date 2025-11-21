package tags

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTagListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取标签列表
func NewGetTagListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTagListLogic {
	return &GetTagListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTagListLogic) GetTagList(req *types.GetTagListRequest) (resp *types.GetTagListResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tags"),
		logx.Field("operation", "get_tag_list"),
		logx.Field("status", "started"),
	).Info("开始获取标签列表")

	// 验证scope值
	if req.Scope != "tenant" && req.Scope != "user" {
		return &types.GetTagListResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "scope参数必须为tenant或user"},
		}, nil
	}

	// 获取当前操作者（用于权限验证）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "get_tag_list"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户失败")
		return &types.GetTagListResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 权限验证：根据scope类型检查读取权限
	var permissionResource core.Resource
	if req.Scope == "tenant" {
		permissionResource = core.ResourceTagTenant
	} else {
		permissionResource = core.ResourceTagUser
	}

	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: permissionResource, Action: core.ActionRead})
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "get_tag_list"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")
		return &types.GetTagListResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "get_tag_list"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")
		return &types.GetTagListResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: fmt.Sprintf("权限不足，需要%s读取权限", permissionResource)},
		}, nil
	}

	// 构建查询过滤器
	filters := make(map[string]interface{})

	// scope条件（必需）
	filters["scope"] = req.Scope

	// 租户隔离条件
	if req.Scope == "user" {
		// 用户标签：只查询当前租户的标签
		filters["tenant_id"] = jwtUser.TenantId
	}
	// 租户标签：tenant_id为NULL，在查询函数中处理

	// 关键词搜索条件
	if req.Keyword != "" {
		filters["keyword"] = req.Keyword
	}

	// 直接查询标签
	tags, total, err := l.queryTagsWithFilters(filters, req.Current, req.PageSize)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "get_tag_list"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询标签列表失败")
		return &types.GetTagListResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "查询标签列表失败"},
		}, nil
	}

	// 数据已经在查询函数中转换完成

	// 获取分页参数
	current, pageSize, _ := l.normalizePagination(req.Current, req.PageSize)

	// 记录操作成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tags"),
		logx.Field("operation", "get_tag_list"),
		logx.Field("status", "success"),
		logx.Field("scope", req.Scope),
		logx.Field("total", total),
		logx.Field("current", current),
		logx.Field("page_size", pageSize),
		logx.Field("user_key", jwtUser.UserKey),
	).Info("获取标签列表成功")

	return &types.GetTagListResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "获取成功"},
		PageParams: types.PageParams{
			Current:  current,
			PageSize: pageSize,
			Total:    total,
		},
		Data: types.TagListData{List: tags},
	}, nil
}

// ========== 内部辅助函数 ==========

// 规范化分页参数
func (l *GetTagListLogic) normalizePagination(current, pageSize int64) (normalizedCurrent, normalizedPageSize, offset int64) {
	if current <= 0 {
		current = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset = (current - 1) * pageSize
	return current, pageSize, offset
}

// 查询标签列表
func (l *GetTagListLogic) queryTagsWithFilters(filters map[string]interface{}, current, pageSize int64) ([]types.TagDefinition, int64, error) {
	// 规范化分页参数
	current, pageSize, offset := l.normalizePagination(current, pageSize)

	// 构建WHERE条件
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	// 应用过滤器
	if scope, ok := filters["scope"].(string); ok && scope != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("scope = $%d", argIndex))
		args = append(args, scope)
		argIndex++
	}

	// 租户隔离
	if scope, ok := filters["scope"].(string); ok {
		if tenantId, hasTenantId := filters["tenant_id"].(string); ok {
			if scope == "tenant" {
				// 租户标签：查询全局标签（tenant_id 为 NULL）
				whereConditions = append(whereConditions, "tenant_id IS NULL")
			} else if hasTenantId {
				// 用户标签：查询当前租户的标签
				whereConditions = append(whereConditions, fmt.Sprintf("tenant_id = $%d", argIndex))
				args = append(args, tenantId)
				argIndex++
			}
		}
	}

	// 关键词搜索
	if keyword, ok := filters["keyword"].(string); ok && keyword != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(label ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex+1))
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
		argIndex += 2
	}

	whereSQL := "WHERE " + strings.Join(whereConditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM system_tag_definitions
		%s
	`, whereSQL)

	var total int64
	err := l.svcCtx.DBConn.QueryRow(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []types.TagDefinition{}, 0, nil
	}

	// 查询标签列表
	listQuery := fmt.Sprintf(`
		SELECT id, scope, label, 
		       COALESCE(description, '') as description,
		       tenant_id, created_at, updated_at
		FROM system_tag_definitions
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	listArgs := append(args, pageSize, offset)

	var tagRows []struct {
		Id          string         `db:"id"`
		Scope       string         `db:"scope"`
		Label       string         `db:"label"`
		Description string         `db:"description"`
		TenantId    sql.NullString `db:"tenant_id"`
		CreatedAt   time.Time      `db:"created_at"`
		UpdatedAt   time.Time      `db:"updated_at"`
	}

	err = l.svcCtx.DBConn.QueryRowsPartial(&tagRows, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}

	// 转换为返回类型
	var tags []types.TagDefinition
	for _, row := range tagRows {
		tags = append(tags, types.TagDefinition{
			Id:          row.Id,
			Scope:       row.Scope,
			Label:       row.Label,
			Description: row.Description,
			CreatedAt:   row.CreatedAt.Unix(),
			UpdatedAt:   row.UpdatedAt.Unix(),
		})
	}

	return tags, total, nil
}
