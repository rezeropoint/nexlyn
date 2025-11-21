package query

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestValuesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取设备字段最新值
func NewGetLatestValuesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestValuesLogic {
	return &GetLatestValuesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLatestValuesLogic) GetLatestValues(req *types.GetLatestValuesRequest) (resp *types.GetLatestValuesResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_query"),
		logx.Field("operation", "get_latest_values"),
		logx.Field("status", "started"),
	).Info("开始获取设备最新值")

	// 1. 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "get_latest_values"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetLatestValuesResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 2. 权限验证：检查设备数据查询权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceIoTDevice, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "get_latest_values"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetLatestValuesResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "get_latest_values"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetLatestValuesResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要设备查看权限"},
		}, nil
	}

	// 3. 查询所有子组织ID（包括当前组织）
	orgIds, err := auth.GetAllChildOrgIds(l.ctx, l.svcCtx.DBConn, req.OrgId, jwtUser.TenantId)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "get_latest_values"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询子组织失败")

		return &types.GetLatestValuesResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "查询组织失败: " + err.Error()},
		}, nil
	}

	// 4. 构建core.LatestValuesQuery
	query := core.LatestValuesQuery{
		TenantID:   jwtUser.TenantId,
		OrgIDs:     orgIds,
		DeviceIDs:  req.DeviceIds,
		FieldNames: svc.ConvertFieldNamesToCore(req.FieldNames),
	}

	// 5. 调用IoT引擎获取最新值
	result, err := l.svcCtx.IoTEngine.GetLatestValues(l.ctx, query)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "get_latest_values"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("IoT引擎获取最新值失败")

		return &types.GetLatestValuesResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 6. 转换结果为API类型
	data := svc.ConvertCoreDeviceLatestValuesToTypes(result)

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_query"),
		logx.Field("operation", "get_latest_values"),
		logx.Field("status", "success"),
		logx.Field("device_count", len(data)),
	).Info("获取设备最新值成功")

	return &types.GetLatestValuesResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: data,
	}, nil
}
