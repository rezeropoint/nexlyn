package block

import (
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
)

// getFactory 返回特定类型和版本的逻辑块工厂
func (r *blockRegistry) getFactory(blockKey core.BlockKey) (core.BlockFactory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 先尝试查找指定版本
	if factory, exists := r.factories[blockKey]; exists {
		return factory, nil
	}

	return nil, core.ErrFactoryNotFound // 未找到逻辑块工厂
}
