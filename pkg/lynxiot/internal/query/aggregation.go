package query

import (
	"fmt"
	"strings"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
)

// buildAggregatedQuery 构建聚合查询SQL（带时间窗口）
// 使用UNION ALL合并多个字段表的聚合结果
func buildAggregatedQuery(query core.TimeSeriesQuery, database string) (string, []any) {
	var queries []string
	var args []any

	// 计算时间窗口大小（秒）
	intervalSeconds := int(query.Interval.Seconds())

	for _, field := range query.FieldNames {
		tableName := buildTableNameByField(field)

		// 根据聚合类型选择聚合函数
		aggFunc := getAggregationFunction(query.Aggregation)

		sqlQuery := fmt.Sprintf(`
			SELECT
				toStartOfInterval(timestamp, INTERVAL %d SECOND) as timestamp,
				device_id,
				any(device_model) as device_model,
				any(device_category) as device_category,
				'%s' as field_name,
				%s as value
			FROM %s.%s
			WHERE tenant_id = ?
			  AND timestamp >= ?
			  AND timestamp <= ?
		`, intervalSeconds, field, aggFunc, database, tableName)

		// 添加设备ID过滤
		if len(query.DeviceIDs) > 0 {
			sqlQuery += fmt.Sprintf(" AND device_id IN (%s)", buildPlaceholders(len(query.DeviceIDs)))
		}

		// 按时间窗口和设备ID分组
		sqlQuery += " GROUP BY timestamp, device_id"

		queries = append(queries, sqlQuery)

		// 添加参数
		args = append(args, query.TenantID, query.StartTime, query.EndTime)
		if len(query.DeviceIDs) > 0 {
			args = append(args, toAnySlice(query.DeviceIDs)...)
		}
	}

	// UNION所有字段查询
	finalSQL := strings.Join(queries, " UNION ALL ")

	// 添加排序
	finalSQL += fmt.Sprintf(" ORDER BY %s %s", query.OrderBy, query.OrderDir)

	// 添加分页
	finalSQL += fmt.Sprintf(" LIMIT %d OFFSET %d", query.Limit, query.Offset)

	return finalSQL, args
}

// buildAggregatedCountQuery 构建聚合查询的总数查询SQL
func buildAggregatedCountQuery(query core.TimeSeriesQuery, database string) (string, []any) {
	var subQueries []string
	var args []any

	// 计算时间窗口大小（秒）
	intervalSeconds := int(query.Interval.Seconds())

	for _, field := range query.FieldNames {
		tableName := buildTableNameByField(field)

		sqlQuery := fmt.Sprintf(`
			SELECT COUNT(DISTINCT (toStartOfInterval(timestamp, INTERVAL %d SECOND), device_id)) as cnt
			FROM %s.%s
			WHERE tenant_id = ?
			  AND timestamp >= ?
			  AND timestamp <= ?
		`, intervalSeconds, database, tableName)

		// 添加设备ID过滤
		if len(query.DeviceIDs) > 0 {
			sqlQuery += fmt.Sprintf(" AND device_id IN (%s)", buildPlaceholders(len(query.DeviceIDs)))
		}

		subQueries = append(subQueries, fmt.Sprintf("(%s)", sqlQuery))

		// 添加参数
		args = append(args, query.TenantID, query.StartTime, query.EndTime)
		if len(query.DeviceIDs) > 0 {
			args = append(args, toAnySlice(query.DeviceIDs)...)
		}
	}

	// 对所有子查询的count求和
	finalSQL := fmt.Sprintf("SELECT sum(cnt) as total FROM (%s)", strings.Join(subQueries, " UNION ALL "))

	return finalSQL, args
}

// getAggregationFunction 根据聚合类型返回对应的ClickHouse聚合函数
func getAggregationFunction(aggType core.AggregationType) string {
	switch aggType {
	case core.AggAvg:
		return "avg(value)"
	case core.AggMax:
		return "max(value)"
	case core.AggMin:
		return "min(value)"
	case core.AggSum:
		return "sum(value)"
	case core.AggCount:
		return "count(*)"
	case core.AggLast:
		// argMax返回timestamp最大时的value值
		return "argMax(value, timestamp)"
	default:
		// 默认返回平均值
		return "avg(value)"
	}
}
