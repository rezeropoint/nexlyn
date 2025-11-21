# Nexlyn 插件开发说明

## 架构概述

Nexlyn 插件作为 monibucax 的**统一网关层**，负责：
- JWT 认证和权限验证
- 路径参数自动提取
- 请求路由和转发
- 与具体插件（GB28181等）的集成

## 核心组件

### 1. 处理器框架
- `NexlynHandlerFunc` - 统一处理器函数签名
- `NexlynContext` - 自动提取路径参数和用户信息的请求上下文
- `PathParamExtractor` - 路径参数自动提取器
- `HandlerConfig` - 处理器权限和方法配置

### 2. 文件组织
```
pkg/nexlyn/
├── index.go                # 插件入口和初始化
├── api.go                  # 路由注册
├── handler_framework.go    # 框架核心（中间件、参数提取）
├── handlers.go             # 通用处理器（测试、工具函数）
├── handlers_gb28281.go     # GB28181 插件处理器
├── handlers_xxx.go         # 其他插件处理器（待添加）
├── jwt.go                  # JWT 认证
└── middleware_permission.go # 权限中间件
```

## 添加新插件处理器

### 1. 创建处理器文件
```go
// handlers_newplugin.go
package plugin_nexlyn

// 标准转发模式：纯路由层，无业务逻辑
func (n *NexlynPlugin) handleNewPluginAction(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
    n.newPlugin.HandleNewAction(w, r, ctx.UserInfo.TenantId)
}

// 带路径参数的转发
func (n *NexlynPlugin) handleNewPluginDetail(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
    n.newPlugin.HandleDetail(w, r, ctx.PathParams["resourceId"], ctx.UserInfo.TenantId)
}
```

### 2. 在 api.go 中注册路由
```go
"/api/newplugin/action": n.JWTAuthMiddleware(n.createHandler("/api/newplugin/action", n.handleNewPluginAction, &HandlerConfig{
    RequireAuth: true,
    RequirePluginCheck: true,
    AllowedMethods: []string{http.MethodPost},
    Resource: "newplugin_resource",
    Action: "write",
})),
```

### 3. 在 index.go 中初始化目标插件连接
```go
// 在 Start() 方法中添加插件查找逻辑
n.newPlugin = n.findNewPlugin()
```

## 开发规范

### 1. 处理器函数命名
- 使用 `handle` + 功能名称：`handleDeviceList`, `handleTagCreate`
- 按插件分组：`handlers_gb28181.go`, `handlers_rtsp.go`

### 2. 纯转发原则
- **禁止参数校验**：不检查 `ctx.DeviceID == ""`，由目标插件处理
- **禁止业务逻辑**：不包含 `switch r.Method` 等业务判断
- **禁止错误处理**：不调用 `n.sendErrorResponse`，让目标插件处理
- **标准函数体**：每个处理器只有一行转发调用

### 3. 权限配置
- `Resource` - 资源类型（如 `gb28181_device`）
- `Action` - 操作类型（`read`, `write`, `delete`）
- `RequireAuth` - 是否需要 JWT 认证
- `RequirePluginCheck` - 是否检查目标插件可用性

### 4. 路径参数
框架自动提取以下参数到 `ctx`：
- `ctx.DeviceID` - 来自 `{deviceId}`
- `ctx.ChannelID` - 来自 `{channelId}`
- `ctx.TagID` - 来自 `{tagId}`
- `ctx.PathParams` - 所有路径参数映射

### 5. 标准转发模式
Nexlyn 仅作为**纯路由层**，所有处理器都遵循统一结构：

```go
// ✓ 正确：标准转发模式（无参数）
func (n *NexlynPlugin) handleDeviceList(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
    n.gb28181Plugin.HandleNexlynGetDeviceList(w, r, ctx.UserInfo.TenantId)
}

// ✓ 正确：标准转发模式（带路径参数）
func (n *NexlynPlugin) handleDeviceDetail(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
    n.gb28181Plugin.HandleNexlynGetDeviceDetail(w, r, ctx.DeviceID, ctx.UserInfo.TenantId)
}

// ✗ 错误：包含参数校验
func (n *NexlynPlugin) handleDeviceDetail(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
    if ctx.DeviceID == "" {  // ❌ 不要在 Nexlyn 中校验
        n.sendErrorResponse(w, http.StatusBadRequest, "设备ID不能为空", nil)
        return
    }
    n.gb28181Plugin.HandleNexlynGetDeviceDetail(w, r, ctx.DeviceID, ctx.UserInfo.TenantId)
}

// ✗ 错误：包含业务逻辑
func (n *NexlynPlugin) handleTagOperations(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
    switch r.Method {  // ❌ 不要在 Nexlyn 中判断 HTTP 方法
    case http.MethodPut:
        n.gb28181Plugin.HandleNexlynUpdateTag(w, r, ctx.TagID, ctx.UserInfo.TenantId)
    case http.MethodDelete:
        n.gb28181Plugin.HandleNexlynDeleteTag(w, r, ctx.TagID, ctx.UserInfo.TenantId)
    }
}
```

## 架构优势

### 1. **职责分离**
- **Nexlyn 层**：认证、权限、路由转发
- **插件层**：业务逻辑、参数校验、错误处理

### 2. **开发效率**
- **标准化**：所有处理器都是 3 行标准结构
- **无重复**：避免每个插件重复实现认证逻辑
- **易维护**：纯转发层，代码简洁清晰

### 3. **扩展性**
- **插件独立**：添加新插件不影响现有功能
- **路由统一**：所有 API 路径统一管理
- **权限集中**：统一的权限控制策略

### 4. **实际效果**
- **代码减少**：`handlers_gb28281.go` 从 133 行优化到 60 行
- **复杂度降低**：去除所有条件判断和错误处理逻辑
- **维护成本**：大幅降低，标准化程度高