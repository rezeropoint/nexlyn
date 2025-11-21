package fields

import (
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
)

// 字段名常量
const (
	FieldNameTemperature core.FieldName = "temperature"
	FieldNameHumidity    core.FieldName = "humidity"
	FieldNameSmoke       core.FieldName = "smoke"
	FieldNameIsOnline    core.FieldName = "is_online"
)

// standardFieldRegistry 标准字段注册表
var standardFieldRegistry = make(map[core.FieldName]core.StandardField)

// RegisterField 注册标准字段
func RegisterField(field core.StandardField) {
	standardFieldRegistry[core.FieldName(field.Name)] = field
}

// GetFieldByName 根据字段名获取预定义的标准字段
// 如果字段不存在，返回 nil
func GetFieldByName(name core.FieldName) *core.StandardField {
	if field, exists := standardFieldRegistry[name]; exists {
		return &field
	}
	return nil
}

// GetAllFields 获取所有预定义的标准字段列表
func GetAllFields() []core.StandardField {
	fields := make([]core.StandardField, 0, len(standardFieldRegistry))
	for _, field := range standardFieldRegistry {
		fields = append(fields, field)
	}
	return fields
}

// GetFieldsByType 根据FieldType获取标准字段列表
func GetFieldsByType(fieldType core.FieldType) []core.StandardField {
	fields := make([]core.StandardField, 0)
	for _, field := range standardFieldRegistry {
		if field.FieldType == fieldType {
			fields = append(fields, field)
		}
	}
	return fields
}

// ValidateFieldExists 验证字段名是否在注册表中
// 如果字段不存在，返回错误信息；否则返回nil
func ValidateFieldExists(name core.FieldName) error {
	if GetFieldByName(name) == nil {
		return fmt.Errorf("字段 '%s' 未在标准字段注册表中注册", name)
	}
	return nil
}
