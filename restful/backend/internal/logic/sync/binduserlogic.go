// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package sync

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 绑定用户到已存在的Skylark用户
func NewBindUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindUserLogic {
	return &BindUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BindUserLogic) BindUser(req *types.BindUserRequest) (resp *types.BindUserResponse, err error) {
	// 1. 检查是否启用Skylark同步
	if !l.svcCtx.Config.SkylarkSyncEnabled || l.svcCtx.AdminSyncClient == nil {
		return &types.BindUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "Skylark同步功能未启用",
			},
		}, nil
	}

	// 2. 参数验证
	if req.RemoteUserId <= 0 {
		return &types.BindUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "远程用户ID必须大于0",
			},
		}, nil
	}

	// 3. 从 JWT 获取当前用户信息（租户ID）
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		return &types.BindUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "获取当前用户信息失败",
			},
		}, nil
	}

	// 4. 验证用户是否存在且有权限访问
	query := `SELECT id FROM system_users WHERE id = $1 AND tenant_id = $2 AND status != 'deleted'`
	var userId string
	err = l.svcCtx.DBConn.QueryRowCtx(l.ctx, &userId, query, req.Id, currentUser.TenantId)
	if err != nil {
		if err == sql.ErrNoRows {
			return &types.BindUserResponse{
				BaseResponse: types.BaseResponse{
					Code: 404,
					Msg:  "用户不存在或无权限",
				},
			}, nil
		}
		logx.Error("查询用户信息失败: ", err)
		return &types.BindUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "查询用户信息失败",
			},
		}, nil
	}

	// 5. 调用 eventsync gRPC 绑定用户
	bindResp, err := l.svcCtx.AdminSyncClient.BindUser(l.ctx, &pb.BindUserReq{
		TenantId:     currentUser.TenantId,
		LocalUserId:  req.Id,
		RemoteUserId: req.RemoteUserId,
	})

	if err != nil {
		logx.Error("调用eventsync绑定用户失败: ", err)
		return &types.BindUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  fmt.Sprintf("绑定失败: %v", err),
			},
		}, nil
	}

	if !bindResp.Success {
		// 根据错误码返回不同的HTTP状态码
		httpCode := 500
		if bindResp.ErrorCode == "NOT_FOUND" {
			httpCode = 404
		} else if bindResp.ErrorCode == "MAPPING_EXISTS" {
			httpCode = 409
		} else if bindResp.ErrorCode == "INVALID_PARAMS" {
			httpCode = 400
		}

		return &types.BindUserResponse{
			BaseResponse: types.BaseResponse{
				Code: int64(httpCode),
				Msg:  bindResp.Message,
			},
		}, nil
	}

	// 6. 返回成功响应
	return &types.BindUserResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "用户绑定成功",
		},
	}, nil
}
