package devices

import (
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
)

func init() {
	// 注册AI算法盒类别
	// AI算法盒只需要在线检测配置，不需要业务数据字段映射
	core.RegisterCategory(
		core.DeviceCategoryInfo{
			Code:        core.CategoryAIBox,
			Name:        "AI算法盒",
			NameEn:      "AI Box",
			Description: "边缘计算AI算法盒，通过MQTT心跳监控在线状态",
			Icon:        "robot",
		},
		[]core.StandardField{
			// AI算法盒不需要业务字段，只需在线检测
			// is_online由在线检测配置自动管理，不在业务配置中映射
		},
	)
}
