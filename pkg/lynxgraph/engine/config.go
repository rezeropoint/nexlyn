package engine

import (
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/block"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/dispatcher"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/scheduler"

	"github.com/rezeropoint/etcdtrigger"
	"github.com/zeromicro/go-zero/core/stores/cache"
)

// EngineConfig 引擎配置
type Config struct {
	RunMode   core.Mode
	KeyPrefix string
	// 图上下文的默认过期时间
	GraphContextTTL time.Duration
	// 信息原子的默认过期时间
	InfoAtomTTL time.Duration

	// 子模块配置
	CacheConf        cache.CacheConf    // infoatom 包内创建 cache
	EtcdConfig       etcdtrigger.Config // graph 包内创建 etcd
	BlockConfig      block.Config       // block 配置（空配置）
	DispatcherConfig dispatcher.Config  // dispatcher 业务配置
	SchedulerConfig  scheduler.Config   // scheduler 定时调度配置
}
