# AI Box MQTT 控制功能开发计划

## 一、核心架构原则

1. **子包解耦**：通过 core 包的函数类型定义实现解耦，engine 层负责组装
2. **统一接口**：同类设备的同一控制行为使用统一接口，型号差异在实现层转换
3. **Template Manager 管理配置**：控制配置作为模板的第三种配置类型
4. **型号驱动**：必须指定型号才能配置控制功能

## 二、开发计划

### Phase 1: 核心定义层（core包）

#### 1. 函数类型定义 (`core/controller.go`)

```go
// 控制相关的函数类型定义（用于解耦）
type GetTemplateControlConfigFunc func(templateID string) (*DeviceControlConfig, bool)
type GetDeviceInfoFunc func(deviceID string) (*DeviceBinding, error)
type ValidateDevicePermissionFunc func(deviceID, tenantID string, orgIDs []string) error

// 命令发送函数
type SendControlCommandFunc func(topic string, payload []byte) error

// 响应处理函数
type HandleControlResponseFunc func(deviceID, commandID string, response []byte) error
```

#### 2. AI Box 控制定义 (`core/devices/ai_box_control.go`)

```go
// AI Box 支持的型号
type AIBoxModel string
const (
    AIBoxModelAI200 AIBoxModel = "AI-200"
    AIBoxModelAI300 AIBoxModel = "AI-300"
)

// 算法任务结构（基于真实设备协议）
type AIBoxAlgorithmTask struct {
    AlgTaskSession string          `json:"alg_task_session"` // 任务会话ID
    TaskDesc       string          `json:"task_desc"`        // 任务描述
    MediaName      string          `json:"media_name"`       // 媒体名称
    AlgInfo        []int           `json:"alg_info"`         // 算法信息列表
    AlgTaskStatus  AIBoxTaskStatus `json:"alg_task_status"`  // 任务运行状态
    AlarmBody      int             `json:"alarm_body"`       // 告警主体
    AlarmProtocol  int             `json:"alarm_protocol"`   // 告警协议
    MetadataUrl    interface{}     `json:"metadata_url"`     // 元数据URL（可能是string或[]string）
    UserData       map[string]interface{} `json:"user_data,omitempty"`
}

// AI Box 算法能力（基于真实设备协议）
type AIBoxCapabilities struct {
    BoardId   string         `json:"board_id"`  // 盒子ID
    Abilities []AIBoxAbility `json:"abilities"` // 完整算法能力列表
}

// AI Box 控制命令类型
type AIBoxCommandType string
const (
    AIBoxCommandCreateTask       AIBoxCommandType = "create_task"
    AIBoxCommandUpdateTask       AIBoxCommandType = "update_task"
    AIBoxCommandDeleteTask       AIBoxCommandType = "delete_task"
    AIBoxCommandListTasks        AIBoxCommandType = "list_tasks"
    AIBoxCommandGetCapabilities  AIBoxCommandType = "get_capabilities"  // 获取算法能力
    AIBoxCommandControlTask      AIBoxCommandType = "control_task"      // 控制任务（启停）
)

// 统一的控制命令结构
type AIBoxControlCommand struct {
    CommandID   string               `json:"command_id"`
    CommandType AIBoxCommandType     `json:"command_type"`
    DeviceID    string               `json:"device_id"`
    Model       AIBoxModel           `json:"model"`
    Payload     interface{}          `json:"payload"`
    Timestamp   int64                `json:"timestamp"`
}

// 控制响应
type AIBoxControlResponse struct {
    CommandID   string      `json:"command_id"`
    Status      string      `json:"status"`      // success/failed/timeout
    Result      interface{} `json:"result"`
    Error       string      `json:"error,omitempty"`
    Timestamp   int64       `json:"timestamp"`
}
```

#### 3. 控制配置结构 (`core/template_control.go`)

