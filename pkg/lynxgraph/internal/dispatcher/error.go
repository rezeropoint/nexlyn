package dispatcher

import "errors"

// 预定义错误
var (
	ErrConfigNil     = errors.New("配置不能为空")
	ErrQueryFuncNil  = errors.New("查询函数不能为空")
	ErrGraphNotFound = errors.New("找不到指定的图")
	ErrNodeNotFound  = errors.New("找不到指定的节点")
	ErrGetFuncNil    = errors.New("获取图函数不能为空")
	ErrDatastoreNil  = errors.New("数据存储不能为空")
)
