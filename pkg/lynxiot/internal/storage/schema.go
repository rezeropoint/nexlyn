package storage

import (
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
)

// ClickHouse表结构定义
// 每个表独立定义为常量，方便单独维护和扩展
// 新增表时：1) 在此文件添加常量 2) 在initializeClickHouse函数中添加createTableIfNotExists调用
//
// 注意：从动态TTL实现开始，优先使用generateCreateTableSQL函数动态生成建表SQL，
// 静态常量schema保留作为备用和兼容性参考

// schemaTemperatureData 温度数据表结构
const schemaTemperatureData = `
CREATE TABLE IF NOT EXISTS nexlyn.iot_temperature_data (
	timestamp DateTime64(3) COMMENT '数据采集时间戳（毫秒精度）',
	device_id String COMMENT '设备唯一标识符',
	device_model String COMMENT '设备型号',
	device_category String COMMENT '设备类别',
	tenant_id String COMMENT '租户ID',
	value Float64 COMMENT '温度值（℃）',
	INDEX idx_device_id device_id TYPE bloom_filter(0.01) GRANULARITY 1,
	INDEX idx_tenant_id tenant_id TYPE bloom_filter(0.01) GRANULARITY 1
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, device_id, timestamp)
TTL toDateTime(timestamp) + INTERVAL 180 DAY
SETTINGS index_granularity = 8192`

// schemaHumidityData 湿度数据表结构
const schemaHumidityData = `
CREATE TABLE IF NOT EXISTS nexlyn.iot_humidity_data (
	timestamp DateTime64(3) COMMENT '数据采集时间戳（毫秒精度）',
	device_id String COMMENT '设备唯一标识符',
	device_model String COMMENT '设备型号',
	device_category String COMMENT '设备类别',
	tenant_id String COMMENT '租户ID',
	value Float64 COMMENT '湿度值（%）',
	INDEX idx_device_id device_id TYPE bloom_filter(0.01) GRANULARITY 1,
	INDEX idx_tenant_id tenant_id TYPE bloom_filter(0.01) GRANULARITY 1
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, device_id, timestamp)
TTL toDateTime(timestamp) + INTERVAL 180 DAY
SETTINGS index_granularity = 8192`

// schemaSmokeData 烟雾浓度数据表结构
const schemaSmokeData = `
CREATE TABLE IF NOT EXISTS nexlyn.iot_smoke_data (
	timestamp DateTime64(3) COMMENT '数据采集时间戳（毫秒精度）',
	device_id String COMMENT '设备唯一标识符',
	device_model String COMMENT '设备型号',
	device_category String COMMENT '设备类别',
	tenant_id String COMMENT '租户ID',
	value Float64 COMMENT '烟雾浓度（ppm）',
	INDEX idx_device_id device_id TYPE bloom_filter(0.01) GRANULARITY 1,
	INDEX idx_tenant_id tenant_id TYPE bloom_filter(0.01) GRANULARITY 1
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, device_id, timestamp)
TTL toDateTime(timestamp) + INTERVAL 180 DAY
SETTINGS index_granularity = 8192`

// schemaIsOnlineData 在线状态数据表结构
const schemaIsOnlineData = `
CREATE TABLE IF NOT EXISTS nexlyn.iot_is_online_data (
	timestamp DateTime64(3) COMMENT '数据采集时间戳（毫秒精度）',
	device_id String COMMENT '设备唯一标识符',
	device_model String COMMENT '设备型号',
	device_category String COMMENT '设备类别',
	tenant_id String COMMENT '租户ID',
	value Bool COMMENT '在线状态（true/false）',
	INDEX idx_device_id device_id TYPE bloom_filter(0.01) GRANULARITY 1,
	INDEX idx_tenant_id tenant_id TYPE bloom_filter(0.01) GRANULARITY 1
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, device_id, timestamp)
TTL toDateTime(timestamp) + INTERVAL 180 DAY
SETTINGS index_granularity = 8192`

// generateCreateTableSQL 动态生成ClickHouse建表SQL
// 参数：
//   - database: 数据库名
//   - fieldName: 字段名（用于构建表名）
//   - fieldType: 字段类型（决定value列类型）
//   - displayName: 字段显示名称（用于注释）
//   - unit: 字段单位（用于注释）
//   - ttlDays: TTL时长（天）
//
// 返回：CREATE TABLE IF NOT EXISTS SQL语句
func generateCreateTableSQL(database string, fieldName core.FieldName, fieldType core.FieldType, displayName, unit string, ttlDays int) string {
	// 构建表名
	tableName := core.BuildClickHouseTableName(fieldName)

	// 根据字段类型决定value列的ClickHouse类型
	var valueType string
	var valueComment string

	switch fieldType {
	case core.FieldTypeNumber:
		valueType = "Float64"
		if unit != "" {
			valueComment = fmt.Sprintf("%s值（%s）", displayName, unit)
		} else {
			valueComment = fmt.Sprintf("%s值", displayName)
		}
	case core.FieldTypeBoolean:
		valueType = "Bool"
		valueComment = fmt.Sprintf("%s（true/false）", displayName)
	case core.FieldTypeString:
		valueType = "String"
		valueComment = fmt.Sprintf("%s", displayName)
	case core.FieldTypeTimestamp, core.FieldTypeDatetime:
		valueType = "DateTime64(3)"
		valueComment = fmt.Sprintf("%s", displayName)
	case core.FieldTypeImage, core.FieldTypeImageBase64:
		valueType = "String"
		valueComment = fmt.Sprintf("%s", displayName)
	default:
		// 默认使用String类型
		valueType = "String"
		valueComment = fmt.Sprintf("%s", displayName)
	}

	// 生成SQL
	sql := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s.%s (
	timestamp DateTime64(3) COMMENT '数据采集时间戳（毫秒精度）',
	device_id String COMMENT '设备唯一标识符',
	device_model String COMMENT '设备型号',
	device_category String COMMENT '设备类别',
	tenant_id String COMMENT '租户ID',
	value %s COMMENT '%s',
	INDEX idx_device_id device_id TYPE bloom_filter(0.01) GRANULARITY 1,
	INDEX idx_tenant_id tenant_id TYPE bloom_filter(0.01) GRANULARITY 1
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, device_id, timestamp)
TTL toDateTime(timestamp) + INTERVAL %d DAY
SETTINGS index_granularity = 8192`, database, tableName, valueType, valueComment, ttlDays)

	return sql
}
