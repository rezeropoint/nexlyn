package config

import (
	"github.com/rezeropoint/go-skylark/v2/admin"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
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

	// Redis配置（用于Skylark缓存）
	RedisConf redis.RedisConf

	// AdminEngine配置
	AdminConfig admin.Config
}
