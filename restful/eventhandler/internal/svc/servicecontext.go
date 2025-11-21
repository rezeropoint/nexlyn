package svc

import (
	"fmt"
	"os"

	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/config"

	_ "github.com/lib/pq" // postgres 支持
	"github.com/rezeropoint/casbinx/core"
	"github.com/rezeropoint/casbinx/engine"
	skylarkengine "github.com/rezeropoint/go-skylark/v2/engine"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	PodName       string
	Config        config.Config
	DBConn        sqlx.SqlConn
	Casbinx       engine.CasbinX
	SkylarkEngine skylarkengine.SkylarkEngine // Skylark引擎
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
		logx.Must(fmt.Errorf("casbin初始化失败: %v", err))
	}

	// 初始化 Redis Client
	redisClient := redis.MustNewRedis(c.RedisConf)

	// 初始化 Skylark 引擎（使用默认配置，具体配置项已内置到各个Manager）
	skylarkCfg := skylarkengine.Config{
		// 新版本的Config结构已模块化，使用空结构体会应用默认配置
		// 如需自定义配置，可以设置Cache、Platform、Event、Mapping、Query、Stats等子配置
	}

	skylarkEngine, err := skylarkengine.NewSkylarkEngine(skylarkCfg, dbConn, redisClient)
	if err != nil {
		logx.Must(fmt.Errorf("skylark引擎初始化失败: %v", err))
	}

	return &ServiceContext{
		PodName:       podName,
		Config:        c,
		DBConn:        dbConn,
		Casbinx:       casbinx,
		SkylarkEngine: skylarkEngine,
	}
}
