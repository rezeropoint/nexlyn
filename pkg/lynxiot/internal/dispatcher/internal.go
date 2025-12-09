package dispatcher

import (
	"context"
	"encoding/json"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/zeromicro/go-zero/core/logx"
)

// dispatchLog 日志分发
func (m *dispatcherManager) dispatchLog(ctx context.Context, data *map[string]core.TypedValue, taskInfo core.TaskInfo) error {
	jsonData, err := json.MarshalIndent(*data, "", "  ")
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("configType", taskInfo.ConfigType),
			logx.Field("task_id", taskInfo.TaskId),
			logx.Field("module", "dispatcher"),
			logx.Field("operation", "dispatch_log"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("JSON格式化失败")
		return err
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("configType", taskInfo.ConfigType),
		logx.Field("task_id", taskInfo.TaskId),
		logx.Field("timestamp", taskInfo.Timestamp),
		logx.Field("module", "dispatcher"),
		logx.Field("operation", "dispatch_log"),
		logx.Field("status", "data"),
		logx.Field("json_data", string(jsonData)),
	).Info("调度器日志数据")

	return nil
}
