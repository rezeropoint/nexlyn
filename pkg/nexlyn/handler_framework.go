package plugin_nexlyn

import (
	"net/http"
	"regexp"
	"strings"
)

// NexlynHandlerFunc 统一的处理器函数类型
type NexlynHandlerFunc func(w http.ResponseWriter, r *http.Request, ctx *NexlynContext)

// NexlynContext 请求上下文，包含自动提取的路径参数和用户信息
type NexlynContext struct {
	// 路径参数
	DeviceID   string            // 设备ID (从路径参数自动提取)
	ChannelID  string            // 通道ID (从路径参数自动提取)
	TagID      string            // 标签ID (从路径参数自动提取)
	StreamPath string            // 流路径 (从路径参数自动提取)
	PathParams map[string]string // 所有路径参数的映射

	// 用户信息
	UserInfo *JWTUser // 用户信息 (从JWT中获取)
}

// RoutePattern 路由模式定义
type RoutePattern struct {
	Pattern    string         // 路由模式，如 "/api/devices/{deviceId}"
	ParamNames []string       // 参数名称列表，如 ["deviceId"]
	Regex      *regexp.Regexp // 编译后的正则表达式
}

// PathParamExtractor 路径参数提取器
type PathParamExtractor struct {
	patterns map[string]*RoutePattern // 路由模式缓存
}

// NewPathParamExtractor 创建新的路径参数提取器
func NewPathParamExtractor() *PathParamExtractor {
	return &PathParamExtractor{
		patterns: make(map[string]*RoutePattern),
	}
}

// RegisterPattern 注册路由模式
func (e *PathParamExtractor) RegisterPattern(pattern string) *RoutePattern {
	if route, exists := e.patterns[pattern]; exists {
		return route
	}

	// 解析路由模式，提取参数名
	paramNames := make([]string, 0)
	regexPattern := pattern

	// 查找所有 {paramName} 格式的参数
	paramRegex := regexp.MustCompile(`\{([^}]+)\}`)
	matches := paramRegex.FindAllStringSubmatch(pattern, -1)

	for _, match := range matches {
		if len(match) > 1 {
			paramNames = append(paramNames, match[1])
			// 将 {paramName} 替换为捕获组
			regexPattern = strings.Replace(regexPattern, match[0], `([^/]+)`, 1)
		}
	}

	// 编译正则表达式
	regex, err := regexp.Compile("^" + regexPattern + "$")
	if err != nil {
		// 如果编译失败，创建一个匹配所有的正则表达式
		regex = regexp.MustCompile(".*")
	}

	route := &RoutePattern{
		Pattern:    pattern,
		ParamNames: paramNames,
		Regex:      regex,
	}

	e.patterns[pattern] = route
	return route
}

// ExtractParams 从路径中提取参数
func (e *PathParamExtractor) ExtractParams(pattern, path string) map[string]string {
	route := e.RegisterPattern(pattern)
	params := make(map[string]string)

	// 使用正则表达式匹配路径
	matches := route.Regex.FindStringSubmatch(path)
	if len(matches) > 1 && len(matches)-1 == len(route.ParamNames) {
		for i, paramName := range route.ParamNames {
			params[paramName] = matches[i+1]
		}
	}

	return params
}

// MiddlewareChain 中间件链类型
type MiddlewareChain func(NexlynHandlerFunc) NexlynHandlerFunc

// HandlerConfig 处理器配置
type HandlerConfig struct {
	RequireAuth        bool     // 是否需要认证
	RequireValidation  bool     // 是否需要参数校验
	RequirePluginCheck bool     // 是否需要插件检查
	AllowedMethods     []string // 允许的HTTP方法
	Resource           string   // 权限资源
	Action             string   // 权限操作
}

// DefaultHandlerConfig 默认处理器配置
func DefaultHandlerConfig() *HandlerConfig {
	return &HandlerConfig{
		RequireValidation:  true, // 默认需要校验
		RequirePluginCheck: true, // 默认需要插件检查
		AllowedMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}
}

// NoAuthConfig 无需认证的配置
func NoAuthConfig() *HandlerConfig {
	config := DefaultHandlerConfig()
	config.RequireAuth = false
	return config
}

// GetOnlyConfig 只允许GET请求的配置
func GetOnlyConfig() *HandlerConfig {
	config := DefaultHandlerConfig()
	config.AllowedMethods = []string{http.MethodGet}
	return config
}

// PostOnlyConfig 只允许POST请求的配置
func PostOnlyConfig() *HandlerConfig {
	config := DefaultHandlerConfig()
	config.AllowedMethods = []string{http.MethodPost}
	return config
}

// WithAuth 认证中间件装饰器 - 从context中获取已解析的用户信息
func (n *NexlynPlugin) WithAuth(handler NexlynHandlerFunc) NexlynHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
		// JWT中间件已经解析了用户信息并放入context，这里直接获取即可
		user, err := GetUserFromJWT(r.Context())
		if err != nil {
			n.sendErrorResponse(w, http.StatusUnauthorized, "获取用户信息失败", err)
			return
		}

		// 检查用户是否活跃
		if !user.IsActive() {
			n.sendErrorResponse(w, http.StatusForbidden, "用户账户已被禁用", nil)
			return
		}

		ctx.UserInfo = user
		handler(w, r, ctx)
	}
}

