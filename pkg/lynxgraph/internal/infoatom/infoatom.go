package infoatom

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// InfoAtom 信息原子注册表接口，用于管理信息原子类型
type InfoAtomRegistry interface {
	InitInfoAtomTable(ctx context.Context) error // 初始化数据库表

	CheckInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey) bool

	GetInfoAtomTypeList(ctx context.Context, LisParams core.InfoAtomList) ([]core.InfoAtomType, int64, int64, int64, error) // 查询信息原子类型
	GetInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey) (core.InfoAtomType, error)                               // 因为只有创建信息原子时才会使用这个接口，因此可以不保存在内存中。因此决定使用SQL数据库+redis来存储
	CreateInfoAtomType(ctx context.Context, infoAtomType core.InfoAtomType) error                                           // 创建新的信息原子类型
	UpdateInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey, infoAtomType core.InfoAtomType) error                 // 更新信息原子类型
	DeleteInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey) error                                                 // 删除信息原子类型

	// CreateInfoAtom 创建信息原子
	//
	// 职责：
	//  1. 解析原始JSON数据
	//  2. 根据InfoAtomType的DataFormat提取字段
	//  3. 验证字段类型和业务规则
	//  4. 解析标签
	//  5. 生成UUID
	//  6. 构造并返回 InfoAtom 对象
	//
	// 参数：
	//  - ctx: 上下文
	//  - req: 信息原子请求
	//  - infoAtomType: 信息原子类型
	//
	// 返回：
	//  - InfoAtom: 构造完成的信息原子对象
	//  - error: 错误信息
	CreateInfoAtom(
		ctx context.Context,
		req core.InfoAtomRequest,
		infoAtomType core.InfoAtomType,
	) (core.InfoAtom, error)
}

// NewInfoAtomRegistry 创建一个新的信息原子注册表
// 参数：
//   - config: 业务配置（包含 CacheConf）
//   - sqlConn: PostgreSQL 连接（外部注入）
//   - getTagNamesFunc: 根据标签ID获取标签名称的函数（用于标签查询）
func NewInfoAtomRegistry(
	config *Config,
	sqlConn sqlx.SqlConn,
	getTagNamesFunc core.GetTagNamesByIDsFunc) (InfoAtomRegistry, error) {
	if config == nil {
		return nil, ErrConfigNil
	}

	return newInfoAtomRegistry(config, sqlConn, getTagNamesFunc), nil
}
