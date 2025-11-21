package queryhistorydata

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/rezeropoint/nexlyn/service/iotquery/iotquery"

	"github.com/zeromicro/go-zero/core/logx"
)

const QueryHistoryDataVersion = "v1"

// Config 历史数据查询积木的配置
type Config struct {
	ServiceName   string   `json:"serviceName" check:"must"` // 服务名称（必须是 "sensor_data"）
	DeviceIDs     []string `json:"deviceIds"`                // 设备ID列表（可选，为空则查询组织下所有设备）
	FieldNames    []string `json:"fieldNames" check:"must"`  // 查询字段列表（必填，最多10个）
	TimeWindowSec int64    `json:"timeWindowSec"`            // 时间窗口（秒，相对于当前时间向前回溯），可选
	// 可选指定绝对时间范围（Unix秒）。如果同时指定了 StartTime/EndTime，则优先使用它们。
	StartTime int64 `json:"startTime"` // 可选，Unix秒
	EndTime   int64 `json:"endTime"`   // 可选，Unix秒

	Aggregation string `json:"aggregation"` // 聚合类型：空/"avg"/"max"/"min"/"sum"/"count"/"last"
	IntervalSec int64  `json:"intervalSec"` // 聚合时间间隔（秒），仅aggregation非空时有效

	Limit    int32  `json:"limit"`    // 限制返回条数（默认1000，最大10000）
	Offset   int32  `json:"offset"`   // 偏移量
	OrderBy  string `json:"orderBy"`  // 排序字段："timestamp"/"device_id"
	OrderDir string `json:"orderDir"` // 排序方向："asc"/"desc"

	SaveToContext string   `json:"saveToContext" check:"must"` // 保存到GraphContext的key
	OrgIDs        []string `json:"orgIds"`                     // 组织ID列表（可选，从ExecutionContext获取）
}

// QueryHistoryDataBlock 实现历史数据查询逻辑块
type QueryHistoryDataBlock struct {
	core.BaseLogicBlock
}

// NewQueryHistoryDataBlock 创建一个新的 QueryHistoryDataBlock 实例
func NewQueryHistoryDataBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &QueryHistoryDataBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeFetchSensorData,
			RawConfig: config,
		},
	}, nil
}

func (b *QueryHistoryDataBlock) GetID() string                { return b.ID }
func (b *QueryHistoryDataBlock) GetType() core.LogicBlockType { return b.Type }
func (b *QueryHistoryDataBlock) GetConfigure() map[string]any { return b.RawConfig }

func (b *QueryHistoryDataBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 验证服务名称
	if cfg.ServiceName != "sensor_data" {
		return fmt.Errorf("不支持的服务名称: %s (仅支持 sensor_data)", cfg.ServiceName)
	}

	// 验证字段数量
	if len(cfg.FieldNames) == 0 {
		return fmt.Errorf("查询字段列表不能为空")
	}
	if len(cfg.FieldNames) > 10 {
		return fmt.Errorf("查询字段数量不能超过10个")
	}

	// 验证时间范围：必须提供 timeWindowSec 或 start/end
	if cfg.TimeWindowSec <= 0 && !(cfg.StartTime > 0 && cfg.EndTime > 0) {
		return fmt.Errorf("需要指定时间窗口(timeWindowSec)或绝对时间范围(startTime/endTime)")
	}

	// 设置默认值
	if cfg.Limit <= 0 {
		cfg.Limit = 1000
	}
	if cfg.Limit > 10000 {
		cfg.Limit = 10000
	}

	b.TypedConfig = &cfg
	return nil
}

