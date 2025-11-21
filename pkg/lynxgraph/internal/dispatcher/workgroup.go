package dispatcher

import "fmt"

// startWorkers 启动工作池
func (r *dispatcherRegistry) startWorkers() {
	r.workerWg.Add(r.config.NumWorkers)
	for i := 0; i < r.config.NumWorkers; i++ {
		go r.worker()
	}
}

// worker 工作者 goroutine，从任务队列获取任务并处理
func (r *dispatcherRegistry) worker() {
	defer r.workerWg.Done()

	for {
		select {
		case <-r.ctx.Done():
			// 根上下文被取消，优雅退出
			return
		case <-r.stopChan:
			// 收到停止信号，退出
			return
		case infoAtom, ok := <-r.taskQueue:
			if !ok {
				// 任务队列已关闭，退出
				return
			}
			// 处理信息原子
			err := r.processInfoAtom(infoAtom)
			if err != nil {
				// 记录错误但不终止工作者
				fmt.Printf("处理信息原子失败: %v\n", err)
			}
		}
	}
}
