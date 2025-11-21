package manager

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/block"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/graph"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/infoatom"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/tag"

	"github.com/zeromicro/go-zero/core/stores/monc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type manager struct {
	infoAtomRegistry infoatom.InfoAtomRegistry
	graphRegistry    graph.GraphRegistry
	blockRegistry    block.BlockRegistry
	tagManager       tag.Manager
}

// newManager 创建一个新的Manager实例
func newManager(
	config *Config,
	sqlConn sqlx.SqlConn,
	mongoDB *monc.Model,
	options ...ManagerOption) (*manager, error) {
	if config == nil {
		return nil, ErrConfigNil
	}

	// 创建新的上下文和取消函数
	ctx, cancel := context.WithCancel(context.Background())

	// 创建 BlockRegistry（不依赖任何连接）
	blockRegistry, err := block.NewBlockRegistry(&block.Config{})
	if err != nil {
		cancel() // 释放 context 资源
		return nil, fmt.Errorf("%w: %v", ErrBlockRegistryFailed, err)
	}

	// 注册标准逻辑块（Manager 模式不需要检查服务依赖，传 nil）
	if err := block.RegisterStandardBlocks(blockRegistry, nil); err != nil {
		cancel() // 释放 context 资源
		return nil, fmt.Errorf("%w: %v", ErrStandardBlocksFailed, err)
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
		return nil, fmt.Errorf("%w: %v", ErrInfoAtomRegistryFailed, err)
	}

	// 创建 GraphRegistry（注入 PostgreSQL、MongoDB 连接、BlockRegistry 函数、标签查询函数）
	graphRegistry, err := graph.NewGraphRegistry(ctx, cancel, &graph.Config{
		RunMode:    config.RunMode,
		EtcdConfig: config.EtcdConfig,
	}, sqlConn, mongoDB, blockRegistry.CreateBlock, tagManager.GetTagNamesByIDs)
	if err != nil {
		cancel() // 释放 context 资源
		return nil, fmt.Errorf("%w: %v", ErrGraphRegistryFailed, err)
	}

	e := &manager{
		infoAtomRegistry: infoAtomRegistry,
		graphRegistry:    graphRegistry,
		blockRegistry:    blockRegistry,
		tagManager:       tagManager,
	}

	// 应用所有选项
	for _, option := range options {
		if err := option(e); err != nil {
			return nil, err
		}
	}

	// 初始化数据库表
	err = e.InitInfoAtomTable(ctx)
	if err != nil {
		return nil, err
	}

	err = e.InitGraphTable(ctx)
	if err != nil {
		return nil, err
	}

	err = e.InitTagTable(ctx)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (m *manager) InitInfoAtomTable(ctx context.Context) error {
	return m.infoAtomRegistry.InitInfoAtomTable(ctx)
}

// CheckInfoAtomType 检查信息原子类型是否存在
func (m *manager) CheckInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey) bool {
	return m.infoAtomRegistry.CheckInfoAtomType(ctx, key)
}

// GetInfoAtomType 获取信息原子类型
func (m *manager) GetInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey) (core.InfoAtomType, error) {
	return m.infoAtomRegistry.GetInfoAtomType(ctx, key)
}

// GetInfoAtomTypeList 获取信息原子类型列表
func (m *manager) GetInfoAtomTypeList(ctx context.Context, ListRequest core.InfoAtomList) ([]core.InfoAtomType, int64, int64, int64, error) {
	return m.infoAtomRegistry.GetInfoAtomTypeList(ctx, ListRequest)
}

// CreateInfoAtomType 创建新的信息原子类型
func (m *manager) CreateInfoAtomType(ctx context.Context, infoAtomType core.InfoAtomType) error {
	return m.infoAtomRegistry.CreateInfoAtomType(ctx, infoAtomType)
}

// UpdateInfoAtomType 更新信息原子类型
func (m *manager) UpdateInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey, infoAtomType core.InfoAtomType) error {
	return m.infoAtomRegistry.UpdateInfoAtomType(ctx, key, infoAtomType)
}