```go
// 设备控制配置（存储在Etcd）
type DeviceControlConfig struct {
    Category         string            `json:"category"`
    Model            string            `json:"model"`           // 设备型号（必填）
    CommandSuffix    string            `json:"command_suffix"`  // 命令主题后缀
    ResponseSuffix   string            `json:"response_suffix"` // 响应主题后缀
    TenantID         string            `json:"tenant_id"`
}

// 构建控制配置的Etcd key
func BuildControlConfigKey(category, model, suffix string) string {
    return fmt.Sprintf("control-config/%s/%s/%s", category, model, suffix)
}
```

### Phase 2: Controller Manager 实现

#### 1. Manager 接口定义 (`internal/controller/controller.go`)

```go
type Manager interface {
    // 生命周期管理
    Start(ctx context.Context) error
    Stop()

    // AI Box 算法任务管理（统一接口）
    CreateAIBoxTask(ctx context.Context, deviceID string, task core.AIBoxAlgorithmTask) (string, error)
    UpdateAIBoxTask(ctx context.Context, deviceID string, taskID string, task core.AIBoxAlgorithmTask) error
    DeleteAIBoxTask(ctx context.Context, deviceID string, taskID string) error
    ListAIBoxTasks(ctx context.Context, deviceID string) ([]core.AIBoxAlgorithmTask, error)

    // AI Box 算法能力查询
    GetAIBoxCapabilities(ctx context.Context, deviceID string) (*core.AIBoxCapabilities, error)
}
```

#### 2. 配置结构 (`internal/controller/config.go`)

```go
type Config struct {
    // MQTT配置（独立客户端）
    MQTTBroker           string
    MQTTClientID         string
    MQTTUsername         string
    MQTTPassword         string

    // 依赖
    RedisClient          *redis.Redis
    ConfigManager        ec.ConfigManager
    ServiceName          string
    PodName              string

    // 函数注入（解耦依赖）
    GetTemplateControlConfig core.GetTemplateControlConfigFunc
    GetDeviceInfo           core.GetDeviceInfoFunc
    ValidateDevicePermission core.ValidateDevicePermissionFunc
}
```

#### 3. 核心实现逻辑

Controller Manager 的核心实现包括：

1. **命令发送流程**：
   - 获取设备信息和控制配置
   - 构建统一命令结构
   - 根据型号转换命令格式
   - 发送MQTT消息
   - 记录命令到Redis（用于响应匹配）

2. **型号转换器**：
   - AI-200：不支持priority字段，需过滤
   - AI-300：支持全部字段
   - 未来型号：只需添加转换逻辑

3. **响应处理**：
   - 订阅响应主题
   - 异步处理响应
   - 匹配命令并返回结果

### Phase 3: Template Manager 扩展

#### 1. 扩展模板元数据

```go
type TemplateMetadata struct {
    // ... 现有字段
    Model                    string   // 设备型号（控制功能必填）
    ControlTopicSuffixes     []string // 控制主题后缀列表
}
```

#### 2. 控制配置管理

Template Manager 需要添加控制配置管理方法：
- `SetControlConfig`：设置控制配置（验证型号）
- `GetControlConfig`：获取控制配置
- `DeleteControlConfig`：删除控制配置

验证逻辑：
- 必须先设置设备型号
- 型号必须支持控制功能
- 控制配置与型号强关联

### Phase 4: Engine 层组装

#### 1. Engine 接口扩展

```go
type IoT interface {
    // ... 现有接口

    // AI Box 算法任务管理（统一接口）
    CreateAIBoxTask(ctx context.Context, tenantID, deviceID string, task core.AIBoxAlgorithmTask) (string, error)
    UpdateAIBoxTask(ctx context.Context, tenantID, deviceID string, taskID string, task core.AIBoxAlgorithmTask) error
    DeleteAIBoxTask(ctx context.Context, tenantID, deviceID string, taskID string) error
    ListAIBoxTasks(ctx context.Context, tenantID, deviceID string) ([]core.AIBoxAlgorithmTask, error)

    // AI Box 算法能力查询
    GetAIBoxCapabilities(ctx context.Context, tenantID, deviceID string) (*core.AIBoxCapabilities, error)
}
```

#### 2. 依赖注入

