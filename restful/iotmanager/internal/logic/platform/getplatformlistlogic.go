package platform

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetPlatformListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取平台配置列表
func NewGetPlatformListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPlatformListLogic {
	return &GetPlatformListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPlatformListLogic) GetPlatformList(req *types.GetPlatformListRequest) (resp *types.GetPlatformListResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_platform"),
		logx.Field("operation", "list_platforms"),
		logx.Field("status", "started"),
		logx.Field("page", req.Current),
		logx.Field("page_size", req.PageSize),
	).Info("开始获取平台配置列表")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_platform"),
			logx.Field("operation", "list_platforms"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetPlatformListResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查平台配置查看权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceIoTPlatform, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_platform"),
			logx.Field("operation", "list_platforms"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetPlatformListResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_platform"),
			logx.Field("operation", "list_platforms"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetPlatformListResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要平台配置查看权限"},
		}, nil
	}

	// 规范化分页参数
	current, pageSize, _ := svc.NormalizePagination(req.Current, req.PageSize)

	// 构建平台类型过滤（如果提供）
	var platformType *core.PlatformType
	if req.Type != "" {
		pt := core.PlatformType(req.Type)
		platformType = &pt
	}

	// 调用IoT引擎查询平台配置列表
	platformList, total, err := l.svcCtx.IoTEngine.ListPlatforms(
		l.ctx,
		jwtUser.TenantId,
		platformType,
		req.Keyword,
		int(current),
		int(pageSize),
	)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_platform"),
			logx.Field("operation", "list_platforms"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("IoT引擎查询平台配置列表失败")

		return &types.GetPlatformListResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 转换为API类型
	list := svc.ConvertCorePlatformListToTypes(platformList)

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_platform"),
		logx.Field("operation", "list_platforms"),
		logx.Field("status", "success"),
		logx.Field("total", total),
		logx.Field("page", current),
		logx.Field("page_size", pageSize),
	).Info("获取平台配置列表成功")

	return &types.GetPlatformListResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		PageParams: types.PageParams{
			Current:  current,
			PageSize: pageSize,
			Total:    total,
		},
		Data: struct {
			List []types.PlatformMetadata `json:"list"`
		}{
			List: list,
		},
	}, nil
}
