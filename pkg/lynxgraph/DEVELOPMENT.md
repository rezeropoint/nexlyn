# LynxGraph 引擎开发指南

> **文档定位**：本文档面向 LynxGraph 引擎的**内部开发者**，阐述架构设计理念、开发规范和最佳实践。
> **使用说明**：API 使用请参考 `README.md`（待创建）。

---

## 📐 架构设计理念

### 1. 四层架构与依赖规则

LynxGraph 采用**清晰分层架构**，严格遵循依赖倒置原则（DIP）：

```
┌─────────────────────────────────────────────┐
│  REST Layer (restful/lynxmanager/)         │  ← 外部接口层
├─────────────────────────────────────────────┤
│  Manager/Engine Layer (manager/, engine/)  │  ← 公共 API 层
├─────────────────────────────────────────────┤
│  Internal Layer (internal/)                │  ← 内部实现层
├─────────────────────────────────────────────┤
│  Core Layer (core/)                        │  ← 核心领域层（最内层）
└─────────────────────────────────────────────┘

依赖方向：外层 → 内层（Core 层不依赖任何外层）
```

#### **Core 层的黄金法则**

**Core 层应该是什么**：
- ✅ **纯粹的领域模型**：实体、值对象、领域服务接口
- ✅ **业务规则的载体**：领域逻辑（如 `ScanConfig()` 校验图结构）
- ✅ **稳定的抽象**：接口定义（如 `InfoAtomType`、`LogicGraph`、`LogicBlock`）
- ✅ **框架无关**：不依赖任何外部框架（go-zero、MongoDB、PostgreSQL）

**Core 层不应该是什么**：
- ❌ **不是 DTO（数据传输对象）**：不应包含"用于前端展示"的字段（如标签名称）
- ❌ **不是 ORM 模型**：不应出现 `bson`、`gorm`、`db` 等存储标签
- ❌ **不是 REST 响应体**：不应为了方便 API 返回而污染领域模型
- ❌ **不包含展示逻辑**：标签名称等展示字段应在 REST 层通过 Converter 查询填充

**参考实现**：
- ✅ 好的例子：`pkg/iot/core/device_binding.go`（纯领域模型，无框架依赖）
- ✅ 好的例子：`pkg/skylarkq/core/platform.go`（清晰的值对象和接口）
- ✅ 好的例子：`pkg/lynxgraph/core/grap.go`（Phase 1 已移除展示层字段 `Tags`）

#### **Manager/Engine 层的职责边界**

Manager/Engine 层是对外暴露的公共 API，负责业务编排：

**职责**：
- ✅ **返回纯 Core 对象**：所有查询方法返回领域模型，不包含展示字段
- ✅ **业务逻辑编排**：协调多个 Internal Manager 完成复杂操作
- ✅ **事务管理**：跨多个存储的事务控制
- ✅ **权限验证**：验证租户和组织权限（通过 TenantId、OrgId 参数）
- ✅ **暴露工具方法**：如 `GetTagNamesByIDs()` 供 REST 层使用

**禁止**：
- ❌ **不返回 DTO**：不负责填充展示字段（如标签名称）
- ❌ **不处理 HTTP**：不依赖 HTTP 协议和 REST 框架
- ❌ **不做格式转换**：不负责时间戳格式、单位转换（秒转毫秒等）

**示例**：
```go
// ✅ 正确：返回 Core 对象
func (m *Manager) GetGraphConfig(ctx, key) (core.GraphConfig, error)

// ❌ 错误：返回 DTO（应在 REST 层转换）
func (m *Manager) GetGraphConfig(ctx, key) (GraphConfigDTO, error)
```

---

#### **Internal 层的职责边界**

Internal 层负责具体实现，采用 **Manager 模式**：

