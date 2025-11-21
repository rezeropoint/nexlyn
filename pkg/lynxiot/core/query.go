// Package core 定义IoT设备管理的核心领域模型
//
// 文件说明：query.go
// 职责：时序数据查询相关的领域模型和类型定义
package core

import "time"

// AggregationType 聚合类型
type AggregationType string

const (
	AggNone  AggregationType = ""      // 原始数据（不聚合）
	AggAvg   AggregationType = "avg"   // 平均值
	AggMax   AggregationType = "max"   // 最大值
	AggMin   AggregationType = "min"   // 最小值
	AggSum   AggregationType = "sum"   // 求和
	AggCount AggregationType = "count" // 计数
	AggLast  AggregationType = "last"  // 最新值（argMax）
)

// TimeSeriesQuery 时序数据查询参数
type TimeSeriesQuery struct {
	// 基础过滤条件
	TenantID   string      `json:"tenantId"`            // 租户ID（必填，用于安全隔离）
	OrgIDs     []string    `json:"orgIds"`              // 组织ID列表（必填，由REST层根据用户权限提供）
	DeviceIDs  []string    `json:"deviceIds,omitempty"` // 设备ID列表（可选，在OrgIDs范围内进一步过滤）
	FieldNames []FieldName `json:"fieldNames"`          // 查询字段列表（必填）
	StartTime  time.Time   `json:"startTime"`           // 开始时间（必填）
	EndTime    time.Time   `json:"endTime"`             // 结束时间（必填）

	// 聚合选项
	Aggregation AggregationType `json:"aggregation,optional"` // 聚合类型（默认为none，返回原始数据）
	Interval    time.Duration   `json:"interval,optional"`    // 聚合时间间隔（如1小时、1天，仅在Aggregation非none时有效）

	// 分页和排序
	Limit    int    `json:"limit,optional"`    // 限制返回条数（默认1000）
	Offset   int    `json:"offset,optional"`   // 偏移量（默认0）
	OrderBy  string `json:"orderBy,optional"`  // 排序字段（timestamp/device_id，默认timestamp）
	OrderDir string `json:"orderDir,optional"` // 排序方向（asc/desc，默认asc）
}

// TimeSeriesData 时序数据结果
type TimeSeriesData struct {
	Timestamp      time.Time      `json:"timestamp"`      // 数据时间戳
	DeviceID       string         `json:"deviceId"`       // 设备唯一标识符
	DeviceModel    string         `json:"deviceModel"`    // 设备型号
	DeviceCategory DeviceCategory `json:"deviceCategory"` // 设备类别
	FieldName      FieldName      `json:"fieldName"`      // 字段名称
	Value          any            `json:"value"`          // 字段值
}

// TimeSeriesResult 查询结果包装（支持分页）
type TimeSeriesResult struct {
	Data     []TimeSeriesData `json:"data"`     // 数据列表
	Total    int64            `json:"total"`    // 总记录数
	Page     int              `json:"page"`     // 当前页码（从查询参数计算）
	PageSize int              `json:"pageSize"` // 每页数量（等于Limit）
}

// LatestValuesQuery 最新值查询参数
type LatestValuesQuery struct {
	TenantID   string      `json:"tenantId"`            // 租户ID（必填）
	OrgIDs     []string    `json:"orgIds"`              // 组织ID列表（必填，由REST层根据用户权限提供）
	DeviceIDs  []string    `json:"deviceIds,omitempty"` // 设备ID列表（可选，在OrgIDs范围内进一步过滤）
	FieldNames []FieldName `json:"fieldNames"`          // 字段列表（必填）
}

// DeviceLatestValues 设备最新值
type DeviceLatestValues struct {
	DeviceID   string                  `json:"deviceId"`   // 设备ID
	Values     map[FieldName]any       `json:"values"`     // 字段最新值
	Timestamps map[FieldName]time.Time `json:"timestamps"` // 字段最新时间戳
}

// DeviceStatisticsQuery 统计查询参数
type DeviceStatisticsQuery struct {
	TenantID   string      `json:"tenantId"`            // 租户ID（必填）
	OrgIDs     []string    `json:"orgIds"`              // 组织ID列表（必填，由REST层根据用户权限提供）
	DeviceIDs  []string    `json:"deviceIds,omitempty"` // 设备ID列表（可选，在OrgIDs范围内进一步过滤）
	FieldNames []FieldName `json:"fieldNames"`          // 字段列表（必填）
	StartTime  time.Time   `json:"startTime"`           // 开始时间（必填）
	EndTime    time.Time   `json:"endTime"`             // 结束时间（必填）
}

// DeviceStatistics 设备统计数据
type DeviceStatistics struct {
	DeviceID    string                  `json:"deviceId"`    // 设备ID
	DeviceModel string                  `json:"deviceModel"` // 设备型号
	FieldStats  map[FieldName]FieldStat `json:"fieldStats"`  // 字段统计信息
}

// FieldStat 字段统计
type FieldStat struct {
	Min       float64   `json:"min"`       // 最小值
	Max       float64   `json:"max"`       // 最大值
	Avg       float64   `json:"avg"`       // 平均值
	Sum       float64   `json:"sum"`       // 总和
	Count     int64     `json:"count"`     // 数据点数量
	LastValue any       `json:"lastValue"` // 最新值
	LastTime  time.Time `json:"lastTime"`  // 最新值时间戳
}
