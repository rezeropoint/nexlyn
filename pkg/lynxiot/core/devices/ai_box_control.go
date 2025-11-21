package devices

// AI Box 型号定义
type AIBoxModel string

const (
	AIBoxModelAI200 AIBoxModel = "AI-200" // AI-200 型号
	AIBoxModelAI300 AIBoxModel = "AI-300" // AI-300 型号
)

// AI Box 控制命令类型
type AIBoxCommandType string

const (
	AIBoxCommandCreateTask      AIBoxCommandType = "create_task"      // 创建算法任务
	AIBoxCommandUpdateTask      AIBoxCommandType = "update_task"      // 更新算法任务
	AIBoxCommandDeleteTask      AIBoxCommandType = "delete_task"      // 删除算法任务
	AIBoxCommandListTasks       AIBoxCommandType = "list_tasks"       // 查询任务列表
	AIBoxCommandGetCapabilities AIBoxCommandType = "get_capabilities" // 获取算法能力
	AIBoxCommandControlTask     AIBoxCommandType = "control_task"     // 控制算法任务（启动/停止）
)

// AI Box 真实设备事件类型（Event字段值）
const (
	AIBoxEventAbilityFetch = "/alg_ability_fetch" // 算法能力获取事件
	AIBoxEventTaskFetch    = "/alg_task_fetch"    // 任务列表查询事件
	AIBoxEventTaskControl  = "/alg_task_control"  // 任务控制事件（启动/停止）
)

// AI Box 控制响应状态
const (
	AIBoxStatusSuccess = "success" // 成功
	AIBoxStatusFailed  = "failed"  // 失败
	AIBoxStatusTimeout = "timeout" // 超时
)

// AIBoxAlgorithmTask 算法任务结构（基于真实设备协议）
type AIBoxAlgorithmTask struct {
	// 任务标识
	AlgTaskSession string `json:"alg_task_session"` // 任务会话ID
	TaskDesc       string `json:"task_desc"`        // 任务描述
	MediaName      string `json:"media_name"`       // 媒体名称

	// 算法信息
	AlgInfo       []int           `json:"alg_info"`        // 算法信息列表
	AlgTaskStatus AIBoxTaskStatus `json:"alg_task_status"` // 任务运行状态

	// 告警配置
	AlarmBody     int `json:"alarm_body"`     // 告警主体
	AlarmProtocol int `json:"alarm_protocol"` // 告警协议

	// 元数据配置
	MetadataUrl interface{} `json:"metadata_url"` // 元数据URL（可能是string或[]string）

	// 用户自定义数据
	UserData map[string]interface{} `json:"user_data,omitempty"`
}

// AIBoxCapabilities AI Box算法能力（基于真实设备协议）
type AIBoxCapabilities struct {
	BoardId   string         `json:"board_id"`  // 盒子ID
	Abilities []AIBoxAbility `json:"abilities"` // 完整算法能力列表
}

// AIBoxControlCommand 统一的控制命令结构
type AIBoxControlCommand struct {
	CommandType AIBoxCommandType `json:"command_type"` // 命令类型
	DeviceID    string           `json:"device_id"`    // 设备ID
	Model       AIBoxModel       `json:"model"`        // 设备型号
	Payload     interface{}      `json:"payload"`      // 命令载荷
	Timestamp   int64            `json:"timestamp"`    // 时间戳（Unix毫秒）
}

// AIBoxControlResponse 控制响应
type AIBoxControlResponse struct {
	DeviceID  string      `json:"device_id"`       // 设备ID（从响应的 BoardId 提取）
	Event     string      `json:"event"`           // 事件类型（从响应的 Event 提取）
	Status    string      `json:"status"`          // 状态（success/failed/timeout）
	Result    interface{} `json:"result"`          // 响应结果
	Error     string      `json:"error,omitempty"` // 错误信息
	Timestamp int64       `json:"timestamp"`       // 时间戳（Unix毫秒）
}

// ============================================================================
// 真实设备协议数据结构（基于设备厂商文档）
// ============================================================================

// AIBoxRealCommand 真实设备命令格式
type AIBoxRealCommand struct {
	BoardId string `json:"BoardId"` // 设备唯一标识
	Event   string `json:"Event"`   // 功能标识（如：/alg_ability_fetch）
}

// AIBoxResult 真实设备响应的通用Result结构
type AIBoxResult struct {
	Code int    `json:"Code"` // 错误标识（0正常，其他错误）
	Desc string `json:"Desc"` // 描述信息
}

// ============================================================================
// 算法能力查询相关结构体
// ============================================================================

// AIBoxParameterOption 算法参数选项（SELECTOR类型使用）
type AIBoxParameterOption struct {
	Enable bool        `json:"enable"` // 当前选项是否可选
	Key    string      `json:"key"`    // 选项标识
	Value  interface{} `json:"value"`  // 选项序号（可能是int或空字符串""）
	Name   string      `json:"name"`   // 选项名称
}