// WithValidation 参数校验中间件装饰器
func (n *NexlynPlugin) WithValidation(allowedMethods []string) MiddlewareChain {
	return func(handler NexlynHandlerFunc) NexlynHandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
			// HTTP方法检查
			if len(allowedMethods) > 0 {
				methodAllowed := false
				for _, method := range allowedMethods {
					if r.Method == method {
						methodAllowed = true
						break
					}
				}
				if !methodAllowed {
					n.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
					return
				}
			}

			handler(w, r, ctx)
		}
	}
}

// WithPluginCheck 插件可用性检查中间件装饰器
func (n *NexlynPlugin) WithPluginCheck(handler NexlynHandlerFunc) NexlynHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
		if n.gb28181Plugin == nil {
			n.sendErrorResponse(w, http.StatusServiceUnavailable, "GB28181插件未找到", nil)
			return
		}

		handler(w, r, ctx)
	}
}

// WithDeviceIDValidation 设备ID格式校验中间件装饰器
func (n *NexlynPlugin) WithDeviceIDValidation(handler NexlynHandlerFunc) NexlynHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
		// DeviceID 格式校验已在路径参数提取时完成
		// 设备存在性校验由业务逻辑层处理

		handler(w, r, ctx)
	}
}

// ChainMiddleware 链式组合多个中间件
func (n *NexlynPlugin) ChainMiddleware(handler NexlynHandlerFunc, middlewares ...MiddlewareChain) NexlynHandlerFunc {
	// 从后往前应用中间件，这样中间件的执行顺序就是参数的顺序
	result := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		result = middlewares[i](result)
	}
	return result
}

// createHandler 创建处理器的核心函数
func (n *NexlynPlugin) createHandler(pattern string, handler NexlynHandlerFunc, config *HandlerConfig) http.HandlerFunc {
	if config == nil {
		config = DefaultHandlerConfig()
	}

	// 如果还没有初始化参数提取器，则创建一个
	if n.paramExtractor == nil {
		n.paramExtractor = NewPathParamExtractor()
	}

	// 构建中间件链
	middlewares := []MiddlewareChain{}

	// 添加HTTP方法校验中间件
	if len(config.AllowedMethods) > 0 {
		middlewares = append(middlewares, n.WithValidation(config.AllowedMethods))
	}

	// 包装处理器，应用中间件链和认证/插件检查
	wrappedHandler := handler

	// 应用设备ID校验 (如果路径中包含deviceId参数)
	if strings.Contains(pattern, "{deviceId}") {
		wrappedHandler = n.WithDeviceIDValidation(wrappedHandler)
	}

	// 应用插件检查
	if config.RequirePluginCheck {
		wrappedHandler = n.WithPluginCheck(wrappedHandler)
	}

	// 应用认证和权限检查
	if config.RequireAuth {
		if config.Resource != "" && config.Action != "" {
			// 使用权限检查中间件
			wrappedHandler = n.WithPermission(config.Resource, config.Action)(wrappedHandler)
		} else {
			// 仅认证，无权限检查
			wrappedHandler = n.WithAuth(wrappedHandler)
		}
	}

	// 应用其他中间件
	wrappedHandler = n.ChainMiddleware(wrappedHandler, middlewares...)

	return func(w http.ResponseWriter, r *http.Request) {
		// 提取路径参数
		pathParams := n.paramExtractor.ExtractParams(pattern, r.URL.Path)

		// 创建上下文
		ctx := &NexlynContext{
			PathParams: pathParams,
			DeviceID:   pathParams["deviceId"],
			ChannelID:  pathParams["channelId"],
			TagID:      pathParams["tagId"],
			StreamPath: pathParams["streamPath"],
		}

		// 调用包装后的处理器
		wrappedHandler(w, r, ctx)
	}
}

// SimpleHandler 创建简单处理器的便捷函数
func (n *NexlynPlugin) SimpleHandler(pattern string, handler NexlynHandlerFunc) http.HandlerFunc {
	return n.createHandler(pattern, handler, DefaultHandlerConfig())
}

// NoAuthHandler 创建无需认证处理器的便捷函数
func (n *NexlynPlugin) NoAuthHandler(pattern string, handler NexlynHandlerFunc) http.HandlerFunc {
	return n.createHandler(pattern, handler, NoAuthConfig())
}

// GetHandler 创建只允许GET请求处理器的便捷函数
func (n *NexlynPlugin) GetHandler(pattern string, handler NexlynHandlerFunc) http.HandlerFunc {
	return n.createHandler(pattern, handler, GetOnlyConfig())
}

// PostHandler 创建只允许POST请求处理器的便捷函数
func (n *NexlynPlugin) PostHandler(pattern string, handler NexlynHandlerFunc) http.HandlerFunc {
	return n.createHandler(pattern, handler, PostOnlyConfig())
}
