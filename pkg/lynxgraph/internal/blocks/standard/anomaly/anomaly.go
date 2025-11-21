package anomaly

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/zeromicro/go-zero/core/logx"
)

const AnomalyDetectorVersion = "v1"

// Config 异常检测积木的配置
type Config struct {
	SourceKey     string   `json:"sourceKey" check:"must"`     // GraphContext中的数据源key
	FieldNames    []string `json:"fieldNames"`                 // 要分析的字段列表（多字段场景，可选）
	Method        string   `json:"method" check:"must"`        // "zscore" 或 "iqr"
	Threshold     float64  `json:"threshold"`                  // Z-score阈值（默认3.0）
	IQRMultiplier float64  `json:"iqrMultiplier"`              // IQR倍数（默认1.5）
	SaveToContext string   `json:"saveToContext" check:"must"` // 保存结果的key
}

// AnomalyPoint 异常点信息
type AnomalyPoint struct {
	Timestamp int64   `json:"timestamp"` // 时间戳
	Value     float64 `json:"value"`     // 值
	Score     float64 `json:"score"`     // 异常分数（Z-score或IQR偏离倍数）
}

// AnomalyDetectorBlock 实现异常检测逻辑块
type AnomalyDetectorBlock struct {
	core.BaseLogicBlock
}

// NewAnomalyDetectorBlock 创建一个新的 AnomalyDetectorBlock 实例
func NewAnomalyDetectorBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &AnomalyDetectorBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeAnomalyDetector,
			RawConfig: config,
		},
	}, nil
}

func (b *AnomalyDetectorBlock) GetID() string                { return b.ID }
func (b *AnomalyDetectorBlock) GetType() core.LogicBlockType { return b.Type }
func (b *AnomalyDetectorBlock) GetConfigure() map[string]any { return b.RawConfig }

func (b *AnomalyDetectorBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 验证method
	if cfg.Method != "zscore" && cfg.Method != "iqr" {
		return fmt.Errorf("不支持的异常检测方法: %s (支持: zscore, iqr)", cfg.Method)
	}

	// 设置默认值
	if cfg.Threshold == 0 {
		cfg.Threshold = 3.0 // Z-score默认阈值
	}
	if cfg.IQRMultiplier == 0 {
		cfg.IQRMultiplier = 1.5 // IQR默认倍数
	}

	b.TypedConfig = &cfg
	return nil
}

