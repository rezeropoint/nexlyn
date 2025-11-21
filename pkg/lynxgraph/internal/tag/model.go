package tag

import "database/sql"

// tagDB 数据库查询结构体（用于Scan操作）
// 说明：使用sql.Null*类型处理可空字段，转换为core层对象时使用helper函数
type tagDB struct {
	ID          string         `db:"id"`
	TenantId    string         `db:"tenant_id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Scope       string         `db:"scope"`
	CreatedBy   sql.NullString `db:"created_by"`
	UpdatedBy   sql.NullString `db:"updated_by"`
	CreatedAt   int64          `db:"created_at"`
	UpdatedAt   int64          `db:"updated_at"`
}

// tagSummaryDB 标签摘要数据库查询结构体
type tagSummaryDB struct {
	ID          string         `db:"id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Scope       string         `db:"scope"`
	CreatedAt   int64          `db:"created_at"`
	UpdatedAt   int64          `db:"updated_at"`
}
