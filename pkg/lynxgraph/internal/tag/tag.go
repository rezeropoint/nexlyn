package tag

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Manager 标签管理器接口
type Manager interface {
	// InitTagTable 初始化数据库表
	InitTagTable(ctx context.Context) error

	// CreateTag 创建标签
	CreateTag(ctx context.Context, metadata core.LynxTagMetadata) (string, error)

	// GetTag 获取标签详情
	GetTag(ctx context.Context, tagID string, tenantID string) (*core.LynxTag, error)

	// ListTags 查询标签列表
	ListTags(ctx context.Context, query core.LynxTagQuery) ([]*core.LynxTagSummary, int64, error)

	// UpdateTag 更新标签
	UpdateTag(ctx context.Context, tagID string, tenantID string, update core.LynxTagUpdate) error

	// DeleteTag 删除标签（硬删除）
	DeleteTag(ctx context.Context, tagID string, tenantID string) error

	// GetTagsByNames 根据标签名称批量查询（用于前端标签自动补全）
	// 返回指定租户指定作用域下匹配的标签列表
	GetTagsByNames(ctx context.Context, tenantID string, scope string, names []string) ([]*core.LynxTagSummary, error)

	// GetTagsByIDs 根据标签ID列表批量查询标签详情
	// 返回指定租户下匹配的标签列表（用于填充关联资源的标签详情）
	GetTagsByIDs(ctx context.Context, tagIDs []string, tenantID string) ([]*core.LynxTagSummary, error)

	// GetTagNamesByIDs 根据标签ID列表批量查询标签名称
	// 返回标签名称列表，顺序与输入的tagIDs顺序一致（用于前端展示）
	// 如果某个ID不存在，对应位置返回空字符串
	GetTagNamesByIDs(ctx context.Context, tagIDs []string, tenantID string) ([]string, error)

	// GetTagUsage 查询标签使用情况
	// 返回该标签被多少个信息原子类型和逻辑图使用，以及使用该标签的资源ID列表
	GetTagUsage(ctx context.Context, tagID string, tenantID string) (*TagUsage, error)

	// DeleteTagWithCheck 删除标签前检查使用情况
	// force=true时强制删除（会先解除所有关联）
	// force=false时如果被使用则返回错误
	DeleteTagWithCheck(ctx context.Context, tagID string, tenantID string, force bool) error
}

// TagUsage 标签使用情况统计
type TagUsage struct {
	InfoAtomTypeCount int      `json:"infoAtomTypeCount"` // 被多少个信息原子类型使用
	GraphConfigCount  int      `json:"graphConfigCount"`  // 被多少个逻辑图使用
	InfoAtomTypeIDs   []string `json:"infoAtomTypeIds"`   // 使用该标签的信息原子类型ID列表
	GraphConfigIDs    []string `json:"graphConfigIds"`    // 使用该标签的逻辑图ID列表
}

// Config 标签管理器配置
type Config struct {
	// 预留配置字段
}

// NewManager 创建标签管理器实例
func NewManager(dbConn sqlx.SqlConn, config Config) (Manager, error) {
	return newTagManager(dbConn, config)
}
