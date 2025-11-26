package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core/devices"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/internal/controller"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/internal/device"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/internal/dispatcher"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/internal/httpreceive"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/internal/metadata"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/internal/mqtt"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/internal/platform"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/internal/query"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/internal/storage"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/internal/tag"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/internal/template"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/internal/workgroup"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	etcdtriggercore "github.com/rezeropoint/etcdtrigger/v2/core"
	etcdtrigger "github.com/rezeropoint/etcdtrigger/v2/engine"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/rezeropoint/go-skylark/v2/engine"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// iotClient IoT引擎客户端实现
type iotClient struct {
	templateManager    template.Manager      // 模板管理器
	deviceManager      device.Manager        // 设备管理器
	tagManager         tag.Manager           // 标签管理器
	platformManager    platform.Manager      // 平台配置管理器
	metadataManager    metadata.Manager      // 元数据管理器
	mqttManager        mqtt.Manager          // MQTT管理器（实现OnlineStatusProvider接口）
	queryManager       query.QueryManager    // 时序数据查询管理器
	controllerManager  controller.Manager    // Controller管理器（设备控制）
	httpReceiveManager httpreceive.Manager   // HTTP数据接收管理器
}

// newIoTClient 创建IoT引擎客户端
func newIoTClient(ctx context.Context, config Config, dbConn sqlx.SqlConn, redisClient *redis.Redis, clickHouseConn driver.Conn, skylarkEngine engine.SkylarkEngine) (*iotClient, error) {
	// 验证配置
	if len(config.EtcdHosts) == 0 {
		return nil, fmt.Errorf("config.EtcdHosts 未设置")
	}

	// 创建 etcd 客户端
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   config.EtcdHosts,
		Username:    config.EtcdUser,
		Password:    config.EtcdPass,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("创建etcd客户端失败: %w", err)
	}

	// 创建共享的 Engine（供Template、Platform、MQTT Manager复用）
	// 注意：
	// - sensor-template-{uuid}：模板配置，仅写入不监听（静态配置，通过API修改）
	// - sensor-online/：在线检测配置，需要监听以触发MQTT订阅更新
	// - sensor-data/：业务数据处理配置，需要监听以触发MQTT订阅更新
	// - platform-{type}/ 平台配置，仅写入不监听（静态配置，通过API修改）
	configStore := etcdtrigger.NewEngine(etcdClient, &etcdtrigger.Config{
		PodName:     config.PodName,
		ServiceName: config.ServiceName,
		// 初始化配置监听器（仅注册需要热更新的配置）
		Configs: []etcdtriggercore.WatchConfig{
			{
				Path:   "sensor-online/", // 在线检测配置（MQTT Manager监听，动态更新订阅）
				Struct: &core.OnlineDetectionConfig{},
			},
			{
				Path:   "sensor-data/", // 业务数据处理配置（MQTT Manager监听，动态更新订阅）
				Struct: &core.DataProcessingConfig{},
			},
			{
				Path:   "control-config/", // 设备控制配置
				Struct: &core.DeviceControlConfig{},
			},
		},
	})

	// 创建模板管理器（复用 Engine）
	templateMgr, err := template.NewManager(dbConn, configStore, template.Config{})
	if err != nil {
		return nil, fmt.Errorf("创建模板管理器失败: %w", err)
	}

	// 注意：MQTT Manager 需要在 Dispatcher Manager 之后创建（依赖 DispatchInfo 函数）
	// Device Manager 需要在 MQTT Manager 之后创建（依赖 MQTT Manager的在线状态查询函数）
	// 先占位为 nil
	var mqttMgr mqtt.Manager
	var deviceMgr device.Manager

	// 创建标签管理器
	tagMgr, err := tag.NewManager(dbConn, tag.Config{})
	if err != nil {
		return nil, fmt.Errorf("创建标签管理器失败: %w", err)
	}

	// 创建平台配置管理器（复用 Engine）
	platformMgr, err := platform.NewManager(dbConn, platform.Config{
		ServiceName: config.ServiceName,
		PodName:     config.PodName,
	}, configStore) // 传入 configStore
	if err != nil {
		return nil, fmt.Errorf("创建平台配置管理器失败: %w", err)
	}

	// 创建元数据管理器
	metadataMgr := metadata.NewManager()

	// 创建 Workgroup Manager（用于异步数据分发）
	workerCount := config.WorkerCount
	if workerCount == 0 {
		workerCount = 10 // 默认10个工作协程
	}
	workgroupMgr := workgroup.NewManager(workgroup.Config{
		WorkerCount: workerCount,
	})

	// 创建 Dispatcher Manager（用于数据分发）
	// 注意：只有在提供了 SkylarkEngine 时才能使用 Skylark 分发功能
	dispatcherMgr := dispatcher.NewManager(
		dispatcher.Config{
			ServiceName:      config.ServiceName,
			PodName:          config.PodName,
			LynxGraphGRPCURL: config.LynxGraphGRPCURL,
		},
		func(id string) (*core.PlatformConfig, bool) {
			// 从 platformManager 获取平台配置
			// 注意：这里需要先查询元数据获取 type 和 tenantID
			// 为了简化，这里返回 nil（如果需要使用，需要扩展 platform.Manager 接口）
			return nil, false
		},
		skylarkEngine,
		workgroupMgr.SubmitAsync,
	)

	// 创建 Storage Manager（用于传感器数据存储）
	// 注意：clickHouseConn 由外部（REST层）创建并传入
	var storageMgr storage.StorageManager
	if clickHouseConn != nil {
		// 在创建Storage Manager前验证TTL配置完整性
		if err := validateTTLConfig(config.TTLOverrides); err != nil {
			return nil, fmt.Errorf("TTL配置验证失败: %w", err)
		}

		storageMgr, err = storage.NewManager(ctx, clickHouseConn, storage.Config{
			Database:     config.ClickHouseDatabase,
			TTLOverrides: config.TTLOverrides,
			ServiceName:  config.ServiceName,
			PodName:      config.PodName,
		})
		if err != nil {
			return nil, fmt.Errorf("创建Storage Manager失败: %w", err)
		}
	}

	// 创建Query Manager（用于时序数据查询）
	// 注意：clickHouseConn 由外部（REST层）创建并传入
	var queryMgr query.QueryManager
	if clickHouseConn != nil {
		queryMgr, err = query.NewQueryManager(dbConn, clickHouseConn, query.Config{
			// ClickHouse配置
			Database:        config.ClickHouseDatabase,
			DefaultLimit:    1000,  // 默认返回1000条数据
			MaxLimit:        10000, // 最多返回10000条数据
			QueryTimeout:    30,    // 查询超时30秒
			MaxFieldsPerReq: 10,    // 单次请求最多10个字段
			ServiceName:     config.ServiceName,
			PodName:         config.PodName,
		})
		if err != nil {
			return nil, fmt.Errorf("创建Query Manager失败: %w", err)
		}
	}

	// 设置MQTT默认值（统一设置，MQTT Manager和Controller Manager共用）
	if config.MQTTBroker != "" {
		if config.MQTTConnectTimeout == 0 {
			config.MQTTConnectTimeout = 10 * time.Second
		}
		if config.MQTTPingTimeout == 0 {
			config.MQTTPingTimeout = 10 * time.Second
		}
		if config.MQTTKeepAlive == 0 {
			config.MQTTKeepAlive = 60 * time.Second
		}
		if config.MQTTMaxReconnectInterval == 0 {
			config.MQTTMaxReconnectInterval = 10 * time.Minute
		}
		if config.MQTTRedisTTL == 0 {
			config.MQTTRedisTTL = 300 // 默认5分钟
		}
	}

	// 创建MQTT管理器（在Dispatcher Manager和Storage Manager之后，需要传递相关函数）
	if config.MQTTBroker != "" {
		// 验证Redis配置（MQTT Manager需要Redis）
		if redisClient == nil {
			return nil, fmt.Errorf("启用MQTT功能需要配置redisClient")
		}

		// 构建StoreSensorDataFunc（如果Storage Manager已创建）
		var storeSensorDataFunc core.StoreSensorDataFunc
		if storageMgr != nil {
			storeSensorDataFunc = storageMgr.StoreSensorData
		}

		mqttMgr, err = mqtt.NewManager(mqtt.Config{
			Broker:               config.MQTTBroker,
			ClientID:             config.MQTTClientID,
			Username:             config.MQTTUsername,
			Password:             config.MQTTPassword,
			ConnectTimeout:       config.MQTTConnectTimeout,
			PingTimeout:          config.MQTTPingTimeout,
			KeepAlive:            config.MQTTKeepAlive,
			MaxReconnectInterval: config.MQTTMaxReconnectInterval,
			ServiceName:          config.ServiceName,
			PodName:              config.PodName,
			RedisTTL:             config.MQTTRedisTTL,
		}, configStore, redisClient, dispatcherMgr.DispatchInfo, storeSensorDataFunc) // 传递数据分发函数和存储函数
		if err != nil {
			return nil, fmt.Errorf("创建MQTT管理器失败: %w", err)
		}

		// 启动MQTT监控
		if err := mqttMgr.Start(context.Background()); err != nil {
			return nil, fmt.Errorf("启动MQTT管理器失败: %w", err)
		}
	}

	// 创建设备管理器（需要在MQTT Manager之后，注入在线状态查询函数）
	if mqttMgr != nil {
		// MQTT Manager已创建，注入其在线状态查询函数
		deviceMgr, err = device.NewManager(
			dbConn,
			redisClient,
			device.Config{},
			mqttMgr.GetDeviceOnlineStatus, // 注入在线状态查询函数
			mqttMgr.GetAllOnlineDevices,   // 注入所有在线设备查询函数
		)
		if err != nil {
			return nil, fmt.Errorf("创建设备管理器失败: %w", err)
		}
	} else {
		// MQTT Manager未启用，Device Manager无法获取在线状态
		// 返回错误，因为Device Manager依赖MQTT Manager
		return nil, fmt.Errorf("创建设备管理器失败: MQTT Manager未启用，Device Manager依赖MQTT Manager提供在线状态查询功能")
	}

	// 创建Controller Manager（用于设备控制）
	var controllerMgr controller.Manager
	if config.MQTTBroker != "" {
		// 验证Redis配置（Controller Manager需要Redis用于命令追踪）
		if redisClient == nil {
			return nil, fmt.Errorf("启用Controller功能需要配置redisClient")
		}

		// 创建GetDeviceInfo适配器函数（通过Device Manager获取，不进行组织权限验证）
		// 注意：权限验证将在Engine层的AI Box控制方法中统一处理
		getDeviceInfoFunc := func(deviceID string) (*core.DeviceBinding, error) {
			return deviceMgr.GetDeviceInfoForController(context.Background(), deviceID)
		}

		// 创建Controller Manager
		controllerMgr, err = controller.NewManager(
			controller.Config{
				Broker:               config.MQTTBroker,
				ClientID:             config.MQTTClientID + "-controller", // 使用不同的ClientID
				Username:             config.MQTTUsername,
				Password:             config.MQTTPassword,
				ConnectTimeout:       config.MQTTConnectTimeout,
				PingTimeout:          config.MQTTPingTimeout,
				KeepAlive:            config.MQTTKeepAlive,
				MaxReconnectInterval: config.MQTTMaxReconnectInterval,
				CapabilitiesCacheTTL: config.CapabilitiesCacheTTL, // 算法能力缓存时长
				ServiceName:          config.ServiceName,
				PodName:              config.PodName,
			},
			redisClient,       // Redis客户端（用于命令追踪）
			configStore,       // 配置存储（用于读取控制配置）
			getDeviceInfoFunc, // GetDeviceInfo函数（用于获取设备信息）
		)
		if err != nil {
			return nil, fmt.Errorf("创建Controller管理器失败: %w", err)
		}

		// 启动Controller Manager
		if err := controllerMgr.Start(context.Background()); err != nil {
			return nil, fmt.Errorf("启动Controller管理器失败: %w", err)
		}
	}

	// 创建 HTTP 数据接收管理器
	httpReceiveMgr, err := httpreceive.NewManager(
		dbConn,
		httpreceive.Config{
			ServiceName: config.ServiceName,
			PodName:     config.PodName,
		},
		configStore,
		dispatcherMgr.DispatchInfo,
	)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP数据接收管理器失败: %w", err)
	}

	return &iotClient{
		templateManager:    templateMgr,
		deviceManager:      deviceMgr,
		tagManager:         tagMgr,
		platformManager:    platformMgr,
		metadataManager:    metadataMgr,
		mqttManager:        mqttMgr,
		queryManager:       queryMgr,
		controllerManager:  controllerMgr,
		httpReceiveManager: httpReceiveMgr,
	}, nil
}

