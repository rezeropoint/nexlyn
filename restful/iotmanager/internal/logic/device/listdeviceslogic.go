package device

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListDevicesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取设备列表
func NewListDevicesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDevicesLogic {
	return &ListDevicesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListDevicesLogic) ListDevices(req *types.ListDevicesRequest) (resp *types.ListDevicesResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_device"),
		logx.Field("operation", "list_devices"),
		logx.Field("status", "started"),
		logx.Field("page", req.Current),
		logx.Field("page_size", req.PageSize),
	).Info("开始获取设备列表")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device"),
			logx.Field("operation", "list_devices"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.ListDevicesResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查设备列表查看权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceIoTDevice, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device"),
			logx.Field("operation", "list_devices"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.ListDevicesResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device"),
			logx.Field("operation", "list_devices"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.ListDevicesResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要设备查看权限"},
		}, nil
	}

	// 规范化分页参数
	current, pageSize, _ := svc.NormalizePagination(req.Current, req.PageSize)

	// 构建查询条件（直接使用前端传入的必填组织ID）
	query := core.DeviceBindingQuery{
		TenantID:       jwtUser.TenantId,
		OrgID:          req.OrgId, // 使用前端传入的组织ID（必填）
		DeviceModel:    req.DeviceModel,
		DeviceCategory: req.DeviceCategory,
		Status:         req.Status,
		Keyword:        req.Keyword,
		Page:           int(current),
		PageSize:       int(pageSize),
	}

	// 处理IsOnline筛选参数（字符串转bool）
	if req.IsOnline != "" {
		isOnline := req.IsOnline == "true"
		query.IsOnline = &isOnline
	}

	// 调用IoT引擎查询设备列表
	summaries, total, err := l.svcCtx.IoTEngine.ListDevices(l.ctx, query)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device"),
			logx.Field("operation", "list_devices"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("IoT引擎查询设备列表失败")

		return &types.ListDevicesResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 转换为API类型
	list := svc.ConvertCoreDeviceBindingSummariesToTypes(summaries)

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_device"),
		logx.Field("operation", "list_devices"),
		logx.Field("status", "success"),
		logx.Field("total", total),
		logx.Field("page", current),
		logx.Field("page_size", pageSize),
	).Info("获取设备列表成功")

	return &types.ListDevicesResponse{
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
			List []types.DeviceBindingSummary `json:"list"`
		}{
			List: list,
		},
	}, nil
}
