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

type DeleteInfoAtomTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除信息原子类型
func NewDeleteInfoAtomTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteInfoAtomTypeLogic {
	return &DeleteInfoAtomTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteInfoAtomTypeLogic) DeleteInfoAtomType(req *types.DeleteInfoAtomTypeRequest) (resp *types.DeleteInfoAtomTypeResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_infoatomtype"),
		logx.Field("operation", "delete_infoatomtype"),
		logx.Field("status", "started"),
		logx.Field("id", req.Id),
	).Info("开始删除信息原子类型")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_infoatomtype"),
			logx.Field("operation", "delete_infoatomtype"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.DeleteInfoAtomTypeResponse{
			BaseResponse: types.BaseResponse{
				Code:    401,
				Message: "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查信息原子类型删除权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceLynxInfoAtomType, Action: casbinxcore.ActionDelete},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_infoatomtype"),
			logx.Field("operation", "delete_infoatomtype"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.DeleteInfoAtomTypeResponse{
			BaseResponse: types.BaseResponse{Code: 500, Message: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_infoatomtype"),
			logx.Field("operation", "delete_infoatomtype"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.DeleteInfoAtomTypeResponse{
			BaseResponse: types.BaseResponse{Code: 403, Message: "权限不足：需要信息原子类型删除权限"},
		}, nil
	}

	// 调用Manager删除信息原子类型
	err = l.svcCtx.LynxManager.DeleteInfoAtomType(l.ctx, core.InfoAtomTypeKey{ID: req.Id})
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_infoatomtype"),
			logx.Field("operation", "delete_infoatomtype"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("Manager删除信息原子类型失败")

		return &types.DeleteInfoAtomTypeResponse{
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
		logx.Field("module", "lynxmanager_infoatomtype"),
		logx.Field("operation", "delete_infoatomtype"),
		logx.Field("status", "success"),
		logx.Field("id", req.Id),
	).Info("删除信息原子类型成功")

	return &types.DeleteInfoAtomTypeResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
	}, nil
}
