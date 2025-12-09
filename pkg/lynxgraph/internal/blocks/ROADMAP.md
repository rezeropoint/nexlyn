# LynxGraph逻辑积木开发路线图

## 职能边界原则

### 边 (Edge) 的职责
- ✅ 简单的布尔条件表达式（`condition`字段）
- ✅ 定义节点间的流转关系
- ✅ 轻量级的路由决策（如：`status == "success"`）
- ❌ **不应包含**：复杂业务逻辑、数据转换、外部调用

### 积木 (Block) 的职责
- ✅ 复杂的业务逻辑处理
- ✅ 数据转换和计算
- ✅ 外部服务集成
- ✅ 状态持久化管理
- ✅ 多条件复杂路由

**设计哲学**：**边应该"轻"（连接关系），积木应该"重"（业务逻辑）**

---

## 积木的通用能力

### 数据访问模式
所有积木都应支持以下数据访问模式（通过ExecutionContext）：

**从ExecutionContext直接访问**：
- **信息原子**：`execCtx.GetInfoAtom().GetPayload()["field"]`
- **租户ID**：`execCtx.GetTenantId()`
- **图信息**：`execCtx.GetGraph().GetName()`
- **节点信息**：`execCtx.GetNode().GetID()`

**从Store读取图上下文**：
```go
graphKey := execCtx.GetGraphKey()
graphContext, err := datastore.GetGraphContext(ctx, execCtx.GetTenantId(), graphKey, contextKey)
```

**嵌套字段访问**：通过辅助函数解析路径如`data.sensor.temperature.value`

### 上下文写入
积木通过 `datastore.SaveGraphContext()` 批量写入键值对到图上下文，供后续节点使用。

### 模板变量
支持在配置中使用模板变量（语法：`${}`）：
- `${value}` - 引用当前处理的值
- `${infoAtom.data.temperature}` - 引用信息原子字段（需通过辅助函数解析）
- `${context.alert_level}` - 引用上下文数据

---

## 已实现的逻辑块

| 积木 | 类型 | 路径 | 版本 |
|------|------|------|------|
| Log (日志打印) | log | standard/log/log.go | v1 |
| Skylark流程创建 | skylark_journey_create | skylark/journeys/create.go | v1 |
| 传感器数据查询 | fetch_sensor_data | standard/queryhistorydata/queryhistorydata.go | v1 |
| 统计分析 | statistical_analyzer | standard/statistical/statistical.go | v1 |
| 变化率计算 | rate_of_change | standard/rateofchange/rateofchange.go | v1 |
| 异常检测 | anomaly_detector | standard/anomaly/anomaly.go | v1 |
| 趋势分析 | trend_analyzer | standard/trend/trend.go | v1 |
| HTTP请求 | http_request | standard/httprequest/httprequest.go | v1 |
| 时间窗口检查 | time_window_check | standard/timewindow/timewindow.go | v1 |
| 条件路由 | switch | standard/switch/switch.go | v1 |
| 假期检查 | holiday_check | standard/holidaycheck/holidaycheck.go | v1 |
| 外部数据库查询 | query_database | standard/querydatabase/querydatabase.go | v1 |
| 去重检查 | dedup_check | standard/dedupcheck/dedupcheck.go | v1 |

---

## 待实现的逻辑块

### 🔴 第一阶段：核心数据处理类（立即开发）

#### 1. RuleMatcher（规则匹配器）⭐ P0

**功能要求**：
- 通用的条件匹配引擎，支持复杂表达式求值
- 根据匹配的规则执行动作（批量设置上下文）
- 支持多级规则、优先级匹配、模板变量

**典型场景**：
- 多级告警（温度阈值判断）
- 设备类型路由
- 复合条件判断（地理位置 + 时间）

**实现要点**：
- 使用 `github.com/antonmedv/expr` 求值条件表达式
- 支持模板变量解析（${} 语法）
- 从ExecutionContext构建求值环境
- 批量写入上下文键值对到Store

**预计工时**：3-4天 | **难度**：⭐⭐⭐

