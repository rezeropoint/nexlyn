package tag

import (
	"context"
	"fmt"
	"strings"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type tagManager struct {
	dbConn sqlx.SqlConn
	config Config
}

func newTagManager(dbConn sqlx.SqlConn, config Config) (Manager, error) {
	return &tagManager{
		dbConn: dbConn,
		config: config,
	}, nil
}

// InitTagTable 初始化数据库表
func (m *tagManager) InitTagTable(ctx context.Context) error {
	// 先检查表是否已存在
	var count int
	err := m.dbConn.QueryRowCtx(ctx, &count, CheckTableExistsSQL)
	if err != nil {
		return fmt.Errorf("检查表存在性失败: %w", err)
	}

	// 如果表已存在，直接返回
	if count > 0 {
		return nil
	}

	// 表不存在，执行创建
	_, err = m.dbConn.ExecCtx(ctx, CreateTableSQL)
	if err != nil {
		return fmt.Errorf("创建表失败: %w", err)
	}
	return nil
}

// CreateTag 创建标签
func (m *tagManager) CreateTag(ctx context.Context, metadata core.LynxTagMetadata) (string, error) {
	// 验证输入参数
	if err := metadata.Validate(); err != nil {
		return "", fmt.Errorf("标签元数据验证失败: %w", err)
	}

	// 生成ID（通过数据库函数）
	var tagID string
	query := "SELECT gen_random_uuid()::text"
	if err := m.dbConn.QueryRowCtx(ctx, &tagID, query); err != nil {
		return "", fmt.Errorf("生成UUID失败: %w", err)
	}

	// 创建标签（created_at和updated_at由数据库自动生成）
	insertQuery := fmt.Sprintf(`
		INSERT INTO %s (id, tenant_id, name, description, scope, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, TableName)

	_, err := m.dbConn.ExecCtx(ctx, insertQuery,
		tagID,
		metadata.TenantID,
		metadata.Name,
		metadata.Description,
		metadata.Scope,
		metadata.CreatedBy,
		metadata.CreatedBy, // updated_by初始值同created_by
	)

	if err != nil {
		logx.Errorf("创建标签失败: %v", err)
		return "", fmt.Errorf("创建标签失败: %w", err)
	}

	return tagID, nil
}

// GetTag 获取标签详情
func (m *tagManager) GetTag(ctx context.Context, tagID string, tenantID string) (*core.LynxTag, error) {
	query := fmt.Sprintf(`
		SELECT id, tenant_id, name, description, scope, created_by, updated_by,
		       EXTRACT(EPOCH FROM created_at)::bigint as created_at,
		       EXTRACT(EPOCH FROM updated_at)::bigint as updated_at
		FROM %s
		WHERE id = $1 AND tenant_id = $2
	`, TableName)

	var tagRow tagDB
	if err := m.dbConn.QueryRowCtx(ctx, &tagRow, query, tagID, tenantID); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, fmt.Errorf("标签不存在")
		}
		logx.Errorf("查询标签失败: %v", err)
		return nil, fmt.Errorf("查询标签失败: %w", err)
	}

	return tagDBToLynxTag(&tagRow), nil
}

// ListTags 查询标签列表
func (m *tagManager) ListTags(ctx context.Context, query core.LynxTagQuery) ([]*core.LynxTagSummary, int64, error) {
	// 验证查询条件
	if err := query.ValidateQuery(); err != nil {
		return nil, 0, fmt.Errorf("查询条件验证失败: %w", err)
	}

	// 构建WHERE条件和参数
	whereClauses := []string{}
	args := []interface{}{}
	paramIdx := 1

	// tenant_id条件
	whereClauses = append(whereClauses, fmt.Sprintf("tenant_id = $%d", paramIdx))
	args = append(args, query.TenantID)
	paramIdx++

	// scope条件
	if query.Scope != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("scope = $%d", paramIdx))
		args = append(args, query.Scope)
		paramIdx++
	}

	// keyword条件
	if query.Keyword != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", paramIdx, paramIdx+1))
		keywordPattern := "%" + query.Keyword + "%"
		args = append(args, keywordPattern, keywordPattern)
		paramIdx += 2
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", TableName, whereSQL)
	var total int64
	if err := m.dbConn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		logx.Errorf("查询标签总数失败: %v", err)
		return nil, 0, fmt.Errorf("查询标签总数失败: %w", err)
	}

	// 查询列表（分页）
	offset := (query.Page - 1) * query.PageSize
	listQuery := fmt.Sprintf(`
		SELECT id, name, description, scope,
		       EXTRACT(EPOCH FROM created_at)::bigint as created_at,
		       EXTRACT(EPOCH FROM updated_at)::bigint as updated_at
		FROM %s
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, TableName, whereSQL, paramIdx, paramIdx+1)

	listArgs := append(args, query.PageSize, offset)
	var tagRows []tagSummaryDB
	if err := m.dbConn.QueryRowsCtx(ctx, &tagRows, listQuery, listArgs...); err != nil {
		logx.Errorf("查询标签列表失败: %v", err)
		return nil, 0, fmt.Errorf("查询标签列表失败: %w", err)
	}

	// 转换结果
	results := make([]*core.LynxTagSummary, 0, len(tagRows))
	for _, row := range tagRows {
		results = append(results, tagSummaryDBToLynxTagSummary(&row))
	}

	return results, total, nil
}

// UpdateTag 更新标签
func (m *tagManager) UpdateTag(ctx context.Context, tagID string, tenantID string, update core.LynxTagUpdate) error {
	// 构建UPDATE语句（跳过空字段）
	updateClauses := []string{}
	args := []interface{}{}
	paramIdx := 1

	if update.Name != "" {
		updateClauses = append(updateClauses, fmt.Sprintf("name = $%d", paramIdx))
		args = append(args, update.Name)
		paramIdx++
	}

	if update.Description != "" {
		updateClauses = append(updateClauses, fmt.Sprintf("description = $%d", paramIdx))
		args = append(args, update.Description)
		paramIdx++
	}

	if len(updateClauses) == 0 {
		return nil // 没有更新字段，直接返回
	}

	// 添加WHERE条件
	args = append(args, tagID, tenantID)

	updateSQL := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = $%d AND tenant_id = $%d",
		TableName,
		strings.Join(updateClauses, ", "),
		paramIdx, paramIdx+1,
	)

	_, err := m.dbConn.ExecCtx(ctx, updateSQL, args...)
	if err != nil {
		logx.Errorf("更新标签失败: %v", err)
		return fmt.Errorf("更新标签失败: %w", err)
	}

	return nil
}

