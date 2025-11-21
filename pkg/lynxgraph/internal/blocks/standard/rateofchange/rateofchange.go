package rateofchange

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/zeromicro/go-zero/core/logx"
)

const RateOfChangeVersion = "v1"

// Config 变化率积木的配置
type Config struct {
	SourceKey     string   `json:"sourceKey" check:"must"`     // GraphContext中的数据源key
	FieldNames    []string `json:"fieldNames"`                 // 要分析的字段列表（多字段场景，可选）
	RateType      string   `json:"rateType" check:"must"`      // "absolute"（绝对变化） 或 "relative"（相对变化）
	TimeWindows   []string `json:"timeWindows"`                // 时间窗口标识（多窗口模式，可选）
	SaveToContext string   `json:"saveToContext" check:"must"` // 保存结果的key
}

// RateOfChangeBlock 实现变化率计算逻辑块
type RateOfChangeBlock struct {
	core.BaseLogicBlock
}

// NewRateOfChangeBlock 创建一个新的 RateOfChangeBlock 实例
func NewRateOfChangeBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &RateOfChangeBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeRateOfChange,
			RawConfig: config,
		},
	}, nil
}

func (b *RateOfChangeBlock) GetID() string                { return b.ID }
func (b *RateOfChangeBlock) GetType() core.LogicBlockType { return b.Type }
func (b *RateOfChangeBlock) GetConfigure() map[string]any { return b.RawConfig }

func (b *RateOfChangeBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 验证rateType
	if cfg.RateType != "absolute" && cfg.RateType != "relative" {
		return fmt.Errorf("不支持的变化率类型: %s (支持: absolute, relative)", cfg.RateType)
	}

	b.TypedConfig = &cfg
	return nil
}

// Execute 执行变化率计算逻辑
func (b *RateOfChangeBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	config, ok := b.TypedConfig.(*Config)
	if !ok {
		return false, core.ErrInvalidConfig
	}

	// 从GraphContext读取数据源
	tenantID := execCtx.GetTenantId()
	graphKey := execCtx.GetGraphKey()

	// 判断是单时间窗口还是多时间窗口模式
	if len(config.TimeWindows) == 0 {
		// 单时间窗口模式：直接从sourceKey读取数据
		return b.executeSingleWindow(ctx, execCtx, datastore, config, tenantID, graphKey)
	} else {
		// 多时间窗口模式：从多个sourceKey读取数据并对比
		return b.executeMultiWindow(ctx, execCtx, datastore, config, tenantID, graphKey)
	}
}

// executeSingleWindow 单时间窗口模式：计算首尾值的变化率
func (b *RateOfChangeBlock) executeSingleWindow(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, config *Config, tenantID string, graphKey core.GraphKey) (bool, error) {
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

	// 计算变化率
	results := make(map[string]interface{})

	if len(datasets) == 1 {
		// 单字段场景
		dataset := datasets[0]
		rate, err := calculateRate(dataset, config.RateType)
		if err != nil {
			return false, err
		}

		results = rate

		logx.Infof("变化率计算完成（单字段）- 租户: %s, 图: %s, 字段: %s, 类型: %s",
			tenantID, graphKey.ID, dataset.FieldName, config.RateType)
	} else {
		// 多字段场景
		for _, dataset := range datasets {
			rate, err := calculateRate(dataset, config.RateType)
			if err != nil {
				return false, err
			}

			results[dataset.FieldName] = rate
		}

		logx.Infof("变化率计算完成（多字段）- 租户: %s, 图: %s, 字段数: %d",
			tenantID, graphKey.ID, len(datasets))
	}

	// 保存结果到GraphContext
	return b.saveResults(ctx, datastore, tenantID, graphKey, config.SaveToContext, results)
}

