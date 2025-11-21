package logic

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/rezeropoint/nexlyn/service/lynxengine/internal/svc"
	"github.com/rezeropoint/nexlyn/service/lynxengine/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReceiveInfoAtomLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReceiveInfoAtomLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReceiveInfoAtomLogic {
	return &ReceiveInfoAtomLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ReceiveInfoAtom 接收信息原子
//
// 职责：
//  1. gRPC类型 → Core类型转换
//  2. 调用 Engine 统一接口
//  3. Core类型 → gRPC类型转换
//
// 注意：
//   - 这是轻量协议转换层，不包含任何业务逻辑
//   - 所有业务逻辑（字段提取、验证、构造、分发）由 Engine 层和 InfoAtomRegistry 层处理
func (l *ReceiveInfoAtomLogic) ReceiveInfoAtom(in *pb.InfoAtomRequest) (*pb.InfoAtomResponse, error) {
	// 1. gRPC类型 → Core类型转换
	req := core.InfoAtomRequest{
		TenantID:       in.TenantId,
		InfoAtomTypeID: in.InfoAtomTypeId,
		Source:         in.Source,
		RawData:        in.RawData,
		Tags:           in.Tags,
		Timestamp:      in.Timestamp,
	}

	// 2. 调用 Engine 统一接口
	resp, err := l.svcCtx.LynxEngine.ReceiveInfoAtom(l.ctx, req)
	if err != nil {
		l.Errorf("引擎处理失败: %v", err)
		return &pb.InfoAtomResponse{
			Success:   false,
			Message:   err.Error(),
			ErrorCode: "ENGINE_ERROR",
		}, nil
	}

	// 3. Core类型 → gRPC类型转换
	return &pb.InfoAtomResponse{
		Success:    resp.Success,
		Message:    resp.Message,
		InfoAtomId: resp.InfoAtomID,
		ErrorCode:  resp.ErrorCode,
	}, nil
}
