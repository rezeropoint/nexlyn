package tag

import (
	"database/sql"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
)

// nullStringToString 将sql.NullString转换为普通字符串
//
// 这是纯函数辅助方法（无 receiver），符合 DEVELOPMENT.md 的 helpers.go 规范
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// tagDBToLynxTag 将数据库查询结果转换为LynxTag
//
// 这是纯函数辅助方法（无 receiver），符合 DEVELOPMENT.md 的 helpers.go 规范
func tagDBToLynxTag(t *tagDB) *core.LynxTag {
	return &core.LynxTag{
		ID:          t.ID,
		Name:        t.Name,
		Description: nullStringToString(t.Description),
		Scope:       t.Scope,
		TenantID:    t.TenantId,
		CreatedBy:   nullStringToString(t.CreatedBy),
		UpdatedBy:   nullStringToString(t.UpdatedBy),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// tagSummaryDBToLynxTagSummary 将数据库查询结果转换为LynxTagSummary
//
// 这是纯函数辅助方法（无 receiver），符合 DEVELOPMENT.md 的 helpers.go 规范
func tagSummaryDBToLynxTagSummary(t *tagSummaryDB) *core.LynxTagSummary {
	return &core.LynxTagSummary{
		ID:          t.ID,
		Name:        t.Name,
		Description: nullStringToString(t.Description),
		Scope:       t.Scope,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