---

#### 2. Transform（数据转换器）P0

**功能要求**：
- 将信息原子的数据格式转换为另一种格式
- 支持字段映射、嵌套字段处理、类型转换

**典型场景**：
- IoT设备数据格式统一（不同厂商设备字段名不一致）
- API响应数据结构转换
- 数据脱敏（删除敏感字段）

**实现要点**：
- 支持嵌套字段路径（如 `data.sensor.temperature`）
- 支持类型转换（string → int、float → string等）
- 字段不存在时使用默认值或跳过

**预计工时**：2-3天 | **难度**：⭐⭐

---

#### 3. Switch（多路路由器）P0

**功能要求**：
- 根据复杂条件将信息原子路由到不同的后续节点
- 支持多个case分支 + 默认分支
- 避免边定义的"爆炸"问题

**典型场景**：
- 根据设备类型路由到不同处理分支
- 根据事件严重级别路由（info/warning/error/critical）
- 复杂的业务规则路由

**实现要点**：
- 使用表达式引擎（如 `github.com/antonmedv/expr`）
- 按顺序匹配case，第一个匹配成功执行
- 支持默认分支（所有case都不匹配时执行）
- 边通过 `condition` 字段匹配Switch的分支标签

**预计工时**：3-4天 | **难度**：⭐⭐⭐

---

#### 4. HTTPCall（HTTP调用）P0

**功能要求**：
- 调用外部HTTP API并将响应写入上下文
- 支持GET/POST/PUT/DELETE等方法
- 支持Header、Query、Body动态构造

**典型场景**：
- 调用Webhook通知外部系统
- 查询第三方API获取数据
- 触发远程服务（如发送短信、推送通知）

**实现要点**：
- 使用go-zero的 `rest` 包或标准 `net/http`
- 支持模板变量替换（从信息原子和上下文提取）
- 错误处理：超时、网络错误、非2xx状态码
- 可选的重试机制（指数退避）
- 响应解析：JSON自动解析，其他格式保存原始字符串

**预计工时**：3-5天 | **难度**：⭐⭐⭐

---

#### 5. Delay（延迟处理）P1

**功能要求**：
- 延迟一段时间后继续执行后续节点
- 支持相对延迟（延迟N秒）和绝对延迟（延迟到指定时间）

**典型场景**：
- 事件触发5分钟后发送提醒
- 延迟到工作日上午9点执行任务
- 冷却期控制（防止频繁操作）

**实现要点**：
- 使用Redis的 `ZADD` + `ZRANGEBYSCORE` 实现延迟队列
- 或使用消息队列（如RabbitMQ的延迟队列插件）
- 定时器协程扫描到期任务并重新触发
- 需要持久化延迟任务，防止重启丢失

**预计工时**：4-6天 | **难度**：⭐⭐⭐⭐

---

#### 6. Notification（通知发送）P1

**功能要求**：
- 发送邮件/短信/推送通知
- 支持模板和多通道

**典型场景**：
- IoT设备告警通知（温度过高、设备离线）
- 事件流程提醒（工单分配、超时告警）
- 系统运维通知

**实现要点**：
- 集成通知服务（如阿里云短信、SendGrid邮件）
- 支持模板引擎（如 `text/template`）
- 异步发送，不阻塞主流程
- 发送失败时记录日志或重试

**预计工时**：3-4天 | **难度**：⭐⭐⭐

---

### 🟡 第二阶段：流程控制类（尽快需要）

#### 7. Debounce（防抖）P1

**功能要求**：
- 在时间窗口内只处理最后一次事件
- 用于高频事件去重

**典型场景**：
- 传感器每100ms上报数据，只处理5秒内最后一次
- 用户快速点击按钮，只处理最后一次
- 搜索输入框，停止输入500ms后才查询

**实现要点**：
- 使用Redis存储最后一次事件的时间戳和数据
- TTL设置为窗口时间 + 一定余量
- 每次事件到达时重置TTL，窗口结束后触发后续节点

