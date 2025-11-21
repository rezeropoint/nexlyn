package svc

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/config"

	"github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/lib/pq" // postgres 支持
	"github.com/rezeropoint/casbinx/core"
	"github.com/rezeropoint/casbinx/engine"
	skylarkengine "github.com/rezeropoint/go-skylark/v2/engine"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	iotengine "github.com/rezeropoint/nexlyn/pkg/lynxiot/engine"
)

type ServiceContext struct {
	PodName   string
	Config    config.Config
	DBConn    sqlx.SqlConn
	Casbinx   engine.CasbinX
	IoTEngine iotengine.IoT // IoT引擎
}

func NewServiceContext(c config.Config) *ServiceContext {
	ctx := context.Background()

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

	// 初始化 Redis Client
	redisClient := redis.MustNewRedis(c.RedisConf)

	// 初始化 Skylark Engine（新版需要dbConn参数）
	skylarkEngine, err := skylarkengine.NewSkylarkEngine(c.SkylarkConfig, dbConn, redisClient)
	if err != nil {
		logx.Must(fmt.Errorf("初始化 Skylark Engine 失败: %w", err))
	}

	// 初始化 ClickHouse 连接
	clickHouseConn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{c.ClickHouseConfig.Address},
		Auth: clickhouse.Auth{
			Database: c.ClickHouseConfig.Database,
			Username: c.ClickHouseConfig.Username,
			Password: c.ClickHouseConfig.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		MaxOpenConns:    c.ClickHouseConfig.MaxOpenConns,
		MaxIdleConns:    c.ClickHouseConfig.MaxIdleConns,
		ConnMaxLifetime: c.ClickHouseConfig.ConnMaxLifetime,
		DialTimeout:     5 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})
	if err != nil {
		logx.Must(fmt.Errorf("连接ClickHouse失败: %w", err))
	}

	// 测试ClickHouse连接
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := clickHouseConn.Ping(pingCtx); err != nil {
		logx.Must(fmt.Errorf("Ping ClickHouse失败: %w", err))
	}

	logx.Infof("ClickHouse连接成功: %s", c.ClickHouseConfig.Address)

	// 初始化 IoT 引擎（传入ClickHouse连接）
	iotEngine, err := iotengine.New(ctx, iotengine.Config{
		// Etcd配置
		EtcdHosts: c.EtcdConfig.Hosts,
		EtcdUser:  c.EtcdConfig.User,
		EtcdPass:  c.EtcdConfig.Pass,

		// MQTT配置
		MQTTBroker:               c.MqttConfig.Broker,
		MQTTClientID:             c.MqttConfig.ClientID + "-iot-engine",
		MQTTUsername:             c.MqttConfig.Username,
		MQTTPassword:             c.MqttConfig.Password,
		MQTTConnectTimeout:       c.MqttConfig.ConnectTimeout,
		MQTTPingTimeout:          c.MqttConfig.PingTimeout,
		MQTTKeepAlive:            c.MqttConfig.KeepAlive,
		MQTTMaxReconnectInterval: c.MqttConfig.MaxReconnectInterval,
		MQTTRedisTTL:             300, // 5分钟

		// 控制配置
		CapabilitiesCacheTTL: c.MqttConfig.CapabilitiesCacheTTL,

		// ClickHouse配置
		ClickHouseDatabase: c.ClickHouseConfig.Database,
		TTLOverrides:       c.ClickHouseConfig.TTLOverrides,

		// LynxGraph配置
		LynxGraphGRPCURL: c.LynxGraphConfig.GRPCURL,

		// Pod信息
		PodName:     podName,
		ServiceName: c.Name,
	}, dbConn, redisClient, clickHouseConn, skylarkEngine)
	if err != nil {
		logx.Must(fmt.Errorf("初始化 IoT 引擎失败: %w", err))
	}

	return &ServiceContext{
		Config:    c,
		DBConn:    dbConn,
		Casbinx:   casbinx,
		PodName:   podName,
		IoTEngine: iotEngine,
	}
}
