package core

// Edge 表示逻辑图中连接两个节点的边
type Edge interface {
	GetID() string        // GetID 返回边的唯一标识符
	GetSourceID() string  // GetSourceID 返回边的源节点ID
	GetTargetID() string  // GetTargetID 返回边的目标节点ID
	GetCondition() string // GetCondition 返回边的条件表达式（如果有）
}

// BasicEdge 提供 Edge 接口的基本实现
type BasicEdge struct {
	ID        string `json:"id"`
	SourceID  string `json:"source_id"`
	TargetID  string `json:"target_id"`
	Condition string `json:"condition"`
	// Evaluate  func(ctx StateContext) bool `json:"-"` // 不序列化，仅运行期构建
	// Metadata  map[string]string           `json:"metadata,omitempty"` //用于前端提示、日志输出、链路追踪标签；
}

func (e *BasicEdge) GetID() string        { return e.ID }
func (e *BasicEdge) GetSourceID() string  { return e.SourceID }
func (e *BasicEdge) GetTargetID() string  { return e.TargetID }
func (e *BasicEdge) GetCondition() string { return e.Condition }
