package core

// BlockSpec 提供积木定义元信息，支持 UI 渲染、配置校验、文档生成
type BlockSpec interface {
	Name() string                        // 积木名称
	Version() string                     // 版本标识
	Description() string                 // 简要说明
	ConfigSchema() map[string]any        // JSON Schema 或结构体定义（输入）
	OutputSchema() map[string]any        // 输出到图上下文的数据结构（用于边条件智能提示）
	Tags() []string                      // 用于标记功能分类，如 "action", "timer", "state", "skylark"
	RequiredServiceTypes() []ServiceType // 声明依赖的服务类型（如 ServiceTypeSkylarkEngine），每个类型只有唯一实例
}

// BasicBlockSpec 提供 BlockSpec 的基本实现
type BasicBlockSpec struct {
	name                 string
	version              string
	description          string
	tags                 []string
	configSchema         map[string]any
	outputSchema         map[string]any // 输出到图上下文的数据结构
	requiredServiceTypes []ServiceType  // 依赖的服务类型列表
}

// NewBasicBlockSpec 创建一个新的 BasicBlockSpec
func NewBasicBlockSpec(name string, version string, description string, tags []string, configSchema map[string]any, requiredServiceTypes []ServiceType) *BasicBlockSpec {
	if tags == nil {
		tags = []string{}
	}
	if configSchema == nil {
		configSchema = map[string]any{}
	}
	if requiredServiceTypes == nil {
		requiredServiceTypes = []ServiceType{}
	}
	return &BasicBlockSpec{
		name:                 name,
		version:              version,
		description:          description,
		tags:                 tags,
		configSchema:         configSchema,
		requiredServiceTypes: requiredServiceTypes,
	}
}

func (s *BasicBlockSpec) Name() string                        { return s.name }
func (s *BasicBlockSpec) Version() string                     { return s.version }
func (s *BasicBlockSpec) Description() string                 { return s.description }
func (s *BasicBlockSpec) ConfigSchema() map[string]any        { return s.configSchema }
func (s *BasicBlockSpec) OutputSchema() map[string]any        { return s.outputSchema }
func (s *BasicBlockSpec) Tags() []string                      { return s.tags }
func (s *BasicBlockSpec) RequiredServiceTypes() []ServiceType { return s.requiredServiceTypes }

// WithOutputSchema 设置输出结构（链式调用）
func (s *BasicBlockSpec) WithOutputSchema(schema map[string]any) *BasicBlockSpec {
	s.outputSchema = schema
	return s
}
