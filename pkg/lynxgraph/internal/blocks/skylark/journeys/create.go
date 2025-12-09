package journeys

import (
	"context"
	"fmt"
	"regexp"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	skylarkCore "github.com/rezeropoint/go-skylark/v2/core"
	"github.com/rezeropoint/go-skylark/v2/engine"
)

const SkylarkJourneyCreateVersion = "v1"

type Config struct {
	App        string            `json:"app" check:"must"`    // Skylark实例，一般是域名
	FlowId     int64             `json:"flowId" check:"must"` // 流程ID
	UserID     int64             `json:"userId" check:"must"` // 用户ID
	AuthHeader string            `json:"authHeader"`          // 认证头（可选，用于外部调用）
	Data       map[string]string `json:"data"`                // 字段映射，key=目标字段名，value=源字段（支持 {{atom.xxx}}）
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

	b.TypedConfig = &cfg
	return nil
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
		// 使用配置的字段映射
		data = convertWithMapping(config.Data, infoAtom)
	} else {
		// 默认行为：使用全部 Payload
		data = convertPayloadToTypedValue(infoAtom)
	}

	// 5. 调用Skylark引擎创建流程
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
				"type":        "object",
				"title":       "字段映射",
				"description": "自定义字段映射，key=目标字段名，value=源字段（支持 {{atom.xxx}}）。不配置则使用全部 InfoAtom 数据",
				"additionalProperties": map[string]any{
					"type": "string",
				},
			},
		},
		[]core.ServiceType{core.ServiceTypeSkylarkEngine},
	)
}

// convertPayloadToTypedValue 将InfoAtom的Payload转换为Skylark的TypedValue格式
func convertPayloadToTypedValue(infoAtom core.InfoAtom) map[string]skylarkCore.TypedValue {
	payload := infoAtom.GetPayload()
	atomType := infoAtom.GetType()

	// 如果没有类型信息，尝试推断类型
	if atomType == nil {
		return convertWithInference(payload)
	}

	dataFormat := atomType.GetDataFormat()
	result := make(map[string]skylarkCore.TypedValue)

	// 根据DataFormat中的字段配置进行转换
	for _, fieldConfig := range dataFormat.Fields {
		value, exists := payload[fieldConfig.FieldKey]
		if !exists {
			continue
		}

		// 根据FieldType转换为skylark的类型字符串
		skylarkType := convertFieldType(fieldConfig.FieldType)
		result[fieldConfig.FieldKey] = skylarkCore.TypedValue{
			Type:  skylarkType,
			Value: value,
		}
	}

	return result
}

// convertFieldType 将core.FieldType转换为Skylark的类型字符串
func convertFieldType(ft core.FieldType) string {
	switch ft {
	case core.FieldTypeString:
		return "string"
	case core.FieldTypeInt:
		return "int"
	case core.FieldTypeFloat:
		return "float"
	case core.FieldTypeBool:
		return "bool"
	default:
		return "string"
	}
}

// convertWithInference 当InfoAtomType为nil时，通过类型推断转换Payload
func convertWithInference(payload map[string]any) map[string]skylarkCore.TypedValue {
	result := make(map[string]skylarkCore.TypedValue)

	for key, value := range payload {
		var typeName string
		switch value.(type) {
		case string:
			typeName = "string"
		case int, int32, int64:
			typeName = "int"
		case float32, float64:
			typeName = "float"
		case bool:
			typeName = "bool"
		default:
			typeName = "string"
		}

		result[key] = skylarkCore.TypedValue{
			Type:  typeName,
			Value: value,
		}
	}

	return result
}

// convertWithMapping 根据配置的字段映射转换数据
// mapping: key=目标字段名, value=源字段表达式（支持 {{atom.xxx}}）
func convertWithMapping(mapping map[string]string, infoAtom core.InfoAtom) map[string]skylarkCore.TypedValue {
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
	re := regexp.MustCompile(`^\{\{atom\.(\w+)\}\}$`)

	for targetKey, sourceExpr := range mapping {
		// 检查是否是简单变量引用
		if match := re.FindStringSubmatch(sourceExpr); match != nil {
			fieldName := match[1]
			if value, ok := payload[fieldName]; ok {
				// 确定类型
				typeName := inferType(value)
				if ft, ok := fieldTypes[fieldName]; ok {
					typeName = convertFieldType(ft)
				}
				result[targetKey] = skylarkCore.TypedValue{
					Type:  typeName,
					Value: value,
				}
			}
		} else {
			// 字符串模板替换
			value := replaceVariables(sourceExpr, payload)
			result[targetKey] = skylarkCore.TypedValue{
				Type:  "string",
				Value: value,
			}
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

// inferType 推断值的类型
func inferType(value any) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int32, int64:
		return "int"
	case float32, float64:
		return "float"
	case bool:
		return "bool"
	default:
		return "string"
	}
}
