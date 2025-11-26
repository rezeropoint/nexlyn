package httpreceive

import (
	"context"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/rezeropoint/etcdtrigger/v2/engine"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 注意：time 包仍需保留，用于 List 方法中时间格式化

// httpReceiveManager HTTP数据接收管理器实现
type httpReceiveManager struct {
	dbConn           sqlx.SqlConn
	configStore      engine.Engine
	config           Config
	dispatchInfoFunc core.DispatchInfoFunc
}

// newHttpReceiveManager 创建HTTP数据接收管理器实例
func newHttpReceiveManager(
	dbConn sqlx.SqlConn,
	config Config,
	configStore engine.Engine,
	dispatchInfoFunc core.DispatchInfoFunc,
) (*httpReceiveManager, error) {
	if configStore == nil {
		return nil, fmt.Errorf("configStore 未设置")
	}

	return &httpReceiveManager{
		dbConn:           dbConn,
		configStore:      configStore,
		config:           config,
		dispatchInfoFunc: dispatchInfoFunc,
	}, nil
}

// Create 创建HTTP接收配置（元数据 + 配置），返回生成的UUID
func (m *httpReceiveManager) Create(ctx context.Context, metadata core.HttpReceiveMetadata, config *core.HttpReceiveConfig) (string, error) {
	// 1. 生成UUID作为配置ID
	configID := uuid.New().String()

	// 2. 填充配置的基本信息
	config.ConfigID = configID
	config.TenantID = metadata.TenantID

	// 3. 验证配置
	if err := config.Validate(); err != nil {
		return "", fmt.Errorf("配置验证失败: %w", err)
	}

	// 4. 使用事务插入PostgreSQL并写入Etcd（任意失败都会回滚事务）
	etcdKey := core.BuildHttpReceiveConfigKey(configID)
	err := m.dbConn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 4.1 插入PostgreSQL元数据
		_, err := session.ExecCtx(ctx, insertHttpReceiveSQL,
			configID, metadata.Name, metadata.Description,
			metadata.Enabled, metadata.TenantID, metadata.CreatedBy,
		)
		if err != nil {
			return fmt.Errorf("插入PostgreSQL失败: %w", err)
		}

		// 4.2 写入Etcd配置（失败会触发事务回滚）
		if err := m.configStore.PutConfig(ctx, etcdKey, config); err != nil {
			return fmt.Errorf("保存配置到Etcd失败: %w", err)
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "http_receive_manager"),
		logx.Field("operation", "create"),
		logx.Field("config_id", configID),
		logx.Field("tenant_id", metadata.TenantID),
		logx.Field("status", "success"),
	).Info("创建HTTP接收配置成功")

	return configID, nil
}

// Get 获取完整的HTTP接收配置信息（元数据 + 配置 + 审计信息）
func (m *httpReceiveManager) Get(ctx context.Context, id string, tenantID string) (*core.HttpReceive, error) {
	// 1. 从PostgreSQL查询元数据和审计信息
	var result httpReceiveDB
	err := m.dbConn.QueryRowCtx(ctx, &result, getHttpReceiveSQL, id, tenantID)
	if err != nil {
		return nil, core.ErrNotFound
	}

	// 2. 构建HttpReceive对象
	httpReceive := &core.HttpReceive{
		ID:       result.ID,
		Name:     result.Name,
		Enabled:  result.Enabled,
		TenantID: result.TenantID,
	}

	// 3. 处理可空字段
	if result.Description.Valid {
		httpReceive.Description = result.Description.String
	}
	if result.CreatedBy.Valid {
		httpReceive.CreatedBy = result.CreatedBy.String
	}
	if result.CreatedAt.Valid {
		httpReceive.CreatedAt = result.CreatedAt.Time.Format(time.RFC3339)
	}
	if result.UpdatedAt.Valid {
		httpReceive.UpdatedAt = result.UpdatedAt.Time.Format(time.RFC3339)
	}

	// 4. 从Etcd读取配置
	etcdKey := core.BuildHttpReceiveConfigKey(id)
	var config core.HttpReceiveConfig
	if m.configStore.GetConfig(etcdKey, &config) {
		httpReceive.Config = &config
	}
	// 配置查询失败不影响主流程，Config 字段为 nil

	return httpReceive, nil
}

