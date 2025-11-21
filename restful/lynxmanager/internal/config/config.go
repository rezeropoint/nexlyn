package config

import (
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/manager"

	casbinxCore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	// go-zero REST服务配置
	rest.RestConf

	// JWT认证配置
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}

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

	// LynxGraph管理器配置（使用pkg/lynxgraph提供的真实配置结构）
	LynxManager manager.Config

	// Casbin权限管理配置
	WatcherConfig casbinxCore.WatcherConfig
}
