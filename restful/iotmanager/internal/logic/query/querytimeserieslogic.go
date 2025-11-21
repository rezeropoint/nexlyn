package query

import (
	"context"
	"time"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryTimeSeriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询时序数据
func NewQueryTimeSeriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryTimeSeriesLogic {
	return &QueryTimeSeriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryTimeSeriesLogic) QueryTimeSeries(req *types.QueryTimeSeriesRequest) (resp *types.QueryTimeSeriesResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_query"),
		logx.Field("operation", "query_timeseries"),
		logx.Field("status", "started"),
	).Info("开始查询时序数据")

	// 1. 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "query_timeseries"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.QueryTimeSeriesResponse{
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
			logx.Field("operation", "query_timeseries"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.QueryTimeSeriesResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "query_timeseries"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.QueryTimeSeriesResponse{
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
			logx.Field("operation", "query_timeseries"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询子组织失败")

		return &types.QueryTimeSeriesResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "查询组织失败: " + err.Error()},
		}, nil
	}

	// 4. 解析时间参数
	startTime, err := svc.ParseTimeString(req.StartTime)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "query_timeseries"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("解析开始时间失败")

		return &types.QueryTimeSeriesResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "开始时间格式错误: " + err.Error()},
		}, nil
	}

	endTime, err := svc.ParseTimeString(req.EndTime)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "query_timeseries"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("解析结束时间失败")

		return &types.QueryTimeSeriesResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "结束时间格式错误: " + err.Error()},
		}, nil
	}

	// 5. 构建core.TimeSeriesQuery
	query := core.TimeSeriesQuery{
		TenantID:    jwtUser.TenantId,
		OrgIDs:      orgIds,
		DeviceIDs:   req.DeviceIds,
		FieldNames:  svc.ConvertFieldNamesToCore(req.FieldNames),
		StartTime:   startTime,
		EndTime:     endTime,
		Aggregation: core.AggregationType(req.Aggregation),
		Interval:    time.Duration(req.Interval) * time.Second,
		Limit:       int(req.Limit),
		Offset:      int(req.Offset),
		OrderBy:     req.OrderBy,
		OrderDir:    req.OrderDir,
	}

	// 6. 调用IoT引擎查询时序数据
	result, err := l.svcCtx.IoTEngine.QueryTimeSeries(l.ctx, query)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_query"),
			logx.Field("operation", "query_timeseries"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("IoT引擎查询时序数据失败")

		return &types.QueryTimeSeriesResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 7. 转换结果为API类型
	list := svc.ConvertCoreTimeSeriesResultToTypes(result)

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_query"),
		logx.Field("operation", "query_timeseries"),
		logx.Field("status", "success"),
		logx.Field("total", result.Total),
		logx.Field("returned", len(list)),
	).Info("查询时序数据成功")

	return &types.QueryTimeSeriesResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: struct {
			List     []types.TimeSeriesDataItem `json:"list"`
			Total    int64                      `json:"total"`
			Page     int64                      `json:"page"`
			PageSize int64                      `json:"pageSize"`
		}{
			List:     list,
			Total:    result.Total,
			Page:     int64(result.Page),
			PageSize: int64(result.PageSize),
		},
	}, nil
}
