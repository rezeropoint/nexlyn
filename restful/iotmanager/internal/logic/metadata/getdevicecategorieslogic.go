package metadata

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeviceCategoriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取所有设备类别
func NewGetDeviceCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeviceCategoriesLogic {
	return &GetDeviceCategoriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDeviceCategoriesLogic) GetDeviceCategories(req *types.GetDeviceCategoriesRequest) (resp *types.GetDeviceCategoriesResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_metadata"),
		logx.Field("operation", "get_device_categories"),
		logx.Field("status", "started"),
	).Info("开始获取设备类别列表")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_metadata"),
			logx.Field("operation", "get_device_categories"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetDeviceCategoriesResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查元数据查看权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceIoTMetadata, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_metadata"),
			logx.Field("operation", "get_device_categories"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetDeviceCategoriesResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_metadata"),
			logx.Field("operation", "get_device_categories"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetDeviceCategoriesResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要元数据查看权限"},
		}, nil
	}

	// 从IoT引擎获取所有设备类别
	categories := l.svcCtx.IoTEngine.GetDeviceCategories()

	// 转换为响应类型
	respCategories := make([]types.DeviceCategoryInfo, 0, len(categories))
	for _, cat := range categories {
		respCategories = append(respCategories, types.DeviceCategoryInfo{
			Code:        string(cat.Code),
			Name:        cat.Name,
			NameEn:      cat.NameEn,
			Description: cat.Description,
			Icon:        cat.Icon,
		})
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_metadata"),
		logx.Field("operation", "get_device_categories"),
		logx.Field("status", "success"),
		logx.Field("count", len(respCategories)),
	).Info("获取设备类别列表成功")

	return &types.GetDeviceCategoriesResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: struct {
			Categories []types.DeviceCategoryInfo `json:"categories"`
		}{
			Categories: respCategories,
		},
	}, nil
}
