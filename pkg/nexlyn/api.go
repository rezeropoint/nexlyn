package plugin_nexlyn

import (
	"net/http"
)

// RegisterHandler 注册HTTP路由处理器 - 使用新的处理器框架
func (n *NexlynPlugin) RegisterHandler() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		// 设备管理接口
		"/api/devices": n.JWTAuthMiddleware(n.createHandler("/api/devices", n.handleDeviceList, &HandlerConfig{
			RequireAuth: true, RequirePluginCheck: true, AllowedMethods: []string{http.MethodGet},
			Resource: "gb28181_device", Action: "read",
		})),
		"/api/devices/bind": n.JWTAuthMiddleware(n.createHandler("/api/devices/bind", n.handleDeviceBind, &HandlerConfig{
			RequireAuth: true, RequirePluginCheck: true, AllowedMethods: []string{http.MethodPost},
			Resource: "gb28181_device", Action: "write",
		})),
		"/api/devices/{deviceId}/unbind": n.JWTAuthMiddleware(n.createHandler("/api/devices/{deviceId}/unbind", n.handleDeviceUnbind, &HandlerConfig{
			RequireAuth: true, RequirePluginCheck: true, AllowedMethods: []string{http.MethodDelete},
			Resource: "gb28181_device", Action: "delete",
		})),
		"/api/devices/{deviceId}/delete": n.JWTAuthMiddleware(n.createHandler("/api/devices/{deviceId}/delete", n.handleDeviceDelete, &HandlerConfig{
			RequireAuth: true, RequirePluginCheck: true, AllowedMethods: []string{http.MethodDelete},
			Resource: "gb28181_device", Action: "delete",
		})),
		"/api/devices/{deviceId}": n.JWTAuthMiddleware(n.createHandler("/api/devices/{deviceId}", n.handleDeviceDetail, &HandlerConfig{
			RequireAuth: true, RequirePluginCheck: true, AllowedMethods: []string{http.MethodGet},
			Resource: "gb28181_device", Action: "read",
		})),
		"/api/devices/{deviceId}/alias": n.JWTAuthMiddleware(n.createHandler("/api/devices/{deviceId}/alias", n.handleDeviceAlias, &HandlerConfig{
			RequireAuth: true, RequirePluginCheck: true, AllowedMethods: []string{http.MethodPut},
			Resource: "gb28181_device", Action: "write",
		})),
		"/api/devices/{deviceId}/tags": n.JWTAuthMiddleware(n.createHandler("/api/devices/{deviceId}/tags", n.handleDeviceTags, &HandlerConfig{
			RequireAuth: true, RequirePluginCheck: true, AllowedMethods: []string{http.MethodPut},
			Resource: "gb28181_device", Action: "write",
		})),

		"/api/records/{deviceId}/{channelId}": n.JWTAuthMiddleware(n.createHandler("/api/records/{deviceId}/{channelId}", n.handleRecordQuery, &HandlerConfig{
			RequireAuth: true, RequirePluginCheck: true, AllowedMethods: []string{http.MethodGet},
			Resource: "gb28181_device", Action: "read",
		})),

		// 标签管理接口
		"/api/tags": n.JWTAuthMiddleware(n.createHandler("/api/tags", n.handleTagList, &HandlerConfig{
			RequireAuth: true, RequirePluginCheck: true, AllowedMethods: []string{http.MethodGet},
			Resource: "gb28181_tag", Action: "read",
		})),
		"/api/tags/create": n.JWTAuthMiddleware(n.createHandler("/api/tags/create", n.handleTagCreate, &HandlerConfig{
			RequireAuth: true, RequirePluginCheck: true, AllowedMethods: []string{http.MethodPost},
			Resource: "gb28181_tag", Action: "write",
		})),
		"/api/tags/{tagId}/update": n.JWTAuthMiddleware(n.createHandler("/api/tags/{tagId}/update", n.handleTagUpdate, &HandlerConfig{
			RequireAuth: true, RequirePluginCheck: true, AllowedMethods: []string{http.MethodPut},
			Resource: "gb28181_tag", Action: "write",
		})),
		"/api/tags/{tagId}/delete": n.JWTAuthMiddleware(n.createHandler("/api/tags/{tagId}/delete", n.handleTagDelete, &HandlerConfig{
			RequireAuth: true, RequirePluginCheck: true, AllowedMethods: []string{http.MethodDelete},
			Resource: "gb28181_tag", Action: "delete",
		})),

		// 统计接口
		"/api/statistics/devices": n.JWTAuthMiddleware(n.createHandler("/api/statistics/devices", n.handleGetDeviceStatistics, &HandlerConfig{
			RequireAuth: true, RequirePluginCheck: true, AllowedMethods: []string{http.MethodGet},
			Resource: "gb28181_stats", Action: "read",
		})),

		// 测试接口（无权限要求）
		"/api/test": n.SimpleHandler("/api/test", n.handleTest),
		"/ping":     n.NoAuthHandler("/ping", n.handlePing),
	}
}
