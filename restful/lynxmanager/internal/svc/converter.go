package svc

import (
	"context"
	"encoding/json"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// ConvertCoreInfoAtomTypeToTypes 转换 core.InfoAtomType 为 types.InfoAtomType（包含标签名称查询）
func (svc *ServiceContext) ConvertCoreInfoAtomTypeToTypes(ctx context.Context, infoAtomType core.InfoAtomType) types.InfoAtomType {
	// 转换 DataFormat
	dataFormat := types.DataFormat{
		DataPlural: infoAtomType.GetDataFormat().DataPlural,
		FieldStart: infoAtomType.GetDataFormat().FieldStart,
		Fields:     make([]types.FieldConfig, 0, len(infoAtomType.GetDataFormat().Fields)),
	}

	// 转换字段列表
	for _, field := range infoAtomType.GetDataFormat().Fields {
		dataFormat.Fields = append(dataFormat.Fields, types.FieldConfig{
			FieldKey:  field.FieldKey,
			FieldPath: field.FieldPath,
			FieldType: string(field.FieldType),
		})
	}

	result := types.InfoAtomType{
		Id:         infoAtomType.GetID(),
		TenantId:   infoAtomType.GetTenantId(),
		Name:       infoAtomType.GetName(),
		Version:    infoAtomType.GetVersion(),
		Tags:       make([]types.LynxTagSummary, 0), // Phase 2.5: 初始化为空切片
		DataFormat: dataFormat,
		CreatedBy:  infoAtomType.GetCreatedBy(),
		UpdatedBy:  infoAtomType.GetUpdatedBy(),
		CreatedAt:  infoAtomType.GetCreatedAt() * 1000, // 秒转毫秒
		UpdatedAt:  infoAtomType.GetUpdatedAt() * 1000, // 秒转毫秒
	}

	// Phase 2.5 重构：REST 层查询完整标签摘要对象（包含 id、name、description、scope 等）
	if len(infoAtomType.GetTagIDs()) > 0 {
		coreTagSummaries, err := svc.LynxManager.GetTagsByIDs(ctx, infoAtomType.GetTagIDs(), infoAtomType.GetTenantId())
		if err == nil {
			// 转换 []*core.LynxTagSummary → []types.LynxTagSummary
			result.Tags = make([]types.LynxTagSummary, 0, len(coreTagSummaries))
			for _, coreTag := range coreTagSummaries {
				result.Tags = append(result.Tags, types.LynxTagSummary{
					Id:          coreTag.ID,
					Name:        coreTag.Name,
					Description: coreTag.Description,
					Scope:       coreTag.Scope,
					CreatedAt:   coreTag.CreatedAt,
					UpdatedAt:   coreTag.UpdatedAt,
				})
			}
		} else {
			logx.Errorf("查询标签信息失败 (InfoAtomType ID=%s): %v", infoAtomType.GetID(), err)
			// 降级处理：保留空数组
		}
	}

	return result
}

// ConvertCoreInfoAtomTypesToTypes 批量转换 core.InfoAtomType 列表为 types.InfoAtomType 列表
func (svc *ServiceContext) ConvertCoreInfoAtomTypesToTypes(ctx context.Context, infoAtomTypes []core.InfoAtomType) []types.InfoAtomType {
	result := make([]types.InfoAtomType, 0, len(infoAtomTypes))
	for _, infoAtomType := range infoAtomTypes {
		result = append(result, svc.ConvertCoreInfoAtomTypeToTypes(ctx, infoAtomType))
	}
	return result
}

// ConvertCoreGraphConfigToMetadata 转换 core.GraphConfig 为 types.GraphConfigMetadata（包含标签名称查询）
func (svc *ServiceContext) ConvertCoreGraphConfigToMetadata(ctx context.Context, config core.GraphConfig) types.GraphConfigMetadata {
	metadata := types.GraphConfigMetadata{
		Id:          config.ID,
		TenantId:    config.TenantId,
		Name:        config.Name,
		Version:     config.Version,
		Description: config.Description,
		Tags:        make([]types.LynxTagSummary, 0), // Phase 2.5: 初始化为空切片
		IsEnabled:   config.Enable,
		Icon:        config.Icon,
		IconColor:   config.IconColor,
		CreatedAt:   config.CreatedAt * 1000, // 秒转毫秒
		UpdatedAt:   config.UpdatedAt * 1000, // 秒转毫秒
		CreatedBy:   config.CreatedBy,
		UpdatedBy:   config.UpdatedBy,
	}

	// Phase 2.5 重构：REST 层查询完整标签摘要对象（包含 id、name、description、scope 等）
	if len(config.TagIDs) > 0 {
		coreTagSummaries, err := svc.LynxManager.GetTagsByIDs(ctx, config.TagIDs, config.TenantId)
		if err == nil {
			// 转换 []*core.LynxTagSummary → []types.LynxTagSummary
			metadata.Tags = make([]types.LynxTagSummary, 0, len(coreTagSummaries))
			for _, coreTag := range coreTagSummaries {
				metadata.Tags = append(metadata.Tags, types.LynxTagSummary{
					Id:          coreTag.ID,
					Name:        coreTag.Name,
					Description: coreTag.Description,
					Scope:       coreTag.Scope,
					CreatedAt:   coreTag.CreatedAt,
					UpdatedAt:   coreTag.UpdatedAt,
				})
			}
		} else {
			logx.Errorf("查询标签信息失败 (GraphConfig ID=%s): %v", config.ID, err)
			// 降级处理：保留空数组
		}
	}

	return metadata
}

// ConvertCoreGraphConfigsToMetadata 批量转换 core.GraphConfig 列表为 types.GraphConfigMetadata 列表
func (svc *ServiceContext) ConvertCoreGraphConfigsToMetadata(ctx context.Context, configs []core.GraphConfig) []types.GraphConfigMetadata {
	result := make([]types.GraphConfigMetadata, 0, len(configs))
	for _, config := range configs {
		result = append(result, svc.ConvertCoreGraphConfigToMetadata(ctx, config))
	}
	return result
}

// ConvertCoreGraphConfigToDetail 转换 core.GraphConfig 为 types.GraphConfigDetail（包含节点和边）
func (svc *ServiceContext) ConvertCoreGraphConfigToDetail(ctx context.Context, config core.GraphConfig) types.GraphConfigDetail {
	// 转换节点列表
	nodes := make([]types.NodeConfig, 0, len(config.Nodes))
	for _, node := range config.Nodes {
		// 将 map[string]string 转换为 []string (标签)
		subscribedLabels := make([]string, 0)
		for key, value := range node.SubscribedLabels {
			subscribedLabels = append(subscribedLabels, key+":"+value)
		}

		// BlockConfig 从 map[string]any 转为 JSON 字符串
		blockConfigJSON := "{}"
		if node.BlockConfig != nil {
			if data, err := json.Marshal(node.BlockConfig); err == nil {
				blockConfigJSON = string(data)
			}
		}

		nodes = append(nodes, types.NodeConfig{
			Id:                        node.ID,
			Type:                      string(node.Type),
			BlockType:                 string(node.BlockType),
			BlockVersion:              node.BlockVersion,
			IsEntryPoint:              node.IsEntryPoint,
			SubscribedInfoAtomTypeIDs: node.SubscribedInfoAtomTypeIDs,
			SubscribedSource:          node.SubscribedSource,
			SubscribedLabels:          subscribedLabels,
			BlockConfig:               blockConfigJSON,
			X:                         node.X,
			Y:                         node.Y,
		})
	}

	// 转换边列表
	edges := make([]types.EdgeConfig, 0, len(config.Edges))
	for _, edge := range config.Edges {
		edges = append(edges, types.EdgeConfig{
			Id:        edge.ID,
			SourceID:  edge.SourceID,
			TargetID:  edge.TargetID,
			Condition: edge.Condition,
		})
	}

	return types.GraphConfigDetail{
		GraphConfigMetadata: svc.ConvertCoreGraphConfigToMetadata(ctx, config),
		Nodes:               nodes,
		Edges:               edges,
	}
}
