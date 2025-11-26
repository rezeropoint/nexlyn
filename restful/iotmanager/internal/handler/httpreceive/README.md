# httpreceive Handler 说明

## 特殊说明

`receivedatahandler.go` 是手动编写的 Handler，**不使用 goctl 生成的标准模式**。

### 为什么不使用 httpx.Parse

标准的 goctl 生成代码使用 `httpx.Parse(r, &req)` 解析请求，但该方法会消耗 `r.Body`，导致后续无法再次读取请求体。

对于 HTTP 数据接收接口，需要：
1. 支持嵌套 JSON 解析（字符串类型的 JSON 会被递归解析为对象）
2. 保持请求体的完整性

因此使用 `EmiyaRegistry.AnalyzingBody` 替代标准解析。

### 代码模式

```go
func ReceiveDataHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. 使用 EmiyaRegistry 解析请求体（支持嵌套 JSON）
        var data map[string]any
        if err := svcCtx.EmiyaRegistry.AnalyzingBody(&r.Body, &data); err != nil {
            httpx.ErrorCtx(r.Context(), w, fmt.Errorf("解析请求体失败: %v", err))
            logx.WithContext(r.Context()).WithFields(...).Error("解析请求体失败")
            return
        }

        // 2. 获取路径参数并调用 Logic 层
        l := httpreceive.NewReceiveDataLogic(r.Context(), svcCtx)
        resp, err := l.ReceiveData(pathvar.Vars(r)["configId"], data)
        // ...
    }
}
```

### 注意事项

- 重新运行 `goctl api go` 会覆盖此文件，需要手动恢复
- 路径参数使用 `pathvar.Vars(r)["configId"]` 获取，不会消耗 Body
- 错误处理顺序：先 `httpx.ErrorCtx` 返回客户端，再 `logx.Error` 记录日志
