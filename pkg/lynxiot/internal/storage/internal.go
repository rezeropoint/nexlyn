package storage

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core/fields"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/zeromicro/go-zero/core/logx"
)

// buildTableNameByField 根据字段名构建ClickHouse表名
// 使用core包的统一函数
func buildTableNameByField(fieldName core.FieldName) string {
	return core.BuildClickHouseTableName(fieldName)
}

// storeFieldData 存储单个字段的数据到ClickHouse
// 注意：值已在MQTT提取时转换为正确类型（float64/string/bool等），直接插入即可
func (m *storageManager) storeFieldData(ctx context.Context, tableName string, record *core.SensorDataRecord, fieldName core.FieldName, value any) error {
	// 构建INSERT SQL（通用表结构：timestamp, device_id, device_model, device_category, tenant_id, value）
	query := fmt.Sprintf(`
		INSERT INTO %s.%s (
			timestamp, device_id, device_model, device_category, tenant_id, value
		) VALUES (?, ?, ?, ?, ?, ?)
	`, m.config.Database, tableName)

	// 执行插入（值类型已在MQTT提取时转换，表结构的value列类型与FieldType匹配）
	err := m.conn.Exec(ctx, query,
		record.Timestamp,
		record.DeviceID,
		record.DeviceModel,
		string(record.DeviceCategory),
		record.TenantID,
		value,
	)

	if err != nil {
		return fmt.Errorf("存储数据到ClickHouse失败: %w", err)
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_storage"),
		logx.Field("operation", "insert"),
		logx.Field("status", "success"),
		logx.Field("table", tableName),
		logx.Field("field_name", string(fieldName)),
		logx.Field("device_id", record.DeviceID),
		logx.Field("device_category", record.DeviceCategory),
		logx.Field("tenant_id", record.TenantID),
		logx.Field("timestamp", record.Timestamp),
		logx.Field("value", value),
	).Debug("成功存储字段数据")

	return nil
}

// initializeClickHouse 初始化ClickHouse数据库和表
// 动态遍历字段注册表，为每个字段创建对应的时序数据表
func initializeClickHouse(ctx context.Context, conn driver.Conn, config Config) error {
	// 1. 创建数据库
	createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", config.Database)
	if err := conn.Exec(ctx, createDBSQL); err != nil {
		return fmt.Errorf("创建数据库失败: %w", err)
	}
	logx.WithContext(ctx).Infof("ClickHouse数据库 '%s' 已存在或已创建", config.Database)

	// 2. 遍历字段注册表，为每个字段动态创建表
	allFields := fields.GetAllFields()
	createdCount := 0

	for _, field := range allFields {
		// 构建表名
		tableName := core.BuildClickHouseTableName(core.FieldName(field.Name))

		// 获取TTL配置（强制要求配置文件提供）
		ttlDays := resolveTTL(field, config.TTLOverrides)

		// 验证TTL有效性
		if ttlDays <= 0 {
			logx.WithContext(ctx).WithFields(
				logx.Field("field_name", field.Name),
				logx.Field("table_name", tableName),
			).Error("字段TTL配置缺失或无效，跳过表创建。请在配置文件的ClickHouseConfig.TTLOverrides中为该字段配置TTL")
			continue // 跳过该字段
		}

		// 动态生成建表SQL
		createSQL := generateCreateTableSQL(
			config.Database,
			core.FieldName(field.Name),
			field.FieldType,
			field.DisplayName,
			field.Unit,
			ttlDays,
		)

		// 检查表是否存在，不存在则创建
		if err := createTableIfNotExists(ctx, conn, config, tableName, createSQL, &createdCount); err != nil {
			return err
		}

		logx.WithContext(ctx).WithFields(
			logx.Field("field_name", field.Name),
			logx.Field("table_name", tableName),
			logx.Field("ttl_days", ttlDays),
			logx.Field("ttl_source", getTTLSource(field, config.TTLOverrides)),
		).Debug("字段表初始化完成")
	}

	if createdCount > 0 {
		logx.WithContext(ctx).Infof("ClickHouse初始化完成，创建了 %d 张新表", createdCount)
	} else {
		logx.WithContext(ctx).Info("ClickHouse所有时序数据表已存在，跳过初始化")
	}

	return nil
}

// resolveTTL 解析字段的TTL配置（强制要求配置文件提供）
func resolveTTL(field core.StandardField, ttlOverrides map[string]int) int {
	// 检查配置文件是否提供了TTL
	if ttlOverrides != nil {
		if ttl, exists := ttlOverrides[field.Name]; exists && ttl > 0 {
			return ttl
		}
	}

	// 未配置时返回0（会在日志中警告）
	return 0
}

// getTTLSource 获取TTL来源（用于日志记录）
func getTTLSource(field core.StandardField, ttlOverrides map[string]int) string {
	if ttlOverrides != nil {
		if _, exists := ttlOverrides[field.Name]; exists {
			return "config_file"
		}
	}

	return "missing_config"
}

// createTableIfNotExists 检查表是否存在，不存在则创建
func createTableIfNotExists(ctx context.Context, conn driver.Conn, config Config, tableName, createSQL string, createdCount *int) error {
	// 检查表是否存在
	var exists uint8
	query := fmt.Sprintf("EXISTS TABLE %s.%s", config.Database, tableName)
	if err := conn.QueryRow(ctx, query).Scan(&exists); err != nil {
		return fmt.Errorf("检查表 %s 是否存在失败: %w", tableName, err)
	}

	// 表不存在则创建
	if exists == 0 {
		logx.WithContext(ctx).Infof("表 %s 不存在，开始创建...", tableName)
		if err := conn.Exec(ctx, createSQL); err != nil {
			return fmt.Errorf("创建表 %s 失败: %w", tableName, err)
		}
		logx.WithContext(ctx).Infof("表 %s 创建成功", tableName)
		*createdCount++
	}

	return nil
}
