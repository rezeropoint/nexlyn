# IoT 引擎开发指南

> **文档定位**：本文档面向 IoT 引擎的**内部开发者**，阐述架构设计理念、开发规范和最佳实践。
> **使用说明**：API 使用请参考 `README.md`（待创建）。

---

## 📐 架构设计理念

### 1. 三层架构与依赖规则

IoT 引擎采用**清晰分层架构**，严格遵循依赖倒置原则（DIP）：

```
┌─────────────────────────────────────────────┐
│  REST Layer (restful/iotmanager/)          │  ← 外部接口层
├─────────────────────────────────────────────┤
│  Engine Layer (engine/)                    │  ← 公共 API 层
├─────────────────────────────────────────────┤
│  Internal Layer (internal/)                │  ← 内部实现层
├─────────────────────────────────────────────┤
│  Core Layer (core/)                        │  ← 核心领域层（最内层）
└─────────────────────────────────────────────┘

依赖方向：外层 → 内层（Core 层不依赖任何外层）
```

#### **Core 层的黄金法则**

**Core 层应该是什么**：
- ✅ **纯粹的领域模型**：设备类别、模板配置、设备绑定等实体
- ✅ **业务规则的载体**：领域逻辑（如设备注册机制、在线状态管理）
- ✅ **稳定的抽象**：接口定义（如 `DeviceCategory`、`SensorTemplate`、`Platform`）
- ✅ **框架无关**：不依赖任何外部框架（go-zero、PostgreSQL、Redis）

**Core 层不应该是什么**：
- ❌ **不是 DTO**：不应包含"用于前端展示"的字段
- ❌ **不是 ORM 模型**：不应出现 `db`、`gorm`、`sqlx` 等存储标签
- ❌ **不是 REST 响应体**：不应为了方便 API 返回而污染领域模型
- ❌ **不包含存储实现**：Redis 键构建、Etcd 路径等应在工具函数中

**参考实现**：
- ✅ 好的例子：`core/device_binding.go`（纯领域模型，无框架依赖）
- ✅ 好的例子：`core/template_metadata.go`（清晰的值对象）
- ✅ 好的例子：`core/platform.go`（元数据与配置分离设计）

#### **Engine 层的职责边界**

Engine 层是对外暴露的公共 API，负责业务编排：

**职责**：
- ✅ **返回纯 Core 对象**：所有查询方法返回领域模型
- ✅ **业务逻辑编排**：协调多个 Internal Manager 完成复杂操作
- ✅ **Manager 生命周期**：初始化、依赖注入、销毁管理
- ✅ **权限验证**：验证租户和组织权限（通过 TenantID、OrgIDs 参数）

**禁止**：
- ❌ **不返回 DTO**：不负责填充展示字段
- ❌ **不处理 HTTP**：不依赖 HTTP 协议和 REST 框架
- ❌ **不做格式转换**：不负责时间戳格式、单位转换

#### **Internal 层的职责边界**

Internal 层负责具体实现，采用 **Manager 模式**：

| Manager | 职责 | 存储方案 | 关键文件 |
|---------|------|----------|----------|
| **Template Manager** | 传感器模板管理 | PostgreSQL + Etcd | `internal/template/` |
| **Device Manager** | 设备绑定管理 | PostgreSQL + Redis | `internal/device/` |
| **MQTT Manager** | 消息处理 | Redis + Etcd | `internal/mqtt/` |
| **Storage Manager** | 时序数据存储 | ClickHouse | `internal/storage/` |
| **Query Manager** | 时序数据查询 | PostgreSQL + ClickHouse | `internal/query/` |
| **Platform Manager** | 平台配置管理 | PostgreSQL + Etcd | `internal/platform/` |
| **Dispatcher Manager** | 数据分发调度 | 内存队列 | `internal/dispatcher/` |
| **Controller Manager** | 设备控制 | MQTT + Redis | `internal/controller/` |
| **Tag Manager** | 标签管理 | PostgreSQL | `internal/tag/` |
| **Metadata Manager** | 元数据查询 | 内存注册表 | `internal/metadata/` |

**文件组织规范**（所有 Manager 必须遵守）：
- `<manager>.go` - 接口定义和核心类型
- `handler.go` - 接口实现和公共业务逻辑
- `internal.go` - 内部方法（有 receiver）和私有业务逻辑
- `helpers.go` - 纯函数辅助方法（无 receiver）
- `model.go` - 数据库查询结构体（使用 `sql.Null*` 类型）
- `config.go` - Manager 配置结构

#### **REST 层的职责边界**

REST 层负责 HTTP 接口和数据转换：

