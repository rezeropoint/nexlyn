package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core/devices"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/zeromicro/go-zero/core/logx"
)

// storeCommandToRedis 存储命令到Redis（用于监控和调试）
func (m *controllerManager) storeCommandToRedis(ctx context.Context, deviceID string, command *devices.AIBoxControlCommand, ttl int) error {
	key := core.BuildControlDeviceRedisKey(deviceID)

	// 序列化命令
	data, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("序列化命令失败: %w", err)
	}

	// 存储到Redis，设置TTL
	if err := m.redisClient.SetexCtx(ctx, key, string(data), ttl); err != nil {
		return fmt.Errorf("存储命令到Redis失败: %w", err)
	}

	return nil
}

// getCommandFromRedis 从Redis获取命令
func (m *controllerManager) getCommandFromRedis(ctx context.Context, deviceID string) (*devices.AIBoxControlCommand, error) {
	key := core.BuildControlDeviceRedisKey(deviceID)

	// 从Redis读取
	data, err := m.redisClient.GetCtx(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("从Redis读取命令失败: %w", err)
	}

	// 反序列化
	var command devices.AIBoxControlCommand
	if err := json.Unmarshal([]byte(data), &command); err != nil {
		return nil, fmt.Errorf("反序列化命令失败: %w", err)
	}

	return &command, nil
}

// deleteCommandFromRedis 从Redis删除命令
func (m *controllerManager) deleteCommandFromRedis(ctx context.Context, deviceID string) error {
	key := core.BuildControlDeviceRedisKey(deviceID)
	_, err := m.redisClient.DelCtx(ctx, key)
	return err
}

// ===============================
// 分布式锁管理
// ===============================

// tryAcquireControlLock 尝试获取设备控制分布式锁
// 使用Redis SET NX EX命令，防止多Pod并发冲突
// 参数：
//   - deviceID: 设备ID
//   - ttl: 锁的过期时间（秒），建议60秒（比命令超时30秒长，防止死锁）
//
// 返回：是否成功获取锁
func (m *controllerManager) tryAcquireControlLock(ctx context.Context, deviceID string, ttl int) (bool, error) {
	lockKey := core.BuildControlLockKey(deviceID)

	// SET key value NX EX ttl
	// NX: 只在键不存在时设置
	// EX: 设置过期时间（秒）
	ok, err := m.redisClient.SetnxExCtx(ctx, lockKey, "1", ttl)
	if err != nil {
		return false, fmt.Errorf("获取设备控制分布式锁失败: %w", err)
	}
	return ok, nil
}

// releaseControlLock 释放设备控制分布式锁
func (m *controllerManager) releaseControlLock(ctx context.Context, deviceID string) error {
	lockKey := core.BuildControlLockKey(deviceID)
	_, err := m.redisClient.DelCtx(ctx, lockKey)
	if err != nil {
		return fmt.Errorf("释放设备控制分布式锁失败: %w", err)
	}
	return nil
}

// waitForResponse 等待命令响应
// 使用设备ID作为匹配key，单设备同时只允许一个命令执行
// expectedEvent: 期望的Event类型（如 /alg_ability_fetch），用于验证响应匹配
//
// 并发安全：使用 Redis分布式锁（跨Pod协调）+ sync.Map（单Pod内快速响应匹配）
func (m *controllerManager) waitForResponse(ctx context.Context, deviceID string, expectedEvent string, timeout time.Duration) (*devices.AIBoxControlResponse, error) {
	// 步骤1：获取分布式锁（跨Pod协调，防止多Pod并发冲突）
	locked, err := m.tryAcquireControlLock(ctx, deviceID, 60) // 60秒TTL，比命令超时30秒长
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, fmt.Errorf("设备正在处理其他命令，请稍后重试")
	}

	// 步骤2：确保释放锁（无论成功、超时还是取消）
	defer func() {
		// 使用Background context确保锁能释放（即使主context被取消）
		releaseCtx := context.Background()
		if err := m.releaseControlLock(releaseCtx, deviceID); err != nil {
			logx.Error("释放设备控制锁失败: ", err)
		}
	}()

	// 步骤3：创建响应channel
	responseCh := make(chan *devices.AIBoxControlResponse, 1)

	// 步骤4：注册到pendingCommands（用于响应匹配，单Pod内快速查找）
	pending := &pendingCommand{
		responseCh:    responseCh,
		expectedEvent: expectedEvent,
	}
	m.pendingCommands.Store(deviceID, pending)

	// 步骤5：确保清理pendingCommands
	defer func() {
		m.pendingCommands.Delete(deviceID)
		close(responseCh)
	}()

	// 步骤6：等待响应或超时
	select {
	case response := <-responseCh:
		// 收到响应
		return response, nil
	case <-time.After(timeout):
		// 超时
		return nil, fmt.Errorf("命令执行超时")
	case <-ctx.Done():
		// Context取消
		return nil, ctx.Err()
	}
}

