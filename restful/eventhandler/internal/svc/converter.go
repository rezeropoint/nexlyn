package svc

import (
	"fmt"

	"github.com/rezeropoint/go-skylark/v2/core"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"
)

// ========== Platform 配置转换 ==========

// ConvertCorePlatformConfigToTypes 转换Core平台配置为API响应（密码和Token不返回）
func ConvertCorePlatformConfigToTypes(cfg *core.PlatformConfig) types.PlatformConfig {
	createdBy := ""
	if cfg.CreatedBy != nil {
		createdBy = *cfg.CreatedBy
	}
	updatedBy := ""
	if cfg.UpdatedBy != nil {
		updatedBy = *cfg.UpdatedBy
	}

	return types.PlatformConfig{
		Id:          cfg.ID,
		TenantId:    cfg.TenantID,
		Host:        cfg.Host,
		Port:        cfg.Port,
		Database:    cfg.Database,
		Username:    cfg.Username,
		Password:    "", // 密码不返回
		NamespaceId: cfg.NamespaceID,
		ApiBaseUrl:  cfg.APIBaseURL, // 新增字段
		ApiToken:    "",             // Token不返回（敏感信息）
		CreatedBy:   createdBy,
		UpdatedBy:   updatedBy,
		CreatedAt:   cfg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   cfg.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ConvertAPIRequestToCorePlatformConfig 转换API请求为Core平台配置
func ConvertAPIRequestToCorePlatformConfig(req *types.SavePlatformConfigRequest, tenantID, userID string) *core.PlatformConfig {
	return &core.PlatformConfig{
		TenantID:    tenantID,
		Host:        req.Host,
		Port:        req.Port,
		Database:    req.Database,
		Username:    req.Username,
		Password:    req.Password,
		NamespaceID: req.NamespaceId,
		APIBaseURL:  req.ApiBaseUrl, // 新增字段
		APIToken:    req.ApiToken,   // 新增字段
		CreatedBy:   &userID,
		UpdatedBy:   &userID,
	}
}

// ========== Flow 信息转换 ==========

// ConvertCoreFlowInfoListToTypes 转换Core FlowInfo列表为API响应
func ConvertCoreFlowInfoListToTypes(flows []*core.FlowInfo) []types.FlowInfo {
	result := make([]types.FlowInfo, 0, len(flows))
	for _, flow := range flows {
		result = append(result, types.FlowInfo{
			Id:          flow.ID,
			Title:       flow.Title,
			NamespaceId: flow.NamespaceID,
		})
	}
	return result
}

// ConvertCoreFieldMetadataListToTypes 转换Core FieldMetadata列表为API响应（分离系统字段和业务字段）
func ConvertCoreFieldMetadataListToTypes(fields []*core.FieldMetadata) (systemFields []types.FieldMetadata, businessFields []types.FieldMetadata) {
	for _, field := range fields {
		fm := types.FieldMetadata{
			FieldName: field.FieldName,
			DataType:  field.DataType,
			IsSystem:  field.IsSystem,
		}
		if field.IsSystem {
			systemFields = append(systemFields, fm)
		} else {
			businessFields = append(businessFields, fm)
		}
	}
	return systemFields, businessFields
}

// ========== Event 配置转换 ==========

// ConvertCoreEventConfigWithFieldsToTypes 转换Core事件配置为API响应
func ConvertCoreEventConfigWithFieldsToTypes(cfg *core.EventAggregate) types.EventConfigWithFields {
	fields := make([]types.FieldConfig, 0, len(cfg.Fields))
	for _, field := range cfg.Fields {
		fields = append(fields, types.FieldConfig{
			Id:            field.ID,
			EventConfigId: field.EventConfigID,
			FieldName:     field.FieldName,
			DisplayName:   field.DisplayName,
			FieldType:     field.FieldType,
			IsVisible:     field.IsVisible,
			DisplayOrder:  field.DisplayOrder,
			IsSearchable:  field.IsSearchable,
			CreatedAt:     field.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     field.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	orgFieldName := ""
	if cfg.EventConfig.OrgFieldName != nil {
		orgFieldName = *cfg.EventConfig.OrgFieldName
	}
	description := ""
	if cfg.EventConfig.Description != nil {
		description = *cfg.EventConfig.Description
	}
	createdBy := ""
	if cfg.EventConfig.CreatedBy != nil {
		createdBy = *cfg.EventConfig.CreatedBy
	}
	updatedBy := ""
	if cfg.EventConfig.UpdatedBy != nil {
		updatedBy = *cfg.EventConfig.UpdatedBy
	}

	return types.EventConfigWithFields{
		Id:           cfg.EventConfig.ID,
		Name:         cfg.EventConfig.Name,
		FlowId:       cfg.EventConfig.FlowID,
		FlowTitle:    cfg.EventConfig.FlowTitle,
		OrgFieldName: orgFieldName,
		Description:  description,
		Enabled:      cfg.EventConfig.Enabled,
		TenantId:     cfg.EventConfig.TenantID,
		CreatedBy:    createdBy,
		UpdatedBy:    updatedBy,
		CreatedAt:    cfg.EventConfig.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    cfg.EventConfig.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Fields:       fields,
	}
}

// ConvertCoreEventConfigListToTypes 转换Core事件配置列表为API响应
func ConvertCoreEventConfigListToTypes(configs []*core.EventAggregate) []types.EventConfigWithFields {
	result := make([]types.EventConfigWithFields, 0, len(configs))
	for _, cfg := range configs {
		result = append(result, ConvertCoreEventConfigWithFieldsToTypes(cfg))
	}
	return result
}

// ConvertCreateEventRequestToCore 转换API创建请求为Core创建请求
func ConvertCreateEventRequestToCore(req *types.CreateEventConfigWithFieldsRequest, tenantID, userID string) *core.EventCreation {
	// 转换字段配置
	fields := make([]core.FieldConfig, 0, len(req.Fields))
	for _, field := range req.Fields {
		fields = append(fields, core.FieldConfig{
			FieldName:    field.FieldName,
			DisplayName:  field.DisplayName,
			FieldType:    field.FieldType,
			IsVisible:    field.IsVisible,
			DisplayOrder: field.DisplayOrder,
			IsSearchable: field.IsSearchable,
		})
	}

	var orgFieldName *string
	if req.OrgFieldName != "" {
		orgFieldName = &req.OrgFieldName
	}
	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	return &core.EventCreation{
		EventConfig: core.EventConfig{
			Name:         req.Name,
			FlowID:       req.FlowId,
			OrgFieldName: orgFieldName,
			Description:  description,
			Enabled:      req.Enabled,
			TenantID:     tenantID,
			CreatedBy:    &userID,
			UpdatedBy:    &userID,
		},
		Fields: fields,
	}
}

// ConvertUpdateEventRequestToCore 转换API更新请求为Core更新请求
// 注意：FlowID 和 FlowTitle 从现有配置中提取，不允许修改
func ConvertUpdateEventRequestToCore(req *types.UpdateEventConfigWithFieldsRequest, id, tenantID, userID string, existingConfig *core.EventAggregate) *core.EventUpdate {
	// 转换字段配置
	fields := make([]core.FieldConfig, 0, len(req.Fields))
	for _, field := range req.Fields {
		fields = append(fields, core.FieldConfig{
			FieldName:    field.FieldName,
			DisplayName:  field.DisplayName,
			FieldType:    field.FieldType,
			IsVisible:    field.IsVisible,
			DisplayOrder: field.DisplayOrder,
			IsSearchable: field.IsSearchable,
		})
	}

	var orgFieldName *string
	if req.OrgFieldName != "" {
		orgFieldName = &req.OrgFieldName
	}
	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	return &core.EventUpdate{
		EventConfig: core.EventConfig{
			ID:           id,
			Name:         req.Name,
			FlowID:       existingConfig.EventConfig.FlowID,    // 从现有配置提取，不允许修改
			FlowTitle:    existingConfig.EventConfig.FlowTitle, // 从现有配置提取，不允许修改
			OrgFieldName: orgFieldName,
			Description:  description,
			Enabled:      req.Enabled,
			TenantID:     tenantID,
			UpdatedBy:    &userID,
		},
		Fields: fields,
	}
}

// ========== OrgMapping 组织映射转换 ==========

// ConvertCoreOrgMappingToTypes 转换Core组织映射为API响应
func ConvertCoreOrgMappingToTypes(mapping *core.OrgMapping) types.OrgMapping {
	return types.OrgMapping{
		Id:             mapping.ID,
		TenantId:       mapping.TenantID,
		RemoteOrgValue: mapping.RemoteOrgValue,
		LocalOrgId:     mapping.LocalOrgID,
		LocalOrgName:   mapping.LocalOrgName,
		CreatedAt:      mapping.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      mapping.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ConvertCoreOrgMappingListToTypes 转换Core组织映射列表为API响应
func ConvertCoreOrgMappingListToTypes(mappings []*core.OrgMapping) []types.OrgMapping {
	result := make([]types.OrgMapping, 0, len(mappings))
	for _, mapping := range mappings {
		result = append(result, ConvertCoreOrgMappingToTypes(mapping))
	}
	return result
}

// ConvertCreateRequestToCoreOrgMapping 转换API创建请求为Core组织映射
func ConvertCreateRequestToCoreOrgMapping(req *types.CreateOrgMappingRequest, tenantID string) *core.OrgMapping {
	return &core.OrgMapping{
		TenantID:       tenantID,
		RemoteOrgValue: req.RemoteOrgValue,
		LocalOrgID:     req.LocalOrgId,
	}
}

// ConvertUpdateRequestToCoreOrgMapping 转换API更新请求为Core组织映射（用于更新字段）
func ConvertUpdateRequestToCoreOrgMapping(req *types.UpdateOrgMappingRequest, existingMapping *core.OrgMapping) *core.OrgMapping {
	// 复制现有映射
	updated := *existingMapping

	// 更新请求中的字段（如果提供）
	if req.RemoteOrgValue != "" {
		updated.RemoteOrgValue = req.RemoteOrgValue
	}
	if req.LocalOrgId != "" {
		updated.LocalOrgID = req.LocalOrgId
	}

	return &updated
}

// ========== Event Stats 事件统计转换 ==========

// ConvertCoreDurationStatsToTypes 转换Core处理时长统计为API响应
func ConvertCoreDurationStatsToTypes(stats *core.DurationStats) types.GetDurationStatsResponse {
	// 从*float64提取值，如果为nil则返回0
	avgDuration := float64(0)
	if stats.AvgDuration != nil {
		avgDuration = *stats.AvgDuration
	}

	minDuration := float64(0)
	if stats.MinDuration != nil {
		minDuration = *stats.MinDuration
	}

	maxDuration := float64(0)
	if stats.MaxDuration != nil {
		maxDuration = *stats.MaxDuration
	}

	return types.GetDurationStatsResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "success"},
		Data: struct {
			AvgDuration    float64 `json:"avgDuration"`
			MinDuration    float64 `json:"minDuration"`
			MaxDuration    float64 `json:"maxDuration"`
			TotalCount     int64   `json:"totalCount"`
			CompletedCount int64   `json:"completedCount"`
		}{
			AvgDuration:    avgDuration,
			MinDuration:    minDuration,
			MaxDuration:    maxDuration,
			TotalCount:     stats.TotalCount,
			CompletedCount: stats.CompletedCount,
		},
	}
}

// ConvertCorePendingStatsToTypes 转换Core待处理事件统计为API响应
func ConvertCorePendingStatsToTypes(stats *core.PendingStats) types.GetPendingStatsResponse {
	// 转换事件统计数组
	eventStats := make([]types.EventPendingStats, 0, len(stats.EventStats))
	for _, es := range stats.EventStats {
		eventStats = append(eventStats, types.EventPendingStats{
			EventConfigID:   es.EventConfigID,
			EventName:       es.EventName,
			PendingCount:    es.PendingCount,
			ProcessingCount: es.ProcessingCount,
			Total:           es.Total,
		})
	}

	return types.GetPendingStatsResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "success"},
		Data: struct {
			EventStats      []types.EventPendingStats `json:"eventStats"`
			TotalPending    int64                     `json:"totalPending"`
			TotalProcessing int64                     `json:"totalProcessing"`
			Total           int64                     `json:"total"`
		}{
			EventStats:      eventStats,
			TotalPending:    stats.TotalPending,
			TotalProcessing: stats.TotalProcessing,
			Total:           stats.Total,
		},
	}
}

// ConvertCoreStatusStatsToTypes 转换Core状态统计为API响应
func ConvertCoreStatusStatsToTypes(stats *core.StatusStats) types.GetStatusStatsResponse {
	statusCounts := make([]types.StatusCount, 0, len(stats.StatusCounts))
	for _, sc := range stats.StatusCounts {
		statusCounts = append(statusCounts, types.StatusCount{
			Status:     sc.Status,
			StatusKey:  sc.StatusKey,
			Count:      sc.Count,
			Percentage: sc.Percentage,
		})
	}

	return types.GetStatusStatsResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "success"},
		Data: struct {
			StatusCounts []types.StatusCount `json:"statusCounts"`
			Total        int64               `json:"total"`
		}{
			StatusCounts: statusCounts,
			Total:        stats.Total,
		},
	}
}

// ConvertCoreTrendStatsToTypes 转换Core趋势统计为API响应
func ConvertCoreTrendStatsToTypes(stats *core.TrendStats) types.GetTrendStatsResponse {
	timePoints := make([]types.TrendPoint, 0, len(stats.TimePoints))
	for _, tp := range stats.TimePoints {
		timePoints = append(timePoints, types.TrendPoint{
			Date:           tp.Date,
			TotalCount:     tp.TotalCount,
			CompletedCount: tp.CompletedCount,
			CompletionRate: tp.CompletionRate,
		})
	}

	return types.GetTrendStatsResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "success"},
		Data: struct {
			TimePoints []types.TrendPoint `json:"timePoints"`
		}{
			TimePoints: timePoints,
		},
	}
}

