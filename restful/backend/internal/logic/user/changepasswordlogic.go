// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type ChangePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 修改当前用户密码
func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangePasswordLogic) ChangePassword(req *types.ChangePasswordRequest) (resp *types.ChangePasswordResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "change_password"),
		logx.Field("status", "started"),
	).Info("开始修改当前用户密码")

	// 获取当前用户信息（从JWT）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "change_password"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取当前用户信息失败")

		return &types.ChangePasswordResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
		}, nil
	}

	// 查询用户的密码哈希
	var user struct {
		PasswordHash string `db:"password_hash"`
		Status       string `db:"status"`
	}

	query := `
		SELECT password_hash, status
		FROM system_users
		WHERE id = $1 AND status = 'active'
	`

	err = l.svcCtx.DBConn.QueryRowCtx(l.ctx, &user, query, jwtUser.UserId)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "change_password"),
			logx.Field("status", "failed"),
			logx.Field("user_id", jwtUser.UserId),
			logx.Field("error", err.Error()),
		).Error("查询用户信息失败")

		return &types.ChangePasswordResponse{
			BaseResponse: types.BaseResponse{
				Code: 404,
				Msg:  "用户不存在",
			},
		}, nil
	}

	// 验证旧密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword))
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "change_password"),
			logx.Field("status", "failed"),
			logx.Field("user_id", jwtUser.UserId),
		).Error("旧密码验证失败")

		return &types.ChangePasswordResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "当前密码错误",
			},
		}, nil
	}

	// 生成新密码哈希
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "change_password"),
			logx.Field("status", "failed"),
			logx.Field("user_id", jwtUser.UserId),
			logx.Field("error", err.Error()),
		).Error("生成密码哈希失败")

		return &types.ChangePasswordResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  config.FormatError(config.ErrMsgUserPasswordHash, err),
			},
		}, nil
	}

	// 更新密码
	updateQuery := `
		UPDATE system_users SET
			password_hash = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status = 'active'
	`

	result, err := l.svcCtx.DBConn.Exec(updateQuery, jwtUser.UserId, string(newPasswordHash))
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "change_password"),
			logx.Field("status", "failed"),
			logx.Field("user_id", jwtUser.UserId),
			logx.Field("error", err.Error()),
		).Error("更新密码失败")

		return &types.ChangePasswordResponse{
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
			logx.Field("operation", "change_password"),
			logx.Field("status", "failed"),
			logx.Field("user_id", jwtUser.UserId),
		).Error("密码修改操作未影响任何记录")

		return &types.ChangePasswordResponse{
			BaseResponse: types.BaseResponse{
				Code: 404,
				Msg:  "用户不存在或已被禁用",
			},
		}, nil
	}

	// 记录成功（不记录密码到日志中，出于安全考虑）
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "change_password"),
		logx.Field("status", "success"),
		logx.Field("user_id", jwtUser.UserId),
	).Info("修改密码成功")

	return &types.ChangePasswordResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "密码修改成功",
		},
	}, nil
}