// DeleteTag 删除标签（硬删除）
func (m *tagManager) DeleteTag(ctx context.Context, tagID string, tenantID string) error {
	deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE id = $1 AND tenant_id = $2", TableName)

	_, err := m.dbConn.ExecCtx(ctx, deleteQuery, tagID, tenantID)
	if err != nil {
		logx.Errorf("删除标签失败: %v", err)
		return fmt.Errorf("删除标签失败: %w", err)
	}

	return nil
}

// GetTagsByNames 根据标签名称批量查询
func (m *tagManager) GetTagsByNames(ctx context.Context, tenantID string, scope string, names []string) ([]*core.LynxTagSummary, error) {
	if len(names) == 0 {
		return []*core.LynxTagSummary{}, nil
	}

	// 构建IN子句
	placeholders := make([]string, len(names))
	args := make([]interface{}, 0, len(names)+2)
	args = append(args, tenantID, scope)

	// 为每个name生成占位符
	for i, name := range names {
		placeholders[i] = fmt.Sprintf("$%d", i+3) // 前两个占位符是tenantID和scope
		args = append(args, name)
	}

	inSQL := strings.Join(placeholders, ", ")
	query := fmt.Sprintf(`
		SELECT id, name, description, scope,
		       EXTRACT(EPOCH FROM created_at)::bigint as created_at,
		       EXTRACT(EPOCH FROM updated_at)::bigint as updated_at
		FROM %s
		WHERE tenant_id = $1 AND scope = $2 AND name IN (%s)
		ORDER BY created_at DESC
	`, TableName, inSQL)

	var tagRows []tagSummaryDB
	if err := m.dbConn.QueryRowsCtx(ctx, &tagRows, query, args...); err != nil {
		logx.Errorf("批量查询标签失败: %v", err)
		return nil, fmt.Errorf("批量查询标签失败: %w", err)
	}

	// 转换结果
	results := make([]*core.LynxTagSummary, 0, len(tagRows))
	for _, row := range tagRows {
		results = append(results, tagSummaryDBToLynxTagSummary(&row))
	}

	return results, nil
}

// GetTagsByIDs 根据标签ID列表批量查询标签详情
func (m *tagManager) GetTagsByIDs(ctx context.Context, tagIDs []string, tenantID string) ([]*core.LynxTagSummary, error) {
	if len(tagIDs) == 0 {
		return []*core.LynxTagSummary{}, nil
	}

	// 使用ANY运算符批量查询
	query := fmt.Sprintf(`
		SELECT id, name, description, scope,
		       EXTRACT(EPOCH FROM created_at)::bigint as created_at,
		       EXTRACT(EPOCH FROM updated_at)::bigint as updated_at
		FROM %s
		WHERE id = ANY($1) AND tenant_id = $2
		ORDER BY created_at DESC
	`, TableName)

	var tagRows []tagSummaryDB
	if err := m.dbConn.QueryRowsCtx(ctx, &tagRows, query, fmt.Sprintf("{%s}", strings.Join(tagIDs, ",")), tenantID); err != nil {
		logx.Errorf("批量查询标签详情失败: %v", err)
		return nil, fmt.Errorf("批量查询标签详情失败: %w", err)
	}

	// 转换结果
	results := make([]*core.LynxTagSummary, 0, len(tagRows))
	for _, row := range tagRows {
		results = append(results, tagSummaryDBToLynxTagSummary(&row))
	}

	return results, nil
}