// Execute 执行异常检测逻辑
func (b *AnomalyDetectorBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
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

	// 解析数据源
	datasets, err := parseDataSource(graphContext.GetPayload(), config.FieldNames)
	if err != nil {
		logx.Errorf("解析数据源失败: %v", err)
		return false, fmt.Errorf("解析数据源失败: %w", err)
	}

	// 异常检测
	results := make(map[string]interface{})

	if len(datasets) == 1 {
		// 单字段场景
		dataset := datasets[0]
		anomalyResult, err := detectAnomalies(dataset, config.Method, config.Threshold, config.IQRMultiplier)
		if err != nil {
			return false, err
		}

		results = anomalyResult

		logx.Infof("异常检测完成（单字段）- 租户: %s, 图: %s, 字段: %s, 方法: %s, 异常率: %.2f%%",
			tenantID, graphKey.ID, dataset.FieldName, config.Method, anomalyResult["anomalyScore"].(float64)*100)
	} else {
		// 多字段场景
		for _, dataset := range datasets {
			anomalyResult, err := detectAnomalies(dataset, config.Method, config.Threshold, config.IQRMultiplier)
			if err != nil {
				return false, err
			}

			results[dataset.FieldName] = anomalyResult
		}

		logx.Infof("异常检测完成（多字段）- 租户: %s, 图: %s, 字段数: %d",
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
		logx.Errorf("保存异常检测结果到GraphContext失败: %v", err)
		return false, fmt.Errorf("保存异常检测结果到GraphContext失败: %w", err)
	}

	// 成功执行，继续后续节点
	return true, nil
}

// detectAnomalies 检测异常值
func detectAnomalies(dataset core.TimeSeriesDataset, method string, threshold float64, iqrMultiplier float64) (map[string]interface{}, error) {
	if dataset.Len() == 0 {
		return nil, fmt.Errorf("数据为空，无法进行异常检测")
	}

	values := dataset.GetValues()
	timestamps := dataset.GetTimestamps()

	var anomalies []AnomalyPoint
	var stats map[string]float64

	if method == "zscore" {
		// Z-score方法
		mean, std := calculateMeanAndStd(values)
		stats = map[string]float64{
			"mean": mean,
			"std":  std,
		}

		for i, val := range values {
			zscore := math.Abs((val - mean) / std)
			if zscore > threshold {
				anomalies = append(anomalies, AnomalyPoint{
					Timestamp: timestamps[i],
					Value:     val,
					Score:     zscore,
				})
			}
		}
	} else {
		// IQR方法
		q1, q3, iqr := calculateIQR(values)
		lowerBound := q1 - iqrMultiplier*iqr
		upperBound := q3 + iqrMultiplier*iqr

		stats = map[string]float64{
			"q1":         q1,
			"q3":         q3,
			"iqr":        iqr,
			"lowerBound": lowerBound,
			"upperBound": upperBound,
		}

		for i, val := range values {
			if val < lowerBound || val > upperBound {
				// 计算偏离倍数作为score
				var score float64
				if val < lowerBound {
					score = (q1 - val) / iqr
				} else {
					score = (val - q3) / iqr
				}

				anomalies = append(anomalies, AnomalyPoint{
					Timestamp: timestamps[i],
					Value:     val,
					Score:     score,
				})
			}
		}
	}

	anomalyScore := float64(len(anomalies)) / float64(len(values))

	result := map[string]interface{}{
		"fieldName":    dataset.FieldName,
		"method":       method,
		"anomalyScore": anomalyScore,
		"anomalyCount": len(anomalies),
		"totalCount":   len(values),
		"anomalies":    anomalies,
		"stats":        stats,
	}

	return result, nil
}

// calculateMeanAndStd 计算均值和标准差
func calculateMeanAndStd(values []float64) (float64, float64) {
	// 计算均值
	var mean float64
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	// 计算标准差
	var variance float64
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))
	std := math.Sqrt(variance)

	return mean, std
}

// calculateIQR 计算四分位距
func calculateIQR(values []float64) (float64, float64, float64) {
	sortedValues := make([]float64, len(values))
	copy(sortedValues, values)
	sort.Float64s(sortedValues)

	n := len(sortedValues)

	// 计算Q1（第25百分位）
	q1Index := n / 4
	q1 := sortedValues[q1Index]

	// 计算Q3（第75百分位）
	q3Index := (n * 3) / 4
	q3 := sortedValues[q3Index]

	// 计算IQR
	iqr := q3 - q1

	return q1, q3, iqr
}

// parseDataSource 解析GraphContext中的数据源（与其他积木相同）
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

// contains 检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetAnomalyDetectorSpec 返回 AnomalyDetectorBlock 的规格
func GetAnomalyDetectorSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"AnomalyDetector",
		AnomalyDetectorVersion,
		"异常检测积木：基于Z-score或IQR方法检测异常值，返回异常分数和异常点列表",
		[]string{"data", "analysis", "anomaly"},
		map[string]any{
			"sourceKey": map[string]any{
				"type":        "string",
				"description": "GraphContext中的数据源key（由FetchSensorData生成）",
				"check":       "must",
			},
			"fieldNames": map[string]any{
				"type":        "array",
				"description": "要分析的字段列表（多字段场景，可选）",
				"check":       "",
				"items": map[string]any{
					"type": "string",
				},
			},
			"method": map[string]any{
				"type":        "string",
				"description": "异常检测方法：zscore（基于标准差）或 iqr（基于四分位距）",
				"check":       "must",
				"enum":        []string{"zscore", "iqr"},
			},
			"threshold": map[string]any{
				"type":        "number",
				"description": "Z-score阈值（默认3.0，仅zscore方法有效）",
				"check":       "",
			},
			"iqrMultiplier": map[string]any{
				"type":        "number",
				"description": "IQR倍数（默认1.5，仅iqr方法有效）",
				"check":       "",
			},
			"saveToContext": map[string]any{
				"type":        "string",
				"description": "保存结果的key",
				"check":       "must",
			},
		},
		[]core.ServiceType{}, // 无外部服务依赖
	)
}
