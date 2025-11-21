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

type CreateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建角色
func NewCreateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRoleLogic {
	return &CreateRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateRoleLogic) CreateRole(req *types.CreateRoleRequest) (resp *types.CreateRoleResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "permission"),
		logx.Field("operation", "create_role"),
		logx.Field("status", "started"),
		logx.Field("role_key", req.RoleKey),
		logx.Field("role_name", req.RoleName),
	).Info("开始创建角色")

	// 从JWT中获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "permission"),
			logx.Field("operation", "create_role"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从JWT获取用户信息失败")

		return &types.CreateRoleResponse{
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
				return &types.CreateRoleResponse{
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

	// 创建角色（使用 CasbinX 的 CreateRole 方法，自动进行安全检查）
	err = l.svcCtx.Casbinx.CreateRole(jwtUser.UserKey, req.RoleKey, req.RoleName, req.Description, jwtUser.TenantKey, permissions)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "permission"),
			logx.Field("operation", "create_role"),
			logx.Field("status", "failed"),
			logx.Field("role_key", req.RoleKey),
			logx.Field("error", err.Error()),
		).Error("创建角色失败")

		return &types.CreateRoleResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  fmt.Sprintf("创建角色失败: %v", err),
			},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "permission"),
		logx.Field("operation", "create_role"),
		logx.Field("status", "success"),
		logx.Field("role_key", req.RoleKey),
		logx.Field("role_name", req.RoleName),
	).Info("创建角色成功")

	return &types.CreateRoleResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "创建角色成功",
		},
	}, nil
}
