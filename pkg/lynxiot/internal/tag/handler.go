package tag

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// tagManager 标签管理器实现
type tagManager struct {
	dbConn sqlx.SqlConn
	config Config
}

// newTagManager 创建标签管理器实例
func newTagManager(dbConn sqlx.SqlConn, config Config) (*tagManager, error) {
	return &tagManager{
		dbConn: dbConn,
		config: config,
	}, nil
}

// CreateTag 创建标签
func (m *tagManager) CreateTag(ctx context.Context, metadata core.DeviceTagMetadata) (string, error) {
	// 1. 验证元数据
	if err := metadata.Validate(); err != nil {
		return "", fmt.Errorf("元数据验证失败: %w", err)
	}

	// 2. 检查标签名称是否已存在（租户内）
	var count int
	checkQuery := "SELECT COUNT(*) FROM iot_tag_definitions WHERE label = $1 AND tenant_id = $2 AND deleted_at IS NULL"
	err := m.dbConn.QueryRowCtx(ctx, &count, checkQuery, metadata.Label, metadata.TenantID)
	if err != nil {
		return "", fmt.Errorf("检查标签是否存在失败: %w", err)
	}
	if count > 0 {
		return "", fmt.Errorf("标签名称已存在")
	}

	// 3. 生成UUID
	id := uuid.New().String()

	// 4. 插入PostgreSQL
	insertQuery := `
		INSERT INTO iot_tag_definitions (id, label, description, color, tenant_id, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = m.dbConn.ExecCtx(ctx, insertQuery,
		id, metadata.Label, metadata.Description, metadata.Color, metadata.TenantID, metadata.CreatedBy,
	)
	if err != nil {
		return "", fmt.Errorf("插入标签失败: %w", err)
	}

	return id, nil
}

// GetTag 获取标签详情
func (m *tagManager) GetTag(ctx context.Context, tagID string, tenantID string) (*core.DeviceTag, error) {
	var result struct {
		ID          string         `db:"id"`
		Label       string         `db:"label"`
		Description sql.NullString `db:"description"`
		Color       sql.NullString `db:"color"`
		TenantID    string         `db:"tenant_id"`
		CreatedBy   sql.NullString `db:"created_by"`
		UpdatedBy   sql.NullString `db:"updated_by"`
		CreatedAt   sql.NullTime   `db:"created_at"`
		UpdatedAt   sql.NullTime   `db:"updated_at"`
	}

	query := `
		SELECT id, label, description, color, tenant_id, created_by, updated_by, created_at, updated_at
		FROM iot_tag_definitions
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`
	err := m.dbConn.QueryRowCtx(ctx, &result, query, tagID, tenantID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("标签不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询标签失败: %w", err)
	}

	tag := &core.DeviceTag{
		ID:          result.ID,
		Label:       result.Label,
		Description: nullStringToString(result.Description),
		Color:       nullStringToString(result.Color),
		TenantID:    result.TenantID,
		CreatedBy:   nullStringToString(result.CreatedBy),
		UpdatedBy:   nullStringToString(result.UpdatedBy),
		CreatedAt:   nullTimeToString(result.CreatedAt),
		UpdatedAt:   nullTimeToString(result.UpdatedAt),
	}

	return tag, nil
}

// ListTags 查询标签列表
func (m *tagManager) ListTags(ctx context.Context, query core.DeviceTagQuery) ([]*core.DeviceTagSummary, int64, error) {
	// 1. 构建SQL查询
	baseQuery := "FROM iot_tag_definitions WHERE tenant_id = $1 AND deleted_at IS NULL"
	var args []any
	args = append(args, query.TenantID)
	argIndex := 2

	// 添加关键词搜索
	if query.Keyword != "" {
		baseQuery += fmt.Sprintf(" AND (label ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+query.Keyword+"%")
		argIndex++
	}

	// 2. 查询总数
	var total int64
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := m.dbConn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}

	// 3. 查询列表
	listQuery := "SELECT id, label, description, color, created_at " + baseQuery
	listQuery += " ORDER BY created_at DESC"

	// 添加分页
	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	listQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)

	var results []struct {
		ID          string         `db:"id"`
		Label       string         `db:"label"`
		Description sql.NullString `db:"description"`
		Color       sql.NullString `db:"color"`
		CreatedAt   sql.NullTime   `db:"created_at"`
	}

	err = m.dbConn.QueryRowsCtx(ctx, &results, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询列表失败: %w", err)
	}

	// 4. 转换结果
	tags := make([]*core.DeviceTagSummary, 0, len(results))
	for _, r := range results {
		t := &core.DeviceTagSummary{
			ID:          r.ID,
			Label:       r.Label,
			Description: nullStringToString(r.Description),
			Color:       nullStringToString(r.Color),
			CreatedAt:   nullTimeToString(r.CreatedAt),
		}
		tags = append(tags, t)
	}

	return tags, total, nil
}

// UpdateTag 更新标签
func (m *tagManager) UpdateTag(ctx context.Context, tagID string, tenantID string, update core.DeviceTagUpdate) error {
	// 1. 检查标签是否存在并验证租户权限
	var existingTenantID string
	checkQuery := "SELECT tenant_id FROM iot_tag_definitions WHERE id = $1 AND deleted_at IS NULL"
	err := m.dbConn.QueryRowCtx(ctx, &existingTenantID, checkQuery, tagID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("标签不存在")
	}
	if err != nil {
		return fmt.Errorf("查询标签失败: %w", err)
	}

	if existingTenantID != tenantID {
		return fmt.Errorf("无权限修改其他租户的标签")
	}

	// 2. 构建更新SQL
	updateQuery := "UPDATE iot_tag_definitions SET updated_at = CURRENT_TIMESTAMP"
	var args []any
	argIndex := 1

	if update.Label != "" {
		updateQuery += fmt.Sprintf(", label = $%d", argIndex)
		args = append(args, update.Label)
		argIndex++
	}
	if update.Description != "" {
		updateQuery += fmt.Sprintf(", description = $%d", argIndex)
		args = append(args, update.Description)
		argIndex++
	}
	if update.Color != "" {
		updateQuery += fmt.Sprintf(", color = $%d", argIndex)
		args = append(args, update.Color)
		argIndex++
	}

	updateQuery += fmt.Sprintf(" WHERE id = $%d AND tenant_id = $%d", argIndex, argIndex+1)
	args = append(args, tagID, tenantID)

	// 3. 执行更新
	_, err = m.dbConn.ExecCtx(ctx, updateQuery, args...)
	if err != nil {
		return fmt.Errorf("更新标签失败: %w", err)
	}

	return nil
}

// DeleteTag 删除标签
func (m *tagManager) DeleteTag(ctx context.Context, tagID string, tenantID string) error {
	// 1. 检查标签是否存在并验证租户权限
	var existingTenantID string
	checkQuery := "SELECT tenant_id FROM iot_tag_definitions WHERE id = $1 AND deleted_at IS NULL"
	err := m.dbConn.QueryRowCtx(ctx, &existingTenantID, checkQuery, tagID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("标签不存在")
	}
	if err != nil {
		return fmt.Errorf("查询标签失败: %w", err)
	}

	if existingTenantID != tenantID {
		return fmt.Errorf("无权限删除其他租户的标签")
	}

	// 2. 软删除
	deleteQuery := "UPDATE iot_tag_definitions SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1"
	_, err = m.dbConn.ExecCtx(ctx, deleteQuery, tagID)
	if err != nil {
		return fmt.Errorf("删除标签失败: %w", err)
	}

	return nil
}