// AIBoxParameter 算法可设置的配置项
type AIBoxParameter struct {
	Class    string                 `json:"class"`             // 设置项数据类型描述（FLOAT/BOOLEAN/SELECTOR/INTEGER）
	Type     int                    `json:"type"`              // 设置项数据类型编号（FLOAT=>2, BOOLEAN=>4, SELECTOR=>5, INTEGER=>0）
	Max      interface{}            `json:"max,omitempty"`     // 可设置最大值（FLOAT/INTEGER时存在）
	Min      interface{}            `json:"min,omitempty"`     // 可设置最小值（FLOAT/INTEGER时存在）
	Default  interface{}            `json:"default"`           // 设置项默认值
	Key      string                 `json:"key"`               // 设置项标识
	Name     string                 `json:"name"`              // 设置项名称
	Required bool                   `json:"required"`          // 是否必填
	Value    interface{}            `json:"value"`             // 设置值
	Options  []AIBoxParameterOption `json:"options,omitempty"` // SELECTOR待选列表
}

// AIBoxAttribute 算法属性配置
type AIBoxAttribute struct {
	LineRequired bool   `json:"lineRequired"` // 是否必须配置辅助线
	LineDesc     string `json:"lineDesc"`     // 配置辅助线的描述文字
	ZoneRequired bool   `json:"zoneRequired"` // 是否必须配置算法专用区域
	ZoneDesc     string `json:"zoneDesc"`     // 配置算法专用区域的描述文字
}

// AIBoxPolicy 算法报警规则
type AIBoxPolicy struct {
	Property string `json:"property"` // 违规的属性名（告警类型）
	Name     string `json:"name"`     // 属性名描述（告警类型的中文描述）
}

// AIBoxAbility 单个算法能力
type AIBoxAbility struct {
	Attribute  AIBoxAttribute   `json:"attribute"`  // 算法属性
	Parameters []AIBoxParameter `json:"parameters"` // 算法可设置的配置项
	Permitted  bool             `json:"permitted"`  // 算法是否已授权
	Code       int              `json:"code"`       // 主算法编号
	Sub        bool             `json:"sub"`        // 是否是子算法
	Name       string           `json:"name"`       // 算法名称
	Desc       string           `json:"desc"`       // 算法描述
	Item       int              `json:"item"`       // 子算法编号
	Policy     []AIBoxPolicy    `json:"policy"`     // 报警规则
}

// AIBoxAbilityResponse 算法能力查询响应
type AIBoxAbilityResponse struct {
	BoardIp string         `json:"BoardIp"` // 盒子网络地址
	BoardId string         `json:"BoardId"` // 盒子唯一标识
	Ability []AIBoxAbility `json:"Ability"` // 算法列表
	Event   string         `json:"Event"`   // 功能标识（/alg_ability_fetch）
	Result  AIBoxResult    `json:"Result"`  // 接口状态返回
}

// ============================================================================
// 算法任务查询相关结构体
// ============================================================================

// AIBoxTaskStatus 任务运行状态
type AIBoxTaskStatus struct {
	Type  int    `json:"type"`  // 状态级别标识（1/2/3/4）
	Style string `json:"style"` // 状态级别（normal/warning/danger/success）
	Label string `json:"label"` // 状态描述
}

// AIBoxTaskInfo 单个任务详情
type AIBoxTaskInfo struct {
	AlarmBody      int                    `json:"AlarmBody"`          // 告警主体
	AlarmProtocol  int                    `json:"AlarmProtocol"`      // 告警协议
	AlgInfo        []int                  `json:"AlgInfo"`            // 算法信息
	AlgTaskSession string                 `json:"AlgTaskSession"`     // 任务会话标识
	AlgTaskStatus  AIBoxTaskStatus        `json:"AlgTaskStatus"`      // 任务运行状态
	MediaName      string                 `json:"MediaName"`          // 媒体名称
	MetadataUrl    interface{}            `json:"MetadataUrl"`        // 元数据URL（可能是string或[]string）
	TaskDesc       string                 `json:"TaskDesc"`           // 任务描述
	UserData       map[string]interface{} `json:"UserData,omitempty"` // 用户数据
}

// AIBoxTaskResponse 算法任务查询响应
type AIBoxTaskResponse struct {
	BoardId string          `json:"BoardId"` // 盒子唯一标识
	BoardIp string          `json:"BoardIp"` // 盒子网络地址
	Content []AIBoxTaskInfo `json:"Content"` // 任务列表
	Event   string          `json:"Event"`   // 功能标识（/alg_task_fetch）
	Result  AIBoxResult     `json:"Result"`  // 接口状态返回
}

// ============================================================================
// 算法任务控制相关结构体（2.1.11 算法任务控制）
// ============================================================================

// AIBoxTaskControlCommand 算法任务控制命令（发送到设备）
type AIBoxTaskControlCommand struct {
	BoardId        string `json:"BoardId"`        // 盒子唯一标识
	Event          string `json:"Event"`          // 功能标识（固定为 /alg_task_control）
	ControlCommand int    `json:"ControlCommand"` // 启停标识（0停止, 1启动）
	AlgTaskSession string `json:"AlgTaskSession"` // 任务标识
}

// AIBoxTaskControlResponse 算法任务控制响应（从设备接收）
type AIBoxTaskControlResponse struct {
	BoardId        string      `json:"BoardId"`        // 盒子唯一标识
	BoardIp        string      `json:"BoardIp"`        // 盒子网络地址
	Event          string      `json:"Event"`          // 功能标识（/alg_task_control）
	AlgTaskSession string      `json:"AlgTaskSession"` // 任务标识
	Result         AIBoxResult `json:"Result"`         // 接口状态返回
}
