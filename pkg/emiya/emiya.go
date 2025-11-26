// Package emiya 提供 HTTP 请求体解析功能，支持嵌套 JSON 递归解析
//
// 主要功能：
//   - 解析 HTTP 请求体中的 JSON 数据
//   - 自动递归解析字符串类型的嵌套 JSON
//   - 使用 sync.Pool 优化内存分配
//
// 使用示例：
//
//	registry := emiya.NewRegistry()
//	var data map[string]any
//	if err := registry.AnalyzingBody(&r.Body, &data); err != nil {
//	    // 处理错误
//	}
package emiya

import (
	"io"
)

// Registry 定义了 Emiya 包的主要接口
type Registry interface {
	// AnalyzingBody 分析和解析请求体中的 JSON 数据
	// 支持递归解析字符串类型的嵌套 JSON
	//
	// 参数：
	//   - body: HTTP 请求体指针，解析完成后会自动关闭
	//   - result: 解析结果存储位置
	//
	// 返回：
	//   - error: 解析失败时返回错误信息
	AnalyzingBody(body *io.ReadCloser, result *map[string]any) error
}

// NewRegistry 创建新的 Registry 实例
func NewRegistry() Registry {
	return newRegistry()
}