**预计工时**：4-5天 | **难度**：⭐⭐⭐⭐

---

#### 8. Throttle（节流）P1

**功能要求**：
- 限制时间窗口内的处理次数
- 用于流量控制和速率限制

**典型场景**：
- 设备1分钟内最多处理10次数据上报
- API限流（每秒最多100次请求）
- 防止异常设备导致的流量风暴

**实现要点**：
- 使用Redis计数器 + 滑动窗口
- 超过限制时直接丢弃或延迟处理

**预计工时**：3-4天 | **难度**：⭐⭐⭐

---

#### 9. Enrich（数据丰富）P1

**功能要求**：
- 从外部源（数据库/缓存/API）获取额外数据并附加到信息原子
- 支持多种数据源和缓存机制

**典型场景**：
- 根据device_id查询设备名称、所属组织
- 根据user_id查询用户权限
- 查询外部系统的配置信息

**实现要点**：
- 支持多种数据源（Redis、PostgreSQL、HTTP）
- 查询结果自动序列化为JSON并写入指定路径
- 支持Redis缓存避免重复查询
- 查询失败时记录日志但不阻塞流程

**预计工时**：3-4天 | **难度**：⭐⭐⭐

---

#### 10. Split（数据拆分）P2

**功能要求**：
- 将一个信息原子拆分为多个
- 支持数组展开和并行处理

**典型场景**：
- 批量事件上报（一次上报多个传感器数据）
- 数组数据展开处理（每个元素单独处理）
- 消息分发（一对多）

**实现要点**：
- 遍历数组字段，为每个元素创建新的信息原子
- FanOut模式：并发触发所有后续节点
- 非FanOut模式：串行处理

**预计工时**：3-4天 | **难度**：⭐⭐⭐

---

#### 11. Cache（缓存操作）P2

**功能要求**：
- 读写Redis缓存
- 支持常见缓存操作（GET/SET/DELETE/INCR/DECR）

**典型场景**：
- 快速访问设备状态
- 计数器（访问次数、错误次数）
- 分布式锁

**实现要点**：
- 使用go-zero的 `cache` 包
- 支持模板变量替换
- 操作失败时记录日志但不中断流程

**预计工时**：2-3天 | **难度**：⭐⭐

---

### 🟢 第三阶段：高级功能类（长期规划）

#### 12. Merge（数据合并/等待门）P3

**功能要求**：
- 等待多个分支的信息原子到达后合并处理
- 复杂事件处理（CEP）的基础能力

**实现难度**：非常高，需要分布式协调、超时处理、死锁检测

**预计工时**：7-10天 | **难度**：⭐⭐⭐⭐⭐

---

#### 13. Retry（重试机制）P3

**功能要求**：
- 包裹其他积木，失败时自动重试

**注意**：可能与工作池的重试机制重复，需评估必要性

**预计工时**：4-5天 | **难度**：⭐⭐⭐⭐

---

#### 14. DatabaseQuery（数据库查询）P3

**功能要求**：
- 执行复杂SQL查询并将结果写入上下文

**与Enrich的区别**：
- **Enrich**：简单查询（单表、主键查询）
- **DatabaseQuery**：复杂查询（联表、聚合、子查询）

**预计工时**：3-4天 | **难度**：⭐⭐⭐

---

#### 15. MQTTPublish（MQTT发布）P2

**功能要求**：
- 向MQTT主题发布消息
- 用于IoT控制指令下发

**实现要点**：
- 集成 `pkg/iot/internal/mqtt/` 的MQTT Manager
- 支持QoS配置

**预计工时**：3-4天 | **难度**：⭐⭐⭐

---

#### 16. PatternMatching（模式匹配）P3

**功能要求**：
- 检测复杂事件模式（CEP）
- 如："事件A发生后5分钟内事件B也发生"

**实现难度**：非常高，建议后期集成专业CEP引擎（如Apache Flink、Esper）

**预计工时**：10-15天 | **难度**：⭐⭐⭐⭐⭐

---

## 实施计划

