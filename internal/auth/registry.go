package auth

import (
	"sync"
)

// PermissionMetadata 权限元数据
type PermissionMetadata struct {
	Resource     string // 资源标识
	ResourceName string // 资源中文名
	Action       string // 操作
	Description  string // 描述
	Category     string // 分类
}

// PermissionRegistry 权限注册中心
type PermissionRegistry struct {
	mu          sync.RWMutex
	permissions map[string]PermissionMetadata // key: resource:action
}

// globalRegistry 全局权限注册中心实例
var globalRegistry = &PermissionRegistry{
	permissions: make(map[string]PermissionMetadata),
}

// RegisterPermission 注册权限
// key 格式: "resource:action" 例如: "iot_template:read"
func RegisterPermission(p PermissionMetadata) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	key := p.Resource + ":" + p.Action
	globalRegistry.permissions[key] = p
}

// GetAllPermissions 获取所有权限（按资源和操作排序）
func GetAllPermissions() []PermissionMetadata {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	permissions := make([]PermissionMetadata, 0, len(globalRegistry.permissions))
	for _, p := range globalRegistry.permissions {
		permissions = append(permissions, p)
	}
	return permissions
}

// GetPermissionsByResource 按资源获取权限
func GetPermissionsByResource(resource string) []PermissionMetadata {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	permissions := make([]PermissionMetadata, 0)
	for _, p := range globalRegistry.permissions {
		if p.Resource == resource {
			permissions = append(permissions, p)
		}
	}
	return permissions
}

// GetPermissionsByCategory 按分类获取权限
func GetPermissionsByCategory(category string) []PermissionMetadata {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	permissions := make([]PermissionMetadata, 0)
	for _, p := range globalRegistry.permissions {
		if p.Category == category {
			permissions = append(permissions, p)
		}
	}
	return permissions
}

// GetPermissionCount 获取权限总数
func GetPermissionCount() int {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	return len(globalRegistry.permissions)
}
