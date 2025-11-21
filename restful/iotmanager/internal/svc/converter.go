package svc

import (
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core/devices"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"
)

// ConvertOnlineDetectionConfigToCore 转换API类型的OnlineDetectionConfig为core类型
// tenantID: 租户ID，由REST层从JWT Token解析后传入
func ConvertOnlineDetectionConfigToCore(config types.OnlineDetectionConfig, tenantID string) *core.OnlineDetectionConfig {
	// 转换字段检查规则
	fieldChecks := make([]core.FieldCheck, 0, len(config.FieldChecks))
	for _, fc := range config.FieldChecks {
		fieldChecks = append(fieldChecks, core.FieldCheck{
			Field:    fc.Field,
			Operator: core.FieldCheckOperator(fc.Operator),
			Value:    fc.Value,
			Values:   fc.Values,
		})
	}

	return &core.OnlineDetectionConfig{
		Category:       config.Category,
		Model:          config.Model,
		TopicSuffix:    config.TopicSuffix,
		TenantID:       tenantID, // 从JWT Token解析的租户ID
		Strategy:       core.OnlineStrategy(config.Strategy),
		TimeoutSeconds: config.TimeoutSeconds,
		FieldChecks:    fieldChecks,
	}
}

// ConvertDataProcessingConfigToCore 转换API类型的DataProcessingConfig为core类型
func ConvertDataProcessingConfigToCore(config types.DataProcessingConfig, tenantID string) *core.DataProcessingConfig {
	// 转换字段映射
	fieldMappings := make([]core.FieldMapping, 0, len(config.FieldMappings))
	for _, fm := range config.FieldMappings {
		fieldMappings = append(fieldMappings, core.FieldMapping{
			StandardField: core.FieldName(fm.StandardField),
			SourcePath:    fm.SourcePath,
			FieldType:     core.FieldType(fm.FieldType),
			Scale:         fm.Scale,
			Offset:        fm.Offset,
			DefaultValue:  fm.DefaultValue,
		})
	}

	// 转换过滤规则
	var filterRules *core.FilterRule
	if config.FilterRules != nil {
		conditions := make([]core.Condition, 0, len(config.FilterRules.Conditions))
		for _, cond := range config.FilterRules.Conditions {
			conditions = append(conditions, core.Condition{
				Field:    cond.Field,
				Operator: cond.Operator,
				Value:    cond.Value,
			})
		}
		filterRules = &core.FilterRule{
			Conditions: conditions,
			Logic:      config.FilterRules.Logic,
		}
	}

	return &core.DataProcessingConfig{
		Category:        config.Category,
		Model:           config.Model,
		TopicSuffix:     config.TopicSuffix,
		TenantID:        tenantID,
		TimestampPath:   config.TimestampPath,
		TimestampFormat: core.TimestampFormat(config.TimestampFormat),
		FieldMappings:   fieldMappings,
		FilterRules:     filterRules,
	}
}

// ConvertCoreTemplateToTypes 转换core.Template为types.SensorTemplateDetail
func ConvertCoreTemplateToTypes(template *core.Template) types.SensorTemplateDetail {
	detail := types.SensorTemplateDetail{
		SensorTemplate: types.SensorTemplate{
			Id:           template.ID,
			Model:        template.Model,
			Name:         template.Name,
			Category:     template.Category,
			Manufacturer: template.Manufacturer,
			Description:  template.Description,
			Version:      template.Version,
			Enabled:      template.Enabled,
			DeviceCount:  template.DeviceCount,
			TenantId:     template.TenantID,
			CreatedBy:    template.CreatedBy,
			UpdatedBy:    "", // core.Template没有UpdatedBy字段
			CreatedAt:    template.CreatedAt,
			UpdatedAt:    template.UpdatedAt,
		},
	}

	// 转换在线检测配置（如果存在且有效）
	if template.OnlineConfig != nil && template.OnlineConfig.TopicSuffix != "" {
		onlineConfig := ConvertCoreOnlineDetectionConfigToTypes(*template.OnlineConfig)
		detail.OnlineConfig = onlineConfig
	}

	// 转换业务数据处理配置（如果存在且有效）
	if template.BusinessConfig != nil && template.BusinessConfig.TopicSuffix != "" {
		businessConfig := ConvertCoreDataProcessingConfigToTypes(*template.BusinessConfig)
		detail.BusinessConfig = businessConfig
	}

	// 转换控制配置（如果存在且有效）
	if template.ControlConfig != nil && template.ControlConfig.CommandSuffix != "" {
		controlConfig := ConvertCoreDeviceControlConfigToTypes(*template.ControlConfig)
		detail.ControlConfig = controlConfig
	}

	return detail
}

