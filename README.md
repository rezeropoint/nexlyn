# Nexlyn

> 为 LynxGraph 逻辑引擎提供的可视化管理平台
>
> **私有项目 - 仅供内部使用**

Nexlyn 是一个现代化的可视化管理平台，专为 LynxGraph 逻辑引擎设计，采用前后端分离架构，提供视频管理、算法管理、物联网管理和可视化规则引擎等核心功能。

## ✨ 特性

- 🎥 **视频管理** - 支持设备接入、分屏监控、媒体管理、录像管理
- 🤖 **算法管理** - 云边端 AI 算法统一管理平台（开发中）
- 🌐 **物联管理** - 物联网设备全生命周期管理（开发中）
- ⚡ **逻辑引擎** - 基于 LynxGraph 的可视化规则引擎（开发中）
- 👥 **系统管理** - 用户权限、租户管理、组织架构等
- 📱 **响应式设计** - 基于 Ant Design Pro 的现代化 UI
- 🔐 **权限管理** - 基于 CasbinX 的灵活权限控制
- 📺 **GB28181** - 完整的视频监控协议支持

## 🏗️ 架构

### 技术栈

**后端**
- Go 1.25.1
- go-zero 框架
- sqlx 数据库操作（go-zero内置）
- CasbinX 权限管理
- m7s.live v5 媒体服务
- GB28181 协议支持

**前端**
- React 18
- TypeScript
- Ant Design Pro
- UmiJS

**数据库**
- PostgreSQL (主数据库)
- Redis (缓存和会话)
- MongoDB (文档存储)
- Etcd (配置中心)

### 核心组件

```
nexlyn/
├── frontend/           # React 前端应用
├── restful/           # 后端服务
│   ├── backend/       # RESTful API 服务
│   └── mediahandler/  # 媒体服务
├── pkg/nexlyn/        # 核心插件包
├── internal/          # 内部包
└── nexlyn-deploy/     # Docker 部署配置
```

## 🚀 快速开始

### 环境要求

- Go 1.25.1+
- Node.js 20+
- PostgreSQL 13+
- Redis 6+
- Docker & Docker Compose (可选)

### 安装依赖

```bash
# 克隆项目
git clone https://codeup.aliyun.com/68d8eb1eab336f9c842ac4f0/nexlyn.git
cd nexlyn

# 安装 Go 依赖
go mod download

# 安装前端依赖
cd frontend
npm install
```

### 开发模式

**启动后端服务**

```bash
# 启动后端 API 服务
go run restful/backend/backend.go -f restful/backend/etc/backend.yaml

# 启动媒体服务
go run restful/mediahandler/mediahandler.go -f restful/mediahandler/etc/mediahandler.yaml

# 启动主程序（LynxGraph 引擎演示）
go run main.go
```

**启动前端**

```bash
cd frontend
npm run start:dev
```

访问 `http://localhost:8000` 查看应用。

### Docker 部署

```bash
cd nexlyn-deploy

# 使用推荐脚本构建所有服务
./build-with-env.ps1

# 或者构建单个服务
./build-with-env.ps1 mediahandler

# 启动服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

## 📖 API 文档

启动后端服务后，访问以下地址查看 API 文档：

- 后端 API: `http://localhost:8888/swagger/`
- 媒体服务 API: `http://localhost:8080/api/docs`

## 🔧 配置

### 数据库配置

```yaml
# restful/backend/etc/backend.yaml
Database:
  Host: localhost
  Port: 5432
  User: nexlyn
  Password: your_password
  DBName: nexlyn

Redis:
  Host: localhost:6379
  Password: your_redis_password
```

### 权限配置

权限模型配置文件位于 `pkg/nexlyn/etc/casbin_model.conf`，基于 RBAC 模型实现。

## 🧪 测试

```bash
# 运行后端测试
go test ./...

# 运行前端测试
cd frontend
npm test

# 代码覆盖率
npm run test:coverage
```

## 🛠️ 开发指南

### 代码规范

- 遵循 Go 官方代码规范
- 前端遵循 Ant Design Pro 约定
- 提交信息遵循 Conventional Commits 规范

### 分支管理

- `master` - 生产环境分支
- `develop` - 开发环境分支
- `feature/*` - 功能开发分支
- `bugfix/*` - 错误修复分支

### CI/CD

项目使用 GitLab CI/CD 进行自动化构建和部署：

- **测试阶段**: 自动运行单元测试和代码检查
- **构建阶段**: 构建 Go 二进制文件和前端资源
- **部署阶段**: 自动部署到开发/生产环境

## 📋 路由

### 主要页面

- `/welcome` - 欢迎页面
- `/video-management` - 视频管理
  - `/video-management/device` - 设备管理
  - `/video-management/screen` - 分屏监控
  - `/video-management/media` - 媒体管理
  - `/video-management/record` - 录像管理
- `/algorithm-management` - 算法管理（开发中）
- `/iot-management` - 物联管理（开发中）
- `/logic-engine` - 逻辑引擎（开发中）
- `/system` - 系统管理
- `/visualization-dashboard` - 可视化大屏

## 🛠️ 开发协作

### 分支策略

1. 从 `dev` 分支创建功能分支 (`git checkout -b feature/功能名称`)
2. 完成开发后提交更改 (`git commit -m '功能描述'`)
3. 推送到远程分支 (`git push origin feature/功能名称`)
4. 在阿里云Code平台创建合并请求（Merge Request）

## 📝 变更日志

查看 [CHANGELOG.md](CHANGELOG.md) 了解版本更新详情。

## 📄 许可证

本项目为私有软件，所有权利保留 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙋‍♂️ 内部支持

如果在开发过程中遇到问题，请联系项目负责人或通过内部渠道沟通。

## 🏆 致谢

- [go-zero](https://github.com/zeromicro/go-zero) - 优秀的 Go 微服务框架
- [Ant Design Pro](https://pro.ant.design/) - 企业级 UI 设计语言
- [m7s.live](https://github.com/langhuihui/monibuca) - 强大的流媒体服务器
- [CasbinX](https://github.com/rezeropoint/casbinx) - 权限管理解决方案

---

**Nexlyn** - 让视频管理和逻辑引擎可视化变得简单高效 🚀