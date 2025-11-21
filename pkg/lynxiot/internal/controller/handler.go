package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core/devices"

	"github.com/rezeropoint/etcdtrigger/v2/engine"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// pendingCommand 待处理命令信息
type pendingCommand struct {
	responseCh    chan *devices.AIBoxControlResponse
	expectedEvent string // 期望的Event类型，用于响应验证
}

// controllerManager Controller管理器实现
type controllerManager struct {
	config Config

	converter *aiboxConverter

	// MQTT客户端
	client mqtt.Client

	// 命令追踪（deviceID -> pendingCommand）
	pendingCommands sync.Map

	// 依赖注入（运行时实例）
	redisClient   *redis.Redis           // Redis客户端（用于命令调试）
	configStore   engine.Engine          // 配置存储（用于读取控制配置）
	getDeviceInfo core.GetDeviceInfoFunc // 获取设备信息（category, model, tenant_id）

	// 状态管理
	mu        sync.RWMutex
	connected bool
	started   bool
}

// newControllerManager 创建Controller管理器实例
func newControllerManager(config Config, redisClient *redis.Redis, configStore engine.Engine, getDeviceInfo core.GetDeviceInfoFunc) (*controllerManager, error) {
	// 验证配置
	if config.Broker == "" {
		return nil, fmt.Errorf("config.Broker 未设置")
	}
	if config.ClientID == "" {
		return nil, fmt.Errorf("config.ClientID 未设置")
	}

	// 验证依赖
	if redisClient == nil {
		return nil, fmt.Errorf("redisClient 未设置")
	}
	if configStore == nil {
		return nil, fmt.Errorf("configStore 未设置")
	}
	if getDeviceInfo == nil {
		return nil, fmt.Errorf("getDeviceInfo 函数未设置")
	}

	manager := &controllerManager{
		config:        config,
		converter:     newAIBoxConverter(),
		redisClient:   redisClient,
		configStore:   configStore,
		getDeviceInfo: getDeviceInfo,
		connected:     false,
		started:       false,
	}

	return manager, nil
}

// Start 启动Controller管理器
func (m *controllerManager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.started {
		m.mu.Unlock()
		return fmt.Errorf("Controller管理器已启动")
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
		logx.Field("module", "iot_controller_manager"),
		logx.Field("operation", "start"),
		logx.Field("broker", m.config.Broker),
		logx.Field("client_id", m.config.ClientID),
	).Info("正在启动IoT Controller管理器")

	// 连接到Broker
	if token := m.client.Connect(); !token.WaitTimeout(m.config.ConnectTimeout) {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller_manager"),
			logx.Field("operation", "start"),
			logx.Field("status", "failed"),
			logx.Field("error", "connect timeout"),
		).Error("连接MQTT Broker超时")
		return context.DeadlineExceeded
	} else if token.Error() != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller_manager"),
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
		logx.Field("module", "iot_controller_manager"),
		logx.Field("operation", "start"),
		logx.Field("status", "success"),
	).Info("IoT Controller管理器启动成功")

	return nil
}

// Stop 停止Controller管理器
func (m *controllerManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return
	}

	if m.client != nil && m.client.IsConnected() {
		logx.WithContext(context.Background()).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller_manager"),
			logx.Field("operation", "stop"),
		).Info("正在停止IoT Controller管理器")

		m.client.Disconnect(250)
		m.connected = false
	}

	m.started = false
}

// IsConnected 检查MQTT连接状态
func (m *controllerManager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected && m.client != nil && m.client.IsConnected()
}

// onConnect 连接成功回调
func (m *controllerManager) onConnect(client mqtt.Client) {
	ctx := context.Background()

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_controller_manager"),
		logx.Field("operation", "on_connect"),
		logx.Field("status", "success"),
	).Info("IoT Controller MQTT客户端已连接")

	m.mu.Lock()
	m.connected = true
	m.mu.Unlock()
}

