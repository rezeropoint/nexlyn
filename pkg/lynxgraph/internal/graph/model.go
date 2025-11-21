package graph

import (
	"database/sql"

	"github.com/lib/pq"
)

// 数据库查询结构体定义
// 说明：这些结构体仅用于数据库 Scan 操作，使用 sql.Null* 类型处理可空字段
// 转换为 core 层对象时，使用 helpers.go 中的转换函数

// graphConfigBasicDB PostgreSQL中存储的GraphConfig基本信息（用于 List 方法）
type graphConfigBasicDB struct {
	ID          string         `db:"id"`
	TenantId    string         `db:"tenant_id"`
	OrgID       string         `db:"org_id"`
	Name        string         `db:"name"`
	Version     string         `db:"version"`
	Description string         `db:"description"`
	TagIDs      pq.StringArray `db:"tag_ids"`
	Enable      bool           `db:"enable"`
	Icon        sql.NullString `db:"icon"`
	IconColor   sql.NullString `db:"icon_color"`
	CreatedBy   sql.NullString `db:"created_by"`
	UpdatedBy   sql.NullString `db:"updated_by"`
	CreatedAt   int64          `db:"created_at"`
	UpdatedAt   int64          `db:"updated_at"`
}

// graphConfigDetail MongoDB中存储的GraphConfig详细信息
//
// 设计原则：
//   - 使用 MongoDB 专用结构体（graphNodeConfigMongo、graphEdgeConfigMongo）
//   - 不直接使用 core 层类型，避免 Core 层被存储层框架约束
//   - 复杂字段（map[string]any）序列化为 JSON 字符串存储
type graphConfigDetail struct {
	ID       string                   `json:"id" bson:"id"`
	TenantId string                   `json:"tenantId" bson:"tenantId"`
	OrgID    string                   `json:"orgId" bson:"orgId"`
	Name     string                   `json:"name" bson:"name"`
	Version  string                   `json:"version" bson:"version"`
	Nodes    []graphNodeConfigMongo   `json:"nodes" bson:"nodes"`
	Edges    []graphEdgeConfigMongo   `json:"edges" bson:"edges"`
}

// graphNodeConfigMongo MongoDB专用的节点配置结构体
//
// 设计说明：
//   - 独立定义MongoDB存储结构，不依赖 core.NodeConfig
//   - BlockConfig 和 SubscribedLabels 存储为 map[string]any（BSON 原生支持）
type graphNodeConfigMongo struct {
	ID                        string            `json:"id" bson:"id"`
	Type                      string            `json:"type" bson:"type"`                                   // 存储为字符串
	BlockType                 string            `json:"blockType" bson:"blockType"`                         // 存储为字符串
	BlockVersion              string            `json:"blockVersion" bson:"blockVersion"`
	BlockConfig               map[string]any    `json:"blockConfig" bson:"blockConfig"`                     // MongoDB 原生支持 map
	IsEntryPoint              bool              `json:"isEntryPoint" bson:"isEntryPoint"`
	SubscribedInfoAtomTypeIDs []string          `json:"subscribed_infoatom_typeids" bson:"subscribed_infoatom_typeids"`
	SubscribedSource          string            `json:"subscribed_source" bson:"subscribed_source"`
	SubscribedLabels          map[string]string `json:"subscribed_labels" bson:"subscribed_labels"`         // MongoDB 原生支持 map
	X                         float64           `json:"x,omitempty" bson:"x,omitempty"`
	Y                         float64           `json:"y,omitempty" bson:"y,omitempty"`
}

// graphEdgeConfigMongo MongoDB专用的边配置结构体
type graphEdgeConfigMongo struct {
	ID        string `json:"id" bson:"id"`
	SourceID  string `json:"sourceId" bson:"sourceId"`
	TargetID  string `json:"targetId" bson:"targetId"`
	Condition string `json:"condition" bson:"condition"`
}
