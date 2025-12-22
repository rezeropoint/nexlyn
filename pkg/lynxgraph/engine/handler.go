package engine

import (
	"context"
	"fmt"
	"sync"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/block"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/dispatcher"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/graph"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/infoatom"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/scheduler"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/tag"

	"github.com/zeromicro/go-zero/core/stores/monc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 引擎实现
type engine struct {
	config *Config

	infoAtomRegistry   infoatom.InfoAtomRegistry     // 信息原子注册表，支持自定义/扩展
	datastore          core.Store                    // 数据存储
	blockRegistry      block.BlockRegistry           // 逻辑块注册表，支持自定义/扩展
	graphRegistry      graph.GraphRegistry           // 逻辑图注册表，支持动态加载/卸载
	dispatcherRegistry dispatcher.DispatcherRegistry // 信息原子分发器
	scheduleRegistry   scheduler.ScheduleRegistry    // 定时调度器（Engine 模式使用）

	stopChan chan struct{} // 停止信号通道

	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool // 引擎运行状态

	mu sync.RWMutex
}

func newwEngine(
	config *Config,
	sqlConn sqlx.SqlConn,
	mongoDB *monc.Model,
	services []ServiceRegistration,
	options ...EngineOption) (*engine, error) {
	if config == nil {
		return nil, ErrConfigNil
	}

	// 创建新的上下文和取消函数
	ctx, cancel := context.WithCancel(context.Background())

	// 创建服务注册表并注册所有服务
	service := core.NewService()
	for _, svc := range services {
		if err := service.Register(svc.ServiceType, svc.Service); err != nil {
			cancel() // 释放 context 资源
			return nil, fmt.Errorf("注册服务 %s 失败: %w", svc.ServiceType, err)
		}
	}

	// 创建 BlockRegistry（不依赖任何连接）
	blockRegistry, err := block.NewBlockRegistry(&config.BlockConfig)
	if err != nil {
		cancel() // 释放 context 资源
		return nil, ErrBlockRegistryFailed
	}

	// 创建 TagManager（注入 PostgreSQL 连接）
	tagManager, err := tag.NewManager(sqlConn, tag.Config{})
	if err != nil {
		cancel() // 释放 context 资源
		return nil, fmt.Errorf("failed to create tag manager: %w", err)
	}

	// 创建 InfoAtomRegistry（注入 PostgreSQL 连接和标签查询函数）
	infoAtomRegistry, err := infoatom.NewInfoAtomRegistry(&infoatom.Config{
		RunMode:   config.RunMode,
		CacheConf: config.CacheConf,
	}, sqlConn, tagManager.GetTagNamesByIDs)
	if err != nil {
		cancel() // 释放 context 资源
		return nil, ErrInfoAtomRegistryFailed
	}

	// 创建 DataStore（使用配置中的 CacheConf）
	datastore, err := core.NewBaseStore(config.KeyPrefix, config.GraphContextTTL, config.InfoAtomTTL, infoAtomRegistry.GetInfoAtomType, config.CacheConf)
	if err != nil {
		cancel() // 释放 context 资源
		return nil, ErrDataStoreFailed
	}

	// 注册 InfoAtomQueryFunc 到 Service（ForEach 积木需要）
	// 注意：必须显式转换为 core.InfoAtomTypeQueryFunc 类型，否则 method value 存入 interface{} 后类型断言会失败
	var infoAtomQueryFunc core.InfoAtomTypeQueryFunc = infoAtomRegistry.GetInfoAtomType
	if err := service.Register(core.ServiceTypeInfoAtomQuery, infoAtomQueryFunc); err != nil {
		cancel()
		return nil, fmt.Errorf("注册 InfoAtomQuery 服务失败: %w", err)
	}

	// 声明变量用于延迟初始化（解决循环依赖）
	var graphRegistryInstance graph.GraphRegistry
	var dispatcherRegistryInstance dispatcher.DispatcherRegistry
	var scheduleRegistryInstance scheduler.ScheduleRegistry

	// 定时任务回调函数（延迟调用 scheduleRegistry 方法）
	var scheduleRegisterFunc core.ScheduleRegisterFunc
	var scheduleUnregisterFunc core.ScheduleUnregisterFunc

	// 仅 Engine 模式需要定时调度器
	if config.RunMode == core.Engine {
		scheduleRegisterFunc = func(graphKey core.GraphKey, nodes []core.NodeConfig) error {
			if scheduleRegistryInstance != nil {
				return scheduleRegistryInstance.RegisterSchedule(graphKey, nodes)
			}
			return nil
		}
		scheduleUnregisterFunc = func(graphKey core.GraphKey) error {
			if scheduleRegistryInstance != nil {
				return scheduleRegistryInstance.UnregisterSchedule(graphKey)
			}
			return nil
		}
	}

	// 创建 DispatcherRegistry（使用闭包延迟引用 graphRegistryInstance）
	// 注意：必须在 GraphRegistry 创建之前创建 Dispatcher，以便先注册积木
	dispatcherRegistryInstance, err = dispatcher.NewDispatcherRegistry(
		ctx,
		cancel,
		&config.DispatcherConfig,
		func(infoAtom core.InfoAtom) (map[core.GraphKey][]core.Node, error) {
			return graphRegistryInstance.FindGraphsByInfoAtom(infoAtom)
		},
		func(graphKey core.GraphKey) (core.LogicGraph, error) {
			return graphRegistryInstance.GetGraph(graphKey)
		},
		datastore,
		service,
	)
	if err != nil {
		cancel() // 释放 context 资源
		return nil, ErrDispatcherRegistryFailed
	}

	// 注册 Dispatcher 到 Service（ForEach 积木需要）
	if err := service.Register(core.ServiceTypeDispatcher, dispatcherRegistryInstance); err != nil {
		cancel()
		return nil, fmt.Errorf("注册 Dispatcher 服务失败: %w", err)
	}

	// 注册标准区块（只注册依赖的服务已就绪的逻辑块）
	// 注意：必须在 GraphRegistry 创建之前调用，因为 GraphRegistry 初始化时会加载图
	if err := block.RegisterStandardBlocks(blockRegistry, service); err != nil {
		cancel()
		return nil, ErrStandardBlocksFailed
	}

	// 创建 GraphRegistry（注入 PostgreSQL、MongoDB 连接、BlockRegistry 函数、标签查询函数、定时任务回调）
	// 注意：必须在 RegisterStandardBlocks 之后创建，否则加载图时找不到积木
	graphRegistryInstance, err = graph.NewGraphRegistry(ctx, cancel, &graph.Config{
		RunMode:    config.RunMode,
		EtcdConfig: config.EtcdConfig,
	}, sqlConn, mongoDB, blockRegistry.CreateBlock, tagManager.GetTagNamesByIDs, scheduleRegisterFunc, scheduleUnregisterFunc)
	if err != nil {
		cancel() // 释放 context 资源
		return nil, ErrGraphRegistryFailed
	}

	// 仅 Engine 模式创建定时调度器
	if config.RunMode == core.Engine {
		// 定时触发回调函数
		scheduleTriggerFunc := func(graphKey core.GraphKey, scheduleConfig *core.ScheduleConfig) error {
			// 获取图配置以获取 tenantId
			logicGraph, err := graphRegistryInstance.GetGraph(graphKey)
			if err != nil {
				return fmt.Errorf("获取图失败: %w", err)
			}

			// 创建虚拟 InfoAtom
			infoAtom := scheduler.CreateScheduledInfoAtom(
				logicGraph.GetTenantId(),
				graphKey,
				scheduleConfig,
			)

			// 使用 DispatchScheduled 直接分发到指定图
			return dispatcherRegistryInstance.DispatchScheduled(graphKey, infoAtom)
		}

		scheduleRegistryInstance, err = scheduler.NewScheduleRegistry(
			ctx,
			cancel,
			&config.SchedulerConfig,
			scheduleTriggerFunc,
			datastore, // 传入 datastore 用于分布式锁
		)
		if err != nil {
			cancel() // 释放 context 资源
			return nil, ErrSchedulerRegistryFailed
		}
	}

	e := &engine{
		config:    config,
		stopChan:  make(chan struct{}),
		isRunning: false,

		infoAtomRegistry:   infoAtomRegistry,
		graphRegistry:      graphRegistryInstance,
		blockRegistry:      blockRegistry,
		dispatcherRegistry: dispatcherRegistryInstance,
		scheduleRegistry:   scheduleRegistryInstance,
		datastore:          datastore,

		ctx:    ctx,
		cancel: cancel,
	}

	// 应用所有选项
	for _, option := range options {
		if err := option(e); err != nil {
			cancel() // 释放 context 资源
			return nil, err
		}
	}

	return e, nil
}

func (e *engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.isRunning {
		return ErrEngineAlreadyRunning
	}

	// 启动信息原子调度器
	if err := e.dispatcherRegistry.Start(); err != nil {
		// 出错时取消上下文
		e.cancel()
		return fmt.Errorf("%w: %v", ErrStartDispatcherFailed, err)
	}

	// 启动定时调度器（仅 Engine 模式）
	if e.scheduleRegistry != nil {
		// 先注册所有已加载图的定时任务（解决初始化顺序问题）
		if err := e.graphRegistry.RegisterAllSchedules(); err != nil {
			_ = e.dispatcherRegistry.Close()
			e.cancel()
			return fmt.Errorf("注册定时任务失败: %w", err)
		}

		if err := e.scheduleRegistry.Start(); err != nil {
			// 定时调度器启动失败，关闭已启动的 dispatcher
			_ = e.dispatcherRegistry.Close()
			e.cancel()
			return fmt.Errorf("%w: %v", ErrStartSchedulerFailed, err)
		}
	}

	// 更新引擎状态
	e.isRunning = true

	return nil
}

func (e *engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.isRunning {
		return ErrEngineNotRunning
	}

	// 先停止定时调度器（避免新的定时任务触发）
	if e.scheduleRegistry != nil {
		if err := e.scheduleRegistry.Stop(); err != nil {
			return fmt.Errorf("%w: %v", ErrStopSchedulerFailed, err)
		}
	}

	// 关闭信息原子调度器
	if err := e.dispatcherRegistry.Close(); err != nil {
		return fmt.Errorf("%w: %v", ErrCloseDispatcherFailed, err)
	}

	// 关闭图注册表
	if err := e.graphRegistry.Close(); err != nil {
		return fmt.Errorf("%w: %v", ErrCloseGraphRegistryFailed, err)
	}

	// 取消上下文
	if e.cancel != nil {
		e.cancel()
		e.cancel = nil // 清空取消函数引用
	}

	// 更新引擎状态
	e.isRunning = false

	return nil
}

func (e *engine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.isRunning
}

func (e *engine) GetInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey) (core.InfoAtomType, error) {
	return e.infoAtomRegistry.GetInfoAtomType(ctx, key)
}