Engine 负责组装所有依赖：
- 创建Controller Manager
- 注入获取配置的函数
- 注入设备信息查询函数
- 注入权限验证函数

## 三、文件结构规划

```
pkg/iot/
├── core/
│   ├── controller.go              # 控制领域模型（主题格式、Redis键、GetDeviceInfoFunc）
│   ├── template_control.go        # 控制配置结构（DeviceControlConfig、默认主题）
│   └── devices/
│       ├── ai_box.go              # AI Box设备类别定义
│       └── ai_box_control.go      # AI Box控制协议（真实协议结构体、常量）
│
├── internal/
│   ├── controller/                # 控制管理器（独立MQTT客户端）
│   │   ├── controller.go          # Manager接口
│   │   ├── config.go              # 配置结构（MQTT配置，不含运行时实例）
│   │   ├── handler.go             # 核心实现（生命周期、AI Box接口）
│   │   ├── internal.go            # 内部函数（Redis操作、配置加载、响应等待）
│   │   └── aibox_converter.go     # 协议转换器（真实设备协议适配）
│   │
│   └── template/
│       └── handler.go             # 扩展控制配置管理
│
└── engine/
    ├── engine.go                  # 添加控制接口
    └── handler.go                 # 集成Controller Manager（依赖注入）
```

## 四、配置存储设计

### PostgreSQL 存储

#### 表结构修改（iot_sensor_templates表）

```sql
ALTER TABLE iot_sensor_templates ADD COLUMN model VARCHAR(50);  -- 设备型号
ALTER TABLE iot_sensor_templates ADD COLUMN control_topic_suffixes TEXT[];  -- 控制主题后缀数组
```

#### 初始化SQL文件修改

需要修改 `nexlyn-deploy/postgres-init-scripts/005-iot-devices-schema.sql`：

```sql
-- 在 CREATE TABLE iot_sensor_templates 语句中添加字段
model VARCHAR(50),                              -- 设备型号
control_topic_suffixes TEXT[],                  -- 控制主题后缀数组
```

### Etcd 存储格式

```
control-config/{category}/{model}/{suffix}  -- 控制配置
```

### Redis 键设计

```
iot:control:device:{device_id}             -- 当前正在执行的命令（用于监控和调试，TTL 60秒）
```

**注意**：
- 响应匹配使用设备ID作为key（存储在内存 sync.Map 中）
- 单设备同时只允许一个命令执行（并发控制）
- Redis存储仅用于调试，不影响响应匹配逻辑

## 五、实现优先级

### P0 - 基础框架（第一阶段）✅ 已完成
- [x] core/controller.go - 函数类型定义、主题格式、Redis键构建
- [x] core/devices/ai_box_control.go - 真实协议结构体、常量定义
- [x] core/template_control.go - 控制配置结构、默认主题

### P1 - Controller实现（第二阶段）✅ 已完成
- [x] internal/controller/controller.go - Manager接口
- [x] internal/controller/config.go - 配置结构
- [x] internal/controller/handler.go - 核心实现
- [x] internal/controller/internal.go - 内部辅助函数（Redis追踪、配置加载）
- [x] internal/controller/aibox_converter.go - AI Box转换器

### P2 - 集成完善（第三阶段）✅ 已完成
- [x] internal/template扩展 - 控制配置管理（Create/Update/Delete/Get方法）
- [x] 数据库初始化SQL修改（添加control_topic_suffixes字段）
- [x] engine集成 - Controller Manager集成（依赖注入、权限验证）
- [x] 响应处理机制（sync.Map + channel异步匹配）
- [x] 真实协议适配（基于厂商文档，GetCapabilities和ListTasks已实现）
- [x] 常量化硬编码（Event、Status等定义为常量）

### P3 - 响应匹配优化（第四阶段）✅ 已完成
- [x] 移除 command_id 字段（设备协议不支持）
- [x] 使用设备ID进行响应匹配（符合真实硬件协议）
- [x] 通过 Event 字段验证响应类型（避免并发命令时响应混淆）
- [x] 并发控制（单设备同时只允许一个命令）
- [x] 大响应日志优化（算法能力查询仅记录元信息）
- [x] 任务控制（启停）功能实现（2.1.11 算法任务控制）
- [x] 算法能力缓存（Redis，可配置TTL）

