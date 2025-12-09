// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package graphconfig

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateGraphConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新逻辑图配置
func NewUpdateGraphConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGraphConfigLogic {
	return &UpdateGraphConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateGraphConfigLogic) UpdateGraphConfig(req *types.UpdateGraphConfigRequest) (resp *types.UpdateGraphConfigResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_graphconfig"),
		logx.Field("operation", "update_graphconfig"),
		logx.Field("status", "started"),
		logx.Field("id", req.Id),
	).Info("开始更新逻辑图配置")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "update_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.UpdateGraphConfigResponse{
			BaseResponse: types.BaseResponse{
				Code:    401,
				Message: "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查逻辑图配置更新权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceLynxGraphConfig, Action: casbinxcore.ActionWrite},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "update_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.UpdateGraphConfigResponse{
			BaseResponse: types.BaseResponse{Code: 500, Message: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "update_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.UpdateGraphConfigResponse{
			BaseResponse: types.BaseResponse{Code: 403, Message: "权限不足：需要逻辑图配置管理权限"},
		}, nil
	}

	// 转换节点配置
	nodes := make([]core.NodeConfig, 0, len(req.Nodes))
	for _, node := range req.Nodes {
		// 将 []string 转换为 map[string]string (标签)
		// 格式：["key1:value1", "key2:value2"] → {"key1": "value1", "key2": "value2"}
		subscribedLabels := make(map[string]string)
		for _, label := range node.SubscribedLabels {
			// 解析 "key:value" 格式
			parts := strings.SplitN(label, ":", 2)
			if len(parts) == 2 {
				subscribedLabels[parts[0]] = parts[1]
			} else {
				// 如果没有冒号，使用标签本身作为key和value
				subscribedLabels[label] = label
			}
		}

		// BlockConfig 从 JSON 字符串转为 map[string]any
		var blockConfig map[string]any
		if node.BlockConfig != "" {
			if err := json.Unmarshal([]byte(node.BlockConfig), &blockConfig); err != nil {
				logx.WithContext(l.ctx).WithFields(
					logx.Field("service", l.svcCtx.Config.RestConf.Name),
					logx.Field("pod", l.svcCtx.PodName),
					logx.Field("module", "lynxmanager_graphconfig"),
					logx.Field("operation", "update_graphconfig"),
					logx.Field("status", "failed"),
					logx.Field("error", err.Error()),
					logx.Field("node_id", node.Id),
				).Error("节点BlockConfig JSON解析失败")

				return &types.UpdateGraphConfigResponse{
					BaseResponse: types.BaseResponse{
						Code:    400,
						Message: "节点配置格式错误: " + err.Error(),
					},
				}, nil
			}
		}

		nodes = append(nodes, core.NodeConfig{
			ID:                        node.Id,
			Type:                      core.NodeType(node.Type),
			BlockType:                 core.LogicBlockType(node.BlockType),
			BlockVersion:              node.BlockVersion,
			BlockConfig:               blockConfig,
			IsEntryPoint:              node.IsEntryPoint,
			SubscribedInfoAtomTypeIDs: node.SubscribedInfoAtomTypeIDs,
			SubscribedSource:          node.SubscribedSource,
			SubscribedLabels:          subscribedLabels,
			X:                         node.X,
			Y:                         node.Y,
		})
	}

	// 转换边配置
	edges := make([]core.EdgeConfig, 0, len(req.Edges))
	for _, edge := range req.Edges {
		edges = append(edges, core.EdgeConfig{
			ID:        edge.Id,
			SourceID:  edge.SourceID,
			TargetID:  edge.TargetID,
			Condition: edge.Condition,
		})
	}

	// 构建GraphConfig
	graphConfig := core.GraphConfig{
		ID:          req.Id,
		TenantId:    req.TenantId,
		Name:        req.Name,
		Version:     req.Version,
		Description: req.Description,
		TagIDs:      req.TagIds,
		Enable:      req.IsEnabled,
		Icon:        req.Icon,
		IconColor:   req.IconColor,
		Nodes:       nodes,
		Edges:       edges,
		UpdatedBy:   jwtUser.UserId, // 使用UserId（UUID）而非UserKey
	}

	// 验证所有节点的配置（包括必填字段校验）
	for _, nodeConfig := range nodes {
		if err := l.svcCtx.LynxManager.ValidateNodeConfig(nodeConfig); err != nil {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "lynxmanager_graphconfig"),
				logx.Field("operation", "update_graphconfig"),
				logx.Field("status", "failed"),
				logx.Field("node_id", nodeConfig.ID),
				logx.Field("block_type", nodeConfig.BlockType),
				logx.Field("error", err.Error()),
			).Error("节点配置验证失败")

			return &types.UpdateGraphConfigResponse{
				BaseResponse: types.BaseResponse{
					Code:    400,
					Message: "节点配置验证失败: " + err.Error(),
				},
			}, nil
		}
	}

	// 调用Manager更新逻辑图配置
	err = l.svcCtx.LynxManager.UpdateGraphConfig(l.ctx, graphConfig)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "update_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("Manager更新逻辑图配置失败")

		return &types.UpdateGraphConfigResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "服务器错误: " + err.Error(),
			},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_graphconfig"),
		logx.Field("operation", "update_graphconfig"),
		logx.Field("status", "success"),
		logx.Field("id", req.Id),
	).Info("更新逻辑图配置成功")

	return &types.UpdateGraphConfigResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
	}, nil
}
