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

type CreateInfoAtomTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建信息原子类型
func NewCreateInfoAtomTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateInfoAtomTypeLogic {
	return &CreateInfoAtomTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateInfoAtomTypeLogic) CreateInfoAtomType(req *types.CreateInfoAtomTypeRequest) (resp *types.CreateInfoAtomTypeResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "lynxmanager_infoatomtype"),
		logx.Field("operation", "create_infoatomtype"),
		logx.Field("status", "started"),
		logx.Field("name", req.Name),
		logx.Field("version", req.Version),
	).Info("开始创建信息原子类型")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_infoatomtype"),
			logx.Field("operation", "create_infoatomtype"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.CreateInfoAtomTypeResponse{
			BaseResponse: types.BaseResponse{
				Code:    401,
				Message: "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查信息原子类型创建权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceLynxInfoAtomType, Action: casbinxcore.ActionWrite},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_infoatomtype"),
			logx.Field("operation", "create_infoatomtype"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.CreateInfoAtomTypeResponse{
			BaseResponse: types.BaseResponse{Code: 500, Message: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_infoatomtype"),
			logx.Field("operation", "create_infoatomtype"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.CreateInfoAtomTypeResponse{
			BaseResponse: types.BaseResponse{Code: 403, Message: "权限不足：需要信息原子类型管理权限"},
		}, nil
	}

	// 转换字段配置
	fields := make([]core.FieldConfig, 0, len(req.DataFormat.Fields))
	for _, field := range req.DataFormat.Fields {
		fields = append(fields, core.FieldConfig{
			FieldKey:  field.FieldKey,
			FieldPath: field.FieldPath,
			FieldType: core.FieldType(field.FieldType),
		})
	}

	// 构建InfoAtomType
	infoAtomType := core.NewInfoAtomType(
		req.TenantId,
		req.Name,
		req.Version,
		req.TagIds,
		core.DataFormat{
			DataPlural: req.DataFormat.DataPlural,
			FieldStart: req.DataFormat.FieldStart,
			Fields:     fields,
		},
	)

	// 设置创建人（从JWT获取）
	if basicType, ok := infoAtomType.(*core.BasicInfoAtomType); ok {
		basicType.CreatedBy = jwtUser.UserKey
	}

	// 调用Manager创建信息原子类型
	err = l.svcCtx.LynxManager.CreateInfoAtomType(l.ctx, infoAtomType)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "lynxmanager_infoatomtype"),
			logx.Field("operation", "create_infoatomtype"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("Manager创建信息原子类型失败")

		return &types.CreateInfoAtomTypeResponse{
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
		logx.Field("operation", "create_infoatomtype"),
		logx.Field("status", "success"),
		logx.Field("id", infoAtomType.GetID()),
	).Info("创建信息原子类型成功")

	return &types.CreateInfoAtomTypeResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Data: struct {
			Id string `json:"id"`
		}{
			Id: infoAtomType.GetID(),
		},
	}, nil
}
