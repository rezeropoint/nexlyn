package template

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/rezeropoint/etcdtrigger/v2/engine"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Manager 模板管理器接口
type Manager interface {
	// 创建模板（支持在线检测配置 + 业务数据配置 + 控制配置）
	// - onlineConfig 为 nil 表示不创建在线检测配置
	// - businessConfig 为 nil 表示不创建业务数据配置
	// - controlConfig 为 nil 表示不创建控制配置
	Create(ctx context.Context, metadata core.TemplateMetadata, onlineConfig *core.OnlineDetectionConfig, businessConfig *core.DataProcessingConfig, controlConfig *core.DeviceControlConfig) (string, error)

	// 更新模板（支持在线检测配置 + 业务数据配置 + 控制配置）
	// - onlineConfig 为 nil 表示删除在线检测配置
	// - businessConfig 为 nil 表示删除业务数据配置
	// - controlConfig 为 nil 表示删除控制配置
	Update(ctx context.Context, id string, metadata core.TemplateMetadata, onlineConfig *core.OnlineDetectionConfig, businessConfig *core.DataProcessingConfig, controlConfig *core.DeviceControlConfig) error

	// 删除模板（级联删除所有关联配置）
	Delete(ctx context.Context, id string) error

	// 获取模板详情
	Get(ctx context.Context, id string) (*core.Template, error)

	// 查询模板列表
	List(ctx context.Context, query core.TemplateQuery) ([]*core.TemplateSummary, int64, error)
}

// NewManager 创建模板管理器
func NewManager(dbConn sqlx.SqlConn, configStore engine.Engine, config Config) (Manager, error) {
	return newTemplateManager(dbConn, configStore, config)
}
