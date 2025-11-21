package query

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// queryManager Query Manager实现
type queryManager struct {
	config Config
	conn   driver.Conn  // ClickHouse连接
	dbConn sqlx.SqlConn // PostgreSQL连接（用于查询设备绑定信息）
}

// newQueryManager 创建Query Manager实例（非导出，内部使用）
// clickHouseConn: ClickHouse连接（由外部创建并传入）
func newQueryManager(dbConn sqlx.SqlConn, clickHouseConn driver.Conn, config Config) (QueryManager, error) {
	ctx := context.Background()

	logx.WithContext(ctx).WithFields(
		logx.Field("service", config.ServiceName),
		logx.Field("pod", config.PodName),
		logx.Field("module", "iot_query"),
		logx.Field("operation", "init"),
		logx.Field("database", config.Database),
	).Info("Query Manager初始化成功")

	return &queryManager{
		config: config,
		conn:   clickHouseConn,
		dbConn: dbConn,
	}, nil
}

// QueryTimeSeries 查询时序数据
func (m *queryManager) QueryTimeSeries(ctx context.Context, query core.TimeSeriesQuery) (*core.TimeSeriesResult, error) {
	// 1. 验证查询参数
	if err := validateQuery(query, m.config); err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "query_timeseries"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询参数验证失败")
		return nil, fmt.Errorf("查询参数验证失败: %w", err)
	}

	// 2. 解析设备ID（强制OrgIDs约束）
	resolvedDeviceIDs, err := resolveDeviceIDs(ctx, m.dbConn, query.TenantID, query.OrgIDs, query.DeviceIDs)
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "query_timeseries"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("解析设备ID失败")
		return nil, fmt.Errorf("解析设备ID失败: %w", err)
	}

	// 将解析后的设备ID列表覆盖到查询参数中
	query.DeviceIDs = resolvedDeviceIDs

	// 3. 标准化查询参数（设置默认值）
	normalizeQuery(&query, m.config)

	startTime := time.Now()

	// 4. 查询总数
	total, err := m.queryTotal(ctx, query)
	if err != nil {
		return nil, err
	}

	// 5. 查询数据
	data, err := m.queryData(ctx, query)
	if err != nil {
		return nil, err
	}

	// 6. 构建结果
	result := &core.TimeSeriesResult{
		Data:     data,
		Total:    total,
		Page:     query.Offset/query.Limit + 1,
		PageSize: query.Limit,
	}

	// 记录查询日志
	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_query"),
		logx.Field("operation", "query_timeseries"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", query.TenantID),
		logx.Field("device_count", len(query.DeviceIDs)),
		logx.Field("field_count", len(query.FieldNames)),
		logx.Field("aggregation", query.Aggregation),
		logx.Field("total_records", total),
		logx.Field("returned_records", len(data)),
		logx.Field("duration_ms", time.Since(startTime).Milliseconds()),
	).Info("时序数据查询成功")

	return result, nil
}

// queryTotal 查询总记录数
func (m *queryManager) queryTotal(ctx context.Context, query core.TimeSeriesQuery) (int64, error) {
	var sqlQuery string
	var args []any

	// 根据是否聚合选择不同的count查询
	if query.Aggregation != "" && query.Aggregation != core.AggNone {
		sqlQuery, args = buildAggregatedCountQuery(query, m.config.Database)
	} else {
		sqlQuery, args = buildCountQuery(query, m.config.Database)
	}

	var total sql.NullInt64
	err := m.conn.QueryRow(ctx, sqlQuery, args...).Scan(&total)
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "query_total"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询总数失败")
		return 0, fmt.Errorf("查询总数失败: %w", err)
	}

	if !total.Valid {
		return 0, nil
	}

	return total.Int64, nil
}

// queryData 查询时序数据
func (m *queryManager) queryData(ctx context.Context, query core.TimeSeriesQuery) ([]core.TimeSeriesData, error) {
	var sqlQuery string
	var args []any

	// 根据是否聚合选择不同的查询
	if query.Aggregation != "" && query.Aggregation != core.AggNone {
		sqlQuery, args = buildAggregatedQuery(query, m.config.Database)
	} else {
		sqlQuery, args = buildRawDataQuery(query, m.config.Database)
	}

	rows, err := m.conn.Query(ctx, sqlQuery, args...)
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "query_data"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询数据失败")
		return nil, fmt.Errorf("查询数据失败: %w", err)
	}
	defer rows.Close()

	// 解析结果
	var data []core.TimeSeriesData
	for rows.Next() {
		var record core.TimeSeriesData
		var fieldNameStr string

		err := rows.Scan(
			&record.Timestamp,
			&record.DeviceID,
			&record.DeviceModel,
			&record.DeviceCategory,
			&fieldNameStr,
			&record.Value,
		)
		if err != nil {
			logx.WithContext(ctx).WithFields(
				logx.Field("service", m.config.ServiceName),
				logx.Field("module", "iot_query"),
				logx.Field("operation", "scan_row"),
				logx.Field("status", "failed"),
				logx.Field("error", err.Error()),
			).Error("扫描查询结果失败")
			continue // 跳过错误行，继续处理其他行
		}

		record.FieldName = core.FieldName(fieldNameStr)
		data = append(data, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历查询结果失败: %w", err)
	}

	return data, nil
}

