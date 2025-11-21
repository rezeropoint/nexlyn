package graph

import (
	"fmt"
	"time"
)

// NotifyMessage etcd 通知消息结构
type NotifyMessage struct {
	Action    string `json:"action"`    // 操作类型：create, update, delete
	ID        string `json:"id"`        // 图的唯一标识符
	Timestamp string `json:"timestamp"` // 时间戳
}

// generateEtcdKey 生成 etcd 键
func generateEtcdKey(prefix, action, id string) string {
	return fmt.Sprintf(EtcdKeyFormat, prefix, action, id)
}

// generateNotifyMessage 生成通知消息
func generateNotifyMessage(action, id string) NotifyMessage {
	return NotifyMessage{
		Action:    action,
		ID:        id,
		Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
	}
}
