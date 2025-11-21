package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/rezeropoint/etcdtrigger/v2/engine"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// mqttManager MQTT管理器实现
type mqttManager struct {
	config Config

	// 配置存储（用于加载模板配置和监听配置变化）
	configStore engine.Engine

	// MQTT客户端
	client              mqtt.Client
	DispatchInfoFunc    core.DispatchInfoFunc
	StoreSensorDataFunc core.StoreSensorDataFunc

	// Redis客户端
	redisClient *redis.Redis

	// 主题订阅缓存（内存中的Set，使用go-zero 1.9泛型特性）
	subscribedTopics *collection.Set[string]

	// 状态管理
	mu            sync.RWMutex
	connected     bool
	started       bool
	watchersAdded sync.Once // 确保监听器只添加一次
}

// newMqttManager 创建MQTT管理器实例
func newMqttManager(config Config, configStore engine.Engine, redisClient *redis.Redis, DispatchInfoFunc core.DispatchInfoFunc, StoreSensorDataFunc core.StoreSensorDataFunc) (*mqttManager, error) {
	// 验证配置
	if config.Broker == "" {
		return nil, fmt.Errorf("config.Broker 未设置")
	}
	if config.ClientID == "" {
		return nil, fmt.Errorf("config.ClientID 未设置")
	}
	if redisClient == nil {
		return nil, fmt.Errorf("config.RedisClient 未设置")
	}
	if configStore == nil {
		return nil, fmt.Errorf("configStore 未设置")
	}

	manager := &mqttManager{
		config:              config,
		configStore:         configStore,
		redisClient:         redisClient,
		DispatchInfoFunc:    DispatchInfoFunc,
		StoreSensorDataFunc: StoreSensorDataFunc,
		subscribedTopics:    collection.NewSet[string](), // 初始化订阅主题缓存Set
		connected:           false,
		started:             false,
	}

	// 注意：不在这里添加配置监听器，而是在Start()方法中MQTT连接成功后再添加
	// 这样可以确保配置加载时MQTT已经连接，订阅不会失败

	return manager, nil
}

// Start 启动MQTT监控
func (m *mqttManager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.started {
		m.mu.Unlock()
		return fmt.Errorf("MQTT管理器已启动")
	}
	m.started = true
	m.mu.Unlock()

	// 创建MQTT客户端选项
	opts := mqtt.NewClientOptions()
	opts.AddBroker(m.config.Broker)
	opts.SetClientID(m.config.ClientID)

	// 设置认证
	if m.config.Username != "" {
		opts.SetUsername(m.config.Username)
		opts.SetPassword(m.config.Password)
	}

	// 设置连接回调
	opts.SetOnConnectHandler(m.onConnect)
	opts.SetConnectionLostHandler(m.onConnectionLost)

	// 设置自动重连
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(m.config.MaxReconnectInterval)

	// 设置清理会话
	opts.SetCleanSession(true)

	// 设置KeepAlive
	opts.SetKeepAlive(m.config.KeepAlive)

	// 设置连接超时
	opts.SetConnectTimeout(m.config.ConnectTimeout)

	// 设置Ping超时
	opts.SetPingTimeout(m.config.PingTimeout)

	// 创建客户端
	m.client = mqtt.NewClient(opts)

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_mqtt_manager"),
		logx.Field("operation", "start"),
		logx.Field("broker", m.config.Broker),
		logx.Field("client_id", m.config.ClientID),
	).Info("正在启动IoT MQTT管理器")

	// 连接到Broker
	if token := m.client.Connect(); !token.WaitTimeout(m.config.ConnectTimeout) {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "start"),
			logx.Field("status", "failed"),
			logx.Field("error", "connect timeout"),
		).Error("连接MQTT Broker超时")
		return context.DeadlineExceeded
	} else if token.Error() != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "start"),
			logx.Field("status", "failed"),
			logx.Field("error", token.Error().Error()),
		).Error("连接MQTT Broker失败")
		return token.Error()
	}

	m.mu.Lock()
	m.connected = true
	m.mu.Unlock()

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_mqtt_manager"),
		logx.Field("operation", "start"),
		logx.Field("status", "success"),
	).Info("IoT MQTT管理器启动成功")

	// 注意：配置监听器将在 onConnect 回调中添加，确保 MQTT 连接完全建立后再处理订阅

	return nil
}