// onConnectionLost 连接丢失回调
func (m *controllerManager) onConnectionLost(client mqtt.Client, err error) {
	logx.WithContext(context.Background()).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_controller_manager"),
		logx.Field("operation", "on_connection_lost"),
		logx.Field("error", err.Error()),
	).Error("IoT Controller MQTT连接丢失，将自动重连")

	m.mu.Lock()
	m.connected = false
	m.mu.Unlock()
}

// CreateAIBoxTask 创建AI Box算法任务
func (m *controllerManager) CreateAIBoxTask(ctx context.Context, deviceID string, task *devices.AIBoxAlgorithmTask) (string, error) {
	// 1. 获取设备信息
	deviceInfo, err := m.getDeviceInfo(deviceID)
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "create_aibox_task"),
			logx.Field("device_id", deviceID),
			logx.Field("error", err.Error()),
		).Error("获取设备信息失败")
		return "", core.ErrDeviceNotFound
	}

	// 2. 检查设备在线状态
	if !deviceInfo.IsOnline {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "create_aibox_task"),
			logx.Field("device_id", deviceID),
		).Error("设备离线")
		return "", core.ErrDeviceOffline
	}

	// 3. 读取控制配置
	controlConfig, err := m.loadControlConfig(ctx, deviceInfo.DeviceCategory, deviceInfo.DeviceModel)
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "create_aibox_task"),
			logx.Field("device_id", deviceID),
			logx.Field("category", deviceInfo.DeviceCategory),
			logx.Field("model", deviceInfo.DeviceModel),
			logx.Field("error", err.Error()),
		).Error("读取控制配置失败")
		return "", err
	}

	// 4. 生成AlgTaskSession（如果未提供）
	if task.AlgTaskSession == "" {
		task.AlgTaskSession = uuid.New().String()
	}

	// 5. 构建命令
	command := &devices.AIBoxControlCommand{
		CommandType: devices.AIBoxCommandCreateTask,
		DeviceID:    deviceID,
		Model:       devices.AIBoxModel(deviceInfo.DeviceModel),
		Payload:     task,
		Timestamp:   time.Now().UnixMilli(),
	}

	// 6. 发送命令并等待响应
	response, err := m.sendCommandAndWait(ctx, deviceInfo, controlConfig, command)
	if err != nil {
		return "", err
	}

	// 7. 检查响应状态
	if response.Status != "success" {
		return "", fmt.Errorf("创建任务失败: %s", response.Error)
	}

	// 8. 返回AlgTaskSession
	return task.AlgTaskSession, nil
}

// UpdateAIBoxTask 更新AI Box算法任务
func (m *controllerManager) UpdateAIBoxTask(ctx context.Context, deviceID, taskID string, task *devices.AIBoxAlgorithmTask) error {
	// 1. 获取设备信息
	deviceInfo, err := m.getDeviceInfo(deviceID)
	if err != nil {
		return core.ErrDeviceNotFound
	}

	// 2. 检查设备在线状态
	if !deviceInfo.IsOnline {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "update_aibox_task"),
			logx.Field("device_id", deviceID),
		).Error("设备离线")
		return core.ErrDeviceOffline
	}

	// 3. 读取控制配置
	controlConfig, err := m.loadControlConfig(ctx, deviceInfo.DeviceCategory, deviceInfo.DeviceModel)
	if err != nil {
		return err
	}

	// 4. 设置AlgTaskSession
	task.AlgTaskSession = taskID

	// 5. 构建命令
	command := &devices.AIBoxControlCommand{
		CommandType: devices.AIBoxCommandUpdateTask,
		DeviceID:    deviceID,
		Model:       devices.AIBoxModel(deviceInfo.DeviceModel),
		Payload:     task,
		Timestamp:   time.Now().UnixMilli(),
	}

	// 6. 发送命令并等待响应
	response, err := m.sendCommandAndWait(ctx, deviceInfo, controlConfig, command)
	if err != nil {
		return err
	}

	// 7. 检查响应状态
	if response.Status != "success" {
		return fmt.Errorf("更新任务失败: %s", response.Error)
	}

	return nil
}

