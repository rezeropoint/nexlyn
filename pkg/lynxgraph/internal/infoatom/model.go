package infoatom

import (
	"database/sql"

	"github.com/lib/pq"
)

// 数据库查询结构体定义
// 说明：这些结构体仅用于数据库 Scan 操作，使用 sql.Null* 类型处理可空字段
// 转换为 core 层对象时，使用 helpers.go 中的转换函数

// infoAtomTypeDB 完整信息原子类型信息（用于 Get 方法）
type infoAtomTypeDB struct {
	ID         string         `db:"id"`
	TenantId   string         `db:"tenant_id"`
	Name       string         `db:"name"`
	Version    string         `db:"version"`
	TagIDs     pq.StringArray `db:"tag_ids"`
	DataFormat string         `db:"data_format"`
	CreatedBy  sql.NullString `db:"created_by"`
	UpdatedBy  sql.NullString `db:"updated_by"`
	CreatedAt  int64          `db:"created_at"`
	UpdatedAt  int64          `db:"updated_at"`
}

// infoAtomTypeSummaryDB 信息原子类型摘要信息（用于 List 方法）
// 与完整结构相同，因为信息原子类型字段较少
type infoAtomTypeSummaryDB = infoAtomTypeDB
