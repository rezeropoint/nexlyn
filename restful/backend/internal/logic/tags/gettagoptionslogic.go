package tags

import (
	"context"
	"fmt"
	"strings"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTagOptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取标签下拉选项
func NewGetTagOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTagOptionsLogic {
	return &GetTagOptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTagOptionsLogic) GetTagOptions(req *types.GetTagOptionsRequest) (resp *types.GetTagOptionsResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tags"),
		logx.Field("operation", "get_tag_options"),
		logx.Field("status", "started"),
	).Info("开始获取标签下拉选项")

	// 验证scope值
	if req.Scope != "tenant" && req.Scope != "user" {
		return &types.GetTagOptionsResponse{
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
			logx.Field("operation", "get_tag_options"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户失败")
		return &types.GetTagOptionsResponse{
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
			logx.Field("operation", "get_tag_options"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")
		return &types.GetTagOptionsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "get_tag_options"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")
		return &types.GetTagOptionsResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: fmt.Sprintf("权限不足，需要%s读取权限", permissionResource)},
		}, nil
	}

	// 设置默认限制
	limit := req.Limit
	if limit <= 0 {
		limit = 50 // 默认返50条
	}
	if limit > 200 {
		limit = 200 // 最大200条
	}

	// 构建查询条件（包含租户隔离）
	var whereClause []string
	var args []interface{}
	argIndex := 1

	// scope条件（必需）
	whereClause = append(whereClause, fmt.Sprintf("td.scope = $%d", argIndex))
	args = append(args, req.Scope)
	argIndex++

	// 租户隔离条件
	if req.Scope == "tenant" {
		// 租户标签：只查询全局标签（tenant_id为NULL）
		whereClause = append(whereClause, "td.tenant_id IS NULL")
	} else {
		// 用户标签：只查询当前租户的标签
		whereClause = append(whereClause, fmt.Sprintf("td.tenant_id = $%d", argIndex))
		args = append(args, jwtUser.TenantId)
		argIndex++
	}

	// 关键词搜索条件
	if req.Keyword != "" {
		keyword := strings.TrimSpace(req.Keyword)
		if keyword != "" {
			whereClause = append(whereClause, fmt.Sprintf("td.label ILIKE $%d", argIndex))
			args = append(args, "%"+keyword+"%")
			argIndex++
		}
	}

	whereSQL := ""
	if len(whereClause) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClause, " AND ")
	}

	// 添加limit参数
	args = append(args, limit)

	// 构建查询SQL：查询标签并统计使用数量
	// 这里将用户标签和租户标签的使用数量加起来
	querySQL := fmt.Sprintf(`
		SELECT 
			td.id,
			td.label,
			(COALESCE(user_usage.count, 0) + COALESCE(tenant_usage.count, 0)) as count
		FROM system_tag_definitions td
		LEFT JOIN (
			SELECT tag_id, COUNT(*) as count 
			FROM user_tag_relations 
			GROUP BY tag_id
		) user_usage ON td.id = user_usage.tag_id
		LEFT JOIN (
			SELECT tag_id, COUNT(*) as count 
			FROM system_tenant_tag_relations 
			GROUP BY tag_id
		) tenant_usage ON td.id = tenant_usage.tag_id
		%s
		ORDER BY count DESC, td.created_at DESC
		LIMIT $%d
	`, whereSQL, argIndex)

	var optionRows []struct {
		Id    string `db:"id"`
		Label string `db:"label"`
		Count int64  `db:"count"`
	}

	err = l.svcCtx.DBConn.QueryRowsPartial(&optionRows, querySQL, args...)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "get_tag_options"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询标签选项失败")
		return &types.GetTagOptionsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "查询标签选项失败"},
		}, nil
	}

	// 转换为返回类型
	var tagOptions []types.TagOption
	for _, row := range optionRows {
		tagOptions = append(tagOptions, types.TagOption{
			Id:    row.Id,
			Label: row.Label,
			Count: row.Count,
		})
	}

	// 记录操作成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tags"),
		logx.Field("operation", "get_tag_options"),
		logx.Field("status", "success"),
		logx.Field("scope", req.Scope),
		logx.Field("count", len(tagOptions)),
		logx.Field("limit", limit),
		logx.Field("user_key", jwtUser.UserKey),
	).Info("获取标签下拉选项成功")

	return &types.GetTagOptionsResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "获取成功"},
		Data: struct {
			List []types.TagOption `json:"list"`
		}{List: tagOptions},
	}, nil
}
