package core

import "fmt"

// DeviceBinding 完整设备绑定信息
type DeviceBinding struct {
	ID             string             `json:"id"`             // 设备绑定记录ID（UUID）
	DeviceID       string             `json:"deviceId"`       // 设备唯一标识符
	DeviceName     string             `json:"deviceName"`     // 设备名称
	DeviceAlias    string             `json:"deviceAlias"`    // 设备别名
	DeviceModel    string             `json:"deviceModel"`    // 设备型号（关联iot_sensor_templates.model）
	DeviceCategory string             `json:"deviceCategory"` // 设备类别（冗余字段）

	// 设备基本信息
	Description      string `json:"description,omitempty"`      // 设备描述
	Location         string `json:"location,omitempty"`         // 安装位置
	InstallationDate string `json:"installationDate,omitempty"` // 安装日期（YYYY-MM-DD）

	// 设备状态
	Status   string `json:"status"`   // 设备状态（active/inactive/maintenance/error/decommissioned）
	IsOnline bool   `json:"isOnline"` // 当前在线状态（从Redis实时查询）

	// 时间追踪
	LastDataAt string `json:"lastDataAt,omitempty"` // 最后接收数据时间

	// 多租户和组织绑定
	TenantID string `json:"tenantId"` // 所属租户ID（UUID）
	OrgID    string `json:"orgId"`    // 所属组织ID（UUID）

	// 标签
	Tags []DeviceTagSummary `json:"tags,omitempty"` // 关联的标签列表

	// 审计字段
	CreatedBy string `json:"createdBy,omitempty"` // 创建者用户ID（UUID）
	UpdatedBy string `json:"updatedBy,omitempty"` // 最后修改者用户ID（UUID）
	CreatedAt string `json:"createdAt"`           // 创建时间
	UpdatedAt string `json:"updatedAt"`           // 更新时间
}

// DeviceBindingSummary 设备绑定摘要信息（用于列表展示）
type DeviceBindingSummary struct {
	ID             string             `json:"id"`             // 设备绑定记录ID
	DeviceID       string             `json:"deviceId"`       // 设备唯一标识符
	DeviceName     string             `json:"deviceName"`     // 设备名称
	DeviceAlias    string             `json:"deviceAlias"`    // 设备别名
	DeviceModel    string             `json:"deviceModel"`    // 设备型号
	DeviceCategory string             `json:"deviceCategory"` // 设备类别
	Location       string             `json:"location"`       // 安装位置
	Status         string             `json:"status"`         // 设备状态
	IsOnline       bool               `json:"isOnline"`       // 当前在线状态
	LastDataAt     string             `json:"lastDataAt"`     // 最后接收数据时间
	OrgID          string             `json:"orgId"`          // 所属组织ID
	OrgName        string             `json:"orgName"`        // 所属组织名称
	Tags           []DeviceTagSummary `json:"tags,omitempty"` // 关联的标签列表
	CreatedAt      string             `json:"createdAt"`      // 创建时间
}

// DeviceBindingMetadata 设备绑定元数据（用于插入PostgreSQL）
type DeviceBindingMetadata struct {
	DeviceID         string   `json:"deviceId"`                   // 设备唯一标识符（必填）
	DeviceName       string   `json:"deviceName"`                 // 设备名称（必填）
	DeviceAlias      string   `json:"deviceAlias,omitempty"`      // 设备别名
	DeviceModel      string   `json:"deviceModel"`                // 设备型号（必填，关联模板）
	Description      string   `json:"description,omitempty"`      // 设备描述
	Location         string   `json:"location,omitempty"`         // 安装位置
	InstallationDate string   `json:"installationDate,omitempty"` // 安装日期
	Status           string   `json:"status,omitempty"`           // 设备状态（默认active）
	TagIDs           []string `json:"tagIds,omitempty"`           // 标签ID列表（绑定时）
	TenantID         string   `json:"tenantId"`                   // 租户ID（UUID，必填）
	OrgID            string   `json:"orgId"`                      // 组织ID（UUID，必填）
	CreatedBy        string   `json:"createdBy,omitempty"`        // 创建者用户ID（UUID）
}

// DeviceBindingUpdate 设备绑定更新信息
type DeviceBindingUpdate struct {
	DeviceName       string   `json:"deviceName,omitempty"`       // 设备名称
	DeviceAlias      string   `json:"deviceAlias,omitempty"`      // 设备别名
	Description      string   `json:"description,omitempty"`      // 设备描述
	Location         string   `json:"location,omitempty"`         // 安装位置
	InstallationDate string   `json:"installationDate,omitempty"` // 安装日期
	Status           string   `json:"status,omitempty"`           // 设备状态
	TagIDs           []string `json:"tagIds,omitempty"`           // 更新标签列表（完全替换）
}

// DeviceBindingQuery 设备绑定查询条件
type DeviceBindingQuery struct {
	TenantID       string  `json:"tenantId"`                 // 租户ID（UUID，必填）
	OrgID          string  `json:"orgId,omitempty"`          // 组织ID筛选
	DeviceModel    string  `json:"deviceModel,omitempty"`    // 设备型号筛选
	DeviceCategory string  `json:"deviceCategory,omitempty"` // 设备类别筛选
	Status         string  `json:"status,omitempty"`         // 设备状态筛选
	IsOnline       *bool   `json:"isOnline,omitempty"`       // 在线状态筛选
	Keyword        string  `json:"keyword,omitempty"`        // 关键词搜索（设备ID、名称、别名）
	Page           int     `json:"page"`                     // 页码（从1开始）
	PageSize       int     `json:"pageSize"`                 // 每页数量
	OrderBy        string  `json:"orderBy,omitempty"`        // 排序字段
	OrderDir       string  `json:"orderDir,omitempty"`       // 排序方向（asc/desc）
}

// Validate 验证设备绑定元数据的有效性
func (metadata *DeviceBindingMetadata) Validate() error {
	// 1. 验证必填字段
	if metadata.DeviceID == "" {
		return fmt.Errorf("设备ID不能为空")
	}
	if metadata.DeviceName == "" {
		return fmt.Errorf("设备名称不能为空")
	}
	if metadata.DeviceModel == "" {
		return fmt.Errorf("设备型号不能为空")
	}
	if metadata.TenantID == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metadata.OrgID == "" {
		return fmt.Errorf("组织ID不能为空")
	}

	// 2. 验证设备状态
	if metadata.Status != "" {
		validStatuses := map[string]bool{
			"active":        true,
			"inactive":      true,
			"maintenance":   true,
			"error":         true,
			"decommissioned": true,
		}
		if !validStatuses[metadata.Status] {
			return fmt.Errorf("无效的设备状态: %s", metadata.Status)
		}
	}

	return nil
}