// ConvertCoreOnlineDetectionConfigToTypes 转换core.OnlineDetectionConfig为types类型
func ConvertCoreOnlineDetectionConfigToTypes(config core.OnlineDetectionConfig) types.OnlineDetectionConfig {
	// 转换字段检查规则
	fieldChecks := make([]types.FieldCheck, 0, len(config.FieldChecks))
	for _, fc := range config.FieldChecks {
		fieldChecks = append(fieldChecks, types.FieldCheck{
			Field:    fc.Field,
			Operator: string(fc.Operator),
			Value:    fc.Value,
			Values:   fc.Values,
		})
	}

	return types.OnlineDetectionConfig{
		Category:       config.Category,
		Model:          config.Model,
		TopicSuffix:    config.TopicSuffix,
		Strategy:       string(config.Strategy),
		TimeoutSeconds: config.TimeoutSeconds,
		FieldChecks:    fieldChecks,
	}
}

// ConvertCoreDataProcessingConfigToTypes 转换core.DataProcessingConfig为types类型
func ConvertCoreDataProcessingConfigToTypes(config core.DataProcessingConfig) types.DataProcessingConfig {
	// 转换字段映射
	fieldMappings := make([]types.FieldMapping, 0, len(config.FieldMappings))
	for _, fm := range config.FieldMappings {
		fieldMappings = append(fieldMappings, types.FieldMapping{
			StandardField: string(fm.StandardField),
			SourcePath:    fm.SourcePath,
			FieldType:     string(fm.FieldType),
			Scale:         fm.Scale,
			Offset:        fm.Offset,
			DefaultValue:  fm.DefaultValue,
		})
	}

	// 转换过滤规则
	var filterRules *types.FilterRule
	if config.FilterRules != nil {
		conditions := make([]types.Condition, 0, len(config.FilterRules.Conditions))
		for _, cond := range config.FilterRules.Conditions {
			conditions = append(conditions, types.Condition{
				Field:    cond.Field,
				Operator: cond.Operator,
				Value:    cond.Value,
			})
		}
		filterRules = &types.FilterRule{
			Conditions: conditions,
			Logic:      config.FilterRules.Logic,
		}
	}

	return types.DataProcessingConfig{
		Category:        config.Category,
		Model:           config.Model,
		TopicSuffix:     config.TopicSuffix,
		TimestampPath:   config.TimestampPath,
		TimestampFormat: string(config.TimestampFormat),
		FieldMappings:   fieldMappings,
		FilterRules:     filterRules,
	}
}

// ConvertCoreTemplateSummariesToTypes 转换core.TemplateSummary列表为types.SensorTemplate列表
func ConvertCoreTemplateSummariesToTypes(summaries []*core.TemplateSummary) []types.SensorTemplate {
	result := make([]types.SensorTemplate, 0, len(summaries))
	for _, summary := range summaries {
		result = append(result, types.SensorTemplate{
			Id:           summary.ID,
			Model:        summary.Model,
			Name:         summary.Name,
			Category:     summary.Category,
			Manufacturer: summary.Manufacturer,
			Description:  "", // TemplateSummary没有Description字段
			Version:      summary.Version,
			Enabled:      summary.Enabled,
			DeviceCount:  summary.DeviceCount,
			TenantId:     "", // TemplateSummary没有TenantID字段
			CreatedBy:    "", // TemplateSummary没有CreatedBy字段
			UpdatedBy:    "", // TemplateSummary没有UpdatedBy字段
			CreatedAt:    summary.CreatedAt,
			UpdatedAt:    summary.UpdatedAt,
		})
	}
	return result
}

