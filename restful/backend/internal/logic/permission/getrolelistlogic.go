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

type GetRoleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取角色列表
func NewGetRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoleListLogic {
	return &GetRoleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRoleListLogic) GetRoleList(req *types.GetRoleListRequest) (resp *types.GetRoleListResponse, err error) {
	// 从JWT中获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		return &types.GetRoleListResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
		}, nil
	}

	// 权限验证：检查是否有角色管理权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: core.ResourceRole, Action: core.ActionWrite})
	if err != nil {
		return &types.GetRoleListResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "系统权限检查失败",
			},
		}, nil
	}

	if !hasPermission {
		return &types.GetRoleListResponse{
			BaseResponse: types.BaseResponse{
				Code: 403,
				Msg:  "权限不足，需要角色管理权限",
			},
		}, nil
	}

	// 构建过滤器
	filter := &core.RoleFilter{}
	if req.RoleKey != "" {
		filter.KeyPattern = req.RoleKey
	}
	if req.RoleName != "" {
		filter.NamePattern = req.RoleName
	}

	// 获取角色列表
	roles, err := l.svcCtx.Casbinx.ListRoles(jwtUser.TenantKey, filter)
	if err != nil {
		return &types.GetRoleListResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  fmt.Sprintf("获取角色列表失败: %v", err),
			},
		}, nil
	}

	// 转换为API响应格式
	var roleList []types.RoleInfo
	for _, role := range roles {
		// 转换权限格式为字符串数组
		permissions := make([]string, len(role.Permissions))
		for i, perm := range role.Permissions {
			permissions[i] = fmt.Sprintf("%s:%s", perm.Resource, perm.Action)
		}

		roleList = append(roleList, types.RoleInfo{
			RoleKey:     role.Key,
			RoleName:    role.Name,
			Description: role.Description,
			Permissions: permissions,
		})
	}

	return &types.GetRoleListResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "获取角色列表成功",
		},
		Data: roleList,
		PageParams: types.PageParams{
			Current:  req.Current,
			PageSize: req.PageSize,
			Total:    int64(len(roleList)),
		},
	}, nil
}
