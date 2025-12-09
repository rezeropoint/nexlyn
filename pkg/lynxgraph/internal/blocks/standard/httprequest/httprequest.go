package httprequest

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/zeromicro/go-zero/core/logx"
)

const HttpRequestVersion = "v1"

// Config HTTP 请求积木的配置
type Config struct {
	Method             string            `json:"method" check:"must"`  // HTTP 方法: GET/POST/PUT/DELETE
	URL                string            `json:"url" check:"must"`     // 请求 URL，支持变量替换
	Headers            map[string]string `json:"headers"`              // 请求头
	Body               string            `json:"body"`                 // 请求体，支持变量替换
	Timeout            int               `json:"timeout"`              // 超时（毫秒），默认 5000
	SuccessStatusCodes []int             `json:"successStatusCodes"`   // 视为成功的状态码列表，默认 200-299
	IgnoreError        bool              `json:"ignoreError"`          // 忽略错误继续执行
	SkipTLSVerify      bool              `json:"skipTLSVerify"`        // 跳过 HTTPS 证书验证
}

// HttpRequestBlock 实现 HTTP 请求逻辑块
type HttpRequestBlock struct {
	core.BaseLogicBlock
}

// NewHttpRequestBlock 创建一个新的 HttpRequestBlock 实例
func NewHttpRequestBlock(id string, config map[string]any) (core.LogicBlock, error) {
	return &HttpRequestBlock{
		BaseLogicBlock: core.BaseLogicBlock{
			ID:        id,
			Type:      core.BlockTypeHttpRequest,
			RawConfig: config,
		},
	}, nil
}

func (b *HttpRequestBlock) GetID() string                 { return b.ID }
func (b *HttpRequestBlock) GetType() core.LogicBlockType  { return b.Type }
func (b *HttpRequestBlock) GetConfigure() map[string]any  { return b.RawConfig }

func (b *HttpRequestBlock) SetConfigure(config map[string]any) error {
	var cfg Config
	if err := core.FillConfig(config, &cfg); err != nil {
		return err
	}

	// 验证 HTTP 方法
	method := strings.ToUpper(cfg.Method)
	if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" {
		return fmt.Errorf("不支持的 HTTP 方法: %s (仅支持 GET/POST/PUT/DELETE)", cfg.Method)
	}
	cfg.Method = method

	// 设置默认超时
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5000
	}

	b.TypedConfig = &cfg
	return nil
}

