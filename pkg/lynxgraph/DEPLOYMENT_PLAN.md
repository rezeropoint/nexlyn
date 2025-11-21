# LynxGraph 微服务部署和集成方案

## 📋 文档概述

本文档描述 LynxGraph 逻辑引擎的微服务架构、IoT 集成和外部数据服务集成方案。

**当前状态**（2025-01-15）：
- ✅ **Phase 1-2**: 微服务架构、前端界面（100% 完成）
- ✅ **Phase 3**: lynxengine 核心服务、IoT 集成（100% 完成）
- ⏳ **Phase 4**: 端到端测试、生产部署（待实施）

**核心目标**：
- ✅ 微服务拆分（Manager模式 + Engine模式）
- ✅ IoT 引擎数据分发集成
- ✅ 外部数据服务集成（历史数据查询、API调用）
- ✅ 前端配置界面

**最后更新**: 2025-01-15

---

## 🏗️ 总体架构

### 微服务拆分

根据 `pkg/lynxgraph/DEVELOPMENT.md` 的设计，拆分为两个独立微服务：

| 微服务 | 路径 | 协议 | 容器端口 | 宿主机端口 | 职责 |
|--------|------|------|----------|------------|------|
| **lynxmanager** | `restful/lynxmanager` | REST API | 8888 | 8891 | 配置管理（信息原子类型、逻辑图CRUD） |
| **lynxengine** | `service/lynxengine` | gRPC | 9999 | 9992 | 运行时执行（接收信息原子、执行逻辑图） |

### 架构决策

1. **服务定位**：遵循 go-zero 规范（REST服务→`restful/`，gRPC服务→`service/`）
2. **通信协议**：gRPC用于高频IoT数据传输，REST用于前端配置管理
3. **配置同步**：通过Etcd实现Manager→Engine的配置热更新

---

## 📦 微服务详细设计

### 1️⃣ lynxmanager 服务（REST API）

**位置**: `restful/lynxmanager/`

**职责**: 配置管理（信息原子类型CRUD、逻辑图配置CRUD、逻辑块规格查询）

**API分组**: 信息原子类型管理、逻辑图配置管理、逻辑块规格查询

**依赖**: PostgreSQL（元数据）、MongoDB（图详情）、Redis（缓存）、Etcd（配置通知）

**端口**: 容器内8888，宿主机8891

---

### 2️⃣ lynxengine 服务（gRPC）

**位置**: `service/lynxengine/`

**职责**: 运行时执行（接收信息原子、执行逻辑图、管理图上下文）

**gRPC接口**: ReceiveInfoAtom（接收信息原子）、HealthCheck（健康检查）

**依赖**: PostgreSQL（元数据）、MongoDB（图详情）、Redis（上下文和缓存）、Etcd（配置热更新）

**端口**: 容器内9999，宿主机9992

---

## 🔌 IoT 集成方案 ✅ 已完成（2025-01-15）

**状态**: IoT引擎通过gRPC将MQTT数据分发到LynxGraph引擎的功能已完整实现

**核心功能**:
- ✅ `DispatchTypeLynxGraph` 分发类型定义
- ✅ `DispatchConfig.InfoAtomTypeID` 字段支持
- ✅ `dispatchToLynxGraph()` gRPC调用实现（`pkg/iot/internal/dispatcher/lynxgraph.go`）
- ✅ gRPC 地址配置支持（`restful/iotmanager/etc/config.yaml`）
- ✅ 前端字段匹配验证（自动检查 IoT 字段与信息原子类型字段一致性）

**数据流**: MQTT消息 → IoT引擎提取字段 → gRPC调用lynxengine → 验证+分发 → 执行逻辑图

**配置示例** (`restful/iotmanager/etc/config.yaml`):
```yaml
LynxGraphConfig:
  grpcUrl: "lynxengine:9999"  # Docker环境使用服务名
```

---

## 🌐 外部数据服务集成方案

### 设计原则

**职责分离**:
- **LynxGraph**: 逻辑编排和流程控制（事件驱动）
- **外部服务**: 数据查询和计算（按需调用）
- **Storage**: 持久化存储（专业引擎）

**核心理念**:
- ✅ 信息原子是触发器，不是数据仓库
- ✅ GraphContext 是短期中转（TTL 1小时）
- ✅ InfoAtom 是短期引用（TTL 24小时）
- ✅ 历史数据由外部服务提供（ClickHouse、PostgreSQL、API）

