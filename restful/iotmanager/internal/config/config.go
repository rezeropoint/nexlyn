package config

import (
	"time"

	casbinxCore "github.com/rezeropoint/casbinx/core"
	"github.com/rezeropoint/go-skylark/v2/engine"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	RequestTimeout   time.Duration
	CacheConf        cache.CacheConf
	RedisConf        redis.RedisConf
	DataSource       DataSource
	EtcdConfig       EtcdConfig
	MqttConfig       MqttConfig
	ClickHouseConfig ClickHouseConfig
	SkylarkConfig    engine.Config
	LynxGraphConfig  LynxGraphConfig
	WatcherConfig    casbinxCore.WatcherConfig
	Auth             struct {
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

// EtcdConfig Etcd 配置
type EtcdConfig struct {
	Hosts    []string `json:"hosts"`
	User     string   `json:"user,optional"`
	Pass     string   `json:"pass,optional"`
	Configs  []string `json:"configs"` // 配置键前缀列表
	RootPath string   `json:"rootPath,optional"`
}

// MqttConfig MQTT Broker 配置
type MqttConfig struct {
	Broker               string        `json:"broker"`
	ClientID             string        `json:"clientId"`
	Username             string        `json:"username,optional"`
	Password             string        `json:"password,optional"`
	ConnectTimeout       time.Duration `json:"connectTimeout,default=5s"`
	PingTimeout          time.Duration `json:"pingTimeout,default=10s"`
	KeepAlive            time.Duration `json:"keepAlive,default=60s"`
	MaxReconnectInterval time.Duration `json:"maxReconnectInterval,default=10s"`
	CapabilitiesCacheTTL int           `json:"capabilitiesCacheTTL,default=7200"` // 算法能力缓存时长（秒），默认7200秒（2小时），设置为0表示禁用缓存
}

// ClickHouseConfig ClickHouse 时序数据库配置
type ClickHouseConfig struct {
	Address         string         `json:"address"`                         // ClickHouse地址（如：localhost:9000）
	Database        string         `json:"database"`                        // 数据库名称
	Username        string         `json:"username,optional"`               // 用户名
	Password        string         `json:"password,optional"`               // 密码
	MaxOpenConns    int            `json:"maxOpenConns,default=10"`         // 最大连接数
	MaxIdleConns    int            `json:"maxIdleConns,default=5"`          // 最大空闲连接数
	ConnMaxLifetime time.Duration  `json:"connMaxLifetime,default=3600s"`   // 连接最大生命周期
	BatchSize       int            `json:"batchSize,default=100"`           // 批量写入大小（预留，当前未使用）
	FlushInterval   time.Duration  `json:"flushInterval,default=1s"`        // 刷新间隔（预留，当前未使用）
	TTLOverrides    map[string]int `json:"ttlOverrides,optional"`           // TTL配置覆盖（字段名 -> TTL天数，可选）
}

// LynxGraphConfig LynxGraph 逻辑引擎配置
type LynxGraphConfig struct {
	GRPCURL string `json:"grpcUrl"` // gRPC 服务地址（如：lynxengine:9999）
}
