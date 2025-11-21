package dispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/service/lynxengine/pb"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// dispatchToLynxGraph 分发到 LynxGraph 逻辑引擎
func (m *dispatcherManager) dispatchToLynxGraph(ctx context.Context, config core.DispatchConfig, data *map[string]core.TypedValue, taskInfo core.TaskInfo) error {
	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("configType", taskInfo.ConfigType),
		logx.Field("task_id", taskInfo.TaskId),
		logx.Field("module", "dispatcher"),
		logx.Field("operation", "dispatch_lynxgraph"),
		logx.Field("info_atom_type_id", config.InfoAtomTypeID),
		logx.Field("status", "started"),
	).Debug("开始执行 LynxGraph 分发")

	// 检查 gRPC 地址配置
	if m.config.LynxGraphGRPCURL == "" {
		err := fmt.Errorf("LynxGraph gRPC 地址未配置")
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("configType", taskInfo.ConfigType),
			logx.Field("task_id", taskInfo.TaskId),
			logx.Field("module", "dispatcher"),
			logx.Field("operation", "dispatch_lynxgraph"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("LynxGraph 分发失败")
		return err
	}

	// 直接转换所有字段为JSON（不需要字段映射，fieldMappings已经定义了要发送的字段）
	mappedData := make(map[string]any)
	for key, value := range *data {
		mappedData[key] = convertTypedValueToAny(value)
	}

	// 将数据转换为 JSON
	rawData, err := json.Marshal(mappedData)
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("configType", taskInfo.ConfigType),
			logx.Field("task_id", taskInfo.TaskId),
			logx.Field("module", "dispatcher"),
			logx.Field("operation", "dispatch_lynxgraph"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("JSON 序列化失败")
		return fmt.Errorf("JSON 序列化失败: %w", err)
	}

	// 创建 gRPC 连接
	conn, err := grpc.Dial(
		m.config.LynxGraphGRPCURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("configType", taskInfo.ConfigType),
			logx.Field("task_id", taskInfo.TaskId),
			logx.Field("module", "dispatcher"),
			logx.Field("operation", "dispatch_lynxgraph"),
			logx.Field("grpc_url", m.config.LynxGraphGRPCURL),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("gRPC 连接失败")
		return fmt.Errorf("gRPC 连接失败: %w", err)
	}
	defer conn.Close()

	// 创建 gRPC 客户端
	client := pb.NewLynxEngineClient(conn)

	// 构造请求
	request := &pb.InfoAtomRequest{
		TenantId:       taskInfo.TenantID,
		InfoAtomTypeId: config.InfoAtomTypeID,
		Source:         taskInfo.DeviceID,
		RawData:        rawData,
		Tags:           extractTags(config),
		Timestamp:      time.Now().UnixMilli(),
	}

	// 调用 gRPC 接口
	response, err := client.ReceiveInfoAtom(ctx, request)
	if err != nil {
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("configType", taskInfo.ConfigType),
			logx.Field("task_id", taskInfo.TaskId),
			logx.Field("module", "dispatcher"),
			logx.Field("operation", "dispatch_lynxgraph"),
			logx.Field("info_atom_type_id", config.InfoAtomTypeID),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("gRPC 调用失败")
		return fmt.Errorf("gRPC 调用失败: %w", err)
	}

	// 检查响应
	if !response.Success {
		err := fmt.Errorf("LynxGraph 处理失败: %s (错误码: %s)", response.Message, response.ErrorCode)
		logx.WithContext(ctx).WithFields(
			logx.Field("service", m.config.ServiceName),
			logx.Field("pod", m.config.PodName),
			logx.Field("configType", taskInfo.ConfigType),
			logx.Field("task_id", taskInfo.TaskId),
			logx.Field("module", "dispatcher"),
			logx.Field("operation", "dispatch_lynxgraph"),
			logx.Field("info_atom_type_id", config.InfoAtomTypeID),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("LynxGraph 分发失败")
		return err
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("configType", taskInfo.ConfigType),
		logx.Field("task_id", taskInfo.TaskId),
		logx.Field("module", "dispatcher"),
		logx.Field("operation", "dispatch_lynxgraph"),
		logx.Field("info_atom_type_id", config.InfoAtomTypeID),
		logx.Field("info_atom_id", response.InfoAtomId),
		logx.Field("status", "success"),
	).Debug("LynxGraph 分发成功")

	return nil
}

// convertTypedValueToAny 将 TypedValue 转换为 any 类型
func convertTypedValueToAny(value core.TypedValue) any {
	return value.Value
}

// extractTags 从配置中提取标签
func extractTags(config core.DispatchConfig) []string {
	// 可以从 ExtraParams 中提取标签，或者使用默认标签
	if tags, ok := config.ExtraParams["tags"].([]string); ok {
		return tags
	}
	// 默认标签
	return []string{"iot"}
}
