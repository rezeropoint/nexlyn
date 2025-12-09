package config

import (
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/engine"

	skylarkEngine "github.com/rezeropoint/go-skylark/v2/engine"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	// go-zero gRPC服务配置
	zrpc.RpcServerConf

	// PostgreSQL数据库配置
	DataSource struct {
		Driver   string
		Host     string
		Port     int
		Database string
		Username string
		Password string
	}

	// MongoDB配置
	MongoDBConf struct {
		URI        string
		DB         string
		Collection string
		CacheConf  cache.CacheConf
	}

	// Redis配置
	RedisConf redis.RedisConf

	// LynxGraph引擎配置（使用pkg/lynxgraph提供的真实配置结构）
	LynxEngine engine.Config

	// Skylark引擎配置
	SkylarkConfig skylarkEngine.Config

	// IoTQuery gRPC客户端配置
	IoTQueryClient zrpc.RpcClientConf

	// 外部数据源配置（用于 QueryDatabase 积木）
	ExternalDataSources []ExternalDataSourceConfig
}

// ExternalDataSourceConfig 外部数据源配置
type ExternalDataSourceConfig struct {
	Name         string // 数据源名称（唯一标识，积木中通过此名称引用）
	Driver       string // 数据库驱动：postgres
	Host         string
	Port         int
	Database     string
	Username     string
	Password     string
	MaxOpenConns int // 最大打开连接数，默认 10
	MaxIdleConns int // 最大空闲连接数，默认 5
}
