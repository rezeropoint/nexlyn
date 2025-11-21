package platform

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// checkPlatformInUse 检查平台配置是否被设备模板使用（非事务版本）
// 返回：使用该平台的模板信息列表，如果没有被使用则返回空列表
func (m *platformManager) checkPlatformInUse(ctx context.Context, platformID string) ([]string, error) {
	// 使用 SQL 查询检查平台配置是否在 used_platform_ids 数组中
	// PostgreSQL 数组查询：$1 = ANY(used_platform_ids)
	query := `
		SELECT model, category
		FROM iot_sensor_templates
		WHERE $1 = ANY(used_platform_ids) AND deleted_at IS NULL
	`

	var results []struct {
		Model    string `db:"model"`
		Category string `db:"category"`
	}

	err := m.dbConn.QueryRowsCtx(ctx, &results, query, platformID)
	if err != nil {
		return nil, fmt.Errorf("查询使用平台配置的模板失败: %w", err)
	}

	// 构建模板信息列表
	usedByTemplates := make([]string, 0, len(results))
	for _, r := range results {
		templateInfo := fmt.Sprintf("%s/%s", r.Category, r.Model)
		usedByTemplates = append(usedByTemplates, templateInfo)
	}

	return usedByTemplates, nil
}

// checkPlatformInUseWithSession 检查平台配置是否被设备模板使用（事务版本）
// 在事务中执行查询，供Delete方法使用
func (m *platformManager) checkPlatformInUseWithSession(ctx context.Context, session sqlx.Session, platformID string) ([]string, error) {
	query := `
		SELECT model, category
		FROM iot_sensor_templates
		WHERE $1 = ANY(used_platform_ids) AND deleted_at IS NULL
	`

	var results []struct {
		Model    string `db:"model"`
		Category string `db:"category"`
	}

	err := session.QueryRowsCtx(ctx, &results, query, platformID)
	if err != nil {
		return nil, fmt.Errorf("查询使用平台配置的模板失败: %w", err)
	}

	// 构建模板信息列表
	usedByTemplates := make([]string, 0, len(results))
	for _, r := range results {
		templateInfo := fmt.Sprintf("%s/%s", r.Category, r.Model)
		usedByTemplates = append(usedByTemplates, templateInfo)
	}

	return usedByTemplates, nil
}
