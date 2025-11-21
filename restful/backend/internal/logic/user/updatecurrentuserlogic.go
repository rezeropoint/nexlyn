// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCurrentUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新当前用户信息(用户修改自己的资料,无需user:write权限)
func NewUpdateCurrentUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCurrentUserLogic {
	return &UpdateCurrentUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateCurrentUserLogic) UpdateCurrentUser(req *types.UpdateCurrentUserRequest) (resp *types.UpdateCurrentUserResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "update_current_user"),
		logx.Field("status", "started"),
	).Info("用户开始更新自己的信息")

	// 从JWT获取当前用户ID（不从请求参数获取，避免越权）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		return &types.UpdateCurrentUserResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	currentUserId := jwtUser.UserId

	// 检查用户是否存在
	var existingUser struct {
		UserKey string `db:"user_key"`
		Status  string `db:"status"`
	}
	checkUserQuery := `SELECT user_key, status FROM system_users WHERE id = $1`
	err = l.svcCtx.DBConn.QueryRow(&existingUser, checkUserQuery, currentUserId)
	if err != nil {
		return &types.UpdateCurrentUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 404,
				Msg:  "用户不存在",
			},
		}, nil
	}

	// 检查用户状态
	if existingUser.Status == "deleted" {
		return &types.UpdateCurrentUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 403,
				Msg:  "用户已被删除，无法修改信息",
			},
		}, nil
	}

	if existingUser.Status == "inactive" {
		return &types.UpdateCurrentUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 403,
				Msg:  "用户已被禁用，无法修改信息",
			},
		}, nil
	}

	// 构建更新SQL（只允许修改个人信息字段）
	// 注意：不允许修改user_name、tenant_id、status等敏感字段
	updateQuery := `
		UPDATE system_users SET
			name = COALESCE(NULLIF($2, ''), name),
			email = COALESCE(NULLIF($3, ''), email),
			phone = COALESCE(NULLIF($4, ''), phone),
			title = COALESCE(NULLIF($5, ''), title),
			avatar = COALESCE(NULLIF($6, ''), avatar),
			signature = COALESCE(NULLIF($7, ''), signature),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status = 'active'
	`

	// 执行更新
	result, err := l.svcCtx.DBConn.ExecCtx(l.ctx, updateQuery, currentUserId,
		req.Name, req.Email, req.Phone, req.Title,
		req.Avatar, req.Signature)

	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "update_current_user"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("更新用户信息失败")

		return &types.UpdateCurrentUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  fmt.Sprintf("更新用户信息失败: %v", err),
			},
		}, nil
	}

	// 检查是否有行被更新
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "update_current_user"),
			logx.Field("status", "no_rows_affected"),
		).Error("没有行被更新")

		return &types.UpdateCurrentUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "更新失败：没有找到可更新的记录",
			},
		}, nil
	}

	// 记录操作成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "update_current_user"),
		logx.Field("status", "success"),
		logx.Field("user_id", currentUserId),
	).Info("用户信息更新成功")

	return &types.UpdateCurrentUserResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "更新成功",
		},
	}, nil
}
