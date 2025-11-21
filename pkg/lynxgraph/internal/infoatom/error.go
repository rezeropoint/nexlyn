package infoatom

import "errors"

// 预定义错误
var (
	ErrConfigNil                  = errors.New("配置不能为空")
	ErrInfoAtomTypeAlreadyExists  = errors.New("信息原子类型已存在")
	ErrInfoAtomTypeNotFound       = errors.New("信息原子类型未找到")
	ErrInvalidInfoAtomTypeSpec    = errors.New("无效的信息原子类型规格")
	ErrInfoAtomTypeIDEmpty        = errors.New("信息原子类型ID不能为空")
	ErrInfoAtomTypeValidateFailed = errors.New("信息原子类型验证失败")

	// SQL相关错误
	ErrSQLConnectionFailed = errors.New("SQL连接失败")
	ErrSQLQueryFailed      = errors.New("SQL查询失败")
	ErrSQLInsertFailed     = errors.New("SQL插入失败")
	ErrSQLUpdateFailed     = errors.New("SQL更新失败")
	ErrSQLDeleteFailed     = errors.New("SQL删除失败")
	ErrSQLInvalidQuery     = errors.New("SQL无效的查询条件")
	ErrSQLInvalidCacheKey  = errors.New("SQL无效的缓存键")
	ErrNotManagerMode      = errors.New("当前模式不是Manager")
)