// List 列出HTTP接收配置（返回摘要信息）
func (m *httpReceiveManager) List(ctx context.Context, query core.HttpReceiveQuery) ([]*core.HttpReceiveSummary, int64, error) {
	// 1. 构建额外条件
	var extraConditions string
	var extraArgs []any
	argIndex := 4 // $1=tenantID, $2=limit, $3=offset

	if query.Keyword != "" {
		extraConditions += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex)
		extraArgs = append(extraArgs, "%"+query.Keyword+"%")
		argIndex++
	}

	// 2. 查询总数
	var total int64
	countQuery := fmt.Sprintf(countHttpReceiveSQL, extraConditions)
	countArgs := append([]any{query.TenantID}, extraArgs...)
	err := m.dbConn.QueryRowCtx(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}

	// 3. 查询列表
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 10
	}
	offset := (query.Page - 1) * query.PageSize

	listQuery := fmt.Sprintf(listHttpReceiveSQL, extraConditions)
	listArgs := append([]any{query.TenantID, query.PageSize, offset}, extraArgs...)

	var results []httpReceiveSummaryDB
	err = m.dbConn.QueryRowsCtx(ctx, &results, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询列表失败: %w", err)
	}

	// 4. 转换结果
	summaries := make([]*core.HttpReceiveSummary, 0, len(results))
	for _, r := range results {
		summary := &core.HttpReceiveSummary{
			ID:      r.ID,
			Name:    r.Name,
			Enabled: r.Enabled,
		}
		// 处理可空字段
		if r.Description.Valid {
			summary.Description = r.Description.String
		}
		if r.CreatedAt.Valid {
			summary.CreatedAt = r.CreatedAt.Time.Format(time.RFC3339)
		}
		if r.UpdatedAt.Valid {
			summary.UpdatedAt = r.UpdatedAt.Time.Format(time.RFC3339)
		}
		summaries = append(summaries, summary)
	}

	return summaries, total, nil
}

