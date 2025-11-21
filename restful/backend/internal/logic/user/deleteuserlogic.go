package user

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DeleteUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteUserLogic) DeleteUser(req *types.DeleteUserRequest) (resp *types.DeleteUserResponse, err error) {
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "delete_user"),
		logx.Field("status", "started"),
		logx.Field("target_user_key", req.Id),
	).Info("开始删除用户")

	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "delete_user"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取操作用户信息失败")

		return &types.DeleteUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
		}, nil
	}

	if jwtUser.UserId == req.Id {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "delete_user"),
			logx.Field("status", "failed"),
			logx.Field("operator_user_key", jwtUser.UserId),
			logx.Field("target_user_key", req.Id),
		).Error("用户尝试删除自己")

		return &types.DeleteUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 403,
				Msg:  "不能删除自己的账户",
			},
		}, nil
	}

	var existingUser struct {
		Name        string `db:"name"`
		Status      string `db:"status"`
		TenantIdStr string `db:"tenant_id_str"`
		TenantKey   string `db:"tenant_key"`
		UserKey     string `db:"user_key"`
	}
	checkUserQuery := `
        SELECT u.name, u.status, COALESCE(u.tenant_id::text, '') as tenant_id_str,
               COALESCE(t.tenant_key, '') as tenant_key, u.user_key
        FROM system_users u
        LEFT JOIN system_tenants t ON u.tenant_id = t.id
        WHERE u.id = $1 AND u.status != 'deleted'
    `
	err = l.svcCtx.DBConn.QueryRow(&existingUser, checkUserQuery, req.Id)
	if err != nil {
		return &types.DeleteUserResponse{
			BaseResponse: types.BaseResponse{Code: 404, Msg: "用户不存在"},
		}, nil
	}

	hasDeletePermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		core.Permission{Resource: core.ResourceUser, Action: core.ActionDelete},
	)
	if err != nil {
		return &types.DeleteUserResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasDeletePermission {
		return &types.DeleteUserResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要用户删除权限"},
		}, nil
	}

	isCrossTenant := existingUser.TenantIdStr != "" && existingUser.TenantIdStr != jwtUser.TenantId
	if isCrossTenant {
		hasCrossTenantPermission, err := l.svcCtx.Casbinx.CheckPermission(
			jwtUser.UserKey,
			jwtUser.TenantKey,
			core.Permission{Resource: "user:cross_tenant", Action: core.ActionDelete},
		)
		if err != nil {
			return &types.DeleteUserResponse{
				BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
			}, nil
		}

		if !hasCrossTenantPermission {
			return &types.DeleteUserResponse{
				BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要跨租户用户删除权限"},
			}, nil
		}
	}

	err = l.svcCtx.DBConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		deleteQuery := `
			UPDATE system_users SET
				status = 'deleted',
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND status != 'deleted'
		`

		result, err := session.ExecCtx(ctx, deleteQuery, req.Id)
		if err != nil {
			return fmt.Errorf("执行用户软删除失败: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("检查删除影响行数失败: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("用户不存在或已被删除")
		}

		return nil
	})

	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "delete_user"),
			logx.Field("status", "failed"),
			logx.Field("target_user_key", req.Id),
			logx.Field("error", err.Error()),
		).Error("用户删除事务失败")

		return &types.DeleteUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  fmt.Sprintf("删除用户失败: %v", err),
			},
		}, nil
	}

	if l.svcCtx.Casbinx != nil {
		err = l.svcCtx.Casbinx.ClearUserPermissions(jwtUser.UserKey, existingUser.UserKey, existingUser.TenantKey)
		if err != nil {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "user"),
				logx.Field("operation", "delete_user"),
				logx.Field("status", "warning"),
				logx.Field("target_user_key", existingUser.UserKey),
				logx.Field("error", err.Error()),
			).Error("清理用户权限数据失败")
		}

		err = l.svcCtx.Casbinx.ClearUserRoles(jwtUser.UserKey, existingUser.UserKey)
		if err != nil {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "user"),
				logx.Field("operation", "delete_user"),
				logx.Field("status", "warning"),
				logx.Field("target_user_key", existingUser.UserKey),
				logx.Field("error", err.Error()),
			).Error("移除用户角色分配失败")
		}
	}

	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "delete_user"),
		logx.Field("status", "success"),
		logx.Field("operator_user_key", jwtUser.UserId),
		logx.Field("target_user_key", req.Id),
		logx.Field("target_user_name", existingUser.Name),
	).Info("删除用户成功")

	return &types.DeleteUserResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "用户删除成功",
		},
	}, nil
}
