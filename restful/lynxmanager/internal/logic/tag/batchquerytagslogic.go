// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tag

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type BatchQueryTagsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 批量查询标签（用于自动补全）
func NewBatchQueryTagsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchQueryTagsLogic {
	return &BatchQueryTagsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchQueryTagsLogic) BatchQueryTags(req *types.BatchQueryTagsRequest) (resp *types.BatchQueryTagsResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_tag"),
		logx.Field("operation", "batch_query_tags"),
		logx.Field("status", "started"),
		logx.Field("scope", req.Scope),
		logx.Field("name_count", len(req.Names)),
	).Info("开始批量查询标签")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_tag"),
			logx.Field("operation", "batch_query_tags"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.BatchQueryTagsResponse{
			BaseResponse: types.BaseResponse{
				Code:    401,
				Message: "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查标签读权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceLynxTag, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_tag"),
			logx.Field("operation", "batch_query_tags"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.BatchQueryTagsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Message: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_tag"),
			logx.Field("operation", "batch_query_tags"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.BatchQueryTagsResponse{
			BaseResponse: types.BaseResponse{Code: 403, Message: "权限不足：需要标签查看权限"},
		}, nil
	}

	// 调用Manager批量查询标签
	tags, err := l.svcCtx.LynxManager.GetTagsByNames(l.ctx, req.TenantId, req.Scope, req.Names)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_tag"),
			logx.Field("operation", "batch_query_tags"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("Manager批量查询标签失败")

		return &types.BatchQueryTagsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Message: "服务器错误: " + err.Error()},
		}, nil
	}

	// 转换结果
	data := make([]types.LynxTagSummary, len(tags))
	for i, tag := range tags {
		data[i] = types.LynxTagSummary{
			Id:          tag.ID,
			Name:        tag.Name,
			Description: tag.Description,
			Scope:       tag.Scope,
			CreatedAt:   tag.CreatedAt,
			UpdatedAt:   tag.UpdatedAt,
		}
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_tag"),
		logx.Field("operation", "batch_query_tags"),
		logx.Field("status", "success"),
		logx.Field("matched_count", len(tags)),
	).Info("批量查询标签成功")

	return &types.BatchQueryTagsResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Data: data,
	}, nil
}
