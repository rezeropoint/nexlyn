package tenant

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

type GetTenantListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantListLogic {
	return &GetTenantListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantListLogic) GetTenantList(req *types.GetTenantListRequest) (resp *types.GetTenantListResponse, err error) {
	// 用户认证
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		return &types.GetTenantListResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 权限验证：检查是否有租户列表查看权限
	hasReadPermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: core.ResourceTenant, Action: core.ActionRead})
	if err != nil {
		return &types.GetTenantListResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}
	if !hasReadPermission {
		return &types.GetTenantListResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要租户查看权限"},
		}, nil
	}

	// 构建查询过滤器
	filters := make(map[string]interface{})

	// 获取用户可访问的租户列表进行过滤
	userTenants, err := l.svcCtx.Casbinx.GetUserTenants(jwtUser.UserKey)
	if err != nil {
		return &types.GetTenantListResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "获取可访问租户列表失败"},
		}, nil
	}

	// 如果用户不能访问所有租户（非超级管理员），则限制查询范围
	if len(userTenants) > 0 {
		filters["accessible_tenants"] = userTenants
	}

	// 添加搜索条件
	if req.Id != "" {
		filters["id"] = req.Id
	}
	if req.TenantKey != "" {
		filters["tenant_key"] = req.TenantKey
	}
	if req.TenantName != "" {
		filters["tenant_name"] = req.TenantName
	}
	if req.Status != "" {
		filters["status"] = req.Status
	}

	// 直接查询租户
	tenants, total, err := l.queryTenantsWithFilters(filters, req.Current, req.PageSize)
	if err != nil {
		return &types.GetTenantListResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "查询失败"},
		}, nil
	}

	// 处理tags过滤条件（如果需要）
	if len(req.Tags) > 0 && len(tenants) > 0 {
		// 过滤出具有指定标签的租户
		tagFilterQuery := `
			SELECT DISTINCT ttr.tenant_id::text
			FROM system_tenant_tag_relations ttr
			JOIN system_tag_definitions td ON ttr.tag_id = td.id
			WHERE td.label = ANY($1::text[])
			AND td.scope = 'tenant'
		`
		var filteredTenantIds []struct {
			TenantId string `db:"tenant_id"`
		}
		err = l.svcCtx.DBConn.QueryRowsPartial(&filteredTenantIds, tagFilterQuery, pq.Array(req.Tags))
		if err != nil {
			logx.WithContext(l.ctx).Info("标签过滤查询失败")
		} else {
			// 过滤tenants列表
			filteredMap := make(map[string]bool)
			for _, id := range filteredTenantIds {
				filteredMap[id.TenantId] = true
			}

			var filteredTenants []types.Tenant
			for _, tenant := range tenants {
				if filteredMap[tenant.Id] {
					filteredTenants = append(filteredTenants, tenant)
				}
			}
			tenants = filteredTenants
		}
	}

	// 获取分页参数
	current, pageSize, _ := l.normalizePagination(req.Current, req.PageSize)

	// 数据已经在查询函数中处理完成，包括标签

	return &types.GetTenantListResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "获取成功"},
		Data:         types.TenantListData{List: tenants},
		PageParams: types.PageParams{
			Current:  current,
			PageSize: pageSize,
			Total:    total,
		},
	}, nil
}

// ========== 内部辅助函数 ==========

