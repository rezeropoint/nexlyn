package user

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type ResetUserPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 重置用户密码
func NewResetUserPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetUserPasswordLogic {
	return &ResetUserPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetUserPasswordLogic) ResetUserPassword(req *types.ResetPasswordRequest) (resp *types.ResetPasswordResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "reset_user_password"),
		logx.Field("status", "started"),
		logx.Field("target_user_id", req.Id),
	).Info("开始重置用户密码")

	// 获取当前操作用户信息（从JWT）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "reset_user_password"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取操作用户信息失败")

		return &types.ResetPasswordResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
		}, nil
	}

	// 检查目标用户是否存在（含租户和角色）
	var targetUser struct {
		Name        string `db:"name"`
		Email       string `db:"email"`
		Status      string `db:"status"`
		TenantIdStr string `db:"tenant_id_str"`
	}
	checkUserQuery := `
        SELECT u.name, u.email, u.status, COALESCE(u.tenant_id::text, '') as tenant_id_str
        FROM system_users u
        LEFT JOIN system_tenants t ON u.tenant_id = t.id
        WHERE u.id = $1 AND u.status = 'active'
    `
	if err := l.svcCtx.DBConn.QueryRow(&targetUser, checkUserQuery, req.Id); err != nil {
		return &types.ResetPasswordResponse{
			BaseResponse: types.BaseResponse{Code: 404, Msg: "用户不存在"},
		}, nil
	}

	// 权限验证：重置密码需要 user:write 权限（仅限管理员）
	// 注意：用户修改自己的密码应使用 changePassword 接口（需验证旧密码）
	hasWritePermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: core.ResourceUser, Action: core.ActionWrite})
	if err != nil {
		return &types.ResetPasswordResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}

	if !hasWritePermission {
		return &types.ResetPasswordResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: config.FormatError(config.ErrMsgPermissionDenied, err)},
		}, nil
	}

	// 权限传递深度检查：防止被授权者重置授权者的密码
	// 获取目标用户的user_key
	var targetUserKey string
	err = l.svcCtx.DBConn.QueryRow(&targetUserKey, `SELECT user_key FROM system_users WHERE id = $1`, req.Id)
	if err != nil {
		return &types.ResetPasswordResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgUserQuery, err)},
		}, nil
	}

	// 检查跨租户权限
	isCrossTenant := targetUser.TenantIdStr != "" && targetUser.TenantIdStr != jwtUser.TenantId
	if isCrossTenant {
		hasCrossTenantPermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: core.ResourceTenant, Action: core.ActionWrite})
		if err != nil {
			return &types.ResetPasswordResponse{
				BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
			}, nil
		}

		if !hasCrossTenantPermission {
			return &types.ResetPasswordResponse{
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
			logx.Field("operation", "reset_user_password"),
			logx.Field("status", "failed"),
			logx.Field("target_user_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("查询目标用户信息失败")

		return &types.ResetPasswordResponse{
			BaseResponse: types.BaseResponse{
				Code: 404,
				Msg:  "用户不存在或已被禁用",
			},
		}, nil
	}

	// 哈希新密码
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "reset_user_password"),
			logx.Field("status", "failed"),
			logx.Field("target_user_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("生成密码哈希失败")

		return &types.ResetPasswordResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  config.FormatError(config.ErrMsgUserPasswordHash, err),
			},
		}, nil
	}

	// 更新用户密码
	updatePasswordQuery := `
		UPDATE system_users SET 
			password_hash = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status = 'active'
	`

	result, err := l.svcCtx.DBConn.Exec(updatePasswordQuery, req.Id, string(passwordHash))
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "reset_user_password"),
			logx.Field("status", "failed"),
			logx.Field("target_user_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("更新用户密码失败")

		return &types.ResetPasswordResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  config.FormatError(config.ErrMsgUserUpdate, err),
			},
		}, nil
	}

	// 检查是否实际更新了密码
	if result == nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "reset_user_password"),
			logx.Field("status", "failed"),
			logx.Field("target_user_id", req.Id),
		).Error("密码重置操作未影响任何记录")

		return &types.ResetPasswordResponse{
			BaseResponse: types.BaseResponse{
				Code: 404,
				Msg:  "用户不存在或已被禁用",
			},
		}, nil
	}

	// 记录成功（不记录新密码到日志中，出于安全考虑）
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "reset_user_password"),
		logx.Field("status", "success"),
		logx.Field("operator_user_id", jwtUser.UserId),
		logx.Field("target_user_id", req.Id),
		logx.Field("target_user_name", targetUser.Name),
		logx.Field("target_user_email", targetUser.Email),
	).Info("重置用户密码成功")

	return &types.ResetPasswordResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "密码重置成功",
		},
	}, nil
}
