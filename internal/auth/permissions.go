package auth

// init 初始化权限注册
// 在应用启动时自动注册所有业务权限到全局注册中心
func init() {
	// ============ IoT 资源权限 ============
	// IoT设备模板
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTTemplate,
		Action:      "read",
		Description: "设备模板查看",
		Category:    "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTTemplate,
		Action:      "write",
		Description: "设备模板创建/修改",
		Category:    "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTTemplate,
		Action:      "delete",
		Description: "设备模板删除",
		Category:    "物联管理",
	})

	// IoT设备
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTDevice,
		Action:      "read",
		Description: "物联设备查看",
		Category:    "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTDevice,
		Action:      "write",
		Description: "物联设备创建/修改",
		Category:    "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTDevice,
		Action:      "delete",
		Description: "物联设备删除",
		Category:    "物联管理",
	})

	// IoT标签
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTTag,
		Action:      "read",
		Description: "物联标签查看",
		Category:    "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTTag,
		Action:      "write",
		Description: "物联标签创建/修改",
		Category:    "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTTag,
		Action:      "delete",
		Description: "物联标签删除",
		Category:    "物联管理",
	})

	// IoT元数据
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTMetadata,
		Action:      "read",
		Description: "元数据查看",
		Category:    "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTMetadata,
		Action:      "write",
		Description: "元数据创建/修改",
		Category:    "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTMetadata,
		Action:      "delete",
		Description: "元数据删除",
		Category:    "物联管理",
	})

	// IoT平台配置
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTPlatform,
		Action:      "read",
		Description: "平台配置查看",
		Category:    "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTPlatform,
		Action:      "write",
		Description: "平台配置创建/修改",
		Category:    "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTPlatform,
		Action:      "delete",
		Description: "平台配置删除",
		Category:    "物联管理",
	})

	// IoT HTTP数据接收配置
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTHttpReceive,
		Action:      "read",
		Description: "HTTP接收配置查看",
		Category:    "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTHttpReceive,
		Action:      "write",
		Description: "HTTP接收配置创建/修改",
		Category:    "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceIoTHttpReceive,
		Action:      "delete",
		Description: "HTTP接收配置删除",
		Category:    "物联管理",
	})

	// ============ GB28181 资源权限 ============
	// GB28181设备
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceGB28181Device,
		Action:      "read",
		Description: "视频设备查看",
		Category:    "视频管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceGB28181Device,
		Action:      "write",
		Description: "视频设备创建/修改",
		Category:    "视频管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceGB28181Device,
		Action:      "delete",
		Description: "视频设备删除",
		Category:    "视频管理",
	})

	// GB28181流
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceGB28181Stream,
		Action:      "read",
		Description: "视频流查看",
		Category:    "视频管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceGB28181Stream,
		Action:      "write",
		Description: "视频流创建/修改",
		Category:    "视频管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceGB28181Stream,
		Action:      "delete",
		Description: "视频流删除",
		Category:    "视频管理",
	})

	// GB28181标签
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceGB28181Tag,
		Action:      "read",
		Description: "视频标签查看",
		Category:    "视频管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceGB28181Tag,
		Action:      "write",
		Description: "视频标签创建/修改",
		Category:    "视频管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceGB28181Tag,
		Action:      "delete",
		Description: "视频标签删除",
		Category:    "视频管理",
	})

	// GB28181统计
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceGB28181Stats,
		Action:      "read",
		Description: "视频统计查看",
		Category:    "视频管理",
	})

	// ============ Skylark事件管理资源权限 ============
	// Skylark平台配置
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceSkylarkPlatform,
		Action:      "read",
		Description: "Skylark平台配置查看",
		Category:    "Skylark事件管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceSkylarkPlatform,
		Action:      "write",
		Description: "Skylark平台配置创建/修改",
		Category:    "Skylark事件管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceSkylarkPlatform,
		Action:      "delete",
		Description: "Skylark平台配置删除",
		Category:    "Skylark事件管理",
	})

	// 事件配置
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceEventConfig,
		Action:      "read",
		Description: "事件配置查看",
		Category:    "事件配置管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceEventConfig,
		Action:      "write",
		Description: "事件配置创建/修改",
		Category:    "事件配置管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceEventConfig,
		Action:      "delete",
		Description: "事件配置删除",
		Category:    "事件配置管理",
	})

	// 事件数据查询
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceEventData,
		Action:      "read",
		Description: "事件数据查询",
		Category:    "事件数据查询",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceEventData,
		Action:      "write",
		Description: "事件数据创建/修改",
		Category:    "事件数据查询",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceEventData,
		Action:      "delete",
		Description: "事件数据删除",
		Category:    "事件数据查询",
	})

	// 组织映射
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceOrgMapping,
		Action:      "read",
		Description: "组织映射查看",
		Category:    "组织映射管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceOrgMapping,
		Action:      "write",
		Description: "组织映射创建/修改",
		Category:    "组织映射管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceOrgMapping,
		Action:      "delete",
		Description: "组织映射删除",
		Category:    "组织映射管理",
	})

	// 流程Journey操作
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceFlowJourney,
		Action:      "read",
		Description: "流程审批查看",
		Category:    "流程管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceFlowJourney,
		Action:      "write",
		Description: "流程审批操作",
		Category:    "流程管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceFlowJourney,
		Action:      "delete",
		Description: "流程审批删除",
		Category:    "流程管理",
	})

	// 流程管理
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceFlow,
		Action:      "read",
		Description: "流程查询",
		Category:    "流程管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceFlow,
		Action:      "write",
		Description: "流程创建",
		Category:    "流程管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceFlow,
		Action:      "delete",
		Description: "流程终止",
		Category:    "流程管理",
	})

	// 表单管理
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceForm,
		Action:      "write",
		Description: "表单提交",
		Category:    "表单管理",
	})

	// ============ LynxGraph逻辑引擎资源权限 ============
	// 信息原子类型
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceLynxInfoAtomType,
		Action:      "read",
		Description: "信息原子类型查看",
		Category:    "LynxGraph逻辑引擎",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceLynxInfoAtomType,
		Action:      "write",
		Description: "信息原子类型创建/修改",
		Category:    "LynxGraph逻辑引擎",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceLynxInfoAtomType,
		Action:      "delete",
		Description: "信息原子类型删除",
		Category:    "LynxGraph逻辑引擎",
	})

	// 逻辑图配置
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceLynxGraphConfig,
		Action:      "read",
		Description: "逻辑图配置查看",
		Category:    "LynxGraph逻辑引擎",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceLynxGraphConfig,
		Action:      "write",
		Description: "逻辑图配置创建/修改",
		Category:    "LynxGraph逻辑引擎",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceLynxGraphConfig,
		Action:      "delete",
		Description: "逻辑图配置删除",
		Category:    "LynxGraph逻辑引擎",
	})

	// 逻辑块规格（只读）
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceLynxBlockSpec,
		Action:      "read",
		Description: "逻辑块规格查看",
		Category:    "LynxGraph逻辑引擎",
	})

	// Lynx标签
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceLynxTag,
		Action:      "read",
		Description: "Lynx标签查看",
		Category:    "LynxGraph逻辑引擎",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceLynxTag,
		Action:      "write",
		Description: "Lynx标签创建/修改",
		Category:    "LynxGraph逻辑引擎",
	})
	RegisterPermission(PermissionMetadata{
		Resource:    ResourceLynxTag,
		Action:      "delete",
		Description: "Lynx标签删除",
		Category:    "LynxGraph逻辑引擎",
	})
}
