package engine

import "errors"

// 预定义错误
var (
	ErrConfigNil                = errors.New("配置不能为空")
	ErrGraphRegistryFailed      = errors.New("创建默认图注册表失败")
	ErrBlockRegistryFailed      = errors.New("创建默认区块注册表失败")
	ErrDataStoreFailed          = errors.New("创建默认数据存储失败")
	ErrStandardBlocksFailed     = errors.New("注册标准区块失败")
	ErrInfoAtomRegistryFailed   = errors.New("创建默认信息原子注册表失败")
	ErrDispatcherRegistryFailed = errors.New("创建默认调度器注册表失败")

	// handler.go 中的错误
	ErrEngineAlreadyRunning     = errors.New("引擎已经在运行中")
	ErrEngineNotRunning         = errors.New("引擎未在运行")
	ErrStartDispatcherFailed    = errors.New("启动调度器失败")
	ErrCloseDispatcherFailed    = errors.New("关闭调度器失败")
	ErrCloseGraphRegistryFailed = errors.New("关闭图注册表失败")
	ErrSaveInfoAtomFailed       = errors.New("保存信息原子失败")
	ErrDispatchInfoAtomFailed   = errors.New("分发信息原子失败")
	ErrDirectRegisterGraph      = errors.New("当前接口不支持直接注册图，请使用LoadGraph方法加载图配置")
	ErrDirectUnregisterGraph    = errors.New("当前接口不支持直接注销图，请使用LoadGraph方法加载图配置")
	ErrGraphConfigNil           = errors.New("图配置不能为空")
	ErrLoadGraphFailed          = errors.New("加载图失败")
)
