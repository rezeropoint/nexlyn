// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tag

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListTagsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询标签列表
func NewListTagsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTagsLogic {
	return &ListTagsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTagsLogic) ListTags(req *types.ListTagsRequest) (resp *types.ListTagsResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_tag"),
		logx.Field("operation", "list_tags"),
		logx.Field("status", "started"),
		logx.Field("page", req.Page),
		logx.Field("pageSize", req.PageSize),
	).Info("开始查询标签列表")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_tag"),
			logx.Field("operation", "list_tags"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.ListTagsResponse{
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
			logx.Field("operation", "list_tags"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.ListTagsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Message: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_tag"),
			logx.Field("operation", "list_tags"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.ListTagsResponse{
			BaseResponse: types.BaseResponse{Code: 403, Message: "权限不足：需要标签查看权限"},
		}, nil
	}

	// 构建查询条件
	query := core.LynxTagQuery{
		TenantID: req.TenantId,
		Scope:    req.Scope,
		Keyword:  req.Keyword,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// 记录查询参数
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_tag"),
		logx.Field("operation", "list_tags"),
		logx.Field("tenant_id", query.TenantID),
		logx.Field("scope", query.Scope),
		logx.Field("keyword", query.Keyword),
		logx.Field("page", query.Page),
		logx.Field("page_size", query.PageSize),
	).Info("查询标签列表参数")

	// 调用Manager查询标签列表
	tags, total, err := l.svcCtx.LynxManager.ListTags(l.ctx, query)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_tag"),
			logx.Field("operation", "list_tags"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("Manager查询标签列表失败")

		return &types.ListTagsResponse{
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
		logx.Field("operation", "list_tags"),
		logx.Field("status", "success"),
		logx.Field("count", len(tags)),
		logx.Field("total", total),
	).Info("查询标签列表成功")

	return &types.ListTagsResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
		PageParams: types.PageParams{
			Page:     query.Page,
			PageSize: query.PageSize,
			Total:    total,
		},
		Data: types.ListTagsData{
			List: data,
		},
	}, nil
}
