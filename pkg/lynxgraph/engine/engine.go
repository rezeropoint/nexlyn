package engine

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/zeromicro/go-zero/core/stores/monc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Engine interface {
	// 基本操作
	Start() error    // 启动引擎
	Stop() error     // 停止引擎
	IsRunning() bool // 判断引擎是否正在运行

	// 信息原子
	GetInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey) (core.InfoAtomType, error)      // 获取信息原子类型
	ReceiveInfoAtom(ctx context.Context, req core.InfoAtomRequest) (*core.InfoAtomResponse, error) // 接收信息原子（统一入口）

	// 事件分发
	// 在网关侧，首先就需要判断信息原子的类型，因为逻辑图根据信息原子运行单个或数个。
	// 因此引擎微服务需要通过gRPC来接收信息原子，而网关侧能进行核验每个接口是否收到了预期的信息原子类型。
	Dispatch(infoAtom core.InfoAtom) error // 分发事件

	// 图管理
	//RegisterGraph(graph core.LogicGraph) error       // 注册图
	//UnregisterGraph(id string, version string) error // 注销图
	//LoadGraph(config *graph.GraphConfig) error       // 从配置加载图
	//GetGraph(id string) (interface{}, bool)               // 获取图
	//ListGraphs() []string                                 // 列出所有图

	// 逻辑块管理
	GetBlockKeyAll() []core.BlockKey
	//RegisterBlock(name string, block LogicBlock) // 注册逻辑块

	// 状态管理
	//SetStateStore(store core.StateStore) // 设置状态存储
}

// 引擎创建后，首先是加载信息原子的类型，然后加载逻辑块、逻辑图。最后是调度器和数据存储。
// 引擎启动后，首先是启动调度器，让传入可查询信息原子和遍历逻辑图的接口让调度器编制索引，然后是启动数据存储。
// 引擎停止后，首先是停止数据存储，然后是停止调度器。

// EngineOption 表示引擎选项函数
type EngineOption func(*engine) error

// ServiceRegistration 服务注册信息
type ServiceRegistration struct {
	ServiceType core.ServiceType
	Service     interface{}
}

// WithService 创建一个服务注册信息
// 参数：
//   - serviceType: 服务类型常量（如 core.ServiceTypeSkylarkEngine）
//   - service: 服务实例（如 SkylarkEngine）
//
// 注意：每个服务类型只能注册一个实例
func WithService(serviceType core.ServiceType, service interface{}) ServiceRegistration {
	return ServiceRegistration{
		ServiceType: serviceType,
		Service:     service,
	}
}

// NewEngine 创建一个新的引擎实例
func NewEngine(
	config *Config,
	sqlConn sqlx.SqlConn,
	mongoDB *monc.Model,
	services []ServiceRegistration,
	options ...EngineOption) (Engine, error) {
	return newwEngine(config, sqlConn, mongoDB, services, options...)
}
