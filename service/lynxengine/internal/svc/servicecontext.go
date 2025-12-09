package svc

import (
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/engine"
	"github.com/rezeropoint/nexlyn/service/lynxengine/internal/config"
	"github.com/rezeropoint/nexlyn/service/lynxengine/internal/service"

	_ "github.com/lib/pq" // postgres 驱动
	skylarkEngine "github.com/rezeropoint/go-skylark/v2/engine"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/mon"
	"github.com/zeromicro/go-zero/core/stores/monc"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ServiceContext struct {
	Config     config.Config
	DBConn     sqlx.SqlConn
	MongoDB    *monc.Model
	LynxEngine engine.Engine
}

func NewServiceContext(c config.Config) *ServiceContext {
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
	// 配置 BSONOptions：让 MongoDB 驱动将文档解码为 map[string]any 而非 bson.D
	mongoClient, err := mongo.Connect(
		options.Client().
			ApplyURI(c.MongoDBConf.URI).
			SetBSONOptions(&options.BSONOptions{
				DefaultDocumentM: true, // 文档解码为 bson.M (即 map[string]any)
			}),
	)
	if err != nil {
		logx.Must(fmt.Errorf("创建 MongoDB 客户端失败: %w", err))
	}

	// 注入预配置的客户端，让 go-zero 的 monc 使用它
	mon.Inject(c.MongoDBConf.URI, mongoClient)

	mongoDB := monc.MustNewModel(
		c.MongoDBConf.URI,
		c.MongoDBConf.DB,
		c.MongoDBConf.Collection,
		c.MongoDBConf.CacheConf,
	)

	logx.Infof("MongoDB连接成功: %s/%s（已配置 map[string]any 解码）", c.MongoDBConf.DB, c.MongoDBConf.Collection)

	// 初始化服务注册列表
	var serviceRegistrations []engine.ServiceRegistration

	// 1. 初始化 IoTQuery gRPC 客户端（传感器数据服务）
	if c.IoTQueryClient.Endpoints != nil && len(c.IoTQueryClient.Endpoints) > 0 {
		iotQueryClient := zrpc.MustNewClient(c.IoTQueryClient)
		logx.Infof("IoTQuery gRPC客户端连接成功: %v", c.IoTQueryClient.Endpoints)

		// 创建传感器数据服务实例
		sensorDataService, err := service.NewSensorDataService(iotQueryClient)
		if err != nil {
			logx.Must(fmt.Errorf("创建传感器数据服务失败: %w", err))
		}

		// 注册服务
		serviceRegistrations = append(serviceRegistrations,
			engine.WithService(core.ServiceTypeSensorData, sensorDataService),
		)
		logx.Info("传感器数据服务已注册")
	} else {
		logx.Info("IoTQuery客户端配置为空，跳过传感器数据服务注册")
	}

	// 2. 初始化 Skylark Engine（流程引擎服务）
	// 初始化 Redis 客户端
	redisClient := redis.MustNewRedis(c.RedisConf)
	logx.Infof("Redis连接成功（Skylark引擎）: %s", c.RedisConf.Host)

	// 初始化 Skylark Engine（使用配置结构体，空配置会使用默认值）
	skylarkEngineInstance, err := skylarkEngine.NewSkylarkEngine(c.SkylarkConfig, dbConn, redisClient)
	if err != nil {
		logx.Must(fmt.Errorf("初始化 Skylark Engine 失败: %w", err))
	}

	// 注册服务
	serviceRegistrations = append(serviceRegistrations,
		engine.WithService(core.ServiceTypeSkylarkEngine, skylarkEngineInstance),
	)
	logx.Info("Skylark 流程引擎服务已注册")

	// 3. 初始化外部数据库服务（可选，用于 QueryDatabase 积木）
	if len(c.ExternalDataSources) > 0 {
		// 转换配置
		dbConfigs := make([]service.ExternalDBConfig, len(c.ExternalDataSources))
		for i, src := range c.ExternalDataSources {
			dbConfigs[i] = service.ExternalDBConfig{
				Name:         src.Name,
				Driver:       src.Driver,
				Host:         src.Host,
				Port:         src.Port,
				Database:     src.Database,
				Username:     src.Username,
				Password:     src.Password,
				MaxOpenConns: src.MaxOpenConns,
				MaxIdleConns: src.MaxIdleConns,
			}
		}

		externalDBService, err := service.NewExternalDBService(dbConfigs)
		if err != nil {
			logx.Must(fmt.Errorf("初始化外部数据库服务失败: %w", err))
		}

		serviceRegistrations = append(serviceRegistrations,
			engine.WithService(core.ServiceTypeExternalDB, externalDBService),
		)
		logx.Infof("外部数据库服务已注册，数据源数量: %d", len(c.ExternalDataSources))
	} else {
		logx.Info("未配置外部数据源，跳过外部数据库服务注册")
	}

	// 初始化 LynxGraph Engine（注入外部连接和服务）
	lynxEngine, err := engine.NewEngine(&c.LynxEngine, dbConn, mongoDB, serviceRegistrations)
	if err != nil {
		logx.Must(fmt.Errorf("初始化 LynxGraph Engine 失败: %w", err))
	}

	// 启动引擎
	if err := lynxEngine.Start(); err != nil {
		logx.Must(fmt.Errorf("启动 LynxGraph Engine 失败: %w", err))
	}

	logx.Info("LynxGraph Engine 初始化并启动成功")

	return &ServiceContext{
		Config:     c,
		DBConn:     dbConn,
		MongoDB:    mongoDB,
		LynxEngine: lynxEngine,
	}
}
