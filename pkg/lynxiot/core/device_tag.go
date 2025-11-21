package core

import "fmt"

// DeviceTag 设备标签定义
type DeviceTag struct {
	ID          string `json:"id"`          // 标签ID（UUID）
	Label       string `json:"label"`       // 标签名称
	Description string `json:"description"` // 标签描述
	Color       string `json:"color"`       // 标签颜色（用于前端展示）
	TenantID    string `json:"tenantId"`    // 租户ID（UUID）
	CreatedBy   string `json:"createdBy"`   // 创建者用户ID（UUID）
	UpdatedBy   string `json:"updatedBy"`   // 最后修改者用户ID（UUID）
	CreatedAt   string `json:"createdAt"`   // 创建时间
	UpdatedAt   string `json:"updatedAt"`   // 更新时间
}

// DeviceTagSummary 设备标签摘要（用于列表展示）
type DeviceTagSummary struct {
	ID          string `json:"id"`          // 标签ID
	Label       string `json:"label"`       // 标签名称
	Description string `json:"description"` // 标签描述
	Color       string `json:"color"`       // 标签颜色
	CreatedAt   string `json:"createdAt"`   // 创建时间
}

// DeviceTagMetadata 设备标签元数据（用于创建标签）
type DeviceTagMetadata struct {
	Label       string `json:"label"`                // 标签名称（必填）
	Description string `json:"description,omitempty"` // 标签描述
	Color       string `json:"color,omitempty"`      // 标签颜色（如：#1890ff）
	TenantID    string `json:"tenantId"`             // 租户ID（UUID，必填）
	CreatedBy   string `json:"createdBy,omitempty"`  // 创建者用户ID（UUID）
}

// DeviceTagUpdate 设备标签更新信息
type DeviceTagUpdate struct {
	Label       string `json:"label,omitempty"`       // 标签名称
	Description string `json:"description,omitempty"` // 标签描述
	Color       string `json:"color,omitempty"`       // 标签颜色
}

// DeviceTagQuery 设备标签查询条件
type DeviceTagQuery struct {
	TenantID string `json:"tenantId"`          // 租户ID（UUID，必填）
	Keyword  string `json:"keyword,omitempty"` // 关键词搜索（标签名称或描述）
	Page     int    `json:"page"`              // 页码（从1开始）
	PageSize int    `json:"pageSize"`          // 每页数量
}

// Validate 验证设备标签元数据的有效性
func (metadata *DeviceTagMetadata) Validate() error {
	if metadata.Label == "" {
		return fmt.Errorf("标签名称不能为空")
	}
	if metadata.TenantID == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	return nil
}
