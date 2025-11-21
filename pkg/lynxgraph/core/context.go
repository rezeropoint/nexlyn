package core

// ========== 执行上下文 (ExecutionContext) ==========

// ExecutionContext 执行上下文接口，包含积木执行所需的所有信息
type ExecutionContext interface {
	GetTenantId() string
	GetInfoAtom() InfoAtom
	GetGraph() LogicGraph
	GetNode() Node
	GetGraphKey() GraphKey // 辅助方法
}

// BaseExecutionContext 提供 ExecutionContext 接口的基本实现
type BaseExecutionContext struct {
	TenantId string     // 租户ID（冗余但方便访问）
	InfoAtom InfoAtom   // 触发执行的信息原子
	Graph    LogicGraph // 当前逻辑图
	Node     Node       // 当前节点
}

// NewExecutionContext 创建执行上下文
func NewExecutionContext(tenantId string, infoAtom InfoAtom, graph LogicGraph, node Node) ExecutionContext {
	return &BaseExecutionContext{
		TenantId: tenantId,
		InfoAtom: infoAtom,
		Graph:    graph,
		Node:     node,
	}
}

func (e *BaseExecutionContext) GetTenantId() string   { return e.TenantId }
func (e *BaseExecutionContext) GetInfoAtom() InfoAtom { return e.InfoAtom }
func (e *BaseExecutionContext) GetGraph() LogicGraph  { return e.Graph }
func (e *BaseExecutionContext) GetNode() Node         { return e.Node }
func (e *BaseExecutionContext) GetGraphKey() GraphKey {
	return GraphKey{ID: e.Graph.GetName()} // TODO: 根据实际情况调整
}

// ========== 图上下文 (GraphContext) ==========

// GraphContext 图上下文接口，存储图执行期间的临时状态
type GraphContext interface {
	GetTenantId() string
	GetGraphKey() GraphKey
	GetContextKey() string
	GetPayload() map[string]any
	SetPayload(payload map[string]any) error
}

// BaseGraphContext 提供 GraphContext 接口的基本实现
type BaseGraphContext struct {
	TenantId   string         `json:"tenantId"`  // 租户ID（必填，用于数据隔离）
	GraphKey   GraphKey       `json:"graph_key"` // 图标识（必填，用于唯一标识图）
	ContextKey string         `json:"contextKey"`
	Payload    map[string]any `json:"payload"` // 事件有效载荷，结构由具体 Type 决定
}

// NewBaseGraphContext 创建基础图上下文
func NewBaseGraphContext(tenantId string, graphKey GraphKey, payload map[string]any) GraphContext {
	return &BaseGraphContext{
		TenantId: tenantId,
		GraphKey: graphKey,
		Payload:  payload,
	}
}

func (c *BaseGraphContext) GetTenantId() string        { return c.TenantId }
func (c *BaseGraphContext) GetGraphKey() GraphKey      { return c.GraphKey }
func (c *BaseGraphContext) GetContextKey() string      { return c.ContextKey }
func (c *BaseGraphContext) GetPayload() map[string]any { return c.Payload }

func (c *BaseGraphContext) SetPayload(payload map[string]any) error {
	c.Payload = payload
	return nil
}