| Manager | 职责 | 存储方案 | 关键文件 |
|---------|------|----------|----------|
| **InfoAtom Registry** | 信息原子类型管理 | PostgreSQL + Redis 缓存 | `internal/infoatom/` |
| **Graph Registry** | 逻辑图配置管理 | PostgreSQL + MongoDB + Etcd | `internal/graph/` |
| **Tag Manager** | 标签定义 CRUD | PostgreSQL | `internal/tag/` |
| **Block Registry** | 逻辑块注册表 | 内存 | `internal/blocks/` |
| **Dispatcher** | 信息原子调度 | 内存工作队列 | `internal/dispatcher/` |
| **Schedule Registry** | 定时任务调度 | 内存 + 分布式锁 | `internal/scheduler/` |

**文件组织规范**（所有 Manager 必须遵守）：
- `<manager>.go` - 接口定义和核心类型
- `handler.go` - 接口实现和公共业务逻辑
- `model.go` - 数据库查询结构体（使用 `sql.Null*` 处理可空字段）
- `helpers.go` - 纯函数辅助方法（无 receiver）
- `internal.go` - 内部方法（有 receiver）和私有业务逻辑
- `sql.go` - SQL 语句常量
- `config.go` - Manager 配置结构

---

#### **REST 层的职责边界**

REST 层负责 HTTP 接口和数据转换：

**职责**：
- ✅ **DTO 转换**：将 Core 对象转换为前端需要的格式（通过 Converter）
- ✅ **填充展示字段**：查询标签名称等展示数据并填充到响应
- ✅ **格式转换**：时间戳单位转换（秒转毫秒）、枚举转字符串等
- ✅ **HTTP 协议处理**：请求解析、响应序列化、错误码映射
- ✅ **参数验证**：验证请求参数的合法性

**禁止**：
- ❌ **不包含业务逻辑**：不直接操作数据库和缓存
- ❌ **不处理权限**：权限验证已在 Manager 层完成

**示例**：
```go
// restful/lynxmanager/internal/svc/converter.go
func (svc *ServiceContext) ConvertCoreGraphConfigToMetadata(
    ctx context.Context,
    config core.GraphConfig,  // Manager 层返回的 Core 对象
) types.GraphConfigMetadata {
    metadata := types.GraphConfigMetadata{
        Id:          config.ID,
        TenantId:    config.TenantId,
        Name:        config.Name,
        Tags:        config.TagIDs,  // 初始为 TagIDs
        // ...
    }

    // ✅ REST 层负责查询标签名称（展示逻辑）
    if len(config.TagIDs) > 0 {
        tags, err := svc.LynxManager.GetTagNamesByIDs(ctx, config.TagIDs, config.TenantId)
        if err == nil {
            metadata.Tags = tags  // 替换为标签名称
        }
    }

    return metadata
}
```

---

## 🎯 核心概念与运行模式

### 核心概念

- **信息原子 (InfoAtom)**：系统中流动的事件数据单元，包含类型、来源、时间戳、标签和载荷
- **逻辑图 (LogicGraph)**：由节点和边组成的有向图，定义业务逻辑流程
- **节点 (Node)**：图中的执行单元，每个节点关联一个逻辑块
- **边 (Edge)**：连接节点的有向边，可附带条件表达式
- **逻辑块 (LogicBlock)**：可执行的业务逻辑单元（如过滤器、状态机、动作执行器）
- **引擎 (Engine)**：负责接收信息原子、分发到对应图并执行
- **管理器 (Manager)**：负责信息原子类型和逻辑图配置的 CRUD 操作

### 运行模式

LynxGraph 支持两种互斥的运行模式：

| 模式 | 用途 | 核心功能 | 禁止操作 |
|------|------|----------|----------|
| **Engine 模式** | 生产环境运行 | 接收信息原子并执行逻辑图、监听 Etcd 热更新、工作池异步调度 | 禁止直接注册/注销图 |
| **Manager 模式** | 配置管理服务 | 信息原子类型 CRUD、逻辑图配置 CRUD、通过 Etcd 通知 Engine | 禁止执行信息原子分发 |

**配置方式**：通过 `Config.RunMode` 字段指定（`core.Engine` 或 `core.Manager`）