// Stop 停止MQTT监控
func (m *mqttManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return
	}

	if m.client != nil && m.client.IsConnected() {
		logx.WithContext(context.Background()).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "stop"),
		).Info("正在停止IoT MQTT管理器")

		m.client.Disconnect(250)
		m.connected = false
	}

	// 清理Redis订阅记录
	if err := m.clearSubscribedTopics(); err != nil {
		logx.WithContext(context.Background()).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "stop"),
			logx.Field("status", "warning"),
			logx.Field("error", err.Error()),
		).Error("清理Redis订阅记录失败")
	}

	m.started = false
}

// IsConnected 检查MQTT连接状态
func (m *mqttManager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected && m.client != nil && m.client.IsConnected()
}

// onConnect 连接成功回调
func (m *mqttManager) onConnect(client mqtt.Client) {
	ctx := context.Background()

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_mqtt_manager"),
		logx.Field("operation", "on_connect"),
		logx.Field("status", "success"),
	).Info("IoT MQTT客户端已连接")

	m.mu.Lock()
	m.connected = true
	m.mu.Unlock()

	// 清理Redis订阅记录（因为CleanSession=true，MQTT重连后会话已清理）
	// 这确保Redis状态与MQTT Broker状态同步，避免残留记录导致跳过订阅
	if err := m.clearSubscribedTopics(); err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "on_connect"),
			logx.Field("status", "warning"),
			logx.Field("error", err.Error()),
		).Error("清理Redis订阅记录失败，将继续执行")
		// 清理失败不影响后续逻辑，日志记录即可
	} else {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "on_connect"),
		).Info("已清理旧的Redis订阅记录")
	}

	// 使用 sync.Once 确保监听器只在首次连接时添加（重连时跳过）
	// AddPrefixWatcher 会主动触发已有配置的回调，实现订阅
	m.watchersAdded.Do(func() {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "on_connect"),
		).Info("首次连接，开始添加配置监听器")

		// 添加在线检测配置监听器（监听sensor-online/前缀的配置变化）
		m.configStore.AddPrefixWatcher(core.OnlineConfigPrefix, m.onOnlineConfigChange)

		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "on_connect"),
		).Info("已添加在线检测配置监听器")

		// 添加业务数据配置监听器（监听sensor-data/前缀的配置变化）
		m.configStore.AddPrefixWatcher(core.DataProcessingConfigPrefix, m.onDataConfigChange)

		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "on_connect"),
		).Info("已添加业务数据配置监听器")
	})

	// 重新订阅所有已加载的配置（首次连接时补充确认，重连时恢复订阅）
	m.resubscribeAllConfigs()
}

