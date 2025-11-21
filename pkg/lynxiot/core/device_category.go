// Package core 定义IoT设备管理的核心领域模型
//
// 文件说明：device_category.go
// 职责：设备类别类型定义和元信息
package core

// DeviceCategory 设备类别类型
type DeviceCategory string

const (
	CategoryTempHumiditySmokeSersor DeviceCategory = "temp_humidity_smoke_sensor" // 温湿度烟雾传感器
	CategoryAIBox                   DeviceCategory = "ai_box"                     // AI算法盒
)

// DeviceCategoryInfo 设备类别元信息
type DeviceCategoryInfo struct {
	Code        DeviceCategory `json:"code"`        // 类别代码
	Name        string         `json:"name"`        // 中文名称
	NameEn      string         `json:"nameEn"`      // 英文名称
	Description string         `json:"description"` // 描述
	Icon        string         `json:"icon"`        // 图标名称（用于前端展示）
}

// StandardField 标准字段定义（增强版）
type StandardField struct {
	Name        string    `json:"name"`                  // 字段名（如：temperature）
	DisplayName string    `json:"displayName"`           // 显示名称（如：温度）
	FieldType   FieldType `json:"fieldType"`             // 字段类型
	Unit        string    `json:"unit,omitempty"`        // 单位（如：℃, %, ppm）
	Description string    `json:"description,omitempty"` // 字段描述
	Required    bool      `json:"required"`              // 是否必填
	MinValue    *float64  `json:"minValue,omitempty"`    // 最小值（数值类型）
	MaxValue    *float64  `json:"maxValue,omitempty"`    // 最大值（数值类型）
	EnumValues  []string  `json:"enumValues,omitempty"`  // 枚举值（枚举类型）
	TTLDays     *int      `json:"ttlDays,omitempty"`     // 数据保存时长（天），仅用于结构体定义，实际TTL由配置文件控制
}

// Float64Ptr 辅助函数：创建float64指针
func Float64Ptr(v float64) *float64 {
	return &v
}

// IntPtr 辅助函数：创建int指针
func IntPtr(v int) *int {
	return &v
}
