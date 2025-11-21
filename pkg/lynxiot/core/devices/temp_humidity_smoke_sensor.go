package devices

import (
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core/fields"
)

func init() {
	// 注册温湿度烟雾传感器类别
	// 使用预定义的标准字段，确保字段定义统一且与ClickHouse表结构一致
	core.RegisterCategory(
		core.DeviceCategoryInfo{
			Code:        core.CategoryTempHumiditySmokeSersor,
			Name:        "温湿度烟雾传感器",
			NameEn:      "Temperature Humidity Smoke Sensor",
			Description: "监测环境温度、湿度和烟雾浓度的复合传感器",
			Icon:        "dashboard",
		},
		[]core.StandardField{
			// 从fields注册表获取标准字段
			*fields.GetFieldByName(fields.FieldNameTemperature),
			*fields.GetFieldByName(fields.FieldNameHumidity),
			*fields.GetFieldByName(fields.FieldNameSmoke),
		},
	)
}