// HandleIoTError 处理IoT引擎错误并转换为HTTP响应
func HandleIoTError(err error) (code int64, msg string) {
	switch err {
	case core.ErrTemplateNotFound:
		return 404, "模板不存在"
	case core.ErrTemplateAlreadyExists:
		return 400, "模板已存在"
	case core.ErrTemplateInUse:
		return 400, "模板正在使用中，无法删除"
	case core.ErrDeviceNotFound:
		return 404, "设备不存在"
	case core.ErrDeviceAlreadyBound:
		return 400, "设备已绑定"
	case core.ErrDeviceModelNotFound:
		return 404, "设备型号模板不存在"
	case core.ErrDeviceUnauthorized:
		return 403, "无权操作该设备"
	case core.ErrTagNotFound:
		return 404, "标签不存在"
	case core.ErrTagNameConflict:
		return 400, "标签名称已存在"
	case core.ErrTagInUse:
		return 400, "标签正在使用中，无法删除"
	case core.ErrPlatformNotFound:
		return 404, "平台配置不存在"
	case core.ErrPlatformAlreadyExists:
		return 400, "平台配置ID已存在"
	case core.ErrPlatformInvalid:
		return 400, "平台配置无效"
	case core.ErrPlatformInUse:
		return 400, "平台配置正在使用中，无法删除"
	case core.ErrPlatformUnauthorized:
		return 403, "无权操作该平台配置"
	default:
		return 500, fmt.Sprintf("服务器错误: %v", err)
	}
}

// ConvertCoreDeviceBindingToTypes 转换core.DeviceBinding为types.DeviceBinding
func ConvertCoreDeviceBindingToTypes(device *core.DeviceBinding) types.DeviceBinding {
	tags := make([]types.DeviceTagSummary, 0, len(device.Tags))
	for _, tag := range device.Tags {
		tags = append(tags, types.DeviceTagSummary{
			Id:          tag.ID,
			Name:        tag.Label,
			Color:       tag.Color,
			Description: tag.Description,
		})
	}

	return types.DeviceBinding{
		Id:               device.ID,
		DeviceId:         device.DeviceID,
		DeviceName:       device.DeviceName,
		DeviceAlias:      device.DeviceAlias,
		DeviceModel:      device.DeviceModel,
		DeviceCategory:   device.DeviceCategory,
		Description:      device.Description,
		Location:         device.Location,
		InstallationDate: device.InstallationDate,
		Status:           device.Status,
		IsOnline:         device.IsOnline,
		LastDataAt:       device.LastDataAt,
		TenantId:         device.TenantID,
		OrgId:            device.OrgID,
		Tags:             tags,
		CreatedBy:        device.CreatedBy,
		UpdatedBy:        device.UpdatedBy,
		CreatedAt:        device.CreatedAt,
		UpdatedAt:        device.UpdatedAt,
	}
}

// ConvertCoreDeviceBindingSummariesToTypes 转换core.DeviceBindingSummary列表为types.DeviceBindingSummary列表
func ConvertCoreDeviceBindingSummariesToTypes(summaries []*core.DeviceBindingSummary) []types.DeviceBindingSummary {
	result := make([]types.DeviceBindingSummary, 0, len(summaries))
	for _, summary := range summaries {
		tags := make([]types.DeviceTagSummary, 0, len(summary.Tags))
		for _, tag := range summary.Tags {
			tags = append(tags, types.DeviceTagSummary{
				Id:          tag.ID,
				Name:        tag.Label,
				Color:       tag.Color,
				Description: tag.Description,
			})
		}

		result = append(result, types.DeviceBindingSummary{
			Id:             summary.ID,
			DeviceId:       summary.DeviceID,
			DeviceName:     summary.DeviceName,
			DeviceAlias:    summary.DeviceAlias,
			DeviceModel:    summary.DeviceModel,
			DeviceCategory: summary.DeviceCategory,
			Location:       summary.Location,
			Status:         summary.Status,
			IsOnline:       summary.IsOnline,
			LastDataAt:     summary.LastDataAt,
			OrgId:          summary.OrgID,
			OrgName:        summary.OrgName,
			Tags:           tags,
			CreatedAt:      summary.CreatedAt,
		})
	}
	return result
}

// ConvertCoreUnboundDevicesToTypes 转换core.UnboundDevice列表为types.UnboundDevice列表
func ConvertCoreUnboundDevicesToTypes(devices []core.UnboundDevice) []types.UnboundDevice {
	result := make([]types.UnboundDevice, 0, len(devices))
	for _, device := range devices {
		result = append(result, types.UnboundDevice{
			DeviceId: device.DeviceID,
			Category: device.Category,
			IsOnline: device.IsOnline,
			LastSeen: device.LastSeen,
		})
	}
	return result
}

// ConvertCoreDeviceTagToTypes 转换core.DeviceTag为types.DeviceTag
func ConvertCoreDeviceTagToTypes(tag *core.DeviceTag) types.DeviceTag {
	return types.DeviceTag{
		Id:          tag.ID,
		Name:        tag.Label,
		Color:       tag.Color,
		Icon:        "", // core层没有Icon字段
		Description: tag.Description,
		TenantId:    tag.TenantID,
		DeviceCount: 0, // 需要单独查询
		CreatedBy:   tag.CreatedBy,
		UpdatedBy:   tag.UpdatedBy,
		CreatedAt:   tag.CreatedAt,
		UpdatedAt:   tag.UpdatedAt,
	}
}

