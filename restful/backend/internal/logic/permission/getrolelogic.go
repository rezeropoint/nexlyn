package permission

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取角色详情
func NewGetRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoleLogic {
	return &GetRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRoleLogic) GetRole(req *types.GetRoleRequest) (resp *types.GetRoleResponse, err error) {
	// 从JWT中获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		return &types.GetRoleResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
		}, nil
	}

	// 权限验证：检查是否有权限读取权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: core.ResourcePermission, Action: core.ActionRead})
	if err != nil {
		return &types.GetRoleResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "系统权限检查失败",
			},
		}, nil
	}

	if !hasPermission {
		return &types.GetRoleResponse{
			BaseResponse: types.BaseResponse{
				Code: 403,
				Msg:  "权限不足，需要权限读取权限",
			},
		}, nil
	}

	// 获取角色详情
	roleInfo, err := l.svcCtx.Casbinx.GetRole(req.RoleKey)
	if err != nil {
		return &types.GetRoleResponse{
			BaseResponse: types.BaseResponse{
				Code: 404,
				Msg:  fmt.Sprintf("角色不存在: %v", err),
			},
		}, nil
	}

	// 转换权限格式为字符串数组
	permissions := make([]string, len(roleInfo.Permissions))
	for i, perm := range roleInfo.Permissions {
		permissions[i] = fmt.Sprintf("%s:%s", perm.Resource, perm.Action)
	}

	return &types.GetRoleResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "获取角色详情成功",
		},
		Data: types.RoleInfo{
			RoleKey:     roleInfo.Key,
			RoleName:    roleInfo.Name,
			Description: roleInfo.Description,
			Permissions: permissions,
		},
	}, nil
}