// Execute 执行 HTTP 请求
func (b *HttpRequestBlock) Execute(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store, service core.Service) (bool, error) {
	config, ok := b.TypedConfig.(*Config)
	if !ok {
		return false, core.ErrInvalidConfig
	}

	// 获取 InfoAtom Payload 用于变量替换
	payload := execCtx.GetInfoAtom().GetPayload()

	// 替换 URL 中的变量
	url := replaceVariables(config.URL, payload)

	// 替换 Body 中的变量
	body := replaceVariables(config.Body, payload)

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Millisecond,
	}

	// 配置 TLS
	if config.SkipTLSVerify {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// 创建请求
	var reqBody io.Reader
	if body != "" {
		reqBody = bytes.NewBufferString(body)
	}

	req, err := http.NewRequestWithContext(ctx, config.Method, url, reqBody)
	if err != nil {
		return b.handleError(config, err, "创建请求失败")
	}

	// 设置请求头
	for key, value := range config.Headers {
		req.Header.Set(key, replaceVariables(value, payload))
	}

	// 如果有 Body 且没有设置 Content-Type，默认设置为 JSON
	if body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 发送请求
	logx.WithContext(ctx).Infof("[HttpRequest] 发送请求: %s %s", config.Method, url)

	resp, err := client.Do(req)
	if err != nil {
		return b.handleError(config, err, "发送请求失败")
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return b.handleError(config, err, "读取响应失败")
	}

	logx.WithContext(ctx).Infof("[HttpRequest] 响应状态: %d, 响应体: %s", resp.StatusCode, string(respBody))

	// 检查响应状态码
	if !isSuccessStatusCode(resp.StatusCode, config.SuccessStatusCodes) {
		err := fmt.Errorf("HTTP 状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
		return b.handleError(config, err, "请求返回非成功状态")
	}

	return true, nil
}

// isSuccessStatusCode 检查状态码是否在成功列表中
func isSuccessStatusCode(statusCode int, successCodes []int) bool {
	// 如果未配置，默认 200-299 为成功
	if len(successCodes) == 0 {
		return statusCode >= 200 && statusCode < 300
	}
	// 检查是否在配置的成功状态码列表中
	for _, code := range successCodes {
		if statusCode == code {
			return true
		}
	}
	return false
}

// handleError 处理错误，根据 IgnoreError 配置决定返回值
func (b *HttpRequestBlock) handleError(config *Config, err error, msg string) (bool, error) {
	fullErr := fmt.Errorf("%s: %w", msg, err)
	if config.IgnoreError {
		logx.Errorf("[HttpRequest] %s (已忽略): %v", msg, err)
		return true, nil
	}
	return false, fullErr
}

// replaceVariables 替换字符串中的变量 {{atom.xxx}}
func replaceVariables(template string, payload map[string]any) string {
	if template == "" {
		return ""
	}

	// 匹配 {{atom.xxx}} 格式的变量
	re := regexp.MustCompile(`\{\{atom\.(\w+)\}\}`)
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		// 提取字段名
		fieldName := re.FindStringSubmatch(match)[1]
		if value, ok := payload[fieldName]; ok {
			return fmt.Sprintf("%v", value)
		}
		return match // 未找到则保留原样
	})

	return result
}

// GetHttpRequestSpec 返回 HttpRequestBlock 的规格
func GetHttpRequestSpec() core.BlockSpec {
	return core.NewBasicBlockSpec(
		"HttpRequest",
		HttpRequestVersion,
		"发送 HTTP 请求到外部系统，支持 GET/POST/PUT/DELETE 方法",
		[]string{"action", "http", "webhook"},
		map[string]any{
			"method": map[string]any{
				"type":        "string",
				"title":       "HTTP 方法",
				"description": "请求方法",
				"check":       "must",
				"enum":        []string{"GET", "POST", "PUT", "DELETE"},
			},
			"url": map[string]any{
				"type":        "string",
				"title":       "请求 URL",
				"description": "目标 URL，支持 {{atom.xxx}} 变量替换",
				"placeholder": "https://example.com/api/webhook",
				"check":       "must",
			},
			"headers": map[string]any{
				"type":        "object",
				"title":       "请求头",
				"description": "HTTP 请求头，键值对格式",
				"check":       "",
			},
			"body": map[string]any{
				"type":        "string",
				"title":       "请求体",
				"description": "POST/PUT 请求的内容，支持 {{atom.xxx}} 变量替换",
				"placeholder": `{"action": "open_door", "person": "{{atom.person_name}}"}`,
				"check":       "",
			},
			"timeout": map[string]any{
				"type":        "integer",
				"title":       "超时时间",
				"description": "请求超时（毫秒），默认 5000",
				"check":       "",
			},
			"successStatusCodes": map[string]any{
				"type":        "array",
				"title":       "成功状态码",
				"description": "视为成功的 HTTP 状态码列表，默认 200-299",
				"placeholder": "[200, 201, 204]",
				"check":       "",
				"items": map[string]any{
					"type": "integer",
				},
			},
			"ignoreError": map[string]any{
				"type":        "boolean",
				"title":       "忽略错误",
				"description": "请求失败时是否继续执行后续节点",
				"check":       "",
			},
			"skipTLSVerify": map[string]any{
				"type":        "boolean",
				"title":       "跳过证书验证",
				"description": "跳过 HTTPS 证书验证（用于自签名证书）",
				"check":       "",
			},
		},
		[]core.ServiceType{}, // 不依赖任何外部服务
	)
}