### Service 接口注册机制

**核心原则**：LynxGraph Core 不定义业务接口，只提供服务注册机制

**注册流程**（参考 Skylark 示例）：
```
1. Engine 初始化时注入外部服务实例（如 go-skylark、iotquery gRPC 客户端）
2. 积木导入外部接口并类型断言（如 engine.SkylarkEngine、iotquery.IoTQuery）
3. 直接调用外部接口方法
```

**已集成服务**：
- `ServiceTypeSkylarkEngine`: go-skylark 流程引擎（`engine.SkylarkEngine`）
- `ServiceTypeSensorData`: IoTQuery gRPC 服务（`iotquery.IoTQuery`）

### 标准积木：FetchSensorData

**功能**: 查询外部服务的历史数据并保存到 GraphContext

**配置参数**:
- `serviceName`: 服务名称（必须是 "sensor_data"）
- `deviceIds`: 设备ID列表（可选，为空则查询组织下所有设备）
- `fieldNames`: 查询字段列表（必填，最多10个）
- `timeWindowSec`: 时间窗口（秒，相对于当前时间向前回溯）
- `aggregation`: 聚合类型（空/"avg"/"max"/"min"/"sum"/"count"/"last"，可选）
- `intervalSec`: 聚合时间间隔（秒，仅aggregation非空时有效）
- `limit`: 限制返回条数（默认1000，最大10000）
- `saveToContext`: 保存到 GraphContext 的 key（必填）
- `orgIds`: 组织ID列表（可选，从ExecutionContext获取）

**典型用法**:
```
入口节点 → FetchSensorData(查询温度历史) → 保存到 GraphContext["temp_history"]
                                           ↓
                                    汇聚节点（读取多个历史数据）
```

### 安全指数逻辑图架构

**核心理念**: 机器提取特征，LLM综合决策

```
温度传感器 → 查询历史 → 特征提取(变化率/异常/趋势) → GraphContext
                                                      ↓
湿度传感器 → 查询历史 → 特征提取(变化率/异常/趋势) ─→ LLMDecisionMaker
                                                      ↑    (综合分析)
烟雾传感器 → 阈值检查 ──────────────────────────────→      ↓
                                              输出: 安全指数
                                                    告警级别
                                                    推理过程
                                                    建议行动
                                                      ↓
                                              Switch条件路由
                                    ┌─────────┬─────────┬─────────┐
                                  紧急      危险      警告      正常
                                    ↓         ↓         ↓         ↓
                              告警+工单    告警     记录日志   记录日志
```

**关键优势**:
- **变化率优先**: 30度陡增 vs 30度缓慢到达，LLM能区分
- **异常检测**: 基于24小时基线的统计学方法
- **智能决策**: LLM理解多传感器关联（温度↑+湿度↓+烟雾=火灾）
- **可解释性**: LLM提供自然语言推理过程

### Redis TTL 配置

| 数据类型 | 推荐TTL | 用途 | 说明 |
|---------|--------|------|------|
| **GraphContext** | 1小时 | 节点间传递 | 短期协调，不作为数据源 |
| **InfoAtom** | 24小时 | 短期引用 | 触发器标识，不存储大量数据 |
| **历史数据** | 不存储 | 外部查询 | 由ClickHouse/PostgreSQL管理 |

### 标准积木设计（安全指数逻辑图导向）

**数据查询积木**:
- `FetchSensorData`: 查询多时间窗口历史数据（5分钟/1小时/24小时）

**特征提取积木**:
- `StatisticalAnalyzer`: 统计分析（均值、标准差、最值）
- `RateOfChange`: 变化率计算（识别陡然变化）
- `AnomalyDetector`: 异常检测（基于历史基线的z-score）
- `TrendAnalyzer`: 趋势分析（线性回归、移动平均）

**综合决策积木** ⭐:
- `LLMDecisionMaker`: **LLM综合决策**（调用LLM分析所有特征，输出安全指数、告警级别、推理过程、建议行动）

**输出积木**:
- `Switch`: 条件路由（根据LLM决策结果）
- `Alert`: 告警通知
- `StateManager`: 状态管理和去重

### 待实现任务

