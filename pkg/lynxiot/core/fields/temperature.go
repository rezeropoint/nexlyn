package fields

import "github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

func init() {
	// 注册温度字段
	RegisterField(core.StandardField{
		Name:        "temperature",
		DisplayName: "温度",
		FieldType:   core.FieldTypeNumber,
		Unit:        "℃",
		Description: "环境温度",
		Required:    true, // 温湿度传感器的核心字段，必填
	})
}
