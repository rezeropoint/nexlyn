package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

// scheduleJob 定时任务信息
type scheduleJob struct {
	EntryID        cron.EntryID         // cron 任务ID
	GraphKey       core.GraphKey        // 图标识
	TriggerIndex   int                  // 触发器索引
	ScheduleConfig *core.ScheduleConfig // 定时配置
}

// scheduleRegistry 定时调度管理器实现
type scheduleRegistry struct {
	config      *Config
	cronRunner  *cron.Cron
	triggerFunc core.ScheduleTriggerFunc
	datastore   core.Store // 用于分布式锁，可为 nil

	// 图到任务的映射：graphID -> []scheduleJob
	jobs map[string][]scheduleJob
	mu   sync.RWMutex

	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	startMu sync.Mutex
}

func newScheduleRegistry(
	ctx context.Context,
	cancel context.CancelFunc,
	config *Config,
	triggerFunc core.ScheduleTriggerFunc,
	datastore core.Store,
) (*scheduleRegistry, error) {
	// 解析默认时区
	defaultTimezone := config.DefaultTimezone
	if defaultTimezone == "" {
		defaultTimezone = "Asia/Shanghai"
	}
	loc, err := time.LoadLocation(defaultTimezone)
	if err != nil {
		loc = time.Local
		logx.Errorf("[ScheduleRegistry] 解析默认时区失败，使用本地时区: %v", err)
	}

	// 创建 cron 实例
	cronOpts := []cron.Option{
		cron.WithLocation(loc),
	}
	if config.EnableSeconds {
		cronOpts = append(cronOpts, cron.WithSeconds())
	}

	return &scheduleRegistry{
		config:      config,
		cronRunner:  cron.New(cronOpts...),
		triggerFunc: triggerFunc,
		datastore:   datastore,
		jobs:        make(map[string][]scheduleJob),
		ctx:         ctx,
		cancel:      cancel,
		started:     false,
	}, nil
}

// Start 启动调度器
func (r *scheduleRegistry) Start() error {
	r.startMu.Lock()
	defer r.startMu.Unlock()

	if r.started {
		return ErrAlreadyStarted
	}

	r.cronRunner.Start()
	r.started = true
	logx.Info("[ScheduleRegistry] 定时调度器已启动")
	return nil
}

// Stop 停止调度器
func (r *scheduleRegistry) Stop() error {
	r.startMu.Lock()
	defer r.startMu.Unlock()

	if !r.started {
		return ErrNotStarted
	}

	// 优雅停止 cron
	ctx := r.cronRunner.Stop()
	<-ctx.Done()

	r.started = false
	logx.Info("[ScheduleRegistry] 定时调度器已停止")
	return nil
}

// RegisterSchedule 注册图的定时触发任务
// 从节点配置中提取定时积木（BlockTypeSchedule）的配置进行注册
func (r *scheduleRegistry) RegisterSchedule(graphKey core.GraphKey, nodes []core.NodeConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 先清理已有任务
	r.unregisterScheduleLocked(graphKey)

	// 从节点中提取定时积木配置
	scheduleConfigs := core.ExtractScheduleConfigsFromNodes(nodes)
	if len(scheduleConfigs) == 0 {
		return nil // 没有定时积木，不需要注册
	}

	var registeredJobs []scheduleJob

	for nodeID, scheduleConfig := range scheduleConfigs {
		// 设置节点 ID，用于后续精确触发
		scheduleConfig.NodeID = nodeID

		// 解析时区
		loc, err := parseTimezone(scheduleConfig.Timezone, r.config.DefaultTimezone)
		if err != nil {
			logx.Errorf("[ScheduleRegistry] 图 %s 节点 %s 时区解析失败: %v", graphKey.ID, nodeID, err)
			continue
		}

		// 创建定时任务（使用带时区的调度器）
		schedule, err := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(scheduleConfig.CronExpr)
		if err != nil {
			logx.Errorf("[ScheduleRegistry] 图 %s 节点 %s cron表达式解析失败: %v", graphKey.ID, nodeID, err)
			continue
		}

		// 创建带时区的调度器
		scheduleWithTZ := &cronScheduleWithLocation{Schedule: schedule, loc: loc}

		entryID := r.cronRunner.Schedule(scheduleWithTZ, r.createJob(graphKey, scheduleConfig))

		registeredJobs = append(registeredJobs, scheduleJob{
			EntryID:        entryID,
			GraphKey:       graphKey,
			TriggerIndex:   0, // 对于定时积木，此字段不再有意义
			ScheduleConfig: scheduleConfig,
		})

		logx.Infof("[ScheduleRegistry] 图 %s 节点 %s 定时任务已注册: cron=%s, timezone=%s",
			graphKey.ID, nodeID, scheduleConfig.CronExpr, scheduleConfig.Timezone)
	}

	if len(registeredJobs) > 0 {
		r.jobs[graphKey.ID] = registeredJobs
	}

	return nil
}

