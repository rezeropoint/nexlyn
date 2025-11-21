package storage

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// StorageManager 传感器数据存储管理器接口
type StorageManager interface {
	// StoreSensorData 存储传感器数据
	// 根据 DeviceCategory 自动路由到对应的ClickHouse表
	StoreSensorData(ctx context.Context, record *core.SensorDataRecord) error

	// Close 关闭存储连接
	Close() error
}

func NewManager(ctx context.Context, clickHouseConn driver.Conn, config Config) (StorageManager, error) {
	return newStorageManager(ctx, clickHouseConn, config)
}
