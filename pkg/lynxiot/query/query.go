package query

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
)

// Query 查询引擎接口
// 提供传感器历史数据查询能力（轻量级，只依赖PostgreSQL和ClickHouse）
// 适用于：
// - 后端服务间调用（如LynxGraph引擎查询传感器数据）
// - 独立的查询微服务
type Query interface {
	// QueryTimeSeries 查询时序数据
	// 支持原始数据查询和6种聚合类型（avg/max/min/sum/count/last）
	// 支持时间窗口聚合、分页、排序
	// 权限约束：通过org_ids参数实现（调用方需先查询用户可见组织）
	QueryTimeSeries(ctx context.Context, query core.TimeSeriesQuery) (*core.TimeSeriesResult, error)

	// GetLatestValues 获取设备字段最新值
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
