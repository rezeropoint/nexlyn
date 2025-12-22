package filterbyset

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/zeromicro/go-zero/core/logx"
)

const FilterBySetVersion = "v1"

// Config FilterBySet 积木配置
type Config struct {
	SourceType    string `json:"sourceType" check:"must"`    // 数据来源: "context" | "atom"
	SourcePath    string `json:"sourcePath" check:"must"`    // 数组路径（如 "employees.data"）
	SetContextKey string `json:"setContextKey" check:"must"` // 集合数据位置（GetDedupSet 的 saveResultTo）
	MatchField    string `json:"matchField" check:"must"`    // 源数组元素的匹配字段（如 "name"）
	SetMatchField string `json:"setMatchField"`              // 集合中的匹配字段（可选，默认同 matchField）
	SaveResultTo  string `json:"saveResultTo" check:"must"`  // 过滤结果保存位置
	InvertFilter  *bool  `json:"invertFilter"`               // 反向过滤（保留存在的，默认 false）
}

// FilterBySetBlock 按集合过滤数组积木
type FilterBySetBlock struct {
	core.BaseLogicBlock
}

// NewFilterBySetBlock 创建一个新的 FilterBySetBlock 实例
func NewFilterBySetBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &FilterBySetBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeFilterBySet,
			RawConfig: config,
		},
	}, nil
}

func (b *FilterBySetBlock) GetID() string                { return b.ID }
func (b *FilterBySetBlock) GetType() core.LogicBlockType { return b.Type }
func (b *FilterBySetBlock) GetConfigure() map[string]any { return b.RawConfig }

func (b *FilterBySetBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 验证 sourceType
	if cfg.SourceType != "context" && cfg.SourceType != "atom" {
		return fmt.Errorf("sourceType 必须是 context 或 atom")
	}

	// 设置默认值
	if cfg.SetMatchField == "" {
		cfg.SetMatchField = cfg.MatchField
	}
	if cfg.InvertFilter == nil {
		defaultFalse := false
		cfg.InvertFilter = &defaultFalse
	}

	b.TypedConfig = &cfg
	return nil
}

// Execute 执行按集合过滤逻辑
func (b *FilterBySetBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	config, ok := b.TypedConfig.(*Config)
	if !ok {
		return false, core.ErrInvalidConfig
	}

	tenantId := execCtx.GetTenantId()
	graphKey := execCtx.GetGraphKey()

	// 1. 获取源数组
	sourceArray, err := b.getSourceArray(ctx, config, execCtx, datastore)
	if err != nil {
		return false, fmt.Errorf("[FilterBySet] 获取源数组失败: %w", err)
	}

	if len(sourceArray) == 0 {
		logx.WithContext(ctx).Infof("[FilterBySet] 源数组为空，跳过过滤")
		// 保存空结果
		return b.saveResult(ctx, datastore, tenantId, graphKey, config.SaveResultTo, []any{}, 0)
	}

	// 2. 获取集合数据（GetDedupSet 的输出）
	setData, err := b.getSetData(ctx, datastore, tenantId, graphKey, config.SetContextKey)
	if err != nil {
		return false, fmt.Errorf("[FilterBySet] 获取集合数据失败: %w", err)
	}

	// 3. 构建集合的 Set（用于快速查找）
	setValues := make(map[string]struct{})
	logx.WithContext(ctx).Debugf("[FilterBySet] 使用 setMatchField=%s 构建集合", config.SetMatchField)
	for i, record := range setData {
		if i < 3 {
			logx.WithContext(ctx).Debugf("[FilterBySet] record[%d] 字段: %v", i, getMapKeys(record))
		}
		if value, ok := record[config.SetMatchField]; ok {
			setValues[value] = struct{}{}
			if i < 3 {
				logx.WithContext(ctx).Debugf("[FilterBySet] 添加到集合: %s", value)
			}
		} else if i < 3 {
			logx.WithContext(ctx).Debugf("[FilterBySet] record[%d] 不包含字段 %s", i, config.SetMatchField)
		}
	}

	logx.WithContext(ctx).Infof("[FilterBySet] 源数组: %d 条, 集合: %d 条 (setMatchField=%s)", len(sourceArray), len(setValues), config.SetMatchField)

	// 4. 遍历源数组，过滤匹配/不匹配的元素
	invertFilter := config.InvertFilter != nil && *config.InvertFilter
	var filteredArray []any
	filteredCount := 0

	for _, item := range sourceArray {
		itemMap, ok := toStringAnyMap(item)
		if !ok {
			continue
		}

		matchValue := getFieldValue(itemMap, config.MatchField)
		_, existsInSet := setValues[matchValue]

		// invertFilter=false: 排除存在于集合中的（保留不存在的）
		// invertFilter=true: 保留存在于集合中的（排除不存在的）
		shouldKeep := existsInSet == invertFilter
		if shouldKeep {
			filteredArray = append(filteredArray, item)
		} else {
			filteredCount++
		}
	}

	logx.WithContext(ctx).Infof("[FilterBySet] 过滤完成: 保留 %d 条, 过滤掉 %d 条", len(filteredArray), filteredCount)

	// 5. 保存结果到 GraphContext
	return b.saveResult(ctx, datastore, tenantId, graphKey, config.SaveResultTo, filteredArray, filteredCount)
}