// ConvertCoreDeviceTagSummariesToTypes 转换core.DeviceTagSummary列表为types.DeviceTagSummary列表
func ConvertCoreDeviceTagSummariesToTypes(summaries []*core.DeviceTagSummary) []types.DeviceTagSummary {
	result := make([]types.DeviceTagSummary, 0, len(summaries))
	for _, summary := range summaries {
		result = append(result, types.DeviceTagSummary{
			Id:          summary.ID,
			Name:        summary.Label,
			Color:       summary.Color,
			Description: summary.Description,
		})
	}
	return result
}

// ConvertPlatformConfigRequestToCore 转换API请求为Core平台配置
// ID由IoT引擎层Manager的Create方法自动生成
func ConvertPlatformConfigRequestToCore(req *types.CreatePlatformRequest, tenantID string) (*core.PlatformMetadata, *core.PlatformConfig) {
	metadata := core.PlatformMetadata{
		Type:        core.PlatformType(req.Type),
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
		TenantID:    tenantID,
	}

	config := &core.PlatformConfig{
		Type:     core.PlatformType(req.Type),
		TenantID: tenantID,
	}

	// 根据平台类型设置特定配置
	switch core.PlatformType(req.Type) {
	case core.PlatformTypeSkylark:
		config.SkylarkDomain = req.SkylarkDomain
		config.SkylarkAuthHeader = req.SkylarkAuthHeader
		config.SkylarkUserID = req.SkylarkUserID
	case core.PlatformTypeWebhook:
		config.WebhookURL = req.WebhookURL
		config.WebhookHeaders = req.WebhookHeaders
		config.WebhookMethod = req.WebhookMethod
	case core.PlatformTypeKafka:
		config.KafkaBrokers = req.KafkaBrokers
		config.KafkaTopic = req.KafkaTopic
	}

	return &metadata, config
}

// ConvertPlatformUpdateRequestToCore 转换更新请求为Core平台配置
func ConvertPlatformUpdateRequestToCore(req *types.UpdatePlatformRequest, tenantID string) (*core.PlatformMetadata, *core.PlatformConfig) {
	metadata := core.PlatformMetadata{
		Type:        core.PlatformType(req.Type),
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
		TenantID:    tenantID,
	}

	config := &core.PlatformConfig{
		Type:     core.PlatformType(req.Type),
		TenantID: tenantID,
	}

	// 根据平台类型设置特定配置
	switch core.PlatformType(req.Type) {
	case core.PlatformTypeSkylark:
		config.SkylarkDomain = req.SkylarkDomain
		config.SkylarkAuthHeader = req.SkylarkAuthHeader
		config.SkylarkUserID = req.SkylarkUserID
	case core.PlatformTypeWebhook:
		config.WebhookURL = req.WebhookURL
		config.WebhookHeaders = req.WebhookHeaders
		config.WebhookMethod = req.WebhookMethod
	case core.PlatformTypeKafka:
		config.KafkaBrokers = req.KafkaBrokers
		config.KafkaTopic = req.KafkaTopic
	}

	return &metadata, config
}

// ConvertCorePlatformToTypes 转换Core平台配置为API详情
func ConvertCorePlatformToTypes(platform *core.Platform) types.PlatformDetail {
	detail := types.PlatformDetail{
		PlatformMetadata: types.PlatformMetadata{
			Id:          platform.ID,
			Type:        string(platform.Type),
			Name:        platform.Name,
			Description: platform.Description,
			Enabled:     platform.Enabled,
			TenantId:    platform.TenantID,
			CreatedBy:   platform.CreatedBy,
			UpdatedBy:   platform.UpdatedBy,
			CreatedAt:   platform.CreatedAt,
			UpdatedAt:   platform.UpdatedAt,
		},
	}

	// 如果有配置信息，根据类型转换
	if platform.Config != nil {
		switch platform.Type {
		case core.PlatformTypeSkylark:
			detail.Config = types.SkylarkConfig{
				SkylarkDomain:     platform.Config.SkylarkDomain,
				SkylarkAuthHeader: platform.Config.SkylarkAuthHeader,
				SkylarkUserID:     platform.Config.SkylarkUserID,
			}
		case core.PlatformTypeWebhook:
			detail.Config = types.WebhookConfig{
				WebhookURL:     platform.Config.WebhookURL,
				WebhookHeaders: platform.Config.WebhookHeaders,
				WebhookMethod:  platform.Config.WebhookMethod,
			}
		case core.PlatformTypeKafka:
			detail.Config = types.KafkaConfig{
				KafkaBrokers: platform.Config.KafkaBrokers,
				KafkaTopic:   platform.Config.KafkaTopic,
			}
		}
	}

	return detail
}

