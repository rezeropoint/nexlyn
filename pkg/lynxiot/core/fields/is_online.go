package fields

import "github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

func init() {
	// 注册在线状态字段
	// 说明：
	// - Required: false - 不强制业务配置映射（由在线检测配置自动管理）
	// - Reserved: true - 系统保留字段，禁止业务配置手动映射
	RegisterField(core.StandardField{
		Name:        "is_online",
		DisplayName: "在线状态",
		FieldType:   core.FieldTypeBoolean,
		Description: "设备是否在线（由在线检测配置自动管理）",
		Required:    false, // 不强制业务配置映射
	})
}