---

## 🔐 权限与多租户机制

### 设计哲学

- **源头控制**：权限在逻辑图创建时确定，不依赖积木配置
- **自动继承**：积木执行时自动继承逻辑图的权限范围（通过 `ExecutionContext.GetVisibleOrgIDs()`）
- **最小权限**：逻辑图只能访问其归属组织及子组织的数据
- **租户隔离**：不同租户的逻辑图完全隔离，无法相互访问

### 权限传递链

```
创建逻辑图（Manager 模式）
  └─> 保存 tenant_id + org_id 到 graph_configs 表

执行逻辑图（Engine 模式）
  └─> 查询可见组织 ID 列表（org_id + 所有子组织）
      └─> 构建 ExecutionContext（注入 visibleOrgIDs）
          └─> 积木自动继承 visibleOrgIDs
              └─> 查询数据时应用权限过滤（WHERE org_id IN (...)）
```

**关键接口**：
- `ExecutionContext.GetTenantId()` - 获取租户 ID
- `ExecutionContext.GetVisibleOrgIDs()` - 获取可见组织 ID 列表（待实现）
- `ExecutionContext.GetOrgId()` - 获取逻辑图归属的组织 ID

**实施状态**：⏳ Phase 3（graph_configs 表需添加 `org_id` 字段，待数据库迁移）

---

## 🏗️ 数据存储架构

LynxGraph 采用**混合存储架构**，根据数据特性选择存储方案：

| 存储层 | 数据类型 | 特性 | 关键表/集合 |
|--------|----------|------|------------|
| **PostgreSQL** | 元数据和索引 | 强一致性、ACID 事务 | `info_atom_types`、`graph_configs`、`lynxgraph_tags` |
| **MongoDB** | 详细配置 | 灵活的文档结构 | `graph` 集合（节点、边配置） |
| **Redis** | 运行时状态 | 高性能缓存、TTL 自动过期 | 图上下文、信息原子缓存、分布式锁 |
| **Etcd** | 配置变更通知 | 分布式一致性、Watch 机制 | `{prefix}/{action}/{id}` |

**混合存储策略**：
- **创建流程**：PostgreSQL（事务开始）→ MongoDB → Etcd 通知 → PostgreSQL（事务提交）
- **查询优化**：List 只查 PostgreSQL，Get 先验证 PostgreSQL 权限再读 MongoDB
- **删除流程**：PostgreSQL（软删除）→ MongoDB → Etcd 通知

**MongoDB 客户端配置**：

必须配置 `DefaultDocumentM: true`，让驱动将文档解码为 `map[string]any` 而非 `bson.D`：

```go
mongoClient, _ := mongo.Connect(
    options.Client().ApplyURI(uri).SetBSONOptions(&options.BSONOptions{
        DefaultDocumentM: true,
    }),
)
mon.Inject(uri, mongoClient)  // 注入到 go-zero
```

> 原因：积木配置使用 `map[string]any` 存储，若不配置则 MongoDB 返回 `bson.D` 类型导致类型断言失败。

**详细 Schema**：参见 `nexlyn-deploy/postgres-init-scripts/010-lynxgraph-schema.sql`

---

## 🧩 逻辑块系统

### 职能边界：积木 vs 边

**设计哲学**：**边应该"轻"（连接关系），积木应该"重"（业务逻辑）**

| 组件 | 职责 | 示例 |
|------|------|------|
| **边 (Edge)** | 简单布尔条件、流转关系 | `status == "success"` |
| **积木 (Block)** | 复杂业务逻辑、数据转换、外部服务集成、状态管理 | Transform、HTTPCall、Switch、FetchSensorData |

### 逻辑块接口

所有逻辑块必须实现 `core.LogicBlock` 接口：

```go
type LogicBlock interface {
    Execute(ctx context.Context,
            execCtx ExecutionContext,    // 执行上下文（信息原子、图、节点、租户、权限）
            datastore Store,              // 图上下文存储（Redis）
            service Service) (bool, error) // 外部服务注册表
}
```

