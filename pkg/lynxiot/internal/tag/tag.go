package tag

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Manager 设备标签管理器接口
type Manager interface {
	// CreateTag 创建标签
	CreateTag(ctx context.Context, metadata core.DeviceTagMetadata) (string, error)

	// GetTag 获取标签详情
	GetTag(ctx context.Context, tagID string, tenantID string) (*core.DeviceTag, error)

	// ListTags 查询标签列表
	ListTags(ctx context.Context, query core.DeviceTagQuery) ([]*core.DeviceTagSummary, int64, error)

	// UpdateTag 更新标签
	UpdateTag(ctx context.Context, tagID string, tenantID string, update core.DeviceTagUpdate) error

	// DeleteTag 删除标签
	DeleteTag(ctx context.Context, tagID string, tenantID string) error
}

// NewManager 创建标签管理器实例
func NewManager(dbConn sqlx.SqlConn, config Config) (Manager, error) {
	return newTagManager(dbConn, config)
}
