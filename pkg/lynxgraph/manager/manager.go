package manager

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/zeromicro/go-zero/core/stores/monc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Manager interface {
	CheckInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey) bool
	GetInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey) (core.InfoAtomType, error)                                 // 因为只有创建信息原子时才会使用这个接口，因此可以不保存在内存中。因此决定使用mongodb+redis来存储
	GetInfoAtomTypeList(ctx context.Context, ListRequest core.InfoAtomList) ([]core.InfoAtomType, int64, int64, int64, error) // 查询信息原子类型
	CreateInfoAtomType(ctx context.Context, infoAtomType core.InfoAtomType) error                                             // 创建新的信息原子类型
	UpdateInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey, infoAtomType core.InfoAtomType) error                   // 更新信息原子类型
	DeleteInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey) error                                                   // 删除信息原子类型

	GetGraphConfig(ctx context.Context, key core.GraphKey) (core.GraphConfig, error)                                    // GetGraph 根据key获取逻辑图
	GetGraphConfigList(ctx context.Context, ListParams core.GraphList) ([]core.GraphConfig, int64, int64, int64, error) // GetGraphConfigs 分页获取逻辑图配置列表
	CreateGraphConfig(ctx context.Context, config core.GraphConfig) error                                               // CreateGraph 创建逻辑图
	UpdateGraphConfig(ctx context.Context, config core.GraphConfig) error                                               // UpdateGraph 更新逻辑图
	DeleteGraphConfig(ctx context.Context, key core.GraphKey) error                                                     // DeleteGraph 删除逻辑图

	GetBlockKeyAll() []core.BlockKey                             // GetBlockKeyAll 获取所有逻辑块的key
	GetBlockSpec(blockKey core.BlockKey) (core.BlockSpec, error) // GetBlockSpec 获取逻辑块规格

	// ValidateNodeConfig 验证节点配置（创建积木实例并调用 SetConfigure 验证必填字段）
	ValidateNodeConfig(nodeConfig core.NodeConfig) error

	// Tag管理接口
	CreateTag(ctx context.Context, metadata core.LynxTagMetadata) (string, error)                                      // 创建标签
	GetTag(ctx context.Context, tagID string, tenantID string) (*core.LynxTag, error)                                  // 获取标签详情
	ListTags(ctx context.Context, query core.LynxTagQuery) ([]*core.LynxTagSummary, int64, error)                      // 查询标签列表
	UpdateTag(ctx context.Context, tagID string, tenantID string, update core.LynxTagUpdate) error                     // 更新标签
	DeleteTag(ctx context.Context, tagID string, tenantID string) error                                                // 删除标签
	GetTagsByNames(ctx context.Context, tenantID string, scope string, names []string) ([]*core.LynxTagSummary, error) // 批量查询标签
	GetTagsByIDs(ctx context.Context, tagIDs []string, tenantID string) ([]*core.LynxTagSummary, error)                // Phase 2.5: 根据标签ID列表批量查询标签摘要（包含 id、name、description、scope）
	GetTagNamesByIDs(ctx context.Context, tagIDs []string, tenantID string) ([]string, error)                          // 根据标签ID列表批量查询标签名称（供 REST 层使用）
}

type ManagerOption func(*manager) error

func NewManager(
	config *Config,
	sqlConn sqlx.SqlConn,
	mongoDB *monc.Model,
	opts ...ManagerOption) (Manager, error) {
	return newManager(config, sqlConn, mongoDB, opts...)
}
