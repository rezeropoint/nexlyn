package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"
	"github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/rezeropoint/casbinx/core"

	"github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type CreateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建新用户
func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateUserLogic) CreateUser(req *types.CreateUserRequest) (resp *types.CreateUserResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "create_user"),
		logx.Field("status", "started"),
		logx.Field("email", req.Email),
		logx.Field("username", req.UserName),
	).Info("开始创建新用户")

	// 简单的邮箱格式验证
	if !strings.Contains(req.Email, "@") {
		return &types.CreateUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "电子邮件格式不正确",
			},
		}, nil
	}

	// 获取当前用户的租户信息
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "create_user"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取当前用户信息失败")

		return &types.CreateUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
		}, nil
	}

	// 权限验证：检查用户创建权限
	hasCreatePermission, err := l.svcCtx.Casbinx.CheckPermission(
		currentUser.UserKey,
		currentUser.TenantKey,
		core.Permission{Resource: core.ResourceUser, Action: core.ActionWrite},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "create_user"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")
		return &types.CreateUserResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasCreatePermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "create_user"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("current_tenant", currentUser.TenantKey),
			logx.Field("target_tenant", req.TenantId),
		).Error("用户权限不足")
		return &types.CreateUserResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要用户创建权限"},
		}, nil
	}

	// 检查跨租户权限：如果不是在同一租户操作，需要跨租户权限
	// 注意：这里应该比较 TenantId (UUID) 而不是 TenantKey (字符串标识符)
	isCrossTenant := currentUser.TenantId != req.TenantId
	if isCrossTenant {
		hasCrossTenantPermission, err := l.svcCtx.Casbinx.CheckPermission(
			currentUser.UserKey,
			currentUser.TenantKey,
			core.Permission{Resource: core.ResourceTenant, Action: core.ActionWrite},
		)
		if err != nil {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "user"),
				logx.Field("operation", "create_user"),
				logx.Field("status", "failed"),
				logx.Field("user_key", currentUser.UserKey),
				logx.Field("error", err.Error()),
			).Error("跨租户权限检查失败")
			return &types.CreateUserResponse{
				BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
			}, nil
		}

		if !hasCrossTenantPermission {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "user"),
				logx.Field("operation", "create_user"),
				logx.Field("status", "failed"),
				logx.Field("user_key", currentUser.UserKey),
				logx.Field("current_tenant", currentUser.TenantId),
				logx.Field("target_tenant", req.TenantId),
			).Error("跨租户权限不足")
			return &types.CreateUserResponse{
				BaseResponse: types.BaseResponse{
					Code: 403,
					Msg:  "权限不足：需要跨租户用户创建权限",
				},
			}, nil
		}
	}

	// 检查邮箱在当前租户是否已存在
	var existingCount int

	// 首先获取租户的UUID和用户限制信息
	var tenantInfo struct {
		Id       string `db:"id"`
		MaxUsers int    `db:"max_users"`
	}
	// 注意：req.TenantId 是 UUID 格式，应该使用 id 字段而不是 tenant_key 字段
	getTenantQuery := `SELECT id, max_users FROM system_tenants WHERE id = $1 AND status = 'active'`
	err = l.svcCtx.DBConn.QueryRow(&tenantInfo, getTenantQuery, req.TenantId)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "create"),
			logx.Field("status", "failed"),
			logx.Field("tenantId", req.TenantId),
			logx.Field("error", err.Error()),
		).Error("租户不存在或已被禁用")

		return &types.CreateUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "指定的租户不存在或已被禁用",
			},
		}, nil
	}

	// 检查租户当前用户数量是否已达到上限
	var currentUserCount int
	countUserQuery := `SELECT COUNT(*) FROM system_users WHERE tenant_id = $1 AND status != 'deleted'`
	err = l.svcCtx.DBConn.QueryRow(&currentUserCount, countUserQuery, tenantInfo.Id)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "create_user"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询租户用户数量失败")

		return &types.CreateUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  config.FormatError(config.ErrMsgTenantUserCheck, err),
			},
		}, nil
	}

	if currentUserCount >= tenantInfo.MaxUsers {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "create_user"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", req.TenantId),
			logx.Field("current_users", currentUserCount),
			logx.Field("max_users", tenantInfo.MaxUsers),
		).Error("租户用户数量已达上限")

		return &types.CreateUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  fmt.Sprintf("租户用户数量已达上限（%d/%d）", currentUserCount, tenantInfo.MaxUsers),
			},
		}, nil
	}

	// 检查邮箱是否已存在
	checkEmailQuery := `SELECT COUNT(*) FROM system_users WHERE email = $1 AND tenant_id = $2 AND status != 'deleted'`
	err = l.svcCtx.DBConn.QueryRow(&existingCount, checkEmailQuery, req.Email, tenantInfo.Id)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "create_user"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("检查邮箱是否存在失败")

		return &types.CreateUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  config.FormatError(config.ErrMsgUserQuery, err),
			},
		}, nil
	}

	if existingCount > 0 {
		return &types.CreateUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 409,
				Msg:  "该邮箱在当前租户中已被使用",
			},
		}, nil
	}

	// 检查用户名是否已全局存在
	var usernameCount int
	checkUsernameQuery := `SELECT COUNT(*) FROM system_users WHERE user_name = $1 AND status != 'deleted'`
	err = l.svcCtx.DBConn.QueryRow(&usernameCount, checkUsernameQuery, req.UserName)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "create_user"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("检查用户名是否存在失败")

		return &types.CreateUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  config.FormatError(config.ErrMsgUserQuery, err),
			},
		}, nil
	}

	if usernameCount > 0 {
		return &types.CreateUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 409,
				Msg:  "该用户名已被使用",
			},
		}, nil
	}

	// 使用用户名作为用户key，确保user_key和user_name一致
	userKey := req.UserName

	// 生成默认密码并哈希（实际应用中可能通过邮件发送临时密码）
	defaultPassword := "User123!"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "create_user"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("生成密码哈希失败")

		return &types.CreateUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  config.FormatError(config.ErrMsgUserPasswordHash, err),
			},
		}, nil
	}

	// 用于存储用户创建结果
	var newUserInternalID string
	var skylarkRemoteUserID string

	// ========================================
	// 事务内同步：创建本地数据 + 同步到 Skylark
	// ========================================
	// 使用事务处理整个用户创建过程（用户创建 + Skylark同步 + 标签关联 + 地理信息）
	err = l.svcCtx.DBConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 插入用户基本信息（不再包含role字段）
		insertUserQuery := `
			INSERT INTO system_users (
				user_key, user_name, password_hash, name, email, phone, avatar, 
	            signature, title, tenant_id,
	            country, address, status
			) VALUES (
	            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 'active'
			) RETURNING id
		`

		err := session.QueryRowCtx(ctx, &newUserInternalID, insertUserQuery,
			userKey, req.UserName, string(passwordHash), req.Name, req.Email,
			req.Phone, req.Avatar, req.Signature, req.Title, tenantInfo.Id,
			req.Country, req.Address)

		if err != nil {
			return fmt.Errorf("插入用户基本信息失败: %w", err)
		}

		// ========================================
		// 事务内同步到 Skylark
		// ========================================
		// 如果启用了 Skylark 同步，在事务内同步用户到 Skylark
		if l.svcCtx.Config.SkylarkSyncEnabled && l.svcCtx.AdminSyncClient != nil {
			syncCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			syncResp, syncErr := l.svcCtx.AdminSyncClient.SyncUser(syncCtx, &pb.SyncUserReq{
				TenantId:    currentUser.TenantId,
				LocalUserId: newUserInternalID, // 使用数据库生成的 UUID
				Name:        req.Name,
				Identifier:  req.Email,
				Phone:       req.Phone,
				Openid:      "",
			})

			if syncErr != nil {
				logx.WithContext(ctx).WithFields(
					logx.Field("service", l.svcCtx.Config.RestConf.Name),
					logx.Field("pod", l.svcCtx.PodName),
					logx.Field("module", "user"),
					logx.Field("operation", "sync_to_skylark"),
					logx.Field("status", "failed"),
					logx.Field("user_id", newUserInternalID),
					logx.Field("error", syncErr.Error()),
				).Error("Skylark 用户同步失败，事务将回滚")

				return fmt.Errorf("用户同步到 Skylark 失败: %v", syncErr)
			}

			if !syncResp.Success {
				logx.WithContext(ctx).WithFields(
					logx.Field("service", l.svcCtx.Config.RestConf.Name),
					logx.Field("pod", l.svcCtx.PodName),
					logx.Field("module", "user"),
					logx.Field("operation", "sync_to_skylark"),
					logx.Field("status", "failed"),
					logx.Field("user_id", newUserInternalID),
					logx.Field("message", syncResp.Message),
					logx.Field("error_code", syncResp.ErrorCode),
				).Error("Skylark 返回同步失败，事务将回滚")

				return fmt.Errorf("用户同步到 Skylark 失败: %s", syncResp.Message)
			}

			skylarkRemoteUserID = syncResp.RemoteUserId

			logx.WithContext(ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "user"),
				logx.Field("operation", "sync_to_skylark"),
				logx.Field("status", "success"),
				logx.Field("user_id", newUserInternalID),
				logx.Field("remote_user_id", skylarkRemoteUserID),
			).Info("用户已成功同步到 Skylark")
		}

		// 校验并关联用户标签（使用标签ID）
		if len(req.TagIds) > 0 {
			// 校验所有标签ID是否均有效且属于 'user' 作用域
			var validCount int
			validateQuery := `SELECT COUNT(*) FROM system_tag_definitions WHERE id = ANY($1::uuid[]) AND scope = 'user'`
			err = session.QueryRowCtx(ctx, &validCount, validateQuery, pq.Array(req.TagIds))
			if err != nil {
				return fmt.Errorf("校验用户标签失败: %w", err)
			}

			if validCount != len(req.TagIds) {
				return fmt.Errorf("一个或多个提供的标签ID无效或不属于'user'作用域")
			}

			// 关联所有有效标签
			for _, tagID := range req.TagIds {
				_, err = session.ExecCtx(ctx,
					`INSERT INTO user_tag_relations (user_id, tag_id) VALUES ($1, $2) ON CONFLICT (user_id, tag_id) DO NOTHING`,
					newUserInternalID, tagID,
				)
				if err != nil {
					return fmt.Errorf("关联用户标签 %s 失败: %w", tagID, err)
				}
			}
		}

		// 插入用户地理信息（如果有）
		if req.Geographic.Province.Key != "" || req.Geographic.City.Key != "" {
			insertGeoQuery := `
				INSERT INTO user_geographic (
					user_id, province_key, province_label, city_key, city_label
				) VALUES ($1, $2, $3, $4, $5)
			`
			_, err = session.ExecCtx(ctx, insertGeoQuery, newUserInternalID,
				req.Geographic.Province.Key, req.Geographic.Province.Label,
				req.Geographic.City.Key, req.Geographic.City.Label)

			if err != nil {
				return fmt.Errorf("插入用户地理信息失败: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user"),
			logx.Field("operation", "create_user"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("用户创建事务失败（包含 Skylark 同步失败时的自动回滚）")

		return &types.CreateUserResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  config.FormatError(config.ErrMsgUserCreate, err),
			},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user"),
		logx.Field("operation", "create_user"),
		logx.Field("status", "success"),
		logx.Field("user_key", userKey),
		logx.Field("username", req.UserName),
		logx.Field("email", req.Email),
		logx.Field("tenantId", req.TenantId),
	).Info("创建新用户成功")

	return &types.CreateUserResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "用户创建成功",
		},
		Data: types.CreateUserData{
			UserKey:  userKey,
			UserName: req.UserName,
			Password: defaultPassword,
		},
	}, nil
}
