# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

Nexlyn 是一个为 LynxGraph 逻辑引擎提供的可视化管理平台，采用前后端分离架构。

**技术栈**：
- **后端**: Go 1.25.1 + go-zero + sqlx + CasbinX
- **前端**: React 18 + TypeScript + Ant Design Pro + UmiJS（需 Node.js 20+）
- **数据库**: PostgreSQL + ClickHouse + Redis + MongoDB + Etcd
- **物联网**: MQTT (Paho) + Skylark 流程引擎（go-skylark/v2）
- **媒体**: m7s.live v5 + GB28181

## 常用命令

### 后端
```bash
# 编译检查
go build ./...

# 运行测试
go test ./...
go test ./pkg/lynxiot/...                           # 指定包
go test -run TestFunctionName ./pkg/lynxiot/internal/template  # 单个测试

# API 代码生成（不加 -style 参数）
goctl api go -api restful/iotmanager/iot.api -dir restful/iotmanager

# gRPC 代码生成
goctl rpc protoc service/{service}/{service}.proto --go_out=. --go-grpc_out=. --zrpc_out=service/{service}
```

### 前端
```bash
cd frontend
npm run lint      # 代码检查
npm run tsc       # TypeScript 类型检查
npm test          # 运行测试

# 以下命令需用户手动执行（Claude Code 不应主动执行）
# npm install       # 安装依赖
# npm run start:dev # 开发模式
# npm run build     # 构建生产版本
```

## 核心服务架构

### REST 服务
| 服务 | 入口 | 配置 | 说明 |
|-----|------|------|------|
| backend | `restful/backend/backend.go` | `etc/config.yaml` | 用户、权限、租户、组织管理 |
| iotmanager | `restful/iotmanager/iot.go` | `etc/config.yaml` | IoT 平台配置 CRUD |
| eventhandler | `restful/eventhandler/eventhandler.go` | `etc/config.yaml` | Skylark 流程操作（15 个接口） |
| mediahandler | `restful/mediahandler/mediahandler.go` | `etc/mediahandler.yaml` | GB28181 视频接入 |

### gRPC 服务
| 服务 | 入口 | 端口 | 说明 |
|-----|------|------|------|
| lynxengine | `service/lynxengine/lynxengine.go` | - | LynxGraph 逻辑引擎 |
| iotquery | `service/iotquery/iotquery.go` | - | ClickHouse 时序查询 |
| eventsync | `service/eventsync/eventsync.go` | 9998 | Skylark 用户/组织同步（7 个 RPC） |

### 核心包
| 包 | 说明 |
|---|------|
| `pkg/lynxiot/` | IoT 引擎（MQTT、设备管理、数据分发） |
| `pkg/lynxgraph/` | LynxGraph 逻辑引擎（逻辑块、逻辑图、信息原子） |
| `pkg/nexlyn/` | m7s 媒体服务插件 |
| `internal/auth/` | JWT 认证、组织权限验证 |

## IoT 引擎架构 (pkg/lynxiot/)

采用三层架构：**core/**（领域模型）→ **engine/**（接口层）→ **internal/**（Manager 实现）

### Manager 模式
- **Template Manager**: 传感器模板（PostgreSQL + Etcd）
- **Device Manager**: 设备绑定（PostgreSQL + Redis 在线状态）
- **MQTT Manager**: 订阅、在线检测、业务数据提取
- **Platform Manager**: 平台配置（Skylark/Webhook/Kafka）
- **Dispatcher Manager**: 异步多平台分发
- **Tag/Metadata Manager**: 标签和元数据 CRUD

### 数据处理流程
1. MQTT 订阅 → 2. 时间窗口去重（Redis 锁）→ 3. 在线检测（更新 Redis TTL 300s）→ 4. 业务数据提取 → 5. 异步分发到平台

### 存储架构
- **PostgreSQL**: 元数据 + 多租户隔离
- **Etcd**: 运行时配置（`sensor-online/`、`sensor-data/`、`platform-`）
- **Redis**: 在线状态（`iot:device:online:{id}`）、分布式锁

### MQTT 主题格式
```
nexlyn/iot/{category}/{model}/{device_id}/{topicSuffix}
```

## LynxGraph 引擎架构 (pkg/lynxgraph/)

事件驱动型逻辑图执行引擎，核心概念：
- **信息原子 (InfoAtom)**: 事件数据单元
- **逻辑图 (LogicGraph)**: 节点 + 边组成的有向图
- **逻辑块 (LogicBlock)**: 8 种类型（filter/state_machine/action 等）

## 开发规范

### Go 后端
- **NULL 值**: 使用 `sql.NullString`/`sql.NullTime`
- **数组字段**: 使用 `pq.Array()` 包装
- **日志**: `logx.Error`（无 Warn 方法）
- **数据库**: 使用 go-zero sqlx，不使用 GORM
- **系统表**: `system_` 前缀（system_users、system_organizations 等）

### IoT 引擎开发
- **设备注册**: 在 `core/devices/` 创建文件，init() 中调用 `core.RegisterCategory()`
- **模板更新**: 禁止修改 model/category 字段（防止 Etcd key 不一致）
- **平台配置**: skylark/webhook/kafka 字段互斥
- **循环依赖**: 使用函数类型解耦（如 `GetPlatformConfigFunc`）

### 前端（详见 `frontend/前端编码规范与最佳实践.md`）
- **主题适配**: 使用 Less 模块 + CSS 变量，禁止硬编码颜色
- **动态样式**: 使用 `theme.useToken()`
- **ProForm**: 使用 formRef，DrawerForm/ModalForm 设置 `destroyOnHidden: true`
- **布局**: 优先 Flex 组件，使用 Splitter 实现可拖拽分栏

## 权限系统

基于 CasbinX，配置位于 `pkg/nexlyn/etc/casbin_model.conf`

**验证流程**：REST 解析 JWT → 查询用户可见组织 ID → Manager 层 SQL 约束（`org_id = ANY($X)`）

## 部署

配置位于 `nexlyn-deploy/`，使用 `build-with-env.ps1` 构建。

**注意**: Claude Code 不应主动构建镜像或启动前端服务器。

## 重要参考文档

- `pkg/lynxiot/DEVELOPMENT.md` - IoT 引擎详细设计
- `pkg/lynxgraph/DEVELOPMENT.md` - LynxGraph 引擎设计
- `restful/eventhandler/SKYLARK_V2_TODO.md` - Skylark V2 实施计划
- `frontend/前端编码规范与最佳实践.md` - 前端规范
- `nexlyn-deploy/postgres-init-scripts/*.sql` - 数据库 Schema

## 常见问题

| 问题 | 解决方案 |
|-----|---------|
| PostgreSQL `converting NULL to string` | 使用 `sql.NullString` 类型 |
| PostgreSQL 数组操作失败 | 使用 `pq.Array()` 包装 |
| Ant Design `destroyOnClose` 警告 | 使用 `destroyOnHidden` 替代 |
| MQTT 消息未触发在线更新 | 检查 Etcd 配置和主题格式 |

### 调试命令
```bash
# 查看 Etcd 配置
etcdctl get --prefix sensor-online/
etcdctl get --prefix sensor-data/

# 查看 Redis 在线设备
redis-cli KEYS "iot:device:online:*"
```
