package foreach

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/dispatcher"
	"github.com/zeromicro/go-zero/core/logx"
)

const ForEachVersion = "v1"

// Config ForEach 积木配置
// 信息原子驱动模式：通过创建信息原子触发订阅了该类型的子图
type Config struct {
	SourceType              string `json:"sourceType" check:"must"`              // 数据来源: "context" | "atom"
	SourcePath              string `json:"sourcePath" check:"must"`              // 数组路径，如 "employees.data" 或 "items"
	IterationInfoAtomTypeId string `json:"iterationInfoAtomTypeId" check:"must"` // 迭代信息原子类型ID
	ItemKey                 string `json:"itemKey"`                              // Payload 中元素的 key，默认 "item"
	IndexKey                string `json:"indexKey"`                             // Payload 中索引的 key，默认 "index"
	ContinueOnError         bool   `json:"continueOnError"`                      // 单元素失败是否继续
}

// ForEachBlock 遍历数组，为每个元素创建信息原子触发子图
type ForEachBlock struct {
	core.BaseLogicBlock
}

// NewForEachBlock 创建一个新的 ForEachBlock 实例
func NewForEachBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &ForEachBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeForEach,
			RawConfig: config,
		},
	}, nil
}

func (b *ForEachBlock) GetID() string                { return b.ID }
func (b *ForEachBlock) GetType() core.LogicBlockType { return b.Type }
func (b *ForEachBlock) GetConfigure() map[string]any { return b.RawConfig }

func (b *ForEachBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 验证 sourceType
	if cfg.SourceType != "context" && cfg.SourceType != "atom" {
		return fmt.Errorf("sourceType 必须是 context 或 atom")
	}

	// 设置默认值
	if cfg.ItemKey == "" {
		cfg.ItemKey = "item"
	}
	if cfg.IndexKey == "" {
		cfg.IndexKey = "index"
	}

	b.TypedConfig = &cfg
	return nil
}

// Execute 执行 ForEach 逻辑
// 信息原子驱动模式：为每个元素创建信息原子并分发，由订阅了该类型的子图自动处理
func (b *ForEachBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	config, ok := b.TypedConfig.(*Config)
	if !ok {
		return false, core.ErrInvalidConfig
	}

	tenantId := execCtx.GetTenantId()
	parentGraphKey := execCtx.GetGraphKey()
	parentNodeId := b.ID

	// 1. 获取源数组
	sourceArray, err := b.getSourceArray(ctx, config, execCtx, datastore)
	if err != nil {
		return false, fmt.Errorf("[ForEach] 获取源数组失败: %w", err)
	}

	if len(sourceArray) == 0 {
		logx.WithContext(ctx).Infof("[ForEach] 源数组为空，跳过执行")
		return true, nil
	}

	logx.WithContext(ctx).Infof("[ForEach] 开始遍历，数组长度: %d", len(sourceArray))

	// 2. 从 Service 获取 Dispatcher
	dispatcherSvc, err := service.GetByType(core.ServiceTypeDispatcher)
	if err != nil {
		return false, fmt.Errorf("[ForEach] 获取 Dispatcher 服务失败: %w", err)
	}

	dispatcherRegistry, ok := dispatcherSvc.(dispatcher.DispatcherRegistry)
	if !ok {
		return false, fmt.Errorf("[ForEach] Dispatcher 服务类型错误")
	}

	// 3. 从 Service 获取信息原子类型查询函数
	infoAtomQuerySvc, err := service.GetByType(core.ServiceTypeInfoAtomQuery)
	if err != nil {
		return false, fmt.Errorf("[ForEach] 获取信息原子类型查询服务失败: %w", err)
	}

	infoAtomQueryFunc, ok := infoAtomQuerySvc.(core.InfoAtomTypeQueryFunc)
	if !ok {
		return false, fmt.Errorf("[ForEach] 信息原子类型查询服务类型错误")
	}

	// 4. 查询迭代信息原子类型
	infoAtomTypeKey := core.InfoAtomTypeKey{ID: config.IterationInfoAtomTypeId}
	infoAtomType, err := infoAtomQueryFunc(ctx, infoAtomTypeKey)
	if err != nil {
		return false, fmt.Errorf("[ForEach] 查询信息原子类型失败: %w", err)
	}

	if infoAtomType == nil {
		return false, fmt.Errorf("[ForEach] 信息原子类型不存在: %s", config.IterationInfoAtomTypeId)
	}

	// 5. 遍历数组，为每个元素创建信息原子并分发
	var lastErr error
	successCount := 0
	now := time.Now().UnixMilli()

	for index, item := range sourceArray {
		// 构建原始 Payload（ForEach 提供的完整数据）
		rawPayload := map[string]any{
			config.ItemKey:   item,
			config.IndexKey:  index,
			"total":          len(sourceArray),
			"parentGraphKey": parentGraphKey.ID,
			"parentNodeId":   parentNodeId,
		}

		// 根据信息原子类型的 DataFormat 提取字段，确保子图收到符合要求的信息原子
		extractedPayload, err := extractFieldsByDataFormat(rawPayload, infoAtomType.GetDataFormat())
		if err != nil {
			logx.WithContext(ctx).Errorf("[ForEach] 字段提取失败 (index=%d): %v", index, err)
			if !config.ContinueOnError {
				return false, fmt.Errorf("[ForEach] 字段提取失败 (index=%d): %w", index, err)
			}
			lastErr = err
			continue
		}

		// 创建迭代信息原子
		iterInfoAtom := &core.BasicInfoAtom{
			TenantId:  tenantId,
			ID:        uuid.New().String(),
			Type:      infoAtomType,
			Source:    fmt.Sprintf("foreach:%s:%s", parentGraphKey.ID, parentNodeId),
			Timestamp: now,
			Labels:    map[string]string{},
			Payload:   extractedPayload,
		}

		// 分发信息原子（异步，由订阅了该类型的子图处理）
		logx.WithContext(ctx).Debugf("[ForEach] 分发迭代信息原子 (index=%d, atomId=%s)", index, iterInfoAtom.ID)

		if err := dispatcherRegistry.Dispatch(iterInfoAtom); err != nil {
			logx.WithContext(ctx).Errorf("[ForEach] 分发信息原子失败 (index=%d): %v", index, err)
			if !config.ContinueOnError {
				return false, fmt.Errorf("[ForEach] 分发信息原子失败 (index=%d): %w", index, err)
			}
			lastErr = err
			continue
		}

		successCount++
	}

	logx.WithContext(ctx).Infof("[ForEach] 分发完成，成功: %d/%d", successCount, len(sourceArray))

	// 如果所有分发都失败了，返回错误
	if successCount == 0 && lastErr != nil {
		return false, lastErr
	}

	return true, nil
}

