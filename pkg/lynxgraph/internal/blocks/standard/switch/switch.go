package switchblock

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/zeromicro/go-zero/core/logx"
)

const SwitchVersion = "v1"

// Condition 条件配置
type Condition struct {
	Name       string `json:"name"`       // 条件名称
	Expression string `json:"expression"` // 条件表达式
}

// Config Switch 积木的配置
type Config struct {
	Conditions []Condition `json:"conditions"` // 条件列表
	ResultKey  string      `json:"resultKey"`  // 保存匹配条件名到 GraphContext 的 key
}

// SwitchBlock 实现条件路由逻辑块
type SwitchBlock struct {
	core.BaseLogicBlock
}

// NewSwitchBlock 创建一个新的 SwitchBlock 实例
func NewSwitchBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &SwitchBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeSwitch,
			RawConfig: config,
		},
	}, nil
}

func (b *SwitchBlock) GetID() string                 { return b.ID }
func (b *SwitchBlock) GetType() core.LogicBlockType  { return b.Type }
func (b *SwitchBlock) GetConfigure() map[string]any  { return b.RawConfig }

func (b *SwitchBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 验证条件列表
	if len(cfg.Conditions) == 0 {
		return fmt.Errorf("条件列表不能为空")
	}

	for i, cond := range cfg.Conditions {
		if cond.Name == "" {
			return fmt.Errorf("条件 %d 缺少名称", i)
		}
		if cond.Expression == "" {
			return fmt.Errorf("条件 %s 缺少表达式", cond.Name)
		}
	}

	b.TypedConfig = &cfg
	return nil
}

// Execute 执行条件路由
func (b *SwitchBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	config, ok := b.TypedConfig.(*Config)
	if !ok {
		return false, core.ErrInvalidConfig
	}

	// 创建表达式求值器
	evaluator := NewExpressionEvaluator(ctx, execCtx, datastore)

	// 依次检查条件
	for _, cond := range config.Conditions {
		logx.WithContext(ctx).Debugf("[Switch] 检查条件: %s = %s", cond.Name, cond.Expression)

		result, err := evaluator.Evaluate(cond.Expression)
		if err != nil {
			logx.WithContext(ctx).Errorf("[Switch] 条件 %s 求值失败: %v", cond.Name, err)
			continue // 跳过求值失败的条件
		}

		if result {
			logx.WithContext(ctx).Infof("[Switch] 匹配条件: %s", cond.Name)

			// 保存结果到 GraphContext
			if config.ResultKey != "" {
				graphContext := &core.BaseGraphContext{
					TenantId:   execCtx.GetTenantId(),
					GraphKey:   execCtx.GetGraphKey(),
					ContextKey: config.ResultKey,
					Payload: map[string]any{
						"condition":  cond.Name,
						"expression": cond.Expression,
					},
				}
				if err := datastore.SaveGraphContext(ctx, graphContext); err != nil {
					logx.WithContext(ctx).Errorf("[Switch] 保存结果到 GraphContext 失败: %v", err)
				}
			}

			return true, nil
		}
	}

	logx.WithContext(ctx).Infof("[Switch] 未匹配任何条件")
	return false, nil
}

// GetSwitchSpec 返回 SwitchBlock 的规格
func GetSwitchSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"Switch",
		SwitchVersion,
		"条件路由积木，根据表达式求值结果路由到不同分支，支持多个条件依次检查",
		[]string{"condition", "routing", "logic"},
		map[string]any{
			"conditions": map[string]any{
				"type":        "array",
				"title":       "条件列表",
				"description": "按顺序检查的条件列表，匹配首个为 true 的条件",
				"check":       "must",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type":        "string",
							"title":       "条件名称",
							"description": "用于标识匹配的条件",
						},
						"expression": map[string]any{
							"type":        "string",
							"title":       "表达式",
							"description": "条件表达式，支持 context.xxx 和 atom.xxx 变量",
							"placeholder": "atom.status == 'online' && context.threshold > 10",
						},
					},
				},
			},
			"resultKey": map[string]any{
				"type":        "string",
				"title":       "结果键名",
				"description": "将匹配的条件信息保存到图上下文的键名",
				"placeholder": "switch_result",
				"check":       "",
			},
		},
		[]core.ServiceType{}, // 不依赖任何外部服务
	).WithOutputSchema(map[string]any{
		"contextKey": "config.resultKey", // 上下文键名来自配置
		"type":       "object",
		"properties": map[string]any{
			"condition": map[string]any{
				"type":        "string",
				"description": "匹配的条件名称",
			},
			"expression": map[string]any{
				"type":        "string",
				"description": "匹配的条件表达式",
			},
		},
	})
}
