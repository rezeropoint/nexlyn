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

type AddRolePermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 为角色添加权限
func NewAddRolePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddRolePermissionLogic {
	return &AddRolePermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddRolePermissionLogic) AddRolePermission(req *types.AddRolePermissionRequest) (resp *types.AddRolePermissionResponse, err error) {
	// 从JWT中获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		return &types.AddRolePermissionResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
		}, nil
	}

	// 为角色添加权限（使用 CasbinX 的 GrantRolePermission 方法，自动进行安全检查）
	// 将字符串Action转换为Action常量
	action, err := core.ParseAction(req.Action)
	if err != nil {
		return &types.AddRolePermissionResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "无效的操作类型",
			},
		}, nil
	}
	resource, err := core.ParseSystemResource(req.Resource)
	if err != nil {
		return &types.AddRolePermissionResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "无效的资源类型",
			},
		}, nil
	}

	err = l.svcCtx.Casbinx.GrantRolePermission(jwtUser.UserKey, req.RoleKey, core.Permission{Resource: resource, Action: action})
	if err != nil {
		return &types.AddRolePermissionResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  fmt.Sprintf("为角色添加权限失败: %v", err),
			},
		}, nil
	}

	return &types.AddRolePermissionResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "为角色添加权限成功",
		},
	}, nil
}
