package tenant

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type CreateTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建租户
func NewCreateTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantLogic {
	return &CreateTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTenantLogic) CreateTenant(req *types.CreateTenantRequest) (resp *types.CreateTenantResponse, err error) {
	// 使用前端传递的租户标识符
	tenantKey := req.TenantKey

	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tenant"),
		logx.Field("operation", "create_tenant"),
		logx.Field("status", "started"),
		logx.Field("tenant_id", tenantKey),
		logx.Field("tenant_name", req.TenantName),
	).Info("开始创建租户")

	// 获取当前操作者（用于权限验证和审计）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "create_tenant"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户失败")
		return &types.CreateTenantResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 权限验证：检查是否有租户创建权限（使用Casbinx）
	hasCreatePermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: core.ResourceTenant, Action: core.ActionWrite})
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "create_tenant"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")
		return &types.CreateTenantResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}

	if !hasCreatePermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "create_tenant"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")
		return &types.CreateTenantResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要租户创建权限"},
		}, nil
	}

	// 设置默认值
	if req.Status == "" {
		req.Status = "active"
	}

	if req.MaxUsers <= 0 {
		req.MaxUsers = 100
	}

	// 验证状态值
	if req.Status != "active" && req.Status != "inactive" {
		return &types.CreateTenantResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "状态值无效：必须为active或inactive"},
		}, nil
	}

	// 预先验证角色权限（在创建任何数据之前）
	if l.svcCtx.Casbinx != nil {
		adminRoleKey := req.AdminRoleKey
		if adminRoleKey == "" {
			return &types.CreateTenantResponse{
				BaseResponse: types.BaseResponse{Code: 400, Msg: "必须指定租户管理员角色（adminRoleKey）"},
			}, nil
		}

		// 检查角色是否存在
		role, err := l.svcCtx.Casbinx.GetRole(adminRoleKey)
		if err != nil {
			return &types.CreateTenantResponse{
				BaseResponse: types.BaseResponse{Code: 400, Msg: fmt.Sprintf("指定的管理员角色 '%s' 不存在", adminRoleKey)},
			}, nil
		}

		// 检查角色权限是否符合租户管理员要求
		hasTenantPermissions := false
		hasUserManagePermissions := false

		for _, perm := range role.Permissions {
			// 检查是否包含租户管理权限（不允许）
			if perm.Resource == core.ResourceTenant {
				hasTenantPermissions = true
			}
			// 检查是否有用户管理权限（必需）
			if perm.Resource == core.ResourceUser && (perm.Action == core.ActionWrite || perm.Action == core.ActionDelete) {
				hasUserManagePermissions = true
			}
		}

		// 验证权限要求
		if hasTenantPermissions {
			return &types.CreateTenantResponse{
				BaseResponse: types.BaseResponse{Code: 400, Msg: fmt.Sprintf("角色 '%s' 包含租户管理权限，租户内管理员不允许跨租户操作", adminRoleKey)},
			}, nil
		}

		if !hasUserManagePermissions {
			return &types.CreateTenantResponse{
				BaseResponse: types.BaseResponse{Code: 400, Msg: fmt.Sprintf("角色 '%s' 缺少用户管理权限，无法作为租户管理员角色", adminRoleKey)},
			}, nil
		}
	}

	// 检查租户标识是否已存在
	var existingKeyCount int
	checkTenantKeyQuery := "SELECT COUNT(*) FROM system_tenants WHERE tenant_key = $1"
	err = l.svcCtx.DBConn.QueryRow(&existingKeyCount, checkTenantKeyQuery, tenantKey)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "create_tenant"),
			logx.Field("status", "failed"),
			logx.Field("tenant_key", tenantKey),
			logx.Field("error", err.Error()),
		).Error("检查租户标识是否存在失败")
		return &types.CreateTenantResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantQuery, err)},
		}, nil
	}

	if existingKeyCount > 0 {
		return &types.CreateTenantResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "租户标识已存在"},
		}, nil
	}

	// 检查租户名称是否已存在
	var existingNameCount int
	checkTenantNameQuery := "SELECT COUNT(*) FROM system_tenants WHERE tenant_name = $1"
	err = l.svcCtx.DBConn.QueryRow(&existingNameCount, checkTenantNameQuery, req.TenantName)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "create_tenant"),
			logx.Field("status", "failed"),
			logx.Field("tenant_key", tenantKey),
			logx.Field("error", err.Error()),
		).Error("检查租户名称是否存在失败")
		return &types.CreateTenantResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantQuery, err)},
		}, nil
	}

	if existingNameCount > 0 {
		return &types.CreateTenantResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "租户名称已存在"},
		}, nil
	}

	// 用于存储租户和用户创建结果
	var tenantResult struct {
		Id string `db:"id"`
	}
	var adminUserKey, adminUserName, adminPassword string
	var adminUserInternalID string

	// 使用单个事务处理整个租户初始化过程（租户创建 + 管理员用户创建）
	err = l.svcCtx.DBConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 插入租户记录
		var expiresAt *string
		if req.ExpiresAt != "" {
			expiresAt = &req.ExpiresAt
		}

		insertTenantQuery := `
			INSERT INTO system_tenants (
				tenant_key, tenant_name, description, contact_email,
				status, max_users, expires_at,
				created_by, updated_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id`

		err := session.QueryRowCtx(ctx, &tenantResult, insertTenantQuery,
			tenantKey, req.TenantName, req.Description, req.ContactEmail,
			req.Status, req.MaxUsers, expiresAt,
			jwtUser.UserId, jwtUser.UserId)

		if err != nil {
			return fmt.Errorf("插入租户记录失败: %w", err)
		}

		// 处理标签关联
		if len(req.Tags) > 0 {
			for _, tagLabel := range req.Tags {
				// 查找或创建标签
				var tagResult struct {
					Id string `db:"id"`
				}
				findTagQuery := "SELECT id FROM system_tag_definitions WHERE scope = $1 AND label = $2"
				err = session.QueryRowCtx(ctx, &tagResult, findTagQuery, "tenant", tagLabel)
				if err != nil {
					// 标签不存在，创建新标签
					createTagQuery := "INSERT INTO system_tag_definitions (scope, label) VALUES ($1, $2) RETURNING id"
					err = session.QueryRowCtx(ctx, &tagResult, createTagQuery, "tenant", tagLabel)
					if err != nil {
						return fmt.Errorf("创建标签 %s 失败: %w", tagLabel, err)
					}
				}

				// 创建租户-标签关联
				insertTagRelationQuery := "INSERT INTO system_tenant_tag_relations (tenant_id, tag_id) VALUES ($1, $2)"
				_, err = session.ExecCtx(ctx, insertTagRelationQuery, tenantResult.Id, tagResult.Id)
				if err != nil {
					return fmt.Errorf("创建租户标签关联失败: %w", err)
				}
			}
		}

		// 为新租户创建默认管理员用户（在同一个事务中）
		if l.svcCtx.Casbinx != nil {
			// 设置管理员用户参数默认值
			adminUserName = req.AdminUserName
			if adminUserName == "" {
				adminUserName = "admin"
			}

			// 生成管理员用户Key（符合租户初始化模式）
			adminUserKey = fmt.Sprintf("%s-%s-001", adminUserName, tenantKey)

			// 生成或使用提供的密码（遵循创建用户的惯例）
			adminPassword = req.AdminPassword
			if adminPassword == "" {
				// 使用与创建用户相同的默认密码格式
				adminPassword = "Admin123!"
			}

			// 哈希密码
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("管理员密码哈希失败: %w", err)
			}

			// 检查管理员用户名是否已全局存在
			var usernameCount int
			checkUsernameQuery := `SELECT COUNT(*) FROM system_users WHERE user_name = $1 AND status != 'deleted'`
			err = session.QueryRowCtx(ctx, &usernameCount, checkUsernameQuery, adminUserName)
			if err != nil {
				return fmt.Errorf("检查管理员用户名是否存在失败: %w", err)
			}

			if usernameCount > 0 {
				return fmt.Errorf("管理员用户名已被使用")
			}

			// 创建管理员用户记录
			adminName := req.AdminName
			if adminName == "" {
				adminName = fmt.Sprintf("%s租户管理员", req.TenantName)
			}

			adminEmail := req.AdminEmail
			if adminEmail == "" {
				adminEmail = fmt.Sprintf("admin@%s.local", tenantKey)
			}

			// 检查管理员邮箱是否已存在
			var emailCount int
			checkEmailQuery := `SELECT COUNT(*) FROM system_users WHERE email = $1 AND tenant_id = $2 AND status != 'deleted'`
			err = session.QueryRowCtx(ctx, &emailCount, checkEmailQuery, adminEmail, tenantResult.Id)
			if err != nil {
				return fmt.Errorf("检查管理员邮箱是否存在失败: %w", err)
			}

			if emailCount > 0 {
				return fmt.Errorf("管理员邮箱在当前租户中已被使用")
			}

			// 插入管理员用户记录
			insertAdminUserQuery := `
				INSERT INTO system_users (
					user_key, user_name, password_hash, name, email,
					tenant_id, status
				) VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id`

			err = session.QueryRowCtx(ctx, &adminUserInternalID, insertAdminUserQuery,
				adminUserKey, adminUserName, string(hashedPassword), adminName, adminEmail,
				tenantResult.Id, "active")
			if err != nil {
				return fmt.Errorf("创建管理员用户失败: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "create_tenant"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", tenantKey),
			logx.Field("error", err.Error()),
		).Error("租户创建事务失败")
		return &types.CreateTenantResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantCreate, err)},
		}, nil
	}

	// 使用 CasbinX 初始化租户并为管理员用户分配角色
	// 使用 InitializeTenant 方法，它会验证角色适用性并绕过系统权限检查
	if l.svcCtx.Casbinx != nil && adminUserKey != "" {
		adminRoleKey := req.AdminRoleKey
		if adminRoleKey == "" {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "tenant"),
				logx.Field("operation", "create_tenant"),
				logx.Field("status", "failed"),
				logx.Field("tenant_id", tenantKey),
			).Error("未指定租户管理员角色")
			return &types.CreateTenantResponse{
				BaseResponse: types.BaseResponse{Code: 400, Msg: "必须指定租户管理员角色（adminRoleKey）"},
			}, nil
		}

		err = l.svcCtx.Casbinx.InitializeTenant(tenantKey, adminUserKey, adminRoleKey)
		if err != nil {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "tenant"),
				logx.Field("operation", "create_tenant"),
				logx.Field("status", "failed"),
				logx.Field("tenant_id", tenantKey),
				logx.Field("admin_user_key", adminUserKey),
				logx.Field("admin_role_key", adminRoleKey),
				logx.Field("error", err.Error()),
			).Error("租户初始化失败")
			return &types.CreateTenantResponse{
				BaseResponse: types.BaseResponse{Code: 500, Msg: fmt.Sprintf("租户初始化失败: %v", err)},
			}, nil
		}

		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tenant"),
			logx.Field("operation", "create_tenant"),
			logx.Field("status", "success"),
			logx.Field("tenant_id", tenantKey),
			logx.Field("admin_user_key", adminUserKey),
			logx.Field("admin_role_key", adminRoleKey),
		).Info("租户管理员用户创建完成，已分配角色")
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tenant"),
		logx.Field("operation", "create_tenant"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", tenantKey),
		logx.Field("tenant_name", req.TenantName),
		logx.Field("created_by", jwtUser.UserId),
	).Info("创建租户成功")

	return &types.CreateTenantResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "创建租户成功",
		},
		Data: struct {
			TenantKey     string `json:"tenantKey"`
			TenantId      string `json:"tenantId"`
			TenantName    string `json:"tenantName"`
			AdminUserKey  string `json:"adminUserKey,omitempty"`
			AdminUserName string `json:"adminUserName,omitempty"`
			AdminPassword string `json:"adminPassword,omitempty"`
		}{
			TenantKey:     tenantKey,
			TenantId:      tenantResult.Id,
			TenantName:    req.TenantName,
			AdminUserKey:  adminUserKey,
			AdminUserName: adminUserName,
			AdminPassword: func() string {
				// 只有生成的密码才返回，用户提供的密码不返回
				if req.AdminPassword == "" {
					return adminPassword
				}
				return ""
			}(),
		},
	}, nil
}

// 注释：状态字段现在直接使用字符串类型存储，不再需要mapStatusToInt函数
