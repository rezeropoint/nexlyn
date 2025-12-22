package core

import "fmt"

// GraphConfig 逻辑图配置的领域模型
//
// 设计原则：
//   - 这是纯粹的领域模型，不包含任何框架依赖（无 bson、gorm、db 等标签）
//   - 存储层实现使用 internal/graph/model.go 中的数据库专用结构体
//   - 不包含任何展示层字段（如标签名称），展示数据由 manager.GraphConfigDTO 承载
//
// 字段说明：
//   - TagIDs: 标签ID列表（持久化字段，存储在 PostgreSQL）
//
// ⚠️ 关于标签名称（Tags）：
//   - Tags 字段已从 Core 层移除，因为它是展示层概念，不属于领域模型
//   - API 响应需要标签名称时，应使用 manager.GraphConfigDTO（详见 Phase 2 重构计划）
//   - Manager 层会调用 GetTagNamesByIDsFunc 将 TagIDs 转换为标签名称并填充到 DTO
//   - 详见 DEVELOPMENT.md 的"Phase 2: 引入 DTO 模式"章节
type GraphConfig struct {
	TenantId    string          `json:"tenantId"`
	ID          string          `json:"id"`                  // 图的唯一标识符
	Name        string          `json:"name"`                // 图的名称
	Version     string          `json:"version"`             // 图的版本
	Description string          `json:"description"`         // 图的描述
	TagIDs      []string        `json:"tagIds"`              // 标签ID列表（持久化字段）
	Enable      bool            `json:"enable"`              // 图是否启用
	Icon        string          `json:"icon,omitempty"`      // 图标名称（用于前端显示）
	IconColor   string          `json:"iconColor,omitempty"` // 图标颜色（用于前端显示）
	Nodes       []NodeConfig    `json:"nodes"`               // 图中的节点
	Edges       []EdgeConfig    `json:"edges"`               // 图中的边
	OrgID       string          `json:"orgId"`               // 所属组织ID（用于权限控制）
	CreatedBy   string          `json:"createdBy"`           // 创建人用户ID
	UpdatedBy   string          `json:"updatedBy"`           // 更新人用户ID
	CreatedAt   int64           `json:"createdAt"`           // 创建时间（Unix时间戳）
	UpdatedAt   int64           `json:"updatedAt"`           // 更新时间（Unix时间戳）
}

// NodeConfig 节点配置
type NodeConfig struct {
	ID                        string            `json:"id"`                          // 节点的唯一标识符
	Type                      NodeType          `json:"type"`                        // 节点类型
	BlockType                 LogicBlockType    `json:"blockType"`                   // 节点关联的逻辑块类型
	BlockVersion              string            `json:"blockVersion"`                // 节点关联的逻辑块版本
	BlockConfig               map[string]any    `json:"blockConfig"`                 // 节点关联的逻辑块配置
	IsEntryPoint              bool              `json:"isEntryPoint"`                // 是否为入口节点
	SubscribedInfoAtomTypeIDs []string          `json:"subscribed_infoatom_typeids"` // 节点订阅的信息原子类型ID列表
	SubscribedSource          string            `json:"subscribed_source"`           // 节点订阅的信息原子来源
	SubscribedLabels          map[string]string `json:"subscribed_labels"`           // 节点订阅的标签列表
	X                         float64           `json:"x,omitempty"`                 // 画布X坐标（用于前端可视化编辑器）
	Y                         float64           `json:"y,omitempty"`                 // 画布Y坐标（用于前端可视化编辑器）
}

// EdgeConfig 边配置
type EdgeConfig struct {
	ID        string `json:"id"`        // 边的唯一标识符
	SourceID  string `json:"sourceId"`  // 边的源节点ID
	TargetID  string `json:"targetId"`  // 边的目标节点ID
	Condition string `json:"condition"` // 边的条件表达式
}

func ScanConfig(config *GraphConfig) error {
	// 基本参数检查
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	if config.Name == "" {
		return fmt.Errorf("图名称不能为空")
	}

	if config.Version == "" {
		return fmt.Errorf("图版本不能为空")
	}

	// 允许创建空图（节点可以在后续通过可视化编辑器添加）
	// 只有当存在节点时才进行节点相关的验证
	if len(config.Nodes) == 0 {
		return nil // 空图有效，直接返回
	}

	// 检查是否至少有一个入口节点
	hasEntryPoint := false
	nodeMap := make(map[string]struct{})

	for _, node := range config.Nodes {
		// 检查节点ID是否为空
		if node.ID == "" {
			return fmt.Errorf("节点ID不能为空")
		}

		// 检查节点ID是否重复
		if _, exists := nodeMap[node.ID]; exists {
			return fmt.Errorf("节点ID重复: %s", node.ID)
		}
		nodeMap[node.ID] = struct{}{}

		// 检查是否有入口节点
		if node.IsEntryPoint {
			hasEntryPoint = true
		}

		// 检查节点类型
		if node.Type == "" {
			return fmt.Errorf("节点 %s 的类型不能为空", node.ID)
		}

		// 检查逻辑块类型
		if node.BlockType == "" {
			return fmt.Errorf("节点 %s 的逻辑块类型不能为空", node.ID)
		}

		// 检查逻辑块版本
		if node.BlockVersion == "" {
			return fmt.Errorf("节点 %s 的逻辑块版本不能为空", node.ID)
		}
	}

	// 确保图有入口节点
	if !hasEntryPoint {
		return fmt.Errorf("图必须至少有一个入口节点")
	}

	// 检查边的有效性
	for _, edge := range config.Edges {
		// 检查边ID是否为空
		if edge.ID == "" {
			return fmt.Errorf("边ID不能为空")
		}

		// 检查源节点是否存在
		if _, exists := nodeMap[edge.SourceID]; !exists {
			return fmt.Errorf("边 %s 的源节点 %s 不存在", edge.ID, edge.SourceID)
		}

		// 检查目标节点是否存在
		if _, exists := nodeMap[edge.TargetID]; !exists {
			return fmt.Errorf("边 %s 的目标节点 %s 不存在", edge.ID, edge.TargetID)
		}
	}

	// 验证定时积木的约束
	// 1. 定时积木必须是入口节点
	// 2. 定时积木不能有上游边
	scheduleNodeIDs := make(map[string]struct{})
	for _, node := range config.Nodes {
		if node.BlockType == BlockTypeSchedule {
			if !node.IsEntryPoint {
				return fmt.Errorf("定时积木 %s 必须设置为入口节点", node.ID)
			}
			scheduleNodeIDs[node.ID] = struct{}{}
		}
	}

	// 检查定时积木不能作为边的目标节点
	for _, edge := range config.Edges {
		if _, isSchedule := scheduleNodeIDs[edge.TargetID]; isSchedule {
			return fmt.Errorf("定时积木 %s 不能有上游边", edge.TargetID)
		}
	}

	return nil
}