// GetLatestValues 获取设备字段的最新值
func (m *queryManager) GetLatestValues(ctx context.Context, query core.LatestValuesQuery) ([]core.DeviceLatestValues, error) {
	// 1. 验证查询参数
	if err := validateLatestValuesQuery(query); err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "get_latest_values"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询参数验证失败")
		return nil, fmt.Errorf("查询参数验证失败: %w", err)
	}

	// 2. 解析设备ID（强制OrgIDs约束）
	resolvedDeviceIDs, err := resolveDeviceIDs(ctx, m.dbConn, query.TenantID, query.OrgIDs, query.DeviceIDs)
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "get_latest_values"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("解析设备ID失败")
		return nil, fmt.Errorf("解析设备ID失败: %w", err)
	}

	startTime := time.Now()
	results := make([]core.DeviceLatestValues, 0, len(resolvedDeviceIDs))

	// 3. 遍历每个设备
	for _, deviceID := range resolvedDeviceIDs {
		values := make(map[core.FieldName]any)
		timestamps := make(map[core.FieldName]time.Time)

		// 3. 查询每个字段的最新值
		for _, field := range query.FieldNames {
			sqlQuery := buildLatestValueQuery(field, m.config.Database)

			var value any
			var ts time.Time
			err := m.conn.QueryRow(ctx, sqlQuery, query.TenantID, deviceID).Scan(&value, &ts)
			if err != nil {
				// 如果查询失败（包括没有数据sql.ErrNoRows），记录Debug日志并继续处理其他字段
				if err != sql.ErrNoRows {
					logx.WithContext(ctx).WithFields(
						logx.Field("service", m.config.ServiceName),
						logx.Field("module", "iot_query"),
						logx.Field("operation", "query_latest_value"),
						logx.Field("device_id", deviceID),
						logx.Field("field_name", field),
						logx.Field("error", err.Error()),
					).Error("查询最新值失败")
				}
				// 继续处理其他字段
				continue
			}

			// 成功查询到数据
			values[field] = value
			timestamps[field] = ts
		}

		results = append(results, core.DeviceLatestValues{
			DeviceID:   deviceID,
			Values:     values,
			Timestamps: timestamps,
		})
	}

	// 记录查询日志
	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_query"),
		logx.Field("operation", "get_latest_values"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", query.TenantID),
		logx.Field("device_count", len(query.DeviceIDs)),
		logx.Field("field_count", len(query.FieldNames)),
		logx.Field("duration_ms", time.Since(startTime).Milliseconds()),
	).Info("获取最新值成功")

	return results, nil
}

// GetDeviceStatistics 获取设备统计信息
func (m *queryManager) GetDeviceStatistics(ctx context.Context, query core.DeviceStatisticsQuery) ([]core.DeviceStatistics, error) {
	// 1. 验证查询参数
	if err := validateStatisticsQuery(query); err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "get_device_statistics"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询参数验证失败")
		return nil, fmt.Errorf("查询参数验证失败: %w", err)
	}

	// 2. 解析设备ID（强制OrgIDs约束）
	resolvedDeviceIDs, err := resolveDeviceIDs(ctx, m.dbConn, query.TenantID, query.OrgIDs, query.DeviceIDs)
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "get_device_statistics"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("解析设备ID失败")
		return nil, fmt.Errorf("解析设备ID失败: %w", err)
	}

	// 将解析后的设备ID列表覆盖到查询参数中
	query.DeviceIDs = resolvedDeviceIDs

	startTime := time.Now()
	statsMap := make(map[string]*core.DeviceStatistics)

	// 3. 遍历每个字段，查询统计信息
	for _, field := range query.FieldNames {
		sql, args := buildStatisticsQuery(field, query, m.config.Database)

		rows, err := m.conn.Query(ctx, sql, args...)
		if err != nil {
			logx.WithContext(ctx).WithFields(
				logx.Field("service", m.config.ServiceName),
				logx.Field("module", "iot_query"),
				logx.Field("operation", "query_statistics"),
				logx.Field("field_name", field),
				logx.Field("error", err.Error()),
			).Error("查询统计信息失败")
			continue // 继续处理其他字段
		}

		// 解析统计结果
		for rows.Next() {
			var deviceID string
			var deviceModel string
			var stat core.FieldStat

			err := rows.Scan(
				&deviceID,
				&deviceModel,
				&stat.Min,
				&stat.Max,
				&stat.Avg,
				&stat.Sum,
				&stat.Count,
				&stat.LastValue,
				&stat.LastTime,
			)
			if err != nil {
				logx.WithContext(ctx).WithFields(
					logx.Field("service", m.config.ServiceName),
					logx.Field("module", "iot_query"),
					logx.Field("operation", "scan_statistics"),
					logx.Field("error", err.Error()),
				).Error("扫描统计结果失败")
				continue
			}

			// 初始化设备统计对象
			if _, exists := statsMap[deviceID]; !exists {
				statsMap[deviceID] = &core.DeviceStatistics{
					DeviceID:    deviceID,
					DeviceModel: deviceModel,
					FieldStats:  make(map[core.FieldName]core.FieldStat),
				}
			}

			// 添加字段统计
			statsMap[deviceID].FieldStats[field] = stat
		}

		rows.Close()
	}

	// 3. 转换为切片
	results := make([]core.DeviceStatistics, 0, len(statsMap))
	for _, stat := range statsMap {
		results = append(results, *stat)
	}

	// 记录查询日志
	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_query"),
		logx.Field("operation", "get_device_statistics"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", query.TenantID),
		logx.Field("device_count", len(results)),
		logx.Field("field_count", len(query.FieldNames)),
		logx.Field("duration_ms", time.Since(startTime).Milliseconds()),
	).Info("获取设备统计成功")

	return results, nil
}

// Close 关闭查询连接
func (m *queryManager) Close() error {
	if m.conn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "close"),
		).Info("关闭Query Manager连接")

		return m.conn.Close()
	}
	return nil
}
