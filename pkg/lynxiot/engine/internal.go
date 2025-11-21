package engine

import (
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core/fields"
)

// validateTTLConfig 验证TTL配置完整性
// 确保所有注册的字段都在配置文件中配置了TTL
func validateTTLConfig(ttlOverrides map[string]int) error {
	// 获取所有注册的字段
	allFields := fields.GetAllFields()
	if len(allFields) == 0 {
		return fmt.Errorf("未找到任何注册的字段，请检查字段注册表")
	}

	// 检查配置是否为空（len(nil map) == 0，无需单独检查nil）
	if len(ttlOverrides) == 0 {
		return fmt.Errorf("TTL配置为空，请在配置文件的ClickHouseConfig.TTLOverrides中为所有字段配置TTL")
	}

	// 检查每个字段是否都有TTL配置
	var missingFields []string
	var invalidFields []string

	for _, field := range allFields {
		ttl, exists := ttlOverrides[field.Name]
		if !exists {
			missingFields = append(missingFields, field.Name)
		} else if ttl <= 0 || ttl > 3650 {
			invalidFields = append(invalidFields, fmt.Sprintf("%s (TTL=%d)", field.Name, ttl))
		}
	}

	// 构建错误信息
	if len(missingFields) > 0 || len(invalidFields) > 0 {
		errMsg := "TTL配置不完整:\n"
		if len(missingFields) > 0 {
			errMsg += fmt.Sprintf("  缺少配置的字段: %v\n", missingFields)
		}
		if len(invalidFields) > 0 {
			errMsg += fmt.Sprintf("  无效的TTL配置: %v (有效范围: 1-3650天)\n", invalidFields)
		}
		errMsg += "请在配置文件的ClickHouseConfig.TTLOverrides中添加或修正这些字段的TTL配置"
		return fmt.Errorf("%s", errMsg)
	}

	return nil
}
