package auth

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// GetAllChildOrgIds 查询组织及其所有子组织的ID列表
// 基于PostgreSQL的path字段（Path Enumeration Pattern）
// 参数:
//   - ctx: 上下文
//   - dbConn: 数据库连接
//   - orgId: 组织ID
//   - tenantId: 租户ID
//
// 返回值:
//   - []string: 组织ID列表（包含传入的orgId及其所有子组织）
//   - error: 错误信息
func GetAllChildOrgIds(ctx context.Context, dbConn sqlx.SqlConn, orgId string, tenantId string) ([]string, error) {
	// 1. 查询组织的path（用于构建子组织查询条件）
	// 说明：go-zero的QueryRowCtx将sql.NullString视为结构体映射，单列扫描会触发
	// "not matching destination to scan"。这里使用COALESCE并扫描到string以避免该问题。
	var orgPath string
	getPathQuery := `
        SELECT COALESCE(path, '')
        FROM system_organizations
        WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
    `
	err := dbConn.QueryRowCtx(ctx, &orgPath, getPathQuery, orgId, tenantId)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果组织不存在，返回空列表
			return []string{}, nil
		}
		return nil, fmt.Errorf("查询组织path失败: %v", err)
	}

	// 如果path为空（不应该发生），返回组织ID本身
	if orgPath == "" {
		return []string{orgId}, nil
	}

	// 2. 查询所有子组织（path以当前组织path开头的所有组织）
	// 注意：path格式是 "/root/" 或 "/root/child1/" 或 "/root/child1/child2/"
	// 所以子组织的path会以父组织path开头，例如：
	//   父: /root/              子: /root/child1/、/root/child1/child2/
	//   父: /root/child1/       子: /root/child1/child2/、/root/child1/child2/child3/
	getAllChildrenQuery := `
		SELECT id
		FROM system_organizations
		WHERE tenant_id = $1
			AND deleted_at IS NULL
			AND (path LIKE $2 || '%' OR path = $2)
		ORDER BY path
	`

	// 使用临时结构体接收查询结果
	type orgResult struct {
		Id string `db:"id"`
	}

	var orgResults []orgResult
	err = dbConn.QueryRowsCtx(ctx, &orgResults, getAllChildrenQuery, tenantId, orgPath)
	if err != nil {
		return nil, fmt.Errorf("查询子组织失败: %v", err)
	}

	// 3. 收集所有组织ID
	orgIds := make([]string, 0, len(orgResults))
	for _, result := range orgResults {
		orgIds = append(orgIds, result.Id)
	}

	// 如果没有找到任何组织（理论上不应该发生，因为至少应该包含传入的组织本身）
	// 返回传入的orgId以确保至少包含用户自己的组织
	if len(orgIds) == 0 {
		return []string{orgId}, nil
	}

	return orgIds, nil
}

// ValidateOrgAccess 验证用户是否有权访问指定组织（检查组织是否在用户权限范围内）
// 权限逻辑：用户可以访问自己所属组织及其所有下级组织的数据
//
// 参数:
//   - ctx: 上下文
//   - dbConn: 数据库连接
//   - targetOrgId: 目标组织ID（用户想要访问的组织）
//   - userOrgId: 用户所属组织ID
//   - tenantId: 租户ID
//
// 返回:
//   - bool: true=有权限访问, false=无权限
//   - error: 错误信息
func ValidateOrgAccess(ctx context.Context, dbConn sqlx.SqlConn, targetOrgId, userOrgId, tenantId string) (bool, error) {
	// 查询用户组织及所有子组织（用户有权访问的组织范围）
	userOrgIds, err := GetAllChildOrgIds(ctx, dbConn, userOrgId, tenantId)
	if err != nil {
		return false, fmt.Errorf("查询用户组织失败: %v", err)
	}

	// 检查目标组织是否在用户权限范围内
	for _, orgId := range userOrgIds {
		if orgId == targetOrgId {
			return true, nil
		}
	}

	return false, nil
}
