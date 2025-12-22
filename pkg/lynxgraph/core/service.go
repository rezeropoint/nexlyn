package core

import (
	"fmt"
	"sync"
)

// ServiceType 服务类型定义
type ServiceType string

// 预定义服务类型常量
const (
	ServiceTypeSkylarkEngine  ServiceType = "skylark_engine"   // Skylark流程引擎
	ServiceTypeSensorData     ServiceType = "sensor_data"      // 传感器历史数据查询服务
	ServiceTypeExternalDB     ServiceType = "external_db"      // 外部数据库查询服务（用于QueryDatabase积木）
	ServiceTypeDispatcher     ServiceType = "dispatcher"       // 图调度器服务（用于ForEach积木调用子图）
	ServiceTypeInfoAtomQuery  ServiceType = "info_atom_query"  // 信息原子类型查询服务（用于ForEach创建信息原子）
)

type Service interface {
	Register(serviceType ServiceType, service interface{}) error // 注册服务（每个类型只能有一个）
	GetByType(serviceType ServiceType) (interface{}, error)      // 通过类型获取服务
	Unregister(serviceType ServiceType) error                    // 注销服务
	HasServiceType(serviceType ServiceType) bool                 // 检查是否有该类型的服务
}

type BasicService struct {
	mu       sync.RWMutex
	services map[ServiceType]interface{} // 服务类型到服务实例的映射（每个类型唯一）
}

func NewService() Service {
	return &BasicService{
		services: make(map[ServiceType]interface{}),
	}
}

// Register 注册一个服务（每个类型只能注册一个）
func (s *BasicService) Register(serviceType ServiceType, service interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.services[serviceType]; exists {
		return fmt.Errorf("服务类型 '%s' 已注册", serviceType)
	}

	s.services[serviceType] = service
	return nil
}

// GetByType 通过服务类型获取服务实例
func (s *BasicService) GetByType(serviceType ServiceType) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	service, exists := s.services[serviceType]
	if !exists {
		return nil, fmt.Errorf("服务类型 '%s' 未注册", serviceType)
	}

	return service, nil
}

// Unregister 注销一个服务
func (s *BasicService) Unregister(serviceType ServiceType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.services[serviceType]; !exists {
		return fmt.Errorf("服务类型 '%s' 不存在", serviceType)
	}

	delete(s.services, serviceType)
	return nil
}

// HasServiceType 检查是否有指定类型的服务已注册
func (s *BasicService) HasServiceType(serviceType ServiceType) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.services[serviceType]
	return exists
}

// ============================================================================
// 服务接口说明
// ============================================================================
//
// LynxGraph Core 不定义具体业务接口，只提供服务注册机制。
//
// 具体业务接口由外部系统定义：
//   - Skylark: 使用 github.com/rezeropoint/go-skylark/v2/engine.SkylarkEngine
//   - SensorData: 使用 nexlyn/service/iotquery/iotquery.IoTQuery (gRPC生成的接口)
//
// 积木使用服务时，直接导入外部接口并类型断言：
//
//   import "github.com/rezeropoint/go-skylark/v2/engine"
//   skylarkEngine, ok := service.GetByType(ServiceTypeSkylarkEngine).(engine.SkylarkEngine)
//
//   import "github.com/rezeropoint/nexlyn/service/iotquery/iotquery"
//   iotQueryClient, ok := service.GetByType(ServiceTypeSensorData).(iotquery.IoTQuery)
//
// ============================================================================