// 实现 IoT 接口 - 模板管理方法

func (c *iotClient) CreateTemplate(ctx context.Context, metadata core.TemplateMetadata, onlineConfig *core.OnlineDetectionConfig, businessConfig *core.DataProcessingConfig, controlConfig *core.DeviceControlConfig) (string, error) {
	return c.templateManager.Create(ctx, metadata, onlineConfig, businessConfig, controlConfig)
}

func (c *iotClient) UpdateTemplate(ctx context.Context, id string, metadata core.TemplateMetadata, onlineConfig *core.OnlineDetectionConfig, businessConfig *core.DataProcessingConfig, controlConfig *core.DeviceControlConfig) error {
	return c.templateManager.Update(ctx, id, metadata, onlineConfig, businessConfig, controlConfig)
}

func (c *iotClient) DeleteTemplate(ctx context.Context, id string) error {
	return c.templateManager.Delete(ctx, id)
}

func (c *iotClient) GetTemplate(ctx context.Context, id string) (*core.Template, error) {
	return c.templateManager.Get(ctx, id)
}

func (c *iotClient) ListTemplates(ctx context.Context, query core.TemplateQuery) ([]*core.TemplateSummary, int64, error) {
	return c.templateManager.List(ctx, query)
}

// 实现 IoT 接口 - 设备绑定管理方法

