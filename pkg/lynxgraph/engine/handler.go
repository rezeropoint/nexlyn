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
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/tag"

	"github.com/zeromicro/go-zero/core/stores/monc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 引擎实现
type engine struct {
	config *Config

	infoAtomRegistry   infoatom.InfoAtomRegistry // 信息原子注册表，支持自定义/扩展
	datastore          core.Store
	blockRegistry      block.BlockRegistry // 逻辑块注册表，支持自定义/扩展
	graphRegistry      graph.GraphRegistry // 逻辑图注册表，支持动态加载/卸载
	dispatcherRegistry dispatcher.DispatcherRegistry

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

	// 注册标准区块（只注册依赖的服务已就绪的逻辑块）
	if err := block.RegisterStandardBlocks(blockRegistry, service); err != nil {
		cancel() // 释放 context 资源
		return nil, ErrStandardBlocksFailed
	}

	// 创建 GraphRegistry（注入 PostgreSQL、MongoDB 连接、BlockRegistry 函数、标签查询函数）
	graphRegistry, err := graph.NewGraphRegistry(ctx, cancel, &graph.Config{
		RunMode:    config.RunMode,
		EtcdConfig: config.EtcdConfig,
	}, sqlConn, mongoDB, blockRegistry.CreateBlock, tagManager.GetTagNamesByIDs)
	if err != nil {
		cancel() // 释放 context 资源
		return nil, ErrGraphRegistryFailed
	}

	// 创建 DataStore（使用配置中的 CacheConf）
	datastore, err := core.NewBaseStore(config.KeyPrefix, config.GraphContextTTL, config.InfoAtomTTL, infoAtomRegistry.GetInfoAtomType, config.CacheConf)
	if err != nil {
		cancel() // 释放 context 资源
		return nil, ErrDataStoreFailed
	}
	dispatcherRegistry, err := dispatcher.NewDispatcherRegistry(
		ctx,
		cancel,
		&config.DispatcherConfig,
		func(infoAtom core.InfoAtom) (map[core.GraphKey][]core.Node, error) {
			return graphRegistry.FindGraphsByInfoAtom(infoAtom)
		},
		func(graphKey core.GraphKey) (core.LogicGraph, error) {
			return graphRegistry.GetGraph(graphKey)
		},
		datastore,
		service,
	)
	if err != nil {
		cancel() // 释放 context 资源
		return nil, ErrDispatcherRegistryFailed
	}

	e := &engine{
		config:    config,
		stopChan:  make(chan struct{}),
		isRunning: false,

		infoAtomRegistry:   infoAtomRegistry,
		graphRegistry:      graphRegistry,
		blockRegistry:      blockRegistry,
		dispatcherRegistry: dispatcherRegistry,
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

	// 启动调度器
	if err := e.dispatcherRegistry.Start(); err != nil {
		// 出错时取消上下文
		e.cancel()
		return fmt.Errorf("%w: %v", ErrStartDispatcherFailed, err)
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

	// 关闭调度器
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
