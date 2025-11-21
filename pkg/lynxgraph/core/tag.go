package core

import (
	"context"
	"fmt"
)

// 标签作用域常量
const (
	ScopeInfoAtom   = "info_atom"  // 信息原子类型作用域
	ScopeLogicGraph = "logic_graph" // 逻辑图作用域
)

// GetTagNamesByIDsFunc 定义根据标签ID列表批量查询标签名称的函数类型
// 用于解耦internal子包之间的依赖，避免循环依赖
type GetTagNamesByIDsFunc func(ctx context.Context, tagIDs []string, tenantID string) ([]string, error)

// LynxTag 标签完整信息
type LynxTag struct {
	ID          string `json:"id"`          // 标签ID（UUID）
	Name        string `json:"name"`        // 标签名称
	Description string `json:"description"` // 标签描述
	Scope       string `json:"scope"`       // 标签作用域：info_atom / logic_graph
	TenantID    string `json:"tenantId"`    // 租户ID（UUID）
	CreatedBy   string `json:"createdBy"`   // 创建者用户ID（UUID）
	UpdatedBy   string `json:"updatedBy"`   // 最后修改者用户ID（UUID）
	CreatedAt   int64  `json:"createdAt"`   // 创建时间（Unix秒）
	UpdatedAt   int64  `json:"updatedAt"`   // 更新时间（Unix秒）
}

// LynxTagSummary 标签摘要（用于列表展示）
type LynxTagSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Scope       string `json:"scope"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

// LynxTagMetadata 标签元数据（用于创建标签）
type LynxTagMetadata struct {
	Name        string `json:"name"`                    // 标签名称（必填）
	Description string `json:"description,omitempty"`  // 标签描述
	Scope       string `json:"scope"`                  // 标签作用域：info_atom / logic_graph（必填）
	TenantID    string `json:"tenantId"`               // 租户ID（UUID，必填）
	CreatedBy   string `json:"createdBy,omitempty"`    // 创建者用户ID（UUID）
}

// LynxTagUpdate 标签更新信息
type LynxTagUpdate struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// LynxTagQuery 标签查询条件
type LynxTagQuery struct {
	TenantID string `json:"tenantId"`        // 租户ID（UUID，必填）
	Scope    string `json:"scope,omitempty"` // 按作用域筛选：info_atom / logic_graph（可选）
	Keyword  string `json:"keyword,omitempty"` // 关键词搜索（标签名称或描述）
	Page     int    `json:"page"`            // 页码（从1开始）
	PageSize int    `json:"pageSize"`        // 每页数量
}

// Validate 验证标签元数据的有效性
func (metadata *LynxTagMetadata) Validate() error {
	if metadata.Name == "" {
		return fmt.Errorf("标签名称不能为空")
	}
	if metadata.TenantID == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metadata.Scope == "" {
		return fmt.Errorf("标签作用域不能为空")
	}
	if metadata.Scope != ScopeInfoAtom && metadata.Scope != ScopeLogicGraph {
		return fmt.Errorf("标签作用域只能是 %s 或 %s", ScopeInfoAtom, ScopeLogicGraph)
	}
	return nil
}

// ValidateQuery 验证查询条件的有效性
func (query *LynxTagQuery) ValidateQuery() error {
	if query.TenantID == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if query.Scope != "" && query.Scope != ScopeInfoAtom && query.Scope != ScopeLogicGraph {
		return fmt.Errorf("标签作用域只能是 %s 或 %s", ScopeInfoAtom, ScopeLogicGraph)
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	return nil
}
