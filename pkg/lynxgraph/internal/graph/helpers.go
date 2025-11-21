package graph

import (
	"database/sql"
	"strings"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/google/uuid"
)

// generateCacheKey 生成缓存键
func generateCacheKey(id string) string {
	return "graph:" + id
}

// nullStringToString 将 sql.NullString 转换为 string
// 如果 NULL 则返回空字符串
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// getAllNodes 获取图中的所有节点
// 该函数通过遍历从入口节点可达的所有节点，获取图中的完整节点集合
func getAllNodes(graph core.LogicGraph) []core.Node {
	// 预先分配一个足够大的容量
	entryNodes := graph.GetEntryNodes()
	nodeMap := make(map[string]core.Node, len(entryNodes)*2) // 预估总节点数为入口节点的2倍
	nodes := make([]core.Node, 0, len(entryNodes)*2)

	// 记录已访问的节点，避免重复
	for _, node := range entryNodes {
		nodeID := node.GetID()
		if _, exists := nodeMap[nodeID]; !exists {
			nodeMap[nodeID] = node
			nodes = append(nodes, node)
		}
	}

	// 从入口节点开始，递归获取所有可达节点
	var visitNode func(nodeID string)
	visitNode = func(nodeID string) {
		nextNodes := graph.GetNextNodes(nodeID)
		for _, nextNode := range nextNodes {
			nextID := nextNode.GetID()
			if _, exists := nodeMap[nextID]; !exists {
				nodeMap[nextID] = nextNode
				nodes = append(nodes, nextNode)
				visitNode(nextID)
			}
		}
	}

	// 从每个入口节点开始遍历
	for _, node := range entryNodes {
		visitNode(node.GetID())
	}

	return nodes
}

// matchLabels 检查信息原子的标签是否满足节点的订阅标签要求
// 匹配规则：
// 1. 如果节点没有订阅标签，则匹配任何信息原子
// 2. 如果节点有订阅标签，则信息原子必须包含所有这些标签且值相同
// 3. 额外的信息原子标签不影响匹配结果
func matchLabels(nodeLabels, atomLabels map[string]string) bool {
	// 快速路径：节点没有指定订阅标签，匹配任何信息原子
	if len(nodeLabels) == 0 {
		return true
	}

	// 快速路径：节点指定了标签但信息原子没有标签
	if len(atomLabels) == 0 {
		return false
	}

	// 检查信息原子是否包含节点订阅的所有标签
	for key, value := range nodeLabels {
		atomValue, exists := atomLabels[key]
		if !exists || atomValue != value {
			return false
		}
	}

	return true
}

// isClientTempID 判断是否为前端临时ID
// 前端临时ID格式：client:node-xxx 或 client:edge-xxx
func isClientTempID(id string) bool {
	return strings.HasPrefix(id, "client:")
}

// processNodeAndEdgeIDs 处理节点和边的ID生成与映射
// 功能：
// 1. 为临时ID或空ID的节点生成UUID
// 2. 为临时ID或空ID的边生成UUID
// 3. 更新边的sourceID和targetID引用（临时ID → 真实UUID）
// 4. 保持已有真实UUID不变（编辑场景）
func processNodeAndEdgeIDs(config *core.GraphConfig) error {
	// 节点ID映射表（临时ID → 真实UUID）
	nodeIDMap := make(map[string]string)

	// 1. 处理节点ID
	for i := range config.Nodes {
		oldID := config.Nodes[i].ID

		// 如果是临时ID或空ID，生成新UUID
		if oldID == "" || isClientTempID(oldID) {
			newID := uuid.New().String()
			config.Nodes[i].ID = newID
			if oldID != "" {
				nodeIDMap[oldID] = newID
			}
		}
		// 真实UUID保持不变
	}

	// 2. 处理边ID和节点引用
	for i := range config.Edges {
		oldEdgeID := config.Edges[i].ID

		// 边ID：临时ID或空ID替换为新UUID
		if oldEdgeID == "" || isClientTempID(oldEdgeID) {
			config.Edges[i].ID = uuid.New().String()
		}

		// 更新sourceID引用
		if newSourceID, exists := nodeIDMap[config.Edges[i].SourceID]; exists {
			config.Edges[i].SourceID = newSourceID
		}

		// 更新targetID引用
		if newTargetID, exists := nodeIDMap[config.Edges[i].TargetID]; exists {
			config.Edges[i].TargetID = newTargetID
		}
	}

	return nil
}

// nodeConfigToMongo 将 Core 层节点配置转换为 MongoDB 存储结构
// 设计说明：
//   - Type 和 BlockType 从枚举转换为字符串
//   - map[string]any 字段直接赋值（MongoDB BSON 原生支持）
func nodeConfigToMongo(node core.NodeConfig) graphNodeConfigMongo {
	return graphNodeConfigMongo{
		ID:                        node.ID,
		Type:                      string(node.Type),      // 枚举 → 字符串
		BlockType:                 string(node.BlockType), // 枚举 → 字符串
		BlockVersion:              node.BlockVersion,
		BlockConfig:               node.BlockConfig, // map 直接赋值
		IsEntryPoint:              node.IsEntryPoint,
		SubscribedInfoAtomTypeIDs: node.SubscribedInfoAtomTypeIDs,
		SubscribedSource:          node.SubscribedSource,
		SubscribedLabels:          node.SubscribedLabels, // map 直接赋值
		X:                         node.X,
		Y:                         node.Y,
	}
}

// mongoToNodeConfig 将 MongoDB 存储结构转换为 Core 层节点配置
// 设计说明：
//   - 字符串转换回枚举类型
//   - map[string]any 字段直接赋值
func mongoToNodeConfig(mongoNode graphNodeConfigMongo) core.NodeConfig {
	return core.NodeConfig{
		ID:                        mongoNode.ID,
		Type:                      core.NodeType(mongoNode.Type),            // 字符串 → 枚举
		BlockType:                 core.LogicBlockType(mongoNode.BlockType), // 字符串 → 枚举
		BlockVersion:              mongoNode.BlockVersion,
		BlockConfig:               mongoNode.BlockConfig, // map 直接赋值
		IsEntryPoint:              mongoNode.IsEntryPoint,
		SubscribedInfoAtomTypeIDs: mongoNode.SubscribedInfoAtomTypeIDs,
		SubscribedSource:          mongoNode.SubscribedSource,
		SubscribedLabels:          mongoNode.SubscribedLabels, // map 直接赋值
		X:                         mongoNode.X,
		Y:                         mongoNode.Y,
	}
}

// edgeConfigToMongo 将 Core 层边配置转换为 MongoDB 存储结构
// 设计说明：
//   - 边的结构简单，字段直接映射即可
func edgeConfigToMongo(edge core.EdgeConfig) graphEdgeConfigMongo {
	return graphEdgeConfigMongo{
		ID:        edge.ID,
		SourceID:  edge.SourceID,
		TargetID:  edge.TargetID,
		Condition: edge.Condition,
	}
}

// mongoToEdgeConfig 将 MongoDB 存储结构转换为 Core 层边配置
func mongoToEdgeConfig(mongoEdge graphEdgeConfigMongo) core.EdgeConfig {
	return core.EdgeConfig{
		ID:        mongoEdge.ID,
		SourceID:  mongoEdge.SourceID,
		TargetID:  mongoEdge.TargetID,
		Condition: mongoEdge.Condition,
	}
}