**外部服务集成**:
- [x] IoTQuery gRPC 服务集成（已完成）
- [x] TimeSeriesDataset 标准数据格式（已完成）
- [ ] LLM 服务集成（待实现）

**标准积木实现**:
- [x] FetchSensorData（查询传感器历史数据）
- [x] StatisticalAnalyzer（统计分析：均值、标准差、最大值、最小值、中位数、方差）
- [x] RateOfChange（变化率：绝对/相对，单/多时间窗口）
- [x] AnomalyDetector（异常检测：Z-score、IQR方法）
- [x] TrendAnalyzer（趋势分析：线性回归、预测）
- [ ] LLMDecisionMaker（LLM决策引擎，核心）
- [ ] Switch（条件路由）/ StateManager（状态管理）/ Alert（告警）

**后续任务**:
- [ ] 安全指数逻辑图端到端测试
- [ ] 性能测试和优化（Redis IO、数据冗余）

---

## 💾 数据库 Schema

**PostgreSQL**:
- `info_atom_types`: 信息原子类型定义（支持UUID、版本管理、租户隔离）
- `graph_configs`: 逻辑图元数据（支持UUID、版本管理、启用状态）

**MongoDB**:
- `graph_details`: 逻辑图详细配置（节点列表、边列表、逻辑块配置）

**Redis**:
- GraphContext: `{prefix}graph:{tenantId}:{graphKey}:{contextKey}` (TTL 1小时)
- InfoAtom: `{prefix}info:{tenantId}:{id}` (TTL 24小时)

详见 `nexlyn-deploy/postgres-init-scripts/009-lynxgraph-schema.sql`

---

## 🔄 核心数据流

### IoT → LynxGraph 数据流

```
MQTT消息 → IoT引擎(字段提取) → gRPC调用 → lynxengine → 验证+构造InfoAtom
→ Engine.Dispatch → 查找订阅图 → 执行逻辑块链
```

### 配置管理流程

```
前端 → lynxmanager REST API → Manager.CreateGraphConfig → PostgreSQL+MongoDB
→ Etcd通知 → lynxengine监听 → LoadGraph → 注册到内存 → 开始处理信息原子
```

### 外部服务查询流程

```
信息原子触发 → FetchSensorData积木 → service.Get("sensor_data")
→ 外部服务查询(ClickHouse) → 结果保存到GraphContext → 后续节点读取
```

---

## 🎨 前端配置界面

**详细文档**: [LynxGraph前端页面设计规范](../../restful/lynxmanager/FRONTEND_DESIGN.md)

**实施状态**: ✅ Phase 1-2 已完成（2025-10-26）

### 核心页面

1. **信息原子类型管理** (`/logic-engine/infoatom-types`)
   - ProTable列表、创建/编辑表单、字段可视化编辑器
   - 支持 JSONPath、字段类型、数组标记

2. **逻辑图配置** (`/logic-engine/graph-config`)
   - 列表页：卡片/表格双视图、图标系统（32个预定义图标）
   - 详情页：5个Tab（基本信息、可视化编辑器、运行时数据、Webhooks、统计分析）
   - 技术栈：AntV X6、react-icons、@ant-design/charts

3. **IoT模板集成配置** (`/iot-management/sensor-template`)
   - LynxGraph分发配置区域
   - 信息原子类型选择、字段完全匹配验证
   - 实时显示字段匹配状态（已匹配/缺失/未使用）

---

## 🐳 Docker 部署

**文件**: `nexlyn-deploy/docker-compose.yml`

**服务**:
- `lynxmanager`（REST，8888→8891）
- `lynxengine`（gRPC，9999→9992）

**依赖**: PostgreSQL、MongoDB、Redis、Etcd

**构建**: `./build-with-env.ps1 lynxmanager|lynxengine`

---

## 📝 开发任务

### ✅ 已完成（Phase 1-2）

**微服务基础架构**:
- [x] lynxmanager 和 lynxengine 微服务创建
- [x] API定义、protobuf生成、依赖注入重构
- [x] lynxmanager Logic 层实现（11个文件）
- [x] 数据库 Schema 创建（PostgreSQL + MongoDB）
- [x] 编译通过

**IoT集成**:
- [x] 扩展 IoT Dispatcher 支持 LynxGraph 分发
- [x] gRPC 调用实现（`pkg/iot/internal/dispatcher/lynxgraph.go`）
- [x] gRPC 地址配置支持（✅ 2025-01-15 完成）
- [x] 字段映射和验证逻辑
- [x] 前端字段匹配验证