// Execute 执行历史数据查询逻辑
func (b *QueryHistoryDataBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	config, ok := b.TypedConfig.(*Config)
	if !ok {
		return false, core.ErrInvalidConfig
	}

	// 从Service获取IoTQuery gRPC客户端（类似Skylark积木获取SkylarkEngine）
	iotQueryServiceRaw, err := service.GetByType(core.ServiceTypeSensorData)
	if err != nil {
		return false, fmt.Errorf("获取传感器数据服务失败: %w", err)
	}

	iotQueryClient, ok := iotQueryServiceRaw.(iotquery.IoTQuery)
	if !ok {
		return false, fmt.Errorf("服务类型不匹配，期望iotquery.IoTQuery")
	}

	// 构建查询请求
	tenantID := execCtx.GetTenantId()
	now := time.Now().Unix()

	var startTime int64
	var endTime int64
	if config.StartTime > 0 && config.EndTime > 0 {
		startTime = config.StartTime
		endTime = config.EndTime
	} else if config.TimeWindowSec > 0 {
		endTime = now
		startTime = now - config.TimeWindowSec
	} else {
		return false, fmt.Errorf("无法确定查询时间范围")
	}

	// 获取组织ID列表（如果配置中未指定，可以从ExecutionContext或其他地方获取）
	orgIDs := config.OrgIDs

	// 默认排序及偏移
	offset := config.Offset
	orderBy := config.OrderBy
	orderDir := config.OrderDir
	if orderBy == "" {
		orderBy = "timestamp"
	}
	if orderDir == "" {
		orderDir = "desc"
	}

	// 构建 gRPC 请求（直接使用 proto 定义的类型）
	grpcReq := &iotquery.QueryTimeSeriesReq{
		TenantId:        tenantID,
		OrgIds:          orgIDs,
		DeviceIds:       config.DeviceIDs,
		FieldNames:      config.FieldNames,
		StartTime:       startTime,
		EndTime:         endTime,
		Aggregation:     config.Aggregation,
		IntervalSeconds: config.IntervalSec,
		Limit:           config.Limit,
		Offset:          offset,
		OrderBy:         orderBy,
		OrderDir:        orderDir,
	}

	// 调用 gRPC 服务
	grpcResp, err := iotQueryClient.QueryTimeSeries(ctx, grpcReq)
	if err != nil {
		logx.Errorf("查询历史数据失败: %v", err)
		return false, fmt.Errorf("查询历史数据失败: %w", err)
	}

	// 将 gRPC 响应转换为标准的 TimeSeriesDataset 格式
	datasets := convertToTimeSeriesDatasets(grpcResp.Records)

	// 构建保存到GraphContext的数据
	graphKey := execCtx.GetGraphKey()
	var resultPayload interface{}

	// 如果只有一个字段，直接保存TimeSeriesDataset（简化后续积木访问）
	// 如果有多个字段，保存为map[string]TimeSeriesDataset（按fieldName索引）
	if len(config.FieldNames) == 1 && len(datasets) == 1 {
		// 单字段查询（推荐用法）
		resultPayload = datasets[0]
		logx.Infof("拉取传感器数据完成（单字段）- 租户: %s, 图: %s, 保存键: %s, 字段: %s, 记录数: %d",
			tenantID, graphKey.ID, config.SaveToContext, datasets[0].FieldName, datasets[0].Len())
	} else {
		// 多字段查询（批量优化场景）
		datasetMap := make(map[string]core.TimeSeriesDataset)
		for _, dataset := range datasets {
			datasetMap[dataset.FieldName] = dataset
		}
		resultPayload = datasetMap
		logx.Infof("拉取传感器数据完成（多字段）- 租户: %s, 图: %s, 保存键: %s, 字段数: %d, 总记录数: %d",
			tenantID, graphKey.ID, config.SaveToContext, len(datasets), len(grpcResp.Records))
	}

	// 保存到GraphContext（BaseGraphContext.Payload类型为map[string]any）
	// 统一包装为 {"data": resultPayload}，后续积木通过 Payload["data"] 访问
	graphContext := &core.BaseGraphContext{
		TenantId:   tenantID,
		GraphKey:   graphKey,
		ContextKey: config.SaveToContext,
		Payload: map[string]any{
			"data": resultPayload, // 单字段：TimeSeriesDataset，多字段：map[string]TimeSeriesDataset
		},
	}

	if err := datastore.SaveGraphContext(ctx, graphContext); err != nil {
		logx.Errorf("保存查询结果到GraphContext失败: %v", err)
		return false, fmt.Errorf("保存查询结果到GraphContext失败: %w", err)
	}

	// 成功执行，继续后续节点
	return true, nil
}

// convertToTimeSeriesDatasets 将gRPC响应转换为标准TimeSeriesDataset格式
//
// 设计说明：
//   - 按 (deviceId, fieldName) 分组，每组创建一个TimeSeriesDataset
//   - 推荐用法：每个FetchSensorData节点只查询一个字段（单一职责）
//   - 多字段场景：返回多个TimeSeriesDataset，调用方按fieldName索引访问
func convertToTimeSeriesDatasets(records []*iotquery.TimeSeriesRecord) []core.TimeSeriesDataset {
	// 按 (deviceId, fieldName) 分组
	groupKey := func(deviceID, fieldName string) string {
		return deviceID + "|" + fieldName
	}

	groups := make(map[string]*core.TimeSeriesDataset)

	for _, record := range records {
		key := groupKey(record.DeviceId, record.FieldName)

		// 如果分组不存在，创建新的TimeSeriesDataset
		if _, exists := groups[key]; !exists {
			groups[key] = &core.TimeSeriesDataset{
				DeviceID:   record.DeviceId,
				FieldName:  record.FieldName,
				DataPoints: make([]core.DataPoint, 0),
				Metadata: map[string]interface{}{
					"deviceModel":    record.DeviceModel,
					"deviceCategory": record.DeviceCategory,
				},
			}
		}

		// 解析数值（统一为float64）
		value := parseFloatValue(record)

		// 添加数据点
		groups[key].DataPoints = append(groups[key].DataPoints, core.DataPoint{
			Timestamp: record.Timestamp,
			Value:     value,
		})
	}

	// 转换为数组
	datasets := make([]core.TimeSeriesDataset, 0, len(groups))
	for _, dataset := range groups {
		datasets = append(datasets, *dataset)
	}

	return datasets
}