**Execute 方法约定**：
- **返回值** `(true, nil)`：继续执行后续节点
- **返回值** `(false, nil)`：正常结束，不继续执行
- **返回值** `(_, error)`：异常中断

**已实现的逻辑块**（7 个）：
- Skylark 集成：`skylark_journey_create`
- 基础工具：`log`
- 数据查询：`fetch_sensor_data`
- 数据分析：`statistical_analyzer`、`rate_of_change`、`anomaly_detector`、`trend_analyzer`

**实现位置**：`internal/blocks/skylark/` 和 `internal/blocks/standard/`

### 服务注册机制

**核心原则**：Core 层不定义业务接口，只定义服务类型常量

**集成流程**：
1. **Core 层**：定义服务类型常量（如 `ServiceTypeSkylarkEngine`）
2. **外部系统**：定义业务接口（如 `github.com/rezeropoint/go-skylark/engine.SkylarkEngine`）
3. **Engine 层**：使用 `engine.WithService()` 注册服务实例
4. **积木层**：通过 `service.GetByType()` 获取并类型断言

**预定义服务类型**：
- `ServiceTypeSkylarkEngine` - Skylark 流程引擎
- `ServiceTypeSensorData` - 传感器历史数据查询（通过 IoTQuery gRPC）
- `ServiceTypeWebhook` - Webhook 客户端
- `ServiceTypeKafka` - Kafka 生产者

**示例代码**：参见 `service/lynxengine/internal/svc/servicecontext.go`

---

## 📋 开发规范

### 1. 数据库查询规范

- **可空字段处理**：在 `model.go` 中使用 `sql.NullString`/`sql.NullTime`
- **转换函数**：在 `helpers.go` 中定义 `nullStringToString()` 等转换函数
- **避免临时结构体**：不在 `handler.go` 中定义一次性结构体

### 2. 错误处理规范

- **错误包装**：使用 `fmt.Errorf("%w", err)` 保留错误链
- **错误定义**：在各模块的 `error.go` 中定义领域错误
- **MongoDB 错误**：使用 `monc.ErrNotFound`（不直接使用 `mongo.ErrNoDocuments`）
- **PostgreSQL 错误**：使用 `sql.ErrNoRows`

### 3. Manager 解耦机制

**设计原则**：Internal 子包之间避免直接依赖，使用接口和函数类型解耦

**解耦策略**：
1. **通过 Core 包接口通信**：所有 Manager 依赖 Core 接口，不直接依赖其他 Manager
2. **通过存储层解耦**：Manager 通过共享的存储层（PostgreSQL/MongoDB/Redis/Etcd）间接通信
3. **事件通知机制**：Manager 通过 Etcd 发送配置变更通知，Engine 监听响应

**函数类型示例**（避免循环依赖）：
```go
// core/tag.go
type GetTagNamesByIDsFunc func(ctx context.Context, tagIDs []string, tenantID string) ([]string, error)

// internal/graph/handler.go
type graphRegistry struct {
    getTagNamesFunc core.GetTagNamesByIDsFunc  // 依赖注入，而不是直接依赖 tag.Manager
}
```

### 4. Config 结构体规范

**Config 只能包含**：
- ✅ 配置参数（字符串、数字、时间等基础类型）
- ✅ Etcd/Redis 配置结构（如 `etcdtrigger.EtcdConfig`）

**Config 禁止包含**：
- ❌ 函数类型（如 `GetTagNamesFunc`）
- ❌ 运行时实例（如 `*redis.Client`、`sqlx.SqlConn`）

**原因**：函数和实例必须通过 `NewManager()` 的独立参数传入，保持 Config 的纯粹性

### 5. 数据库连接依赖注入

**注入流程**：
```
外部层（REST 或 Main）
  └─> 创建数据库连接（sqlx.SqlConn、*monc.Model、Redis、Etcd）
      └─> 注入到 Manager/Engine 构造函数
          └─> 传递给 Internal 层 Registry
```

