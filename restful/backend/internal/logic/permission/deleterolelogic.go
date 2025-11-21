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

type DeleteRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRoleLogic {
	return &DeleteRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteRoleLogic) DeleteRole(req *types.DeleteRoleRequest) (resp *types.DeleteRoleResponse, err error) {
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		return &types.DeleteRoleResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
		}, nil
	}

	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: core.ResourceRole, Action: core.ActionWrite})
	if err != nil {
		return &types.DeleteRoleResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "系统权限检查失败",
			},
		}, nil
	}

	if !hasPermission {
		return &types.DeleteRoleResponse{
			BaseResponse: types.BaseResponse{
				Code: 403,
				Msg:  "权限不足，需要角色管理权限",
			},
		}, nil
	}

	usersUsingRole, err := l.svcCtx.Casbinx.GetUsersWithRole(req.RoleKey, jwtUser.TenantKey)
	if err != nil {
		return &types.DeleteRoleResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "检查角色使用情况失败",
			},
		}, nil
	}

	if len(usersUsingRole) > 0 {
		return &types.DeleteRoleResponse{
			BaseResponse: types.BaseResponse{
				Code: 409,
				Msg:  fmt.Sprintf("角色正在被 %d 个用户使用，无法删除。请先移除这些用户的角色分配", len(usersUsingRole)),
			},
		}, nil
	}

	err = l.svcCtx.Casbinx.DeleteRole(req.RoleKey)
	if err != nil {
		return &types.DeleteRoleResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  fmt.Sprintf("删除角色失败: %v", err),
			},
		}, nil
	}

	return &types.DeleteRoleResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "删除角色成功",
		},
	}, nil
}
