package emiya

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// registry 实现了 Registry 接口
type registry struct {
	bufferPool sync.Pool
}

// newRegistry 创建新的 registry 实例（私有构造函数）
func newRegistry() *registry {
	return &registry{
		bufferPool: sync.Pool{
			New: func() any {
				buf := make([]byte, 0, defaultBufferSize)
				return bytes.NewBuffer(buf)
			},
		},
	}
}

// AnalyzingBody 分析和解析请求体中的 JSON 数据，支持嵌套 JSON 解析
func (r *registry) AnalyzingBody(body *io.ReadCloser, result *map[string]any) error {
	if body == nil {
		return fmt.Errorf("请求体不能为空")
	}

	// 确保在函数结束时关闭 body
	defer func() {
		if *body != nil {
			(*body).Close()
		}
	}()

	// 从缓冲区池获取缓冲区
	buf := r.getBuffer()
	defer r.putBuffer(buf)

	// 从请求体读取数据
	if _, err := buf.ReadFrom(*body); err != nil {
		return fmt.Errorf("读取请求体失败: %w. 初始数据: %s", err, buf.String())
	}

	// 检查数据长度限制
	if buf.Len() > maxBodyLen {
		return fmt.Errorf("请求体过大，超过最大限制 %d 字节", maxBodyLen)
	}

	// 将请求体解析为 JSON 数据
	if err := json.Unmarshal(buf.Bytes(), result); err != nil {
		return fmt.Errorf("无法解析请求体中的 JSON 数据: %w. 初始数据: %s", err, buf.String())
	}

	// 解析嵌套 JSON 数据
	if err := analyzingJSON(*result); err != nil {
		return fmt.Errorf("嵌套 JSON 数据解析失败: %w. 初始数据: %s", err, buf.String())
	}

	return nil
}
