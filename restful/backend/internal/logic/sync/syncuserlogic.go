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

type SyncUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 同步单个用户到Skylark
func NewSyncUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncUserLogic {
	return &SyncUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncUserLogic) SyncUser(req *types.SyncUserRequest) (resp *types.SyncUserResponse, err error) {
	// 1. 检查是否启用Skylark同步
	if !l.svcCtx.Config.SkylarkSyncEnabled || l.svcCtx.AdminSyncClient == nil {
		return &types.SyncUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "Skylark同步功能未启用",
			},
		}, nil
	}

	// 2. 从 JWT 获取当前用户信息（租户ID）
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		return &types.SyncUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "获取当前用户信息失败",
			},
		}, nil
	}

	// 3. 从数据库查询用户信息
	query := `SELECT user_key, name, email, phone FROM system_users WHERE id = $1 AND tenant_id = $2 AND status != 'deleted'`
	var userInfo struct {
		UserKey string         `db:"user_key"`
		Name    string         `db:"name"`
		Email   sql.NullString `db:"email"`
		Phone   sql.NullString `db:"phone"`
	}
	err = l.svcCtx.DBConn.QueryRowCtx(l.ctx, &userInfo, query, req.Id, currentUser.TenantId)
	if err != nil {
		logx.Error("查询用户信息失败: ", err)
		if err == sql.ErrNoRows {
			return &types.SyncUserResponse{
				BaseResponse: types.BaseResponse{
					Code: 404,
					Msg:  "用户不存在或无权限",
				},
			}, nil
		}
		return &types.SyncUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "查询用户信息失败",
			},
		}, nil
	}

	// 4. 处理可选字段
	email := ""
	if userInfo.Email.Valid {
		email = userInfo.Email.String
	}
	phone := ""
	if userInfo.Phone.Valid {
		phone = userInfo.Phone.String
	}

	// 5. 调用 eventsync gRPC 同步用户
	syncResp, err := l.svcCtx.AdminSyncClient.SyncUser(l.ctx, &pb.SyncUserReq{
		TenantId:    currentUser.TenantId,
		LocalUserId: req.Id,
		Name:        userInfo.Name,
		Identifier:  email,
		Phone:       phone,
		Openid:      "",
	})

	if err != nil {
		logx.Error("调用eventsync同步用户失败: ", err)
		return &types.SyncUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  fmt.Sprintf("同步失败: %v", err),
			},
		}, nil
	}

	if !syncResp.Success {
		return &types.SyncUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  fmt.Sprintf("同步失败: %s", syncResp.Message),
			},
		}, nil
	}

	// 6. 返回成功响应
	return &types.SyncUserResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "同步成功",
		},
		Data: struct {
			RemoteUserId string `json:"remoteUserId"`
		}{
			RemoteUserId: syncResp.RemoteUserId,
		},
	}, nil
}