func (c *iotClient) BindDevice(ctx context.Context, metadata core.DeviceBindingMetadata) (string, error) {
	return c.deviceManager.Bind(ctx, metadata)
}

func (c *iotClient) GetDevice(ctx context.Context, deviceID string, tenantID string, orgIDs []string) (*core.DeviceBinding, error) {
	return c.deviceManager.Get(ctx, deviceID, tenantID, orgIDs)
}

func (c *iotClient) ListDevices(ctx context.Context, query core.DeviceBindingQuery) ([]*core.DeviceBindingSummary, int64, error) {
	return c.deviceManager.List(ctx, query)
}

func (c *iotClient) ListUnboundDevices(ctx context.Context, tenantID string) ([]core.UnboundDevice, error) {
	return c.deviceManager.ListUnboundDevices(ctx, tenantID)
}

func (c *iotClient) UpdateDevice(ctx context.Context, deviceID string, tenantID string, orgIDs []string, update core.DeviceBindingUpdate) error {
	return c.deviceManager.Update(ctx, deviceID, tenantID, orgIDs, update)
}

func (c *iotClient) DeleteDevice(ctx context.Context, deviceID string, tenantID string, orgIDs []string) error {
	return c.deviceManager.Delete(ctx, deviceID, tenantID, orgIDs)
}

