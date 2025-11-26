// Package httpreceive 提供 HTTP 数据接收功能的 Manager 实现
//
// 本包负责：
// - HTTP 接收配置的 CRUD 操作
// - HTTP 接收数据的处理和分发
//
// 遵循 IoT 引擎的三层架构设计，位于 Internal 层。
package httpreceive

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/rezeropoint/etcdtrigger/v2/engine"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Manager HTTP数据接收管理器接口
// 遵循 Template Manager 的设计模式
type Manager interface {
	// 配置 CRUD
	// Create 创建HTTP接收配置
	// - metadata: 元数据（名称、描述、启用状态等）
	// - config: HTTP接收配置（字段映射、分发配置等）
	// - 返回生成的配置ID
	Create(ctx context.Context, metadata core.HttpReceiveMetadata, config *core.HttpReceiveConfig) (string, error)

	// Update 更新HTTP接收配置
	// - id: 配置ID
	// - tenantID: 租户ID（用于权限验证）
	// - metadata: 元数据
	// - config: HTTP接收配置
	Update(ctx context.Context, id string, tenantID string, metadata core.HttpReceiveMetadata, config *core.HttpReceiveConfig) error

	// Delete 删除HTTP接收配置
	Delete(ctx context.Context, id string, tenantID string) error

	// Get 获取完整的HTTP接收配置信息（元数据 + 配置 + 审计信息）
	Get(ctx context.Context, id string, tenantID string) (*core.HttpReceive, error)

	// List 查询HTTP接收配置列表（返回摘要信息）
	List(ctx context.Context, query core.HttpReceiveQuery) ([]*core.HttpReceiveSummary, int64, error)

	// 数据处理
	ProcessData(ctx context.Context, configId string, data map[string]any) error

	// 生命周期
	Close() error
}

// NewManager 创建HTTP数据接收管理器
func NewManager(
	dbConn sqlx.SqlConn,
	config Config,
	configStore engine.Engine,
	dispatchInfoFunc core.DispatchInfoFunc,
) (Manager, error) {
	return newHttpReceiveManager(dbConn, config, configStore, dispatchInfoFunc)
}
