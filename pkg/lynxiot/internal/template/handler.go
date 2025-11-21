package template

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rezeropoint/etcdtrigger/v2/engine"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// templateManager 模板管理器实现
type templateManager struct {
	dbConn      sqlx.SqlConn
	configStore engine.Engine
	config      Config
}

// newTemplateManager 创建模板管理器实例
func newTemplateManager(dbConn sqlx.SqlConn, configStore engine.Engine, config Config) (*templateManager, error) {
	if configStore == nil {
		return nil, fmt.Errorf("configStore 未设置")
	}

	return &templateManager{
		dbConn:      dbConn,
		configStore: configStore,
		config:      config,
	}, nil
}

// Create 创建模板（支持在线检测配置 + 业务数据配置 + 控制配置）
func (m *templateManager) Create(ctx context.Context, metadata core.TemplateMetadata, onlineConfig *core.OnlineDetectionConfig, businessConfig *core.DataProcessingConfig, controlConfig *core.DeviceControlConfig) (string, error) {
	// 1. 验证配置
	if onlineConfig != nil {
		if err := onlineConfig.Validate(); err != nil {
			return "", fmt.Errorf("在线检测配置验证失败: %w", err)
		}
	}
	if businessConfig != nil {
		if err := businessConfig.Validate(); err != nil {
			return "", fmt.Errorf("业务数据配置验证失败: %w", err)
		}
	}
	if controlConfig != nil {
		if err := controlConfig.Validate(); err != nil {
			return "", fmt.Errorf("设备控制配置验证失败: %w", err)
		}
	}

	// 2. 生成UUID
	id := uuid.New().String()

	// 3. 准备 topic_suffixes 和 used_platform_ids（如果提供了配置）
	var onlineTopicSuffixes []string
	var businessTopicSuffixes []string
	var controlTopicSuffixes []string
	var usedPlatformIDs []string
	if onlineConfig != nil {
		onlineTopicSuffixes = []string{onlineConfig.TopicSuffix}
	}
	if businessConfig != nil {
		businessTopicSuffixes = []string{businessConfig.TopicSuffix}
		// 提取平台ID列表
		usedPlatformIDs = extractPlatformIDs(businessConfig)
	}
	if controlConfig != nil {
		controlTopicSuffixes = []string{controlConfig.CommandSuffix}
	}

	// 4. 使用事务执行PostgreSQL操作
	err := m.dbConn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 4.1 检查model是否已存在
		var count int
		checkQuery := "SELECT COUNT(*) FROM iot_sensor_templates WHERE model = $1 AND tenant_id = $2 AND deleted_at IS NULL"
		err := session.QueryRowCtx(ctx, &count, checkQuery, metadata.Model, metadata.TenantID)
		if err != nil {
			return fmt.Errorf("检查模板是否存在失败: %w", err)
		}
		if count > 0 {
			return core.ErrTemplateAlreadyExists
		}

		// 4.2 插入PostgreSQL
		insertQuery := `
			INSERT INTO iot_sensor_templates
			(id, model, name, category, manufacturer, description, version, enabled, online_topic_suffixes, business_topic_suffixes, control_topic_suffixes, used_platform_ids, device_count, tenant_id, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 0, $13, $14)
		`
		_, err = session.ExecCtx(ctx, insertQuery,
			id, metadata.Model, metadata.Name, metadata.Category, metadata.Manufacturer,
			metadata.Description, metadata.Version, metadata.Enabled, pq.Array(onlineTopicSuffixes), pq.Array(businessTopicSuffixes), pq.Array(controlTopicSuffixes), pq.Array(usedPlatformIDs),
			metadata.TenantID, metadata.CreatedBy,
		)
		if err != nil {
			return fmt.Errorf("插入模板到PostgreSQL失败: %w", err)
		}

		// 4.3 处理标签关联（在事务中）
		if len(metadata.TagIDs) > 0 {
			err = m.replaceTemplateTags(ctx, session, id, metadata.TagIDs, metadata.TenantID)
			if err != nil {
				return fmt.Errorf("创建标签关联失败: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// 5. 事务成功后，写入Etcd配置（如果提供）
	// 填充配置的基本信息
	if onlineConfig != nil {
		onlineConfig.Category = metadata.Category
		onlineConfig.Model = metadata.Model
		onlineConfig.TenantID = metadata.TenantID

		err = m.saveOnlineDetectionConfig(ctx, onlineConfig)
		if err != nil {
			// Etcd写入失败，需要清理PostgreSQL数据
			m.dbConn.ExecCtx(ctx, "DELETE FROM iot_sensor_templates WHERE id = $1", id)
			return "", fmt.Errorf("写入在线检测配置失败: %w", err)
		}
	}

	if businessConfig != nil {
		businessConfig.Category = metadata.Category
		businessConfig.Model = metadata.Model

		err = m.saveDataProcessingConfig(ctx, businessConfig)
		if err != nil {
			// 回滚PostgreSQL和已写入的Etcd配置
			m.dbConn.ExecCtx(ctx, "DELETE FROM iot_sensor_templates WHERE id = $1", id)
			if onlineConfig != nil {
				onlineKey := core.BuildOnlineDetectionKey(metadata.Category, metadata.Model, onlineConfig.TopicSuffix)
				m.configStore.DeleteConfig(ctx, onlineKey)
			}
			return "", fmt.Errorf("写入业务数据配置失败: %w", err)
		}
	}

	if controlConfig != nil {
		controlConfig.Category = metadata.Category
		controlConfig.Model = metadata.Model
		controlConfig.TenantID = metadata.TenantID

		err = m.saveControlConfig(ctx, controlConfig)
		if err != nil {
			// 回滚PostgreSQL和所有已写入的Etcd配置
			m.dbConn.ExecCtx(ctx, "DELETE FROM iot_sensor_templates WHERE id = $1", id)
			if onlineConfig != nil {
				onlineKey := core.BuildOnlineDetectionKey(metadata.Category, metadata.Model, onlineConfig.TopicSuffix)
				m.configStore.DeleteConfig(ctx, onlineKey)
			}
			if businessConfig != nil {
				businessKey := core.BuildDataProcessingKey(metadata.Category, metadata.Model, businessConfig.TopicSuffix)
				m.configStore.DeleteConfig(ctx, businessKey)
			}
			return "", fmt.Errorf("写入控制配置失败: %w", err)
		}
	}

	return id, nil
}

// Update 更新模板（支持在线检测配置 + 业务数据配置 + 控制配置）
// 说明：
// - onlineConfig 为nil表示删除在线检测配置
// - businessConfig 为nil表示删除业务数据配置
// - controlConfig 为nil表示删除控制配置
func (m *templateManager) Update(ctx context.Context, id string, metadata core.TemplateMetadata, onlineConfig *core.OnlineDetectionConfig, businessConfig *core.DataProcessingConfig, controlConfig *core.DeviceControlConfig) error {
	// 1. 验证配置
	if onlineConfig != nil {
		if err := onlineConfig.Validate(); err != nil {
			return fmt.Errorf("在线检测配置验证失败: %w", err)
		}
	}
	if businessConfig != nil {
		if err := businessConfig.Validate(); err != nil {
			return fmt.Errorf("业务数据配置验证失败: %w", err)
		}
	}
	if controlConfig != nil {
		if err := controlConfig.Validate(); err != nil {
			return fmt.Errorf("控制配置验证失败: %w", err)
		}
	}

	// 2. 准备变量存储旧模板信息（用于Etcd操作）
	var oldTemplate oldTemplateDB

	// 3. 准备新的配置数据
	var newOnlineTopicSuffixes []string
	var newBusinessTopicSuffixes []string
	var newControlTopicSuffixes []string
	var newUsedPlatformIDs []string

	if onlineConfig != nil {
		newOnlineTopicSuffixes = []string{onlineConfig.TopicSuffix}
	}
	if businessConfig != nil {
		newBusinessTopicSuffixes = []string{businessConfig.TopicSuffix}
		newUsedPlatformIDs = extractPlatformIDs(businessConfig)
	}
	if controlConfig != nil {
		newControlTopicSuffixes = []string{controlConfig.CommandSuffix}
	}

	// 4. 使用事务执行所有PostgreSQL操作
	err := m.dbConn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 4.1 查询旧模板信息
		checkQuery := "SELECT tenant_id, model, category, online_topic_suffixes, business_topic_suffixes, control_topic_suffixes FROM iot_sensor_templates WHERE id = $1 AND deleted_at IS NULL"
		err := session.QueryRowCtx(ctx, &oldTemplate, checkQuery, id)
		if err == sql.ErrNoRows {
			return core.ErrTemplateNotFound
		}
		if err != nil {
			return fmt.Errorf("查询模板失败: %w", err)
		}

		// 4.2 验证租户权限
		if oldTemplate.TenantID != metadata.TenantID {
			return fmt.Errorf("无权限修改其他租户的模板")
		}

		// 4.3 严格验证 model 和 category（必须传递且不允许修改）
		if metadata.Model == "" {
			return fmt.Errorf("设备型号不能为空")
		}
		if metadata.Category == "" {
			return fmt.Errorf("设备类别不能为空")
		}
		if metadata.Model != oldTemplate.Model {
			return fmt.Errorf("不允许修改设备型号")
		}
		if metadata.Category != oldTemplate.Category {
			return fmt.Errorf("不允许修改设备类别")
		}

		// 4.4 更新PostgreSQL元数据（所有字段，值已验证未改变）
		updateQuery := `
			UPDATE iot_sensor_templates
			SET model = $2, name = $3, category = $4, manufacturer = $5, description = $6,
			    version = $7, enabled = $8, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
		`
		_, err = session.ExecCtx(ctx, updateQuery,
			id, metadata.Model, metadata.Name, metadata.Category, metadata.Manufacturer,
			metadata.Description, metadata.Version, metadata.Enabled,
		)
		if err != nil {
			return fmt.Errorf("更新PostgreSQL失败: %w", err)
		}

		// 4.5 更新online_topic_suffixes
		updateOnlineSuffixesQuery := "UPDATE iot_sensor_templates SET online_topic_suffixes = $1 WHERE id = $2"
		_, err = session.ExecCtx(ctx, updateOnlineSuffixesQuery, pq.Array(newOnlineTopicSuffixes), id)
		if err != nil {
			return fmt.Errorf("更新online_topic_suffixes失败: %w", err)
		}

		// 4.6 更新business_topic_suffixes和used_platform_ids
		updateBusinessQuery := "UPDATE iot_sensor_templates SET business_topic_suffixes = $1, used_platform_ids = $2 WHERE id = $3"
		_, err = session.ExecCtx(ctx, updateBusinessQuery, pq.Array(newBusinessTopicSuffixes), pq.Array(newUsedPlatformIDs), id)
		if err != nil {
			return fmt.Errorf("更新business_topic_suffixes和used_platform_ids失败: %w", err)
		}

		// 4.7 更新control_topic_suffixes
		updateControlQuery := "UPDATE iot_sensor_templates SET control_topic_suffixes = $1 WHERE id = $2"
		_, err = session.ExecCtx(ctx, updateControlQuery, pq.Array(newControlTopicSuffixes), id)
		if err != nil {
			return fmt.Errorf("更新control_topic_suffixes失败: %w", err)
		}

		// 4.8 处理标签关联（如果提供了TagIDs则替换，nil则不修改）
		if metadata.TagIDs != nil {
			err = m.replaceTemplateTags(ctx, session, id, metadata.TagIDs, metadata.TenantID)
			if err != nil {
				return fmt.Errorf("更新标签关联失败: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 5. 事务成功后，处理Etcd配置的删除和写入
	// 5.1 处理在线检测配置
	// 先删除所有旧的在线检测配置
	for _, suffix := range oldTemplate.OnlineTopicSuffixes {
		onlineKey := core.BuildOnlineDetectionKey(metadata.Category, metadata.Model, suffix)
		m.configStore.DeleteConfig(ctx, onlineKey)
	}
	// 如果提供了新配置，则写入
	if onlineConfig != nil {
		onlineConfig.Category = metadata.Category
		onlineConfig.Model = metadata.Model
		onlineConfig.TenantID = oldTemplate.TenantID

		err = m.saveOnlineDetectionConfig(ctx, onlineConfig)
		if err != nil {
			return fmt.Errorf("更新在线检测配置失败: %w", err)
		}
	}

	// 5.2 处理业务数据配置
	// 先删除所有旧的业务数据配置
	for _, suffix := range oldTemplate.BusinessTopicSuffixes {
		businessKey := core.BuildDataProcessingKey(metadata.Category, metadata.Model, suffix)
		m.configStore.DeleteConfig(ctx, businessKey)
	}
	// 如果提供了新配置，则写入
	if businessConfig != nil {
		businessConfig.Category = metadata.Category
		businessConfig.Model = metadata.Model

		err = m.saveDataProcessingConfig(ctx, businessConfig)
		if err != nil {
			return fmt.Errorf("更新业务数据配置失败: %w", err)
		}
	}

	// 5.3 处理控制配置
	// 先删除所有旧的控制配置
	for _, suffix := range oldTemplate.ControlTopicSuffixes {
		controlKey := core.BuildControlConfigKey(metadata.Category, metadata.Model, suffix)
		m.configStore.DeleteConfig(ctx, controlKey)
	}
	// 如果提供了新配置，则写入
	if controlConfig != nil {
		controlConfig.Category = metadata.Category
		controlConfig.Model = metadata.Model
		controlConfig.TenantID = oldTemplate.TenantID

		err = m.saveControlConfig(ctx, controlConfig)
		if err != nil {
			return fmt.Errorf("更新控制配置失败: %w", err)
		}
	}

	return nil
}

// Delete 删除模板（级联删除所有关联配置）
func (m *templateManager) Delete(ctx context.Context, id string) error {
	// 1. 查询模板信息（用于后续删除所有配置）
	var templateInfo templateInfoDB
	checkQuery := "SELECT device_count, model, category, online_topic_suffixes, business_topic_suffixes, control_topic_suffixes FROM iot_sensor_templates WHERE id = $1 AND deleted_at IS NULL"
	err := m.dbConn.QueryRowCtx(ctx, &templateInfo, checkQuery, id)
	if err == sql.ErrNoRows {
		return core.ErrTemplateNotFound
	}
	if err != nil {
		return fmt.Errorf("查询模板失败: %w", err)
	}

	// 2. 检查是否有关联设备
	if templateInfo.DeviceCount > 0 {
		return core.ErrTemplateInUse
	}

	// 3. 先删除所有Etcd配置（优先清理外部依赖）
	// 3.1 删除在线检测配置 (sensor-online/{category}/{model}/{topicSuffix})
	for _, suffix := range templateInfo.OnlineTopicSuffixes {
		onlineKey := core.BuildOnlineDetectionKey(templateInfo.Category, templateInfo.Model, suffix)
		m.configStore.DeleteConfig(ctx, onlineKey)
	}

	// 3.2 删除业务数据配置 (sensor-data/{category}/{model}/{topicSuffix})
	for _, suffix := range templateInfo.BusinessTopicSuffixes {
		businessKey := core.BuildDataProcessingKey(templateInfo.Category, templateInfo.Model, suffix)
		m.configStore.DeleteConfig(ctx, businessKey)
	}

	// 3.3 删除控制配置 (control-config/{category}/{model}/{suffix})
	for _, suffix := range templateInfo.ControlTopicSuffixes {
		controlKey := core.BuildControlConfigKey(templateInfo.Category, templateInfo.Model, suffix)
		m.configStore.DeleteConfig(ctx, controlKey)
	}

	// 4. 最后软删除PostgreSQL记录（使用事务）
	err = m.dbConn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		deleteQuery := "UPDATE iot_sensor_templates SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1"
		_, err := session.ExecCtx(ctx, deleteQuery, id)
		if err != nil {
			return fmt.Errorf("删除PostgreSQL记录失败: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// Get 获取模板详情
func (m *templateManager) Get(ctx context.Context, id string) (*core.Template, error) {
	// 1. 从PostgreSQL查询元数据
	var result templateDB

	query := `
		SELECT id, model, name, category, manufacturer, description, version, enabled,
		       device_count, tenant_id, created_by, created_at, updated_at,
		       online_topic_suffixes, business_topic_suffixes, control_topic_suffixes
		FROM iot_sensor_templates
		WHERE id = $1 AND deleted_at IS NULL
	`
	err := m.dbConn.QueryRowCtx(ctx, &result, query, id)
	if err == sql.ErrNoRows {
		return nil, core.ErrTemplateNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询模板失败: %w", err)
	}

	// 2. 构建模板对象
	template := &core.Template{
		ID:           result.ID,
		Model:        result.Model,
		Name:         result.Name,
		Category:     result.Category,
		Manufacturer: result.Manufacturer,
		Description:  result.Description,
		Version:      result.Version,
		Enabled:      result.Enabled,
		DeviceCount:  result.DeviceCount,
		TenantID:     result.TenantID,
		CreatedBy:    result.CreatedBy,
	}

	// 3. 处理时间字段
	if result.CreatedAt.Valid {
		template.CreatedAt = result.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if result.UpdatedAt.Valid {
		template.UpdatedAt = result.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}

	// 4. 查询并填充标签
	tags, err := m.getTemplateTags(ctx, id)
	if err != nil {
		// 标签查询失败不影响主流程，记录错误
		tags = []core.DeviceTagSummary{}
	}
	template.Tags = tags

	// 5. 从Etcd查询在线检测配置（取第一个）
	if len(result.OnlineTopicSuffixes) > 0 {
		onlineConfig, found := m.getOnlineDetectionConfig(result.Category, result.Model, result.OnlineTopicSuffixes[0])
		if found && onlineConfig != nil {
			template.OnlineConfig = onlineConfig
		}
		// 配置查询失败不影响主流程
	}

	// 6. 从Etcd查询业务数据处理配置（取第一个）
	if len(result.BusinessTopicSuffixes) > 0 {
		businessConfig, found := m.getDataProcessingConfig(result.Category, result.Model, result.BusinessTopicSuffixes[0])
		if found && businessConfig != nil {
			template.BusinessConfig = businessConfig
		}
		// 配置查询失败不影响主流程
	}

	// 7. 从Etcd查询控制配置（取第一个）
	if len(result.ControlTopicSuffixes) > 0 {
		controlConfig, found := m.getControlConfig(result.Category, result.Model, result.ControlTopicSuffixes[0])
		if found && controlConfig != nil {
			template.ControlConfig = controlConfig
		}
		// 配置查询失败不影响主流程
	}

	return template, nil
}

// List 查询模板列表
func (m *templateManager) List(ctx context.Context, query core.TemplateQuery) ([]*core.TemplateSummary, int64, error) {
	// 1. 构建SQL查询
	baseQuery := `
		FROM iot_sensor_templates
		WHERE tenant_id = $1 AND deleted_at IS NULL
	`
	var args []any
	args = append(args, query.TenantID)
	argIndex := 2

	// 添加过滤条件
	if query.Category != "" {
		baseQuery += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, query.Category)
		argIndex++
	}
	if query.Enabled != nil {
		baseQuery += fmt.Sprintf(" AND enabled = $%d", argIndex)
		args = append(args, *query.Enabled)
		argIndex++
	}
	if query.Keyword != "" {
		baseQuery += fmt.Sprintf(" AND (model ILIKE $%d OR name ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+query.Keyword+"%")
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
		SELECT id, model, name, category, manufacturer, version, enabled,
		       device_count, created_at, updated_at
	` + baseQuery

	// 添加排序
	orderBy := query.OrderBy
	if orderBy == "" {
		orderBy = "created_at"
	}
	orderDir := query.OrderDir
	if orderDir == "" {
		orderDir = "DESC"
	}
	listQuery += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)

	// 添加分页
	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	listQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)

	var results []templateSummaryDB

	err = m.dbConn.QueryRowsCtx(ctx, &results, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询列表失败: %w", err)
	}

	// 4. 转换结果
	templates := make([]*core.TemplateSummary, 0, len(results))
	for _, r := range results {
		t := &core.TemplateSummary{
			ID:           r.ID,
			Model:        r.Model,
			Name:         r.Name,
			Category:     r.Category,
			Manufacturer: r.Manufacturer,
			Version:      r.Version,
			Enabled:      r.Enabled,
			DeviceCount:  r.DeviceCount,
		}

		if r.CreatedAt.Valid {
			t.CreatedAt = r.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		if r.UpdatedAt.Valid {
			t.UpdatedAt = r.UpdatedAt.Time.Format("2006-01-02 15:04:05")
		}

		// 查询并填充标签
		tags, err := m.getTemplateTags(ctx, r.ID)
		if err != nil {
			// 标签查询失败不影响主流程
			tags = []core.DeviceTagSummary{}
		}
		t.Tags = tags

		templates = append(templates, t)
	}

	return templates, total, nil
}
