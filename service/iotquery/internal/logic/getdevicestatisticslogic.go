package logic

import (
	"context"

	"github.com/rezeropoint/nexlyn/service/iotquery/pb"

	"github.com/rezeropoint/nexlyn/service/iotquery/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetDeviceStatisticsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDeviceStatisticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeviceStatisticsLogic {
	return &GetDeviceStatisticsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetDeviceStatistics 获取设备统计信息
func (l *GetDeviceStatisticsLogic) GetDeviceStatistics(in *pb.GetDeviceStatisticsReq) (*pb.GetDeviceStatisticsResp, error) {
	// 1. 转换protobuf请求 → core.DeviceStatisticsQuery
	query := svc.ConvertDeviceStatisticsReqToCore(in)

	// 2. 调用查询引擎
	result, err := l.svcCtx.QueryEngine.GetDeviceStatistics(l.ctx, query)
	if err != nil {
		logx.Errorf("Failed to get device statistics: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 3. 转换结果 → protobuf响应
	return svc.ConvertDeviceStatisticsToProto(result), nil
}
