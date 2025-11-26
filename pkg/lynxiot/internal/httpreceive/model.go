package httpreceive

import (
	"database/sql"
)

// httpReceiveDB 完整HTTP接收配置信息（用于 Get 方法）
type httpReceiveDB struct {
	ID          string         `db:"id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Enabled     bool           `db:"enabled"`
	TenantID    string         `db:"tenant_id"`
	CreatedBy   sql.NullString `db:"created_by"`
	CreatedAt   sql.NullTime   `db:"created_at"`
	UpdatedAt   sql.NullTime   `db:"updated_at"`
}

// httpReceiveSummaryDB HTTP接收配置摘要信息（用于 List 方法）
type httpReceiveSummaryDB struct {
	ID          string         `db:"id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Enabled     bool           `db:"enabled"`
	CreatedAt   sql.NullTime   `db:"created_at"`
	UpdatedAt   sql.NullTime   `db:"updated_at"`
}

// SQL 语句常量
const (
	insertHttpReceiveSQL = `
		INSERT INTO iot_http_receive_configs (id, name, description, enabled, tenant_id, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	updateHttpReceiveSQL = `
		UPDATE iot_http_receive_configs
		SET name = $1, description = $2, enabled = $3, updated_at = NOW()
		WHERE id = $4 AND tenant_id = $5
	`

	deleteHttpReceiveSQL = `
		DELETE FROM iot_http_receive_configs
		WHERE id = $1 AND tenant_id = $2
	`

	getHttpReceiveSQL = `
		SELECT id, name, description, enabled, tenant_id, created_by, created_at, updated_at
		FROM iot_http_receive_configs
		WHERE id = $1 AND tenant_id = $2
	`

	listHttpReceiveSQL = `
		SELECT id, name, description, enabled, created_at, updated_at
		FROM iot_http_receive_configs
		WHERE tenant_id = $1
		%s
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	countHttpReceiveSQL = `
		SELECT COUNT(*)
		FROM iot_http_receive_configs
		WHERE tenant_id = $1
		%s
	`

	checkHttpReceiveExistsSQL = `
		SELECT 1 FROM iot_http_receive_configs WHERE id = $1 AND tenant_id = $2
	`
)
