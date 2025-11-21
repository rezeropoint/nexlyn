package platform

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/rezeropoint/etcdtrigger/v2/engine"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// platformManager 平台配置管理器实现
type platformManager struct {
	dbConn      sqlx.SqlConn
	configStore engine.Engine
	serviceName string
	podName     string
}

// newPlatformManager 创建平台配置管理器
func newPlatformManager(dbConn sqlx.SqlConn, config Config, configStore engine.Engine) (*platformManager, error) {
	manager := &platformManager{
		dbConn:      dbConn,
		configStore: configStore,
		serviceName: config.ServiceName,
		podName:     config.PodName,
	}

	return manager, nil
}

// Create 创建平台配置（元数据 + 配置），返回生成的UUID
func (m *platformManager) Create(ctx context.Context, metadata core.PlatformMetadata, config *core.PlatformConfig) (string, error) {
	// 1. 生成UUID作为平台配置ID
	platformID := uuid.New().String()

	// 2. 验证配置
	if config != nil {
		if err := config.Validate(); err != nil {
			return "", fmt.Errorf("配置验证失败: %w", err)
		}
		// 验证元数据与配置的一致性
		if metadata.Type != config.Type || metadata.TenantID != config.TenantID {
			return "", fmt.Errorf("元数据与配置的 Type/TenantID 不一致")
		}
	}

	// 3. 使用事务插入PostgreSQL元数据
	err := m.dbConn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		insertQuery := `
			INSERT INTO iot_platform_configs
			(id, name, type, description, enabled, tenant_id, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err := session.ExecCtx(ctx, insertQuery,
			platformID, metadata.Name, metadata.Type, metadata.Description,
			metadata.Enabled, metadata.TenantID, metadata.CreatedBy,
		)
		if err != nil {
			return fmt.Errorf("插入平台配置到PostgreSQL失败: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	// 4. 事务成功后，保存完整配置到Etcd（如果提供）
	if config != nil {
		etcdKey := core.BuildPlatformKey(metadata.Type, platformID)
		if err := m.configStore.PutConfig(ctx, etcdKey, config); err != nil {
			// Etcd写入失败，回滚PostgreSQL（事务已提交，需手动回滚）
			m.dbConn.ExecCtx(ctx, "DELETE FROM iot_platform_configs WHERE id = $1 AND tenant_id = $2", platformID, metadata.TenantID)
			return "", fmt.Errorf("保存平台配置到Etcd失败: %w", err)
		}
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.serviceName),
		logx.Field("pod", m.podName),
		logx.Field("module", "platform_manager"),
		logx.Field("operation", "create"),
		logx.Field("platform_type", metadata.Type),
		logx.Field("platform_id", platformID),
		logx.Field("tenant_id", metadata.TenantID),
		logx.Field("status", "success"),
	).Info("创建平台配置成功")

	return platformID, nil
}

// Get 获取平台配置（包含元数据 + 完整配置）
func (m *platformManager) Get(ctx context.Context, platformType core.PlatformType, id string, tenantID string) (*core.Platform, error) {
	// 1. 从数据库读取元数据
	var result struct {
		ID          string `db:"id"`
		Type        string `db:"type"`
		Name        string `db:"name"`
		Description string `db:"description"`
		Enabled     bool   `db:"enabled"`
		TenantID    string `db:"tenant_id"`
		CreatedBy   string `db:"created_by"`
		CreatedAt   string `db:"created_at"`
		UpdatedAt   string `db:"updated_at"`
	}
	query := `
		SELECT id, type, name, description, enabled, tenant_id, created_by, created_at, updated_at
		FROM iot_platform_configs
		WHERE id = $1 AND type = $2 AND deleted_at IS NULL
	`
	err := m.dbConn.QueryRowCtx(ctx, &result, query, id, platformType)
	if err != nil {
		return nil, core.ErrNotFound
	}

	// 2. 验证租户权限
	if result.TenantID != tenantID {
		return nil, fmt.Errorf("无权限访问该平台配置")
	}

	// 3. 构建平台信息
	platform := &core.Platform{
		ID:          result.ID,
		Type:        core.PlatformType(result.Type),
		Name:        result.Name,
		Description: result.Description,
		Enabled:     result.Enabled,
		TenantID:    result.TenantID,
		CreatedBy:   result.CreatedBy,
		UpdatedBy:   "", // 查询时未包含 UpdatedBy
		CreatedAt:   result.CreatedAt,
		UpdatedAt:   result.UpdatedAt,
	}

	// 4. 从Etcd读取完整配置（如果存在）
	etcdKey := core.BuildPlatformKey(platformType, id)
	var config core.PlatformConfig
	if m.configStore.GetConfig(etcdKey, &config) {
		platform.Config = &config
	}

	return platform, nil
}

// GetByID 根据ID获取平台配置（用于 Dispatcher，不检查租户）
func (m *platformManager) GetByID(id string) (*core.PlatformConfig, bool) {
	// 1. 从数据库查询 platformType（避免遍历所有类型）
	var platformType string
	query := "SELECT type FROM iot_platform_configs WHERE id = $1 AND deleted_at IS NULL"
	err := m.dbConn.QueryRowCtx(context.Background(), &platformType, query, id)
	if err != nil {
		return nil, false
	}

	// 2. 根据 type 从 Etcd 读取完整配置
	etcdKey := core.BuildPlatformKey(core.PlatformType(platformType), id)
	var config core.PlatformConfig
	if m.configStore.GetConfig(etcdKey, &config) {
		return &config, true
	}

	return nil, false
}

// List 列出平台配置（从数据库查询，返回Platform列表，不包含Config）
func (m *platformManager) List(ctx context.Context, tenantID string, platformType *core.PlatformType, keyword string, page, pageSize int) ([]*core.Platform, int64, error) {
	// 1. 构建SQL查询
	baseQuery := `
		FROM iot_platform_configs
		WHERE tenant_id = $1 AND deleted_at IS NULL
	`
	var args []any
	args = append(args, tenantID)
	argIndex := 2

	// 添加过滤条件
	if platformType != nil {
		baseQuery += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, *platformType)
		argIndex++
	}
	if keyword != "" {
		baseQuery += fmt.Sprintf(" AND (id ILIKE $%d OR name ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+keyword+"%")
		argIndex++
	}

	// 2. 查询总数
	var total int64
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := m.dbConn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}

	// 3. 查询列表
	listQuery := `
		SELECT id, name, type, description, enabled, tenant_id, created_by, created_at, updated_at
	` + baseQuery

	// 添加排序
	listQuery += " ORDER BY created_at DESC"

	// 添加分页
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	listQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)

	var results []struct {
		ID          string `db:"id"`
		Name        string `db:"name"`
		Type        string `db:"type"`
		Description string `db:"description"`
		Enabled     bool   `db:"enabled"`
		TenantID    string `db:"tenant_id"`
		CreatedBy   string `db:"created_by"`
		CreatedAt   string `db:"created_at"`
		UpdatedAt   string `db:"updated_at"`
	}

	err = m.dbConn.QueryRowsCtx(ctx, &results, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询列表失败: %w", err)
	}

	// 4. 转换结果（返回Platform列表，不包含Config）
	platformList := make([]*core.Platform, 0, len(results))
	for _, r := range results {
		platform := &core.Platform{
			ID:          r.ID,
			Type:        core.PlatformType(r.Type),
			Name:        r.Name,
			Description: r.Description,
			Enabled:     r.Enabled,
			TenantID:    r.TenantID,
			CreatedBy:   r.CreatedBy,
			UpdatedBy:   "", // List查询未包含UpdatedBy
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
			Config:      nil, // List不返回配置信息
		}
		platformList = append(platformList, platform)
	}

	return platformList, total, nil
}

// Update 更新平台配置（元数据 + 配置）
// - config 为nil表示删除Etcd配置
// - config 不为nil表示更新Etcd配置
func (m *platformManager) Update(ctx context.Context, platformType core.PlatformType, id string, tenantID string, metadata core.PlatformMetadata, config *core.PlatformConfig) error {
	// 1. 验证新配置
	if config != nil {
		if err := config.Validate(); err != nil {
			return fmt.Errorf("配置验证失败: %w", err)
		}
		// 验证元数据与配置的一致性
		if metadata.Type != config.Type || metadata.TenantID != config.TenantID {
			return fmt.Errorf("元数据与配置的 Type/TenantID 不一致")
		}
	}

	// 2. 使用事务执行所有PostgreSQL操作
	err := m.dbConn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 2.1 从数据库查询旧配置并验证权限
		var oldMetadata struct {
			Type     string `db:"type"`
			TenantID string `db:"tenant_id"`
		}
		checkQuery := "SELECT type, tenant_id FROM iot_platform_configs WHERE id = $1 AND deleted_at IS NULL"
		err := session.QueryRowCtx(ctx, &oldMetadata, checkQuery, id)
		if err != nil {
			return core.ErrNotFound
		}

		// 2.2 验证租户权限
		if oldMetadata.TenantID != tenantID {
			return fmt.Errorf("无权限修改其他租户的平台配置")
		}

		// 2.3 验证关键字段不可修改
		if string(metadata.Type) != oldMetadata.Type {
			return fmt.Errorf("不允许修改平台类型")
		}
		if metadata.TenantID != oldMetadata.TenantID {
			return fmt.Errorf("不允许修改租户ID")
		}

		// 2.4 更新PostgreSQL元数据
		updateQuery := `
			UPDATE iot_platform_configs
			SET name = $1, description = $2, enabled = $3, updated_by = $4, updated_at = CURRENT_TIMESTAMP
			WHERE id = $5 AND type = $6
		`
		_, err = session.ExecCtx(ctx, updateQuery,
			metadata.Name, metadata.Description, metadata.Enabled, metadata.UpdatedBy, id, platformType,
		)
		if err != nil {
			return fmt.Errorf("更新PostgreSQL失败: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 3. 事务成功后，处理Etcd配置
	etcdKey := core.BuildPlatformKey(platformType, id)
	if config != nil {
		// 更新Etcd配置
		if err := m.configStore.PutConfig(ctx, etcdKey, config); err != nil {
			return fmt.Errorf("更新平台配置到Etcd失败: %w", err)
		}
	} else {
		// 删除Etcd配置
		m.configStore.DeleteConfig(ctx, etcdKey)
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.serviceName),
		logx.Field("pod", m.podName),
		logx.Field("module", "platform_manager"),
		logx.Field("operation", "update"),
		logx.Field("platform_type", platformType),
		logx.Field("platform_id", id),
		logx.Field("tenant_id", tenantID),
		logx.Field("status", "success"),
	).Info("更新平台配置成功")

	return nil
}

// Delete 删除平台配置
func (m *platformManager) Delete(ctx context.Context, platformType core.PlatformType, id string, tenantID string) error {
	// 1. 使用事务执行所有PostgreSQL操作
	err := m.dbConn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1.1 从数据库查询验证权限
		var dbTenantID string
		checkQuery := "SELECT tenant_id FROM iot_platform_configs WHERE id = $1 AND type = $2 AND deleted_at IS NULL"
		err := session.QueryRowCtx(ctx, &dbTenantID, checkQuery, id, platformType)
		if err != nil {
			return core.ErrNotFound
		}

		// 1.2 验证租户权限
		if dbTenantID != tenantID {
			return fmt.Errorf("无权限删除其他租户的平台配置")
		}

		// 1.3 检查平台配置是否被设备模板使用
		usedByTemplates, err := m.checkPlatformInUseWithSession(ctx, session, id)
		if err != nil {
			return fmt.Errorf("检查平台配置使用情况失败: %w", err)
		}
		if len(usedByTemplates) > 0 {
			// 构建详细错误信息
			templatesInfo := ""
			for i, template := range usedByTemplates {
				if i > 0 {
					templatesInfo += ", "
				}
				templatesInfo += template
			}
			return fmt.Errorf("%w: 以下设备模板正在使用该平台配置: %s", core.ErrPlatformInUse, templatesInfo)
		}

		// 1.4 软删除PostgreSQL记录
		deleteQuery := "UPDATE iot_platform_configs SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND type = $2"
		_, err = session.ExecCtx(ctx, deleteQuery, id, platformType)
		if err != nil {
			return fmt.Errorf("删除PostgreSQL记录失败: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 2. 事务成功后，从Etcd删除配置
	etcdKey := core.BuildPlatformKey(platformType, id)
	if err := m.configStore.DeleteConfig(ctx, etcdKey); err != nil {
		return fmt.Errorf("从Etcd删除平台配置失败: %w", err)
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.serviceName),
		logx.Field("pod", m.podName),
		logx.Field("module", "platform_manager"),
		logx.Field("operation", "delete"),
		logx.Field("platform_type", platformType),
		logx.Field("platform_id", id),
		logx.Field("tenant_id", tenantID),
		logx.Field("status", "success"),
	).Info("删除平台配置成功")

	return nil
}

// Close 关闭管理器
func (m *platformManager) Close() error {
	// sqlx.SqlConn 不需要显式关闭
	return nil
}
