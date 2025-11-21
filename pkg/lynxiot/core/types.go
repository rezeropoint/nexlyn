package core

// FieldName 字段名类型
type FieldName string

// FieldType 字段类型
type FieldType string

// TypedValue 带类型的值
type TypedValue struct {
	Type  string
	Value any
}
