# Emiya 开发指南

> HTTP 请求体解析库，支持嵌套 JSON 递归解析。

---

## 📑 快速导航

| 场景 | 跳转 |
|------|------|
| 使用示例 | [使用方式](#使用方式) |
| 理解架构 | [包结构](#包结构) |
| 扩展功能 | [开发规范](#开发规范) |

---

## 📋 包结构

```
pkg/emiya/
├── emiya.go      # 接口定义 + 公开工厂函数 NewRegistry()
├── handler.go    # 接口实现 + 私有构造函数 newRegistry()
├── internal.go   # 私有方法（嵌套 JSON 解析）
└── config.go     # 常量配置
```

### 文件职责

| 文件 | 职责 |
|------|------|
| `emiya.go` | 定义 `Registry` 接口和公开工厂函数 |
| `handler.go` | `registry` 结构体实现、私有构造函数、sync.Pool 管理 |
| `internal.go` | `analyzingJSON` 递归解析逻辑 |
| `config.go` | `maxBodyLen`、`defaultBufferSize` 常量 |

---

## 使用方式

### 基本用法

```go
import "github.com/rezeropoint/nexlyn/pkg/emiya"

// 创建 Registry 实例（通常在 ServiceContext 中初始化一次）
registry := emiya.NewRegistry()

// 在 Handler 中使用
func MyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var data map[string]any
        if err := svcCtx.EmiyaRegistry.AnalyzingBody(&r.Body, &data); err != nil {
            httpx.ErrorCtx(r.Context(), w, fmt.Errorf("解析请求体失败: %v", err))
            return
        }
        // 使用 data...
    }
}
```

### 嵌套 JSON 解析

假设请求体为：
```json
{
    "name": "test",
    "nested": "{\"key\": \"value\", \"inner\": \"{\\\"deep\\\": 123}\"}"
}
```

解析后 `data` 结构为：
```go
map[string]any{
    "name": "test",
    "nested": map[string]any{
        "key": "value",
        "inner": map[string]any{
            "deep": 123,
        },
    },
}
```

---

## 开发规范

### 1. 包结构规范

遵循 `pkg/lynxiot/DEVELOPMENT.md` 的 Manager 包规范：

| # | 检查项 | 要求 |
|---|--------|------|
| 1 | 接口文件 | `emiya.go` 定义 `Registry` 接口 |
| 2 | 公开工厂 | `emiya.go` 定义 `NewRegistry() Registry` |
| 3 | 私有构造 | `handler.go` 定义 `newRegistry() *registry` |
| 4 | 返回类型 | 公开返回接口，私有返回具体类型 |
| 5 | 结构体 | 实现结构体小写（私有）：`type registry struct` |
| 6 | 无全局变量 | sync.Pool 等状态存储在结构体内 |

### 2. 错误处理

- 包装错误：`fmt.Errorf("xxx失败: %w", err)`
- 错误信息包含初始数据便于调试

### 3. 性能优化

- 使用 `sync.Pool` 复用 `bytes.Buffer`（存储在 registry 结构体内）
- 避免频繁内存分配
- 请求体大小限制（10MB）

---

## 📐 核心流程

```
AnalyzingBody(body, result)
    │
    ├─► 验证 body 非空
    │
    ├─► 从 registry.bufferPool 获取 buffer
    │
    ├─► 读取请求体到 buffer
    │
    ├─► 检查大小限制（maxBodyLen）
    │
    ├─► json.Unmarshal 解析
    │
    ├─► analyzingJSON 递归解析嵌套 JSON
    │
    └─► 归还 buffer 到 pool
```

### analyzingJSON 递归逻辑

```go
for key, value := range data {
    if str, ok := value.(string); ok {
        // 尝试解析为 JSON 对象
        if json.Unmarshal(str) 成功 {
            data[key] = 解析后的对象
            递归调用 analyzingJSON
        }
    } else if nestedMap, ok := value.(map[string]any); ok {
        递归调用 analyzingJSON(nestedMap)
    }
}
```

---

## 📚 配置说明

| 常量 | 值 | 说明 |
|------|-----|------|
| `maxBodyLen` | 10MB | 请求体最大长度限制 |
| `defaultBufferSize` | 4KB | 缓冲区默认大小 |

---

**版本**：v1.0 | **更新**：2025-01-26
