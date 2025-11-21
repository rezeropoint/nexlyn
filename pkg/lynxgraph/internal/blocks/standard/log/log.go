package log

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/zeromicro/go-zero/core/logx"
)

const LogVersion = "v1"

// Config 日志打印积木的配置
type Config struct {
	Level         string `json:"level" check:"must"` // 日志级别: info/error (logx没有warn方法)
	Message       string `json:"message"`            // 固定日志消息（可选）
	ContextKey    string `json:"contextKey"`         // 从图上下文读取的key（可选）
	PrintInfoAtom bool   `json:"printInfoAtom"`      // 是否打印信息原子payload（可选）
}

// LogBlock 实现日志打印逻辑块
type LogBlock struct {
	core.BaseLogicBlock
}

// NewLogBlock 创建一个新的 LogBlock 实例
func NewLogBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &LogBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeLog,
			RawConfig: config,
		},
	}, nil
}

func (b *LogBlock) GetID() string                { return b.ID }
func (b *LogBlock) GetType() core.LogicBlockType { return b.Type }
func (b *LogBlock) GetConfigure() map[string]any { return b.RawConfig }

func (b *LogBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 验证日志级别
	if cfg.Level != "info" && cfg.Level != "error" {
		return fmt.Errorf("不支持的日志级别: %s (仅支持 info/error)", cfg.Level)
	}

	b.TypedConfig = &cfg
	return nil
}

// Execute 执行日志打印逻辑
func (b *LogBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	config, ok := b.TypedConfig.(*Config)
	if !ok {
		return false, core.ErrInvalidConfig
	}

	// 构建日志字段
	fields := []logx.LogField{
		logx.Field("module", "lynxgraph_block_log"),
		logx.Field("block_id", b.ID),
		logx.Field("tenant_id", execCtx.GetTenantId()),
		logx.Field("graph_name", execCtx.GetGraph().GetName()),
		logx.Field("node_id", execCtx.GetNode().GetID()),
	}

	// 日志消息
	logMessage := config.Message
	if logMessage == "" {
		logMessage = "LynxGraph 逻辑图执行日志"
	}

	// 从图上下文读取
	if config.ContextKey != "" {
		graphKey := execCtx.GetGraphKey()
		graphContext, err := datastore.GetGraphContext(ctx, execCtx.GetTenantId(), graphKey, config.ContextKey)
		if err == nil && graphContext != nil {
			payload := graphContext.GetPayload()
			fields = append(fields, logx.Field("context_"+config.ContextKey, payload))
		} else {
			fields = append(fields, logx.Field("context_"+config.ContextKey, fmt.Sprintf("<读取失败: %v>", err)))
		}
	}

	// 从信息原子读取（直接输出整个payload的JSON）
	if config.PrintInfoAtom {
		payload := execCtx.GetInfoAtom().GetPayload()
		fields = append(fields, logx.Field("infoAtom_payload", payload))
	}

	// 根据级别打印日志
	logger := logx.WithContext(ctx).WithFields(fields...)
	switch config.Level {
	case "info":
		logger.Info(logMessage)
	case "error":
		logger.Error(logMessage)
	default:
		return false, fmt.Errorf("不支持的日志级别: %s", config.Level)
	}

	// 成功执行，继续后续节点
	return true, nil
}

// GetLogSpec 返回 LogBlock 的规格
func GetLogSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"Log",
		LogVersion,
		"打印日志消息，支持固定消息或从信息原子/上下文读取内容，可同时配置多个来源",
		[]string{"debug", "output"},
		map[string]any{ // 配置模式
			"level": map[string]any{
				"type":        "string",
				"title":       "日志级别",
				"description": "选择日志输出级别（info-普通信息，error-错误告警）",
				"placeholder": "请选择日志级别",
				"check":       "must",
				"enum":        []string{"info", "error"},
			},
			"message": map[string]any{
				"type":        "string",
				"title":       "固定消息",
				"description": "静态日志内容，适合固定格式的日志输出（如不填写，需至少配置contextKey或printInfoAtom之一）",
				"placeholder": "例如：设备状态更新通知、传感器数据异常告警",
				"check":       "",
			},
			"contextKey": map[string]any{
				"type":        "string",
				"title":       "上下文键名",
				"description": "从图执行上下文中读取动态内容的字段名，可获取流程中的变量值",
				"placeholder": "例如：device_name、temperature_value、alarm_type",
				"check":       "",
			},
			"printInfoAtom": map[string]any{
				"type":        "boolean",
				"title":       "打印信息原子",
				"description": "是否输出完整的信息原子payload（JSON格式），用于调试或记录完整事件数据",
				"check":       "",
			},
		},
		[]core.ServiceType{}, // 不依赖任何外部服务
	)
}
