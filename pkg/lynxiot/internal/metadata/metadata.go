package metadata

import (
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
)

// Manager 元数据管理器接口
type Manager interface {
	// 获取所有设备类别
	GetDeviceCategories() []core.DeviceCategoryInfo

	// 获取指定类别的标准字段
	GetStandardFields(category core.DeviceCategory) []core.StandardField
}

// NewManager 创建元数据管理器
func NewManager() Manager {
	return newMetadataManager()
}
