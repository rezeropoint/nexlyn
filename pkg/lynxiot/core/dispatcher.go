package core

import (
	"context"
	"fmt"
)

// GetPlatformConfigFunc 获取平台配置的函数类型（由 platform.Manager.GetByID 提供）
type GetPlatformConfigFunc func(id string) (*PlatformConfig, bool)

// DispatchInfoFunc 数据分发函数类型
type DispatchInfoFunc func(ctx context.Context, configs []DispatchConfig, data *map[string]TypedValue, taskInfo TaskInfo) error

// DispatchType 分发类型
type DispatchType string

const (
	DispatchTypeSkylarkFlows DispatchType = "skylark_flows" // Skylark 流程
	DispatchTypeSkylarkForms DispatchType = "skylark_forms" // Skylark 表单
	DispatchTypeLynxGraph    DispatchType = "lynxgraph"     // LynxGraph 逻辑引擎
	DispatchTypeLog          DispatchType = "log"           // 日志记录
)

// DispatchConfig 分发配置（通用化设计，支持多平台扩展）
type DispatchConfig struct {
	Type       DispatchType `json:"type" title:"分发类型"`                        // 分发操作类型
	PlatformID string       `json:"platformID,optional,omitempty" title:"平台配置ID"` // 平台配置ID（关联 PlatformConfig.ID）

	// Skylark 平台操作参数
	FlowID int64 `json:"flowID,optional,omitempty" title:"流程ID"` // Skylark 流程ID（skylark_flows 使用）
	FormID int64 `json:"formID,optional,omitempty" title:"表单ID"` // Skylark 表单ID（skylark_forms 使用）

	// LynxGraph 平台操作参数
	InfoAtomTypeID string `json:"infoAtomTypeID,optional,omitempty" title:"信息原子类型ID"` // LynxGraph 信息原子类型ID

	// 通用扩展参数（预留其他平台使用）
	ExtraParams map[string]any `json:"extraParams,optional,omitempty" title:"扩展参数"` // 其他平台特定参数
}

// Validate 验证分发配置的有效性
func (c *DispatchConfig) Validate() error {
	// 验证分发类型
	validTypes := map[DispatchType]bool{
		DispatchTypeSkylarkFlows: true,
		DispatchTypeSkylarkForms: true,
		DispatchTypeLynxGraph:    true,
		DispatchTypeLog:          true,
	}
	if !validTypes[c.Type] {
		return fmt.Errorf("无效的分发类型: %s", c.Type)
	}

	// 验证平台操作的配置
	switch c.Type {
	case DispatchTypeSkylarkFlows:
		if c.PlatformID == "" {
			return fmt.Errorf("skylark_flows 分发必须指定 platformID")
		}
		if c.FlowID <= 0 {
			return fmt.Errorf("skylark_flows 分发必须指定有效的 flowID")
		}
	case DispatchTypeSkylarkForms:
		if c.PlatformID == "" {
			return fmt.Errorf("skylark_forms 分发必须指定 platformID")
		}
		if c.FormID <= 0 {
			return fmt.Errorf("skylark_forms 分发必须指定有效的 formID")
		}
	case DispatchTypeLynxGraph:
		if c.InfoAtomTypeID == "" {
			return fmt.Errorf("lynxgraph 分发必须指定 infoAtomTypeID")
		}
		// LynxGraph 是内部服务，不需要 platformID
	case DispatchTypeLog:
		// 日志类型无需额外验证
	}

	return nil
}
