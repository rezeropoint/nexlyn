package config

import (
	"github.com/rezeropoint/etcdtrigger"

	casbinxCore "github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	CacheConf     cache.CacheConf
	DataSource    DataSource
	EtcdConfig    etcdtrigger.Config
	WatcherConfig casbinxCore.WatcherConfig
	Auth          struct {
		AccessSecret string
		AccessExpire int64
	}
	OSSConfig         OSSConfig
	SkylarkSyncEnabled bool             `json:",default=false"` // 是否启用Skylark同步
	AdminSyncClient   zrpc.RpcClientConf `json:",optional"`     // AdminSync gRPC客户端配置
}

type DataSource struct {
	Driver   string
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// OSSConfig 对象存储配置（仅支持阿里云OSS）
type OSSConfig struct {
	Enabled         bool   `json:",default=false"`   // 是否启用OSS功能
	Endpoint        string `json:",optional"`        // OSS端点（如：oss-cn-hangzhou.aliyuncs.com）
	AccessKeyID     string `json:",optional"`        // 访问密钥ID
	AccessKeySecret string `json:",optional"`        // 访问密钥Secret
	BucketName      string `json:",optional"`        // 存储桶名称
	CDNDomain       string `json:",optional"`        // CDN域名（可选，用于加速访问）
	MaxFileSize     int64  `json:",default=5242880"` // 最大文件大小（字节），默认5MB
}