// Update 更新HTTP接收配置
func (m *httpReceiveManager) Update(ctx context.Context, id string, tenantID string, metadata core.HttpReceiveMetadata, config *core.HttpReceiveConfig) error {
	// 1. 填充配置的基本信息
	config.ConfigID = id
	config.TenantID = tenantID

	// 2. 验证配置
	if err := config.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 3. 使用事务更新PostgreSQL并写入Etcd（任意失败都会回滚事务）
	etcdKey := core.BuildHttpReceiveConfigKey(id)
	err := m.dbConn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 3.1 验证权限
		var exists int
		err := session.QueryRowCtx(ctx, &exists, checkHttpReceiveExistsSQL, id, tenantID)
		if err != nil {
			return core.ErrNotFound
		}

		// 3.2 更新PostgreSQL元数据
		_, err = session.ExecCtx(ctx, updateHttpReceiveSQL,
			metadata.Name, metadata.Description, metadata.Enabled,
			id, tenantID,
		)
		if err != nil {
			return fmt.Errorf("更新PostgreSQL失败: %w", err)
		}

		// 3.3 更新Etcd配置（失败会触发事务回滚）
		if err := m.configStore.PutConfig(ctx, etcdKey, config); err != nil {
			return fmt.Errorf("更新Etcd配置失败: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "http_receive_manager"),
		logx.Field("operation", "update"),
		logx.Field("config_id", id),
		logx.Field("tenant_id", tenantID),
		logx.Field("status", "success"),
	).Info("更新HTTP接收配置成功")

	return nil
}

// Delete 删除HTTP接收配置
func (m *httpReceiveManager) Delete(ctx context.Context, id string, tenantID string) error {
	// 1. 使用事务删除PostgreSQL并清理Etcd（任意失败都会回滚事务）
	etcdKey := core.BuildHttpReceiveConfigKey(id)
	err := m.dbConn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1.1 验证权限
		var exists int
		err := session.QueryRowCtx(ctx, &exists, checkHttpReceiveExistsSQL, id, tenantID)
		if err != nil {
			return core.ErrNotFound
		}

		// 1.2 删除PostgreSQL记录
		_, err = session.ExecCtx(ctx, deleteHttpReceiveSQL, id, tenantID)
		if err != nil {
			return fmt.Errorf("删除PostgreSQL记录失败: %w", err)
		}

		// 1.3 删除Etcd配置（失败会触发事务回滚）
		if err := m.configStore.DeleteConfig(ctx, etcdKey); err != nil {
			return fmt.Errorf("删除Etcd配置失败: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "http_receive_manager"),
		logx.Field("operation", "delete"),
		logx.Field("config_id", id),
		logx.Field("tenant_id", tenantID),
		logx.Field("status", "success"),
	).Info("删除HTTP接收配置成功")

	return nil
}

// ProcessData 处理HTTP接收的数据
func (m *httpReceiveManager) ProcessData(ctx context.Context, configId string, data map[string]any) error {
	// 1. 从Etcd获取配置
	etcdKey := core.BuildHttpReceiveConfigKey(configId)
	var config core.HttpReceiveConfig
	if !m.configStore.GetConfig(etcdKey, &config) {
		return fmt.Errorf("配置不存在: %s", configId)
	}

	// 2. 从数据库检查配置是否启用
	var enabled bool
	checkQuery := "SELECT enabled FROM iot_http_receive_configs WHERE id = $1"
	err := m.dbConn.QueryRowCtx(ctx, &enabled, checkQuery, configId)
	if err != nil {
		return fmt.Errorf("查询配置状态失败: %w", err)
	}
	if !enabled {
		return fmt.Errorf("配置已禁用: %s", configId)
	}

	// 3. 提取设备ID（从请求体）
	var deviceID string
	if config.DeviceIDPath != "" {
		if value, exists := getFieldValue(data, config.DeviceIDPath); exists {
			deviceID = normalizeValue(value)
		}
	}
	if deviceID == "" {
		deviceID = "unknown"
	}

	// 4. 解析时间戳
	timestamp := m.parseTimestamp(ctx, data, config.TimestampPath, config.TimestampFormat)

	// 5. 字段映射提取
	extractedFields := m.extractFieldsByMapping(ctx, data, config.FieldMappings)

	// 6. 转换为 TypedValue 格式
	typedData := m.convertToTypedValue(extractedFields)

	// 7. 数据分发
	if len(config.DispatchConfigs) > 0 && m.dispatchInfoFunc != nil {
		taskInfo := core.TaskInfo{
			TaskId:     fmt.Sprintf("http/%s/%s-%d", configId, deviceID, timestamp),
			ConfigType: "http_receive",
			TenantID:   config.TenantID,
			DeviceID:   deviceID,
		}
		if err := m.dispatchInfoFunc(ctx, config.DispatchConfigs, &typedData, taskInfo); err != nil {
			logx.WithContext(ctx).WithFields(
				logx.Field("service", m.config.ServiceName),
				logx.Field("pod", m.config.PodName),
				logx.Field("module", "http_receive"),
				logx.Field("operation", "dispatch"),
				logx.Field("config_id", configId),
				logx.Field("device_id", deviceID),
				logx.Field("error", err.Error()),
			).Error("数据分发失败")
			// 分发失败不阻塞返回，继续处理
		}
	}

	logx.WithContext(ctx).WithFields(
		logx.Field("service", m.config.ServiceName),
		logx.Field("pod", m.config.PodName),
		logx.Field("module", "http_receive"),
		logx.Field("operation", "process_data"),
		logx.Field("config_id", configId),
		logx.Field("device_id", deviceID),
		logx.Field("status", "success"),
	).Info("HTTP数据处理成功")

	return nil
}

// Close 关闭管理器
func (m *httpReceiveManager) Close() error {
	return nil
}