// DeleteInfoAtomType 删除信息原子类型
func (m *manager) DeleteInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey) error {
	return m.infoAtomRegistry.DeleteInfoAtomType(ctx, key)
}

func (m *manager) InitGraphTable(ctx context.Context) error {
	return m.graphRegistry.InitGraphTable(ctx)
}

func (m *manager) InitTagTable(ctx context.Context) error {
	return m.tagManager.InitTagTable(ctx)
}

// GetGraphConfig 获取逻辑图配置
func (m *manager) GetGraphConfig(ctx context.Context, key core.GraphKey) (core.GraphConfig, error) {
	return m.graphRegistry.GetGraphConfig(ctx, key)
}

// GetGraphConfigList 获取逻辑图配置列表
func (m *manager) GetGraphConfigList(ctx context.Context, ListParams core.GraphList) ([]core.GraphConfig, int64, int64, int64, error) {
	return m.graphRegistry.GetGraphConfigList(ctx, ListParams)
}

// CreateGraph 创建逻辑图
func (m *manager) CreateGraphConfig(ctx context.Context, config core.GraphConfig) error {
	return m.graphRegistry.CreateGraphConfig(ctx, config)
}

// UpdateGraph 更新逻辑图
func (m *manager) UpdateGraphConfig(ctx context.Context, config core.GraphConfig) error {
	return m.graphRegistry.UpdateGraphConfig(ctx, config)
}

// DeleteGraph 删除逻辑图
func (m *manager) DeleteGraphConfig(ctx context.Context, key core.GraphKey) error {
	return m.graphRegistry.DeleteGraphConfig(ctx, key)
}

func (m *manager) GetBlockKeyAll() []core.BlockKey {
	return m.blockRegistry.GetBlockKeyAll()
}

func (m *manager) GetBlockSpec(blockKey core.BlockKey) (core.BlockSpec, error) {
	return m.blockRegistry.GetSpec(blockKey)
}

// Tag Manager 方法代理
// CreateTag 创建标签
func (m *manager) CreateTag(ctx context.Context, metadata core.LynxTagMetadata) (string, error) {
	return m.tagManager.CreateTag(ctx, metadata)
}

// GetTag 获取标签详情
func (m *manager) GetTag(ctx context.Context, tagID string, tenantID string) (*core.LynxTag, error) {
	return m.tagManager.GetTag(ctx, tagID, tenantID)
}

// ListTags 查询标签列表
func (m *manager) ListTags(ctx context.Context, query core.LynxTagQuery) ([]*core.LynxTagSummary, int64, error) {
	return m.tagManager.ListTags(ctx, query)
}

// UpdateTag 更新标签
func (m *manager) UpdateTag(ctx context.Context, tagID string, tenantID string, update core.LynxTagUpdate) error {
	return m.tagManager.UpdateTag(ctx, tagID, tenantID, update)
}

// DeleteTag 删除标签
func (m *manager) DeleteTag(ctx context.Context, tagID string, tenantID string) error {
	return m.tagManager.DeleteTag(ctx, tagID, tenantID)
}

// GetTagsByNames 批量查询标签
func (m *manager) GetTagsByNames(ctx context.Context, tenantID string, scope string, names []string) ([]*core.LynxTagSummary, error) {
	return m.tagManager.GetTagsByNames(ctx, tenantID, scope, names)
}

// GetTagsByIDs Phase 2.5: 根据标签ID列表批量查询标签摘要（包含 id、name、description、scope）
func (m *manager) GetTagsByIDs(ctx context.Context, tagIDs []string, tenantID string) ([]*core.LynxTagSummary, error) {
	return m.tagManager.GetTagsByIDs(ctx, tagIDs, tenantID)
}

// GetTagNamesByIDs 根据标签ID列表批量查询标签名称（供 REST 层使用）
func (m *manager) GetTagNamesByIDs(ctx context.Context, tagIDs []string, tenantID string) ([]string, error) {
	return m.tagManager.GetTagNamesByIDs(ctx, tagIDs, tenantID)
}
