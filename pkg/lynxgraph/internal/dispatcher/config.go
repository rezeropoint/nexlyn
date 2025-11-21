package dispatcher

// Config 调度器配置
type Config struct {
	// 是否并行执行不同图的处理
	// 如果为true，则不同图的处理会并行执行
	// 如果为false，则按顺序执行
	ParallelExecution bool

	// 是否在节点执行失败时继续处理同图中的其他节点
	// 如果为true，则某个节点失败不会影响其他节点的处理
	// 如果为false，则某个节点失败会导致整个图的处理终止
	ContinueOnNodeFailure bool

	// 执行超时时间（毫秒）
	// 如果为0，则不设置超时
	TimeoutMs int

	// 最大并行处理的图数量
	// 如果ParallelExecution为true，此参数控制最大并行数
	// 如果为0，则不限制
	MaxParallelGraphs int

	// 工作池配置
	// 工作池中的工作者数量
	NumWorkers int
	// 工作队列的大小
	WorkerQueueSize int
}