**职责**：
- ✅ **JWT 认证**：解析 Token 获取用户信息
- ✅ **权限查询**：查询用户可见的组织树
- ✅ **参数验证**：验证请求参数的合法性
- ✅ **创建数据库连接**：初始化并验证连接可用性
- ✅ **错误码映射**：将引擎错误转换为 HTTP 状态码

**禁止**：
- ❌ **不包含业务逻辑**：不直接操作数据库
- ❌ **不处理存储**：所有存储操作通过引擎完成

---

## 🎯 核心功能与概念

### 核心功能清单

| 功能模块 | 说明 | 状态 |
|---------|------|------|
| **设备模板管理** | 模板 CRUD、双存储架构、配置热更新 | ✅ 完成 |
| **设备绑定管理** | 设备绑定/解绑、组织权限、在线状态 | ✅ 完成 |
| **MQTT 数据处理** | 在线检测、业务数据提取、分布式锁 | ✅ 完成 |
| **时序数据存储** | ClickHouse 按字段分表、批量写入、TTL | ✅ 完成 |
| **时序数据查询** | 混合查询、6 种聚合类型、权限约束 | ✅ 完成 |
| **平台集成** | Skylark、LynxGraph、Webhook、Kafka | 🚧 部分完成 |
| **设备控制** | AI Box 控制协议、并发控制、在线检查 | ✅ 完成 |
| **标签系统** | 标签 CRUD、模板/设备关联 | ✅ 完成 |

### 核心概念

- **设备类别 (DeviceCategory)**：设备的分类定义，如温湿度传感器、AI Box
- **设备模板 (SensorTemplate)**：定义设备的数据采集和处理规则
- **设备绑定 (DeviceBinding)**：将物理设备绑定到组织和模板
- **在线检测配置 (OnlineDetectionConfig)**：定义如何判断设备在线
- **业务数据配置 (DataProcessingConfig)**：定义如何提取和处理业务数据
- **平台配置 (Platform)**：外部平台集成配置（Skylark、Webhook 等）
- **数据分发 (Dispatch)**：将处理后的数据分发到多个平台

### 设备注册机制

```
1. core/device_category.go     - 定义设备类别类型
2. core/devices/*.go           - 每个设备一个文件
3. init() 调用 RegisterCategory() - 自动注册
4. engine/handler.go           - import 触发注册
```

**添加新设备**：在 `core/devices/` 创建文件，定义常量和标准字段，自动注册生效。

---

## 🔐 权限与多租户机制

### 三层安全架构

```
JWT Token（REST 层）
  └─> 解析 TenantID + UserID + PrimaryOrgId
      └─> 查询组织树（获取所有子组织 ID）
          └─> 传递 OrgIDs 到 Engine 层
              └─> Manager 层 SQL 约束（WHERE org_id = ANY($X)）
```

### 权限规则矩阵

| 操作 | TenantID 来源 | OrgID 处理 | SQL 约束 |
|------|--------------|-----------|----------|
| **查询/修改/删除** | JWT | 前端提供 → 查询组织树 | `org_id = ANY($X)` |
| **绑定设备** | JWT | 使用用户默认组织 | `org_id = $1` |
| **未绑定设备查询** | JWT | 不需要（租户级） | `tenant_id = $1` |
| **模板操作** | JWT | 不需要（租户级） | `tenant_id = $1` |

**关键原则**：
- ⚠️ **TenantID 永远不由前端提供**，必须从 JWT 解析
- ⚠️ **绑定操作使用默认组织**，防止越权绑定
- ⚠️ **SQL 层强制权限约束**，即使上层失效也能保证安全

---

## 🏗️ 数据存储架构

### 四层存储策略

| 存储层 | 用途 | 数据类型 | 关键表/键 |
|--------|------|----------|----------|
| **PostgreSQL** | 元数据和索引 | 模板、设备绑定、标签、平台元数据 | `iot_sensor_templates`、`iot_device_bindings` |
| **ClickHouse** | 时序数据 | 传感器数据（温度、湿度等） | `iot_temperature_data`、`iot_humidity_data` |
| **Etcd** | 运行时配置 | MQTT 配置、平台敏感信息 | `sensor-online/`、`sensor-data/`、`platform-` |
| **Redis** | 运行时状态 | 在线状态、分布式锁、订阅主题 | `iot:device:online:{id}`、`iot:mqtt:lock:{topic}` |

### 混合存储模式

**创建流程**：
```
PostgreSQL（事务开始）
  └─> 写入元数据
      └─> Etcd（写入配置）
          └─> PostgreSQL（事务提交/回滚）
```

**查询优化**：
- List 只查 PostgreSQL（快速索引）
- Get 先验证 PostgreSQL 权限，再读 Etcd 完整配置

**职责分离原则**：
- **PostgreSQL** = 关系数据 + 事务 + 索引
- **ClickHouse** = 时序数据 + 列存储 + 自动分区
- **Etcd** = 配置热更新 + 分布式共享
- **Redis** = 状态缓存 + TTL + 分布式锁

