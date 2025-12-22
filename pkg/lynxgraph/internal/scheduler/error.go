package scheduler

import "errors"

var (
	// ErrConfigNil 配置不能为空
	ErrConfigNil = errors.New("配置不能为空")
	// ErrTriggerFuncNil 触发函数不能为空
	ErrTriggerFuncNil = errors.New("触发函数不能为空")
	// ErrNotStarted 调度器未启动
	ErrNotStarted = errors.New("调度器未启动")
	// ErrAlreadyStarted 调度器已启动
	ErrAlreadyStarted = errors.New("调度器已启动")
	// ErrAlreadyStopped 调度器已停止
	ErrAlreadyStopped = errors.New("调度器已停止")
)
