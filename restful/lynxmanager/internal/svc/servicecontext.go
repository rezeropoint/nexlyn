// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"fmt"
	"os"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/manager"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/config"

	_ "github.com/lib/pq" // postgres 驱动
	"github.com/rezeropoint/casbinx/core"
	"github.com/rezeropoint/casbinx/engine"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/monc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	PodName     string
	Config      config.Config
	DBConn      sqlx.SqlConn
	MongoDB     *monc.Model
	Casbinx     engine.CasbinX
	LynxManager manager.Manager
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 获取Pod名称
	podName := os.Getenv("POD_NAME")
	if podName == "" {
		podName = "none"
	}

	// 初始化 PostgreSQL 连接 - 使用 URL 格式的 DSN
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

	logx.Infof("PostgreSQL连接成功: %s:%d/%s", c.DataSource.Host, c.DataSource.Port, c.DataSource.Database)

	// 初始化 MongoDB 连接（带缓存）
	mongoDB := monc.MustNewModel(
		c.MongoDBConf.URI,
		c.MongoDBConf.DB,
		c.MongoDBConf.Collection,
		c.MongoDBConf.CacheConf,
	)

	logx.Infof("MongoDB连接成功: %s/%s", c.MongoDBConf.DB, c.MongoDBConf.Collection)

	// 初始化 CasbinX 权限管理（用于GORM的DSN格式）
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
			"../../pkg/nexlyn/etc/casbin_model.conf", // 开发环境路径
		},
		Watcher: c.WatcherConfig,
	})
	if err != nil {
		logx.Must(fmt.Errorf("初始化 CasbinX 失败: %w", err))
	}

	logx.Info("CasbinX 权限管理初始化成功")

	// 初始化 LynxGraph Manager（注入外部连接）
	lynxManager, err := manager.NewManager(&c.LynxManager, dbConn, mongoDB)
	if err != nil {
		logx.Must(fmt.Errorf("初始化 LynxGraph Manager 失败: %w", err))
	}

	logx.Info("LynxGraph Manager 初始化成功")

	return &ServiceContext{
		PodName:     podName,
		Config:      c,
		DBConn:      dbConn,
		MongoDB:     mongoDB,
		Casbinx:     casbinx,
		LynxManager: lynxManager,
	}
}
