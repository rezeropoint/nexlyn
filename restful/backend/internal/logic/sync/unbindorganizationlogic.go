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

type UnbindOrganizationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 解绑组织的Skylark映射
func NewUnbindOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnbindOrganizationLogic {
	return &UnbindOrganizationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnbindOrganizationLogic) UnbindOrganization(req *types.UnbindOrganizationRequest) (resp *types.UnbindOrganizationResponse, err error) {
	// 1. 检查是否启用Skylark同步
	if !l.svcCtx.Config.SkylarkSyncEnabled || l.svcCtx.AdminSyncClient == nil {
		return &types.UnbindOrganizationResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "Skylark同步功能未启用",
			},
		}, nil
	}

	// 2. 从 JWT 获取当前用户信息（租户ID）
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		return &types.UnbindOrganizationResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "获取当前用户信息失败",
			},
		}, nil
	}

	// 3. 验证组织是否存在且有权限访问
	query := `SELECT id FROM system_organizations WHERE id = $1 AND tenant_id = $2 AND status != 'deleted'`
	var orgId string
	err = l.svcCtx.DBConn.QueryRowCtx(l.ctx, &orgId, query, req.Id, currentUser.TenantId)
	if err != nil {
		if err == sql.ErrNoRows {
			return &types.UnbindOrganizationResponse{
				BaseResponse: types.BaseResponse{
					Code: 404,
					Msg:  "组织不存在或无权限",
				},
			}, nil
		}
		logx.Error("查询组织信息失败: ", err)
		return &types.UnbindOrganizationResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "查询组织信息失败",
			},
		}, nil
	}

	// 4. 调用 eventsync gRPC 解绑组织
	unbindResp, err := l.svcCtx.AdminSyncClient.UnbindOrganization(l.ctx, &pb.UnbindOrganizationReq{
		TenantId:   currentUser.TenantId,
		LocalOrgId: req.Id,
	})

	if err != nil {
		logx.Error("调用eventsync解绑组织失败: ", err)
		return &types.UnbindOrganizationResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  fmt.Sprintf("解绑失败: %v", err),
			},
		}, nil
	}

	if !unbindResp.Success {
		// 根据错误码返回不同的HTTP状态码
		httpCode := 500
		if unbindResp.ErrorCode == "MAPPING_NOT_FOUND" {
			httpCode = 404
		} else if unbindResp.ErrorCode == "INVALID_PARAMS" {
			httpCode = 400
		}

		return &types.UnbindOrganizationResponse{
			BaseResponse: types.BaseResponse{
				Code: int64(httpCode),
				Msg:  unbindResp.Message,
			},
		}, nil
	}

	// 5. 返回成功响应
	return &types.UnbindOrganizationResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "组织解绑成功",
		},
	}, nil
}
