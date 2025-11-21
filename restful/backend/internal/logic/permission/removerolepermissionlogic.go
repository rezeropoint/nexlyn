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

type RemoveRolePermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 为角色移除权限
func NewRemoveRolePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveRolePermissionLogic {
	return &RemoveRolePermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveRolePermissionLogic) RemoveRolePermission(req *types.RemoveRolePermissionRequest) (resp *types.RemoveRolePermissionResponse, err error) {
	// 从JWT中获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		return &types.RemoveRolePermissionResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
		}, nil
	}

	// 为角色移除权限（使用 CasbinX 的 RevokeRolePermission 方法，自动进行安全检查）
	// 将字符串Action转换为Action常量
	action, err := core.ParseAction(req.Action)
	if err != nil {
		return &types.RemoveRolePermissionResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "无效的操作类型",
			},
		}, nil
	}

	resource, err := core.ParseSystemResource(req.Resource)
	if err != nil {
		return &types.RemoveRolePermissionResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "无效的资源类型",
			},
		}, nil
	}

	err = l.svcCtx.Casbinx.RevokeRolePermission(jwtUser.UserKey, req.RoleKey, core.Permission{Resource: resource, Action: action})
	if err != nil {
		return &types.RemoveRolePermissionResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  fmt.Sprintf("为角色移除权限失败: %v", err),
			},
		}, nil
	}

	return &types.RemoveRolePermissionResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "为角色移除权限成功",
		},
	}, nil
}
