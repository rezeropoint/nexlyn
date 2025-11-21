package query

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// QueryManager 时序数据查询管理器接口
type QueryManager interface {
	// QueryTimeSeries 查询时序数据
	// 支持原始数据查询和时间窗口聚合
	// 返回结果按时间戳排序，支持分页
	QueryTimeSeries(ctx context.Context, query core.TimeSeriesQuery) (*core.TimeSeriesResult, error)

	// GetLatestValues 获取设备字段的最新值
	// 用于实时监控和可视化大屏展示
	// 每个设备返回指定字段的最新值和时间戳
	GetLatestValues(ctx context.Context, query core.LatestValuesQuery) ([]core.DeviceLatestValues, error)

	// GetDeviceStatistics 获取设备统计信息
	// 计算指定时间范围内的min/max/avg/sum/count等统计指标
	// 支持多设备、多字段统计
	GetDeviceStatistics(ctx context.Context, query core.DeviceStatisticsQuery) ([]core.DeviceStatistics, error)

	// Close 关闭查询连接
	Close() error
}

// NewQueryManager 创建Query Manager实例
// dbConn: PostgreSQL连接（用于查询设备绑定信息）
// clickHouseConn: ClickHouse连接（用于查询时序数据，由外部创建并传入）
func NewQueryManager(dbConn sqlx.SqlConn, clickHouseConn driver.Conn, config Config) (QueryManager, error) {
	return newQueryManager(dbConn, clickHouseConn, config)
}
