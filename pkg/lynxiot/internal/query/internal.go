package query

import (
	"context"
	"fmt"
	"strings"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// buildTableNameByField 根据字段名构建ClickHouse表名
// 格式: iot_{field_name}_data
// 例如：temperature -> iot_temperature_data
func buildTableNameByField(fieldName core.FieldName) string {
	return fmt.Sprintf("iot_%s_data", string(fieldName))
}

// buildPlaceholders 构建SQL占位符字符串
// 例如：buildPlaceholders(3) 返回 "?, ?, ?"
func buildPlaceholders(count int) string {
	if count <= 0 {
		return ""
	}
	placeholders := make([]string, count)
	for i := range placeholders {
		placeholders[i] = "?"
	}
	return strings.Join(placeholders, ", ")
}

// toAnySlice 将字符串切片转换为any切片（用于SQL参数）
func toAnySlice(strs []string) []any {
	result := make([]any, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result
}

// validateQuery 验证查询参数的有效性
func validateQuery(query core.TimeSeriesQuery, config Config) error {
	// 1. 验证必填字段
	if query.TenantID == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if len(query.OrgIDs) == 0 {
		return fmt.Errorf("组织ID列表不能为空")
	}
	if len(query.FieldNames) == 0 {
		return fmt.Errorf("查询字段列表不能为空")
	}
	if query.StartTime.IsZero() {
		return fmt.Errorf("开始时间不能为空")
	}
	if query.EndTime.IsZero() {
		return fmt.Errorf("结束时间不能为空")
	}

	// 2. 验证时间范围
	if query.EndTime.Before(query.StartTime) {
		return fmt.Errorf("结束时间不能早于开始时间")
	}

	// 3. 验证查询限制
	if len(query.FieldNames) > config.MaxFieldsPerReq {
		return fmt.Errorf("单次请求最多支持查询%d个字段", config.MaxFieldsPerReq)
	}

	// 4. 验证分页参数
	if query.Limit < 0 {
		return fmt.Errorf("Limit不能为负数")
	}
	if query.Limit > config.MaxLimit {
		return fmt.Errorf("Limit不能超过%d", config.MaxLimit)
	}
	if query.Offset < 0 {
		return fmt.Errorf("Offset不能为负数")
	}

	// 5. 验证聚合参数
	if query.Aggregation != "" && query.Aggregation != core.AggNone {
		validAggs := map[core.AggregationType]bool{
			core.AggAvg:   true,
			core.AggMax:   true,
			core.AggMin:   true,
			core.AggSum:   true,
			core.AggCount: true,
			core.AggLast:  true,
		}
		if !validAggs[query.Aggregation] {
			return fmt.Errorf("无效的聚合类型: %s", query.Aggregation)
		}

		if query.Interval <= 0 {
			return fmt.Errorf("聚合查询需要指定Interval（聚合时间间隔）")
		}
	}

	return nil
}

// validateLatestValuesQuery 验证最新值查询参数
func validateLatestValuesQuery(query core.LatestValuesQuery) error {
	if query.TenantID == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if len(query.OrgIDs) == 0 {
		return fmt.Errorf("组织ID列表不能为空")
	}
	if len(query.FieldNames) == 0 {
		return fmt.Errorf("查询字段列表不能为空")
	}
	return nil
}

// validateStatisticsQuery 验证统计查询参数
func validateStatisticsQuery(query core.DeviceStatisticsQuery) error {
	if query.TenantID == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if len(query.OrgIDs) == 0 {
		return fmt.Errorf("组织ID列表不能为空")
	}
	if len(query.FieldNames) == 0 {
		return fmt.Errorf("查询字段列表不能为空")
	}
	if query.StartTime.IsZero() {
		return fmt.Errorf("开始时间不能为空")
	}
	if query.EndTime.IsZero() {
		return fmt.Errorf("结束时间不能为空")
	}
	if query.EndTime.Before(query.StartTime) {
		return fmt.Errorf("结束时间不能早于开始时间")
	}
	return nil
}

// normalizeQuery 标准化查询参数（设置默认值）
func normalizeQuery(query *core.TimeSeriesQuery, config Config) {
	// 设置默认Limit
	if query.Limit <= 0 {
		query.Limit = config.DefaultLimit
	}
	if query.Limit > config.MaxLimit {
		query.Limit = config.MaxLimit
	}

	// 设置默认Offset
	if query.Offset < 0 {
		query.Offset = 0
	}

	// 设置默认排序
	if query.OrderBy == "" {
		query.OrderBy = "timestamp"
	}
	if query.OrderDir == "" {
		query.OrderDir = "ASC"
	} else {
		// 统一转换为大写
		query.OrderDir = strings.ToUpper(query.OrderDir)
	}
}

// resolveDeviceIDs 根据OrgIDs和DeviceIDs解析出最终的设备ID列表
// 安全约束：DeviceIDs必须在OrgIDs范围内，防止通过指定设备ID逃脱组织权限约束
func resolveDeviceIDs(ctx context.Context, dbConn sqlx.SqlConn, tenantID string, orgIDs []string, deviceIDs []string) ([]string, error) {
	// 1. 从PostgreSQL查询OrgIDs范围内的所有设备ID
	validDeviceIDs, err := queryDeviceIDsByOrgs(ctx, dbConn, tenantID, orgIDs)
	if err != nil {
		return nil, fmt.Errorf("查询组织设备ID失败: %w", err)
	}

	// 2. 如果未指定DeviceIDs，直接返回组织内所有设备ID
	if len(deviceIDs) == 0 {
		return validDeviceIDs, nil
	}

	// 3. 如果指定了DeviceIDs，取交集（强制约束在OrgIDs范围内）
	finalDeviceIDs := intersect(validDeviceIDs, deviceIDs)

	return finalDeviceIDs, nil
}

// queryDeviceIDsByOrgs 从PostgreSQL查询指定组织列表中的所有设备ID
func queryDeviceIDsByOrgs(ctx context.Context, dbConn sqlx.SqlConn, tenantID string, orgIDs []string) ([]string, error) {
	// 构建查询SQL（使用PostgreSQL的ANY语法）
	query := `
		SELECT device_id
		FROM iot_device_bindings
		WHERE tenant_id = $1
		  AND org_id = ANY($2)
		  AND deleted_at IS NULL
	`

	var results []struct {
		DeviceID string `db:"device_id"`
	}

	err := dbConn.QueryRowsCtx(ctx, &results, query, tenantID, pq.Array(orgIDs))
	if err != nil {
		return nil, fmt.Errorf("查询设备绑定失败: %w", err)
	}

	deviceIDs := make([]string, 0, len(results))
	for _, r := range results {
		deviceIDs = append(deviceIDs, r.DeviceID)
	}

	return deviceIDs, nil
}

// intersect 计算两个字符串切片的交集
func intersect(a, b []string) []string {
	// 将第一个切片转换为map以提高查找效率
	setA := make(map[string]bool)
	for _, item := range a {
		setA[item] = true
	}

	// 遍历第二个切片，只保留在第一个切片中存在的元素
	var result []string
	seen := make(map[string]bool) // 去重
	for _, item := range b {
		if setA[item] && !seen[item] {
			result = append(result, item)
			seen[item] = true
		}
	}

	return result
}

// buildRawDataQuery 构建原始数据查询SQL（不聚合）
// 使用UNION ALL合并多个字段表的查询结果
func buildRawDataQuery(query core.TimeSeriesQuery, database string) (string, []any) {
	var queries []string
	var args []any

	for _, field := range query.FieldNames {
		tableName := buildTableNameByField(field)

		sqlQuery := fmt.Sprintf(`
			SELECT
				timestamp,
				device_id,
				device_model,
				device_category,
				'%s' as field_name,
				value
			FROM %s.%s
			WHERE tenant_id = ?
			  AND timestamp >= ?
			  AND timestamp <= ?
		`, field, database, tableName)

		// 添加设备ID过滤
		if len(query.DeviceIDs) > 0 {
			sqlQuery += fmt.Sprintf(" AND device_id IN (%s)", buildPlaceholders(len(query.DeviceIDs)))
		}

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

// buildCountQuery 构建总数查询SQL
func buildCountQuery(query core.TimeSeriesQuery, database string) (string, []any) {
	var subQueries []string
	var args []any

	for _, field := range query.FieldNames {
		tableName := buildTableNameByField(field)

		sqlQuery := fmt.Sprintf(`
			SELECT COUNT(*) as cnt
			FROM %s.%s
			WHERE tenant_id = ?
			  AND timestamp >= ?
			  AND timestamp <= ?
		`, database, tableName)

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

// buildLatestValueQuery 构建单个字段的最新值查询SQL
func buildLatestValueQuery(fieldName core.FieldName, database string) string {
	tableName := buildTableNameByField(fieldName)

	return fmt.Sprintf(`
		SELECT value, timestamp
		FROM %s.%s
		WHERE tenant_id = ? AND device_id = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`, database, tableName)
}

// buildStatisticsQuery 构建单个字段的统计查询SQL
func buildStatisticsQuery(fieldName core.FieldName, query core.DeviceStatisticsQuery, database string) (string, []any) {
	tableName := buildTableNameByField(fieldName)

	sqlQuery := fmt.Sprintf(`
		SELECT
			device_id,
			any(device_model) as device_model,
			min(value) as min_val,
			max(value) as max_val,
			avg(value) as avg_val,
			sum(value) as sum_val,
			count(*) as count_val,
			argMax(value, timestamp) as last_val,
			max(timestamp) as last_time
		FROM %s.%s
		WHERE tenant_id = ?
		  AND timestamp >= ?
		  AND timestamp <= ?
	`, database, tableName)

	args := []any{query.TenantID, query.StartTime, query.EndTime}

	// 添加设备ID过滤
	if len(query.DeviceIDs) > 0 {
		sqlQuery += fmt.Sprintf(" AND device_id IN (%s)", buildPlaceholders(len(query.DeviceIDs)))
		args = append(args, toAnySlice(query.DeviceIDs)...)
	}

	sqlQuery += " GROUP BY device_id"

	return sqlQuery, args
}
