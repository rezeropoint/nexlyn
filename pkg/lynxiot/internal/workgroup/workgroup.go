package workgroup

type Manager interface {
	Close()                                   // 关闭工作池
	SubmitAsync(tasks []func() error) []error // 提交任务（异步）
}

func NewManager(config Config) Manager {
	return newWorkerManager(config)
}
