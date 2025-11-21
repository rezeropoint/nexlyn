package block

import (
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
)

type BlockRegistry interface {
	// Register 注册一个逻辑块工厂和对应的规格
	Register(blockKey core.BlockKey, factory core.BlockFactory, spec core.BlockSpec) error
	// GetBlockKeyAll 返回所有逻辑块的key
	GetBlockKeyAll() []core.BlockKey
	// GetSpec 返回特定类型和版本的逻辑块规格信息
	GetSpec(blockKey core.BlockKey) (core.BlockSpec, error)
	// CreateBlock 创建一个逻辑块实例，支持指定版本
	CreateBlock(id string, blockKey core.BlockKey, config map[string]any) (core.LogicBlock, error)
}

// NewBlockRegistry 创建一个新的逻辑块注册表
func NewBlockRegistry(config *Config) (BlockRegistry, error) {
	if config == nil {
		return nil, core.ErrConfigNil
	}
	return newBlockRegistry(), nil
}
