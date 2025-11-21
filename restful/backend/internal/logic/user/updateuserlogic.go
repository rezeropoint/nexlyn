package user

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新用户信息
func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserLogic) UpdateUser(req *types.UpdateUserRequest) (resp *types.UpdateUserResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "update_user"),
		logx.Field("status", "started"),
		logx.Field("target_user_id", req.Id),
	).Info("开始更新用户信息")

	// 获取当前操作者（用于租户隔离）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		return &types.UpdateUserResponse{BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"}}, nil
	}

	// 检查用户是否存在
	var existingUser struct {
		UserKey string `db:"user_key"`
	}
	checkUserQuery := `SELECT user_key FROM system_users WHERE id = $1 AND status != 'deleted'`
	err = l.svcCtx.DBConn.QueryRow(&existingUser, checkUserQuery, req.Id)
	if err != nil {
		return &types.UpdateUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 404,
				Msg:  "用户不存在",
			},
		}, nil
	}

	// 构建更新SQL
	updateQuery := `
		UPDATE system_users SET 
			name = COALESCE(NULLIF($2, ''), name),
			email = COALESCE(NULLIF($3, ''), email),
			user_name = COALESCE(NULLIF($4, ''), user_name),
			phone = COALESCE(NULLIF($5, ''), phone),
			title = COALESCE(NULLIF($6, ''), title),
			avatar = COALESCE(NULLIF($7, ''), avatar),
			signature = COALESCE(NULLIF($8, ''), signature),
			country = COALESCE(NULLIF($9, ''), country),
			address = COALESCE(NULLIF($10, ''), address),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status != 'deleted'
	`

	// 权限验证：检查用户更新权限（仅限管理员使用，需要user:write权限）
	// 注意：用户修改自己的信息应使用 updateCurrentUser 接口（无需权限验证）
	hasWritePermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		core.Permission{Resource: core.ResourceUser, Action: core.ActionWrite},
	)
	if err != nil {
		return &types.UpdateUserResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasWritePermission {
		return &types.UpdateUserResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要用户更新权限"},
		}, nil
	}

	// 读取目标用户租户
	var targetTenantId string
	_ = l.svcCtx.DBConn.QueryRow(&targetTenantId, `SELECT COALESCE(t.tenant_key, '') FROM system_users u LEFT JOIN system_tenants t ON u.tenant_id = t.id WHERE u.id = $1`, req.Id)

	// 检查跨租户权限
	isCrossTenant := targetTenantId != "" && targetTenantId != jwtUser.TenantKey
	if isCrossTenant {
		hasCrossTenantPermission, err := l.svcCtx.Casbinx.CheckPermission(
			jwtUser.UserKey,
			jwtUser.TenantKey,
			core.Permission{Resource: core.ResourceTenant, Action: core.ActionWrite},
		)
		if err != nil {
			return &types.UpdateUserResponse{
				BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
			}, nil
		}

		if !hasCrossTenantPermission {
			return &types.UpdateUserResponse{
				BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要跨租户用户更新权限"},
			}, nil
		}
	}

	// 使用事务处理整个用户更新过程（用户更新 + 标签更新 + 地理信息更新）
	err = l.svcCtx.DBConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 更新用户基本信息
		_, err := session.ExecCtx(ctx, updateQuery, req.Id,
			req.Name, req.Email, req.UserName, req.Phone, req.Title,
			req.Avatar, req.Signature, req.Country, req.Address)

		if err != nil {
			return fmt.Errorf("更新用户基本信息失败: %w", err)
		}

		// 获取用户内部ID用于更新标签和地理信息
		userInternalID := req.Id
		if userInternalID != "" {
			// 覆盖式更新用户标签（使用标签ID）
			if req.TagIds != nil {
				if len(req.TagIds) > 0 {
					// 校验标签ID均为 user 作用域
					validateQuery := `SELECT COUNT(*) FROM system_tag_definitions WHERE id = ANY($1::uuid[]) AND scope = 'user'`
					var validCount int
					err = session.QueryRowCtx(ctx, &validCount, validateQuery, pq.Array(req.TagIds))
					if err != nil {
						return fmt.Errorf("校验用户标签失败: %w", err)
					}
					if validCount != len(req.TagIds) {
						return fmt.Errorf("部分标签不存在或不属于 user 作用域")
					}

					// 所有标签都有效，删除现有关联
					_, err = session.ExecCtx(ctx, `DELETE FROM user_tag_relations WHERE user_id = $1`, userInternalID)
					if err != nil {
						return fmt.Errorf("删除现有标签关联失败: %w", err)
					}

					// 添加新的标签关联
					for _, tagID := range req.TagIds {
						_, err = session.ExecCtx(ctx,
							`INSERT INTO user_tag_relations (user_id, tag_id) VALUES ($1, $2) ON CONFLICT (user_id, tag_id) DO NOTHING`,
							userInternalID, tagID,
						)
						if err != nil {
							return fmt.Errorf("插入标签关联 %s 失败: %w", tagID, err)
						}
					}
				} else {
					// TagIds 为空数组，删除所有现有关联
					_, err = session.ExecCtx(ctx, `DELETE FROM user_tag_relations WHERE user_id = $1`, userInternalID)
					if err != nil {
						return fmt.Errorf("清空标签关联失败: %w", err)
					}
				}
			}

			// 更新地理信息
			if req.Geographic.Province.Key != "" || req.Geographic.City.Key != "" {
				// 尝试更新，如果不存在则插入
				updateGeoQuery := `
					UPDATE user_geographic SET 
						province_key = $2, province_label = $3,
						city_key = $4, city_label = $5,
						updated_at = CURRENT_TIMESTAMP
					WHERE user_id = $1
				`
				result, err := session.ExecCtx(ctx, updateGeoQuery, userInternalID,
					req.Geographic.Province.Key, req.Geographic.Province.Label,
					req.Geographic.City.Key, req.Geographic.City.Label)

				// 如果更新失败或没有影响行数，则插入新记录
				var rowsAffected int64 = 0
				if err == nil && result != nil {
					rowsAffected, _ = result.RowsAffected()
				}

				if err != nil || rowsAffected == 0 {
					insertGeoQuery := `
						INSERT INTO user_geographic (
							user_id, province_key, province_label, city_key, city_label
						) VALUES ($1, $2, $3, $4, $5)
					`
					_, err = session.ExecCtx(ctx, insertGeoQuery, userInternalID,
						req.Geographic.Province.Key, req.Geographic.Province.Label,
						req.Geographic.City.Key, req.Geographic.City.Label)
					if err != nil {
						return fmt.Errorf("插入用户地理信息失败: %w", err)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "update_user"),
			logx.Field("status", "failed"),
			logx.Field("target_user_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("用户更新事务失败")

		return &types.UpdateUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  fmt.Sprintf("更新用户信息失败: %v", err),
			},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "update_user"),
		logx.Field("status", "success"),
		logx.Field("target_user_id", req.Id),
	).Info("更新用户信息成功")

	return &types.UpdateUserResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "用户信息更新成功",
		},
	}, nil
}
