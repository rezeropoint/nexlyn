package template

import (
	"database/sql"

	"github.com/lib/pq"
)

// 数据库查询结构体定义
// 说明：这些结构体仅用于数据库 Scan 操作，使用 sql.Null* 类型处理可空字段

// oldTemplateDB 旧模板信息（用于 Update 方法查询旧模板配置）
type oldTemplateDB struct {
	TenantID              string         `db:"tenant_id"`
	Model                 string         `db:"model"`
	Category              string         `db:"category"`
	OnlineTopicSuffixes   pq.StringArray `db:"online_topic_suffixes"`
	BusinessTopicSuffixes pq.StringArray `db:"business_topic_suffixes"`
	ControlTopicSuffixes  pq.StringArray `db:"control_topic_suffixes"`
}

// templateInfoDB 模板信息（用于 Delete 方法查询设备计数和主题后缀）
type templateInfoDB struct {
	DeviceCount           int            `db:"device_count"`
	Model                 string         `db:"model"`
	Category              string         `db:"category"`
	OnlineTopicSuffixes   pq.StringArray `db:"online_topic_suffixes"`
	BusinessTopicSuffixes pq.StringArray `db:"business_topic_suffixes"`
	ControlTopicSuffixes  pq.StringArray `db:"control_topic_suffixes"`
}

// templateDB 完整模板信息（用于 Get 方法）
type templateDB struct {
	ID                    string         `db:"id"`
	Model                 string         `db:"model"`
	Name                  string         `db:"name"`
	Category              string         `db:"category"`
	Manufacturer          string         `db:"manufacturer"`
	Description           string         `db:"description"`
	Version               string         `db:"version"`
	Enabled               bool           `db:"enabled"`
	DeviceCount           int            `db:"device_count"`
	TenantID              string         `db:"tenant_id"`
	CreatedBy             string         `db:"created_by"`
	CreatedAt             sql.NullTime   `db:"created_at"`
	UpdatedAt             sql.NullTime   `db:"updated_at"`
	OnlineTopicSuffixes   pq.StringArray `db:"online_topic_suffixes"`
	BusinessTopicSuffixes pq.StringArray `db:"business_topic_suffixes"`
	ControlTopicSuffixes  pq.StringArray `db:"control_topic_suffixes"`
}

// templateSummaryDB 模板摘要信息（用于 List 方法）
type templateSummaryDB struct {
	ID           string       `db:"id"`
	Model        string       `db:"model"`
	Name         string       `db:"name"`
	Category     string       `db:"category"`
	Manufacturer string       `db:"manufacturer"`
	Version      string       `db:"version"`
	Enabled      bool         `db:"enabled"`
	DeviceCount  int          `db:"device_count"`
	CreatedAt    sql.NullTime `db:"created_at"`
	UpdatedAt    sql.NullTime `db:"updated_at"`
}

// templateTagDB 模板标签信息（用于 getTemplateTags 方法）
type templateTagDB struct {
	ID          string         `db:"id"`
	Label       string         `db:"label"`
	Description sql.NullString `db:"description"`
	Color       sql.NullString `db:"color"`
	CreatedAt   sql.NullTime   `db:"created_at"`
}
