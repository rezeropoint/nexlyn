package svc

import (
	"fmt"

	"github.com/rezeropoint/nexlyn/service/eventsync/internal/config"

	_ "github.com/lib/pq"
	"github.com/rezeropoint/go-skylark/v2/admin"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config      config.Config
	AdminEngine admin.AdminEngine
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 1. 初始化 PostgreSQL 连接
	dbConn := sqlx.NewSqlConn(c.DataSource.Driver, fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
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

	// 2. 初始化 Redis 客户端
	redisClient := redis.MustNewRedis(c.RedisConf)
	logx.Infof("Redis连接成功: %s", c.RedisConf.Host)

	// 3. 初始化 AdminEngine
	adminEngine, err := admin.NewAdminEngine(c.AdminConfig, dbConn, redisClient)
	if err != nil {
		logx.Must(fmt.Errorf("初始化 AdminEngine 失败: %w", err))
	}

	logx.Info("AdminEngine 初始化成功")

	return &ServiceContext{
		Config:      c,
		AdminEngine: adminEngine,
	}
}
