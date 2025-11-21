package trend

import (
	"context"
	"fmt"
	"math"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/zeromicro/go-zero/core/logx"
)

const TrendAnalyzerVersion = "v1"

// Config 趋势分析积木的配置
type Config struct {
	SourceKey            string   `json:"sourceKey" check:"must"`     // GraphContext中的数据源key
	FieldNames           []string `json:"fieldNames"`                 // 要分析的字段列表（多字段场景，可选）
	StableSlopeThreshold float64  `json:"stableSlopeThreshold"`       // 平稳趋势的斜率阈值（默认0.01）
	SaveToContext        string   `json:"saveToContext" check:"must"` // 保存结果的key
}

// TrendAnalyzerBlock 实现趋势分析逻辑块
type TrendAnalyzerBlock struct {
	core.BaseLogicBlock
}

// NewTrendAnalyzerBlock 创建一个新的 TrendAnalyzerBlock 实例
func NewTrendAnalyzerBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &TrendAnalyzerBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeTrendAnalyzer,
			RawConfig: config,
		},
	}, nil
}

func (b *TrendAnalyzerBlock) GetID() string                { return b.ID }
func (b *TrendAnalyzerBlock) GetType() core.LogicBlockType { return b.Type }
func (b *TrendAnalyzerBlock) GetConfigure() map[string]any { return b.RawConfig }

func (b *TrendAnalyzerBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 设置默认值
	if cfg.StableSlopeThreshold == 0 {
		cfg.StableSlopeThreshold = 0.01
	}

	b.TypedConfig = &cfg
	return nil
}

// Execute 执行趋势分析逻辑
func (b *TrendAnalyzerBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
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

	// 趋势分析
	results := make(map[string]interface{})

	if len(datasets) == 1 {
		// 单字段场景
		dataset := datasets[0]
		trendResult, err := analyzeTrend(dataset, config.StableSlopeThreshold)
		if err != nil {
			return false, err
		}

		results = trendResult

		logx.Infof("趋势分析完成（单字段）- 租户: %s, 图: %s, 字段: %s, 趋势: %s",
			tenantID, graphKey.ID, dataset.FieldName, trendResult["trend"])
	} else {
		// 多字段场景
		for _, dataset := range datasets {
			trendResult, err := analyzeTrend(dataset, config.StableSlopeThreshold)
			if err != nil {
				return false, err
			}

			results[dataset.FieldName] = trendResult
		}

		logx.Infof("趋势分析完成（多字段）- 租户: %s, 图: %s, 字段数: %d",
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
		logx.Errorf("保存趋势分析结果到GraphContext失败: %v", err)
		return false, fmt.Errorf("保存趋势分析结果到GraphContext失败: %w", err)
	}

	// 成功执行，继续后续节点
	return true, nil
}

// analyzeTrend 使用线性回归分析趋势
func analyzeTrend(dataset core.TimeSeriesDataset, stableThreshold float64) (map[string]interface{}, error) {
	if dataset.Len() < 2 {
		return nil, fmt.Errorf("数据点不足，无法进行趋势分析（至少需要2个点）")
	}

	timestamps := dataset.GetTimestamps()
	values := dataset.GetValues()

	// 线性回归计算（最小二乘法）
	// y = slope * x + intercept
	slope, intercept, rSquared := linearRegression(timestamps, values)

	// 判断趋势方向
	var trend string
	var confidence string

	if math.Abs(slope) < stableThreshold {
		trend = "stable"
	} else if slope > 0 {
		trend = "rising"
	} else {
		trend = "falling"
	}

	// 根据R²值判断置信度
	if rSquared >= 0.8 {
		confidence = "high"
	} else if rSquared >= 0.5 {
		confidence = "medium"
	} else {
		confidence = "low"
	}

	// 预测未来值（基于线性模型）
	lastTimestamp := timestamps[len(timestamps)-1]
	next1HourTimestamp := lastTimestamp + 3600   // 1小时后
	next24HourTimestamp := lastTimestamp + 86400 // 24小时后

	next1HourValue := slope*float64(next1HourTimestamp) + intercept
	next24HourValue := slope*float64(next24HourTimestamp) + intercept

	result := map[string]interface{}{
		"fieldName":  dataset.FieldName,
		"trend":      trend,
		"slope":      slope,
		"intercept":  intercept,
		"rSquared":   rSquared,
		"confidence": confidence,
		"prediction": map[string]interface{}{
			"next1hour":  next1HourValue,
			"next24hour": next24HourValue,
		},
		"dataPoints": dataset.Len(),
		"timeRange": map[string]interface{}{
			"start": timestamps[0],
			"end":   timestamps[len(timestamps)-1],
		},
	}

	return result, nil
}

// linearRegression 线性回归（最小二乘法）
//
// 返回：slope（斜率）、intercept（截距）、rSquared（R²值）
func linearRegression(x []int64, y []float64) (float64, float64, float64) {
	n := len(x)

	// 将int64时间戳转为float64（避免溢出）
	xFloat := make([]float64, n)
	for i, v := range x {
		xFloat[i] = float64(v)
	}

	// 计算均值
	var sumX, sumY float64
	for i := 0; i < n; i++ {
		sumX += xFloat[i]
		sumY += y[i]
	}
	meanX := sumX / float64(n)
	meanY := sumY / float64(n)

	// 计算斜率和截距
	var numerator, denominator float64
	for i := 0; i < n; i++ {
		numerator += (xFloat[i] - meanX) * (y[i] - meanY)
		denominator += (xFloat[i] - meanX) * (xFloat[i] - meanX)
	}

	var slope, intercept float64
	if denominator != 0 {
		slope = numerator / denominator
		intercept = meanY - slope*meanX
	}

	// 计算R²值（拟合优度）
	var ssRes, ssTot float64
	for i := 0; i < n; i++ {
		predicted := slope*xFloat[i] + intercept
		ssRes += (y[i] - predicted) * (y[i] - predicted)
		ssTot += (y[i] - meanY) * (y[i] - meanY)
	}

	var rSquared float64
	if ssTot != 0 {
		rSquared = 1 - (ssRes / ssTot)
	}

	return slope, intercept, rSquared
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

// GetTrendAnalyzerSpec 返回 TrendAnalyzerBlock 的规格
func GetTrendAnalyzerSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"TrendAnalyzer",
		TrendAnalyzerVersion,
		"趋势分析积木：使用线性回归分析数据趋势，判断上升/下降/平稳，并预测未来值",
		[]string{"data", "analysis", "trend"},
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
				"stableSlopeThreshold": map[string]any{
					"type":        "number",
					"title":       "平稳阈值",
					"description": "平稳趋势的斜率阈值（默认0.01），绝对值小于此阈值视为平稳",
					"default":     0.01,
				},
				"saveToContext": map[string]any{
					"type":        "string",
					"title":       "保存到Context的Key",
					"description": "保存结果的key",
				},
			},
			"required": []string{"sourceKey", "saveToContext"},
		},
		[]core.ServiceType{}, // 无外部服务依赖
	)
}
