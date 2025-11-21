package logic

import (
	"context"

	"github.com/rezeropoint/nexlyn/service/iotquery/internal/svc"
	"github.com/rezeropoint/nexlyn/service/iotquery/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type QueryTimeSeriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryTimeSeriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryTimeSeriesLogic {
	return &QueryTimeSeriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryTimeSeries 查询时序数据
func (l *QueryTimeSeriesLogic) QueryTimeSeries(in *pb.QueryTimeSeriesReq) (*pb.QueryTimeSeriesResp, error) {
	// 1. 转换protobuf请求 → core.TimeSeriesQuery
	query := svc.ConvertTimeSeriesReqToCore(in)

	// 2. 调用查询引擎（引擎内部会验证参数）
	result, err := l.svcCtx.QueryEngine.QueryTimeSeries(l.ctx, query)
	if err != nil {
		logx.Errorf("Failed to query time series: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 3. 转换结果 → protobuf响应
	return svc.ConvertTimeSeriesResultToProto(result), nil
}
