package core

import (
	"errors"
	"fmt"
)

// Error 错误定义
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// 通用错误
var (
	ErrInvalidParameter = Error{Code: "INVALID_PARAMETER", Message: "无效参数"}
	ErrNotFound         = Error{Code: "NOT_FOUND", Message: "资源不存在"}
	ErrAlreadyExists    = Error{Code: "ALREADY_EXISTS", Message: "资源已存在"}
	ErrInternalError    = Error{Code: "INTERNAL_ERROR", Message: "内部错误"}
	ErrNotImplemented   = Error{Code: "NOT_IMPLEMENTED", Message: "功能暂未实现"}
)

// 模板相关错误
var (
	ErrTemplateNotFound      = Error{Code: "TEMPLATE_NOT_FOUND", Message: "设备模板不存在"}
	ErrTemplateAlreadyExists = Error{Code: "TEMPLATE_ALREADY_EXISTS", Message: "设备型号已存在"}
	ErrTemplateInvalid       = Error{Code: "TEMPLATE_INVALID", Message: "设备模板配置无效"}
	ErrTemplateInUse         = Error{Code: "TEMPLATE_IN_USE", Message: "设备模板正在使用中，无法删除"}
)

// 设备类别相关错误
var (
	ErrInvalidDeviceCategory = Error{Code: "INVALID_DEVICE_CATEGORY", Message: "无效的设备类别"}
	ErrCategoryNotFound      = Error{Code: "CATEGORY_NOT_FOUND", Message: "设备类别不存在"}
)

// 字段映射相关错误
var (
	ErrInvalidFieldMapping   = Error{Code: "INVALID_FIELD_MAPPING", Message: "无效的字段映射配置"}
	ErrRequiredFieldMissing  = Error{Code: "REQUIRED_FIELD_MISSING", Message: "必填字段缺失"}
	ErrStandardFieldNotFound = Error{Code: "STANDARD_FIELD_NOT_FOUND", Message: "标准字段不存在"}
)

// 设备绑定相关错误
var (
	ErrDeviceNotFound      = Error{Code: "DEVICE_NOT_FOUND", Message: "设备不存在"}
	ErrDeviceAlreadyBound  = Error{Code: "DEVICE_ALREADY_BOUND", Message: "设备已绑定"}
	ErrDeviceInvalid       = Error{Code: "DEVICE_INVALID", Message: "设备信息无效"}
	ErrDeviceModelNotFound = Error{Code: "DEVICE_MODEL_NOT_FOUND", Message: "设备型号模板不存在"}
	ErrDeviceUnauthorized  = Error{Code: "DEVICE_UNAUTHORIZED", Message: "无权限操作该设备"}
	ErrDeviceOffline       = Error{Code: "DEVICE_OFFLINE", Message: "设备离线，无法执行控制命令"}
)

// 标签相关错误
var (
	ErrTagNotFound     = Error{Code: "TAG_NOT_FOUND", Message: "标签不存在"}
	ErrTagNameConflict = Error{Code: "TAG_NAME_CONFLICT", Message: "标签名称已存在"}
	ErrTagInUse        = Error{Code: "TAG_IN_USE", Message: "标签正在使用中"}
	ErrTagInvalid      = Error{Code: "TAG_INVALID", Message: "标签信息无效"}
	ErrTagUnauthorized = Error{Code: "TAG_UNAUTHORIZED", Message: "无权限操作该标签"}
)

// 平台配置相关错误
var (
	ErrPlatformNotFound      = Error{Code: "PLATFORM_NOT_FOUND", Message: "平台配置不存在"}
	ErrPlatformAlreadyExists = Error{Code: "PLATFORM_ALREADY_EXISTS", Message: "平台配置ID已存在"}
	ErrPlatformInvalid       = Error{Code: "PLATFORM_INVALID", Message: "平台配置无效"}
	ErrPlatformInUse         = Error{Code: "PLATFORM_IN_USE", Message: "平台配置正在使用中，无法删除"}
	ErrPlatformUnauthorized  = Error{Code: "PLATFORM_UNAUTHORIZED", Message: "无权限操作该平台配置"}
)

// HTTP数据接收配置相关错误
var (
	ErrHttpReceiveNotFound     = Error{Code: "HTTP_RECEIVE_NOT_FOUND", Message: "HTTP接收配置不存在"}
	ErrHttpReceiveInvalid      = Error{Code: "HTTP_RECEIVE_INVALID", Message: "HTTP接收配置无效"}
	ErrHttpReceiveUnauthorized = Error{Code: "HTTP_RECEIVE_UNAUTHORIZED", Message: "无权限操作该HTTP接收配置"}
	ErrHttpReceiveDisabled     = Error{Code: "HTTP_RECEIVE_DISABLED", Message: "HTTP接收配置已禁用"}
)

// 定义工作池错误
var (
	ErrTaskQueueFull    = errors.New("任务队列已满")
	ErrWorkerPoolClosed = errors.New("工作池已关闭")
)
