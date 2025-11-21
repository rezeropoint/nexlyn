package core

import (
	"errors"
)

// GraphKey 用于唯一标识图的键
type GraphKey struct {
	ID string `json:"id"`
	// TenantId string `json:"tenantId"`
	// Name      string `json:"name"`
	// Version   string `json:"version"`
}

// InfoAtomQueryFunc 定义了查询信息原子依赖图和入口节点的函数类型
type InfoAtomQueryFunc func(infoAtom InfoAtom) (map[GraphKey][]Node, error)
type GraphQueryFunc func(graphKey GraphKey) (LogicGraph, error)

// LogicGraph 表示一个完整的逻辑图，由节点和边组成
type LogicGraph interface {
	GetTenantId() string                     // 返回逻辑图的租户ID
	GetName() string                         // 返回逻辑图的名称
	GetVersion() string                      // 返回逻辑图的版本
	GetDescription() string                  // 返回逻辑图的描述
	GetEnable() bool                         // 返回逻辑图是否启用
	GetNextNodes(currentID string) []Node    // 获取从指定节点出发可到达的所有后继节点
	GetEntryNodes() []Node                   // 获取图的所有入口节点
	GetIncomingNodes(targetID string) []Node // 获取所有指向指定节点的前驱节点
	GetNode(id string) (Node, error)         // 通过ID获取节点
	GetEdge(id string) (Edge, error)         // 通过ID获取边
	AddNode(node Node)                       // 添加节点到图中
	AddEdge(edge Edge)                       // 添加边到图中
	Validate() (bool, error)                 // 验证图结构的有效性
	SetEnable(enable bool)                   // 设置逻辑图是否启用
	GetOutgoingEdges(nodeID string) []Edge   // 获取指定节点的所有出边
	GetAllEdgeIDs() []string                 // 获取所有边的ID
	FindNodeByID(nodeID string) (Node, bool) // the通过ID查找节点（返回节点和是否存在）
}

// BasicLogicGraph 提供 LogicGraph 接口的基本实现
type BasicLogicGraph struct {
	TenantId                   string `json:"tenantId"` // 租户ID（必填，用于数据隔离）
	Name, Version, Description string
	Enable                     bool
	Nodes                      map[string]Node
	Edges                      map[string]Edge
	EntryNodes                 []Node
	Outgoing                   map[string][]Edge // 邻接表，用于高效遍历
	Incoming                   map[string][]Edge // 反向边表，用于聚合判断
	Metadata                   map[string]string // 用于支持可视化/标签控制的灵活扩展
}

// NewBasicLogicGraph 创建一个新的逻辑图
func NewBasicLogicGraph(tenantId, name, version, description string, enable bool) LogicGraph {
	return &BasicLogicGraph{
		TenantId:    tenantId,
		Name:        name,
		Version:     version,
		Description: description,
		Enable:      enable,
		Nodes:       make(map[string]Node),
		Edges:       make(map[string]Edge),
		EntryNodes:  []Node{},
		Outgoing:    make(map[string][]Edge),
		Incoming:    make(map[string][]Edge),
		Metadata:    make(map[string]string),
	}
}

func (g *BasicLogicGraph) GetTenantId() string    { return g.TenantId }
func (g *BasicLogicGraph) GetName() string        { return g.Name }
func (g *BasicLogicGraph) GetVersion() string     { return g.Version }
func (g *BasicLogicGraph) GetDescription() string { return g.Description }
func (g *BasicLogicGraph) GetEnable() bool        { return g.Enable }

// GetNode 实现 LogicGraph 接口
func (g *BasicLogicGraph) GetNode(id string) (Node, error) {
	if node, ok := g.Nodes[id]; ok {
		return node, nil
	}
	return nil, errors.New("节点不存在")
}

// GetEdge 实现 LogicGraph 接口
func (g *BasicLogicGraph) GetEdge(id string) (Edge, error) {
	if edge, ok := g.Edges[id]; ok {
		return edge, nil
	}
	return nil, errors.New("边不存在")
}

// AddNode 实现 LogicGraph 接口
func (g *BasicLogicGraph) AddNode(node Node) {
	nodeID := node.GetID()
	g.Nodes[nodeID] = node
	if node.IsEntryPoint() {
		g.EntryNodes = append(g.EntryNodes, node)
	}
}

// AddEdge 实现 LogicGraph 接口
func (g *BasicLogicGraph) AddEdge(edge Edge) {
	edgeID := edge.GetID()
	g.Edges[edgeID] = edge

	// 更新邻接表
	sourceID := edge.GetSourceID()
	targetID := edge.GetTargetID()

	// 添加到出边列表
	g.Outgoing[sourceID] = append(g.Outgoing[sourceID], edge)

	// 添加到入边列表
	g.Incoming[targetID] = append(g.Incoming[targetID], edge)
}

// GetNextNodes 获取从指定节点出发可到达的所有后继节点
// 使用邻接表高效查询，减少遍历开销
func (g *BasicLogicGraph) GetNextNodes(currentID string) []Node {
	outEdges, exists := g.Outgoing[currentID]
	if !exists {
		return []Node{}
	}

	nextNodes := make([]Node, 0, len(outEdges))
	for _, edge := range outEdges {
		targetID := edge.GetTargetID()
		if node, ok := g.Nodes[targetID]; ok {
			nextNodes = append(nextNodes, node)
		}
	}

	return nextNodes
}

// GetEntryNodes 获取图的所有入口节点
// 直接返回缓存的入口节点列表，无需重复计算
func (g *BasicLogicGraph) GetEntryNodes() []Node {
	return g.EntryNodes
}

// GetIncomingNodes 获取所有指向指定节点的前驱节点
// 使用反向边表高效查询，支持聚合判断
func (g *BasicLogicGraph) GetIncomingNodes(targetID string) []Node {
	inEdges, exists := g.Incoming[targetID]
	if !exists {
		return []Node{}
	}

	incomingNodes := make([]Node, 0, len(inEdges))
	for _, edge := range inEdges {
		sourceID := edge.GetSourceID()
		if node, ok := g.Nodes[sourceID]; ok {
			incomingNodes = append(incomingNodes, node)
		}
	}

	return incomingNodes
}

// Validate 验证图结构的有效性
// 检查图的完整性和正确性，如孤立节点、入口可达性等
func (g *BasicLogicGraph) Validate() (bool, error) {
	// TODO:
	// 需要注意需要调用事件的接口来确保图的事件存在
	return true, nil
}

func (g *BasicLogicGraph) SetEnable(enable bool) {
	g.Enable = enable
}

func (g *BasicLogicGraph) GetOutgoingEdges(nodeID string) []Edge {
	if edges, ok := g.Outgoing[nodeID]; ok {
		return edges
	}
	return nil
}

// GetAllEdgeIDs 获取图中所有边的ID
func (g *BasicLogicGraph) GetAllEdgeIDs() []string {
	edgeIDs := make([]string, 0, len(g.Edges))
	for id := range g.Edges {
		edgeIDs = append(edgeIDs, id)
	}
	return edgeIDs
}

// FindNodeByID 通过ID查找节点
func (g *BasicLogicGraph) FindNodeByID(nodeID string) (Node, bool) {
	node, ok := g.Nodes[nodeID]
	return node, ok
}
