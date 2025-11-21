package core

import (
	"context"

	"github.com/google/uuid"
)

// InfoAtomType 描述一类信息原子（Type）的领域接口
//
// 设计原则：
//   - 这是纯粹的领域接口，不包含任何框架依赖
//   - 不包含展示层方法（如 GetTags()），展示数据由 manager.InfoAtomTypeDTO 承载
type InfoAtomType interface {
	GetTenantId() string                   // 租户ID
	GetID() string                         // 信息原子类型ID
	GetName() string                       // Type 名称，如 "face.detected"
	GetVersion() string                    // 版本，便于灰度/演进
	GetTagIDs() []string                   // 标签ID列表（持久化字段）
	GetDataFormat() DataFormat             // 返回 JSON Schema 字节
	GetCreatedBy() string                  // 创建人用户ID
	GetUpdatedBy() string                  // 更新人用户ID
	GetCreatedAt() int64                   // 创建时间（Unix时间戳）
	GetUpdatedAt() int64                   // 更新时间（Unix时间戳）
	Validate(payload map[string]any) error // 校验具体事件负载
}

// BasicInfoAtomType 信息原子类型的基本实现
//
// 设计原则：
//   - 这是纯粹的领域模型，不包含任何框架依赖（无 bson、gorm、db 等标签）
//   - 存储层实现使用 internal/infoatom/model.go 中的数据库专用结构体
//   - 不包含任何展示层字段（如标签名称），展示数据由 manager.InfoAtomTypeDTO 承载
//
// ⚠️ 关于标签名称（Tags）：
//   - Tags 字段已从 Core 层移除，因为它是展示层概念，不属于领域模型
//   - API 响应需要标签名称时，应使用 manager.InfoAtomTypeDTO（详见 Phase 2 重构计划）
//   - Manager 层会调用 GetTagNamesByIDsFunc 将 TagIDs 转换为标签名称并填充到 DTO
//   - 详见 DEVELOPMENT.md 的"Phase 2: 引入 DTO 模式"章节
type BasicInfoAtomType struct {
	TenantId   string     `json:"tenantId"`
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Version    string     `json:"version"`
	TagIDs     []string   `json:"tagIds"` // 标签ID列表（持久化字段）
	DataFormat DataFormat `json:"dataFormat"`
	CreatedBy  string     `json:"createdBy"` // 创建人用户ID
	UpdatedBy  string     `json:"updatedBy"` // 更新人用户ID
	CreatedAt  int64      `json:"createdAt"` // 创建时间（Unix时间戳）
	UpdatedAt  int64      `json:"updatedAt"` // 更新时间（Unix时间戳）
}

// DataFormat 数据格式描述
type DataFormat struct {
	DataPlural bool          `json:"dataPlural"` // 数据是否为复数
	FieldStart string        `json:"fieldStart"` // 数据是复数时，数组的字段
	Fields     []FieldConfig `json:"fields"`
}

// FieldConfig 字段配置结构体
type FieldConfig struct {
	FieldKey  string    `json:"fieldKey"`  // 字段键
	FieldPath string    `json:"fieldPath"` // 字段路径
	FieldType FieldType `json:"fieldType"` // 字段类型
}

// FieldType 表示字段的类型
type FieldType string

const (
	FieldTypeString FieldType = "string" // 字符串类型
	FieldTypeInt    FieldType = "int"    // 整型
	FieldTypeFloat  FieldType = "float"  // 浮点型
	FieldTypeBool   FieldType = "bool"   // 布尔型
	// 可按需扩展
)

// InfoAtomTypeKey 用于唯一标识信息原子类型的键
type InfoAtomTypeKey struct {
	ID string
	// TenantId string
	// Name      string
	// Version   string
}

type InfoAtomTypeQueryFunc func(ctx context.Context, key InfoAtomTypeKey) (InfoAtomType, error)
type CheckInfoAtomFunc func(ctx context.Context, key InfoAtomTypeKey) bool

func NewInfoAtomType(tenantId, name, version string, tagIDs []string, dataFormat DataFormat) InfoAtomType {
	return &BasicInfoAtomType{
		TenantId:   tenantId,
		ID:         uuid.New().String(),
		Name:       name,
		Version:    version,
		TagIDs:     tagIDs,
		DataFormat: dataFormat,
	}
}

func (i *BasicInfoAtomType) GetTenantId() string { return i.TenantId }
func (i *BasicInfoAtomType) GetID() string       { return i.ID }
func (i *BasicInfoAtomType) GetName() string     { return i.Name }
func (i *BasicInfoAtomType) GetVersion() string  { return i.Version }
func (i *BasicInfoAtomType) GetTagIDs() []string { return i.TagIDs }

// GetTags() 已移除 - 使用 manager.InfoAtomTypeDTO（Phase 2 重构）
func (i *BasicInfoAtomType) GetDataFormat() DataFormat { return i.DataFormat }
func (i *BasicInfoAtomType) GetCreatedBy() string      { return i.CreatedBy }
func (i *BasicInfoAtomType) GetUpdatedBy() string      { return i.UpdatedBy }
func (i *BasicInfoAtomType) GetCreatedAt() int64       { return i.CreatedAt }
func (i *BasicInfoAtomType) GetUpdatedAt() int64       { return i.UpdatedAt }

func (i *BasicInfoAtomType) Validate(payload map[string]any) error {
	return nil
}

// InfoAtomRequest 接收信息原子的请求结构（Engine层统一入口）
//
// 设计原则：
//   - 用于 gRPC/REST 等协议层向 Engine 层传递信息原子数据
//   - RawData 为原始 JSON 字节，由 Engine 层协调 InfoAtomRegistry 进行解析和验证
//   - Timestamp 为 Unix 毫秒时间戳
type InfoAtomRequest struct {
	TenantID       string   `json:"tenantId"`       // 租户ID
	InfoAtomTypeID string   `json:"infoAtomTypeId"` // 信息原子类型ID
	Source         string   `json:"source"`         // 来源标识
	RawData        []byte   `json:"rawData"`        // 原始JSON数据
	Tags           []string `json:"tags"`           // 标签列表（格式：key:value）
	Timestamp      int64    `json:"timestamp"`      // Unix毫秒时间戳
}

// InfoAtomResponse 接收信息原子的响应结构
//
// 设计原则：
//   - 用于 Engine 层向协议层返回处理结果
//   - Success 表示是否成功接收和分发信息原子
//   - ErrorCode 用于标准化错误处理
type InfoAtomResponse struct {
	Success    bool   `json:"success"`    // 是否成功
	Message    string `json:"message"`    // 响应消息
	InfoAtomID string `json:"infoAtomId"` // 生成的信息原子ID（成功时返回）
	ErrorCode  string `json:"errorCode"`  // 错误码（失败时返回）
}