### ClickHouse 表设计

| 字段 | 类型 | 说明 |
|------|------|------|
| `timestamp` | DateTime64(3) | 毫秒精度时间戳 |
| `device_id` | String | 设备 ID |
| `device_model` | String | 设备型号 |
| `device_category` | String | 设备类别 |
| `tenant_id` | String | 租户 ID |
| `value` | Float64 | 数值 |

**优化策略**：
- 按月分区：`PARTITION BY toYYYYMM(timestamp)`
- 180 天 TTL：`TTL timestamp + INTERVAL 180 DAY`
- 布隆过滤器索引：`INDEX idx_device TYPE bloom_filter`

---

## 📡 MQTT 数据处理流程

### 完整处理链路

```
设备 → MQTT Broker → MQTT Manager
  ├─> 1. 时间窗口去重（Redis 分布式锁）
  ├─> 2. 在线检测（更新 Redis TTL 300s）
  ├─> 3. 业务数据提取（字段映射 + Scale/Offset）
  ├─> 4. 时间戳解析（支持 10+ 格式）
  ├─> 5. 过滤规则（9 种操作符）
  ├─> 6. 存储时序数据（ClickHouse）
  └─> 7. 异步分发（Dispatcher → 多平台）
```

### 关键机制详解

#### 分布式锁机制

**目的**：多副本部署时避免重复处理

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `DedupWindowMs` | 100ms | 去重时间窗口 |
| `LockTTLSeconds` | 30s | 锁过期时间 |

**锁 Key 格式**：`iot:mqtt:lock:{topic}:{time_window_ms}`

#### 主题格式规范

```
nexlyn/iot/{category}/{model}/{device_id}/{topicSuffix}
```

**示例**：`nexlyn/iot/ai_box/AI-200/box001/board_ping`

#### 配置查询优化

通过主题信息直接构建 Etcd key（**O(1)** 查询）：
- 在线检测：`sensor-online/{category}/{model}/{suffix}`
- 业务数据：`sensor-data/{category}/{model}/{suffix}`

---

## 🎮 设备控制架构

### AI Box 控制特性

| 功能 | 实现方式 | 说明 |
|------|----------|------|
| **并发控制** | Redis 锁 + sync.Map | 单设备单命令 |
| **在线检查** | 必须检查 | 离线返回 `ErrDeviceOffline` |
| **超时处理** | 可配置 | 默认 30 秒 |
| **响应匹配** | Event 字段 | 确保响应对应正确命令 |

### 控制流程

```
控制请求
  └─> 检查设备在线状态（Redis）
      └─> 获取分布式锁（防并发）
          └─> 发送 MQTT 命令
              └─> 等待响应（sync.Map）
                  └─> 释放锁
```

### Redis 键设计

| 键模式 | TTL | 用途 |
|--------|-----|------|
| `iot:control:device:{id}` | 60s | 命令追踪 |
| `iot:control:lock:{id}` | 60s | 并发锁 |

---

## 🔄 平台集成架构

### 已支持平台

| 平台类型 | 标识 | 功能 | 状态 |
|----------|------|------|------|
| **Skylark** | `skylark_flows` | 流程触发 | ✅ 完成 |
| **Skylark** | `skylark_forms` | 表单提交 | ✅ 完成 |
| **LynxGraph** | `lynxgraph` | 逻辑引擎 | ✅ 完成 |
| **Webhook** | `webhook` | HTTP 回调 | 🚧 开发中 |
| **Kafka** | `kafka` | 消息队列 | 🚧 开发中 |
| **Log** | `log` | 结构化日志 | ✅ 完成 |

### 分发配置示例

**LynxGraph 分发**：
```json
{
  "dispatchConfigs": [{
    "type": "lynxgraph",
    "infoAtomTypeID": "temperature-alarm",
    "fieldMapping": {
      "temperature": "temp",
      "device_id": "source"
    }
  }]
}
```

**Skylark 分发**：
```json
{
  "dispatchConfigs": [{
    "type": "skylark_flows",
    "platformID": "platform-001",
    "flowID": 123
  }]
}
```

---

## 📋 开发规范

### 1. Manager 文件组织规范

每个 Manager 包必须遵循统一的文件结构：

```
internal/template/
├── template.go     # 接口定义
├── handler.go      # 接口实现
├── internal.go     # 私有方法
├── helpers.go      # 工具函数
├── model.go        # 数据模型
└── config.go       # 配置结构
```

### 2. 数据库查询规范

- **可空字段**：使用 `sql.NullString`/`sql.NullTime`
- **数组字段**：使用 `pq.Array()` 包装
- **事务处理**：使用 `TransactCtx` 确保原子性
- **辅助函数**：接收 `sqlx.Session` 在事务中执行