**优势**：
- 连接复用：多个 Manager 共享同一个连接池
- 易于测试：可以注入 Mock 连接
- 配置集中：连接配置统一管理

---

## 🔄 配置管理与热更新

### Etcd 通知机制

**Key 格式**：`{prefix}/{action}/{id}`

**Action 类型**：
- `create` - 创建逻辑图，Engine 自动加载
- `update` - 更新逻辑图，Engine 自动重载
- `delete` - 删除逻辑图，Engine 自动卸载

**流程**：
```
Manager.CreateGraphConfig()
  └─> PostgreSQL 事务 + MongoDB 写入
      └─> Etcd.Put("{prefix}/create/{id}")
          └─> Engine 监听到变更
              └─> GraphRegistry.LoadGraph(id)
```

### 节点和边 ID 自动生成

前端创建/编辑逻辑图时可使用临时 ID（格式：`client:node-xxx`、`client:edge-xxx`），后端自动替换为 UUID：

- **临时 ID 识别**：`isClientTempID()` 检查 `client:` 前缀
- **ID 处理**：`processNodeAndEdgeIDs()` 生成 UUID 并更新边的节点引用
- **自环防护**：校验边的源节点不能等于目标节点

**实现位置**：`internal/graph/helpers.go:120-180`

---

## ⏰ 定时调度机制

### 概述

LynxGraph 支持在逻辑图中配置定时触发积木（`schedule`），实现按 cron 表达式周期性执行逻辑图。

**核心组件**：

| 组件 | 位置 | 说明 |
|------|------|------|
| `ScheduleConfig` | `core/schedule.go` | 定时配置结构（NodeID、CronExpr、Timezone、InitialPayload） |
| `ScheduleRegistry` | `internal/scheduler/` | 定时调度管理器（cron 调度、分布式锁） |
| `ScheduleBlock` | `internal/blocks/standard/schedule/` | 定时触发积木（画布可视化） |

### 多 Schedule 节点精确触发

**设计要点**：一个逻辑图可以包含多个 `schedule` 入口节点，每个节点有独立的 cron 表达式。当某个 schedule 触发时，只执行该节点对应的后续链路，不影响其他 schedule 节点。

**实现机制**：

1. **注册阶段**：`ScheduleConfig.NodeID` 保存触发节点的 ID
2. **触发阶段**：创建 InfoAtom 时将 `schedule_node_id` 写入 labels
3. **分发阶段**：`DispatchScheduled` 根据 `schedule_node_id` 只触发对应节点

```
注册 schedule 任务
  └─> RegisterSchedule(graphKey, nodes)
      └─> 遍历所有 schedule 类型节点
          └─> scheduleConfig.NodeID = nodeID
              └─> cronRunner.Schedule(cron, job)

定时触发
  └─> cron 触发 job
      └─> scheduleTriggerFunc(graphKey, scheduleConfig)
          └─> CreateScheduledInfoAtom() // labels["schedule_node_id"] = NodeID
              └─> DispatchScheduled(graphKey, infoAtom)
                  └─> 只选择 block.GetID() == targetNodeID 的节点
```

**示例配置**（一个图两个定时器）：

```json
{
  "nodes": [
    {"id": "node-morning", "blockType": "schedule", "blockConfig": {"cronExpr": "31 9 * * 1-5"}},
    {"id": "node-evening", "blockType": "schedule", "blockConfig": {"cronExpr": "30 17 * * 1-5"}}
  ]
}
```

- 09:31 触发时：只执行 `node-morning` 后续链路
- 17:30 触发时：只执行 `node-evening` 后续链路

### 分布式锁

多副本部署时，使用 Redis 分布式锁确保同一时刻只有一个实例执行定时任务：

- **锁 Key 格式**：`schedule:{graphID}:{cronExpr}`
- **锁 TTL**：5 分钟（防止异常时锁无法释放）
- **获取失败**：跳过本次执行，等待下次触发

