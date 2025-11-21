package block

import (
	"sync"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
)

type blockRegistry struct {
	mu        sync.RWMutex
	factories map[core.BlockKey]core.BlockFactory // 存储每个逻辑块的工厂函数
	specs     map[core.BlockKey]core.BlockSpec    // 存储每个逻辑块的规格信息
}

// newBlockRegistry 创建一个新的逻辑块注册表
func newBlockRegistry() *blockRegistry {
	return &blockRegistry{
		factories: make(map[core.BlockKey]core.BlockFactory),
		specs:     make(map[core.BlockKey]core.BlockSpec),
	}
}

// Register 注册一个逻辑块工厂和对应的规格
func (r *blockRegistry) Register(blockKey core.BlockKey, factory core.BlockFactory, spec core.BlockSpec) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 检测是否已经注册过
	if _, exists := r.factories[blockKey]; exists {
		return core.ErrBlockAlreadyRegistered // 逻辑块类型已注册
	}

	// 注册工厂函数
	r.factories[blockKey] = factory

	// 如果提供了规格信息，则注册规格信息
	if spec != nil {
		r.specs[blockKey] = spec
	}

	return nil
}

// GetBlockKeyAll 返回所有逻辑块的key
func (r *blockRegistry) GetBlockKeyAll() []core.BlockKey {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys := make([]core.BlockKey, 0, len(r.factories))
	for key := range r.factories {
		keys = append(keys, key)
	}

	return keys
}

// GetSpec 返回特定类型和版本的逻辑块规格信息
func (r *blockRegistry) GetSpec(blockKey core.BlockKey) (core.BlockSpec, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 先尝试查找指定版本
	if spec, exists := r.specs[blockKey]; exists {
		return spec, nil
	}

	// 如果找不到指定版本，回退到默认版本
	defaultKey := core.BlockKey{BlockType: blockKey.BlockType, Version: ""}
	if spec, exists := r.specs[defaultKey]; exists {
		return spec, nil
	}

	return nil, core.ErrSpecNotFound // 未找到逻辑块规格
}

// CreateBlock 创建一个逻辑块实例，支持指定版本
func (r *blockRegistry) CreateBlock(id string, blockKey core.BlockKey, config map[string]any) (core.LogicBlock, error) {
	factory, err := r.getFactory(blockKey)
	if err != nil {
		return nil, err
	}

	return factory(id, config)
}