// resubscribeAllConfigs 重新订阅所有配置（在线检测 + 业务数据）
func (m *mqttManager) resubscribeAllConfigs() {
	ctx := context.Background()

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_mqtt_manager"),
		logx.Field("operation", "resubscribe_all_configs"),
	).Info("开始重新订阅所有配置")

	// 1. 获取所有sensor-online/前缀的配置key
	onlineConfigKeys := m.configStore.GetAllKeys(core.OnlineConfigPrefix)

	onlineSubscribeCount := 0
	for _, key := range onlineConfigKeys {
		// 从ConfigManager获取配置
		var config core.OnlineDetectionConfig
		if !m.configStore.GetConfig(key, &config) {
			continue
		}

		// 构建订阅主题
		topic := core.BuildDeviceTopicPattern(core.DeviceCategory(config.Category), config.Model, config.TopicSuffix)

		// 订阅主题
		m.subscribeIfNeeded(topic)
		onlineSubscribeCount++
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_mqtt_manager"),
		logx.Field("operation", "resubscribe_all_configs"),
		logx.Field("config_count", len(onlineConfigKeys)),
		logx.Field("subscribe_count", onlineSubscribeCount),
	).Info("完成重新订阅所有在线检测配置")

	// 2. 获取所有sensor-data/前缀的配置key
	dataConfigKeys := m.configStore.GetAllKeys(core.DataProcessingConfigPrefix)

	dataSubscribeCount := 0
	for _, key := range dataConfigKeys {
		// 从ConfigManager获取配置
		var config core.DataProcessingConfig
		if !m.configStore.GetConfig(key, &config) {
			continue
		}

		// 构建订阅主题
		topic := core.BuildDeviceTopicPattern(core.DeviceCategory(config.Category), config.Model, config.TopicSuffix)

		// 订阅主题
		m.subscribeIfNeeded(topic)
		dataSubscribeCount++
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_mqtt_manager"),
		logx.Field("operation", "resubscribe_all_configs"),
		logx.Field("status", "success"),
		logx.Field("online_config_count", len(onlineConfigKeys)),
		logx.Field("online_subscribe_count", onlineSubscribeCount),
		logx.Field("data_config_count", len(dataConfigKeys)),
		logx.Field("data_subscribe_count", dataSubscribeCount),
	).Info("完成重新订阅所有业务数据配置")
}

// onConnectionLost 连接丢失回调
func (m *mqttManager) onConnectionLost(client mqtt.Client, err error) {
	logx.WithContext(context.Background()).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_mqtt_manager"),
		logx.Field("operation", "on_connection_lost"),
		logx.Field("error", err.Error()),
	).Error("IoT MQTT连接丢失，将自动重连")

	m.mu.Lock()
	m.connected = false
	m.mu.Unlock()
}

// subscribeIfNeeded 按需订阅主题（如果尚未订阅）
func (m *mqttManager) subscribeIfNeeded(topic string) {
	if !m.IsConnected() {
		logx.WithContext(context.Background()).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "subscribe_if_needed"),
			logx.Field("status", "skipped"),
			logx.Field("topic", topic),
			logx.Field("reason", "MQTT未连接"),
		).Error("MQTT未连接，无法订阅主题")
		return
	}

	// 检查是否已订阅（从Redis查询）
	subscribed, err := m.isTopicSubscribed(topic)
	if err != nil {
		logx.WithContext(context.Background()).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "subscribe_if_needed"),
			logx.Field("status", "warning"),
			logx.Field("topic", topic),
			logx.Field("error", err.Error()),
		).Error("检查主题订阅状态失败，将尝试订阅")
	} else if subscribed {
		return // 已订阅，跳过
	}

	// 订阅主题
	logx.WithContext(context.Background()).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_mqtt_manager"),
		logx.Field("operation", "subscribe"),
		logx.Field("topic", topic),
	).Info("订阅设备在线检测主题")

	if token := m.client.Subscribe(topic, 0, m.handleMessage); !token.WaitTimeout(m.config.PingTimeout) {
		logx.WithContext(context.Background()).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "subscribe"),
			logx.Field("status", "failed"),
			logx.Field("topic", topic),
			logx.Field("error", "timeout"),
		).Error("订阅主题超时")
		return
	} else if token.Error() != nil {
		logx.WithContext(context.Background()).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "subscribe"),
			logx.Field("status", "failed"),
			logx.Field("topic", topic),
			logx.Field("error", token.Error().Error()),
		).Error("订阅主题失败")
		return
	}

	// 标记已订阅（写入Redis）
	if err := m.markTopicSubscribed(topic); err != nil {
		logx.WithContext(context.Background()).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "subscribe"),
			logx.Field("status", "warning"),
			logx.Field("topic", topic),
			logx.Field("error", err.Error()),
		).Error("标记主题已订阅失败")
	}

	logx.WithContext(context.Background()).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_mqtt_manager"),
		logx.Field("operation", "subscribe"),
		logx.Field("status", "success"),
		logx.Field("topic", topic),
	).Info("成功订阅设备在线检测主题")
}

