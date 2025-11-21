package infoatom

import (
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/zeromicro/go-zero/core/stores/cache"
)

// Config 信息原子注册表配置
type Config struct {
	RunMode   core.Mode
	CacheConf cache.CacheConf // 缓存配置（包内创建缓存）
}

// 常量定义
const (
	TableName                  = "lynxgraph_info_atom_types" // 信息原子类型表名
	DefaultInfoAtomTypeVersion = "v1"                        // 默认信息原子类型版本
)