// ConvertCoreNodeStatsToTypes 转换Core节点统计为API响应
func ConvertCoreNodeStatsToTypes(stats *core.NodeStats) types.GetNodeStatsResponse {
	nodeMetrics := make([]types.NodeMetric, 0, len(stats.NodeMetrics))
	for _, nm := range stats.NodeMetrics {
		nodeMetrics = append(nodeMetrics, types.NodeMetric{
			VertexId:    nm.VertexID,
			VertexName:  nm.VertexName,
			Count:       nm.Count,
			AvgDuration: nm.AvgDuration,
		})
	}

	return types.GetNodeStatsResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "success"},
		Data: struct {
			NodeMetrics []types.NodeMetric `json:"nodeMetrics"`
		}{
			NodeMetrics: nodeMetrics,
		},
	}
}

// ConvertCoreUserStatsToTypes 转换Core处理人统计为API响应
func ConvertCoreUserStatsToTypes(stats *core.UserStats) types.GetUserStatsResponse {
	userMetrics := make([]types.UserMetric, 0, len(stats.UserMetrics))
	for _, um := range stats.UserMetrics {
		userMetrics = append(userMetrics, types.UserMetric{
			UserId:   um.UserID,
			UserName: um.UserName,
			Count:    um.Count,
			Rank:     um.Rank,
		})
	}

	return types.GetUserStatsResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "success"},
		Data: struct {
			UserMetrics []types.UserMetric `json:"userMetrics"`
			Total       int64              `json:"total"`
		}{
			UserMetrics: userMetrics,
			Total:       stats.Total,
		},
	}
}