// handleMessage 处理MQTT消息（统一预处理，按需调用处理函数）
func (m *mqttManager) handleMessage(client mqtt.Client, msg mqtt.Message) {
	ctx := context.Background()

	// 1. 构建分布式锁key（基于topic和当前时间窗口）
	timestampMs := time.Now().UnixMilli()
	lockKey := core.BuildMessageLockKey(msg.Topic(), timestampMs, m.config.DedupWindowMs)

	// 2. 获取锁TTL配置（默认30秒）
	lockTTL := m.config.LockTTLSeconds
	if lockTTL <= 0 {
		lockTTL = 30
	}

	// 3. 尝试获取分布式锁
	acquired, err := m.tryAcquireLock(ctx, lockKey, lockTTL)
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "handle_message"),
			logx.Field("status", "warning"),
			logx.Field("topic", msg.Topic()),
			logx.Field("lock_key", lockKey),
			logx.Field("error", err.Error()),
		).Error("获取分布式锁失败，将继续处理")
		// 获取锁失败但不影响处理，继续执行
	} else if !acquired {
		// 未能获取锁，说明其他副本正在处理该时间窗口的消息
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "handle_message"),
			logx.Field("status", "skipped"),
			logx.Field("topic", msg.Topic()),
			logx.Field("lock_key", lockKey),
		).Debug("其他副本正在处理该时间窗口的消息，跳过")
		return
	}

	// 4. 确保处理完成后释放锁
	defer func() {
		if acquired {
			if err := m.releaseLock(ctx, lockKey); err != nil {
				logx.WithContext(ctx).WithFields(
					logx.Field("service", m.config.ServiceName),
					logx.Field("pod", m.config.PodName),
					logx.Field("module", "iot_mqtt_manager"),
					logx.Field("operation", "handle_message"),
					logx.Field("status", "warning"),
					logx.Field("lock_key", lockKey),
					logx.Field("error", err.Error()),
				).Error("释放分布式锁失败")
			}
		}
	}()

	// 5. 解析主题一次（避免重复解析）
	category, model, deviceID, topicSuffix, err := parseDeviceTopic(msg.Topic())
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "handle_message"),
			logx.Field("status", "failed"),
			logx.Field("topic", msg.Topic()),
			logx.Field("error", err.Error()),
		).Error("解析主题失败")
		return
	}

	// 6. 构建配置key
	onlineKey := core.BuildOnlineDetectionKey(string(category), model, topicSuffix)
	dataKey := core.BuildDataProcessingKey(string(category), model, topicSuffix)

	// 7. 检查配置存在性（快速内存查询）
	var onlineConfig core.OnlineDetectionConfig
	var dataConfig core.DataProcessingConfig

	hasOnline := m.configStore.GetConfig(onlineKey, &onlineConfig)
	hasData := m.configStore.GetConfig(dataKey, &dataConfig)

	// 8. 如果两个配置都不存在，记录一次DEBUG日志并返回
	if !hasOnline && !hasData {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "handle_message"),
			logx.Field("status", "skipped"),
			logx.Field("topic", msg.Topic()),
			logx.Field("category", category),
			logx.Field("model", model),
			logx.Field("device_id", deviceID),
			logx.Field("topic_suffix", topicSuffix),
			logx.Field("reason", "该主题未配置在线检测和业务数据处理"),
		).Debug("该主题未配置任何处理规则，跳过")
		return
	}

	// 9. 解析JSON消息体一次（按需，如果至少有一个配置存在）
	var data map[string]any
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_mqtt_manager"),
			logx.Field("operation", "handle_message"),
			logx.Field("status", "failed"),
			logx.Field("topic", msg.Topic()),
			logx.Field("device_id", deviceID),
			logx.Field("error", err.Error()),
		).Error("解析消息体JSON失败")
		return
	}

	// 10. 按需处理在线检测
	if hasOnline {
		m.processOnlineDetection(ctx, &onlineConfig, category, model, deviceID, onlineKey, data)
	}

	// 11. 按需处理业务数据提取
	if hasData {
		m.processBusinessData(ctx, &dataConfig, category, model, deviceID, topicSuffix, dataKey, data)
	}
}