// executeMultiWindow 多时间窗口模式：对比多个时间窗口的变化率
func (b *RateOfChangeBlock) executeMultiWindow(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, config *Config, tenantID string, graphKey core.GraphKey) (bool, error) {
	// 从多个GraphContext读取数据
	windowDatasets := make(map[string][]core.TimeSeriesDataset)

	for _, window := range config.TimeWindows {
		// 构建sourceKey：{sourceKey}_{window}（例如：temp_data_5min）
		sourceKey := fmt.Sprintf("%s_%s", config.SourceKey, window)

		graphContext, err := datastore.GetGraphContext(ctx, tenantID, graphKey, sourceKey)
		if err != nil {
			logx.Errorf("读取时间窗口 %s 的GraphContext失败: %v", window, err)
			continue
		}

		datasets, err := parseDataSource(graphContext.GetPayload(), config.FieldNames)
		if err != nil {
			logx.Errorf("解析时间窗口 %s 的数据源失败: %v", window, err)
			continue
		}

		windowDatasets[window] = datasets
	}

	if len(windowDatasets) == 0 {
		return false, fmt.Errorf("未找到任何时间窗口的有效数据")
	}

	// 计算每个时间窗口的变化率并对比
	results := make(map[string]interface{})

	// 获取字段名列表（从第一个时间窗口）
	var firstDatasets []core.TimeSeriesDataset
	for _, datasets := range windowDatasets {
		firstDatasets = datasets
		break
	}

	if len(firstDatasets) == 1 {
		// 单字段场景
		fieldName := firstDatasets[0].FieldName
		rates := make(map[string]interface{})

		for window, datasets := range windowDatasets {
			for _, dataset := range datasets {
				if dataset.FieldName == fieldName {
					rate, err := calculateRate(dataset, config.RateType)
					if err != nil {
						logx.Errorf("计算时间窗口 %s 的变化率失败: %v", window, err)
						continue
					}
					rates[window] = rate
				}
			}
		}

		results = map[string]interface{}{
			"fieldName": fieldName,
			"rates":     rates,
		}

		logx.Infof("多时间窗口变化率计算完成（单字段）- 租户: %s, 图: %s, 字段: %s, 窗口数: %d",
			tenantID, graphKey.ID, fieldName, len(rates))
	} else {
		// 多字段场景
		for _, dataset := range firstDatasets {
			fieldName := dataset.FieldName
			rates := make(map[string]interface{})

			for window, datasets := range windowDatasets {
				for _, ds := range datasets {
					if ds.FieldName == fieldName {
						rate, err := calculateRate(ds, config.RateType)
						if err != nil {
							logx.Errorf("计算时间窗口 %s 的变化率失败: %v", window, err)
							continue
						}
						rates[window] = rate
					}
				}
			}

			results[fieldName] = map[string]interface{}{
				"rates": rates,
			}
		}

		logx.Infof("多时间窗口变化率计算完成（多字段）- 租户: %s, 图: %s, 字段数: %d, 窗口数: %d",
			tenantID, graphKey.ID, len(results), len(config.TimeWindows))
	}

	// 保存结果到GraphContext
	return b.saveResults(ctx, datastore, tenantID, graphKey, config.SaveToContext, results)
}

// calculateRate 计算变化率
func calculateRate(dataset core.TimeSeriesDataset, rateType string) (map[string]interface{}, error) {
	if dataset.Len() < 2 {
		return nil, fmt.Errorf("数据点不足，无法计算变化率（至少需要2个点）")
	}

	// 获取首尾值
	startPoint := dataset.DataPoints[0]
	endPoint := dataset.DataPoints[dataset.Len()-1]

	startValue := startPoint.Value
	endValue := endPoint.Value
	duration := endPoint.Timestamp - startPoint.Timestamp

	// 计算绝对变化和相对变化
	absoluteChange := endValue - startValue
	var relativeChange float64
	if startValue != 0 {
		relativeChange = absoluteChange / startValue
	}

	result := map[string]interface{}{
		"fieldName":  dataset.FieldName,
		"rateType":   rateType,
		"startValue": startValue,
		"endValue":   endValue,
		"duration":   duration,
		"startTime":  startPoint.Timestamp,
		"endTime":    endPoint.Timestamp,
		"dataPoints": dataset.Len(),
	}

	if rateType == "absolute" {
		result["rate"] = absoluteChange
		result["absolute"] = absoluteChange
	} else {
		result["rate"] = relativeChange
		result["relative"] = relativeChange
	}

	return result, nil
}

// saveResults 保存结果到GraphContext
func (b *RateOfChangeBlock) saveResults(ctx context.Context, datastore core.Store, tenantID string, graphKey core.GraphKey, contextKey string, results map[string]interface{}) (bool, error) {
	saveContext := &core.BaseGraphContext{
		TenantId:   tenantID,
		GraphKey:   graphKey,
		ContextKey: contextKey,
		Payload: map[string]any{
			"data": results,
		},
	}

	if err := datastore.SaveGraphContext(ctx, saveContext); err != nil {
		logx.Errorf("保存变化率结果到GraphContext失败: %v", err)
		return false, fmt.Errorf("保存变化率结果到GraphContext失败: %w", err)
	}

	// 成功执行，继续后续节点
	return true, nil
}

// parseDataSource 解析GraphContext中的数据源（与StatisticalAnalyzer相同）
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

// GetRateOfChangeSpec 返回 RateOfChangeBlock 的规格
func GetRateOfChangeSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"RateOfChange",
		RateOfChangeVersion,
		"变化率计算积木：计算绝对/相对变化率，支持单时间窗口和多时间窗口对比",
		[]string{"data", "analysis", "rate"},
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
				"rateType": map[string]any{
					"type":        "string",
					"title":       "变化率类型",
					"description": "变化率类型：absolute（绝对变化）或 relative（相对变化，百分比）",
					"enum":        []string{"absolute", "relative"},
					"default":     "absolute",
				},
				"timeWindows": map[string]any{
					"type":        "array",
					"title":       "时间窗口",
					"description": "时间窗口标识列表（多窗口模式，如：['5min', '1hour', '24hour']），可选",
					"items": map[string]any{
						"type": "string",
					},
				},
				"saveToContext": map[string]any{
					"type":        "string",
					"title":       "保存到Context的Key",
					"description": "保存结果的key",
				},
			},
			"required": []string{"sourceKey", "rateType", "saveToContext"},
		},
		[]core.ServiceType{}, // 无外部服务依赖
	)
}
