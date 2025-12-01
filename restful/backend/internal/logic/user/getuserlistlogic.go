package user

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

	"github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取用户列表
func NewGetUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserListLogic {
	return &GetUserListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserListLogic) GetUserList(req *types.GetUserListRequest) (resp *types.GetUserListResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "get_user_list"),
		logx.Field("status", "started"),
	).Info("开始获取用户列表")

	// 获取当前操作者（用于租户隔离）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "get_user_list"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户失败")
		return &types.GetUserListResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 构建查询过滤器
	filters := make(map[string]interface{})

	// 按用户ID列表查询（优先级最高，传入时忽略其他搜索条件）
	if len(req.Ids) > 0 {
		filters["ids"] = req.Ids
	} else {
		// 根据搜索条件添加过滤器
		if req.Name != "" {
			filters["name"] = req.Name
		}
		if req.Email != "" {
			filters["email"] = req.Email
		}
		if req.UserName != "" {
			filters["user_name"] = req.UserName
		}
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
				logx.Field("module", "user"),
				logx.Field("operation", "get_user_list"),
				logx.Field("status", "warning"),
				logx.Field("status_filter", req.Status),
			).Info("无效的状态过滤值")
		} else {
			filters["status"] = req.Status
		}
	}

	// 权限验证：检查用户查看权限
	hasReadPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		core.Permission{Resource: core.ResourceUser, Action: core.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "get_user_list"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")
		return &types.GetUserListResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasReadPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "get_user_list"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")
		return &types.GetUserListResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要用户查看权限"},
		}, nil
	}

	// 多租户过滤：
	// - 有跨租户权限：可查看所有租户；若指定 tenantId，则按指定租户过滤
	// - 无跨租户权限：强制限定为其自身租户
	hasCrossTenantPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		core.Permission{Resource: core.ResourceTenant, Action: core.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "get_user_list"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("跨租户权限检查失败")
		return &types.GetUserListResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if hasCrossTenantPermission {
		// 有跨租户权限，可以查看指定租户或所有租户
		if req.TenantId != "" {
			filters["tenant_key"] = req.TenantId
		}
	} else {
		// 无跨租户权限，只能查看自己租户（使用TenantKey匹配tenant_key字段）
		filters["tenant_key"] = jwtUser.TenantKey
	}

	// 标签过滤（按标签ID，包含任一）
	if len(req.TagIds) > 0 {
		filters["tag_ids"] = req.TagIds
	}

	// 直接查询用户
	users, total, err := l.queryUsersWithFilters(filters, req.Current, req.PageSize)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "get_user_list"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询用户列表失败")

		return &types.GetUserListResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "查询用户列表失败",
			},
		}, nil
	}

	// 用户数据已经在查询函数中处理完成，包括标签

	// 获取分页参数
	current, pageSize, _ := l.normalizePagination(req.Current, req.PageSize)

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "get_user_list"),
		logx.Field("status", "success"),
		logx.Field("total", total),
		logx.Field("current", current),
		logx.Field("page_size", pageSize),
	).Info("获取用户列表成功")

	return &types.GetUserListResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "获取用户列表成功",
		},
		PageParams: types.PageParams{
			Current:  current,
			PageSize: pageSize,
			Total:    total,
		},
		Data: types.UserListData{
			List: users,
		},
	}, nil
}

// ========== 内部辅助函数 ==========