**前端页面**:
- [x] 信息原子类型管理页面（ProTable、字段编辑器）
- [x] 逻辑图配置页面（卡片/表格双视图、AntV X6编辑器、5个Tab）
- [x] IoT模板配置增强（LynxGraph分发配置、字段验证）

### 🔨 待实现（Phase 3-4）

**lynxengine 服务**:
- [x] 实现 ReceiveInfoAtomLogic ✅ 已完成（`service/lynxengine/internal/logic/receiveinfoatomlogic.go`, 241行）
- [x] 实现 HealthCheckLogic ✅ 已完成
- [ ] 端到端测试（IoT → LynxGraph 数据流）

**外部数据服务集成**:
- [x] 定义 `SensorDataService` 接口（`core.ServiceTypeSensorData`，`core.SensorDataService`）
- [x] 实现 IoTQuery gRPC 查询服务封装（`service/lynxengine/internal/service/sensordata.go`）
- [x] 创建 `FetchSensorData` 标准积木（`pkg/lynxgraph/internal/blocks/standard/queryhistorydata`）
- [x] Engine 初始化时注入 Service 注册表（`service/lynxengine/internal/svc/servicecontext.go`）
- [x] 更新配置文件（`service/lynxengine/internal/config/config.go`，`etc/config.yaml`）

**测试和部署**:
- [ ] 单元测试（依赖注入验证）
- [ ] 集成测试（完整数据流）
- [ ] Dockerfile 创建
- [ ] 生产环境部署验证

---

## 🎯 关键技术决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| **微服务架构** |
| 服务拆分 | REST→`restful/`, gRPC→`service/` | 遵循 go-zero 规范 |
| 配置热更新 | Etcd | 分布式一致性，前缀监听 |
| 依赖注入 | 外部注入数据库连接 | 连接复用，易于测试 |
| **数据架构** |
| 信息原子职责 | 触发器（非数据仓库） | 事件驱动模型 |
| GraphContext TTL | 1小时 | 短期协调，节点间传递 |
| InfoAtom TTL | 24小时 | 短期引用，不存储大量数据 |
| 历史数据存储 | 外部服务（ClickHouse） | 专业时序数据库 |
| **外部服务集成** |
| 服务注册机制 | core.Service接口 | 解耦，支持多种服务类型 |
| 数据查询时机 | 按需调用（积木触发） | 避免预加载，减少内存 |
| 查询结果存储 | GraphContext | 短期中转，供后续节点使用 |
| **逻辑积木设计** |
| 特征提取 | 机器计算（统计/变化率/异常） | 精确、快速、可靠 |
| 综合决策 | LLM分析 | 灵活、智能、可解释 |
| 职责分离 | 机器提取特征，LLM做决策 | 发挥各自优势 |
| **IoT集成** |
| 字段数据来源 | fieldMappings（唯一） | 统一数据源，简化维护 |
| 信息原子类型作用 | 验证约束 | 前端字段完全匹配验证 |
| **前端架构** |
| 列表页架构 | 卡片+表格双视图 | 配置+监控混合场景 |
| 图形编辑器 | AntV X6 | Ant Design体系一致 |
| 样式管理 | CSS变量+Less模块 | 主题切换支持 |

---

## ⚖️ 架构设计权衡（已知约束和后续优化方向）

### 📊 GraphContext管道模式的设计权衡

**设计选择**：采用管道模式（Pipeline Pattern），每个逻辑块通过GraphContext（Redis）传递数据

**优势**：
- ✅ **职责单一**：每个积木只关注自己的业务逻辑，不关心上下游
- ✅ **可观测性强**：GraphContext存储在Redis，可通过CLI实时查看中间结果
- ✅ **容错性好**：节点失败时已保存的数据不会丢失，可从任意节点重试
- ✅ **解耦设计**：积木之间无直接依赖，易于单元测试和独立开发

**已知约束**：

#### 1️⃣ Redis IO开销

**现象**：
```
10个节点的逻辑图 = 10次Redis写入 + 10次Redis读取 = 20次网络IO
高频触发场景（如每秒触发）会产生大量Redis IO
```