// getSourceArray 从 context 或 atom 获取源数组
func (b *FilterBySetBlock) getSourceArray(ctx context.Context, config *Config, execCtx core.ExecutionContext, datastore core.Store) ([]any, error) {
	var rawValue any

	if config.SourceType == "atom" {
		// 从 InfoAtom payload 获取
		payload := execCtx.GetInfoAtom().GetPayload()
		rawValue = getNestedValue(payload, config.SourcePath)
	} else {
		// 从 GraphContext 获取
		// sourcePath 格式: "contextKey.field1.field2" 或 "contextKey"
		parts := strings.SplitN(config.SourcePath, ".", 2)
		contextKey := parts[0]

		graphContext, err := datastore.GetGraphContext(
			ctx,
			execCtx.GetTenantId(),
			execCtx.GetGraphKey(),
			contextKey,
		)
		if err != nil {
			return nil, fmt.Errorf("获取 GraphContext %s 失败: %w", contextKey, err)
		}

		payload := graphContext.GetPayload()
		if len(parts) > 1 {
			rawValue = getNestedValue(payload, parts[1])
		} else {
			rawValue = payload
		}
	}

	if rawValue == nil {
		return nil, nil
	}

	// 转换为 []any
	switch v := rawValue.(type) {
	case []any:
		return v, nil
	case []map[string]any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, nil
	case []string:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, nil
	default:
		return nil, fmt.Errorf("sourcePath %s 的值不是数组类型: %T", config.SourcePath, rawValue)
	}
}

// getSetData 获取集合数据（GetDedupSet 的输出）
func (b *FilterBySetBlock) getSetData(ctx context.Context, datastore core.Store, tenantId string, graphKey core.GraphKey, setContextKey string) ([]map[string]string, error) {
	graphContext, err := datastore.GetGraphContext(ctx, tenantId, graphKey, setContextKey)
	if err != nil {
		if err == core.ErrItemNotFound {
			// 集合不存在，视为空集合
			logx.WithContext(ctx).Infof("[FilterBySet] 集合数据不存在: %s，视为空集合", setContextKey)
			return nil, nil
		}
		return nil, err
	}

	payload := graphContext.GetPayload()
	if payload == nil {
		logx.WithContext(ctx).Infof("[FilterBySet] 集合 payload 为空: %s", setContextKey)
		return nil, nil
	}

	// GetDedupSet 输出格式: { "records": [{"person_name": "张三"}, ...], ... }
	recordsRaw, ok := payload["records"]
	if !ok {
		logx.WithContext(ctx).Infof("[FilterBySet] 集合数据缺少 records 字段: %s，视为空集合", setContextKey)
		return nil, nil
	}

	logx.WithContext(ctx).Debugf("[FilterBySet] records 原始类型: %T", recordsRaw)

	// 转换为 []map[string]string，支持多种类型
	result := make([]map[string]string, 0)

	switch records := recordsRaw.(type) {
	case []any:
		logx.WithContext(ctx).Debugf("[FilterBySet] 匹配类型 []any, 长度: %d", len(records))
		for _, recordRaw := range records {
			if record, ok := toStringMap(recordRaw); ok {
				result = append(result, record)
			}
		}
	case []map[string]string:
		logx.WithContext(ctx).Debugf("[FilterBySet] 匹配类型 []map[string]string, 长度: %d", len(records))
		result = records
	case []map[string]any:
		logx.WithContext(ctx).Debugf("[FilterBySet] 匹配类型 []map[string]any, 长度: %d", len(records))
		for _, record := range records {
			strMap := make(map[string]string)
			for k, v := range record {
				strMap[k] = fmt.Sprintf("%v", v)
			}
			result = append(result, strMap)
		}
	default:
		// 尝试使用反射处理未知类型
		logx.WithContext(ctx).Infof("[FilterBySet] records 类型未直接匹配: %T，尝试反射处理", recordsRaw)
		result = convertToStringMapSlice(recordsRaw)
		if len(result) == 0 {
			logx.WithContext(ctx).Errorf("[FilterBySet] records 字段类型不支持: %T", recordsRaw)
		}
	}

	logx.WithContext(ctx).Debugf("[FilterBySet] 解析集合数据: %d 条记录", len(result))
	return result, nil
}