// 规范化分页参数
func (l *GetUserListLogic) normalizePagination(current, pageSize int64) (normalizedCurrent, normalizedPageSize, offset int64) {
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

// 查询用户列表
func (l *GetUserListLogic) queryUsersWithFilters(filters map[string]interface{}, current, pageSize int64) ([]types.UserBrief, int64, error) {
	// 规范化分页参数
	current, pageSize, offset := l.normalizePagination(current, pageSize)

	// 构建WHERE条件
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	// 基础条件：不查询已删除用户
	whereConditions = append(whereConditions, "u.status != 'deleted'")

	// 添加过滤条件
	// 按用户ID列表查询（优先级最高）
	if ids, ok := filters["ids"].([]string); ok && len(ids) > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("u.id = ANY($%d::uuid[])", argIndex))
		args = append(args, pq.Array(ids))
		argIndex++
	} else {
		// 其他搜索条件
		if name, ok := filters["name"].(string); ok && name != "" {
			whereConditions = append(whereConditions, fmt.Sprintf("u.name ILIKE $%d", argIndex))
			args = append(args, "%"+name+"%")
			argIndex++
		}
		if email, ok := filters["email"].(string); ok && email != "" {
			whereConditions = append(whereConditions, fmt.Sprintf("u.email ILIKE $%d", argIndex))
			args = append(args, "%"+email+"%")
			argIndex++
		}
		if userName, ok := filters["user_name"].(string); ok && userName != "" {
			whereConditions = append(whereConditions, fmt.Sprintf("u.user_name ILIKE $%d", argIndex))
			args = append(args, "%"+userName+"%")
			argIndex++
		}
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("u.status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}
	if tenantKey, ok := filters["tenant_key"].(string); ok && tenantKey != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("t.tenant_key = $%d", argIndex))
		args = append(args, tenantKey)
		argIndex++
	}
	if tagIds, ok := filters["tag_ids"].([]string); ok && len(tagIds) > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("EXISTS (SELECT 1 FROM user_tag_relations utr WHERE utr.user_id = u.id AND utr.tag_id = ANY($%d::uuid[]))", argIndex))
		args = append(args, pq.Array(tagIds))
		argIndex++
	}

	whereSQL := "WHERE " + strings.Join(whereConditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM system_users u
		LEFT JOIN system_tenants t ON u.tenant_id = t.id
		%s
	`, whereSQL)

	var total int64
	err := l.svcCtx.DBConn.QueryRow(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []types.UserBrief{}, 0, nil
	}

	// 查询用户列表数据
	listQuery := fmt.Sprintf(`
		SELECT 
			u.id, u.user_key, u.user_name, u.name, u.email, 
			COALESCE(u.phone, '') as phone,
			COALESCE(u.avatar, '') as avatar, 
			COALESCE(u.signature, '') as signature,
			COALESCE(u.title, '') as title,
			COALESCE(u.country, '') as country,
			COALESCE(u.address, '') as address,
			u.status,
			u.tenant_id::text as tenant_id,
			COALESCE(t.tenant_name, '') as tenant_name,
			u.created_at,
			u.updated_at,
			u.last_login_at
		FROM system_users u
		LEFT JOIN system_tenants t ON u.tenant_id = t.id
		%s
		ORDER BY u.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	listArgs := append(args, pageSize, offset)

	var userRows []struct {
		Id          string       `db:"id"`
		UserKey     string       `db:"user_key"`
		UserName    string       `db:"user_name"`
		Name        string       `db:"name"`
		Email       string       `db:"email"`
		Phone       string       `db:"phone"`
		Avatar      string       `db:"avatar"`
		Signature   string       `db:"signature"`
		Title       string       `db:"title"`
		Country     string       `db:"country"`
		Address     string       `db:"address"`
		Status      string       `db:"status"`
		TenantId    string       `db:"tenant_id"`
		TenantName  string       `db:"tenant_name"`
		CreatedAt   time.Time    `db:"created_at"`
		UpdatedAt   time.Time    `db:"updated_at"`
		LastLoginAt sql.NullTime `db:"last_login_at"`
	}

	err = l.svcCtx.DBConn.QueryRowsPartial(&userRows, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}

	// 转换为响应格式并收集用户ID
	var users []types.UserBrief
	userIds := make([]string, 0, len(userRows))
	for _, userRow := range userRows {
		user := types.UserBrief{
			Id:       userRow.Id,
			UserKey:  userRow.UserKey,
			UserName: userRow.UserName,
			Name:     userRow.Name,
			Email:    userRow.Email,
			Avatar:   userRow.Avatar,
			Status:   userRow.Status,
			IsActive: userRow.Status == "active",
			TenantInfo: types.TenantInfo{
				TenantId:   userRow.TenantId,
				TenantName: userRow.TenantName,
			},
			Tags:      []types.UserTag{}, // 后续查询填充
			CreatedAt: userRow.CreatedAt.Unix(),
			UpdatedAt: userRow.UpdatedAt.Unix(),
		}

		if userRow.LastLoginAt.Valid {
			user.LastLoginAt = userRow.LastLoginAt.Time.Unix()
		}

		users = append(users, user)
		userIds = append(userIds, userRow.Id)
	}

	// 批量获取用户标签
	if len(userIds) > 0 {
		userTagsMap, err := l.batchQueryUserTags(userIds)
		if err != nil {
			// 标签查询失败不影响主要功能，记录日志但继续处理
			logx.WithContext(l.ctx).WithFields(
				logx.Field("error", err.Error()),
			).Info("获取用户标签失败，继续返回用户列表但不包含标签")
		} else {
			// 填充用户的标签信息
			for i := range users {
				if tags, exists := userTagsMap[users[i].Id]; exists {
					users[i].Tags = tags
				}
			}
		}
	}

	return users, total, nil
}

// 批量查询用户标签
func (l *GetUserListLogic) batchQueryUserTags(userIds []string) (map[string][]types.UserTag, error) {
	if len(userIds) == 0 {
		return map[string][]types.UserTag{}, nil
	}

	query := `
		SELECT utr.user_id::text, td.id::text as tag_id, td.label
		FROM user_tag_relations utr
		JOIN system_tag_definitions td ON utr.tag_id = td.id
		WHERE utr.user_id = ANY($1::uuid[])
		AND td.scope = 'user'
		ORDER BY utr.user_id, td.label
	`

	var tagRows []struct {
		UserId string `db:"user_id"`
		TagId  string `db:"tag_id"`
		Label  string `db:"label"`
	}

	err := l.svcCtx.DBConn.QueryRowsPartial(&tagRows, query, pq.Array(userIds))
	if err != nil {
		return nil, err
	}

	// 构建映射
	result := make(map[string][]types.UserTag)
	for _, row := range tagRows {
		result[row.UserId] = append(result[row.UserId], types.UserTag{
			Id:    row.TagId,
			Label: row.Label,
		})
	}

	return result, nil
}