// notifyResponse 通知响应到达（由handleResponse调用）
// 验证Event类型是否匹配，避免并发命令时响应混淆
func (m *controllerManager) notifyResponse(deviceID string, event string, response *devices.AIBoxControlResponse) bool {
	if val, ok := m.pendingCommands.Load(deviceID); ok {
		pending := val.(*pendingCommand)

		// 验证Event是否匹配期望值
		if pending.expectedEvent != event {
			// Event不匹配，可能是旧的或错误的响应
			return false
		}

		// Event匹配，发送响应
		select {
		case pending.responseCh <- response:
			// 成功发送响应
			return true
		default:
			// channel已满或已关闭，忽略
			return false
		}
	}
	return false
}

// loadControlConfig 从Etcd加载控制配置
// 注意：由于配置可能使用不同的CommandSuffix，需要通过前缀查询来查找
func (m *controllerManager) loadControlConfig(ctx context.Context, category, model string) (*core.DeviceControlConfig, error) {
	// 构建Etcd前缀
	prefix := core.BuildControlConfigPrefix(category, model)

	// 从ConfigManager获取前缀下的所有键
	keys := m.configStore.GetAllKeys(prefix)
	if len(keys) == 0 {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "load_control_config"),
			logx.Field("prefix", prefix),
			logx.Field("category", category),
			logx.Field("model", model),
		).Error("未找到控制配置：Etcd中不存在匹配的键")
		return nil, fmt.Errorf("设备未配置控制功能")
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_controller"),
		logx.Field("operation", "load_control_config"),
		logx.Field("prefix", prefix),
		logx.Field("keys_found", len(keys)),
		logx.Field("first_key", keys[0]),
	).Info("找到控制配置")

	// 读取第一个配置（通常每个型号只有一个控制配置）
	var config core.DeviceControlConfig
	for _, key := range keys {
		if m.configStore.GetConfig(key, &config) {
			// 成功读取配置
			logx.WithContext(ctx).WithFields(
				logx.Field("service", m.config.ServiceName),
				logx.Field("pod", m.config.PodName),
				logx.Field("module", "iot_controller"),
				logx.Field("operation", "load_control_config"),
				logx.Field("key", key),
				logx.Field("command_suffix", config.CommandSuffix),
				logx.Field("response_suffix", config.ResponseSuffix),
			).Info("成功加载控制配置")
			return &config, nil
		}
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_controller"),
		logx.Field("operation", "load_control_config"),
		logx.Field("prefix", prefix),
	).Error("控制配置读取失败：所有键都无法解析")
	return nil, fmt.Errorf("设备控制配置读取失败")
}

// sendCommandAndWait 发送命令并等待响应（核心逻辑）
func (m *controllerManager) sendCommandAndWait(ctx context.Context, deviceInfo *core.DeviceBinding, controlConfig *core.DeviceControlConfig, command *devices.AIBoxControlCommand) (*devices.AIBoxControlResponse, error) {
	// 1. 检查MQTT连接
	if !m.IsConnected() {
		return nil, fmt.Errorf("MQTT未连接")
	}

	// 2. 型号转换
	payload, err := m.converter.ConvertCommand(deviceInfo.DeviceModel, command)
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "send_command"),
			logx.Field("device_id", deviceInfo.DeviceID),
			logx.Field("command_type", command.CommandType),
			logx.Field("error", err.Error()),
		).Error("型号转换失败")
		return nil, err
	}

	// 3. 构建主题
	commandTopic := core.BuildControlTopic(deviceInfo.DeviceCategory, deviceInfo.DeviceModel, deviceInfo.DeviceID, controlConfig.CommandSuffix)
	responseTopic := core.BuildControlTopic(deviceInfo.DeviceCategory, deviceInfo.DeviceModel, deviceInfo.DeviceID, controlConfig.ResponseSuffix)

	// 4. 订阅响应主题（如果尚未订阅）
	if token := m.client.Subscribe(responseTopic, 0, m.handleResponse); token.Wait() && token.Error() != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "subscribe_response"),
			logx.Field("topic", responseTopic),
			logx.Field("error", token.Error().Error()),
		).Error("订阅响应主题失败")
		return nil, token.Error()
	}

	// 5. 存储命令到Redis
	if err := m.storeCommandToRedis(ctx, deviceInfo.DeviceID, command, 60); err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "store_command"),
			logx.Field("device_id", deviceInfo.DeviceID),
			logx.Field("error", err.Error()),
		).Error("存储命令到Redis失败")
	}

	// 6. 发送MQTT消息
	if token := m.client.Publish(commandTopic, 0, false, payload); token.Wait() && token.Error() != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "publish_command"),
			logx.Field("topic", commandTopic),
			logx.Field("device_id", deviceInfo.DeviceID),
			logx.Field("command_type", command.CommandType),
			logx.Field("error", token.Error().Error()),
		).Error("发送MQTT命令失败")
		return nil, token.Error()
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_controller"),
		logx.Field("operation", "publish_command"),
		logx.Field("status", "success"),
		logx.Field("topic", commandTopic),
		logx.Field("device_id", deviceInfo.DeviceID),
		logx.Field("command_type", command.CommandType),
	).Info("成功发送控制命令")

	// 7. 确定期望的Event类型（用于响应验证）
	var expectedEvent string
	switch command.CommandType {
	case devices.AIBoxCommandGetCapabilities:
		expectedEvent = devices.AIBoxEventAbilityFetch
	case devices.AIBoxCommandListTasks:
		expectedEvent = devices.AIBoxEventTaskFetch
	case devices.AIBoxCommandControlTask:
		expectedEvent = devices.AIBoxEventTaskControl
	default:
		expectedEvent = "" // 未知命令类型
	}

	// 8. 等待响应（30秒超时，验证Event匹配）
	response, err := m.waitForResponse(ctx, deviceInfo.DeviceID, expectedEvent, 30*time.Second)
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "wait_response"),
			logx.Field("device_id", deviceInfo.DeviceID),
			logx.Field("expected_event", expectedEvent),
			logx.Field("error", err.Error()),
		).Error("等待命令响应失败")
		return nil, err
	}

	// 9. 清理Redis命令记录
	m.deleteCommandFromRedis(ctx, deviceInfo.DeviceID)

	return response, nil
}

