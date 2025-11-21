package graph

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/zeromicro/go-zero/core/stores/monc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// GraphRegistry 是逻辑图注册表接口，用于管理和创建逻辑图
type GraphRegistry interface {
	InitGraphTable(ctx context.Context) error // 初始化数据库表

	// 注意：加载图时不需要信息原子类型都定义，只是未定义会无法创建信息原子而无法调度
	GetGraphConfig(ctx context.Context, key core.GraphKey) (core.GraphConfig, error)                                    // GetGraph 根据key获取逻辑图配置
	GetGraphConfigList(ctx context.Context, ListParams core.GraphList) ([]core.GraphConfig, int64, int64, int64, error) // GetGraphConfigs 分页获取逻辑图配置列表
	CreateGraphConfig(ctx context.Context, config core.GraphConfig) error                                               // CreateGraph 创建逻辑图配置
	UpdateGraphConfig(ctx context.Context, config core.GraphConfig) error                                               // UpdateGraph 更新逻辑图配置
	DeleteGraphConfig(ctx context.Context, key core.GraphKey) error                                                     // DeleteGraph 删除逻辑图配置

	GetGraph(key core.GraphKey) (core.LogicGraph, error)                                // GetGraph 根据key获取逻辑图
	FindGraphsByInfoAtom(infoAtom core.InfoAtom) (map[core.GraphKey][]core.Node, error) // FindGraphsByInfoAtom 根据信息原子查找逻辑图

	Close() error // Close 关闭图注册表，清理资源
}

// NewGraphRegistry 创建一个新的逻辑图注册表
// 参数：
//   - ctx, cancel: 上下文和取消函数
//   - config: 业务配置（包含 EtcdConfig）
//   - sqlConn: PostgreSQL 连接（外部注入）
//   - mongoDB: MongoDB Model（外部注入）
//   - createBlockFunc: 逻辑块创建函数
//   - getTagNamesFunc: 根据标签ID获取标签名称的函数（用于标签查询）
func NewGraphRegistry(
	ctx context.Context,
	cancel context.CancelFunc,
	config *Config,
	sqlConn sqlx.SqlConn,
	mongoDB *monc.Model,
	createBlockFunc core.CreateBlockFunc,
	getTagNamesFunc core.GetTagNamesByIDsFunc) (GraphRegistry, error) {
	if config == nil {
		return nil, ErrConfigNil
	}

	return newGraphRegistry(ctx, cancel, config, sqlConn, mongoDB, createBlockFunc, getTagNamesFunc)
}
