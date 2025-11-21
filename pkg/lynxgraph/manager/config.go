package manager

import (
	"errors"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/rezeropoint/etcdtrigger"

	"github.com/zeromicro/go-zero/core/stores/cache"
)

type Config struct {
	RunMode    core.Mode
	CacheConf  cache.CacheConf    // infoatom 包内创建 cache
	EtcdConfig etcdtrigger.Config // graph 包内创建 etcd
}

var (
	ErrConfigNil              = errors.New("配置不能为空")
	ErrInfoAtomRegistryFailed = errors.New("创建信息原子注册表失败")
	ErrGraphRegistryFailed    = errors.New("创建图注册表失败")
	ErrBlockRegistryFailed    = errors.New("创建块注册表失败")
	ErrStandardBlocksFailed   = errors.New("注册标准逻辑块失败")
)
