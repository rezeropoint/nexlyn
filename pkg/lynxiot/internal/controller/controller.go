package controller

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core/devices"

	"github.com/rezeropoint/etcdtrigger/v2/engine"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// Manager Controller管理器接口
type Manager interface {
	// Start 启动Controller管理器（连接MQTT Broker）
	Start(ctx context.Context) error

	// Stop 停止Controller管理器
	Stop()

	// IsConnected 检查MQTT连接状态
	IsConnected() bool

	// AI Box 算法任务管理（统一接口）
	CreateAIBoxTask(ctx context.Context, deviceID string, task *devices.AIBoxAlgorithmTask) (string, error) // 创建AI Box算法任务
	UpdateAIBoxTask(ctx context.Context, deviceID, taskID string, task *devices.AIBoxAlgorithmTask) error   // 更新AI Box算法任务
	DeleteAIBoxTask(ctx context.Context, deviceID, taskID string) error                                     // 删除AI Box算法任务
	ListAIBoxTasks(ctx context.Context, deviceID string) ([]devices.AIBoxAlgorithmTask, error)              // 列出AI Box算法任务

	// AI Box 算法能力查询
	GetAIBoxCapabilities(ctx context.Context, deviceID string) (*devices.AIBoxCapabilities, error) // 获取AI Box算法能力

	// AI Box 算法任务控制
	ControlAIBoxTask(ctx context.Context, deviceID, taskID string, controlCommand int) error // 控制AI Box算法任务（启动/停止）
}

// NewManager 创建Controller管理器实例
func NewManager(config Config, redisClient *redis.Redis, configStore engine.Engine, getDeviceInfo core.GetDeviceInfoFunc) (Manager, error) {
	return newControllerManager(config, redisClient, configStore, getDeviceInfo)
}