// 实现 IoT 接口 - 标签管理方法

func (c *iotClient) CreateTag(ctx context.Context, metadata core.DeviceTagMetadata) (string, error) {
	return c.tagManager.CreateTag(ctx, metadata)
}

func (c *iotClient) GetTag(ctx context.Context, tagID string, tenantID string) (*core.DeviceTag, error) {
	return c.tagManager.GetTag(ctx, tagID, tenantID)
}

func (c *iotClient) ListTags(ctx context.Context, query core.DeviceTagQuery) ([]*core.DeviceTagSummary, int64, error) {
	return c.tagManager.ListTags(ctx, query)
}

func (c *iotClient) UpdateTag(ctx context.Context, tagID string, tenantID string, update core.DeviceTagUpdate) error {
	return c.tagManager.UpdateTag(ctx, tagID, tenantID, update)
}

func (c *iotClient) DeleteTag(ctx context.Context, tagID string, tenantID string) error {
	return c.tagManager.DeleteTag(ctx, tagID, tenantID)
}

// 实现 IoT 接口 - 元数据查询方法

func (c *iotClient) GetDeviceCategories() []core.DeviceCategoryInfo {
	return c.metadataManager.GetDeviceCategories()
}

func (c *iotClient) GetStandardFields(category core.DeviceCategory) []core.StandardField {
	return c.metadataManager.GetStandardFields(category)
}

