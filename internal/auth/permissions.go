package auth

// init 初始化权限注册
// 在应用启动时自动注册所有业务权限到全局注册中心
func init() {
	// ============ IoT 资源权限 ============
	// IoT设备模板
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTTemplate,
		ResourceName: "设备模板",
		Action:       "read",
		Description:  "设备模板查看",
		Category:     "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTTemplate,
		ResourceName: "设备模板",
		Action:       "write",
		Description:  "设备模板创建/修改",
		Category:     "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTTemplate,
		ResourceName: "设备模板",
		Action:       "delete",
		Description:  "设备模板删除",
		Category:     "物联管理",
	})

	// IoT设备
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTDevice,
		ResourceName: "物联设备",
		Action:       "read",
		Description:  "物联设备查看",
		Category:     "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTDevice,
		ResourceName: "物联设备",
		Action:       "write",
		Description:  "物联设备创建/修改",
		Category:     "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTDevice,
		ResourceName: "物联设备",
		Action:       "delete",
		Description:  "物联设备删除",
		Category:     "物联管理",
	})

	// IoT标签
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTTag,
		ResourceName: "物联标签",
		Action:       "read",
		Description:  "物联标签查看",
		Category:     "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTTag,
		ResourceName: "物联标签",
		Action:       "write",
		Description:  "物联标签创建/修改",
		Category:     "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTTag,
		ResourceName: "物联标签",
		Action:       "delete",
		Description:  "物联标签删除",
		Category:     "物联管理",
	})

	// IoT元数据
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTMetadata,
		ResourceName: "元数据",
		Action:       "read",
		Description:  "元数据查看",
		Category:     "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTMetadata,
		ResourceName: "元数据",
		Action:       "write",
		Description:  "元数据创建/修改",
		Category:     "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTMetadata,
		ResourceName: "元数据",
		Action:       "delete",
		Description:  "元数据删除",
		Category:     "物联管理",
	})

	// IoT平台配置
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTPlatform,
		ResourceName: "平台配置",
		Action:       "read",
		Description:  "平台配置查看",
		Category:     "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTPlatform,
		ResourceName: "平台配置",
		Action:       "write",
		Description:  "平台配置创建/修改",
		Category:     "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTPlatform,
		ResourceName: "平台配置",
		Action:       "delete",
		Description:  "平台配置删除",
		Category:     "物联管理",
	})

	// IoT HTTP数据接收配置
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTHttpReceive,
		ResourceName: "HTTP接收配置",
		Action:       "read",
		Description:  "HTTP接收配置查看",
		Category:     "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTHttpReceive,
		ResourceName: "HTTP接收配置",
		Action:       "write",
		Description:  "HTTP接收配置创建/修改",
		Category:     "物联管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceIoTHttpReceive,
		ResourceName: "HTTP接收配置",
		Action:       "delete",
		Description:  "HTTP接收配置删除",
		Category:     "物联管理",
	})

	// ============ GB28181 资源权限 ============
	// GB28181设备
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceGB28181Device,
		ResourceName: "视频设备",
		Action:       "read",
		Description:  "视频设备查看",
		Category:     "视频管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceGB28181Device,
		ResourceName: "视频设备",
		Action:       "write",
		Description:  "视频设备创建/修改",
		Category:     "视频管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceGB28181Device,
		ResourceName: "视频设备",
		Action:       "delete",
		Description:  "视频设备删除",
		Category:     "视频管理",
	})

	// GB28181流
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceGB28181Stream,
		ResourceName: "视频流",
		Action:       "read",
		Description:  "视频流查看",
		Category:     "视频管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceGB28181Stream,
		ResourceName: "视频流",
		Action:       "write",
		Description:  "视频流创建/修改",
		Category:     "视频管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceGB28181Stream,
		ResourceName: "视频流",
		Action:       "delete",
		Description:  "视频流删除",
		Category:     "视频管理",
	})

	// GB28181标签
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceGB28181Tag,
		ResourceName: "视频标签",
		Action:       "read",
		Description:  "视频标签查看",
		Category:     "视频管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceGB28181Tag,
		ResourceName: "视频标签",
		Action:       "write",
		Description:  "视频标签创建/修改",
		Category:     "视频管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceGB28181Tag,
		ResourceName: "视频标签",
		Action:       "delete",
		Description:  "视频标签删除",
		Category:     "视频管理",
	})

	// GB28181统计
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceGB28181Stats,
		ResourceName: "视频统计",
		Action:       "read",
		Description:  "视频统计查看",
		Category:     "视频管理",
	})

	// ============ 事件管理资源权限 ============
	// Skylark平台配置
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceSkylarkPlatform,
		ResourceName: "Skylark平台配置",
		Action:       "read",
		Description:  "Skylark平台配置查看",
		Category:     "事件管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceSkylarkPlatform,
		ResourceName: "Skylark平台配置",
		Action:       "write",
		Description:  "Skylark平台配置创建/修改",
		Category:     "事件管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceSkylarkPlatform,
		ResourceName: "Skylark平台配置",
		Action:       "delete",
		Description:  "Skylark平台配置删除",
		Category:     "事件管理",
	})

	// 事件配置
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceEventConfig,
		ResourceName: "事件配置",
		Action:       "read",
		Description:  "事件配置查看",
		Category:     "事件管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceEventConfig,
		ResourceName: "事件配置",
		Action:       "write",
		Description:  "事件配置创建/修改",
		Category:     "事件管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceEventConfig,
		ResourceName: "事件配置",
		Action:       "delete",
		Description:  "事件配置删除",
		Category:     "事件管理",
	})

	// 事件数据查询
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceEventData,
		ResourceName: "事件数据",
		Action:       "read",
		Description:  "事件数据查询",
		Category:     "事件管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceEventData,
		ResourceName: "事件数据",
		Action:       "write",
		Description:  "事件数据创建/修改",
		Category:     "事件管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceEventData,
		ResourceName: "事件数据",
		Action:       "delete",
		Description:  "事件数据删除",
		Category:     "事件管理",
	})

	// 组织映射
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceOrgMapping,
		ResourceName: "组织映射",
		Action:       "read",
		Description:  "组织映射查看",
		Category:     "事件管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceOrgMapping,
		ResourceName: "组织映射",
		Action:       "write",
		Description:  "组织映射创建/修改",
		Category:     "事件管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceOrgMapping,
		ResourceName: "组织映射",
		Action:       "delete",
		Description:  "组织映射删除",
		Category:     "事件管理",
	})

	// 流程Journey操作
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceFlowJourney,
		ResourceName: "流程审批",
		Action:       "read",
		Description:  "流程审批查看",
		Category:     "事件管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceFlowJourney,
		ResourceName: "流程审批",
		Action:       "write",
		Description:  "流程审批操作",
		Category:     "事件管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceFlowJourney,
		ResourceName: "流程审批",
		Action:       "delete",
		Description:  "流程审批删除",
		Category:     "事件管理",
	})

	// 事件流程管理
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceFlow,
		ResourceName: "事件流程",
		Action:       "read",
		Description:  "流程查询",
		Category:     "事件管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceFlow,
		ResourceName: "事件流程",
		Action:       "write",
		Description:  "流程创建",
		Category:     "事件管理",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceFlow,
		ResourceName: "事件流程",
		Action:       "delete",
		Description:  "流程终止",
		Category:     "事件管理",
	})

	// ============ LynxGraph逻辑引擎资源权限 ============
	// 信息原子类型
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceLynxInfoAtomType,
		ResourceName: "信息原子类型",
		Action:       "read",
		Description:  "信息原子类型查看",
		Category:     "LynxGraph逻辑引擎",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceLynxInfoAtomType,
		ResourceName: "信息原子类型",
		Action:       "write",
		Description:  "信息原子类型创建/修改",
		Category:     "LynxGraph逻辑引擎",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceLynxInfoAtomType,
		ResourceName: "信息原子类型",
		Action:       "delete",
		Description:  "信息原子类型删除",
		Category:     "LynxGraph逻辑引擎",
	})

	// 逻辑图配置
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceLynxGraphConfig,
		ResourceName: "逻辑图配置",
		Action:       "read",
		Description:  "逻辑图配置查看",
		Category:     "LynxGraph逻辑引擎",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceLynxGraphConfig,
		ResourceName: "逻辑图配置",
		Action:       "write",
		Description:  "逻辑图配置创建/修改",
		Category:     "LynxGraph逻辑引擎",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceLynxGraphConfig,
		ResourceName: "逻辑图配置",
		Action:       "delete",
		Description:  "逻辑图配置删除",
		Category:     "LynxGraph逻辑引擎",
	})

	// 逻辑块规格（只读）
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceLynxBlockSpec,
		ResourceName: "逻辑块规格",
		Action:       "read",
		Description:  "逻辑块规格查看",
		Category:     "LynxGraph逻辑引擎",
	})

	// Lynx标签
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceLynxTag,
		ResourceName: "Lynx标签",
		Action:       "read",
		Description:  "Lynx标签查看",
		Category:     "LynxGraph逻辑引擎",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceLynxTag,
		ResourceName: "Lynx标签",
		Action:       "write",
		Description:  "Lynx标签创建/修改",
		Category:     "LynxGraph逻辑引擎",
	})
	RegisterPermission(PermissionMetadata{
		Resource:     ResourceLynxTag,
		ResourceName: "Lynx标签",
		Action:       "delete",
		Description:  "Lynx标签删除",
		Category:     "LynxGraph逻辑引擎",
	})
}
