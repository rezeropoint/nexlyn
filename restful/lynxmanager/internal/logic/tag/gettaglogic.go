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

type GetTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取标签详情
func NewGetTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTagLogic {
	return &GetTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTagLogic) GetTag(req *types.GetTagRequest) (resp *types.GetTagResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_tag"),
		logx.Field("operation", "get_tag"),
		logx.Field("status", "started"),
		logx.Field("tag_id", req.Id),
	).Info("开始获取标签详情")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_tag"),
			logx.Field("operation", "get_tag"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetTagResponse{
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
			logx.Field("operation", "get_tag"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetTagResponse{
			BaseResponse: types.BaseResponse{Code: 500, Message: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_tag"),
			logx.Field("operation", "get_tag"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetTagResponse{
			BaseResponse: types.BaseResponse{Code: 403, Message: "权限不足：需要标签查看权限"},
		}, nil
	}

	// 调用Manager获取标签
	tag, err := l.svcCtx.LynxManager.GetTag(l.ctx, req.Id, jwtUser.TenantId)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_tag"),
			logx.Field("operation", "get_tag"),
			logx.Field("status", "failed"),
			logx.Field("tag_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("Manager获取标签失败")

		return &types.GetTagResponse{
			BaseResponse: types.BaseResponse{Code: 500, Message: "服务器错误: " + err.Error()},
		}, nil
	}

	if tag == nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_tag"),
			logx.Field("operation", "get_tag"),
			logx.Field("status", "not_found"),
			logx.Field("tag_id", req.Id),
		).Error("标签不存在")

		return &types.GetTagResponse{
			BaseResponse: types.BaseResponse{Code: 404, Message: "标签不存在"},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_tag"),
		logx.Field("operation", "get_tag"),
		logx.Field("status", "success"),
		logx.Field("tag_id", req.Id),
	).Info("获取标签详情成功")

	return &types.GetTagResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Data: types.LynxTag{
			Id:          tag.ID,
			Name:        tag.Name,
			Description: tag.Description,
			Scope:       tag.Scope,
			TenantId:    tag.TenantID,
			CreatedBy:   tag.CreatedBy,
			UpdatedBy:   tag.UpdatedBy,
			CreatedAt:   tag.CreatedAt,
			UpdatedAt:   tag.UpdatedAt,
		},
	}, nil
}
