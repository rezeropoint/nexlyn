// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package block

import (
	"context"
	"encoding/json"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlockSpecsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询逻辑块规格列表
func NewGetBlockSpecsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlockSpecsLogic {
	return &GetBlockSpecsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBlockSpecsLogic) GetBlockSpecs(req *types.GetBlockSpecsRequest) (resp *types.GetBlockSpecsResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_block"),
		logx.Field("operation", "get_block_specs"),
		logx.Field("status", "started"),
	).Info("开始获取逻辑块规格列表")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_block"),
			logx.Field("operation", "get_block_specs"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetBlockSpecsResponse{
			BaseResponse: types.BaseResponse{
				Code:    401,
				Message: "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查逻辑块规格查看权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceLynxBlockSpec, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_block"),
			logx.Field("operation", "get_block_specs"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetBlockSpecsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Message: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_block"),
			logx.Field("operation", "get_block_specs"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetBlockSpecsResponse{
			BaseResponse: types.BaseResponse{Code: 403, Message: "权限不足：需要逻辑块规格查看权限"},
		}, nil
	}

	// 调用Manager获取所有逻辑块规格
	blockKeys := l.svcCtx.LynxManager.GetBlockKeyAll()

	// 构建返回列表（包含完整的BlockSpec信息）
	list := make([]types.BlockSpec, 0, len(blockKeys))
	for _, key := range blockKeys {
		// 获取完整的逻辑块规格
		spec, err := l.svcCtx.LynxManager.GetBlockSpec(key)
		if err != nil {
			// 如果获取规格失败，记录日志但继续处理其他逻辑块
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "lynxmanager_block"),
				logx.Field("operation", "get_block_specs"),
				logx.Field("blockType", key.BlockType),
				logx.Field("version", key.Version),
				logx.Field("error", err.Error()),
			).Error("获取逻辑块规格失败")
			continue
		}

		// 获取分类（使用Tags的第一个作为主分类）
		category := ""
		tags := spec.Tags()
		if len(tags) > 0 {
			category = tags[0]
		}

		// 序列化ConfigSchema为JSON字符串
		configSchemaStr := ""
		if configSchema := spec.ConfigSchema(); configSchema != nil {
			if data, err := json.Marshal(configSchema); err == nil {
				configSchemaStr = string(data)
			}
		}

		// 将core.BlockSpec转换为types.BlockSpec
		list = append(list, types.BlockSpec{
			BlockType:    string(key.BlockType),
			Version:      key.Version,
			Name:         spec.Name(),
			Description:  spec.Description(),
			Category:     category,
			InputSchema:  "",              // 暂不实现InputSchema序列化
			OutputSchema: "",              // 暂不实现OutputSchema序列化
			ConfigSchema: configSchemaStr, // 序列化为JSON字符串
		})
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_block"),
		logx.Field("operation", "get_block_specs"),
		logx.Field("status", "success"),
		logx.Field("count", len(list)),
	).Info("获取逻辑块规格列表成功")

	return &types.GetBlockSpecsResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Data: struct {
			List []types.BlockSpec `json:"list"`
		}{
			List: list,
		},
	}, nil
}