// getSourceArray 从 context 或 atom 获取源数组
func (b *ForEachBlock) getSourceArray(ctx context.Context, config *Config, execCtx core.ExecutionContext, datastore core.Store) ([]any, error) {
	var rawValue any

	if config.SourceType == "atom" {
		// 从 InfoAtom payload 获取
		payload := execCtx.GetInfoAtom().GetPayload()
		rawValue = getNestedValue(payload, config.SourcePath)
	} else {
		// 从 GraphContext 获取
		// sourcePath 格式: "contextKey.field1.field2" 或 "contextKey"
		parts := strings.SplitN(config.SourcePath, ".", 2)
		contextKey := parts[0]

		graphContext, err := datastore.GetGraphContext(
			ctx,
			execCtx.GetTenantId(),
			execCtx.GetGraphKey(),
			contextKey,
		)
		if err != nil {
			return nil, fmt.Errorf("获取 GraphContext %s 失败: %w", contextKey, err)
		}

		payload := graphContext.GetPayload()
		if len(parts) > 1 {
			rawValue = getNestedValue(payload, parts[1])
		} else {
			// 整个 payload 作为数组（如果它是数组的话）
			rawValue = payload
		}
	}

	if rawValue == nil {
		return nil, nil
	}

	// 转换为 []any
	switch v := rawValue.(type) {
	case []any:
		return v, nil
	case []map[string]any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, nil
	case []string:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, nil
	default:
		return nil, fmt.Errorf("sourcePath %s 的值不是数组类型: %T", config.SourcePath, rawValue)
	}
}

// getNestedValue 从 map 中获取嵌套值
func getNestedValue(data map[string]any, path string) any {
	if data == nil || path == "" {
		return data
	}

	parts := strings.Split(path, ".")
	current := any(data)

	for _, part := range parts {
		if m, ok := current.(map[string]any); ok {
			current = m[part]
		} else if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}

	return current
}

// extractFieldsByDataFormat 根据 DataFormat 从原始数据中提取字段
// 这是 ForEach 积木内部使用的辅助函数，确保子图收到的信息原子符合其 DataFormat 定义
func extractFieldsByDataFormat(rawPayload map[string]any, dataFormat core.DataFormat) (map[string]any, error) {
	// 如果 DataFormat 没有定义字段，直接返回原始 payload
	if len(dataFormat.Fields) == 0 {
		return rawPayload, nil
	}

	payload := make(map[string]any)

	for _, fieldCfg := range dataFormat.Fields {
		value := getNestedValue(rawPayload, fieldCfg.FieldPath)
		if value == nil {
			return nil, fmt.Errorf("缺失字段: %s (路径: %s)", fieldCfg.FieldKey, fieldCfg.FieldPath)
		}
		payload[fieldCfg.FieldKey] = value
	}

	return payload, nil
}

// GetForEachSpec 返回 ForEachBlock 的规格
func GetForEachSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"ForEach",
		ForEachVersion,
		"遍历数组，为每个元素创建信息原子触发订阅的子图",
		[]string{"input", "output"},
		map[string]any{
			"sourceType": map[string]any{
				"type":        "string",
				"title":       "数据来源",
				"description": "选择从 GraphContext 还是 InfoAtom 中获取数组",
				"enum":        []string{"context", "atom"},
				"enumLabels":  []string{"图上下文", "信息原子"},
				"check":       "must",
				"default":     "context",
			},
			"sourcePath": map[string]any{
				"type":        "string",
				"title":       "数组路径",
				"description": "数组的访问路径，如 'employees.data' 或 'items'",
				"check":       "must",
				"placeholder": "employees.data",
			},
			"iterationInfoAtomTypeId": map[string]any{
				"type":        "string",
				"title":       "迭代信息原子类型",
				"description": "选择用于触发子图的信息原子类型，子图入口节点需订阅该类型",
				"check":       "must",
				"source":      "infoAtomTypeSelector", // 前端使用信息原子类型选择器
			},
			"itemKey": map[string]any{
				"type":        "string",
				"title":       "元素变量名",
				"description": "当前遍历元素在 Payload 中的 key",
				"default":     "item",
				"placeholder": "item",
			},
			"indexKey": map[string]any{
				"type":        "string",
				"title":       "索引变量名",
				"description": "当前索引在 Payload 中的 key",
				"default":     "index",
				"placeholder": "index",
			},
			"continueOnError": map[string]any{
				"type":        "boolean",
				"title":       "失败时继续",
				"description": "单个元素分发失败时是否继续处理后续元素",
				"default":     false,
			},
		},
		[]core.ServiceType{core.ServiceTypeDispatcher, core.ServiceTypeInfoAtomQuery},
	)
}