// ConvertCorePlatformListToTypes 转换Core平台列表为API列表（不包含Config）
func ConvertCorePlatformListToTypes(platformList []*core.Platform) []types.PlatformMetadata {
	result := make([]types.PlatformMetadata, 0, len(platformList))
	for _, platform := range platformList {
		result = append(result, types.PlatformMetadata{
			Id:          platform.ID,
			Type:        string(platform.Type),
			Name:        platform.Name,
			Description: platform.Description,
			Enabled:     platform.Enabled,
			TenantId:    platform.TenantID,
			CreatedBy:   platform.CreatedBy,
			UpdatedBy:   platform.UpdatedBy,
			CreatedAt:   platform.CreatedAt,
			UpdatedAt:   platform.UpdatedAt,
		})
	}
	return result
}

// ===== 查询相关转换函数 =====

// ConvertFieldNamesToCore 转换字符串数组为FieldName数组
func ConvertFieldNamesToCore(fieldNames []string) []core.FieldName {
	result := make([]core.FieldName, 0, len(fieldNames))
	for _, name := range fieldNames {
		result = append(result, core.FieldName(name))
	}
	return result
}

// ConvertCoreTimeSeriesResultToTypes 转换Core时序数据查询结果为API类型
func ConvertCoreTimeSeriesResultToTypes(result *core.TimeSeriesResult) []types.TimeSeriesDataItem {
	list := make([]types.TimeSeriesDataItem, 0, len(result.Data))
	for _, data := range result.Data {
		list = append(list, types.TimeSeriesDataItem{
			Timestamp:      data.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			DeviceId:       data.DeviceID,
			DeviceModel:    data.DeviceModel,
			DeviceCategory: string(data.DeviceCategory),
			FieldName:      string(data.FieldName),
			Value:          data.Value,
		})
	}
	return list
}

// ConvertCoreDeviceLatestValuesToTypes 转换Core最新值结果为API类型
func ConvertCoreDeviceLatestValuesToTypes(devices []core.DeviceLatestValues) []types.DeviceLatestValuesItem {
	result := make([]types.DeviceLatestValuesItem, 0, len(devices))
	for _, device := range devices {
		fields := make(map[string]types.FieldValue)
		for fieldName, value := range device.Values {
			timestamp := device.Timestamps[fieldName]
			fields[string(fieldName)] = types.FieldValue{
				Value:     value,
				Timestamp: timestamp.Format("2006-01-02T15:04:05Z07:00"),
			}
		}

		result = append(result, types.DeviceLatestValuesItem{
			DeviceId: device.DeviceID,
			Fields:   fields,
		})
	}
	return result
}

// ConvertCoreDeviceStatisticsToTypes 转换Core设备统计结果为API类型
func ConvertCoreDeviceStatisticsToTypes(stats []core.DeviceStatistics) []types.DeviceStatisticsItem {
	result := make([]types.DeviceStatisticsItem, 0, len(stats))
	for _, stat := range stats {
		fields := make(map[string]types.FieldStatItem)
		for fieldName, fieldStat := range stat.FieldStats {
			fields[string(fieldName)] = types.FieldStatItem{
				Min:       fieldStat.Min,
				Max:       fieldStat.Max,
				Avg:       fieldStat.Avg,
				Sum:       fieldStat.Sum,
				Count:     fieldStat.Count,
				LastValue: fieldStat.LastValue,
				LastTime:  fieldStat.LastTime.Format("2006-01-02T15:04:05Z07:00"),
			}
		}

		result = append(result, types.DeviceStatisticsItem{
			DeviceId:    stat.DeviceID,
			DeviceModel: stat.DeviceModel,
			Fields:      fields,
		})
	}
	return result
}

// ===== AI Box 控制相关转换函数 =====