### 第一阶段（立即开发，预计2-3周）
1. RuleMatcher（规则匹配器）
2. Transform（数据转换器）
3. Switch（多路路由器）
4. HTTPCall（HTTP调用）

### 第二阶段（1-2个月）
5. Debounce（防抖）
6. Throttle（节流）
7. Notification（通知发送）
8. Enrich（数据丰富）
9. Cache（缓存操作）

### 第三阶段（长期规划，3-6个月）
10. 其他高级功能积木（根据业务需求优先级调整）

---

## 开发规范

### 目录结构
```
internal/blocks/
├── block/              # 逻辑块注册表
├── skylark/            # Skylark集成类积木
├── standard/           # 标准逻辑块
│   ├── log/
│   ├── queryhistorydata/
│   ├── statistical/
│   ├── rateofchange/
│   ├── anomaly/
│   ├── trend/
│   └── ...
└── ROADMAP.md
```

### 代码规范

1. **配置结构体**：
   - **必须**使用 `json` 标签定义字段名（`FillConfig` 只处理有 json 标签的字段）
   - 使用 `check:"must"` 标签标记必填字段
   - 使用 `core.FillConfig()` 进行配置填充和校验

2. **逻辑块实现**：
   - 实现 `core.LogicBlock` 接口
   - **Execute方法签名**：
     ```go
     func (b *MyBlock) Execute(ctx context.Context, execCtx core.ExecutionContext,
                               datastore core.Store, service core.Service) (bool, error)
     ```
   - 通过ExecutionContext访问执行上下文
   - 通过datastore读写图上下文
   - 嵌入 `core.BaseLogicBlock` 提供基本实现
   - 提供工厂函数（`NewXXX`）
   - 定义 `GetXxxSpec()` 方法返回 `BlockSpec` 规格信息

3. **JSON Schema定义**：
   - 通过 `core.NewBasicBlockSpec()` 创建BlockSpec
   - 在第5个参数中定义完整的configSchema
   - configSchema中每个字段必须包含：
     - `type`：字段类型（string/integer/array/number/boolean）
     - `description`：字段用户提示
     - `check`：是否必填（"must"或""）
     - `enum`（可选）：枚举值列表
     - `items`（可选）：数组元素类型定义
     - `default`（可选）：默认值

4. **错误处理**：
   - 使用 `fmt.Errorf("%w", err)` 包装错误
   - 在 `core/error.go` 中定义核心错误

5. **注册逻辑块**：
   - 在 `internal/blocks/block/loader.go` 中的 `RegisterStandardBlocks()` 注册
   - 检查服务依赖是否满足

### 测试规范
1. 单元测试覆盖率 > 80%
2. 集成测试验证与Redis/PostgreSQL的交互
3. 性能测试验证高并发场景

---

## 职能边界速查表

| 组件 | 职责 | 不应负责 | 示例 |
|------|------|----------|------|
| **Edge（边）** | 简单条件、流转关系 | 复杂逻辑、数据转换、外部调用 | `status == "success"` |
| **Switch积木** | 多路复杂路由 | 数据处理 | 根据10种设备类型路由 |
| **Transform积木** | 数据格式转换 | 外部调用 | IoT数据映射 |
| **HTTPCall积木** | HTTP协议调用 | 业务状态管理 | 调用Webhook |
| **Enrich积木** | 数据查询和附加 | 数据转换 | 查询设备名称 |
| **Debounce积木** | 时间窗口去重 | 数据处理 | 高频事件防抖 |
| **Throttle积木** | 速率限制 | 数据处理 | 流量控制 |

---

## 参考资料
- LynxGraph架构文档：`pkg/lynxgraph/DEVELOPMENT.md`
- 逻辑块接口定义：`pkg/lynxgraph/core/block.go`
- 逻辑块规格定义：`pkg/lynxgraph/core/spec.go`
- Skylark集成示例：`pkg/lynxgraph/internal/blocks/skylark/journeys/create.go`
- Log积木示例：`pkg/lynxgraph/internal/blocks/standard/log/log.go`
