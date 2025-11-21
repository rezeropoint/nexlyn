package query

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/internal/query"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// queryEngine 查询引擎实现
type queryEngine struct {
	queryManager query.QueryManager
	config       Config
}

// New 创建查询引擎实例
// 依赖注入：
//   - dbConn: PostgreSQL连接（用于查询设备绑定信息，获取设备ID范围）
//   - clickHouseConn: ClickHouse连接（用于查询时序数据）
//   - config: 查询引擎配置
//
// 注意：只需要PostgreSQL和ClickHouse，不需要Redis、Etcd、MQTT等依赖
func New(config Config, dbConn sqlx.SqlConn, clickHouseConn driver.Conn) (Query, error) {
	// 构建internal/query的配置
	queryConfig := query.Config{
		Database:        config.ClickHouseDatabase,
		DefaultLimit:    config.DefaultLimit,
		MaxLimit:        config.MaxLimit,
		QueryTimeout:    config.QueryTimeout,
		MaxFieldsPerReq: config.MaxFieldsPerReq,
		ServiceName:     config.ServiceName,
		PodName:         config.PodName,
	}

	// 创建query Manager
	queryManager, err := query.NewQueryManager(dbConn, clickHouseConn, queryConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create query manager: %w", err)
	}

	return &queryEngine{
		queryManager: queryManager,
		config:       config,
	}, nil
}

// QueryTimeSeries 查询时序数据
func (e *queryEngine) QueryTimeSeries(ctx context.Context, q core.TimeSeriesQuery) (*core.TimeSeriesResult, error) {
	return e.queryManager.QueryTimeSeries(ctx, q)
}

// GetLatestValues 获取设备字段最新值
func (e *queryEngine) GetLatestValues(ctx context.Context, q core.LatestValuesQuery) ([]core.DeviceLatestValues, error) {
	return e.queryManager.GetLatestValues(ctx, q)
}

// GetDeviceStatistics 获取设备统计信息
func (e *queryEngine) GetDeviceStatistics(ctx context.Context, q core.DeviceStatisticsQuery) ([]core.DeviceStatistics, error) {
	return e.queryManager.GetDeviceStatistics(ctx, q)
}

// Close 关闭查询连接
func (e *queryEngine) Close() error {
	return e.queryManager.Close()
}
