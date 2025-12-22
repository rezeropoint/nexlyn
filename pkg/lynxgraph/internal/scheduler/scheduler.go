package scheduler

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
)

// ScheduleRegistry 定时调度管理器接口
type ScheduleRegistry interface {
	// Start 启动调度器
	Start() error

	// Stop 停止调度器，取消所有定时任务
	Stop() error

	// RegisterSchedule 注册图的定时触发任务
	// 从节点配置中提取定时积木（BlockTypeSchedule）的配置进行注册
	RegisterSchedule(graphKey core.GraphKey, nodes []core.NodeConfig) error

	// UnregisterSchedule 注销图的所有定时任务
	UnregisterSchedule(graphKey core.GraphKey) error

	// GetScheduledGraphs 获取所有已注册定时任务的图
	GetScheduledGraphs() []core.GraphKey

	// IsRunning 检查调度器是否正在运行
	IsRunning() bool
}

// NewScheduleRegistry 创建定时调度管理器
// 参数：
//   - ctx, cancel: 上下文和取消函数
//   - config: 调度器配置
//   - triggerFunc: 定时触发回调函数
//   - datastore: 数据存储（用于分布式锁），可为 nil 表示单实例模式
func NewScheduleRegistry(
	ctx context.Context,
	cancel context.CancelFunc,
	config *Config,
	triggerFunc core.ScheduleTriggerFunc,
	datastore core.Store,
) (ScheduleRegistry, error) {
	if config == nil {
		return nil, ErrConfigNil
	}
	if triggerFunc == nil {
		return nil, ErrTriggerFuncNil
	}
	return newScheduleRegistry(ctx, cancel, config, triggerFunc, datastore)
}
