package workgroup

import "github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

// start 启动工作协程
func (wp *workGroupManager) start() {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

// worker 工作协程
func (wp *workGroupManager) worker() {
	defer wp.wg.Done()

	for {
		select {
		case task := <-wp.taskChan:
			if task != nil {
				task()
			}
		case <-wp.ctx.Done():
			return
		}
	}
}

// Submit 提交任务到工作池
func (wp *workGroupManager) submit(task func()) error {
	select {
	case wp.taskChan <- task:
		// 任务成功提交
		return nil
	case <-wp.ctx.Done():
		// 工作池已关闭
		return wp.ctx.Err()
	default:
		// 任务队列已满，立即返回错误而不阻塞
		return core.ErrTaskQueueFull
	}
}
