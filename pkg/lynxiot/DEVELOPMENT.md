# IoT 引擎开发指南

> 面向 IoT 引擎内部开发者的规范和最佳实践。

---

## 📑 快速导航

| 场景 | 跳转 |
|------|------|
| 创建新 Manager | [Manager 规范](#1-manager-包规范) |
| 添加新设备类别 | [添加设备类别](#添加新设备类别) |
| 添加新字段 | [添加字段](#添加新字段) |
| 扩展平台集成 | [扩展平台](#扩展平台集成) |
| 调试问题 | [调试指南](#-调试指南) |
| 理解架构 | [架构概览](#-架构概览) |

---

## 📐 架构设计理念

### 1. 四层架构与依赖规则

IoT 引擎采用**清晰分层架构**,严格遵循依赖倒置原则(DIP):

```
┌─────────────────────────────────────────────┐
│  REST Layer (restful/iotmanager/)          │  ← 外部接口层
├─────────────────────────────────────────────┤
│  Engine Layer (engine/)                    │  ← 公共 API 层
├─────────────────────────────────────────────┤
│  Internal Layer (internal/)                │  ← 内部实现层
├─────────────────────────────────────────────┤
│  Core Layer (core/)                        │  ← 核心领域层(最内层)
└─────────────────────────────────────────────┘

依赖方向:外层 → 内层(Core 层不依赖任何外层)
```

#### **Core 层的黄金法则**

**Core 层应该是什么**:
- ✅ **纯粹的领域模型**:设备类别、模板配置、设备绑定等实体
- ✅ **业务规则的载体**:领域逻辑(如设备注册机制、在线状态管理)
- ✅ **稳定的抽象**:接口定义(如 `DeviceCategory`、`SensorTemplate`、`Platform`)
- ✅ **框架无关**:不依赖任何外部框架(go-zero、PostgreSQL、Redis)

**Core 层不应该是什么**:
- ❌ **不是 DTO**:不应包含"用于前端展示"的字段
- ❌ **不是 ORM 模型**:不应出现 `db`、`gorm`、`sqlx` 等存储标签
- ❌ **不是 REST 响应体**:不应为了方便 API 返回而污染领域模型
- ❌ **不包含存储实现**:Redis 键构建、Etcd 路径等应在工具函数中

**参考实现**:
- ✅ 好的例子:`core/device_binding.go`(纯领域模型,无框架依赖)
- ✅ 好的例子:`core/template_metadata.go`(清晰的值对象)
- ✅ 好的例子:`core/platform.go`(元数据与配置分离设计)

#### **Engine 层的职责边界**

Engine 层是对外暴露的公共 API,负责业务编排:

**职责**:
- ✅ **返回纯 Core 对象**:所有查询方法返回领域模型
- ✅ **业务逻辑编排**:协调多个 Internal Manager 完成复杂操作
- ✅ **Manager 生命周期**:初始化、依赖注入、销毁管理
- ✅ **权限验证**:验证租户和组织权限(通过 TenantID、OrgIDs 参数)

**禁止**:
- ❌ **不返回 DTO**:不负责填充展示字段
- ❌ **不处理 HTTP**:不依赖 HTTP 协议和 REST 框架
- ❌ **不做格式转换**:不负责时间戳格式、单位转换

#### **REST 层的职责边界**

REST 层负责 HTTP 接口和数据转换:

**职责**:
- ✅ **JWT 认证**:解析 Token 获取用户信息
- ✅ **权限查询**:查询用户可见的组织树
- ✅ **参数验证**:验证请求参数的合法性
- ✅ **创建数据库连接**:初始化并验证连接可用性
- ✅ **错误码映射**:将引擎错误转换为 HTTP 状态码

**禁止**:
- ❌ **不包含业务逻辑**:不直接操作数据库
- ❌ **不处理存储**:所有存储操作通过引擎完成

---

## 📋 开发规范

### 1. Manager 包规范

每个 Manager 包遵循统一结构：

```
internal/<name>/
├── <name>.go      # 接口 + NewManager 工厂函数
├── handler.go     # 实现 + newXxxManager 私有构造
├── internal.go    # 私有方法（有 receiver）
├── helpers.go     # 工具函数（无 receiver）
├── model.go       # 数据库模型（sql.Null* 类型）
└── config.go      # 配置结构
```

#### 检查清单

| # | 检查项 | 要求 |
|---|--------|------|
| 1 | 接口文件 | `<name>.go` 定义 `Manager` 接口 |
| 2 | 公开工厂 | `<name>.go` 定义 `NewManager(...) (Manager, error)` |
| 3 | 私有构造 | `handler.go` 定义 `new<Name>Manager(...) (*xxxManager, error)` |
| 4 | 返回类型 | 公开返回接口，私有返回具体类型 |
| 5 | 参数验证 | 私有构造验证必要依赖（如 `configStore == nil`） |
| 6 | 结构体 | 实现结构体小写（私有）：`type xxxManager struct` |

**标准模式**：
```go
// <name>.go
func NewManager(dbConn sqlx.SqlConn, config Config) (Manager, error) {
    return newXxxManager(dbConn, config)
}

// handler.go
func newXxxManager(dbConn sqlx.SqlConn, config Config) (*xxxManager, error) {
    if dbConn == nil {
        return nil, fmt.Errorf("dbConn 未设置")
    }
    return &xxxManager{dbConn: dbConn, config: config}, nil
}
```

### 2. 数据库规范

| 场景 | 规范 |
|------|------|
| 可空字段 | `sql.NullString` / `sql.NullTime` |
| 数组字段 | `pq.Array()` 包装 |
| 事务 | `TransactCtx` 确保原子性 |
| 事务内辅助函数 | 接收 `sqlx.Session` 参数 |
| **PostgreSQL + Etcd 双写** | **Etcd 操作必须在事务闭包内，任意失败自动回滚** |

#### PostgreSQL + Etcd 事务处理

**核心原则**：将 Etcd 操作放入 `TransactCtx` 闭包内，利用 go-zero 事务机制确保一致性。

**标准模式**：
```go
// ✅ 正确：Etcd 操作在事务闭包内
etcdKey := BuildKey(id)
err := m.dbConn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
    // 1. 写 PostgreSQL
    _, err := session.ExecCtx(ctx, insertSQL, ...)
    if err != nil {
        return err  // 失败会回滚事务
    }

    // 2. 写 Etcd（在闭包内）
    if err := m.configStore.PutConfig(ctx, etcdKey, config); err != nil {
        return err  // 失败会触发 PostgreSQL 回滚
    }

    return nil  // 成功则提交事务
})
```

**错误模式**：
```go
// ❌ 错误：Etcd 在事务外操作
err := m.dbConn.TransactCtx(ctx, func(...) error {
    session.ExecCtx(ctx, insertSQL, ...)  // PostgreSQL 已提交
    return nil
})
m.configStore.PutConfig(ctx, etcdKey, config)  // Etcd 失败无法回滚
```

**极低概率边界情况**：
- Etcd 写入成功 → 事务提交失败 → PostgreSQL 回滚成功 → **Etcd 孤立配置**
- **影响**：可忽略（业务逻辑以 PostgreSQL 为准，孤立配置不会被使用）
- **处理**：可选的定期清理任务（非必需）

### 3. Manager 解耦

Internal 子包之间**禁止直接依赖**，通过以下方式解耦：

```go
// core/platform.go - 定义函数类型
type GetPlatformConfigFunc func(ctx context.Context, tenantID, platformID string) (*PlatformConfig, error)

// internal/template/template.go - 通过构造函数注入
func NewManager(db sqlx.SqlConn, getPlatformConfig core.GetPlatformConfigFunc) (Manager, error)
```

### 4. 错误处理

- 包装错误：`fmt.Errorf("xxx失败: %w", err)`
- 领域错误定义在 `core/errors.go`
- 保留完整错误链

### 5. 安全规范

| 规则 | 说明 |
|------|------|
| TenantID 来源 | **必须**从 JWT 解析，禁止前端提供 |
| 组织权限 | SQL 层强制 `org_id = ANY($X)` 约束 |
| 绑定操作 | 使用用户默认组织，防止越权 |

---

## 🔐 权限与多租户机制

### 三层安全架构

```
JWT Token(REST 层)
  └─> 解析 TenantID + UserID + PrimaryOrgId
      └─> 查询组织树(获取所有子组织 ID)
          └─> 传递 OrgIDs 到 Engine 层
              └─> Manager 层 SQL 约束(WHERE org_id = ANY($X))
```

### 权限规则矩阵

| 操作 | TenantID 来源 | OrgID 处理 | SQL 约束 |
|------|--------------|-----------|----------|
| **查询/修改/删除** | JWT | 前端提供 → 查询组织树 | `org_id = ANY($X)` |
| **绑定设备** | JWT | 使用用户默认组织 | `org_id = $1` |
| **未绑定设备查询** | JWT | 不需要(租户级) | `tenant_id = $1` |
| **模板操作** | JWT | 不需要(租户级) | `tenant_id = $1` |

**关键原则**:
- ⚠️ **TenantID 永远不由前端提供**,必须从 JWT 解析
- ⚠️ **绑定操作使用默认组织**,防止越权绑定
- ⚠️ **SQL 层强制权限约束**,即使上层失效也能保证安全

---

## 📡 MQTT 数据处理流程

### 完整处理链路

```
设备 → MQTT Broker → MQTT Manager
  ├─> 1. 时间窗口去重(Redis 分布式锁)
  ├─> 2. 在线检测(更新 Redis TTL 300s)
  ├─> 3. 业务数据提取(字段映射 + Scale/Offset)
  ├─> 4. 时间戳解析(支持 10+ 格式)
  ├─> 5. 过滤规则(9 种操作符)
  ├─> 6. 存储时序数据(ClickHouse)
  └─> 7. 异步分发(Dispatcher → 多平台)
```

### 关键机制详解

#### 分布式锁机制

**目的**:多副本部署时避免重复处理

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `DedupWindowMs` | 100ms | 去重时间窗口 |
| `LockTTLSeconds` | 30s | 锁过期时间 |

**锁 Key 格式**:`iot:mqtt:lock:{topic}:{time_window_ms}`

#### 主题格式规范

```
nexlyn/iot/{category}/{model}/{device_id}/{topicSuffix}
```

**示例**:`nexlyn/iot/ai_box/AI-200/box001/board_ping`

#### 配置查询优化

通过主题信息直接构建 Etcd key(**O(1)** 查询):
- 在线检测:`sensor-online/{category}/{model}/{suffix}`
- 业务数据:`sensor-data/{category}/{model}/{suffix}`

---

## 🛠️ 常见开发任务

### 添加新设备类别

```bash
# 1. 创建设备文件
touch core/devices/new_device.go
```

```go
// core/devices/new_device.go
package devices

import "github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

const CategoryNewDevice core.DeviceCategory = "new_device"

func init() {
    core.RegisterCategory(CategoryNewDevice, core.CategoryInfo{
        Name:        "新设备",
        Description: "设备描述",
        Fields:      []string{fields.Temperature, fields.Humidity},
    })
}
```

完成后自动注册生效。

### 添加新字段

1. `core/fields/` 创建字段定义
2. `core/fields/registry.go` 添加字段名常量
3. ClickHouse 创建对应表 `iot_{field}_data`
4. 设备定义中引用新字段

### 扩展平台集成

1. `internal/dispatcher/` 添加平台实现
2. `core/dispatcher.go` 添加类型常量
3. 实现 `Dispatch(ctx, data, config)` 方法

---

## 📐 架构概览

### 分层架构

```
REST Layer (restful/iotmanager/)     → HTTP 接口、JWT 认证、错误码映射
Engine Layer (engine/)               → 公共 API、业务编排、Manager 生命周期
Internal Layer (internal/)           → Manager 实现（见下表）
Core Layer (core/)                   → 领域模型、接口定义（无框架依赖）
```

**依赖方向**：外层 → 内层（Core 层不依赖任何外层）

### Manager 一览

| Manager | 职责 | 存储 |
|---------|------|------|
| Template | 传感器模板 CRUD | PostgreSQL + Etcd |
| Device | 设备绑定管理 | PostgreSQL + Redis |
| MQTT | 消息处理、在线检测 | Redis + Etcd |
| Storage | 时序数据写入 | ClickHouse |
| Query | 时序数据查询 | PostgreSQL + ClickHouse |
| Platform | 平台配置 | PostgreSQL + Etcd |
| Dispatcher | 数据分发 | 内存队列 |
| Controller | 设备控制 | MQTT + Redis |
| Tag | 标签管理 | PostgreSQL |
| HttpReceive | HTTP 数据接收 | PostgreSQL + Etcd |

### 存储职责

| 存储 | 职责 | 示例 Key/表 |
|------|------|-------------|
| PostgreSQL | 元数据、索引、事务 | `iot_sensor_templates` |
| ClickHouse | 时序数据 | `iot_temperature_data` |
| Etcd | 运行时配置 | `sensor-online/{category}/{model}/{suffix}` |
| Redis | 状态缓存、分布式锁 | `iot:device:online:{id}` |

### 混合存储模式

**创建流程（事务保护）**:
```
TransactCtx 闭包开始
  ├─> 写入 PostgreSQL 元数据（session.ExecCtx）
  │   └─> 失败 → 返回 error → 事务回滚 → 结束
  │
  ├─> 写入 Etcd 配置（configStore.PutConfig）
  │   └─> 失败 → 返回 error → 触发 PostgreSQL 回滚 → 结束
  │
  └─> 返回 nil → 事务提交 → 成功
```

**更新/删除流程**:
```
TransactCtx 闭包开始
  ├─> 验证权限（session.QueryRowCtx）
  ├─> 更新/删除 PostgreSQL 记录（session.ExecCtx）
  ├─> 更新/删除 Etcd 配置（configStore.PutConfig/DeleteConfig）
  │   └─> 任意步骤失败 → 事务回滚
  └─> 返回 nil → 事务提交
```

**查询优化**:
- **List**：只查 PostgreSQL（快速索引，返回摘要）
- **Get**：先验证 PostgreSQL 权限，再读 Etcd 完整配置
- **运行时读取**：优先读 Etcd（热配置），不存在则降级查 PostgreSQL

**数据一致性原则**:
- **PostgreSQL 为权威数据源**：所有查询和权限验证以 PostgreSQL 为准
- **Etcd 为运行时缓存**：提供高性能配置读取，可短暂缺失
- **容忍 Etcd 孤立配置**：极低概率出现，业务逻辑不会使用（无对应 PostgreSQL 记录）
- **不容忍 PostgreSQL 孤立记录**：事务机制确保不会发生

**职责分离原则**:
- **PostgreSQL** = 关系数据 + 事务 + 索引 + 权威元数据
- **ClickHouse** = 时序数据 + 列存储 + 自动分区
- **Etcd** = 配置热更新 + 分布式共享 + 运行时缓存
- **Redis** = 状态缓存 + TTL + 分布式锁

### 权限流程

```
JWT → TenantID + UserID → 查询组织树 → OrgIDs → SQL 约束 (org_id = ANY($X))
```

---

## 🔍 调试指南

### Etcd 配置

```bash
etcdctl get --prefix sensor-online/    # 在线检测配置
etcdctl get --prefix sensor-data/      # 业务数据配置
etcdctl get --prefix platform-         # 平台配置
etcdctl get --prefix http-receive/     # HTTP 接收配置
```

### Redis 状态

```bash
redis-cli KEYS "iot:device:online:*"                      # 在线设备
redis-cli SMEMBERS "iot:mqtt:subscribed:{pod}:topics"     # 已订阅主题
redis-cli KEYS "iot:mqtt:lock:*"                          # 消息锁
```

### ClickHouse 数据

```sql
-- 最新数据
SELECT * FROM iot_temperature_data WHERE device_id = 'xxx' ORDER BY timestamp DESC LIMIT 10;

-- 分区信息
SELECT partition, name, rows FROM system.parts WHERE table = 'iot_temperature_data';
```

---

## 📚 关键文件索引

| 分类 | 路径 |
|------|------|
| 领域模型 | `core/device_category.go`, `core/template_*.go`, `core/platform.go` |
| 设备定义 | `core/devices/`, `core/fields/` |
| 引擎接口 | `engine/engine.go`, `engine/config.go` |
| 数据库 Schema | `nexlyn-deploy/postgres-init-scripts/005-*.sql`, `006-*.sql` |

---

**版本**：v3.0 | **更新**：2025-01-25 | **变更**：重构文档结构，突出开发规范和常见任务
