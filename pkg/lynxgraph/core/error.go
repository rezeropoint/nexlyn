package core

import "errors"

var (
	ErrConfigNil              = errors.New("配置不能为空")
	ErrRedisClientNil         = errors.New("Redis客户端不能为空")
	ErrTenantIdEmpty          = errors.New("租户ID不能为空")
	ErrKeyEmpty               = errors.New("键不能为空")
	ErrIDEmpty                = errors.New("ID不能为空")
	ErrItemNotFound           = errors.New("项目未找到")
	ErrMarshalFailed          = errors.New("序列化失败")
	ErrUnmarshalFailed        = errors.New("反序列化失败")
	ErrBlockAlreadyRegistered = errors.New("逻辑块已注册")
	ErrSpecNotFound           = errors.New("未找到逻辑块规格")
	ErrFactoryNotFound        = errors.New("未找到逻辑块工厂")
	ErrRegistryNil            = errors.New("BlockRegistry 不能为空")
	ErrRegisterStandardBlock  = errors.New("注册标准积木失败")
	ErrTmplNotStruct          = errors.New("tmpl 必须是结构体或结构体指针")
	ErrNestedFieldMissing     = errors.New("缺少嵌套字段")
	ErrNestedFieldType        = errors.New("嵌套字段类型应为 map[string]any")
	ErrFieldMissingOrZero     = errors.New("字段缺失或为零值")
	ErrNestedFieldCheckFailed = errors.New("嵌套字段校验失败")
	ErrFieldTypeMismatch      = errors.New("字段类型不匹配")
	ErrInvalidConfig          = errors.New("配置文件类型错误")

	// store.go 中的错误
	ErrGraphContextImplement = errors.New("graphContext 必须实现 GetTenantId 方法")
	ErrGraphKeyImplement     = errors.New("graphContext 必须实现 GetGraphKey 方法")
	ErrContextKeyImplement   = errors.New("graphContext 必须实现 GetContextKey 方法")
	ErrSaveGraphContext      = errors.New("保存图上下文失败")
	ErrDataTypeNotByteArray  = errors.New("返回的数据类型不是字节数组")
	ErrDeleteGraphContext    = errors.New("删除图上下文失败")
	ErrInfoAtomTypeEmpty     = errors.New("信息原子的类型不能为空")
	ErrSaveInfoAtom          = errors.New("保存信息原子失败")
	ErrGetInfoAtomType       = errors.New("获取信息原子类型失败")
	ErrDeleteInfoAtom        = errors.New("删除信息原子失败")
	ErrResourceKeyEmpty      = errors.New("资源键不能为空")
	ErrLockOperationEmpty    = errors.New("锁操作返回空")
	ErrLockResultType        = errors.New("获取锁结果类型错误")
	ErrReleaseLock           = errors.New("释放锁失败")

	// Phase 1 重构：删除约束检查错误
	ErrGraphInUse        = errors.New("逻辑图正在运行中")
	ErrInfoAtomTypeInUse = errors.New("信息原子类型正在使用中")
)