**影响评估**：
- **低频场景**（每分钟触发）：影响可忽略，Redis单机可轻松支持10K+ QPS
- **中频场景**（每秒触发）：需要关注Redis性能，建议使用Redis Cluster
- **高频场景**（每秒100+触发）：需要优化（见下方优化方案）

**后续优化方案**（按优先级）：
1. **Level 1（轻量优化）**：使用 Redis Pipeline 批量读写（减少50%网络往返）
2. **Level 2（中等优化）**：在 ExecutionContext 中缓存 GraphContext（内存缓存，仅在同一次执行内有效）
3. **Level 3（重度优化）**：引入"热路径"标记 - 关键数据直接传递（内存），非关键数据走Redis（持久化）

**当前策略**：✅ **保持现状，待实际性能测试后决定优化级别**

---

#### 2️⃣ 数据冗余

**现象**：
```
temp_history (原始数据):  {records: 10000条} → 约 1MB Redis内存
temp_stats (统计特征):    {mean/std/max/min}  → 约 1KB Redis内存
temp_rate (变化率):       {rate_5min/rate_1h} → 约 1KB Redis内存

原始数据占用99.9%的内存，但后续节点可能只需要统计特征
```

**影响评估**：
- **安全指数逻辑图**（典型场景）：
  - 3个传感器 × 1小时历史（3600条） × 100字节/条 ≈ 1MB
  - 统一1小时TTL，Redis内存占用可控（< 100MB/租户）
- **大规模场景**（100个逻辑图，每分钟触发）：
  - 100个图 × 1MB × 60次/小时 = 6GB/小时（需要优化）

**后续优化方案**（按优先级）：
1. **差异化TTL**（推荐）：
   - 原始数据：30分钟TTL（满足短期分析需求）
   - 统计特征：2小时TTL（供LLM决策使用）
   - 决策结果：24小时TTL（供审计和可视化使用）

2. **数据压缩**（可选）：
   - 大数据量场景启用压缩（如 gzip）
   - 配置项：`GraphContextOptions{Compressed: true}`

3. **按需加载**（可选）：
   - 原始数据不保存到GraphContext，而是保存查询参数
   - 后续节点需要时重新查询（适用于数据量极大的场景）

**当前策略**：✅ **统一1小时TTL，待实际内存监控后启用差异化TTL**

---

#### 3️⃣ 数据格式转换开销

**现象**：
```
FetchSensorData 返回: gRPC响应（Protocol Buffers格式）
            ↓ 转换（O(n)遍历 + JSON解析）
保存到GraphContext: TimeSeriesDataset（标准格式）
            ↓ JSON序列化（Redis存储）
StatisticalAnalyzer读取: JSON反序列化 + 类型断言
```

**优化已完成**：✅
- 定义标准 `TimeSeriesDataset` 格式（`pkg/lynxgraph/core/dataset.go`）
- 统一数据结构，避免每个积木重复转换
- 提供工具方法（`GetValues()`、`GetTimestamps()`），简化后续开发

**剩余开销**：
- JSON序列化/反序列化（Redis存储必须）
- 类型断言（Go语言限制，`interface{}`转具体类型）

**可接受理由**：
- JSON序列化是Redis存储的必要开销
- 类型断言开销极小（纳秒级）
- 标准格式带来的开发效率提升远大于性能开销

---

### 🎯 总结：设计哲学

**核心理念**：**可观测性 > 性能优化**（先保证正确性和可调试性，再优化性能）

**验证策略**：
1. ✅ **Phase 3**：完成基础实现，保证功能正确性（当前阶段）
2. ⏳ **Phase 5**：性能测试，监控Redis IO和内存占用
3. ⏳ **Phase 6**：根据实际数据决定优化级别（避免过早优化）

**后续优化触发条件**：
- Redis CPU使用率 > 70%（启用Pipeline批量读写）
- Redis内存使用 > 50%（启用差异化TTL）
- 单次执行耗时 > 1秒（启用ExecutionContext内存缓存）

---

## 📚 参考文档

- `pkg/lynxgraph/DEVELOPMENT.md` - LynxGraph 开发指南
- `pkg/iot/DEVELOPMENT.md` - IoT 引擎开发指南
- `restful/lynxmanager/FRONTEND_DESIGN.md` - 前端页面设计规范
- `go-zero`: https://go-zero.dev/
- `AntV X6`: https://x6.antv.antgroup.com/

---

**文档结束** | 最后更新: 2025-10-29