### P4 - 高级功能（未来规划）
- [ ] 命令重试机制
- [ ] 批量命令支持
- [ ] 命令历史审计

## 六、响应匹配机制

### 核心设计
- **匹配Key**: 使用设备ID（从响应的 `BoardId` 字段提取）
- **并发控制**: 单设备同时只允许一个命令执行
- **响应提取**: 从MQTT响应中提取 `BoardId` 和 `Event` 字段
- **响应验证**: 通过 `Event` 字段验证响应类型

### 工作流程
1. 发送命令前检查设备是否有待处理命令
2. 将设备ID注册到 `pendingCommands` (sync.Map)
3. 发送MQTT命令到设备
4. 等待响应（30秒超时）
5. 收到响应后，从 `BoardId` 提取设备ID
6. 使用设备ID查找对应的channel并发送响应
7. 清理 `pendingCommands` 和 Redis 记录

### 日志优化
- 大响应（如算法能力查询）只记录元信息
- 记录算法数量、响应大小等统计信息
- 避免日志系统被数千行数据淹没

## 七、关键设计优势

1. **符合真实硬件协议**
   - 设备用 `BoardId` 识别响应，不支持 `command_id`
   - 消除无法实现的抽象层
   - 提升系统可靠性

2. **完全解耦**
   - Controller不直接依赖其他Manager
   - 通过函数类型注入依赖
   - 符合现有架构模式

3. **统一接口，型号适配**
   - 所有AI Box使用相同的任务管理接口
   - 型号差异在ModelConverter中处理
   - API层无需关心型号差异

4. **扩展性强**
   - 新型号只需添加转换逻辑
   - 新设备类别添加新的控制接口
   - 不影响现有功能

5. **配置灵活**
   - 控制主题后缀可自定义
   - 支持配置热更新
   - 型号与控制功能强关联

6. **并发安全**
   - 明确的单设备单命令限制
   - 防止命令冲突和响应混乱
   - 提供清晰的错误提示

## 八、注意事项

1. **向后兼容**：确保新功能不影响现有的数据采集功能
2. **错误处理**：完善的错误定义和处理机制
3. **日志优化**：大响应只记录元信息，避免日志过载
4. **性能考虑**：控制命令使用独立的Redis键空间
5. **安全验证**：严格的权限验证和设备状态检查
6. **并发限制**：单设备单命令，防止冲突
7. **测试覆盖**：单元测试和集成测试

## 九、开发历程

### 初始实现
1. 实现核心数据结构定义（core包）
2. 实现型号转换器机制
3. 开发Controller Manager基础框架
4. 实现MQTT客户端和命令发送
5. 扩展Template Manager支持控制配置
6. 在Engine层完成集成

### 重构优化（2025-01-22）
1. **问题发现**：command_id 响应匹配失败（设备不返回此字段）
2. **根因分析**：设备协议用 `BoardId` 识别响应，不支持 `command_id`
3. **架构调整**：移除 command_id，改用设备ID进行响应匹配
4. **日志优化**：大响应只记录元信息（算法数量、数据大小）
5. **并发控制**：单设备单命令限制，防止响应混乱
6. **文档更新**：更新 MQTT_CONTROL_PLAN.md 和 DEVELOPMENT.md

### 任务控制实现（2025-01-22）
1. **新功能**：算法任务控制（启停）接口
2. **Event验证**：改进响应匹配，通过 Event 字段验证避免并发响应混淆
3. **类型修正**：MetadataUrl 和 AIBoxParameterOption.Value 保持 interface{} 类型
4. **缓存优化**：算法能力查询增加 Redis 缓存（可配置 TTL）

### 关键收获
- ✅ 不要强加不被支持的抽象层
- ✅ 遵循真实硬件协议设计系统
- ✅ 日志系统需考虑数据量级
- ✅ 明确的并发控制语义
- ✅ Event 字段验证确保响应正确匹配