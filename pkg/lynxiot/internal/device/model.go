package device

import "database/sql"

// 数据库查询结构体定义
// 说明：这些结构体仅用于数据库 Scan 操作，使用 sql.Null* 类型处理可空字段
// 转换为 core 层对象时，使用 helpers.go 中的转换函数

// templateInfoDB 模板信息（用于 Bind 方法查询设备型号模板）
type templateInfoDB struct {
	Category sql.NullString `db:"category"`
	TenantID string         `db:"tenant_id"`
}

// deviceBindingDB 完整设备绑定信息（用于 Get 方法）
type deviceBindingDB struct {
	ID               string         `db:"id"`
	DeviceID         string         `db:"device_id"`
	DeviceName       string         `db:"device_name"`
	DeviceAlias      string         `db:"device_alias"`
	DeviceModel      string         `db:"device_model"`
	DeviceCategory   sql.NullString `db:"device_category"`
	Description      sql.NullString `db:"description"`
	Location         sql.NullString `db:"location"`
	InstallationDate sql.NullString `db:"installation_date"`
	Status           string         `db:"status"`
	IsOnline         bool           `db:"is_online"`
	LastDataAt       sql.NullTime   `db:"last_data_at"`
	TenantID         string         `db:"tenant_id"`
	OrgID            string         `db:"org_id"`
	CreatedBy        sql.NullString `db:"created_by"`
	UpdatedBy        sql.NullString `db:"updated_by"`
	CreatedAt        sql.NullTime   `db:"created_at"`
	UpdatedAt        sql.NullTime   `db:"updated_at"`
}

// deviceBindingSummaryDB 设备绑定摘要信息（用于 List 方法）
type deviceBindingSummaryDB struct {
	ID             string         `db:"id"`
	DeviceID       string         `db:"device_id"`
	DeviceName     string         `db:"device_name"`
	DeviceAlias    string         `db:"device_alias"`
	DeviceModel    string         `db:"device_model"`
	DeviceCategory sql.NullString `db:"device_category"`
	Location       sql.NullString `db:"location"`
	Status         string         `db:"status"`
	IsOnline       bool           `db:"is_online"`
	LastDataAt     sql.NullTime   `db:"last_data_at"`
	OrgID          string         `db:"org_id"`
	OrgName        string         `db:"org_name"`
	CreatedAt      sql.NullTime   `db:"created_at"`
}

// deviceTagDB 设备标签信息（用于 getDeviceTags 方法）
type deviceTagDB struct {
	ID          string         `db:"id"`
	Label       string         `db:"label"`
	Description sql.NullString `db:"description"`
	Color       sql.NullString `db:"color"`
	CreatedAt   sql.NullTime   `db:"created_at"`
}

// boundDeviceDB 已绑定设备ID（用于 listBoundDeviceIDs 和 listBoundDeviceIDsForTenant 方法）
type boundDeviceDB struct {
	DeviceID string `db:"device_id"`
}

// deviceInfoForControllerDB 设备基本信息（用于 GetDeviceInfoForController 方法）
// 只包含Controller Manager需要的字段：category、model、tenant_id
type deviceInfoForControllerDB struct {
	DeviceID       string `db:"device_id"`
	DeviceModel    string `db:"device_model"`
	DeviceCategory string `db:"device_category"`
	TenantID       string `db:"tenant_id"`
}
