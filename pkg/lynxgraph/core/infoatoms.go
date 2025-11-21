package core

// InfoAtom 表示一个标准化的事件
type InfoAtom interface {
	GetTenantId() string
	GetID() string
	GetType() InfoAtomType
	GetSource() string
	GetTimestamp() int64
	GetLabels() map[string]string
	GetPayload() map[string]any
}

type BasicInfoAtom struct {
	// ——固定识别维度——
	TenantId  string       `json:"tenantId"`  // 租户ID（必填，用于数据隔离）
	ID        string       `json:"id"`        // 唯一标识（必填，用于去重）
	Type      InfoAtomType `json:"kind"`      // 事件类型（如 "smoke.alarm"、"face.detected"）
	Source    string       `json:"source"`    // 具体来源（设备编号、系统名、数据流 ID 等）
	Timestamp int64        `json:"timestamp"` // 产生时间
	// ——可选识别维度——
	Labels map[string]string `json:"labels"` // 任意键值维度：{ "zone":"A3", "building":"T2", "floor":"5F" }
	// ——数据负载——
	Payload map[string]any `json:"payload"` // 事件有效载荷，结构由具体 Type 决定
}

func NewInfoAtom(tenantId, id string, kind InfoAtomType, source string, timestamp int64, labels map[string]string, payload map[string]any) InfoAtom {
	return &BasicInfoAtom{
		TenantId:  tenantId,
		ID:        id,
		Type:      kind,
		Source:    source,
		Timestamp: timestamp,
		Labels:    labels,
		Payload:   payload,
	}
}

func (i *BasicInfoAtom) GetTenantId() string          { return i.TenantId }
func (i *BasicInfoAtom) GetID() string                { return i.ID }
func (i *BasicInfoAtom) GetType() InfoAtomType        { return i.Type }
func (i *BasicInfoAtom) GetSource() string            { return i.Source }
func (i *BasicInfoAtom) GetTimestamp() int64          { return i.Timestamp }
func (i *BasicInfoAtom) GetLabels() map[string]string { return i.Labels }
func (i *BasicInfoAtom) GetPayload() map[string]any   { return i.Payload }
