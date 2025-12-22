package dispatcher

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
)

// DispatcherRegistry 调度器接口，负责将信息原子分发到相应的逻辑图进行处理
type DispatcherRegistry interface {
	// Dispatch 将信息原子分发到相应的逻辑图处理
	Dispatch(infoAtom core.InfoAtom) error
	// DispatchScheduled 分发定时触发的信息原子到指定图
	// 与 Dispatch 不同，此方法只触发 schedule 类型的入口节点，不经过 InfoAtom 类型匹配
	DispatchScheduled(graphKey core.GraphKey, infoAtom core.InfoAtom) error
	// DispatchSubGraph 同步执行子图（用于 ForEach 积木）
	// 与 DispatchScheduled 不同，此方法触发子图的所有入口节点，不限定节点类型
	// ctx: 执行上下文（用于超时控制和取消）
	// graphKey: 子图的 GraphKey
	// infoAtom: 传递给子图的信息原子
	DispatchSubGraph(ctx context.Context, graphKey core.GraphKey, infoAtom core.InfoAtom) error
	// Start 启动调度器
	Start() error
	// Close 关闭调度器，停止所有工作者
	Close() error
}

// NewDispatcherRegistry 创建一个新的调度器
// ctx 参数是引擎的根上下文，cancel 参数是对应的取消函数，用于统一的生命周期管理
// queryFunc 参数是用于查询信息原子被哪些图和节点依赖的函数
func NewDispatcherRegistry(ctx context.Context, cancel context.CancelFunc, config *Config, queryFunc core.InfoAtomQueryFunc, getFunc core.GraphQueryFunc, datastore core.Store, service core.Service) (DispatcherRegistry, error) {
	if config == nil {
		return nil, ErrConfigNil
	}

	if queryFunc == nil {
		return nil, ErrQueryFuncNil
	}

	if getFunc == nil {
		return nil, ErrGetFuncNil
	}

	if datastore == nil {
		return nil, ErrDatastoreNil
	}

	return newDispatcherRegistry(ctx, cancel, config, queryFunc, getFunc, datastore, service), nil
}