// ConvertCoreOrgStatsToTypes 转换Core组织统计为API响应
func ConvertCoreOrgStatsToTypes(stats *core.OrgStats) types.GetOrgStatsResponse {
	orgMetrics := make([]types.OrgMetric, 0, len(stats.OrgMetrics))
	for _, om := range stats.OrgMetrics {
		orgMetrics = append(orgMetrics, types.OrgMetric{
			OrgValue:    om.OrgValue,
			Count:       om.Count,
			AvgDuration: om.AvgDuration,
		})
	}

	return types.GetOrgStatsResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "success"},
		Data: struct {
			OrgMetrics []types.OrgMetric `json:"orgMetrics"`
		}{
			OrgMetrics: orgMetrics,
		},
	}
}

// ========== Event 事件详情转换 ==========

// ConvertCoreDetailResponseToTypes 转换Core事件详情为API响应
func ConvertCoreDetailResponseToTypes(detail *core.DetailResponse) types.GetEventDetailResponse {
	flowHistory := make([]types.FlowNode, 0, len(detail.FlowHistory))
	for _, node := range detail.FlowHistory {
		flowHistory = append(flowHistory, types.FlowNode{
			VertexId:    node.VertexID,
			VertexName:  node.VertexName,
			VertexAlias: node.VertexAlias,
			UserIds:     node.UserIDs,   // 数组类型，支持多人处理
			UserNames:   node.UserNames, // 数组类型，与UserIDs一一对应
			CreatedAt:   node.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   node.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return types.GetEventDetailResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "success"},
		Data: struct {
			JourneyId          int                    `json:"journeyId"`
			CurrentStatus      string                 `json:"currentStatus"`
			InitiatorUserId    string                 `json:"initiatorUserId"`
			InitiatorUserName  string                 `json:"initiatorUserName"`
			InitiatedAt        string                 `json:"initiatedAt"`
			LatestBusinessData map[string]interface{} `json:"latestBusinessData"`
			FlowHistory        []types.FlowNode       `json:"flowHistory"`
		}{
			JourneyId:          detail.JourneyID,
			CurrentStatus:      detail.CurrentStatus,
			InitiatorUserId:    detail.InitiatorUserID,
			InitiatorUserName:  detail.InitiatorUserName,
			InitiatedAt:        detail.InitiatedAt.Format("2006-01-02T15:04:05Z07:00"),
			LatestBusinessData: detail.LatestBusinessData,
			FlowHistory:        flowHistory,
		},
	}
}

// ========== FlowJourney 操作转换 ==========

// ConvertTypesDataToCoreData 转换API的TypedValue数据为Core的TypedValue数据
func ConvertTypesDataToCoreData(data map[string]types.TypedValue) map[string]core.TypedValue {
	if data == nil {
		return nil
	}

	coreData := make(map[string]core.TypedValue, len(data))
	for key, val := range data {
		coreData[key] = core.TypedValue{
			Type:  val.Type,
			Value: val.Value,
		}
	}
	return coreData
}

// ========== User Assignments 用户任务转换 ==========

// ConvertCoreAssignmentToTypes 转换Core Assignment为API响应
func ConvertCoreAssignmentToTypes(assignment *core.Assignment) types.Assignment {
	flowId := int64(0)
	if assignment.FlowID != nil {
		flowId = *assignment.FlowID
	}
	flowTitle := ""
	if assignment.FlowTitle != nil {
		flowTitle = *assignment.FlowTitle
	}

	return types.Assignment{
		Id:         assignment.ID,
		AssigneeId: assignment.AssigneeID,
		Status:     assignment.Status,
		Category:   assignment.Category,
		VertexId:   assignment.VertexID,
		JourneyId:  assignment.JourneyID,
		CreatedAt:  assignment.CreatedAt, // SDK已经返回ISO 8601格式字符串
		UpdatedAt:  assignment.UpdatedAt, // SDK已经返回ISO 8601格式字符串
		FlowId:     flowId,
		FlowTitle:  flowTitle,
	}
}

// ConvertCoreAssignmentListToTypes 转换Core Assignment列表为API响应
func ConvertCoreAssignmentListToTypes(assignments []*core.Assignment) []types.Assignment {
	result := make([]types.Assignment, 0, len(assignments))
	for _, assignment := range assignments {
		result = append(result, ConvertCoreAssignmentToTypes(assignment))
	}
	return result
}

// ConvertCoreJourneyToTypes 转换Core Journey为API响应
func ConvertCoreJourneyToTypes(journey *core.Journey) types.Journey {
	user := types.FlowUser{}
	if journey.User != nil {
		user.Id = journey.User.ID
		user.Name = journey.User.Name
		// core.FlowUser 只有 ID, Name, Email 三个字段，没有 Nickname 和 Identifier
		// types.FlowUser 的 Nickname 和 Identifier 保持空值
	}

	return types.Journey{
		Id:              journey.ID,
		Sn:              journey.SN,
		Status:          journey.Status,
		CurrentVertexId: journey.CurrentVertexID,
		FlowId:          journey.FlowID,
		CreatedAt:       journey.CreatedAt, // SDK已经返回ISO 8601格式字符串
		UpdatedAt:       journey.UpdatedAt, // SDK已经返回ISO 8601格式字符串
		User:            user,
	}
}

// ConvertCoreJourneyListToTypes 转换Core Journey列表为API响应
func ConvertCoreJourneyListToTypes(journeys []*core.Journey) []types.Journey {
	result := make([]types.Journey, 0, len(journeys))
	for _, journey := range journeys {
		result = append(result, ConvertCoreJourneyToTypes(journey))
	}
	return result
}

// ========== Flow Metadata 流程元数据转换 ==========

// ConvertCoreFlowDetailToTypes 转换Core FlowDetail为API响应
func ConvertCoreFlowDetailToTypes(detail *core.FlowDetail) types.FlowDetailData {
	// 转换字段列表
	// 注意：core.FlowField 只有 ID 和 Title 两个字段
	// types.FlowField 的其他字段（Name、Type、Required、Description）需要前端从其他接口获取
	fields := make([]types.FlowField, 0, len(detail.Fields))
	for _, field := range detail.Fields {
		fields = append(fields, types.FlowField{
			Name:        "",          // core.FlowField 无此字段，保留空值
			Title:       field.Title, // 使用 Title
			Type:        "",          // core.FlowField 无此字段，保留空值
			Required:    false,       // core.FlowField 无此字段，默认false
			Description: "",          // core.FlowField 无此字段，保留空值
		})
	}

	// 转换节点列表
	// 注意：core.FlowVertex 只有 ID、Name、Type 三个字段
	vertices := make([]types.FlowVertex, 0, len(detail.Vertices))
	for _, vertex := range detail.Vertices {
		vertices = append(vertices, types.FlowVertex{
			Id:          vertex.ID,
			Title:       vertex.Name, // core使用Name，types使用Title
			Type:        vertex.Type,
			Description: "", // core.FlowVertex 无此字段，保留空值
		})
	}

	// 转换边列表
	// 注意：core.FlowEdge 只有 FromVertexID 和 ToVertexID
	edges := make([]types.FlowEdge, 0, len(detail.Edges))
	for i, edge := range detail.Edges {
		edges = append(edges, types.FlowEdge{
			Id:           int64(i + 1),      // core无ID字段，使用索引+1作为ID
			SourceVertex: edge.FromVertexID, // core使用FromVertexID
			TargetVertex: edge.ToVertexID,   // core使用ToVertexID
			Condition:    "",                // core.FlowEdge 无此字段，保留空值
		})
	}

	return types.FlowDetailData{
		Id:       detail.ID,
		Title:    detail.Title,
		Fields:   fields,
		Vertices: vertices,
		Edges:    edges,
	}
}

// ========== Journey Detail 流程详情转换 ==========

// ConvertCoreJourneyDetailToTypes 转换Core JourneyDetail为API响应
func ConvertCoreJourneyDetailToTypes(detail *core.JourneyDetail) types.JourneyDetail {
	// 转换发起人信息
	// 注意：core.FlowUser 只有 ID, Name, Email 三个字段，没有 Nickname 和 Identifier
	initiator := types.FlowUser{}
	if detail.Initiator != nil {
		initiator.Id = detail.Initiator.ID
		initiator.Name = detail.Initiator.Name
		// types.FlowUser 的 Nickname 和 Identifier 保持空值（core无此字段）
	}

	// 转换审批节点ID列表（可能为空）
	reviewerVertexIds := make([]int64, 0)
	if detail.ReviewerVertexIDs != nil {
		reviewerVertexIds = detail.ReviewerVertexIDs
	}

	// 处理可选字段
	// 注意：core.JourneyURL 是 string，不是 *string
	currentDurationThreshold := ""
	if detail.CurrentDurationThreshold != nil {
		currentDurationThreshold = *detail.CurrentDurationThreshold
	}

	return types.JourneyDetail{
		Id:                       detail.ID,
		Sn:                       detail.SN,
		Status:                   detail.Status,
		CurrentVertexId:          detail.CurrentVertexID,
		FlowId:                   detail.FlowID,
		CreatedAt:                detail.CreatedAt,  // SDK已经返回ISO 8601格式字符串
		UpdatedAt:                detail.UpdatedAt,  // SDK已经返回ISO 8601格式字符串
		JourneyUrl:               detail.JourneyURL, // core 是 string，直接赋值
		ReviewerVertexIds:        reviewerVertexIds,
		CurrentDurationThreshold: currentDurationThreshold,
		Initiator:                initiator,
		BusinessData:             detail.BusinessData, // 图片字段返回 URL
	}
}

// ========== Moment 审批历史转换 ==========

// ConvertCoreMomentToTypes 转换Core Moment为API响应
func ConvertCoreMomentToTypes(moment *core.Moment) types.Moment {
	// 处理可选字段
	vertexName := ""
	if moment.VertexName != nil {
		vertexName = *moment.VertexName
	}

	operatorName := ""
	if moment.OperatorName != nil {
		operatorName = *moment.OperatorName
	}

	comment := ""
	if moment.Comment != nil {
		comment = *moment.Comment
	}

	return types.Moment{
		Id:           moment.ID,
		AssignmentId: moment.AssignmentID,
		JourneyId:    moment.JourneyID,
		VertexId:     moment.VertexID,
		VertexName:   vertexName,
		Status:       moment.Status,
		OperatorId:   moment.OperatorID, // core 已经是 string（本地用户ID），直接赋值
		OperatorName: operatorName,
		Comment:      comment,
		CreatedAt:    moment.CreatedAt, // SDK已经返回ISO 8601格式字符串
		UpdatedAt:    moment.UpdatedAt, // SDK已经返回ISO 8601格式字符串
	}
}

// ConvertCoreMomentListToTypes 转换Core Moment列表为API响应
func ConvertCoreMomentListToTypes(moments []*core.Moment) []types.Moment {
	result := make([]types.Moment, 0, len(moments))
	for _, moment := range moments {
		result = append(result, ConvertCoreMomentToTypes(moment))
	}
	return result
}

// ========== Processing User 当前处理人转换 ==========

// ConvertCoreProcessingUserToTypes 转换Core ProcessingUser为API响应
func ConvertCoreProcessingUserToTypes(user *core.ProcessingUser) types.ProcessingUser {
	// 处理可选字段
	nickname := ""
	if user.Nickname != nil {
		nickname = *user.Nickname
	}

	phone := ""
	if user.Phone != nil {
		phone = *user.Phone
	}

	identifier := ""
	if user.Identifier != nil {
		identifier = *user.Identifier
	}

	headimgurl := ""
	if user.Headimgurl != nil {
		headimgurl = *user.Headimgurl
	}

	// 转换标签数组（可能为空）
	tags := make([]string, 0)
	if user.Tags != nil {
		tags = user.Tags
	}

	return types.ProcessingUser{
		Id:         user.ID, // core 已经是 string（本地用户ID），直接赋值
		Name:       user.Name,
		Nickname:   nickname,
		Phone:      phone,
		Identifier: identifier,
		Headimgurl: headimgurl,
		Tags:       tags,
	}
}

// ConvertCoreProcessingUserListToTypes 转换Core ProcessingUser列表为API响应
func ConvertCoreProcessingUserListToTypes(users []*core.ProcessingUser) []types.ProcessingUser {
	result := make([]types.ProcessingUser, 0, len(users))
	for _, user := range users {
		result = append(result, ConvertCoreProcessingUserToTypes(user))
	}
	return result
}

// ========== Journey Full Detail 转换 ==========

// ConvertCoreFieldOptionToTypes 转换Core FieldOption为API响应
func ConvertCoreFieldOptionToTypes(option core.FieldOption) types.FieldOption {
	return types.FieldOption{
		Id:       option.ID,
		Value:    option.Value,
		Settings: option.Settings,
		Position: option.Position,
	}
}

// ConvertCoreVertexFieldToTypes 转换Core VertexField为API响应
func ConvertCoreVertexFieldToTypes(field *core.VertexField) types.VertexField {
	// 转换选项列表
	options := make([]types.FieldOption, 0, len(field.Options))
	for _, opt := range field.Options {
		options = append(options, ConvertCoreFieldOptionToTypes(opt))
	}

	return types.VertexField{
		Id:          field.ID,
		IdentityKey: field.IdentityKey,
		Title:       field.Title,
		Type:        field.Type,
		Required:    field.Required,
		Editable:    field.Editable,
		MaxLength:   field.MaxLength,
		Options:     options,
	}
}

// ConvertCorePendingNodeToTypes 转换Core PendingNode为API响应
func ConvertCorePendingNodeToTypes(node *core.PendingNode) types.PendingNode {
	// 转换字段列表
	var fields []types.VertexField
	if node.Fields != nil {
		fields = make([]types.VertexField, 0, len(node.Fields))
		for _, field := range node.Fields {
			if field != nil {
				fields = append(fields, ConvertCoreVertexFieldToTypes(field))
			}
		}
	} else {
		fields = []types.VertexField{}
	}

	return types.PendingNode{
		VertexId:    node.VertexID,
		VertexName:  node.VertexName,
		AssigneeIds: node.AssigneeIDs,
		CreatedAt:   node.CreatedAt,
		Fields:      fields,
	}
}

// ConvertCorePendingNodeListToTypes 转换Core PendingNode列表为API响应
func ConvertCorePendingNodeListToTypes(nodes []*core.PendingNode) []types.PendingNode {
	result := make([]types.PendingNode, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, ConvertCorePendingNodeToTypes(node))
	}
	return result
}

