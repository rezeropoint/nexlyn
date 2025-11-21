package statistical

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/zeromicro/go-zero/core/logx"
)

const StatisticalAnalyzerVersion = "v1"

// Config 统计分析积木的配置
type Config struct {
	SourceKey     string   `json:"sourceKey" check:"must"`     // GraphContext中的数据源key
	FieldNames    []string `json:"fieldNames"`                 // 要分析的字段列表（多字段场景，可选）
	Metrics       []string `json:"metrics" check:"must"`       // 统计指标：["mean","std","max","min","median","variance"]
	SaveToContext string   `json:"saveToContext" check:"must"` // 保存结果的key
}

// StatisticalAnalyzerBlock 实现统计分析逻辑块
type StatisticalAnalyzerBlock struct {
	core.BaseLogicBlock
}

// NewStatisticalAnalyzerBlock 创建一个新的 StatisticalAnalyzerBlock 实例
func NewStatisticalAnalyzerBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &StatisticalAnalyzerBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeStatisticalAnalyzer,
			RawConfig: config,
		},
	}, nil
}

func (b *StatisticalAnalyzerBlock) GetID() string                { return b.ID }
func (b *StatisticalAnalyzerBlock) GetType() core.LogicBlockType { return b.Type }
func (b *StatisticalAnalyzerBlock) GetConfigure() map[string]any { return b.RawConfig }

func (b *StatisticalAnalyzerBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 验证统计指标
	validMetrics := map[string]bool{
		"mean": true, "std": true, "max": true, "min": true, "median": true, "variance": true,
	}
	for _, metric := range cfg.Metrics {
		if !validMetrics[metric] {
			return fmt.Errorf("不支持的统计指标: %s (支持: mean, std, max, min, median, variance)", metric)
		}
	}

	b.TypedConfig = &cfg
	return nil
}

// Execute 执行统计分析逻辑
func (b *StatisticalAnalyzerBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	config, ok := b.TypedConfig.(*Config)
	if !ok {
		return false, core.ErrInvalidConfig
	}

	// 从GraphContext读取数据源
	tenantID := execCtx.GetTenantId()
	graphKey := execCtx.GetGraphKey()

	graphContext, err := datastore.GetGraphContext(ctx, tenantID, graphKey, config.SourceKey)
	if err != nil {
		logx.Errorf("读取GraphContext失败: %v", err)
		return false, fmt.Errorf("读取GraphContext失败: %w", err)
	}

	// 解析数据源（支持单字段和多字段场景）
	datasets, err := parseDataSource(graphContext.GetPayload(), config.FieldNames)
	if err != nil {
		logx.Errorf("解析数据源失败: %v", err)
		return false, fmt.Errorf("解析数据源失败: %w", err)
	}

	// 计算统计特征
	results := make(map[string]interface{})

	if len(datasets) == 1 {
		// 单字段场景：直接返回统计结果
		dataset := datasets[0]
		stats, err := calculateStatistics(dataset.GetValues(), config.Metrics)
		if err != nil {
			return false, err
		}

		results = map[string]interface{}{
			"fieldName": dataset.FieldName,
			"metrics":   stats,
			"count":     dataset.Len(),
		}

		logx.Infof("统计分析完成（单字段）- 租户: %s, 图: %s, 字段: %s, 记录数: %d",
			tenantID, graphKey.ID, dataset.FieldName, dataset.Len())
	} else {
		// 多字段场景：返回map[fieldName]stats
		for _, dataset := range datasets {
			stats, err := calculateStatistics(dataset.GetValues(), config.Metrics)
			if err != nil {
				return false, err
			}

			results[dataset.FieldName] = map[string]interface{}{
				"metrics": stats,
				"count":   dataset.Len(),
			}
		}

		logx.Infof("统计分析完成（多字段）- 租户: %s, 图: %s, 字段数: %d",
			tenantID, graphKey.ID, len(datasets))
	}

	// 保存结果到GraphContext
	saveContext := &core.BaseGraphContext{
		TenantId:   tenantID,
		GraphKey:   graphKey,
		ContextKey: config.SaveToContext,
		Payload: map[string]any{
			"data": results,
		},
	}

	if err := datastore.SaveGraphContext(ctx, saveContext); err != nil {
		logx.Errorf("保存统计结果到GraphContext失败: %v", err)
		return false, fmt.Errorf("保存统计结果到GraphContext失败: %w", err)
	}

	// 成功执行，继续后续节点
	return true, nil
}