// 实现 IoT 接口 - 平台配置管理方法

func (c *iotClient) CreatePlatform(ctx context.Context, metadata core.PlatformMetadata, config *core.PlatformConfig) (string, error) {
	return c.platformManager.Create(ctx, metadata, config)
}

func (c *iotClient) GetPlatform(ctx context.Context, platformType core.PlatformType, id string, tenantID string) (*core.Platform, error) {
	return c.platformManager.Get(ctx, platformType, id, tenantID)
}

func (c *iotClient) ListPlatforms(ctx context.Context, tenantID string, platformType *core.PlatformType, keyword string, page, pageSize int) ([]*core.Platform, int64, error) {
	return c.platformManager.List(ctx, tenantID, platformType, keyword, page, pageSize)
}

func (c *iotClient) UpdatePlatform(ctx context.Context, platformType core.PlatformType, id string, tenantID string, metadata core.PlatformMetadata, config *core.PlatformConfig) error {
	return c.platformManager.Update(ctx, platformType, id, tenantID, metadata, config)
}

func (c *iotClient) DeletePlatform(ctx context.Context, platformType core.PlatformType, id string, tenantID string) error {
	return c.platformManager.Delete(ctx, platformType, id, tenantID)
}

// 实现 IoT 接口 - 时序数据查询方法

func (c *iotClient) QueryTimeSeries(ctx context.Context, query core.TimeSeriesQuery) (*core.TimeSeriesResult, error) {
	if c.queryManager == nil {
		return nil, fmt.Errorf("Query Manager未初始化，请检查ClickHouse配置")
	}
	return c.queryManager.QueryTimeSeries(ctx, query)
}

func (c *iotClient) GetLatestValues(ctx context.Context, query core.LatestValuesQuery) ([]core.DeviceLatestValues, error) {
	if c.queryManager == nil {
		return nil, fmt.Errorf("Query Manager未初始化，请检查ClickHouse配置")
	}
	return c.queryManager.GetLatestValues(ctx, query)
}

func (c *iotClient) GetDeviceStatistics(ctx context.Context, query core.DeviceStatisticsQuery) ([]core.DeviceStatistics, error) {
	if c.queryManager == nil {
		return nil, fmt.Errorf("Query Manager未初始化，请检查ClickHouse配置")
	}
	return c.queryManager.GetDeviceStatistics(ctx, query)
}

// 实现 IoT 接口 - AI Box算法任务管理方法

func (c *iotClient) CreateAIBoxTask(ctx context.Context, tenantID, deviceID string, orgIDs []string, task *devices.AIBoxAlgorithmTask) (string, error) {
	// 1. 验证Controller Manager是否已初始化
	if c.controllerManager == nil {
		return "", fmt.Errorf("Controller Manager未初始化，请检查MQTT配置")
	}

	// 2. 验证设备权限（确保设备属于指定租户和组织）
	_, err := c.deviceManager.Get(ctx, deviceID, tenantID, orgIDs)
	if err != nil {
		return "", fmt.Errorf("设备权限验证失败: %w", err)
	}

	// 3. 调用Controller Manager创建任务
	return c.controllerManager.CreateAIBoxTask(ctx, deviceID, task)
}

func (c *iotClient) UpdateAIBoxTask(ctx context.Context, tenantID, deviceID string, orgIDs []string, taskID string, task *devices.AIBoxAlgorithmTask) error {
	// 1. 验证Controller Manager是否已初始化
	if c.controllerManager == nil {
		return fmt.Errorf("Controller Manager未初始化，请检查MQTT配置")
	}

	// 2. 验证设备权限（确保设备属于指定租户和组织）
	_, err := c.deviceManager.Get(ctx, deviceID, tenantID, orgIDs)
	if err != nil {
		return fmt.Errorf("设备权限验证失败: %w", err)
	}

	// 3. 调用Controller Manager更新任务
	return c.controllerManager.UpdateAIBoxTask(ctx, deviceID, taskID, task)
}

