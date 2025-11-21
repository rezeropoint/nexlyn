package core

import (
	"encoding/json"
	"fmt"
)

// ============================================================================
// TimeSeriesDataset 时序数据集标准格式
// ============================================================================

// TimeSeriesDataset 时序数据集（用于GraphContext数据传递）
//
// 设计理念：
//   - 统一数据格式，避免每个积木都做类型转换
//   - 单一职责：每个数据集只包含一个设备的一个字段
//   - 类型安全：Value统一为float64，简化后续计算
//   - 元数据分离：业务数据和元数据解耦
//
// 使用示例：
//
//	dataset := TimeSeriesDataset{
//	    DeviceID:  "temp001",
//	    FieldName: "temperature",
//	    DataPoints: []DataPoint{
//	        {Timestamp: 1234567890, Value: 25.5},
//	        {Timestamp: 1234567891, Value: 26.0},
//	    },
//	    Metadata: map[string]interface{}{
//	        "deviceModel":    "AI-200",
//	        "deviceCategory": "ai_box",
//	        "unit":           "℃",
//	    },
//	}
type TimeSeriesDataset struct {
	DeviceID   string                 `json:"deviceId"`   // 设备唯一标识符
	FieldName  string                 `json:"fieldName"`  // 字段名称（如 "temperature"）
	DataPoints []DataPoint            `json:"dataPoints"` // 数据点列表（按时间戳升序排列）
	Metadata   map[string]interface{} `json:"metadata"`   // 元数据（设备型号、类别、单位等）
}

// DataPoint 标准数据点
type DataPoint struct {
	Timestamp int64   `json:"timestamp"` // Unix时间戳（秒）
	Value     float64 `json:"value"`     // 数值（统一为float64）
}

// GetValues 提取所有数值（用于统计分析）
func (ds *TimeSeriesDataset) GetValues() []float64 {
	values := make([]float64, len(ds.DataPoints))
	for i, dp := range ds.DataPoints {
		values[i] = dp.Value
	}
	return values
}

// GetTimestamps 提取所有时间戳（用于时间序列分析）
func (ds *TimeSeriesDataset) GetTimestamps() []int64 {
	timestamps := make([]int64, len(ds.DataPoints))
	for i, dp := range ds.DataPoints {
		timestamps[i] = dp.Timestamp
	}
	return timestamps
}

// Len 返回数据点数量
func (ds *TimeSeriesDataset) Len() int {
	return len(ds.DataPoints)
}

// IsEmpty 检查数据集是否为空
func (ds *TimeSeriesDataset) IsEmpty() bool {
	return len(ds.DataPoints) == 0
}

// GetMetadataString 获取字符串类型的元数据
func (ds *TimeSeriesDataset) GetMetadataString(key string) (string, bool) {
	if val, ok := ds.Metadata[key]; ok {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

// ToJSON 转换为JSON字符串（用于日志打印）
func (ds *TimeSeriesDataset) ToJSON() string {
	data, err := json.Marshal(ds)
	if err != nil {
		return fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
	}
	return string(data)
}

// ============================================================================
// MultiFieldTimeSeriesDataset 多字段时序数据集（可选，用于批量查询优化）
// ============================================================================

// MultiFieldTimeSeriesDataset 多字段时序数据集
//
// 使用场景：当需要同时处理同一设备的多个字段时（如温度+湿度+压力）
//
// 注意：当前版本不使用此结构，保留供后续批量查询优化使用
type MultiFieldTimeSeriesDataset struct {
	DeviceID   string                        `json:"deviceId"`   // 设备唯一标识符
	FieldNames []string                      `json:"fieldNames"` // 字段名称列表
	DataPoints []MultiFieldDataPoint         `json:"dataPoints"` // 数据点列表
	Metadata   map[string]interface{}        `json:"metadata"`   // 元数据
	FieldMeta  map[string]map[string]string  `json:"fieldMeta"`  // 字段级别元数据（如单位）
}

// MultiFieldDataPoint 多字段数据点
type MultiFieldDataPoint struct {
	Timestamp int64              `json:"timestamp"` // Unix时间戳（秒）
	Values    map[string]float64 `json:"values"`    // 字段名 -> 数值映射
}

// SplitByField 将多字段数据集拆分为多个单字段数据集
func (mds *MultiFieldTimeSeriesDataset) SplitByField() []TimeSeriesDataset {
	datasets := make([]TimeSeriesDataset, 0, len(mds.FieldNames))

	for _, fieldName := range mds.FieldNames {
		dataset := TimeSeriesDataset{
			DeviceID:   mds.DeviceID,
			FieldName:  fieldName,
			DataPoints: make([]DataPoint, 0, len(mds.DataPoints)),
			Metadata:   mds.Metadata,
		}

		// 提取该字段的所有数据点
		for _, mdp := range mds.DataPoints {
			if val, ok := mdp.Values[fieldName]; ok {
				dataset.DataPoints = append(dataset.DataPoints, DataPoint{
					Timestamp: mdp.Timestamp,
					Value:     val,
				})
			}
		}

		datasets = append(datasets, dataset)
	}

	return datasets
}