// parseDataSource 解析GraphContext中的数据源
//
// 支持两种格式：
//  1. 单字段：{"data": TimeSeriesDataset}
//  2. 多字段：{"data": map[string]TimeSeriesDataset}
func parseDataSource(payload map[string]any, fieldNames []string) ([]core.TimeSeriesDataset, error) {
	data, ok := payload["data"]
	if !ok {
		return nil, fmt.Errorf("GraphContext中缺少 'data' 字段")
	}

	// 尝试解析为单字段TimeSeriesDataset
	if datasetMap, ok := data.(map[string]interface{}); ok {
		// 检查是否是TimeSeriesDataset（包含deviceId和fieldName）
		if _, hasDeviceID := datasetMap["deviceId"]; hasDeviceID {
			dataset, err := parseTimeSeriesDataset(datasetMap)
			if err != nil {
				return nil, err
			}
			return []core.TimeSeriesDataset{dataset}, nil
		}

		// 多字段场景：map[string]TimeSeriesDataset
		var datasets []core.TimeSeriesDataset
		for fieldName, fieldData := range datasetMap {
			// 如果指定了fieldNames，只处理指定的字段
			if len(fieldNames) > 0 && !contains(fieldNames, fieldName) {
				continue
			}

			fieldDataMap, ok := fieldData.(map[string]interface{})
			if !ok {
				continue
			}

			dataset, err := parseTimeSeriesDataset(fieldDataMap)
			if err != nil {
				logx.Errorf("解析字段 %s 失败: %v", fieldName, err)
				continue
			}

			datasets = append(datasets, dataset)
		}

		if len(datasets) == 0 {
			return nil, fmt.Errorf("未找到有效的TimeSeriesDataset数据")
		}

		return datasets, nil
	}

	return nil, fmt.Errorf("无法解析数据源格式")
}

// parseTimeSeriesDataset 解析map为TimeSeriesDataset
func parseTimeSeriesDataset(data map[string]interface{}) (core.TimeSeriesDataset, error) {
	dataset := core.TimeSeriesDataset{}

	// 解析deviceId
	if deviceID, ok := data["deviceId"].(string); ok {
		dataset.DeviceID = deviceID
	} else {
		return dataset, fmt.Errorf("缺少deviceId字段")
	}

	// 解析fieldName
	if fieldName, ok := data["fieldName"].(string); ok {
		dataset.FieldName = fieldName
	} else {
		return dataset, fmt.Errorf("缺少fieldName字段")
	}

	// 解析dataPoints
	if dataPointsRaw, ok := data["dataPoints"].([]interface{}); ok {
		dataset.DataPoints = make([]core.DataPoint, 0, len(dataPointsRaw))
		for _, dpRaw := range dataPointsRaw {
			if dpMap, ok := dpRaw.(map[string]interface{}); ok {
				dp := core.DataPoint{}
				if ts, ok := dpMap["timestamp"].(float64); ok {
					dp.Timestamp = int64(ts)
				}
				if val, ok := dpMap["value"].(float64); ok {
					dp.Value = val
				}
				dataset.DataPoints = append(dataset.DataPoints, dp)
			}
		}
	}

	// 解析metadata（可选）
	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		dataset.Metadata = metadata
	}

	return dataset, nil
}

// calculateStatistics 计算统计特征
func calculateStatistics(values []float64, metrics []string) (map[string]float64, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("数据为空，无法计算统计特征")
	}

	stats := make(map[string]float64)

	// 计算均值
	var mean float64
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	// 计算方差和标准差
	var variance float64
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))
	std := math.Sqrt(variance)

	// 计算最大值和最小值
	max := values[0]
	min := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}

	// 计算中位数
	sortedValues := make([]float64, len(values))
	copy(sortedValues, values)
	sort.Float64s(sortedValues)

	var median float64
	n := len(sortedValues)
	if n%2 == 0 {
		median = (sortedValues[n/2-1] + sortedValues[n/2]) / 2
	} else {
		median = sortedValues[n/2]
	}

	// 根据配置返回指定的统计指标
	for _, metric := range metrics {
		switch metric {
		case "mean":
			stats["mean"] = mean
		case "std":
			stats["std"] = std
		case "max":
			stats["max"] = max
		case "min":
			stats["min"] = min
		case "median":
			stats["median"] = median
		case "variance":
			stats["variance"] = variance
		}
	}

	return stats, nil
}

// contains 检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetStatisticalAnalyzerSpec 返回 StatisticalAnalyzerBlock 的规格
func GetStatisticalAnalyzerSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"StatisticalAnalyzer",
		StatisticalAnalyzerVersion,
		"统计分析积木：计算均值、标准差、最大值、最小值、中位数、方差等统计特征",
		[]string{"data", "analysis", "statistics"},
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"sourceKey": map[string]any{
					"type":        "string",
					"title":       "数据源Key",
					"description": "GraphContext中的数据源key（由FetchSensorData生成）",
				},
				"fieldNames": map[string]any{
					"type":        "array",
					"title":       "字段列表",
					"description": "要分析的字段列表（多字段场景，可选）",
					"items": map[string]any{
						"type": "string",
					},
				},
				"metrics": map[string]any{
					"type":        "array",
					"title":       "统计指标",
					"description": "统计指标：mean（均值）/std（标准差）/max（最大值）/min（最小值）/median（中位数）/variance（方差）",
					"items": map[string]any{
						"type": "string",
						"enum": []string{"mean", "std", "max", "min", "median", "variance"},
					},
				},
				"saveToContext": map[string]any{
					"type":        "string",
					"title":       "保存到Context的Key",
					"description": "保存结果的key",
				},
			},
			"required": []string{"sourceKey", "metrics", "saveToContext"},
		},
		[]core.ServiceType{}, // 无外部服务依赖
	)
}