func (c *iotClient) DeleteAIBoxTask(ctx context.Context, tenantID, deviceID string, orgIDs []string, taskID string) error {
	// 1. 验证Controller Manager是否已初始化
	if c.controllerManager == nil {
		return fmt.Errorf("Controller Manager未初始化，请检查MQTT配置")
	}

	// 2. 验证设备权限（确保设备属于指定租户和组织）
	_, err := c.deviceManager.Get(ctx, deviceID, tenantID, orgIDs)
	if err != nil {
		return fmt.Errorf("设备权限验证失败: %w", err)
	}

	// 3. 调用Controller Manager删除任务
	return c.controllerManager.DeleteAIBoxTask(ctx, deviceID, taskID)
}

func (c *iotClient) ListAIBoxTasks(ctx context.Context, tenantID, deviceID string, orgIDs []string) ([]devices.AIBoxAlgorithmTask, error) {
	// 1. 验证Controller Manager是否已初始化
	if c.controllerManager == nil {
		return nil, fmt.Errorf("Controller Manager未初始化，请检查MQTT配置")
	}

	// 2. 验证设备权限（确保设备属于指定租户和组织）
	_, err := c.deviceManager.Get(ctx, deviceID, tenantID, orgIDs)
	if err != nil {
		return nil, fmt.Errorf("设备权限验证失败: %w", err)
	}

	// 3. 调用Controller Manager查询任务列表
	return c.controllerManager.ListAIBoxTasks(ctx, deviceID)
}

func (c *iotClient) GetAIBoxCapabilities(ctx context.Context, tenantID, deviceID string, orgIDs []string) (*devices.AIBoxCapabilities, error) {
	// 1. 验证Controller Manager是否已初始化
	if c.controllerManager == nil {
		return nil, fmt.Errorf("Controller Manager未初始化，请检查MQTT配置")
	}

	// 2. 验证设备权限（确保设备属于指定租户和组织）
	_, err := c.deviceManager.Get(ctx, deviceID, tenantID, orgIDs)
	if err != nil {
		return nil, fmt.Errorf("设备权限验证失败: %w", err)
	}

	// 3. 调用Controller Manager获取算法能力
	return c.controllerManager.GetAIBoxCapabilities(ctx, deviceID)
}

func (c *iotClient) ControlAIBoxTask(ctx context.Context, tenantID, deviceID string, orgIDs []string, taskID string, controlCommand int) error {
	// 1. 验证Controller Manager是否已初始化
	if c.controllerManager == nil {
		return fmt.Errorf("Controller Manager未初始化，请检查MQTT配置")
	}

	// 2. 验证设备权限（确保设备属于指定租户和组织）
	_, err := c.deviceManager.Get(ctx, deviceID, tenantID, orgIDs)
	if err != nil {
		return fmt.Errorf("设备权限验证失败: %w", err)
	}

	// 3. 调用Controller Manager控制任务
	return c.controllerManager.ControlAIBoxTask(ctx, deviceID, taskID, controlCommand)
}

// 实现 IoT 接口 - HTTP数据接收配置管理方法

func (c *iotClient) CreateHttpReceiveConfig(ctx context.Context, metadata core.HttpReceiveMetadata, config *core.HttpReceiveConfig) (string, error) {
	return c.httpReceiveManager.Create(ctx, metadata, config)
}

func (c *iotClient) GetHttpReceiveConfig(ctx context.Context, id string, tenantID string) (*core.HttpReceive, error) {
	return c.httpReceiveManager.Get(ctx, id, tenantID)
}

func (c *iotClient) ListHttpReceiveConfigs(ctx context.Context, query core.HttpReceiveQuery) ([]*core.HttpReceiveSummary, int64, error) {
	return c.httpReceiveManager.List(ctx, query)
}

func (c *iotClient) UpdateHttpReceiveConfig(ctx context.Context, id string, tenantID string, metadata core.HttpReceiveMetadata, config *core.HttpReceiveConfig) error {
	return c.httpReceiveManager.Update(ctx, id, tenantID, metadata, config)
}

func (c *iotClient) DeleteHttpReceiveConfig(ctx context.Context, id string, tenantID string) error {
	return c.httpReceiveManager.Delete(ctx, id, tenantID)
}

func (c *iotClient) ProcessHttpReceiveData(ctx context.Context, configId string, data map[string]any) error {
	return c.httpReceiveManager.ProcessData(ctx, configId, data)
}