// saveResult 保存过滤结果到 GraphContext
func (b *FilterBySetBlock) saveResult(ctx context.Context, datastore core.Store, tenantId string, graphKey core.GraphKey, saveResultTo string, filteredArray []any, filteredCount int) (bool, error) {
	resultPayload := map[string]any{
		"data":          filteredArray,
		"count":         len(filteredArray),
		"filteredCount": filteredCount,
	}

	newContext := &core.BaseGraphContext{
		TenantId:   tenantId,
		GraphKey:   graphKey,
		ContextKey: saveResultTo,
		Payload:    resultPayload,
	}

	if err := datastore.SaveGraphContext(ctx, newContext); err != nil {
		return false, fmt.Errorf("[FilterBySet] 保存结果失败: %w", err)
	}

	return true, nil
}

// getNestedValue 从 map 中获取嵌套值
func getNestedValue(data map[string]any, path string) any {
	if data == nil || path == "" {
		return data
	}

	parts := strings.Split(path, ".")
	current := any(data)

	for _, part := range parts {
		if m, ok := current.(map[string]any); ok {
			current = m[part]
		} else if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}

	return current
}

// toStringAnyMap 将 any 转换为 map[string]any
func toStringAnyMap(v any) (map[string]any, bool) {
	if m, ok := v.(map[string]any); ok {
		return m, true
	}
	return nil, false
}

// toStringMap 将 any 转换为 map[string]string
func toStringMap(v any) (map[string]string, bool) {
	switch m := v.(type) {
	case map[string]string:
		return m, true
	case map[string]any:
		result := make(map[string]string)
		for k, val := range m {
			result[k] = fmt.Sprintf("%v", val)
		}
		return result, true
	default:
		return nil, false
	}
}

// getFieldValue 从 map 中获取字段值（支持嵌套路径）
func getFieldValue(data map[string]any, field string) string {
	if data == nil {
		return ""
	}

	// 支持嵌套路径（如 "employee.name"）
	parts := strings.Split(field, ".")
	current := any(data)

	for _, part := range parts {
		if m, ok := current.(map[string]any); ok {
			current = m[part]
		} else {
			return ""
		}
	}

	if current == nil {
		return ""
	}

	return fmt.Sprintf("%v", current)
}

// getMapKeys 获取 map 的所有 key
func getMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// convertToStringMapSlice 使用反射将未知切片类型转换为 []map[string]string
func convertToStringMapSlice(v any) []map[string]string {
	if v == nil {
		return nil
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return nil
	}

	result := make([]map[string]string, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		strMap := convertToStringMap(elem)
		if strMap != nil {
			result = append(result, strMap)
		}
	}
	return result
}

// convertToStringMap 使用反射将未知 map 类型转换为 map[string]string
func convertToStringMap(v any) map[string]string {
	if v == nil {
		return nil
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Map {
		return nil
	}

	// 检查 key 是否是 string 类型
	if rv.Type().Key().Kind() != reflect.String {
		return nil
	}

	result := make(map[string]string)
	for _, key := range rv.MapKeys() {
		val := rv.MapIndex(key)
		result[key.String()] = fmt.Sprintf("%v", val.Interface())
	}
	return result
}

// GetFilterBySetSpec 返回 FilterBySetBlock 的规格
func GetFilterBySetSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"FilterBySet",
		FilterBySetVersion,
		"按集合过滤数组，排除（或保留）存在于集合中的元素",
		[]string{"input", "output"},
		map[string]any{
			"sourceType": map[string]any{
				"type":        "string",
				"title":       "数据来源",
				"description": "选择从 GraphContext 还是 InfoAtom 中获取数组",
				"enum":        []string{"context", "atom"},
				"enumLabels":  []string{"图上下文", "信息原子"},
				"check":       "must",
				"default":     "context",
			},
			"sourcePath": map[string]any{
				"type":        "string",
				"title":       "数组路径",
				"description": "数组的访问路径，如 'employees.data'",
				"check":       "must",
				"placeholder": "employees.data",
			},
			"setContextKey": map[string]any{
				"type":        "string",
				"title":       "集合数据位置",
				"description": "GetDedupSet 积木的 saveResultTo 值",
				"check":       "must",
				"source":      "contextKeys",
				"placeholder": "checked_employees",
			},
			"matchField": map[string]any{
				"type":        "string",
				"title":       "源数组匹配字段",
				"description": "源数组元素中用于匹配的字段名",
				"check":       "must",
				"placeholder": "name",
			},
			"setMatchField": map[string]any{
				"type":        "string",
				"title":       "集合匹配字段",
				"description": "集合 records 中用于匹配的字段名（默认与源数组相同）",
				"placeholder": "person_name",
			},
			"saveResultTo": map[string]any{
				"type":        "string",
				"title":       "结果保存位置",
				"description": "过滤结果保存到 GraphContext 的 key",
				"check":       "must",
				"placeholder": "filtered_employees",
			},
			"invertFilter": map[string]any{
				"type":        "boolean",
				"title":       "反向过滤",
				"description": "勾选后保留存在于集合中的元素（而非排除）",
				"default":     false,
			},
		},
		[]core.ServiceType{},
	)
}
