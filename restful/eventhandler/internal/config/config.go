package config

import (
	"time"

	casbinxCore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	RedisConf      redis.RedisConf
	DataSource     DataSource
	WatcherConfig  casbinxCore.WatcherConfig
	SkylarkEngine  SkylarkEngineConfig
	Auth           struct {
		AccessSecret string
		AccessExpire int64
	}
}

// DataSource 数据库配置
type DataSource struct {
	Driver   string
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// SkylarkEngineConfig Skylark查询引擎配置
type SkylarkEngineConfig struct {
	// ========== 查询限制 ==========
	MaxPageSize     int           `json:",default=1000"`  // 最大分页大小
	DefaultPageSize int           `json:",default=20"`    // 默认分页大小
	QueryTimeout    time.Duration `json:",default=30s"`   // 查询超时时间
	
	// ========== 远程数据库连接池配置 ==========
	MaxOpenConns    int           `json:",default=10"`    // 最大打开连接数
	MaxIdleConns    int           `json:",default=5"`     // 最大空闲连接数
	ConnMaxLifetime time.Duration `json:",default=30m"`   // 连接最大生命周期
	ConnMaxIdleTime time.Duration `json:",default=10m"`   // 连接最大空闲时间
	
	// ========== Redis缓存配置 ==========
	EnableCache       bool          `json:",default=true"`  // 是否启用Redis缓存
	UserCacheTTL      time.Duration `json:",default=24h"`   // 用户名缓存过期时间
	FlowListCacheTTL  time.Duration `json:",default=1h"`    // flows列表缓存过期时间
	FlowFieldCacheTTL time.Duration `json:",default=1h"`    // flow字段列表缓存过期时间
	StatsCacheTTL     time.Duration `json:",default=5m"`    // 统计数据缓存过期时间
}
