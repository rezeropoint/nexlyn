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

type CreateGraphConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建逻辑图配置
func NewCreateGraphConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGraphConfigLogic {
	return &CreateGraphConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateGraphConfigLogic) CreateGraphConfig(req *types.CreateGraphConfigRequest) (resp *types.CreateGraphConfigResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_graphconfig"),
		logx.Field("operation", "create_graphconfig"),
		logx.Field("status", "started"),
		logx.Field("name", req.Name),
		logx.Field("version", req.Version),
	).Info("开始创建逻辑图配置")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "create_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.CreateGraphConfigResponse{
			BaseResponse: types.BaseResponse{
				Code:    401,
				Message: "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查逻辑图配置创建权限
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
			logx.Field("operation", "create_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.CreateGraphConfigResponse{
			BaseResponse: types.BaseResponse{Code: 500, Message: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "create_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.CreateGraphConfigResponse{
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
					logx.Field("operation", "create_graphconfig"),
					logx.Field("status", "failed"),
					logx.Field("error", err.Error()),
					logx.Field("node_id", node.Id),
				).Error("节点BlockConfig JSON解析失败")

				return &types.CreateGraphConfigResponse{
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

	// 自动填充OrgId（从JWT获取，确保用户只能为自己的组织创建逻辑图）
	orgId := req.OrgId
	if orgId == "" {
		orgId = jwtUser.PrimaryOrgId
	}

	// 构建GraphConfig
	graphConfig := core.GraphConfig{
		TenantId:    req.TenantId,
		OrgID:       orgId, // 使用自动填充的组织ID
		Name:        req.Name,
		Version:     req.Version,
		Description: req.Description,
		TagIDs:      req.TagIds,
		Enable:      req.IsEnabled,
		Icon:        req.Icon,
		IconColor:   req.IconColor,
		Nodes:       nodes,
		Edges:       edges,
		CreatedBy:   jwtUser.UserId, // 使用UserId（UUID）而非UserKey
	}

	// 调用Manager创建逻辑图配置
	err = l.svcCtx.LynxManager.CreateGraphConfig(l.ctx, graphConfig)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_graphconfig"),
			logx.Field("operation", "create_graphconfig"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("Manager创建逻辑图配置失败")

		return &types.CreateGraphConfigResponse{
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
		logx.Field("operation", "create_graphconfig"),
		logx.Field("status", "success"),
		logx.Field("id", graphConfig.ID),
	).Info("创建逻辑图配置成功")

	return &types.CreateGraphConfigResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Data: struct {
			Id string `json:"id"`
		}{
			Id: graphConfig.ID,
		},
	}, nil
}
