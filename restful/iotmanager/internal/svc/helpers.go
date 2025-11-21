package svc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// GetTenantIdByKey 根据租户Key查询租户ID
// 参数:
//   - ctx: 上下文
//   - dbConn: 数据库连接
//   - tenantKey: 租户Key
//
// 返回:
//   - string: 租户ID
//   - error: 错误信息
func GetTenantIdByKey(ctx context.Context, dbConn sqlx.SqlConn, tenantKey string) (string, error) {
	var tenantId string
	getTenantQuery := `SELECT id FROM system_tenants WHERE tenant_key = $1 AND status = 'active'`
	err := dbConn.QueryRowCtx(ctx, &tenantId, getTenantQuery, tenantKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("租户不存在或未激活: %s", tenantKey)
		}
		return "", fmt.Errorf("查询租户失败: %w", err)
	}
	return tenantId, nil
}

// ParseTimeString 解析时间字符串为time.Time
// 支持多种时间格式：
//   - ISO8601: "2024-01-01T00:00:00Z"
//   - RFC3339: "2024-01-01T00:00:00+08:00"
//   - RFC3339Nano: "2024-01-01T00:00:00.123456789+08:00"
//
// 参数:
//   - timeStr: 时间字符串
//
// 返回:
//   - time.Time: 解析后的时间
//   - error: 错误信息
func ParseTimeString(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("时间字符串不能为空")
	}

	// 尝试多种时间格式
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	var lastErr error
	for _, format := range formats {
		t, err := time.Parse(format, timeStr)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}

	return time.Time{}, fmt.Errorf("无法解析时间字符串 '%s': %w", timeStr, lastErr)
}

// BuildDeviceIdsList 将设备ID列表转换为PostgreSQL数组参数
// 参数:
//   - deviceIds: 设备ID列表
//
// 返回:
//   - pq.StringArray: PostgreSQL数组参数
func BuildDeviceIdsList(deviceIds []string) pq.StringArray {
	if deviceIds == nil {
		return pq.StringArray{}
	}
	return pq.StringArray(deviceIds)
}