// 规范化分页参数
func (l *GetTenantListLogic) normalizePagination(current, pageSize int64) (normalizedCurrent, normalizedPageSize, offset int64) {
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

// 查询租户列表
func (l *GetTenantListLogic) queryTenantsWithFilters(filters map[string]interface{}, current, pageSize int64) ([]types.Tenant, int64, error) {
	// 规范化分页参数
	current, pageSize, offset := l.normalizePagination(current, pageSize)

	// 构建WHERE条件
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	// 基础条件：不查询已删除租户
	whereConditions = append(whereConditions, "status != 'deleted'")

	// 添加过滤条件
	if id, ok := filters["id"].(string); ok && id != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("id::text = $%d", argIndex))
		args = append(args, id)
		argIndex++
	}
	if tenantKey, ok := filters["tenant_key"].(string); ok && tenantKey != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("tenant_key = $%d", argIndex))
		args = append(args, tenantKey)
		argIndex++
	}
	if tenantName, ok := filters["tenant_name"].(string); ok && tenantName != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("tenant_name ILIKE $%d", argIndex))
		args = append(args, "%"+tenantName+"%")
		argIndex++
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}
	if tenantKeyFilter, ok := filters["tenant_key_filter"].(string); ok && tenantKeyFilter != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("tenant_key = $%d", argIndex))
		args = append(args, tenantKeyFilter)
		argIndex++
	}
	if accessibleTenants, ok := filters["accessible_tenants"].([]string); ok && len(accessibleTenants) > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("tenant_key = ANY($%d::text[])", argIndex))
		args = append(args, pq.Array(accessibleTenants))
		argIndex++
	}

	whereSQL := "WHERE " + strings.Join(whereConditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM system_tenants
		%s
	`, whereSQL)

	var total int64
	err := l.svcCtx.DBConn.QueryRow(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []types.Tenant{}, 0, nil
	}

	// 查询租户列表
	listQuery := fmt.Sprintf(`
		SELECT 
			id::text as id,
			tenant_key,
			tenant_name,
			COALESCE(description, '') as description,
			COALESCE(contact_email, '') as contact_email,
			COALESCE(status, 'active') as status,
			created_at,
			updated_at,
			expires_at,
			COALESCE(created_by, '') as created_by,
			COALESCE(updated_by, '') as updated_by,
			COALESCE(max_users, 0) as max_users
		FROM system_tenants
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	listArgs := append(args, pageSize, offset)

	var tenantRows []struct {
		Id           string       `db:"id"`
		TenantKey    string       `db:"tenant_key"`
		TenantName   string       `db:"tenant_name"`
		Description  string       `db:"description"`
		ContactEmail string       `db:"contact_email"`
		Status       string       `db:"status"`
		CreatedAt    time.Time    `db:"created_at"`
		UpdatedAt    time.Time    `db:"updated_at"`
		ExpiresAt    sql.NullTime `db:"expires_at"`
		CreatedBy    string       `db:"created_by"`
		UpdatedBy    string       `db:"updated_by"`
		MaxUsers     int          `db:"max_users"`
	}

	err = l.svcCtx.DBConn.QueryRowsPartial(&tenantRows, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}

	// 转换为目标结构体
	var tenants []types.Tenant
	tenantIds := make([]string, 0, len(tenantRows))

	for _, row := range tenantRows {
		tenant := types.Tenant{
			Id:           row.Id,
			TenantKey:    row.TenantKey,
			TenantName:   row.TenantName,
			Description:  row.Description,
			Tags:         []string{}, // 后续查询填充
			ContactEmail: row.ContactEmail,
			Status:       row.Status,
			CreatedAt:    row.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    row.UpdatedAt.Format(time.RFC3339),
			CreatedBy:    row.CreatedBy,
			UpdatedBy:    row.UpdatedBy,
			MaxUsers:     row.MaxUsers,
		}

		if row.ExpiresAt.Valid {
			tenant.ExpiresAt = row.ExpiresAt.Time.Format(time.RFC3339)
		}

		tenants = append(tenants, tenant)
		tenantIds = append(tenantIds, row.Id)
	}

	// 批量查询所有租户的标签
	if len(tenantIds) > 0 {
		// 批量查询标签
		tagsMap, err := l.batchQueryTenantTags(tenantIds)
		if err != nil {
			logx.WithContext(l.ctx).Info("批量查询租户标签失败，继续返回租户列表但不包含标签")
		} else {
			// 填充租户的标签信息
			for i := range tenants {
				if tags, exists := tagsMap[tenants[i].Id]; exists {
					tenants[i].Tags = tags
				}
			}
		}
	}

	return tenants, total, nil
}

// 批量查询租户标签
func (l *GetTenantListLogic) batchQueryTenantTags(tenantIds []string) (map[string][]string, error) {
	if len(tenantIds) == 0 {
		return map[string][]string{}, nil
	}

	query := `
		SELECT ttr.tenant_id::text, td.label
		FROM system_tenant_tag_relations ttr
		JOIN system_tag_definitions td ON ttr.tag_id = td.id
		WHERE ttr.tenant_id = ANY($1::uuid[])
		AND td.scope = 'tenant'
		ORDER BY ttr.tenant_id, td.label
	`

	var tagRows []struct {
		TenantId string `db:"tenant_id"`
		Label    string `db:"label"`
	}

	err := l.svcCtx.DBConn.QueryRowsPartial(&tagRows, query, pq.Array(tenantIds))
	if err != nil {
		return nil, err
	}

	// 构建映射
	result := make(map[string][]string)
	for _, row := range tagRows {
		result[row.TenantId] = append(result[row.TenantId], row.Label)
	}

	return result, nil
}
