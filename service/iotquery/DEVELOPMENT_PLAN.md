# IoT查询服务开发计划

## 📋 文档概述

本文档描述IoT查询服务（service/iotquery）的架构设计、实施计划和开发指南。

**核心目标**：
- 提供独立的gRPC查询服务，供后端服务调用（如LynxGraph引擎）
- 轻量级架构，只依赖PostgreSQL + ClickHouse
- 与IoT配置管理服务（restful/iotmanager）职责分离
- 复用pkg/iot/query查询引擎接口层

**最后更新**: 2025-10-30

---

## 🏗️ 总体架构

### 服务定位

| 维度 | restful/iotmanager | service/iotquery |
|------|-------------------|-----------------|
| **协议** | REST API | gRPC |
| **调用方** | 前端（Web UI） | 后端服务（LynxGraph等） |
| **职责** | 配置管理（模板、设备、平台CRUD） | 纯查询（时序数据、最新值、统计） |
| **依赖** | PostgreSQL + Redis + ClickHouse + Etcd + MQTT | PostgreSQL + ClickHouse（仅2个） |
| **端口** | 8888（REST） | 9993（gRPC） |
| **引擎** | pkg/iot/engine（全功能） | pkg/iot/query（轻量级） |

### 架构优势

1. **职责清晰**：查询与配置管理、MQTT处理完全解耦
2. **依赖最小**：iotquery服务只需要PostgreSQL+ClickHouse，不需要：
   - MQTT Broker连接
   - Redis分布式锁
   - Etcd配置监听
   - Skylark引擎
3. **独立扩缩容**：LynxGraph查询量大时，只扩iotquery服务即可
4. **高性能**：gRPC二进制协议，适合微服务间高频调用

---

## 📦 目录结构

```
service/iotquery/
├── iotquery.proto              # gRPC接口定义
├── iotquery.go                 # 服务入口
├── pb/                         # protobuf生成代码
│   ├── iotquery.pb.go         # 消息定义
│   └── iotquery_grpc.pb.go    # gRPC服务定义
├── etc/
│   └── iotquery.yaml          # 服务配置
├── internal/
│   ├── config/
│   │   └── config.go          # 配置结构定义
│   ├── svc/
│   │   └── servicecontext.go  # 服务上下文（依赖注入）
│   ├── logic/                 # 业务逻辑层
│   │   ├── querytimeserieslogic.go     # 时序数据查询
│   │   ├── getlatestvalueslogic.go     # 最新值查询
│   │   └── getdevicestatisticslogic.go # 统计查询
│   └── server/
│       └── iotqueryserver.go  # gRPC服务器实现
└── Dockerfile                 # Docker构建文件（待创建）
```

---

## 🔌 gRPC接口定义

### 3个核心查询接口

#### 1. QueryTimeSeries - 时序数据查询
```protobuf
rpc QueryTimeSeries(QueryTimeSeriesReq) returns (QueryTimeSeriesResp);
```

**功能**：
- 支持原始数据查询和6种聚合类型（avg/max/min/sum/count/last）
- 支持时间窗口聚合（如5分钟、1小时）
- 支持分页和排序

**请求参数**：
- `tenant_id`：租户ID（必填）
- `org_ids`：组织ID列表（必填，权限过滤）
- `device_ids`：设备ID列表（可选）
- `field_names`：字段列表（必填，最多10个）
- `start_time`/`end_time`：时间范围（Unix时间戳，秒）
- `aggregation`：聚合类型（空/"avg"/"max"/"min"/"sum"/"count"/"last"）
- `interval_seconds`：聚合间隔（秒）
- `limit`/`offset`：分页参数
- `order_by`/`order_dir`：排序参数

#### 2. GetLatestValues - 获取最新值
```protobuf
rpc GetLatestValues(GetLatestValuesReq) returns (GetLatestValuesResp);
```

**功能**：
- 获取设备指定字段的最新值和时间戳
- 用于实时监控和可视化大屏

#### 3. GetDeviceStatistics - 获取统计信息
```protobuf
rpc GetDeviceStatistics(GetDeviceStatisticsReq) returns (GetDeviceStatisticsResp);
```

**功能**：
- 计算指定时间范围内的统计指标（min/max/avg/sum/count/last_value）
- 支持多设备、多字段统计

---

## 💾 数据架构

### 数据源

- **PostgreSQL**：
  - `iot_device_bindings`表：查询设备ID范围（基于org_ids权限过滤）
  - 获取设备元信息（device_model、device_category）

- **ClickHouse**：
  - 按字段分表存储：`iot_{field_name}_data`
  - 表结构：`timestamp`, `device_id`, `device_model`, `device_category`, `tenant_id`, `value`
  - 分区策略：按月分区（`PARTITION BY toYYYYMM(timestamp)`）
  - 数据保留：180天TTL自动清理

### 权限验证机制

```
1. 调用方（如LynxGraph）提供 tenant_id 和 org_ids（已在其执行上下文中验证）
2. iotquery服务接收请求 → 传递给pkg/iot/query引擎
3. Query Manager查询PostgreSQL：WHERE tenant_id = ? AND org_id = ANY(?)
4. 获取设备ID范围 → 在ClickHouse中查询时序数据（只查询已验证的设备）
```

**安全保障**：即使调用方传入错误的org_ids，也无法查询到无权限的数据。

---

## 🔧 实施计划

### ✅ 已完成（Phase 1-3）

1. **pkg/iot/query查询引擎接口层**：
   - [x] query.go：Query接口定义
   - [x] config.go：配置结构
   - [x] handler.go：实现（调用internal/query/Manager）