// DeleteAIBoxTask 删除AI Box算法任务
func (m *controllerManager) DeleteAIBoxTask(ctx context.Context, deviceID, taskID string) error {
	// 1. 获取设备信息
	deviceInfo, err := m.getDeviceInfo(deviceID)
	if err != nil {
		return core.ErrDeviceNotFound
	}

	// 2. 检查设备在线状态
	if !deviceInfo.IsOnline {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "delete_aibox_task"),
			logx.Field("device_id", deviceID),
		).Error("设备离线")
		return core.ErrDeviceOffline
	}

	// 3. 读取控制配置
	controlConfig, err := m.loadControlConfig(ctx, deviceInfo.DeviceCategory, deviceInfo.DeviceModel)
	if err != nil {
		return err
	}

	// 4. 构建命令
	command := &devices.AIBoxControlCommand{
		CommandType: devices.AIBoxCommandDeleteTask,
		DeviceID:    deviceID,
		Model:       devices.AIBoxModel(deviceInfo.DeviceModel),
		Payload:     map[string]string{"task_id": taskID},
		Timestamp:   time.Now().UnixMilli(),
	}

	// 5. 发送命令并等待响应
	response, err := m.sendCommandAndWait(ctx, deviceInfo, controlConfig, command)
	if err != nil {
		return err
	}

	// 6. 检查响应状态
	if response.Status != "success" {
		return fmt.Errorf("删除任务失败: %s", response.Error)
	}

	return nil
}

// ListAIBoxTasks 查询AI Box算法任务列表
func (m *controllerManager) ListAIBoxTasks(ctx context.Context, deviceID string) ([]devices.AIBoxAlgorithmTask, error) {
	// 1. 获取设备信息
	deviceInfo, err := m.getDeviceInfo(deviceID)
	if err != nil {
		return nil, core.ErrDeviceNotFound
	}

	// 2. 检查设备在线状态
	if !deviceInfo.IsOnline {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "list_aibox_tasks"),
			logx.Field("device_id", deviceID),
		).Error("设备离线")
		return nil, core.ErrDeviceOffline
	}

	// 3. 读取控制配置
	controlConfig, err := m.loadControlConfig(ctx, deviceInfo.DeviceCategory, deviceInfo.DeviceModel)
	if err != nil {
		return nil, err
	}

	// 4. 构建命令
	command := &devices.AIBoxControlCommand{
		CommandType: devices.AIBoxCommandListTasks,
		DeviceID:    deviceID,
		Model:       devices.AIBoxModel(deviceInfo.DeviceModel),
		Payload:     nil,
		Timestamp:   time.Now().UnixMilli(),
	}

	// 5. 发送命令并等待响应
	response, err := m.sendCommandAndWait(ctx, deviceInfo, controlConfig, command)
	if err != nil {
		return nil, err
	}

	// 6. 检查响应状态
	if response.Status != "success" {
		return nil, fmt.Errorf("查询任务列表失败: %s", response.Error)
	}

	// 7. 解析任务列表
	tasks, ok := response.Result.([]devices.AIBoxAlgorithmTask)
	if !ok {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	return tasks, nil
}

