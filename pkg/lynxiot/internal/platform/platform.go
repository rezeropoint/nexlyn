package platform

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/rezeropoint/etcdtrigger/v2/engine"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Manager 平台配置管理器接口
type Manager interface {
	Create(ctx context.Context, metadata core.PlatformMetadata, config *core.PlatformConfig) (string, error)                                                   // Create 创建平台配置（元数据+配置），返回生成的UUID
	Get(ctx context.Context, platformType core.PlatformType, id string, tenantID string) (*core.Platform, error)                                               // Get 获取平台配置（包含完整配置）
	GetByID(id string) (*core.PlatformConfig, bool)                                                                                                            // GetByID 根据ID获取平台配置（用于 Dispatcher getConfig 回调）
	List(ctx context.Context, tenantID string, platformType *core.PlatformType, keyword string, page, pageSize int) ([]*core.Platform, int64, error)           // List 列出平台配置（返回Platform，不包含Config）
	Update(ctx context.Context, platformType core.PlatformType, id string, tenantID string, metadata core.PlatformMetadata, config *core.PlatformConfig) error // Update 更新平台配置
	Delete(ctx context.Context, platformType core.PlatformType, id string, tenantID string) error                                                              // Delete 删除平台配置
	Close() error                                                                                                                                              // Close 关闭管理器
}

// NewManager 创建平台配置管理器
func NewManager(dbConn sqlx.SqlConn, config Config, configStore engine.Engine) (Manager, error) {
	return newPlatformManager(dbConn, config, configStore)
}
