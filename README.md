# Nexlyn

> 为 LynxGraph 逻辑引擎提供的可视化管理平台
>
> **私有项目 - 仅供内部使用**

Nexlyn 是一个现代化的可视化管理平台，专为 LynxGraph 逻辑引擎设计，采用前后端分离架构，提供视频管理、物联网管理、流程引擎和可视化规则引擎等核心功能。

## 特性

- **视频管理** - GB28181 视频接入、分屏监控、媒体管理、录像管理
- **物联管理** - IoT 设备全生命周期管理、MQTT 接入、多平台数据分发
- **流程引擎** - 基于 Skylark V2 的事件驱动流程引擎
- **逻辑引擎** - 基于 LynxGraph 的可视化规则引擎
- **系统管理** - 用户权限、租户管理、组织架构
- **权限控制** - 基于 CasbinX 的 RBAC 权限管理

## 技术栈

| 层级 | 技术 |
|-----|------|
| **后端** | Go 1.25.1 + go-zero + sqlx + CasbinX |
| **前端** | React 18 + TypeScript + Ant Design Pro + UmiJS |
| **数据库** | PostgreSQL + ClickHouse + Redis + MongoDB + Etcd |
| **物联网** | MQTT (Paho) + Skylark 流程引擎 (go-skylark/v2) |
| **媒体** | m7s.live v5 + GB28181 |

## 项目结构

```
nexlyn/
├── frontend/              # React 前端应用
├── restful/               # REST 服务
│   ├── backend/           # 用户、权限、租户、组织管理
│   ├── iotmanager/        # IoT 平台配置 CRUD
│   ├── eventhandler/      # Skylark 流程操作
│   ├── mediahandler/      # GB28181 视频接入
│   └── lynxmanager/       # LynxGraph 管理
├── service/               # gRPC 服务
│   ├── lynxengine/        # LynxGraph 逻辑引擎
│   ├── iotquery/          # ClickHouse 时序查询
│   └── eventsync/         # Skylark 用户/组织同步
├── pkg/                   # 核心包
│   ├── lynxiot/           # IoT 引擎
│   ├── lynxgraph/         # LynxGraph 逻辑引擎
│   ├── nexlyn/            # m7s 媒体服务插件
│   ├── imageutil/         # 图像处理工具
│   └── ossutil/           # OSS 工具
├── internal/              # 内部包
│   └── auth/              # JWT 认证、组织权限验证
└── nexlyn-deploy/         # Docker 部署配置
```

## 快速开始

### 环境要求

- Go 1.25.1+
- Node.js 20+
- PostgreSQL 13+
- Redis 6+
- Docker & Docker Compose (可选)

### 安装

```bash
# 克隆项目
git clone https://codeup.aliyun.com/68d8eb1eab336f9c842ac4f0/nexlyn.git
cd nexlyn

# 安装 Go 依赖
go mod download

# 安装前端依赖
cd frontend && npm install
```

### 开发模式

**后端服务**

```bash
# 后端 API 服务
go run restful/backend/backend.go -f restful/backend/etc/config.yaml

# IoT 管理服务
go run restful/iotmanager/iotmanager.go -f restful/iotmanager/etc/config.yaml

# 媒体服务
go run restful/mediahandler/mediahandler.go -f restful/mediahandler/etc/mediahandler.yaml

# 流程服务
go run restful/eventhandler/eventhandler.go -f restful/eventhandler/etc/config.yaml
```

**前端**

```bash
cd frontend
npm run start:dev
```

访问 `http://localhost:8000` 查看应用。

### Docker 部署

```bash
cd nexlyn-deploy

# 构建所有服务
./build-with-env.ps1

# 启动服务
docker-compose up -d
```

## 常用命令

### 后端

```bash
# 编译检查
go build ./...

# 运行测试
go test ./...

# API 代码生成
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
```

## 核心模块

### IoT 引擎 (pkg/lynxiot)

采用三层架构：**core/** -> **engine/** -> **internal/**

- **Template Manager**: 传感器模板（PostgreSQL + Etcd）
- **Device Manager**: 设备绑定（PostgreSQL + Redis 在线状态）
- **MQTT Manager**: 订阅、在线检测、业务数据提取
- **Platform Manager**: 平台配置（Skylark/Webhook/Kafka）
- **Dispatcher Manager**: 异步多平台分发

### LynxGraph 引擎 (pkg/lynxgraph)

事件驱动型逻辑图执行引擎：

- **信息原子 (InfoAtom)**: 事件数据单元
- **逻辑图 (LogicGraph)**: 节点 + 边组成的有向图
- **逻辑块 (LogicBlock)**: 8 种类型（filter/state_machine/action 等）

## 开发协作

### 分支策略

1. 从 `dev` 分支创建功能分支 (`git checkout -b feature/功能名称`)
2. 完成开发后提交更改 (`git commit -m '功能描述'`)
3. 推送到远程分支 (`git push origin feature/功能名称`)
4. 在阿里云 Code 平台创建合并请求

### 代码规范

- 遵循 Go 官方代码规范
- 前端遵循 Ant Design Pro 约定
- 提交信息遵循 Conventional Commits 规范

## 参考文档

- `pkg/lynxiot/DEVELOPMENT.md` - IoT 引擎详细设计
- `pkg/lynxgraph/DEVELOPMENT.md` - LynxGraph 引擎设计
- `frontend/前端编码规范与最佳实践.md` - 前端规范
- `nexlyn-deploy/postgres-init-scripts/*.sql` - 数据库 Schema

## 许可证

本项目为私有软件，所有权利保留。

## 致谢

- [go-zero](https://github.com/zeromicro/go-zero) - Go 微服务框架
- [Ant Design Pro](https://pro.ant.design/) - 企业级 UI 设计语言
- [m7s.live](https://github.com/langhuihui/monibuca) - 流媒体服务器
- [CasbinX](https://github.com/rezeropoint/casbinx) - 权限管理解决方案
