package svc

import (
	"errors"
	"fmt"
	"os"

	"github.com/rezeropoint/nexlyn/pkg/ossutil"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/rezeropoint/casbinx/core"
	"github.com/rezeropoint/casbinx/engine"

	_ "github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/zrpc"
)

var (
	// can't use one SingleFlight per conn, because multiple conns may share the same cache key.
	singleFlight = syncx.NewSingleFlight()
	stats        = cache.NewStat("backend")
)

type ServiceContext struct {
	PodName         string
	Config          config.Config
	Cache           cache.Cache
	DBConn          sqlx.SqlConn
	Casbinx         engine.CasbinX
	OSSClient       ossutil.Client
	AdminSyncClient pb.AdminSyncClient // AdminSync gRPC客户端
}

func NewServiceContext(c config.Config) *ServiceContext {

	podName := os.Getenv("POD_NAME")
	if podName == "" {
		podName = "none"
	}
	// 初始化 PostgreSQL - 使用 URL 格式的 DSN
	dbConn := sqlx.NewSqlConn(c.DataSource.Driver, fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DataSource.Username,
		c.DataSource.Password,
		c.DataSource.Host,
		c.DataSource.Port,
		c.DataSource.Database,
	))
	// 验证数据库连接
	if rawDB, err := dbConn.RawDB(); err != nil {
		logx.Must(fmt.Errorf("获取原生数据库连接失败: %w", err))
	} else if err := rawDB.Ping(); err != nil {
		logx.Must(fmt.Errorf("数据库连接验证失败: %w", err))
	}

	// 初始化 GORM 数据库连接（用于Casbin）
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		c.DataSource.Host,
		c.DataSource.Username,
		c.DataSource.Password,
		c.DataSource.Database,
		c.DataSource.Port,
	)
	casbinx, err := engine.NewCasbinx(core.Config{
		Dsn: dsn,
		PossiblePaths: []string{
			"etc/casbin_model.conf",
		},
		Watcher: c.WatcherConfig,
	})
	if err != nil {
		logx.Must(fmt.Errorf("Casbin初始化失败: %v", err))
	}

	// 初始化OSS客户端（可选，如果未启用则为nil）
	var ossClient ossutil.Client
	if c.OSSConfig.Enabled {
		// 验证OSS配置（仅支持阿里云）
		ossConfig := ossutil.Config{
			Endpoint:        c.OSSConfig.Endpoint,
			AccessKeyID:     c.OSSConfig.AccessKeyID,
			AccessKeySecret: c.OSSConfig.AccessKeySecret,
			BucketName:      c.OSSConfig.BucketName,
			CDNDomain:       c.OSSConfig.CDNDomain,
		}

		if err := ossutil.ValidateConfig(ossConfig); err != nil {
			logx.Errorf("OSS配置验证失败，头像上传功能将不可用: %v", err)
		} else {
			// 创建OSS客户端
			ossClient, err = ossutil.NewClient(ossConfig)
			if err != nil {
				logx.Errorf("创建OSS客户端失败，头像上传功能将不可用: %v", err)
			} else {
				logx.Info("OSS客户端初始化成功")
			}
		}
	} else {
		logx.Info("OSS功能未启用，头像上传功能将不可用")
	}

	// 初始化 AdminSync gRPC 客户端（可选，根据配置决定是否启用）
	var adminSyncClient pb.AdminSyncClient
	if c.SkylarkSyncEnabled && len(c.AdminSyncClient.Endpoints) > 0 {
		conn := zrpc.MustNewClient(c.AdminSyncClient)
		adminSyncClient = pb.NewAdminSyncClient(conn.Conn())
		logx.Infof("AdminSync gRPC客户端连接成功: %v", c.AdminSyncClient.Endpoints)
	} else {
		if !c.SkylarkSyncEnabled {
			logx.Info("Skylark同步功能未启用")
		} else {
			logx.Info("AdminSync客户端配置为空，同步功能将不可用")
		}
	}

	// 先构造 ServiceContext，返回地址；m7s 实例异步就绪后再注入字段
	return &ServiceContext{
		Config:          c,
		Cache:           cache.New(c.CacheConf, singleFlight, stats, errors.New("缓存未命中")),
		DBConn:          dbConn,
		Casbinx:         casbinx,
		PodName:         podName,
		OSSClient:       ossClient,
		AdminSyncClient: adminSyncClient,
	}

}
