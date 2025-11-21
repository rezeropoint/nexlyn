package dispatcher

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/zeromicro/go-zero/core/logx"
)

// dispatchSkylarkFlows 分发到 Skylark 流程
func (m *dispatcherManager) dispatchSkylarkFlows(ctx context.Context, config core.DispatchConfig, data *map[string]core.TypedValue, taskInfo core.TaskInfo) error {
	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("configType", taskInfo.ConfigType),
		logx.Field("task_id", taskInfo.TaskId),
		logx.Field("module", "dispatcher"),
		logx.Field("operation", "dispatch_skylark_flows"),
		logx.Field("platform_id", config.PlatformID),
		logx.Field("flow_id", config.FlowID),
		logx.Field("status", "started"),
	).Debug("开始执行 Skylark 流程调度")

	// 获取平台配置
	platformConfig, found := m.getPlatformConfig(config.PlatformID)
	if !found {
		err := fmt.Errorf("未找到平台配置: %s", config.PlatformID)
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("configType", taskInfo.ConfigType),
			logx.Field("task_id", taskInfo.TaskId),
			logx.Field("module", "dispatcher"),
			logx.Field("operation", "dispatch_skylark_flows"),
			logx.Field("platform_id", config.PlatformID),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("Skylark 流程调度失败")
		return err
	}

	// 验证平台类型
	if platformConfig.Type != core.PlatformTypeSkylark {
		err := fmt.Errorf("平台类型不匹配，期望 skylark，实际 %s", platformConfig.Type)
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("configType", taskInfo.ConfigType),
			logx.Field("task_id", taskInfo.TaskId),
			logx.Field("module", "dispatcher"),
			logx.Field("operation", "dispatch_skylark_flows"),
			logx.Field("platform_id", config.PlatformID),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("Skylark 流程调度失败")
		return err
	}

	// 转换数据格式（从 core.TypedValue 到 skylark core.TypedValue）
	converted := convertToSkylarkTypedValue(data)

	// 调用 Skylark 引擎创建流程
	err := m.skylarkEngine.CreateFlow(
		ctx,
		platformConfig.SkylarkDomain,
		config.FlowID,
		platformConfig.SkylarkUserID,
		platformConfig.SkylarkAuthHeader,
		converted,
	)
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("configType", taskInfo.ConfigType),
			logx.Field("task_id", taskInfo.TaskId),
			logx.Field("module", "dispatcher"),
			logx.Field("operation", "dispatch_skylark_flows"),
			logx.Field("platform_id", config.PlatformID),
			logx.Field("flow_id", config.FlowID),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("Skylark 流程调度失败")
		return err
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("configType", taskInfo.ConfigType),
		logx.Field("task_id", taskInfo.TaskId),
		logx.Field("module", "dispatcher"),
		logx.Field("operation", "dispatch_skylark_flows"),
		logx.Field("platform_id", config.PlatformID),
		logx.Field("flow_id", config.FlowID),
		logx.Field("status", "success"),
	).Debug("Skylark 流程调度成功")

	return nil
}

// dispatchSkylarkForms 分发到 Skylark 表单
func (m *dispatcherManager) dispatchSkylarkForms(ctx context.Context, config core.DispatchConfig, data *map[string]core.TypedValue, taskInfo core.TaskInfo) error {
	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("configType", taskInfo.ConfigType),
		logx.Field("task_id", taskInfo.TaskId),
		logx.Field("module", "dispatcher"),
		logx.Field("operation", "dispatch_skylark_forms"),
		logx.Field("platform_id", config.PlatformID),
		logx.Field("form_id", config.FormID),
		logx.Field("status", "started"),
	).Debug("开始执行 Skylark 表单调度")

	// 获取平台配置
	platformConfig, found := m.getPlatformConfig(config.PlatformID)
	if !found {
		err := fmt.Errorf("未找到平台配置: %s", config.PlatformID)
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("configType", taskInfo.ConfigType),
			logx.Field("task_id", taskInfo.TaskId),
			logx.Field("module", "dispatcher"),
			logx.Field("operation", "dispatch_skylark_forms"),
			logx.Field("platform_id", config.PlatformID),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("Skylark 表单调度失败")
		return err
	}

	// 验证平台类型
	if platformConfig.Type != core.PlatformTypeSkylark {
		err := fmt.Errorf("平台类型不匹配，期望 skylark，实际 %s", platformConfig.Type)
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("configType", taskInfo.ConfigType),
			logx.Field("task_id", taskInfo.TaskId),
			logx.Field("module", "dispatcher"),
			logx.Field("operation", "dispatch_skylark_forms"),
			logx.Field("platform_id", config.PlatformID),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("Skylark 表单调度失败")
		return err
	}

	// 转换数据格式（从 core.TypedValue 到 skylark core.TypedValue）
	converted := convertToSkylarkTypedValue(data)

	// 调用 Skylark 引擎创建表单行
	err := m.skylarkEngine.CreateFormRow(
		ctx,
		platformConfig.SkylarkDomain,
		config.FormID,
		platformConfig.SkylarkUserID,
		platformConfig.SkylarkAuthHeader,
		converted,
	)
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("configType", taskInfo.ConfigType),
			logx.Field("task_id", taskInfo.TaskId),
			logx.Field("module", "dispatcher"),
			logx.Field("operation", "dispatch_skylark_forms"),
			logx.Field("platform_id", config.PlatformID),
			logx.Field("form_id", config.FormID),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("Skylark 表单调度失败")
		return err
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("configType", taskInfo.ConfigType),
		logx.Field("task_id", taskInfo.TaskId),
		logx.Field("module", "dispatcher"),
		logx.Field("operation", "dispatch_skylark_forms"),
		logx.Field("platform_id", config.PlatformID),
		logx.Field("form_id", config.FormID),
		logx.Field("status", "success"),
	).Debug("Skylark 表单调度成功")

	return nil
}
