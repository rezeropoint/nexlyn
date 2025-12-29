package dispatcher

import (
	"context"
	"fmt"
	"sync"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
)

// dispatcherRegistry 实现 Dispatcher 接口
type dispatcherRegistry struct {
	config    *Config
	queryFunc core.InfoAtomQueryFunc
	getFunc   core.GraphQueryFunc
	datastore core.Store
	service   core.Service
	mu        sync.RWMutex

	// 上下文管理
	ctx    context.Context
	cancel context.CancelFunc

	// 工作池相关
	taskQueue chan core.InfoAtom
	stopChan  chan struct{}
	workerWg  sync.WaitGroup
	started   bool
	startMu   sync.Mutex
}

// newDispatcherRegistry 创建一个新的调度器实例
func newDispatcherRegistry(ctx context.Context, cancel context.CancelFunc, config *Config, queryFunc core.InfoAtomQueryFunc, getFunc core.GraphQueryFunc, datastore core.Store, service core.Service) *dispatcherRegistry {
	return &dispatcherRegistry{
		config:    config,
		queryFunc: queryFunc,
		getFunc:   getFunc,
		datastore: datastore,
		service:   service,
		ctx:       ctx,
		cancel:    cancel,
		taskQueue: make(chan core.InfoAtom, config.WorkerQueueSize),
		stopChan:  make(chan struct{}),
	}
}

// Start 启动调度器
func (r *dispatcherRegistry) Start() error {
	r.startMu.Lock()
	defer r.startMu.Unlock()

	if r.started {
		return fmt.Errorf("调度器已启动")
	}

	// 启动工作池
	r.startWorkers()
	r.started = true

	return nil
}

// Dispatch 将信息原子分发到工作池
func (r *dispatcherRegistry) Dispatch(infoAtom core.InfoAtom) error {
	r.startMu.Lock()
	defer r.startMu.Unlock()

	// 检查调度器是否已启动
	if !r.started {
		return fmt.Errorf("调度器未启动")
	}

	// 在持有锁的情况下检查是否已关闭
	select {
	case <-r.stopChan:
		return fmt.Errorf("调度器已关闭")
	default:
		// 继续执行
	}

	// 将任务发送到工作队列（非阻塞）
	select {
	case r.taskQueue <- infoAtom:
		return nil
	default:
		return fmt.Errorf("任务队列已满")
	}
}

// Close 关闭调度器，停止所有工作者
func (r *dispatcherRegistry) Close() error {
	r.startMu.Lock()
	defer r.startMu.Unlock()

	if !r.started {
		return fmt.Errorf("调度器未启动")
	}

	// 确保只关闭一次
	select {
	case <-r.stopChan:
		return fmt.Errorf("调度器已关闭")
	default:
		// 取消上下文，发送级联取消信号
		r.cancel()
		close(r.stopChan)
	}

	// 等待所有工作者退出
	r.workerWg.Wait()

	// 关闭任务队列
	close(r.taskQueue)

	r.started = false
	return nil
}

// DispatchScheduled 分发定时触发的信息原子到指定图
// 与 Dispatch 不同，此方法只触发 schedule 类型的入口节点，不经过 InfoAtom 类型匹配
func (r *dispatcherRegistry) DispatchScheduled(graphKey core.GraphKey, infoAtom core.InfoAtom) error {
	r.startMu.Lock()
	started := r.started
	r.startMu.Unlock()

	if !started {
		return fmt.Errorf("调度器未启动")
	}

	// 检查是否已关闭
	select {
	case <-r.stopChan:
		return fmt.Errorf("调度器已关闭")
	default:
		// 继续执行
	}

	// 直接获取指定图
	graph, err := r.getFunc(graphKey)
	if err != nil {
		return fmt.Errorf("获取图失败: %w", err)
	}

	// 获取入口节点
	entryNodes := graph.GetEntryNodes()

	// 从 InfoAtom labels 中获取触发的 schedule 节点 ID
	targetNodeID := ""
	if labels := infoAtom.GetLabels(); labels != nil {
		targetNodeID = labels["schedule_node_id"]
	}

	// 只选择 schedule 类型的入口节点，避免触发其他入口（如信息原子订阅入口）
	// 如果指定了 targetNodeID，则只触发该节点
	var scheduleEntryNodes []core.Node
	for _, node := range entryNodes {
		block := node.GetBlock()
		if block == nil {
			continue
		}
		if block.GetType() == core.BlockTypeSchedule {
			// 如果指定了目标节点 ID，则只匹配该节点
			if targetNodeID != "" && block.GetID() != targetNodeID {
				continue
			}
			scheduleEntryNodes = append(scheduleEntryNodes, node)
		}
	}

	if len(scheduleEntryNodes) == 0 {
		if targetNodeID != "" {
			return fmt.Errorf("图没有匹配的 schedule 入口节点: %s", targetNodeID)
		}
		return fmt.Errorf("图没有 schedule 类型的入口节点 (共 %d 个入口节点)", len(entryNodes))
	}

	// 构造 graphNodesMap 并使用复用的处理逻辑
	graphNodesMap := map[core.GraphKey][]core.Node{
		graphKey: scheduleEntryNodes,
	}

	return r.processInfoAtomWithGraphs(infoAtom, graphNodesMap)
}

// DispatchSubGraph 同步执行子图（用于 ForEach 积木）
// 与 DispatchScheduled 不同，此方法触发子图的所有入口节点，不限定节点类型
func (r *dispatcherRegistry) DispatchSubGraph(ctx context.Context, graphKey core.GraphKey, infoAtom core.InfoAtom) error {
	r.startMu.Lock()
	started := r.started
	r.startMu.Unlock()

	if !started {
		return fmt.Errorf("调度器未启动")
	}

	// 检查是否已关闭
	select {
	case <-r.stopChan:
		return fmt.Errorf("调度器已关闭")
	default:
		// 继续执行
	}

	// 获取子图
	graph, err := r.getFunc(graphKey)
	if err != nil {
		return fmt.Errorf("获取子图失败: %w", err)
	}

	// 获取所有入口节点（不限定类型）
	entryNodes := graph.GetEntryNodes()
	if len(entryNodes) == 0 {
		return fmt.Errorf("子图没有入口节点")
	}

	// 构造 graphNodesMap 并使用复用的处理逻辑
	graphNodesMap := map[core.GraphKey][]core.Node{
		graphKey: entryNodes,
	}

	return r.processInfoAtomWithGraphsCtx(ctx, infoAtom, graphNodesMap)
}