// handleResponse 处理MQTT响应消息
func (m *controllerManager) handleResponse(client mqtt.Client, msg mqtt.Message) {
	ctx := context.Background()

	// 1. 先提取设备ID和事件类型（用于响应匹配）
	var rawResp struct {
		BoardId string `json:"BoardId"`
		Event   string `json:"Event"`
	}
	if err := json.Unmarshal(msg.Payload(), &rawResp); err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "handle_response"),
			logx.Field("topic", msg.Topic()),
			logx.Field("error", err.Error()),
		).Error("解析响应消息失败")
		return
	}

	deviceID := rawResp.BoardId
	if deviceID == "" {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "handle_response"),
			logx.Field("topic", msg.Topic()),
		).Error("响应中缺少 BoardId 字段")
		return
	}

	// 2. 使用转换器转换为统一格式
	response, err := m.converter.ConvertResponse(deviceID, msg.Payload())
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "handle_response"),
			logx.Field("device_id", deviceID),
			logx.Field("event", rawResp.Event),
			logx.Field("error", err.Error()),
		).Error("转换响应失败")
		return
	}

	// 3. 记录日志（优化：大响应只记录元信息）
	logFields := []logx.LogField{
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "iot_controller"),
		logx.Field("operation", "handle_response"),
		logx.Field("device_id", deviceID),
		logx.Field("event", rawResp.Event),
		logx.Field("status", response.Status),
		logx.Field("response_size_bytes", len(msg.Payload())),
	}

	// 如果是算法能力查询，记录算法数量而非完整响应
	if rawResp.Event == devices.AIBoxEventAbilityFetch {
		if caps, ok := response.Result.(*devices.AIBoxCapabilities); ok {
			logFields = append(logFields, logx.Field("abilities_count", len(caps.Abilities)))
		}
	}

	logx.WithContext(ctx).WithFields(logFields...).Info("收到控制命令响应")

	// 4. 通知等待的goroutine（验证Event是否匹配）
	matched := m.notifyResponse(deviceID, rawResp.Event, response)
	if !matched {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("module", "iot_controller"),
			logx.Field("operation", "handle_response"),
			logx.Field("device_id", deviceID),
			logx.Field("event", rawResp.Event),
		).Error("响应Event不匹配或无待处理命令，可能是并发命令导致的响应混淆")
	}
}

// getCapabilitiesFromCache 从Redis获取算法能力缓存
func (m *controllerManager) getCapabilitiesFromCache(ctx context.Context, deviceID string) (*devices.AIBoxCapabilities, error) {
	// 检查是否启用缓存
	if m.config.CapabilitiesCacheTTL <= 0 {
		return nil, fmt.Errorf("缓存已禁用")
	}

	key := core.BuildCapabilitiesCacheKey(deviceID)

	// 从Redis读取
	data, err := m.redisClient.GetCtx(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("从Redis读取缓存失败: %w", err)
	}

	// 反序列化
	var capabilities devices.AIBoxCapabilities
	if err := json.Unmarshal([]byte(data), &capabilities); err != nil {
		return nil, fmt.Errorf("反序列化缓存数据失败: %w", err)
	}

	return &capabilities, nil
}

// storeCapabilitiesToCache 存储算法能力到Redis缓存
func (m *controllerManager) storeCapabilitiesToCache(ctx context.Context, deviceID string, capabilities *devices.AIBoxCapabilities) error {
	// 检查是否启用缓存
	if m.config.CapabilitiesCacheTTL <= 0 {
		return nil // 缓存已禁用，跳过
	}

	key := core.BuildCapabilitiesCacheKey(deviceID)

	// 序列化
	data, err := json.Marshal(capabilities)
	if err != nil {
		return fmt.Errorf("序列化算法能力失败: %w", err)
	}

	// 存储到Redis，设置TTL
	if err := m.redisClient.SetexCtx(ctx, key, string(data), m.config.CapabilitiesCacheTTL); err != nil {
		return fmt.Errorf("存储到Redis失败: %w", err)
	}

	return nil
}
