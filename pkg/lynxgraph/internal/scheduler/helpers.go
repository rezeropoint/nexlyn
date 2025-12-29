package scheduler

import (
	"time"

	"github.com/google/uuid"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
)

// parseTimezone 解析时区
func parseTimezone(timezone, defaultTimezone string) (*time.Location, error) {
	if timezone == "" {
		timezone = defaultTimezone
	}
	if timezone == "" {
		timezone = "Asia/Shanghai"
	}
	return time.LoadLocation(timezone)
}

// CreateScheduledInfoAtom 创建定时触发的虚拟信息原子
// 用于将定时任务的触发转换为标准的 InfoAtom 处理流程
func CreateScheduledInfoAtom(
	tenantId string,
	graphKey core.GraphKey,
	scheduleConfig *core.ScheduleConfig,
) core.InfoAtom {
	now := time.Now()

	// 构建载荷：合并初始载荷和系统字段
	payload := make(map[string]any)
	for k, v := range scheduleConfig.InitialPayload {
		payload[k] = v
	}

	// 添加系统字段
	payload["_scheduled"] = true
	payload["_scheduleTime"] = now.UnixMilli()
	payload["_graphId"] = graphKey.ID

	// 创建虚拟信息原子类型
	atomType := &core.BasicInfoAtomType{
		TenantId: tenantId,
		ID:       "system:scheduled_trigger",
		Name:     "scheduled_trigger",
		Version:  "1.0",
	}

	return core.NewInfoAtom(
		tenantId,
		uuid.New().String(),
		atomType,
		"scheduler:"+graphKey.ID,
		now.UnixMilli(),
		map[string]string{
			"trigger_type":      "schedule",
			"graph_id":          graphKey.ID,
			"schedule_node_id":  scheduleConfig.NodeID, // 用于精确触发对应的 schedule 节点
		},
		payload,
	)
}
