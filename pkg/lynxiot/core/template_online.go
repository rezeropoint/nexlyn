// Package core 定义IoT设备管理的核心领域模型
//
// 文件说明：template_online.go
// 职责：在线检测配置（独立的配置类型）
//
// ⚠️ 注意：此文件仅包含在线检测相关的类型定义
// - 元数据结构 → 见 template_metadata.go
// - 业务数据处理配置 → 见 template_data.go
package core

import "fmt"

// OnlineStrategy 在线检测策略类型
type OnlineStrategy string

const (
	OnlineStrategyFieldCheck OnlineStrategy = "field_check" // 检查字段（支持exists/equals/in操作符）
	OnlineStrategyAnyData    OnlineStrategy = "any_data"    // 收到任何数据即认为在线
)

// FieldCheckOperator 字段检查操作符
type FieldCheckOperator string

const (
	FieldCheckOperatorExists FieldCheckOperator = "exists" // 检查字段是否存在
	FieldCheckOperatorEquals FieldCheckOperator = "equals" // 检查字段值是否等于指定值
	FieldCheckOperatorIn     FieldCheckOperator = "in"     // 检查字段值是否在指定值列表中
)

// FieldCheck 字段检查规则
type FieldCheck struct {
	Field    string             `json:"field" title:"字段路径"`                      // 字段路径（点分隔），如 "status.online", "Key", "BoardId"
	Operator FieldCheckOperator `json:"operator,optional" title:"检查操作"`          // 操作符：exists(存在), equals(等于), in(包含于)，默认equals
	Value    string             `json:"value,optional,omitempty" title:"期望值"`    // 期望值（operator=equals时使用）
	Values   []string           `json:"values,optional,omitempty" title:"期望值列表"` // 期望值列表（operator=in时使用）
}

// OnlineDetectionConfig 独立的在线检测配置（存储在Etcd中）
//
// Etcd Key格式: sensor-online/{category}/{model}/{topicSuffix}
// 例如: sensor-online/ai_box/AI-200/board_ping
//
// 说明：
// - category/model/topicSuffix 组合即为 Etcd key，实现 O(1) 查询
// - 一个主题对应一个在线检测配置（通过 category/model/topicSuffix 唯一标识）
// - 配置变更会触发 MQTT Manager 自动更新订阅
// - TenantID 用于存储到ClickHouse时的多租户隔离
//
// ⚠️ 安全说明：
// TenantID 不由前端直接提供，而是在REST API层通过JWT Token解析后填充
type OnlineDetectionConfig struct {
	Category       string         `json:"category"`              // 设备类别
	Model          string         `json:"model"`                 // 设备型号（关联 iot_sensor_templates.model）
	TopicSuffix    string         `json:"topicSuffix"`           // 主题后缀
	TenantID       string         `json:"tenantId"`              // 租户ID（用于ClickHouse多租户隔离，由REST层从JWT解析后填充）
	Strategy       OnlineStrategy `json:"strategy"`              // 检测策略
	TimeoutSeconds int            `json:"timeoutSeconds"`        // 超时时间（秒）
	FieldChecks    []FieldCheck   `json:"fieldChecks,omitempty"` // 字段检查规则（strategy=field_check时必填）
}

// Etcd配置前缀常量
const (
	OnlineConfigPrefix = "sensor-online/" // 在线检测配置根前缀
)

// BuildOnlineConfigPrefix 构建在线检测配置的型号级前缀
// 格式: sensor-online/{category}/{model}/
// 用于: GetAllKeys 前缀查询、AddPrefixWatcher 监听
func BuildOnlineConfigPrefix(category, model string) string {
	return fmt.Sprintf("%s%s/%s/", OnlineConfigPrefix, category, model)
}

// BuildOnlineDetectionKey 构建在线检测配置的Etcd key
// 格式: sensor-online/{category}/{model}/{topicSuffix}
// 例如: sensor-online/ai_box/AI-200/board_ping
func BuildOnlineDetectionKey(category, model, topicSuffix string) string {
	return fmt.Sprintf("%s%s/%s/%s", OnlineConfigPrefix, category, model, topicSuffix)
}

// Validate 验证在线检测配置的有效性
func (config *OnlineDetectionConfig) Validate() error {
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

	// 验证检测策略
	validStrategies := map[OnlineStrategy]bool{
		OnlineStrategyFieldCheck: true,
		OnlineStrategyAnyData:    true,
	}
	if !validStrategies[config.Strategy] {
		return fmt.Errorf("无效的在线检测策略: %s", config.Strategy)
	}

	// 验证超时时间
	if config.TimeoutSeconds <= 0 {
		return fmt.Errorf("超时时间必须大于0，当前值: %d", config.TimeoutSeconds)
	}

	// 设置默认主题后缀
	if config.TopicSuffix == "" {
		config.TopicSuffix = "status/online"
	}

	// 根据策略验证字段检查配置
	switch config.Strategy {
	case OnlineStrategyFieldCheck:
		// field_check策略：必须配置字段检查
		if len(config.FieldChecks) == 0 {
			return fmt.Errorf("field_check策略需要配置fieldChecks")
		}

		// 验证每个字段检查规则
		for i, check := range config.FieldChecks {
			if check.Field == "" {
				return fmt.Errorf("fieldChecks[%d]: 字段路径不能为空", i)
			}

			// 设置默认operator为exists
			if check.Operator == "" {
				config.FieldChecks[i].Operator = FieldCheckOperatorExists
			}

			// 验证operator是否有效
			validOperators := map[FieldCheckOperator]bool{
				FieldCheckOperatorExists: true,
				FieldCheckOperatorEquals: true,
				FieldCheckOperatorIn:     true,
			}
			if !validOperators[check.Operator] {
				return fmt.Errorf("fieldChecks[%d]: 无效的操作符: %s", i, check.Operator)
			}

			// 验证值配置
			if check.Operator == FieldCheckOperatorEquals && check.Value == "" {
				return fmt.Errorf("fieldChecks[%d]: equals操作符需要配置value", i)
			}
			if check.Operator == FieldCheckOperatorIn && len(check.Values) == 0 {
				return fmt.Errorf("fieldChecks[%d]: in操作符需要配置values", i)
			}
		}

	case OnlineStrategyAnyData:
		// any_data策略：无需字段检查配置
		// 允许配置fieldChecks，但会被忽略
	}

	return nil
}