// ConvertCoreFlowVertexToTypes 转换Core FlowVertex为API响应
// 注意：core.FlowVertex.Name 映射到 types.FlowVertex.Title
func ConvertCoreFlowVertexToTypes(vertex *core.FlowVertex) types.FlowVertex {
	return types.FlowVertex{
		Id:    vertex.ID,
		Title: vertex.Name,  // core中是Name，types中是Title
		Type:  vertex.Type,
	}
}

// ConvertCoreFlowVertexMapToTypes 转换Core FlowVertex Map为API响应
// 注意：core 中的 Vertices 是 map[int64]*FlowVertex，API 中是 map[string]FlowVertex
func ConvertCoreFlowVertexMapToTypes(vertices map[int64]*core.FlowVertex) map[string]types.FlowVertex {
	result := make(map[string]types.FlowVertex, len(vertices))
	for id, vertex := range vertices {
		result[intToString(id)] = ConvertCoreFlowVertexToTypes(vertex)
	}
	return result
}

// ConvertCoreJourneyFullDetailToTypes 转换Core JourneyFullDetail为API响应
func ConvertCoreJourneyFullDetailToTypes(detail *core.JourneyFullDetail) types.JourneyFullDetail {
	return types.JourneyFullDetail{
		BasicInfo:    ConvertCoreJourneyDetailToTypes(detail.BasicInfo),
		History:      ConvertCoreMomentListToTypes(detail.History),
		PendingNodes: ConvertCorePendingNodeListToTypes(detail.PendingNodes),
		Vertices:     ConvertCoreFlowVertexMapToTypes(detail.Vertices),
	}
}

// intToString 将 int64 转换为字符串
func intToString(i int64) string {
	return fmt.Sprintf("%d", i)
}