// cronScheduleWithLocation 带时区的 cron 调度器
type cronScheduleWithLocation struct {
	cron.Schedule
	loc *time.Location
}

func (s *cronScheduleWithLocation) Next(t time.Time) time.Time {
	return s.Schedule.Next(t.In(s.loc))
}

// UnregisterSchedule 注销图的所有定时任务
func (r *scheduleRegistry) UnregisterSchedule(graphKey core.GraphKey) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.unregisterScheduleLocked(graphKey)
	return nil
}

// unregisterScheduleLocked 内部注销方法（调用者需持有锁）
func (r *scheduleRegistry) unregisterScheduleLocked(graphKey core.GraphKey) {
	jobs, exists := r.jobs[graphKey.ID]
	if !exists {
		return
	}

	for _, job := range jobs {
		r.cronRunner.Remove(job.EntryID)
		logx.Debugf("[ScheduleRegistry] 图 %s 触发器 %d 已注销", graphKey.ID, job.TriggerIndex)
	}

	delete(r.jobs, graphKey.ID)
}

// GetScheduledGraphs 获取所有已注册定时任务的图
func (r *scheduleRegistry) GetScheduledGraphs() []core.GraphKey {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys := make([]core.GraphKey, 0, len(r.jobs))
	for graphID := range r.jobs {
		keys = append(keys, core.GraphKey{ID: graphID})
	}
	return keys
}

// IsRunning 检查调度器是否正在运行
func (r *scheduleRegistry) IsRunning() bool {
	r.startMu.Lock()
	defer r.startMu.Unlock()
	return r.started
}

// createJob 创建定时任务的执行函数
func (r *scheduleRegistry) createJob(graphKey core.GraphKey, scheduleConfig *core.ScheduleConfig) cron.Job {
	return cron.FuncJob(func() {
		// 分布式锁：确保多副本情况下只有一个实例执行
		lockKey := fmt.Sprintf("schedule:%s:%s", graphKey.ID, scheduleConfig.CronExpr)
		lockTTL := 5 * time.Minute // 锁的过期时间，防止执行异常时锁无法释放

		if r.datastore != nil {
			acquired, err := r.datastore.Lock(r.ctx, lockKey, lockTTL)
			if err != nil {
				logx.Errorf("[ScheduleRegistry] 获取分布式锁失败: 图=%s, 错误=%v", graphKey.ID, err)
				return
			}
			if !acquired {
				logx.Debugf("[ScheduleRegistry] 未获取到分布式锁，跳过执行: 图=%s", graphKey.ID)
				return
			}
			// 任务执行完毕后释放锁
			defer func() {
				if err := r.datastore.Unlock(r.ctx, lockKey); err != nil {
					logx.Errorf("[ScheduleRegistry] 释放分布式锁失败: 图=%s, 错误=%v", graphKey.ID, err)
				}
			}()
		}

		logx.Infof("[ScheduleRegistry] 定时触发: 图=%s, cron=%s", graphKey.ID, scheduleConfig.CronExpr)

		// 调用触发函数
		if err := r.triggerFunc(graphKey, scheduleConfig); err != nil {
			logx.Errorf("[ScheduleRegistry] 定时触发执行失败: 图=%s, 错误=%v", graphKey.ID, err)
		}
	})
}