// parseFloatValue 解析字段值为float64
//
// 支持的格式（从 gRPC proto 的 value_json 字段）：
//   - {"value": 25.5} → 25.5
//   - {"value": "25.5"} → 25.5
//   - 25.5 → 25.5
//   - "25.5" → 25.5
//   - 其他格式 → 0.0（记录警告日志）
func parseFloatValue(record *iotquery.TimeSeriesRecord) float64 {
	// gRPC proto 的 TimeSeriesRecord 只有 value_json 字段
	if record.ValueJson == "" {
		logx.Errorf("字段值为空 - DeviceID: %s, FieldName: %s", record.DeviceId, record.FieldName)
		return 0.0
	}

	// 1. 尝试解析为 {"value": ...} 格式
	var valueMap map[string]interface{}
	if err := json.Unmarshal([]byte(record.ValueJson), &valueMap); err == nil {
		if val, ok := valueMap["value"]; ok {
			switch v := val.(type) {
			case float64:
				return v
			case int:
				return float64(v)
			case string:
				var f float64
				if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
					return f
				}
			}
		}
	}

	// 2. 尝试直接解析为 float64
	var f float64
	if err := json.Unmarshal([]byte(record.ValueJson), &f); err == nil {
		return f
	}

	// 3. 解析失败，记录警告并返回0.0
	logx.Errorf("无法解析字段值为float64 - DeviceID: %s, FieldName: %s, ValueJSON: %s",
		record.DeviceId, record.FieldName, record.ValueJson)
	return 0.0
}

// GetQueryHistoryDataSpec 返回 QueryHistoryDataBlock 的规格
func GetQueryHistoryDataSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"FetchSensorData",
		QueryHistoryDataVersion,
		"拉取传感器数据并保存到GraphContext，支持时间窗口/绝对时间/聚合/分页/排序",
		[]string{"data", "query"},
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"serviceName": map[string]any{
					"type":        "string",
					"title":       "服务名称",
					"description": "服务名称（必须是 'sensor_data'）",
					"enum":        []string{"sensor_data"},
					"default":     "sensor_data",
				},
				"deviceIds": map[string]any{
					"type":        "array",
					"title":       "设备ID列表",
					"description": "设备ID列表（可选，为空则查询组织下所有设备）",
					"items": map[string]any{
						"type": "string",
					},
				},
				"fieldNames": map[string]any{
					"type":        "array",
					"title":       "查询字段列表",
					"description": "查询字段列表（必填，最多10个）",
					"items": map[string]any{
						"type": "string",
					},
				},
				"timeWindowSec": map[string]any{
					"type":        "integer",
					"title":       "时间窗口",
					"description": "时间窗口（秒，相对于当前时间向前回溯），可选",
				},
				"startTime": map[string]any{
					"type":        "integer",
					"title":       "开始时间",
					"description": "绝对开始时间（Unix秒），可选，优先于 timeWindowSec",
				},
				"endTime": map[string]any{
					"type":        "integer",
					"title":       "结束时间",
					"description": "绝对结束时间（Unix秒），可选，优先于 timeWindowSec",
				},
				"aggregation": map[string]any{
					"type":        "string",
					"title":       "聚合类型",
					"description": "聚合类型：空/'avg'/'max'/'min'/'sum'/'count'/'last'",
					"enum":        []string{"", "avg", "max", "min", "sum", "count", "last"},
				},
				"intervalSec": map[string]any{
					"type":        "integer",
					"title":       "聚合时间间隔",
					"description": "聚合时间间隔（秒），仅aggregation非空时有效",
				},
				"limit": map[string]any{
					"type":        "integer",
					"title":       "返回条数限制",
					"description": "限制返回条数（默认1000，最大10000）",
					"default":     1000,
				},
				"offset": map[string]any{
					"type":        "integer",
					"title":       "偏移量",
					"description": "偏移量（可选）",
					"default":     0,
				},
				"orderBy": map[string]any{
					"type":        "string",
					"title":       "排序字段",
					"description": "排序字段：'timestamp'/'device_id'（可选）",
					"enum":        []string{"timestamp", "device_id"},
				},
				"orderDir": map[string]any{
					"type":        "string",
					"title":       "排序方向",
					"description": "排序方向：'asc'/'desc'（可选）",
					"enum":        []string{"asc", "desc"},
				},
				"saveToContext": map[string]any{
					"type":        "string",
					"title":       "保存到Context的Key",
					"description": "保存到GraphContext的key",
				},
				"orgIds": map[string]any{
					"type":        "array",
					"title":       "组织ID列表",
					"description": "组织ID列表（可选，从ExecutionContext获取）",
					"items": map[string]any{
						"type": "string",
					},
				},
			},
			"required": []string{"serviceName", "fieldNames", "saveToContext"},
		},
		[]core.ServiceType{core.ServiceTypeSensorData}, // 依赖传感器数据服务
	)
}
