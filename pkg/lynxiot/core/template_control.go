// Package core 定义IoT设备管理的核心领域模型
//
// 文件说明：template_control.go
// 职责：设备控制配置（MQTT命令/响应主题配置）
//
// ⚠️ 注意：此文件仅包含控制配置相关的类型定义
// - 元数据结构 → 见 template_metadata.go
// - 在线检测配置 → 见 template_online.go
// - 业务数据处理配置 → 见 template_data.go
package core

import "fmt"

// 默认主题后缀常量
// 注意：这些默认值基于AI Box设备的真实协议
const (
	DefaultCommandSuffix  = "/edge_app_controller"       // 默认命令主题后缀（AI Box真实协议）
	DefaultResponseSuffix = "/edge_app_controller_reply" // 默认响应主题后缀（AI Box真实协议）
)

// DeviceControlConfig 设备控制配置（存储在Etcd中）
//
// Etcd Key格式: control-config/{category}/{model}/{suffix}
// 例如: control-config/ai_box/AI-200//edge_app_controller
//
// 说明：
// - category/model/suffix 组合即为 Etcd key，实现 O(1) 查询
// - 一个型号对应一个控制配置（通过 category/model/suffix 唯一标识）
// - 配置变更会触发 Controller Manager 自动更新订阅
// - TenantID 用于权限验证和多租户隔离
//
// ⚠️ 安全说明：
// TenantID 不由前端直接提供，而是在REST API层通过JWT Token解析后填充
type DeviceControlConfig struct {
	Category       string `json:"category"`        // 设备类别
	Model          string `json:"model"`           // 设备型号（必填，关联 iot_sensor_templates.model）
	CommandSuffix  string `json:"command_suffix"`  // 命令主题后缀（如：/edge_app_controller）
	ResponseSuffix string `json:"response_suffix"` // 响应主题后缀（如：/edge_app_controller_reply）
	TenantID       string `json:"tenant_id"`       // 租户ID（由REST层从JWT解析后填充）
}

// Etcd配置前缀常量
const (
	ControlConfigPrefix = "control-config/" // 控制配置根前缀
)

// BuildControlConfigPrefix 构建控制配置的型号级前缀
// 格式: control-config/{category}/{model}/
// 用于: GetAllKeys 前缀查询
func BuildControlConfigPrefix(category, model string) string {
	return fmt.Sprintf("%s%s/%s/", ControlConfigPrefix, category, model)
}

// BuildControlConfigKey 构建控制配置的Etcd key
// 格式: control-config/{category}/{model}/{suffix}
// 示例: control-config/ai_box/AI-200//edge_app_controller
func BuildControlConfigKey(category, model, suffix string) string {
	return fmt.Sprintf("%s%s/%s/%s", ControlConfigPrefix, category, model, suffix)
}

// Validate 验证控制配置的有效性
func (config *DeviceControlConfig) Validate() error {
	// 验证必填字段
	if config.Category == "" {
		return fmt.Errorf("设备类别不能为空")
	}
	if config.Model == "" {
		return fmt.Errorf("设备型号不能为空")
	}
	if config.TenantID == "" {
		return fmt.Errorf("租户ID不能为空")
	}

	// 设置默认主题后缀
	if config.CommandSuffix == "" {
		config.CommandSuffix = DefaultCommandSuffix
	}
	if config.ResponseSuffix == "" {
		config.ResponseSuffix = DefaultResponseSuffix
	}

	return nil
}
