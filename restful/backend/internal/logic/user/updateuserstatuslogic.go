package user

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新用户状态
func NewUpdateUserStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserStatusLogic {
	return &UpdateUserStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserStatusLogic) UpdateUserStatus(req *types.UpdateUserStatusRequest) (resp *types.UpdateUserStatusResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "update_user_status"),
		logx.Field("status", "started"),
		logx.Field("target_user_id", req.Id),
	).Info("开始更新用户状态")

	// 获取当前操作用户信息（从JWT）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "update_user_status"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取操作用户信息失败")

		return &types.UpdateUserStatusResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
		}, nil
	}

	// 防止用户修改自己的状态
	if jwtUser.UserId == req.Id {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "update_user_status"),
			logx.Field("status", "failed"),
			logx.Field("operator_user_id", jwtUser.UserId),
			logx.Field("target_user_id", req.Id),
		).Error("用户尝试修改自己的状态")

		return &types.UpdateUserStatusResponse{
			BaseResponse: types.BaseResponse{
				Code: 403,
				Msg:  "不能修改自己的账户状态",
			},
		}, nil
	}

	// 检查目标用户是否存在
	var targetUser struct {
		Name        string `db:"name"`
		Email       string `db:"email"`
		Status      string `db:"status"`
		TenantIdStr string `db:"tenant_id_str"`
	}
	checkUserQuery := `
        SELECT u.name, u.email, u.status,
               COALESCE(u.tenant_id::text, '') as tenant_id_str
        FROM system_users u
        LEFT JOIN system_tenants t ON u.tenant_id = t.id
        WHERE u.id = $1 AND u.status != 'deleted'`
	if err := l.svcCtx.DBConn.QueryRow(&targetUser, checkUserQuery, req.Id); err != nil {
		return &types.UpdateUserStatusResponse{
			BaseResponse: types.BaseResponse{Code: 404, Msg: "用户不存在"},
		}, nil
	}

	// 权限验证：检查用户更新权限
	hasWritePermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		core.Permission{Resource: core.ResourceUser, Action: core.ActionWrite},
	)
	if err != nil {
		return &types.UpdateUserStatusResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasWritePermission {
		return &types.UpdateUserStatusResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要用户更新权限"},
		}, nil
	}

	// 权限传递深度检查：防止被授权者更改授权者的状态
	// 获取目标用户的user_key
	var targetUserKey string
	err = l.svcCtx.DBConn.QueryRow(&targetUserKey, `SELECT user_key FROM system_users WHERE id = $1`, req.Id)
	if err != nil {
		return &types.UpdateUserStatusResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "获取用户信息失败"},
		}, nil
	}

	// 检查跨租户权限
	isCrossTenant := targetUser.TenantIdStr != "" && targetUser.TenantIdStr != jwtUser.TenantId
	if isCrossTenant {
		hasCrossTenantPermission, err := l.svcCtx.Casbinx.CheckPermission(
			jwtUser.UserKey,
			jwtUser.TenantKey,
			core.Permission{Resource: core.ResourceTenant, Action: core.ActionWrite},
		)
		if err != nil {
			return &types.UpdateUserStatusResponse{
				BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
			}, nil
		}

		if !hasCrossTenantPermission {
			return &types.UpdateUserStatusResponse{
				BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要跨租户用户更新权限"},
			}, nil
		}
	}

	// 查询用户失败的错误处理
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "update_user_status"),
			logx.Field("status", "failed"),
			logx.Field("target_user_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("查询目标用户信息失败")

		return &types.UpdateUserStatusResponse{
			BaseResponse: types.BaseResponse{
				Code: 404,
				Msg:  "用户不存在",
			},
		}, nil
	}

	// 确定新状态
	newStatus := targetUser.Status
	statusAction := "no_change"
	if req.Status != "" {
		// 验证状态值的有效性
		validStatuses := map[string]bool{
			"active":   true,
			"inactive": true,
			"deleted":  true,
		}

		if !validStatuses[req.Status] {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "user"),
				logx.Field("operation", "update_user_status"),
				logx.Field("status", "failed"),
				logx.Field("target_user_id", req.Id),
				logx.Field("invalid_status", req.Status),
			).Error("无效的状态值")

			return &types.UpdateUserStatusResponse{
				BaseResponse: types.BaseResponse{
					Code: 400,
					Msg:  "无效的状态值",
				},
			}, nil
		}

		if req.Status != targetUser.Status {
			newStatus = req.Status
			if req.Status == "active" {
				statusAction = "activate"
			} else if req.Status == "inactive" {
				statusAction = "disable"
			} else if req.Status == "deleted" {
				statusAction = "delete"
			}
		}
	}

	// 更新用户状态
	updateQuery := `
        UPDATE system_users SET 
            status = $2,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1 AND status != 'deleted'
	`

	result, err := l.svcCtx.DBConn.Exec(updateQuery, req.Id, newStatus)

	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "update_user_status"),
			logx.Field("status", "failed"),
			logx.Field("target_user_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("更新用户状态失败")

		return &types.UpdateUserStatusResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "更新用户状态失败",
			},
		}, nil
	}

	// 检查是否实际更新了记录
	if result == nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "update_user_status"),
			logx.Field("status", "failed"),
			logx.Field("target_user_id", req.Id),
		).Error("状态更新操作未影响任何记录")

		return &types.UpdateUserStatusResponse{
			BaseResponse: types.BaseResponse{
				Code: 404,
				Msg:  "用户不存在或已被删除",
			},
		}, nil
	}

	// 确定响应消息
	var successMsg string
	switch statusAction {
	case "activate":
		successMsg = "用户账户已激活"
	case "disable":
		successMsg = "用户账户已禁用"
	default:
		successMsg = "用户状态更新成功"
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "update_user_status"),
		logx.Field("status", "success"),
		logx.Field("operator_user_id", jwtUser.UserId),
		logx.Field("target_user_id", req.Id),
		logx.Field("target_user_name", targetUser.Name),
		logx.Field("target_user_email", targetUser.Email),
		logx.Field("old_status", targetUser.Status),
		logx.Field("new_status", newStatus),
		logx.Field("status_action", statusAction),
	).Info("更新用户状态成功")

	return &types.UpdateUserStatusResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  successMsg,
		},
	}, nil
}
