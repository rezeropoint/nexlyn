// Package core 定义IoT设备管理的核心领域模型
//
// 文件说明：storage.go
// 职责：存储相关的领域模型和函数类型定义
package core

import (
	"context"
	"fmt"
	"time"
)

// ClickHouse 表名相关常量
const (
	// ClickHouseTablePrefix ClickHouse表名前缀
	ClickHouseTablePrefix = "iot_"
	// ClickHouseTableSuffix ClickHouse表名后缀
	ClickHouseTableSuffix = "_data"
)

// FieldType 常量定义
const (
	FieldTypeString      FieldType = "string"
	FieldTypeNumber      FieldType = "number"      // 数值类型（统一处理整数和浮点数）
	FieldTypeBoolean     FieldType = "boolean"     // 布尔类型
	FieldTypeTimestamp   FieldType = "timestamp"   // 时间戳
	FieldTypeDatetime    FieldType = "datetime"    // 日期时间
	FieldTypeImage       FieldType = "imageURL"    // 图片URL
	FieldTypeImageBase64 FieldType = "imageBase64" // 图片Base64
)

// SensorDataRecord 传感器数据记录（领域模型）
// 用于在MQTT业务数据处理和Storage Manager之间传递数据
type SensorDataRecord struct {
	Timestamp      time.Time         `json:"timestamp"`      // 数据采集时间戳
	DeviceID       string            `json:"deviceId"`       // 设备唯一标识符（从MQTT主题解析）
	DeviceModel    string            `json:"deviceModel"`    // 设备型号
	DeviceCategory DeviceCategory    `json:"deviceCategory"` // 设备类别
	TenantID       string            `json:"tenantId"`       // 租户ID
	Fields         map[FieldName]any `json:"fields"`         // 提取的字段（key: 标准字段名, value: 字段值）
}

// StoreSensorDataFunc 存储传感器数据的函数类型
// 用于解耦 storage 和 mqtt 包（避免循环依赖）
// 由 Storage Manager 实现，注入到 MQTT Manager
type StoreSensorDataFunc func(ctx context.Context, record *SensorDataRecord) error

// BuildClickHouseTableName 根据字段名构建ClickHouse表名
// 格式: iot_{field_name}_data
// 例如：temperature -> iot_temperature_data
func BuildClickHouseTableName(fieldName FieldName) string {
	return fmt.Sprintf("%s%s%s", ClickHouseTablePrefix, string(fieldName), ClickHouseTableSuffix)
}
