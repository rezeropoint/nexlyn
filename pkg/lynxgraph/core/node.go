package core

// NodeType 节点的类型
type NodeType string

// Node 表示逻辑图中的一个节点
type Node interface {
	GetID() string                          // GetID 返回节点的唯一标识符
	GetType() NodeType                      // GetType 返回节点的类型
	GetBlock() LogicBlock                   // GetBlock 返回节点关联的逻辑块
	IsEntryPoint() bool                     // IsEntryPoint 判断节点是否为入口节点
	GetSubscribedInfoAtomTypeIDs() []string // GetSubscribedInfoAtomTypeIDs 获取节点的属性
	GetSubscribedSource() string            // GetSubscribedSource 获取节点订阅的信息原子来源
	GetSubscribedLabels() map[string]string // GetSubscribedLabels 获取节点的标签
}

// BasicNode 提供 Node 接口的基本实现
type BasicNode struct {
	ID                        string            `json:"id"`
	Type                      NodeType          `json:"type"`
	Block                     LogicBlock        `json:"block"`
	EntryPoint                bool              `json:"entry_point"`
	SubscribedInfoAtomTypeIDs []string          `json:"subscribed_info_atom_typeids"`
	SubscribedSource          string            `json:"subscribedSource"`
	SubscribedLabels          map[string]string `json:"subscribed_labels"`
}

func (n *BasicNode) GetID() string                          { return n.ID }
func (n *BasicNode) GetType() NodeType                      { return n.Type }
func (n *BasicNode) GetBlock() LogicBlock                   { return n.Block }
func (n *BasicNode) IsEntryPoint() bool                     { return n.EntryPoint }
func (n *BasicNode) GetSubscribedInfoAtomTypeIDs() []string { return n.SubscribedInfoAtomTypeIDs }
func (n *BasicNode) GetSubscribedSource() string            { return n.SubscribedSource }
func (n *BasicNode) GetSubscribedLabels() map[string]string { return n.SubscribedLabels }
