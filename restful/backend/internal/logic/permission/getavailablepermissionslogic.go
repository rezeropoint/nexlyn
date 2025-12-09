package permission

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAvailablePermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取可用权限列表
func NewGetAvailablePermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAvailablePermissionsLogic {
	return &GetAvailablePermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAvailablePermissionsLogic) GetAvailablePermissions() (resp *types.GetAvailablePermissionsResponse, err error) {
	// 从JWT中获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		return &types.GetAvailablePermissionsResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
		}, nil
	}

	// 权限验证：检查是否有权限读取权限列表
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: core.ResourcePermission, Action: core.ActionRead})
	if err != nil {
		return &types.GetAvailablePermissionsResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "系统权限检查失败",
			},
		}, nil
	}

	if !hasPermission {
		return &types.GetAvailablePermissionsResponse{
			BaseResponse: types.BaseResponse{
				Code: 403,
				Msg:  "权限不足，需要权限读取权限",
			},
		}, nil
	}

	// 从权限注册中心获取所有权限
	permissionMetadata := auth.GetAllPermissions()

	// 转换为API响应格式
	permissions := make([]types.PermissionItem, len(permissionMetadata))
	for i, pm := range permissionMetadata {
		permissions[i] = types.PermissionItem{
			Resource:     pm.Resource,
			ResourceName: pm.ResourceName,
			Action:       pm.Action,
			Description:  pm.Description,
			Category:     pm.Category,
		}
	}

	return &types.GetAvailablePermissionsResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "获取可用权限列表成功",
		},
		Data: permissions,
	}, nil
}
