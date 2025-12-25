package journeys

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/zeromicro/go-zero/core/logx"

	skylarkCore "github.com/rezeropoint/go-skylark/v2/core"
	"github.com/rezeropoint/go-skylark/v2/engine"
)

const SkylarkJourneyCreateVersion = "v1"

// DataMapping 字段映射配置
type DataMapping struct {
	Key   string `json:"key"`   // 目标字段名（Skylark 表单字段）
	Value string `json:"value"` // 源字段表达式，支持 {{atom.xxx}} 和 {{context.xxx.yyy}}
}

type Config struct {
	App        string        `json:"app" check:"must"`    // Skylark实例，一般是域名
	FlowId     int64         `json:"flowId" check:"must"` // 流程ID
	UserID     int64         `json:"userId" check:"must"` // 用户ID
	AuthHeader string        `json:"authHeader"`          // 认证头（可选，用于外部调用）
	Data       []DataMapping `json:"data"`                // 字段映射列表
}

// SkylarkJourneyCreateBlock 实现一个动作逻辑块
type SkylarkJourneyCreateBlock struct {
	core.BaseLogicBlock
}

// NewSkylarkJourneyCreateBlock 创建一个新的 SkylarkJourneyCreateBlock 实例
func NewSkylarkJourneyCreateBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &SkylarkJourneyCreateBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeSkylarkJourneyCreate,
			RawConfig: config,
		},
	}, nil
}

func (b *SkylarkJourneyCreateBlock) GetID() string                { return b.ID }
func (b *SkylarkJourneyCreateBlock) GetType() core.LogicBlockType { return b.Type }
func (b *SkylarkJourneyCreateBlock) GetConfigure() map[string]any { return b.RawConfig }

func (b *SkylarkJourneyCreateBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 兼容旧格式：如果 data 是 map[string]string 或 JSON 字符串，转换为 []DataMapping
	if rawData, ok := config["data"]; ok && cfg.Data == nil {
		cfg.Data = convertToDataMappings(rawData)
	}

	b.TypedConfig = &cfg
	return nil
}

// convertToDataMappings 将各种格式的 data 转换为 []DataMapping
// 支持：map[string]string、map[string]any、JSON 字符串
func convertToDataMappings(rawData any) []DataMapping {
	var result []DataMapping

	switch v := rawData.(type) {
	case []DataMapping:
		return v
	case []any:
		// 标准数组格式 [{key: "x", value: "y"}, ...]
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				key, _ := m["key"].(string)
				value, _ := m["value"].(string)
				if key != "" {
					result = append(result, DataMapping{Key: key, Value: value})
				}
			}
		}
	case map[string]any:
		// 旧格式 map[string]any
		for k, val := range v {
			if strVal, ok := val.(string); ok {
				result = append(result, DataMapping{Key: k, Value: strVal})
			}
		}
	case map[string]string:
		// 旧格式 map[string]string
		for k, val := range v {
			result = append(result, DataMapping{Key: k, Value: val})
		}
	case string:
		// JSON 字符串格式，尝试解析
		if v != "" {
			// 尝试解析为数组格式
			var arr []DataMapping
			if err := json.Unmarshal([]byte(v), &arr); err == nil {
				return arr
			}
			// 尝试解析为 map 格式
			var m map[string]string
			if err := json.Unmarshal([]byte(v), &m); err == nil {
				for k, val := range m {
					result = append(result, DataMapping{Key: k, Value: val})
				}
			}
		}
	}

	return result
}

// Execute 执行动作逻辑
func (b *SkylarkJourneyCreateBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	// 1. 从配置获取参数
	config, ok := b.TypedConfig.(*Config)
	if !ok {
		return false, core.ErrInvalidConfig
	}

	// 2. 从Service获取SkylarkEngine实例
	skylarkService, err := service.GetByType(core.ServiceTypeSkylarkEngine)
	if err != nil {
		return false, fmt.Errorf("获取Skylark服务失败: %w", err)
	}

	skylarkEngine, ok := skylarkService.(engine.SkylarkEngine)
	if !ok {
		return false, fmt.Errorf("服务类型不匹配，期望SkylarkEngine")
	}

	// 3. 从ExecutionContext直接获取InfoAtom
	infoAtom := execCtx.GetInfoAtom()

	// 4. 将InfoAtom的Payload转换为TypedValue格式
	var data map[string]skylarkCore.TypedValue
	if len(config.Data) > 0 {
		// 使用配置的字段映射（支持 {{atom.xxx}} 和 {{context.xxx}} 语法）
		data = convertWithMapping(ctx, config.Data, infoAtom, datastore, execCtx)
	} else {
		// 默认行为：使用全部 Payload
		data = convertPayloadToTypedValue(infoAtom)
	}

	// 5. 调用Skylark引擎创建流程
	logx.Infof("[SkylarkJourneyCreate] flowId=%d, mappings=%+v, payload=%+v", config.FlowId, config.Data, infoAtom.GetPayload())
	for k, v := range data {
		logx.Infof("[SkylarkJourneyCreate] field=%s, type=%s, valueType=%T, value=%v", k, v.Type, v.Value, v.Value)
	}
	err = skylarkEngine.CreateFlow(
		ctx,
		config.App,
		config.FlowId,
		config.UserID,
		config.AuthHeader,
		data,
	)
	if err != nil {
		return false, fmt.Errorf("创建Skylark流程失败: %w", err)
	}

	return true, nil
}

