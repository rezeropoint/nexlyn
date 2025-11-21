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