---

## 🛠️ 常见开发任务

### 添加新的逻辑块

1. 在 `internal/blocks/standard/` 创建新包（如 `myblock/`）
2. 实现 `core.LogicBlock` 接口
3. 定义配置结构体（使用 `check:"must"` 标签标记必填字段）
4. 定义 `BlockSpec` 规格（声明服务依赖）
5. 在 `internal/blocks/block/loader.go` 中注册

**参考实现**：`internal/blocks/standard/log/log.go`

### 添加新的服务类型

1. 在 `core/service.go` 的 `const` 块添加服务类型常量
2. 在 Engine 创建时使用 `engine.WithService()` 注册服务实例
3. 更新本文档的服务类型列表

### 扩展 ExecutionContext

1. 在 `core/context.go` 接口中添加新方法
2. 在 `engine/context.go` 中实现 `BaseExecutionContext`
3. 确保所有积木可以通过新方法访问所需信息

---

## 🔍 调试指南

### 查看 Etcd 配置变更

```bash
# 查看配置变更通知
etcdctl get --prefix lynxgraph/

# 监听配置变更
etcdctl watch --prefix lynxgraph/
```

### 查看 Redis 缓存

```bash
# 查看图上下文
redis-cli GET "lynxgraph:context:{graph_id}"

# 查看信息原子缓存
redis-cli GET "lynxgraph:infoatom:{type_id}"
```

### 查询 MongoDB 数据

```javascript
// 查看逻辑图配置
db.graph.find({"graph_id": "xxx"})

// 查看节点配置
db.graph.aggregate([
  {$match: {"graph_id": "xxx"}},
  {$project: {"nodes": 1}}
])
```

---

## 📚 关键文件索引

| 分类 | 文件路径 | 说明 |
|------|----------|------|
| **核心接口** | `core/infoatom.go` | 信息原子类型接口 |
| | `core/logic.go` | 逻辑图接口 |
| | `core/block.go` | 逻辑块接口和配置工具 |
| | `core/context.go` | 执行上下文接口 |
| | `core/service.go` | 外部服务注册接口 |
| | `core/tag.go` | 标签相关类型和函数类型 |
| **Manager 实现** | `internal/graph/handler.go` | 逻辑图 CRUD 实现 |
| | `internal/infoatom/handler.go` | 信息原子类型 CRUD 实现 |
| | `internal/tag/handler.go` | 标签 CRUD 实现 |
| | `internal/blocks/block/handler.go` | 逻辑块注册表 |
| | `internal/dispatcher/handler.go` | 信息原子调度器 |
| | `internal/scheduler/handler.go` | 定时任务调度器 |
| | `internal/scheduler/helpers.go` | 定时调度辅助函数 |
| **对外接口** | `manager/manager.go` | Manager 模式接口定义 |
| | `engine/engine.go` | Engine 模式接口定义 |
| **逻辑块** | `internal/blocks/standard/` | 标准逻辑块实现 |
| | `internal/blocks/skylark/` | Skylark 集成逻辑块 |
| **配置** | `manager/config.go` | Manager 配置结构 |
| | `engine/config.go` | Engine 配置结构 |

---

## 📖 相关文档

- **API 实现**：`restful/lynxmanager/`（待实现）
- **数据库 Schema**：
  - `nexlyn-deploy/postgres-init-scripts/010-lynxgraph-schema.sql`
  - `nexlyn-deploy/mongodb-init-scripts/`（MongoDB 初始化）
- **重构状态**：`REFACTORING_STATUS.md`（Phase 1/2/3 进度跟踪）
- **逻辑块路线图**：`internal/blocks/ROADMAP.md`（积木开发计划）
- **项目总体架构**：`/CLAUDE.md`

---

**文档版本**：v2.2
**最后更新**：2025-12-29
**修订说明**：新增定时调度机制章节，说明多 schedule 节点精确触发设计
**维护者**：LynxGraph 引擎开发团队
