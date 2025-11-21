package journeys

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	skylarkCore "github.com/rezeropoint/go-skylark/v2/core"
	"github.com/rezeropoint/go-skylark/v2/engine"
)

const SkylarkJourneyCreateVersion = "v1"

type Config struct {
	App        string `json:"app" check:"must"`        // Skylark实例，一般是域名
	FlowId     int64  `json:"flowId" check:"must"`     // 流程ID
	UserID     int64  `json:"userID" check:"must"`     // 用户ID
	AuthHeader string `json:"authHeader" check:"must"` // 认证头
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
	data := convertPayloadToTypedValue(infoAtom)

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
		"创建一个新的 Skylark 流程记录，将信息原子转换为 Skylark 流程记录（自动从ExecutionContext获取InfoAtom）",
		[]string{"input", "output"},
		map[string]any{ // 配置模式
			"app": map[string]any{
				"type":        "string",
				"description": "Skylark实例，一般是域名",
				"check":       "must",
			},
			"flowId": map[string]any{
				"type":        "integer",
				"description": "流程ID",
				"check":       "must",
			},
			"userID": map[string]any{
				"type":        "integer",
				"description": "用户ID",
				"check":       "must",
			},
			"authHeader": map[string]any{
				"type":        "string",
				"description": "认证头",
				"check":       "must",
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