// GetSkylarkJourneyCreateSpec 返回 SkylarkJourneyCreateBlock 的规格
func GetSkylarkJourneyCreateSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"SkylarkJourneyCreate",
		SkylarkJourneyCreateVersion,
		"创建一个新的 Skylark 流程记录，将信息原子转换为 Skylark 流程记录",
		[]string{"action", "skylark", "workflow"},
		map[string]any{
			"app": map[string]any{
				"type":        "string",
				"title":       "Skylark 实例",
				"description": "Skylark 实例标识，一般是域名",
				"placeholder": "smp.example.com",
				"check":       "must",
			},
			"flowId": map[string]any{
				"type":        "integer",
				"title":       "流程 ID",
				"description": "要触发的 Skylark 流程 ID",
				"check":       "must",
			},
			"userId": map[string]any{
				"type":        "integer",
				"title":       "用户 ID",
				"description": "触发流程的用户 ID",
				"check":       "must",
			},
			"authHeader": map[string]any{
				"type":        "string",
				"title":       "认证头",
				"description": "Skylark API 认证头（可选，用于外部调用）",
				"placeholder": "Bearer xxx",
			},
			"data": map[string]any{
				"type":        "array",
				"title":       "字段映射",
				"description": "自定义字段映射列表。支持 {{atom.xxx}}（信息原子字段）和 {{context.xxx.yyy}}（图上下文字段）。不配置则使用全部 InfoAtom 数据",
				"items": map[string]any{
					"type":  "object",
					"title": "映射",
					"properties": map[string]any{
						"key": map[string]any{
							"type":        "string",
							"title":       "目标字段",
							"description": "Skylark 表单字段名",
							"placeholder": "name",
							"check":       "must",
						},
						"value": map[string]any{
							"type":        "string",
							"title":       "值表达式",
							"description": "支持 {{atom.xxx}} 或 {{context.xxx.yyy}}",
							"placeholder": "{{atom.person_name}}",
							"check":       "must",
						},
					},
					"required": []string{"key", "value"},
				},
			},
		},
		[]core.ServiceType{core.ServiceTypeSkylarkEngine},
	)
}

// convertPayloadToTypedValue 将InfoAtom的Payload转换为Skylark的TypedValue格式
// Skylark TypedValue.Type 只支持 string/imageURL/imageBase64，所有值都转为字符串
func convertPayloadToTypedValue(infoAtom core.InfoAtom) map[string]skylarkCore.TypedValue {
	payload := infoAtom.GetPayload()
	result := make(map[string]skylarkCore.TypedValue)

	for key, value := range payload {
		result[key] = skylarkCore.TypedValue{
			Type:  "string",
			Value: fmt.Sprintf("%v", value),
		}
	}

	return result
}

