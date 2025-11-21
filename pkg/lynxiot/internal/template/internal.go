package template

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// replaceTemplateTags 完全替换模板的标签关联（在事务中执行）
// 参数session: 事务会话，用于在事务中执行数据库操作
func (m *templateManager) replaceTemplateTags(ctx context.Context, session sqlx.Session, templateID string, tagIDs []string, tenantID string) error {
	// 1. 删除现有关联
	deleteQuery := "DELETE FROM iot_sensor_template_tag_relations WHERE template_id = $1"
	_, err := session.ExecCtx(ctx, deleteQuery, templateID)
	if err != nil {
		return fmt.Errorf("删除现有标签关联失败: %w", err)
	}

	// 2. 如果tagIDs为空，则只删除不添加
	if len(tagIDs) == 0 {
		return nil
	}

	// 3. 为每个标签创建新关联
	for _, tagID := range tagIDs {
		// 验证标签是否存在且属于该租户
		var count int
		checkQuery := "SELECT COUNT(*) FROM iot_tag_definitions WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL"
		err = session.QueryRowCtx(ctx, &count, checkQuery, tagID, tenantID)
		if err != nil {
			return fmt.Errorf("查询标签失败: %w", err)
		}
		if count == 0 {
			return fmt.Errorf("标签 %s 不存在或无权限访问", tagID)
		}

		// 创建关联
		insertQuery := "INSERT INTO iot_sensor_template_tag_relations (id, template_id, tag_id) VALUES ($1, $2, $3)"
		_, err = session.ExecCtx(ctx, insertQuery, uuid.New().String(), templateID, tagID)
		if err != nil {
			return fmt.Errorf("创建标签关联失败: %w", err)
		}
	}

	return nil
}

// getTemplateTags 获取模板的所有标签
func (m *templateManager) getTemplateTags(ctx context.Context, templateID string) ([]core.DeviceTagSummary, error) {
	query := `
		SELECT t.id, t.label, t.description, t.color, t.created_at
		FROM iot_tag_definitions t
		INNER JOIN iot_sensor_template_tag_relations r ON t.id = r.tag_id
		WHERE r.template_id = $1 AND t.deleted_at IS NULL
		ORDER BY r.created_at DESC
	`

	var results []templateTagDB

	err := m.dbConn.QueryRowsCtx(ctx, &results, query, templateID)
	if err != nil {
		return nil, fmt.Errorf("查询模板标签失败: %w", err)
	}

	tags := make([]core.DeviceTagSummary, 0, len(results))
	for _, r := range results {
		tag := core.DeviceTagSummary{
			ID:    r.ID,
			Label: r.Label,
		}
		if r.Description.Valid {
			tag.Description = r.Description.String
		}
		if r.Color.Valid {
			tag.Color = r.Color.String
		}
		if r.CreatedAt.Valid {
			tag.CreatedAt = r.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// saveOnlineDetectionConfig 保存在线检测配置到Etcd
func (m *templateManager) saveOnlineDetectionConfig(ctx context.Context, config *core.OnlineDetectionConfig) error {
	// 验证配置
	if err := config.Validate(); err != nil {
		return fmt.Errorf("在线检测配置验证失败: %w", err)
	}

	// 构建Etcd key
	key := core.BuildOnlineDetectionKey(config.Category, config.Model, config.TopicSuffix)

	// 写入Etcd
	err := m.configStore.PutConfig(ctx, key, config)
	if err != nil {
		return fmt.Errorf("写入在线检测配置到Etcd失败: %w", err)
	}

	return nil
}

// saveDataProcessingConfig 保存业务数据处理配置到Etcd
func (m *templateManager) saveDataProcessingConfig(ctx context.Context, config *core.DataProcessingConfig) error {
	// 验证配置
	if err := config.Validate(); err != nil {
		return fmt.Errorf("业务数据配置验证失败: %w", err)
	}

	// 构建Etcd key
	key := core.BuildDataProcessingKey(config.Category, config.Model, config.TopicSuffix)

	// 写入Etcd
	err := m.configStore.PutConfig(ctx, key, config)
	if err != nil {
		return fmt.Errorf("写入业务数据配置到Etcd失败: %w", err)
	}

	return nil
}

// getOnlineDetectionConfig 获取在线检测配置
func (m *templateManager) getOnlineDetectionConfig(category, model, topicSuffix string) (*core.OnlineDetectionConfig, bool) {
	// 构建Etcd key
	key := core.BuildOnlineDetectionKey(category, model, topicSuffix)

	// 从Etcd读取配置
	var config core.OnlineDetectionConfig
	found := m.configStore.GetConfig(key, &config)
	if !found {
		return nil, false // 配置不存在
	}

	return &config, true
}

// getDataProcessingConfig 获取业务数据处理配置
func (m *templateManager) getDataProcessingConfig(category, model, topicSuffix string) (*core.DataProcessingConfig, bool) {
	// 构建Etcd key
	key := core.BuildDataProcessingKey(category, model, topicSuffix)

	// 从Etcd读取配置
	var config core.DataProcessingConfig
	found := m.configStore.GetConfig(key, &config)
	if !found {
		return nil, false // 配置不存在
	}

	return &config, true
}

// extractPlatformIDs 从业务数据配置中提取所有使用的平台ID
// 用于在数据库中记录平台配置使用情况，以便删除时检查约束
func extractPlatformIDs(businessConfig *core.DataProcessingConfig) []string {
	if businessConfig == nil || len(businessConfig.DispatchConfigs) == 0 {
		return []string{}
	}

	// 使用 map 去重
	platformIDMap := make(map[string]bool)
	for _, dispatchConfig := range businessConfig.DispatchConfigs {
		if dispatchConfig.PlatformID != "" {
			platformIDMap[dispatchConfig.PlatformID] = true
		}
	}

	// 转换为切片
	platformIDs := make([]string, 0, len(platformIDMap))
	for platformID := range platformIDMap {
		platformIDs = append(platformIDs, platformID)
	}

	return platformIDs
}

// saveControlConfig 保存控制配置到Etcd
func (m *templateManager) saveControlConfig(ctx context.Context, config *core.DeviceControlConfig) error {
	// 验证配置
	if err := config.Validate(); err != nil {
		return fmt.Errorf("控制配置验证失败: %w", err)
	}

	// 构建Etcd key（使用CommandSuffix作为主键）
	key := core.BuildControlConfigKey(config.Category, config.Model, config.CommandSuffix)

	// 写入Etcd
	err := m.configStore.PutConfig(ctx, key, config)
	if err != nil {
		return fmt.Errorf("写入控制配置到Etcd失败: %w", err)
	}

	return nil
}

// getControlConfig 获取控制配置
func (m *templateManager) getControlConfig(category, model, suffix string) (*core.DeviceControlConfig, bool) {
	// 构建Etcd key
	key := core.BuildControlConfigKey(category, model, suffix)

	// 从Etcd读取配置
	var config core.DeviceControlConfig
	found := m.configStore.GetConfig(key, &config)
	if !found {
		return nil, false // 配置不存在
	}

	return &config, true
}
