package logic

import (
	"context"

	"github.com/rezeropoint/nexlyn/service/iotquery/internal/svc"
	"github.com/rezeropoint/nexlyn/service/iotquery/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetLatestValuesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLatestValuesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestValuesLogic {
	return &GetLatestValuesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetLatestValues 获取设备字段最新值
func (l *GetLatestValuesLogic) GetLatestValues(in *pb.GetLatestValuesReq) (*pb.GetLatestValuesResp, error) {
	// 1. 转换protobuf请求 → core.LatestValuesQuery
	query := svc.ConvertLatestValuesReqToCore(in)

	// 2. 调用查询引擎
	result, err := l.svcCtx.QueryEngine.GetLatestValues(l.ctx, query)
	if err != nil {
		logx.Errorf("Failed to get latest values: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 3. 转换结果 → protobuf响应
	return svc.ConvertLatestValuesToProto(result), nil
}
