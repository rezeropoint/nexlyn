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

type SyncOrganizationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 同步单个组织到Skylark
func NewSyncOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncOrganizationLogic {
	return &SyncOrganizationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncOrganizationLogic) SyncOrganization(req *types.SyncOrganizationRequest) (resp *types.SyncOrganizationResponse, err error) {
	// 1. 检查是否启用Skylark同步
	if !l.svcCtx.Config.SkylarkSyncEnabled || l.svcCtx.AdminSyncClient == nil {
		return &types.SyncOrganizationResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "Skylark同步功能未启用",
			},
		}, nil
	}

	// 2. 从 JWT 获取当前用户信息（租户ID）
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		return &types.SyncOrganizationResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "获取当前用户信息失败",
			},
		}, nil
	}

	// 3. 从数据库查询组织信息
	query := `SELECT parent_id, name, description, created_by FROM system_organizations WHERE id = $1 AND tenant_id = $2 AND status != 'deleted'`
	var orgInfo struct {
		ParentId    sql.NullString `db:"parent_id"`
		Name        string         `db:"name"`
		Description sql.NullString `db:"description"`
		CreatedBy   sql.NullString `db:"created_by"` // UUID类型，在Go中使用string接收
	}
	err = l.svcCtx.DBConn.QueryRowCtx(l.ctx, &orgInfo, query, req.Id, currentUser.TenantId)
	if err != nil {
		logx.Error("查询组织信息失败: ", err)
		if err == sql.ErrNoRows {
			return &types.SyncOrganizationResponse{
				BaseResponse: types.BaseResponse{
					Code: 404,
					Msg:  "组织不存在或无权限",
				},
			}, nil
		}
		return &types.SyncOrganizationResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "查询组织信息失败",
			},
		}, nil
	}

	// 4. 处理可选字段
	parentIdStr := ""
	if orgInfo.ParentId.Valid {
		parentIdStr = orgInfo.ParentId.String
	}
	descriptionStr := ""
	if orgInfo.Description.Valid {
		descriptionStr = orgInfo.Description.String
	}
	// 处理创建者ID：如果为空（NULL），使用当前操作用户ID
	createdByStr := currentUser.UserId // 默认使用当前用户
	if orgInfo.CreatedBy.Valid && orgInfo.CreatedBy.String != "" {
		createdByStr = orgInfo.CreatedBy.String
	}

	// 5. 调用 eventsync gRPC 同步组织
	syncResp, err := l.svcCtx.AdminSyncClient.SyncOrganization(l.ctx, &pb.SyncOrganizationReq{
		TenantId:         currentUser.TenantId,
		LocalOrgId:       req.Id,
		ParentLocalOrgId: parentIdStr,
		Name:             orgInfo.Name,
		Description:      descriptionStr,
		FounderId:        createdByStr,
	})

	if err != nil {
		logx.Error("调用eventsync同步组织失败: ", err)
		return &types.SyncOrganizationResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  fmt.Sprintf("同步失败: %v", err),
			},
		}, nil
	}

	if !syncResp.Success {
		return &types.SyncOrganizationResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  fmt.Sprintf("同步失败: %s", syncResp.Message),
			},
		}, nil
	}

	// 6. 返回成功响应
	return &types.SyncOrganizationResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "同步成功",
		},
		Data: struct {
			RemoteOrgId string `json:"remoteOrgId"`
		}{
			RemoteOrgId: syncResp.RemoteOrgId,
		},
	}, nil
}