// GetTagNamesByIDs 根据标签ID列表批量查询标签名称
func (m *tagManager) GetTagNamesByIDs(ctx context.Context, tagIDs []string, tenantID string) ([]string, error) {
	if len(tagIDs) == 0 {
		return []string{}, nil
	}

	// 使用ANY运算符批量查询
	query := fmt.Sprintf(`
		SELECT id, name
		FROM %s
		WHERE id = ANY($1) AND tenant_id = $2
	`, TableName)

	var tagRows []struct {
		ID   string `db:"id"`
		Name string `db:"name"`
	}
	if err := m.dbConn.QueryRowsCtx(ctx, &tagRows, query, fmt.Sprintf("{%s}", strings.Join(tagIDs, ",")), tenantID); err != nil {
		logx.Errorf("批量查询标签名称失败: %v", err)
		return nil, fmt.Errorf("批量查询标签名称失败: %w", err)
	}

	// 创建ID->Name映射
	idToName := make(map[string]string, len(tagRows))
	for _, row := range tagRows {
		idToName[row.ID] = row.Name
	}

	// 按照输入顺序返回名称列表，不存在的ID返回空字符串
	results := make([]string, len(tagIDs))
	for i, id := range tagIDs {
		if name, exists := idToName[id]; exists {
			results[i] = name
		} else {
			results[i] = ""
		}
	}

	return results, nil
}

// GetTagUsage 查询标签使用情况
func (m *tagManager) GetTagUsage(ctx context.Context, tagID string, tenantID string) (*TagUsage, error) {
	// 首先验证标签是否存在且属于该租户
	_, err := m.GetTag(ctx, tagID, tenantID)
	if err != nil {
		return nil, err
	}

	usage := &TagUsage{
		InfoAtomTypeIDs: []string{},
		GraphConfigIDs:  []string{},
	}

	// 查询信息原子类型使用情况
	infoAtomQuery := `SELECT info_atom_type_id FROM info_atom_type_tags WHERE tag_id = $1`
	var infoAtomTypeIDs []string
	if err := m.dbConn.QueryRowsCtx(ctx, &infoAtomTypeIDs, infoAtomQuery, tagID); err != nil && err != sqlx.ErrNotFound {
		logx.Errorf("查询标签在信息原子类型中的使用情况失败: %v", err)
		return nil, fmt.Errorf("查询标签使用情况失败: %w", err)
	}
	usage.InfoAtomTypeIDs = infoAtomTypeIDs
	usage.InfoAtomTypeCount = len(infoAtomTypeIDs)

	// 查询逻辑图使用情况
	graphConfigQuery := `SELECT graph_config_id FROM graph_config_tags WHERE tag_id = $1`
	var graphConfigIDs []string
	if err := m.dbConn.QueryRowsCtx(ctx, &graphConfigIDs, graphConfigQuery, tagID); err != nil && err != sqlx.ErrNotFound {
		logx.Errorf("查询标签在逻辑图中的使用情况失败: %v", err)
		return nil, fmt.Errorf("查询标签使用情况失败: %w", err)
	}
	usage.GraphConfigIDs = graphConfigIDs
	usage.GraphConfigCount = len(graphConfigIDs)

	return usage, nil
}

// DeleteTagWithCheck 删除标签前检查使用情况
func (m *tagManager) DeleteTagWithCheck(ctx context.Context, tagID string, tenantID string, force bool) error {
	// 查询标签使用情况
	usage, err := m.GetTagUsage(ctx, tagID, tenantID)
	if err != nil {
		return err
	}

	// 如果标签正在被使用
	isUsed := usage.InfoAtomTypeCount > 0 || usage.GraphConfigCount > 0
	if isUsed && !force {
		return fmt.Errorf("标签正在被使用：%d个信息原子类型，%d个逻辑图。请先解除关联或使用force参数强制删除",
			usage.InfoAtomTypeCount, usage.GraphConfigCount)
	}

	// 如果force=true且标签被使用，先删除所有关联
	if isUsed && force {
		// 删除信息原子类型关联
		if usage.InfoAtomTypeCount > 0 {
			deleteInfoAtomTagsSQL := `DELETE FROM info_atom_type_tags WHERE tag_id = $1`
			if _, err := m.dbConn.ExecCtx(ctx, deleteInfoAtomTagsSQL, tagID); err != nil {
				logx.Errorf("删除信息原子类型标签关联失败: %v", err)
				return fmt.Errorf("删除标签关联失败: %w", err)
			}
		}

		// 删除逻辑图关联
		if usage.GraphConfigCount > 0 {
			deleteGraphTagsSQL := `DELETE FROM graph_config_tags WHERE tag_id = $1`
			if _, err := m.dbConn.ExecCtx(ctx, deleteGraphTagsSQL, tagID); err != nil {
				logx.Errorf("删除逻辑图标签关联失败: %v", err)
				return fmt.Errorf("删除标签关联失败: %w", err)
			}
		}
	}

	// 删除标签本身
	return m.DeleteTag(ctx, tagID, tenantID)
}