### 3. Manager 解耦机制

**设计原则**：Internal 子包之间避免直接依赖

**解耦策略**：
1. **通过 Core 包接口**：定义公共接口和函数类型
2. **通过存储层**：Manager 间通过存储间接通信
3. **函数类型注入**：避免循环依赖

**示例**：
```go
// core/platform.go
type GetPlatformConfigFunc func(ctx context.Context,
    tenantID, platformID string) (*PlatformConfig, error)

// Template Manager 通过构造函数接收
func NewTemplateManager(db *sqlx.DB, etcd *clientv3.Client,
    getPlatformConfig GetPlatformConfigFunc) TemplateManager {
    // ...
}
```

### 4. Config 规范

**Config 只能包含**：
- ✅ 配置参数（基础类型）
- ✅ 配置结构（如 Redis 配置）

**Config 禁止包含**：
- ❌ 函数类型
- ❌ 运行时实例（数据库连接等）

### 5. 错误处理规范

- **错误包装**：使用 `fmt.Errorf("%w", err)`
- **领域错误**：在 `core/errors.go` 定义
- **错误传递**：保留完整错误链

### 6. 安全规范

1. **TenantID 管理**：
   - 必须从 JWT 解析，不由前端提供
   - REST 层填充到配置对象

2. **组织权限**：
   - 查询操作需要 OrgID 参数
   - 绑定操作使用默认组织
   - SQL 层强制约束

3. **配置验证**：
   - REST 层验证 JWT
   - Manager 层验证配置完整性
   - MQTT 层使用验证后的配置

---

## 🛠️ 常见开发任务

### 添加新字段

1. 在 `core/fields/` 创建字段定义文件
2. 在 `fields/registry.go` 添加字段名常量
3. 在 ClickHouse 创建对应表
4. 在设备定义中引用新字段
5. 设置字段属性（Required、MinValue 等）

### 添加新设备类别

1. 在 `core/devices/` 创建新文件
2. 定义设备常量和标准字段
3. 在 `init()` 中调用 `RegisterCategory()`
4. 自动注册生效，无需其他修改

### 扩展平台集成

1. 在 `internal/dispatcher/` 添加平台实现
2. 实现分发逻辑
3. 在 `core/dispatcher.go` 添加类型常量
4. 更新配置示例

---

## 📚 关键文件索引

| 分类 | 文件路径 | 说明 |
|------|----------|------|
| **核心领域** | `core/device_category.go` | 设备类别定义 |
| | `core/template_*.go` | 模板配置模型 |
| | `core/platform.go` | 平台配置模型 |
| | `core/errors.go` | 错误定义 |
| **设备定义** | `core/devices/` | 预定义设备 |
| | `core/fields/` | 标准字段定义 |
| **引擎接口** | `engine/engine.go` | 公共 API |
| | `engine/config.go` | 引擎配置 |
| **Manager 实现** | `internal/template/` | 模板管理 |
| | `internal/device/` | 设备管理 |
| | `internal/mqtt/` | MQTT 处理 |
| | `internal/controller/` | 设备控制 |

---

## 🔍 调试指南

### 查看 Etcd 配置

```bash
# 在线检测配置
etcdctl get --prefix sensor-online/

# 业务数据配置
etcdctl get --prefix sensor-data/

# 平台配置
etcdctl get --prefix platform-
```

### 查看 Redis 状态

```bash
# 在线设备
redis-cli KEYS "iot:device:online:*"

# 已订阅主题
redis-cli SMEMBERS "iot:mqtt:subscribed:{pod_name}:topics"

# 消息锁
redis-cli KEYS "iot:mqtt:lock:*"
```

### 查询 ClickHouse 数据

```sql
-- 查看最新数据
SELECT * FROM iot_temperature_data
WHERE device_id = 'xxx'
ORDER BY timestamp DESC
LIMIT 10;

-- 查看分区信息
SELECT partition, name, rows
FROM system.parts
WHERE table = 'iot_temperature_data';
```

---

## 📖 相关文档

- **API 实现**：`restful/iotmanager/`
- **数据库 Schema**：
  - `nexlyn-deploy/postgres-init-scripts/005-iot-devices-schema.sql`
  - `nexlyn-deploy/postgres-init-scripts/006-iot-platform-configs-schema.sql`
  - `nexlyn-deploy/clickhouse-init-scripts/001-iot-timeseries-schema.sql`
- **MQTT 控制计划**：`MQTT_CONTROL_PLAN.md`
- **项目总体架构**：`/CLAUDE.md`

---

**文档版本**：v2.0
**最后更新**：2025-01-21
**修订说明**：统一文档格式，采用表格化展示，增加 emoji 图标提升可读性
**维护者**：IoT 引擎开发团队