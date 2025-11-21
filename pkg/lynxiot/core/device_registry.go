package core

// 设备类别注册表（对 internal/metadata 包可见）
var deviceCategoryRegistry = make(map[DeviceCategory]DeviceCategoryInfo)

// 标准字段注册表（对 internal/metadata 包可见）
var standardFieldsRegistry = make(map[DeviceCategory][]StandardField)

// 设备类别顺序（用于排序输出，对 internal/metadata 包可见）
var orderedCategories []DeviceCategory

// RegisterCategory 注册设备类别
func RegisterCategory(info DeviceCategoryInfo, fields []StandardField) {
	deviceCategoryRegistry[info.Code] = info
	standardFieldsRegistry[info.Code] = fields
	orderedCategories = append(orderedCategories, info.Code)
}

// GetDeviceCategoryRegistry 获取设备类别注册表（供 internal/metadata 使用）
func GetDeviceCategoryRegistry() map[DeviceCategory]DeviceCategoryInfo {
	return deviceCategoryRegistry
}

// GetStandardFieldsRegistry 获取标准字段注册表（供 internal/metadata 使用）
func GetStandardFieldsRegistry() map[DeviceCategory][]StandardField {
	return standardFieldsRegistry
}

// GetOrderedCategories 获取设备类别顺序（供 internal/metadata 使用）
func GetOrderedCategories() []DeviceCategory {
	return orderedCategories
}

// IsValidDeviceCategory 检查设备类别代码是否有效（用于内部验证）
func IsValidDeviceCategory(code string) bool {
	_, exists := deviceCategoryRegistry[DeviceCategory(code)]
	return exists
}

// GetStandardFieldsByCategory 根据设备类别获取标准字段列表（用于内部验证）
func GetStandardFieldsByCategory(category DeviceCategory) []StandardField {
	if fields, exists := standardFieldsRegistry[category]; exists {
		return fields
	}
	return []StandardField{}
}
