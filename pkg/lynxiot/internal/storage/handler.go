package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/zeromicro/go-zero/core/logx"
)

// storageManager Storage Manager实现
type storageManager struct {
	config Config
	conn   driver.Conn // ClickHouse连接
}

// NewStorageManager 创建Storage Manager实例
func newStorageManager(ctx context.Context, clickHouseConn driver.Conn, config Config) (*storageManager, error) {
	// 初始化数据库和表
	if err := initializeClickHouse(ctx, clickHouseConn, config); err != nil {
		return nil, fmt.Errorf("初始化ClickHouse失败: %w", err)
	}

	return &storageManager{
		config: config,
		conn:   clickHouseConn,
	}, nil
}

// StoreSensorData 存储传感器数据（按字段名分表存储）
func (m *storageManager) StoreSensorData(ctx context.Context, record *core.SensorDataRecord) error {
	// 遍历所有提取的字段，按字段名分别存储到对应的表
	for fieldName, value := range record.Fields {
		// 1. 根据字段名构建表名
		tableName := buildTableNameByField(fieldName)

		// 2. 存储单个字段的数据（值已经在MQTT提取时转换为正确类型）
		if err := m.storeFieldData(ctx, tableName, record, fieldName, value); err != nil {
			logx.WithContext(ctx).WithFields(
				logx.Field("service", m.config.ServiceName),
				logx.Field("pod", m.config.PodName),
				logx.Field("module", "iot_storage"),
				logx.Field("operation", "store_field_data"),
				logx.Field("status", "failed"),
				logx.Field("table", tableName),
				logx.Field("field_name", fieldName),
				logx.Field("device_id", record.DeviceID),
				logx.Field("error", err.Error()),
			).Error("存储字段数据失败")
			// 继续存储其他字段
		}
	}

	return nil
}

// Close 关闭存储连接
func (m *storageManager) Close() error {
	if m.conn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_storage"),
			logx.Field("operation", "close"),
		).Info("关闭Storage Manager连接")

		return m.conn.Close()
	}
	return nil
}