// ConvertCoreAIBoxTasksToTypes 转换core.AIBoxAlgorithmTask列表为types.AIBoxAlgorithmTask列表
func ConvertCoreAIBoxTasksToTypes(tasks []devices.AIBoxAlgorithmTask) []types.AIBoxAlgorithmTask {
	result := make([]types.AIBoxAlgorithmTask, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, types.AIBoxAlgorithmTask{
			AlgTaskSession: task.AlgTaskSession,
			TaskDesc:       task.TaskDesc,
			MediaName:      task.MediaName,
			AlgInfo:        task.AlgInfo,
			AlgTaskStatus: types.AIBoxTaskStatus{
				Type:  task.AlgTaskStatus.Type,
				Style: task.AlgTaskStatus.Style,
				Label: task.AlgTaskStatus.Label,
			},
			AlarmBody:     task.AlarmBody,
			AlarmProtocol: task.AlarmProtocol,
			MetadataUrl:   task.MetadataUrl, // 保持 interface{} 类型，由前端处理
			UserData:      task.UserData,
		})
	}
	return result
}

// ConvertCoreAIBoxCapabilitiesToTypes 转换core.AIBoxCapabilities为types.AIBoxCapabilities
func ConvertCoreAIBoxCapabilitiesToTypes(capabilities *devices.AIBoxCapabilities) types.AIBoxCapabilities {
	abilities := make([]types.AIBoxAbility, 0, len(capabilities.Abilities))
	for _, ability := range capabilities.Abilities {
		abilities = append(abilities, ConvertCoreAIBoxAbilityToTypes(ability))
	}

	return types.AIBoxCapabilities{
		BoardId:   capabilities.BoardId,
		Abilities: abilities,
	}
}

// ConvertCoreAIBoxAbilityToTypes 转换core.AIBoxAbility为types.AIBoxAbility
func ConvertCoreAIBoxAbilityToTypes(ability devices.AIBoxAbility) types.AIBoxAbility {
	// 转换Parameters列表
	parameters := make([]types.AIBoxParameter, 0, len(ability.Parameters))
	for _, param := range ability.Parameters {
		// 转换Options列表
		options := make([]types.AIBoxParameterOption, 0, len(param.Options))
		for _, opt := range param.Options {
			options = append(options, types.AIBoxParameterOption{
				Enable: opt.Enable,
				Key:    opt.Key,
				Value:  opt.Value,
				Name:   opt.Name,
			})
		}

		parameters = append(parameters, types.AIBoxParameter{
			Class:    param.Class,
			Type:     param.Type,
			Max:      param.Max,
			Min:      param.Min,
			Default:  param.Default,
			Key:      param.Key,
			Name:     param.Name,
			Required: param.Required,
			Value:    param.Value,
			Options:  options,
		})
	}

	// 转换Policy列表
	policies := make([]types.AIBoxPolicy, 0, len(ability.Policy))
	for _, policy := range ability.Policy {
		policies = append(policies, types.AIBoxPolicy{
			Property: policy.Property,
			Name:     policy.Name,
		})
	}

	return types.AIBoxAbility{
		Attribute: types.AIBoxAttribute{
			LineRequired: ability.Attribute.LineRequired,
			LineDesc:     ability.Attribute.LineDesc,
			ZoneRequired: ability.Attribute.ZoneRequired,
			ZoneDesc:     ability.Attribute.ZoneDesc,
		},
		Parameters: parameters,
		Permitted:  ability.Permitted,
		Code:       ability.Code,
		Sub:        ability.Sub,
		Name:       ability.Name,
		Desc:       ability.Desc,
		Item:       ability.Item,
		Policy:     policies,
	}
}

// ConvertDeviceControlConfigToCore 转换API类型的DeviceControlConfig为core类型
// tenantID: 租户ID，由REST层从JWT Token解析后传入
func ConvertDeviceControlConfigToCore(config types.DeviceControlConfig, tenantID string) *core.DeviceControlConfig {
	return &core.DeviceControlConfig{
		Category:       config.Category,
		Model:          config.Model,
		CommandSuffix:  config.CommandSuffix,
		ResponseSuffix: config.ResponseSuffix,
		TenantID:       tenantID, // 从JWT Token解析的租户ID
	}
}

// ConvertCoreDeviceControlConfigToTypes 转换core.DeviceControlConfig为types类型
func ConvertCoreDeviceControlConfigToTypes(config core.DeviceControlConfig) types.DeviceControlConfig {
	return types.DeviceControlConfig{
		Category:       config.Category,
		Model:          config.Model,
		CommandSuffix:  config.CommandSuffix,
		ResponseSuffix: config.ResponseSuffix,
	}
}