// GetAIBoxCapabilities 获取AI Box算法能力
func (m *controllerManager) GetAIBoxCapabilities(ctx context.Context, deviceID string) (*devices.AIBoxCapabilities, error) {
	// 1. 尝试从缓存获取
	if cachedCapabilities, err := m.getCapabilitiesFromCache(ctx, deviceID); err == nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "get_aibox_capabilities"),
			logx.Field("device_id", deviceID),
			logx.Field("from_cache", true),
		).Info("从缓存返回算法能力")
		return cachedCapabilities, nil
	}

	// 2. 缓存未命中，从设备查询
	deviceInfo, err := m.getDeviceInfo(deviceID)
	if err != nil {
		return nil, core.ErrDeviceNotFound
	}

	// 3. 检查设备在线状态
	if !deviceInfo.IsOnline {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "get_aibox_capabilities"),
			logx.Field("device_id", deviceID),
		).Error("设备离线")
		return nil, core.ErrDeviceOffline
	}

	// 4. 读取控制配置
	controlConfig, err := m.loadControlConfig(ctx, deviceInfo.DeviceCategory, deviceInfo.DeviceModel)
	if err != nil {
		return nil, err
	}

	// 5. 构建命令
	command := &devices.AIBoxControlCommand{
		CommandType: devices.AIBoxCommandGetCapabilities,
		DeviceID:    deviceID,
		Model:       devices.AIBoxModel(deviceInfo.DeviceModel),
		Payload:     nil,
		Timestamp:   time.Now().UnixMilli(),
	}

	// 6. 发送命令并等待响应
	response, err := m.sendCommandAndWait(ctx, deviceInfo, controlConfig, command)
	if err != nil {
		return nil, err
	}

	// 7. 检查响应状态
	if response.Status != "success" {
		return nil, fmt.Errorf("获取算法能力失败: %s", response.Error)
	}

	// 8. 解析能力信息
	capabilities, ok := response.Result.(*devices.AIBoxCapabilities)
	if !ok {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	// 9. 存储到缓存（失败不影响返回，仅记录日志）
	if err := m.storeCapabilitiesToCache(ctx, deviceID, capabilities); err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "get_aibox_capabilities"),
			logx.Field("device_id", deviceID),
			logx.Field("error", err.Error()),
		).Error("存储算法能力到缓存失败")
	} else {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "get_aibox_capabilities"),
			logx.Field("device_id", deviceID),
			logx.Field("cache_ttl", m.config.CapabilitiesCacheTTL),
		).Info("成功缓存算法能力")
	}

	return capabilities, nil
}

// ControlAIBoxTask 控制AI Box算法任务（启动/停止）
func (m *controllerManager) ControlAIBoxTask(ctx context.Context, deviceID, taskID string, controlCommand int) error {
	// 1. 获取设备信息
	deviceInfo, err := m.getDeviceInfo(deviceID)
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "control_aibox_task"),
			logx.Field("device_id", deviceID),
			logx.Field("error", err.Error()),
		).Error("获取设备信息失败")
		return core.ErrDeviceNotFound
	}

	// 2. 检查设备在线状态
	if !deviceInfo.IsOnline {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "control_aibox_task"),
			logx.Field("device_id", deviceID),
		).Error("设备离线")
		return core.ErrDeviceOffline
	}

	// 3. 读取控制配置
	controlConfig, err := m.loadControlConfig(ctx, deviceInfo.DeviceCategory, deviceInfo.DeviceModel)
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "control_aibox_task"),
			logx.Field("device_id", deviceID),
			logx.Field("category", deviceInfo.DeviceCategory),
			logx.Field("model", deviceInfo.DeviceModel),
			logx.Field("error", err.Error()),
		).Error("读取控制配置失败")
		return err
	}

	// 4. 构建命令
	command := &devices.AIBoxControlCommand{
		CommandType: devices.AIBoxCommandControlTask,
		DeviceID:    deviceID,
		Model:       devices.AIBoxModel(deviceInfo.DeviceModel),
		Payload: map[string]interface{}{
			"task_id":         taskID,
			"control_command": controlCommand,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	// 5. 发送命令并等待响应
	response, err := m.sendCommandAndWait(ctx, deviceInfo, controlConfig, command)
	if err != nil {
		return err
	}

	// 6. 检查响应状态
	if response.Status != "success" {
		return fmt.Errorf("控制任务失败: %s", response.Error)
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_controller"),
		logx.Field("operation", "control_aibox_task"),
		logx.Field("device_id", deviceID),
		logx.Field("task_id", taskID),
		logx.Field("control_command", controlCommand),
	).Info("控制AI Box任务成功")

	return nil
}
