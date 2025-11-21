package svc

import (
	"fmt"
	"os"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/query"
	"github.com/rezeropoint/nexlyn/service/iotquery/internal/config"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config      config.Config
	QueryEngine query.Query // pkg/iot/query查询引擎实例
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 1. 初始化PostgreSQL连接
	pgDSN := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable pool_max_conns=%d pool_max_conn_lifetime=%ds",
		c.PostgreSQL.Host,
		c.PostgreSQL.Port,
		c.PostgreSQL.Username,
		c.PostgreSQL.Password,
		c.PostgreSQL.Database,
		c.PostgreSQL.MaxConns,
		c.PostgreSQL.MaxLifetime,
	)
	dbConn := sqlx.NewSqlConn("postgres", pgDSN)
	logx.Infof("Connected to PostgreSQL: %s:%d/%s (MaxConns=%d)",
		c.PostgreSQL.Host, c.PostgreSQL.Port, c.PostgreSQL.Database, c.PostgreSQL.MaxConns)

	// 2. 初始化ClickHouse连接
	clickHouseConn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", c.ClickHouse.Host, c.ClickHouse.Port)},
		Auth: clickhouse.Auth{
			Database: c.ClickHouse.Database,
			Username: c.ClickHouse.Username,
			Password: c.ClickHouse.Password,
		},
		MaxOpenConns:    c.ClickHouse.MaxOpenConns,
		MaxIdleConns:    c.ClickHouse.MaxIdleConns,
		ConnMaxLifetime: time.Duration(c.ClickHouse.MaxLifetime) * time.Second,
		DialTimeout:     time.Duration(c.ClickHouse.ConnTimeout) * time.Second,
		ReadTimeout:     time.Duration(c.ClickHouse.ReadTimeout) * time.Second,
	})
	if err != nil {
		logx.Severef("Failed to connect to ClickHouse: %v", err)
		panic(err)
	}
	logx.Infof("Connected to ClickHouse: %s:%d/%s (MaxConns=%d, MaxIdleConns=%d)",
		c.ClickHouse.Host, c.ClickHouse.Port, c.ClickHouse.Database,
		c.ClickHouse.MaxOpenConns, c.ClickHouse.MaxIdleConns)

	// 3. 创建查询引擎
	queryEngine, err := query.New(query.Config{
		ClickHouseDatabase: c.ClickHouse.Database,
		DefaultLimit:       1000,
		MaxLimit:           10000,
		QueryTimeout:       time.Duration(c.Timeout) * time.Millisecond, // 使用配置的超时
		MaxFieldsPerReq:    10,
		ServiceName:        c.Name,
		PodName:            os.Getenv("POD_NAME"),
	}, dbConn, clickHouseConn)
	if err != nil {
		logx.Severef("Failed to create query engine: %v", err)
		panic(err)
	}
	logx.Infof("Query engine initialized successfully (QueryTimeout=%dms)", c.Timeout)

	return &ServiceContext{
		Config:      c,
		QueryEngine: queryEngine,
	}
}
