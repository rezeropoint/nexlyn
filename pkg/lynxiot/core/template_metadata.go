// Package core 定义IoT设备管理的核心领域模型
//
// 文件说明：template_metadata.go
// 职责：模板元数据结构（PostgreSQL存储）
package core

// TemplateMetadata 模板元数据（PostgreSQL存储）
type TemplateMetadata struct {
	Model                 string   `json:"model"`                           // 设备型号（唯一标识）
	Name                  string   `json:"name"`                            // 模板名称
	Category              string   `json:"category"`                        // 设备类别
	Manufacturer          string   `json:"manufacturer"`                    // 设备厂商
	Description           string   `json:"description"`                     // 描述
	Version               string   `json:"version"`                         // 版本号
	Enabled               bool     `json:"enabled"`                         // 是否启用
	OnlineTopicSuffixes   []string `json:"onlineTopicSuffixes,omitempty"`   // 在线检测主题后缀列表（内部管理，前端不传）
	BusinessTopicSuffixes []string `json:"businessTopicSuffixes,omitempty"` // 业务数据主题后缀列表（预留;内部管理，前端不传）
	ControlTopicSuffixes  []string `json:"controlTopicSuffixes,omitempty"`  // 控制主题后缀列表（内部管理，前端不传）
	TagIDs                []string `json:"tagIds,omitempty"`                // 标签ID列表（创建/更新时）
	TenantID              string   `json:"tenantId"`                        // 租户ID（UUID）
	CreatedBy             string   `json:"createdBy"`                       // 创建者ID（UUID）
}

// Template 完整模板信息（元数据 + 配置）
type Template struct {
	ID           string             `json:"id"`             // 模板ID
	Model        string             `json:"model"`          // 设备型号
	Name         string             `json:"name"`           // 模板名称
	Category     string             `json:"category"`       // 设备类别
	Manufacturer string             `json:"manufacturer"`   // 设备厂商
	Description  string             `json:"description"`    // 描述
	Version      string             `json:"version"`        // 版本号
	Enabled      bool               `json:"enabled"`        // 是否启用
	DeviceCount  int                `json:"deviceCount"`    // 关联设备数量
	Tags         []DeviceTagSummary `json:"tags,omitempty"` // 关联的标签列表
	TenantID     string             `json:"tenantId"`       // 租户ID（UUID）
	CreatedBy    string             `json:"createdBy"`      // 创建者ID（UUID）
	CreatedAt    string             `json:"createdAt"`      // 创建时间
	UpdatedAt    string             `json:"updatedAt"`      // 更新时间

	// 配置信息（从Etcd查询）
	OnlineConfig   *OnlineDetectionConfig `json:"onlineConfig,omitempty"`   // 在线检测配置
	BusinessConfig *DataProcessingConfig  `json:"businessConfig,omitempty"` // 业务数据处理配置
	ControlConfig  *DeviceControlConfig   `json:"controlConfig,omitempty"`  // 控制配置
}

// TemplateSummary 模板摘要信息（用于列表展示）
type TemplateSummary struct {
	ID           string             `json:"id"`             // 模板ID
	Model        string             `json:"model"`          // 设备型号
	Name         string             `json:"name"`           // 模板名称
	Category     string             `json:"category"`       // 设备类别
	Manufacturer string             `json:"manufacturer"`   // 设备厂商
	Version      string             `json:"version"`        // 版本号
	Enabled      bool               `json:"enabled"`        // 是否启用
	DeviceCount  int                `json:"deviceCount"`    // 关联设备数量
	Tags         []DeviceTagSummary `json:"tags,omitempty"` // 关联的标签列表
	CreatedAt    string             `json:"createdAt"`      // 创建时间
	UpdatedAt    string             `json:"updatedAt"`      // 更新时间
}

// TemplateQuery 模板查询条件
type TemplateQuery struct {
	TenantID string `json:"tenantId"`           // 租户ID（UUID，必填）
	Category string `json:"category,omitempty"` // 设备类别筛选
	Enabled  *bool  `json:"enabled,omitempty"`  // 启用状态筛选
	Keyword  string `json:"keyword,omitempty"`  // 关键词搜索（模型或名称）
	Page     int    `json:"page"`               // 页码（从1开始）
	PageSize int    `json:"pageSize"`           // 每页数量
	OrderBy  string `json:"orderBy,omitempty"`  // 排序字段
	OrderDir string `json:"orderDir,omitempty"` // 排序方向（asc/desc）
}
