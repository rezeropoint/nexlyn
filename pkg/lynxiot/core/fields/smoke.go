package fields

import "github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

func init() {
	// 注册烟雾浓度字段
	RegisterField(core.StandardField{
		Name:        "smoke",
		DisplayName: "烟雾浓度",
		FieldType:   core.FieldTypeNumber,
		Unit:        "ppm",
		Description: "烟雾浓度值",
		Required:    true, // 温湿度烟雾传感器的核心字段，必填
		MinValue:    core.Float64Ptr(0),
	})
}
