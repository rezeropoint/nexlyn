// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package infoatomtype

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetInfoAtomTypeListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询信息原子类型列表
func NewGetInfoAtomTypeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInfoAtomTypeListLogic {
	return &GetInfoAtomTypeListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetInfoAtomTypeListLogic) GetInfoAtomTypeList(req *types.GetInfoAtomTypeListRequest) (resp *types.GetInfoAtomTypeListResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_infoatomtype"),
		logx.Field("operation", "list_infoatomtypes"),
		logx.Field("status", "started"),
		logx.Field("page", req.Page),
		logx.Field("page_size", req.PageSize),
		logx.Field("tenant_id", req.TenantId),
	).Info("开始获取信息原子类型列表")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_infoatomtype"),
			logx.Field("operation", "list_infoatomtypes"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetInfoAtomTypeListResponse{
			BaseResponse: types.BaseResponse{
				Code:    401,
				Message: "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查信息原子类型列表查看权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceLynxInfoAtomType, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_infoatomtype"),
			logx.Field("operation", "list_infoatomtypes"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetInfoAtomTypeListResponse{
			BaseResponse: types.BaseResponse{Code: 500, Message: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_infoatomtype"),
			logx.Field("operation", "list_infoatomtypes"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetInfoAtomTypeListResponse{
			BaseResponse: types.BaseResponse{Code: 403, Message: "权限不足：需要信息原子类型查看权限"},
		}, nil
	}

	// 规范化分页参数
	page, pageSize, _ := svc.NormalizePagination(req.Page, req.PageSize)

	// 构建查询条件
	query := core.InfoAtomList{
		Page:     int64(page),
		PageSize: int64(pageSize),
		TenantId: req.TenantId,
		Name:     req.Name,
		Tags:     req.Tags,
	}

	// 调用Manager查询信息原子类型列表
	infoAtomTypes, total, _, _, err := l.svcCtx.LynxManager.GetInfoAtomTypeList(l.ctx, query)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_infoatomtype"),
			logx.Field("operation", "list_infoatomtypes"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("Manager查询信息原子类型列表失败")

		return &types.GetInfoAtomTypeListResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "服务器错误: " + err.Error(),
			},
		}, nil
	}

	// Phase 2 重构：转换为API类型（包含标签名称）
	list := l.svcCtx.ConvertCoreInfoAtomTypesToTypes(l.ctx, infoAtomTypes)

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_infoatomtype"),
		logx.Field("operation", "list_infoatomtypes"),
		logx.Field("status", "success"),
		logx.Field("total", total),
		logx.Field("page", page),
		logx.Field("page_size", pageSize),
	).Info("获取信息原子类型列表成功")

	return &types.GetInfoAtomTypeListResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
		PageParams: types.PageParams{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
		Data: struct {
			List []types.InfoAtomType `json:"list"`
		}{
			List: list,
		},
	}, nil
}