2. **service/iotquery服务骨架**：
   - [x] iotquery.proto：gRPC接口定义
   - [x] 使用goctl生成服务结构
   - [x] 修复import路径（pb包位置）

3. **Phase 3核心实现**：
   - [x] 扩展配置文件（etc/config.yaml）：PostgreSQL + ClickHouse配置
   - [x] 扩展Config结构：添加数据库配置字段
   - [x] 实现ServiceContext：初始化数据库连接和查询引擎
   - [x] 实现converter.go：protobuf与core类型转换（参考iotmanager）
   - [x] 实现3个Logic层接口：QueryTimeSeries、GetLatestValues、GetDeviceStatistics
   - [x] 编译验证：服务编译通过

### ✅ 已完成（Phase 4）

#### Phase 4：部署和测试

1. **创建Dockerfile**（`service/iotquery/Dockerfile`）：
   - ✅ 使用多阶段构建（builder + alpine）
   - ✅ 支持SSH私钥访问私有仓库（rezeropoint/etcdtrigger等）
   - ✅ 配置文件从etc/config.yaml复制
   - ✅ 暴露9993端口

2. **更新docker-compose.yml**：
   - ✅ 添加iotquery服务配置
   - ✅ 端口映射：9993:9993（gRPC）
   - ✅ 依赖：postgres（健康检查）+ clickhouse（健康检查）+ etcd（启动即可）
   - ✅ 环境变量：POD_NAME="nexlyn-iotquery-docker"

3. **待完成：集成测试**
   - 使用grpcurl测试3个查询接口
   - 验证权限过滤生效
   - 验证查询结果正确

---

## 🎯 关键技术决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| **协议** | gRPC | 二进制协议，高性能，适合微服务间调用 |
| **数据格式** | JSON字符串 | protobuf不支持动态类型，用JSON字符串传递value |
| **时间戳** | Unix时间戳（秒） | 简单、通用、避免时区问题 |
| **权限验证** | 假定调用方已验证 | 服务间调用，调用方（LynxGraph）已在执行上下文验证 |
| **依赖注入** | 外部创建连接 | 连接复用，易于测试 |
| **错误处理** | gRPC status codes | 标准gRPC错误码 |

---

## 📊 性能优化

### 查询限制
- **默认限制**: 1000条/请求
- **最大限制**: 10000条/请求
- **查询超时**: 30秒
- **字段限制**: 最多10个字段/请求

### 优化策略
1. **PostgreSQL索引**：复合索引(tenant_id, org_id, device_id)
2. **ClickHouse优化**：按字段分表、时间分区、主键排序、布隆过滤器
3. **连接池管理**：合理配置MaxOpenConns（默认10）
4. **缓存策略**（可选）：对于聚合查询，可在Redis缓存结果（TTL 5分钟）

---

## 🔗 LynxGraph集成示例

### LynxGraph调用iotquery服务

**1. 创建gRPC客户端** (`service/lynxengine/internal/svc/servicecontext.go`)：
```go
import pb "github.com/rezeropoint/nexlyn/service/iotquery/pb"

type ServiceContext struct {
    Config          config.Config
    IoTQueryClient  pb.IoTQueryClient  // 新增
}

func NewServiceContext(c config.Config) *ServiceContext {
    // 连接iotquery服务
    conn, err := grpc.Dial(c.IoTQueryGrpcAddr, grpc.WithInsecure())
    if err != nil {
        panic(err)
    }

    return &ServiceContext{
        Config:         c,
        IoTQueryClient: pb.NewIoTQueryClient(conn),
    }
}
```

**2. QueryHistoryData积木调用** (`pkg/lynxgraph/internal/blocks/standard/query_history_data.go`)：
```go
func (b *QueryHistoryDataBlock) Execute(ctx core.ExecutionContext) error {
    // 从ExecutionContext获取租户ID和组织ID
    tenantID := ctx.TenantID
    orgIDs := ctx.UserOrgIDs  // LynxGraph已在执行上下文中验证

    // 调用iotquery gRPC服务
    resp, err := iotQueryClient.QueryTimeSeries(ctx, &pb.QueryTimeSeriesReq{
        TenantId:   tenantID,
        OrgIds:     orgIDs,
        DeviceIds:  []string{b.DeviceID},
        FieldNames: []string{"temperature"},
        StartTime:  time.Now().Add(-1 * time.Hour).Unix(),
        EndTime:    time.Now().Unix(),
        Aggregation: "avg",
        IntervalSeconds: 300,  // 5分钟聚合
    })

    // 将结果保存到GraphContext
    ctx.GraphContext.Set(b.SaveToContext, resp.Records)
    return nil
}
```

---

## 📝 验收标准

- [ ] 服务可以独立启动（不依赖MQTT、Redis等）
- [ ] 3个查询接口返回数据正确
- [ ] 权限验证生效（org_ids过滤）
- [ ] 查询性能达标（单次查询<1秒）
- [ ] Docker容器正常启动
- [ ] LynxGraph可以通过gRPC调用iotquery

---

## 📚 参考文档

- `pkg/iot/query/`：查询引擎接口层实现
- `pkg/iot/internal/query/`：Query Manager实现
- `pkg/iot/DEVELOPMENT.md`：IoT引擎开发指南
- `pkg/lynxgraph/DEPLOYMENT_PLAN.md`：LynxGraph部署计划（架构参考）
- go-zero文档：https://go-zero.dev/
- gRPC文档：https://grpc.io/

---

**文档结束** | 最后更新: 2025-10-29
