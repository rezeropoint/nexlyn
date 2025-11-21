package permission

import (
	"context"
	"fmt"
	"strings"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新角色
func NewUpdateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRoleLogic {
	return &UpdateRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateRoleLogic) UpdateRole(req *types.UpdateRoleRequest) (resp *types.UpdateRoleResponse, err error) {
	// 从JWT中获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		return &types.UpdateRoleResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
		}, nil
	}

	// 转换权限格式从字符串到 Permission 结构体
	var permissions []core.Permission
	for _, permStr := range req.Permissions {
		// 权限字符串格式： "resource:action"
		parts := strings.Split(permStr, ":")
		if len(parts) == 2 {
			action, err := core.ParseAction(strings.TrimSpace(parts[1]))
			if err != nil {
				return &types.UpdateRoleResponse{
					BaseResponse: types.BaseResponse{
						Code: 400,
						Msg:  fmt.Sprintf("无效的权限操作: %s", parts[1]),
					},
				}, nil
			}
			permissions = append(permissions, core.Permission{
				Resource: core.Resource(strings.TrimSpace(parts[0])),
				Action:   action,
			})
		}
	}

	// 更新角色（使用 CasbinX 的 UpdateRole 方法，自动进行安全检查）
	err = l.svcCtx.Casbinx.UpdateRole(jwtUser.UserKey, req.RoleKey, req.RoleName, req.Description, jwtUser.TenantKey, permissions)
	if err != nil {
		return &types.UpdateRoleResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  fmt.Sprintf("更新角色失败: %v", err),
			},
		}, nil
	}

	return &types.UpdateRoleResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "更新角色成功",
		},
	}, nil
}
