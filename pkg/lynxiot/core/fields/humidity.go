package fields

import "github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

func init() {
	// 注册湿度字段
	RegisterField(core.StandardField{
		Name:        "humidity",
		DisplayName: "湿度",
		FieldType:   core.FieldTypeNumber,
		Unit:        "%",
		Description: "环境湿度",
		Required:    true, // 温湿度传感器的核心字段，必填
		MinValue:    core.Float64Ptr(0),
		MaxValue:    core.Float64Ptr(100),
	})
}
