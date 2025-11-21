package auth

// 资源类型常量定义
// 用于 CasbinX 权限验证

const (
	// ============ IoT 资源 ============

	// ResourceIoTTemplate IoT设备模板资源
	// 权限: iot_template:read, iot_template:write, iot_template:delete
	ResourceIoTTemplate = "iot_template"

	// ResourceIoTDevice IoT设备资源
	// 权限: iot_device:read, iot_device:write, iot_device:delete
	ResourceIoTDevice = "iot_device"

	// ResourceIoTTag IoT标签资源
	// 权限: iot_tag:read, iot_tag:write, iot_tag:delete
	ResourceIoTTag = "iot_tag"

	// ResourceIoTMetadata IoT元数据资源（设备类别、标准字段等）
	// 权限: iot_metadata:read, iot_metadata:write, iot_metadata:delete
	ResourceIoTMetadata = "iot_metadata"

	// ResourceIoTPlatform IoT平台配置资源
	// 权限: iot_platform:read, iot_platform:write, iot_platform:delete
	ResourceIoTPlatform = "iot_platform"

	// ============ GB28181 资源 ============

	// ResourceGB28181Device GB28181设备资源
	// 权限: gb28181_device:read, gb28181_device:write, gb28181_device:delete
	ResourceGB28181Device = "gb28181_device"

	// ResourceGB28181Stream GB28181流资源
	// 权限: gb28181_stream:read, gb28181_stream:write, gb28181_stream:delete
	ResourceGB28181Stream = "gb28181_stream"

	// ResourceGB28181Tag GB28181标签资源
	// 权限: gb28181_tag:read, gb28181_tag:write, gb28181_tag:delete
	ResourceGB28181Tag = "gb28181_tag"

	// ResourceGB28181Stats GB28181统计资源
	// 权限: gb28181_stats:read, gb28181_stats:write, gb28181_stats:delete
	ResourceGB28181Stats = "gb28181_stats"

	// ============ Skylark事件管理资源 ============

	// ResourceSkylarkPlatform Skylark平台配置资源
	// 权限: skylark_platform:read, skylark_platform:write, skylark_platform:delete
	ResourceSkylarkPlatform = "skylark_platform"

	// ResourceEventConfig 事件配置资源
	// 权限: event_config:read, event_config:write, event_config:delete
	ResourceEventConfig = "event_config"

	// ResourceEventData 事件数据查询资源
	// 权限: event_data:read (查询权限受组织架构限制)
	ResourceEventData = "event_data"

	// ResourceOrgMapping 组织映射资源（租户级别全局配置）
	// 权限: org_mapping:read, org_mapping:write, org_mapping:delete
	ResourceOrgMapping = "org_mapping"

	// ResourceFlowJourney 流程Journey操作资源
	// 权限: flow_journey:write (审批操作权限)
	ResourceFlowJourney = "flow_journey"

	// ResourceFlow 流程管理资源
	// 权限: flow:read (查询流程), flow:write (创建流程), flow:delete (终止流程)
	ResourceFlow = "flow"

	// ResourceForm 表单管理资源
	// 权限: form:write (提交表单)
	ResourceForm = "form"

	// ============ LynxGraph逻辑引擎资源 ============

	// ResourceLynxInfoAtomType 信息原子类型资源
	// 权限: lynx_infoatomtype:read, lynx_infoatomtype:write, lynx_infoatomtype:delete
	ResourceLynxInfoAtomType = "lynx_infoatomtype"

	// ResourceLynxGraphConfig 逻辑图配置资源
	// 权限: lynx_graphconfig:read, lynx_graphconfig:write, lynx_graphconfig:delete
	ResourceLynxGraphConfig = "lynx_graphconfig"

	// ResourceLynxBlockSpec 逻辑块规格资源（只读）
	// 权限: lynx_blockspec:read
	ResourceLynxBlockSpec = "lynx_blockspec"

	// ResourceLynxTag Lynx标签资源
	// 权限: lynx_tag:read, lynx_tag:write, lynx_tag:delete
	ResourceLynxTag = "lynx_tag"
)
