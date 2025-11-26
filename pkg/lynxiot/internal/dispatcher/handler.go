package dispatcher

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/rezeropoint/go-skylark/v2/engine"
	"go.opentelemetry.io/otel/trace"
)

type dispatcherManager struct {
	config            Config
	getPlatformConfig core.GetPlatformConfigFunc
	skylarkEngine     engine.SkylarkEngine
	submitAsyncFunc   core.SubmitAsyncFunc
}

func newManager(config Config, getPlatformConfig core.GetPlatformConfigFunc, skylarkEngine engine.SkylarkEngine, submitAsync core.SubmitAsyncFunc) *dispatcherManager {
	return &dispatcherManager{
		config:            config,
		getPlatformConfig: getPlatformConfig,
		skylarkEngine:     skylarkEngine,
		submitAsyncFunc:   submitAsync,
	}
}

func (m *dispatcherManager) DispatchInfo(ctx context.Context, configs []core.DispatchConfig, data *map[string]core.TypedValue, taskInfo core.TaskInfo) error {
	if len(configs) == 0 {
		return nil
	}

	// 为异步任务创建独立的 context，避免 HTTP 请求结束后 context 被取消
	// 同时保留 trace 信息用于日志关联
	asyncCtx := detachContext(ctx)

	// 创建任务函数切片
	tasks := make([]func() error, len(configs))
	for i, config := range configs {
		cfg := config // 避免闭包问题
		tasks[i] = func() error {
			return m.dispatch(asyncCtx, cfg, data, taskInfo)
		}
	}

	// 异步提交任务，只返回提交错误，不等待执行结果
	submitErrors := m.submitAsyncFunc(tasks)

	// 检查是否有提交错误
	for _, err := range submitErrors {
		if err != nil {
			return err
		}
	}

	return nil
}

// detachContext 创建一个独立的 context，保留 trace 信息但不受原 context 取消影响
func detachContext(ctx context.Context) context.Context {
	// 提取原 context 中的 span，用于 trace 关联
	span := trace.SpanFromContext(ctx)
	// 创建新的独立 context，不会因为原 context 取消而取消
	return trace.ContextWithSpan(context.Background(), span)
}

// Close 关闭调度器资源
func (m *dispatcherManager) Close() error {
	return nil
}

func (m *dispatcherManager) dispatch(ctx context.Context, config core.DispatchConfig, data *map[string]core.TypedValue, taskInfo core.TaskInfo) error {
	switch config.Type {
	case core.DispatchTypeLog:
		return m.dispatchLog(ctx, data, taskInfo)

	case core.DispatchTypeSkylarkFlows:
		return m.dispatchSkylarkFlows(ctx, config, data, taskInfo)

	case core.DispatchTypeSkylarkForms:
		return m.dispatchSkylarkForms(ctx, config, data, taskInfo)

	case core.DispatchTypeLynxGraph:
		return m.dispatchToLynxGraph(ctx, config, data, taskInfo)

	default:
		return fmt.Errorf("未知的分发类型: %s", config.Type)
	}
}
