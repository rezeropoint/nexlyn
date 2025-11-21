package workgroup

import (
	"context"
	"sync"
)

// WorkerPool 工作池结构体
type workGroupManager struct {
	workerCount int            // 工作协程数量
	taskChan    chan func()    // 任务通道
	wg          sync.WaitGroup // 等待组
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewWorkerPool 创建新的工作池
func newWorkerManager(config Config) *workGroupManager {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &workGroupManager{
		workerCount: config.WorkerCount,
		taskChan:    make(chan func(), config.WorkerCount*2), // 缓冲通道，容量为工作协程数的2倍
		ctx:         ctx,
		cancel:      cancel,
	}

	// 启动工作协程
	pool.start()

	return pool
}

// Close 关闭工作池
func (wp *workGroupManager) Close() {
	wp.cancel()
	close(wp.taskChan)
	wp.wg.Wait()
}

// SubmitAsync 异步提交任务，只返回提交错误，不等待任务执行
func (wp *workGroupManager) SubmitAsync(tasks []func() error) []error {
	if len(tasks) == 0 {
		return nil
	}

	results := make([]error, len(tasks))

	// 提交所有任务，只返回提交时的错误
	for i, task := range tasks {
		taskFunc := task
		err := wp.submit(func() {
			// 执行任务，但不返回执行结果
			taskFunc()
		})
		results[i] = err
	}

	return results
}
