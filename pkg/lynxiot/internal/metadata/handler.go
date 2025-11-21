package metadata

import (
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
)

// metadataManager 元数据管理器实现
type metadataManager struct{}

// newMetadataManager 创建元数据管理器实例
func newMetadataManager() *metadataManager {
	return &metadataManager{}
}

// GetDeviceCategories 获取所有设备类别
func (m *metadataManager) GetDeviceCategories() []core.DeviceCategoryInfo {
	categories := make([]core.DeviceCategoryInfo, 0, len(core.GetDeviceCategoryRegistry()))

	// 按固定顺序返回
	for _, code := range core.GetOrderedCategories() {
		if info, exists := core.GetDeviceCategoryRegistry()[code]; exists {
			categories = append(categories, info)
		}
	}

	return categories
}

// GetStandardFields 获取指定类别的标准字段
func (m *metadataManager) GetStandardFields(category core.DeviceCategory) []core.StandardField {
	if fields, exists := core.GetStandardFieldsRegistry()[category]; exists {
		return fields
	}
	return []core.StandardField{}
}