// convertWithMapping 根据配置的字段映射转换数据
// mappings: 字段映射列表，每项包含 key（目标字段名）和 value（源字段表达式）
// 支持 {{atom.xxx}}（信息原子字段）和 {{context.xxx.yyy}}（图上下文字段）
func convertWithMapping(ctx context.Context, mappings []DataMapping, infoAtom core.InfoAtom, datastore core.Store, execCtx core.ExecutionContext) map[string]skylarkCore.TypedValue {
	result := make(map[string]skylarkCore.TypedValue)
	payload := infoAtom.GetPayload()
	atomType := infoAtom.GetType()

	// 构建字段类型映射（用于确定 TypedValue 的类型）
	fieldTypes := make(map[string]core.FieldType)
	if atomType != nil {
		for _, field := range atomType.GetDataFormat().Fields {
			fieldTypes[field.FieldKey] = field.FieldType
		}
	}

	// 匹配 {{atom.xxx}} 格式的变量
	reAtom := regexp.MustCompile(`^\{\{atom\.(\w+)\}\}$`)
	// 匹配 {{context.xxx.yyy}} 格式的变量（支持多级路径）
	reContext := regexp.MustCompile(`^\{\{context\.([a-zA-Z0-9_.]+)\}\}$`)

	for _, mapping := range mappings {
		targetKey := mapping.Key
		sourceExpr := mapping.Value

		if targetKey == "" {
			continue
		}

		// 1. 检查是否是 {{atom.xxx}} 变量引用
		if match := reAtom.FindStringSubmatch(sourceExpr); match != nil {
			fieldName := match[1]
			if value, ok := payload[fieldName]; ok {
				// Skylark TypedValue.Type 只支持 string/imageURL/imageBase64
				// 默认使用 string，Value 必须转为字符串
				result[targetKey] = skylarkCore.TypedValue{
					Type:  "string",
					Value: fmt.Sprintf("%v", value),
				}
			}
			continue
		}

		// 2. 检查是否是 {{context.xxx.yyy}} 变量引用
		if match := reContext.FindStringSubmatch(sourceExpr); match != nil {
			path := match[1] // e.g., "late_time.lateMinutes"
			value, err := getValueFromContext(ctx, path, datastore, execCtx)
			if err == nil && value != nil {
				// Skylark TypedValue.Type 只支持 string/imageURL/imageBase64
				// 默认使用 string，Value 必须转为字符串
				result[targetKey] = skylarkCore.TypedValue{
					Type:  "string",
					Value: fmt.Sprintf("%v", value),
				}
			}
			continue
		}

		// 3. 字符串模板替换（支持混合变量）
		value := replaceVariablesExtended(ctx, sourceExpr, payload, datastore, execCtx)
		result[targetKey] = skylarkCore.TypedValue{
			Type:  "string",
			Value: value,
		}
	}

	return result
}

// replaceVariables 替换字符串中的 {{atom.xxx}} 变量
func replaceVariables(template string, payload map[string]any) string {
	re := regexp.MustCompile(`\{\{atom\.(\w+)\}\}`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		fieldName := re.FindStringSubmatch(match)[1]
		if value, ok := payload[fieldName]; ok {
			return fmt.Sprintf("%v", value)
		}
		return match
	})
}

// getValueFromContext 从 GraphContext 获取值
// path 格式: "contextKey.field" 或 "contextKey.field1.field2"
func getValueFromContext(ctx context.Context, path string, datastore core.Store, execCtx core.ExecutionContext) (any, error) {
	parts := strings.SplitN(path, ".", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("context 路径格式无效: %s (应为 contextKey.field 格式)", path)
	}

	contextKey := parts[0]
	fieldPath := parts[1]

	// 从 datastore 获取 GraphContext
	graphContext, err := datastore.GetGraphContext(ctx, execCtx.GetTenantId(), execCtx.GetGraphKey(), contextKey)
	if err != nil {
		return nil, fmt.Errorf("获取 GraphContext 失败: %w", err)
	}

	payload := graphContext.GetPayload()

	// 解析字段路径（支持 nested.field 格式）
	return getNestedValue(payload, fieldPath)
}

// getNestedValue 获取嵌套字段值
func getNestedValue(data map[string]any, path string) (any, error) {
	parts := strings.Split(path, ".")
	current := any(data)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			val, ok := v[part]
			if !ok {
				return nil, fmt.Errorf("字段 %s 不存在", part)
			}
			current = val
		default:
			return nil, fmt.Errorf("无法访问 %s 的子字段", part)
		}
	}

	return current, nil
}

// replaceVariablesExtended 替换字符串中的 {{atom.xxx}} 和 {{context.xxx.yyy}} 变量
func replaceVariablesExtended(ctx context.Context, template string, payload map[string]any, datastore core.Store, execCtx core.ExecutionContext) string {
	// 替换 {{atom.xxx}}
	reAtom := regexp.MustCompile(`\{\{atom\.(\w+)\}\}`)
	result := reAtom.ReplaceAllStringFunc(template, func(match string) string {
		fieldName := reAtom.FindStringSubmatch(match)[1]
		if value, ok := payload[fieldName]; ok {
			return fmt.Sprintf("%v", value)
		}
		return match
	})

	// 替换 {{context.xxx.yyy}}
	reContext := regexp.MustCompile(`\{\{context\.([a-zA-Z0-9_.]+)\}\}`)
	result = reContext.ReplaceAllStringFunc(result, func(match string) string {
		path := reContext.FindStringSubmatch(match)[1]
		value, err := getValueFromContext(ctx, path, datastore, execCtx)
		if err == nil && value != nil {
			return fmt.Sprintf("%v", value)
		}
		return match
	})

	return result
}