func (e *engine) Dispatch(infoAtom core.InfoAtom) error {
	if !e.IsRunning() {
		return ErrEngineNotRunning
	}

	// 首先保存信息原子到数据存储
	if err := e.datastore.SaveInfoAtom(e.ctx, infoAtom); err != nil {
		return fmt.Errorf("%w: %v", ErrSaveInfoAtomFailed, err)
	}

	// 使用调度器分发信息原子
	if err := e.dispatcherRegistry.Dispatch(infoAtom); err != nil {
		return fmt.Errorf("%w: %v", ErrDispatchInfoAtomFailed, err)
	}

	return nil
}

func (e *engine) RegisterGraph(graph core.LogicGraph) error {
	return ErrDirectRegisterGraph
}

func (e *engine) UnregisterGraph(id string, version string) error {
	return ErrDirectUnregisterGraph
}

func (e *engine) GetBlockKeyAll() []core.BlockKey {
	return e.blockRegistry.GetBlockKeyAll()
}

// ReceiveInfoAtom 接收信息原子（Engine层统一入口）
//
// 职责：
//  1. 参数验证
//  2. 查询信息原子类型
//  3. 创建InfoAtom对象（调用InfoAtomRegistry）
//  4. 分发到调度器
func (e *engine) ReceiveInfoAtom(ctx context.Context, req core.InfoAtomRequest) (*core.InfoAtomResponse, error) {
	// 1. 参数验证（基础检查）
	if req.TenantID == "" || req.InfoAtomTypeID == "" {
		return &core.InfoAtomResponse{
			Success:   false,
			Message:   "参数验证失败: TenantID 和 InfoAtomTypeID 不能为空",
			ErrorCode: "INVALID_PARAMS",
		}, nil
	}

	// 2. 查询信息原子类型
	infoAtomType, err := e.infoAtomRegistry.GetInfoAtomType(ctx, core.InfoAtomTypeKey{
		ID: req.InfoAtomTypeID,
	})
	if err != nil {
		return &core.InfoAtomResponse{
			Success:   false,
			Message:   fmt.Sprintf("信息原子类型不存在: %v", err),
			ErrorCode: "TYPE_NOT_FOUND",
		}, nil
	}

	// 3. 创建InfoAtom对象（调用InfoAtomRegistry）
	infoAtom, err := e.infoAtomRegistry.CreateInfoAtom(ctx, req, infoAtomType)
	if err != nil {
		return &core.InfoAtomResponse{
			Success:   false,
			Message:   fmt.Sprintf("创建信息原子失败: %v", err),
			ErrorCode: "CREATION_FAILED",
		}, nil
	}

	// 4. 分发到调度器
	if err := e.Dispatch(infoAtom); err != nil {
		return &core.InfoAtomResponse{
			Success:    false,
			Message:    fmt.Sprintf("信息原子分发失败: %v", err),
			ErrorCode:  "DISPATCH_FAILED",
			InfoAtomID: infoAtom.GetID(),
		}, nil
	}

	return &core.InfoAtomResponse{
		Success:    true,
		Message:    "信息原子接收成功",
		InfoAtomID: infoAtom.GetID(),
	}, nil
}
