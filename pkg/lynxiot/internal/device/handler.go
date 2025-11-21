package device

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// deviceManager 设备管理器实现
type deviceManager struct {
	config      Config
	dbConn      sqlx.SqlConn
	redisClient *redis.Redis

	// 设备在线状态缓存函数（由MQTT Manager提供）
	getDeviceOnlineStatus core.GetDeviceOnlineStatusFunc
	getAllOnlineDevices   core.GetAllOnlineDevicesFunc
}

// newDeviceManager 创建设备管理器实例
func newDeviceManager(dbConn sqlx.SqlConn, redisClient *redis.Redis, config Config, getDeviceOnlineStatus core.GetDeviceOnlineStatusFunc, getAllOnlineDevices core.GetAllOnlineDevicesFunc) (*deviceManager, error) {

	if dbConn == nil {
		return nil, fmt.Errorf("dbConn 未设置")
	}

	if redisClient == nil {
		return nil, fmt.Errorf("redisClient 未设置")
	}

	// 验证在线状态缓存函数（必需）
	if getDeviceOnlineStatus == nil {
		return nil, fmt.Errorf("getDeviceOnlineStatus 函数未设置")
	}
	if getAllOnlineDevices == nil {
		return nil, fmt.Errorf("getAllOnlineDevices 函数未设置")
	}

	return &deviceManager{
		config:                config,
		dbConn:                dbConn,
		redisClient:           redisClient,
		getDeviceOnlineStatus: getDeviceOnlineStatus,
		getAllOnlineDevices:   getAllOnlineDevices,
	}, nil
}

// Bind 绑定设备
func (m *deviceManager) Bind(ctx context.Context, metadata core.DeviceBindingMetadata) (string, error) {
	// 1. 验证元数据
	if err := metadata.Validate(); err != nil {
		return "", fmt.Errorf("元数据验证失败: %w", err)
	}

	// 2. 检查设备是否已绑定（同一租户和组织下）
	var count int
	checkQuery := `
		SELECT COUNT(*) FROM iot_device_bindings
		WHERE device_id = $1 AND tenant_id = $2 AND org_id = $3
	`
	err := m.dbConn.QueryRowCtx(ctx, &count, checkQuery, metadata.DeviceID, metadata.TenantID, metadata.OrgID)
	if err != nil {
		return "", fmt.Errorf("检查设备是否已绑定失败: %w", err)
	}
	if count > 0 {
		return "", core.ErrDeviceAlreadyBound
	}

	// 3. 验证设备型号模板是否存在，并获取设备类别
	var templateInfo templateInfoDB
	templateQuery := `
		SELECT category, tenant_id FROM iot_sensor_templates
		WHERE model = $1
	`
	err = m.dbConn.QueryRowCtx(ctx, &templateInfo, templateQuery, metadata.DeviceModel)
	if err == sql.ErrNoRows {
		return "", core.ErrDeviceModelNotFound
	}
	if err != nil {
		return "", fmt.Errorf("查询设备模板失败: %w", err)
	}

	// 验证模板是否属于同一租户
	if templateInfo.TenantID != metadata.TenantID {
		return "", fmt.Errorf("设备型号模板不属于当前租户")
	}

	// 验证设备类别与模板类别是否匹配（从Redis获取设备实际类别）
	if m.redisClient != nil && templateInfo.Category.Valid {
		redisKey := core.BuildDeviceOnlineRedisKey(metadata.DeviceID)
		value, err := m.redisClient.Get(redisKey)
		if err == nil && value != "" {
			var onlineData core.DeviceOnlineRedisData
			if err := json.Unmarshal([]byte(value), &onlineData); err == nil && onlineData.Category != "" {
				// 验证设备类别与模板类别是否一致
				if onlineData.Category != templateInfo.Category.String {
					return "", fmt.Errorf("设备类别(%s)与所选型号的类别(%s)不匹配，请选择正确的设备型号",
						onlineData.Category, templateInfo.Category.String)
				}
			}
		}
	}

	// 4. 生成UUID
	id := uuid.New().String()

	// 5. 设置默认值
	status := metadata.Status
	if status == "" {
		status = "active"
	}

	// 6. 插入PostgreSQL
	insertQuery := `
		INSERT INTO iot_device_bindings
		(id, device_id, device_name, device_alias, device_model, device_category,
		 description, location, installation_date, status, is_online,
		 tenant_id, org_id, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, FALSE, $11, $12, $13)
	`
	_, err = m.dbConn.ExecCtx(ctx, insertQuery,
		id, metadata.DeviceID, metadata.DeviceName, metadata.DeviceAlias, metadata.DeviceModel,
		templateInfo.Category, metadata.Description, metadata.Location, toNullString(metadata.InstallationDate),
		status, metadata.TenantID, metadata.OrgID, metadata.CreatedBy,
	)
	if err != nil {
		return "", fmt.Errorf("插入设备绑定到PostgreSQL失败: %w", err)
	}

	// 7. 处理标签关联
	if len(metadata.TagIDs) > 0 {
		err = m.replaceDeviceTags(ctx, metadata.DeviceID, metadata.TagIDs, metadata.TenantID)
		if err != nil {
			// 标签关联失败不回滚，只返回错误
			return id, fmt.Errorf("创建标签关联失败: %w", err)
		}
	}

	return id, nil
}

// Get 获取设备详情
func (m *deviceManager) Get(ctx context.Context, deviceID string, tenantID string, orgIDs []string) (*core.DeviceBinding, error) {
	// 从PostgreSQL查询设备详情
	var result deviceBindingDB

	query := `
		SELECT id, device_id, device_name, device_alias, device_model, device_category,
		       description, location, installation_date, status, is_online, last_data_at,
		       tenant_id, org_id, created_by, updated_by,
		       created_at, updated_at
		FROM iot_device_bindings
		WHERE device_id = $1 AND tenant_id = $2 AND org_id = ANY($3)
	`
	err := m.dbConn.QueryRowCtx(ctx, &result, query, deviceID, tenantID, pq.Array(orgIDs))
	if err == sql.ErrNoRows {
		return nil, core.ErrDeviceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}

	// 构建设备对象
	device := &core.DeviceBinding{
		ID:               result.ID,
		DeviceID:         result.DeviceID,
		DeviceName:       result.DeviceName,
		DeviceAlias:      result.DeviceAlias,
		DeviceModel:      result.DeviceModel,
		DeviceCategory:   nullStringToString(result.DeviceCategory),
		Description:      nullStringToString(result.Description),
		Location:         nullStringToString(result.Location),
		InstallationDate: nullStringToString(result.InstallationDate),
		Status:           result.Status,
		IsOnline:         result.IsOnline,
		LastDataAt:       nullTimeToString(result.LastDataAt),
		TenantID:         result.TenantID,
		OrgID:            result.OrgID,
		CreatedBy:        nullStringToString(result.CreatedBy),
		UpdatedBy:        nullStringToString(result.UpdatedBy),
		CreatedAt:        nullTimeToString(result.CreatedAt),
		UpdatedAt:        nullTimeToString(result.UpdatedAt),
	}

	// 查询并填充标签
	tags, err := m.getDeviceTags(ctx, deviceID)
	if err != nil {
		// 标签查询失败不影响主流程
		tags = []core.DeviceTagSummary{}
	}
	device.Tags = tags

	// 从Redis获取实时在线状态（覆盖数据库中的is_online字段）
	device.IsOnline = m.getDeviceOnlineStatus(deviceID)

	return device, nil
}

// List 查询设备列表
func (m *deviceManager) List(ctx context.Context, query core.DeviceBindingQuery) ([]*core.DeviceBindingSummary, int64, error) {
	// 1. 构建SQL查询
	baseQuery := `
		FROM iot_device_bindings
		WHERE tenant_id = $1
	`
	var args []any
	args = append(args, query.TenantID)
	argIndex := 2

	// 添加过滤条件
	if query.OrgID != "" {
		baseQuery += fmt.Sprintf(" AND org_id = $%d", argIndex)
		args = append(args, query.OrgID)
		argIndex++
	}
	if query.DeviceModel != "" {
		baseQuery += fmt.Sprintf(" AND device_model = $%d", argIndex)
		args = append(args, query.DeviceModel)
		argIndex++
	}
	if query.DeviceCategory != "" {
		baseQuery += fmt.Sprintf(" AND device_category = $%d", argIndex)
		args = append(args, query.DeviceCategory)
		argIndex++
	}
	if query.Status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, query.Status)
		argIndex++
	}
	if query.IsOnline != nil {
		baseQuery += fmt.Sprintf(" AND is_online = $%d", argIndex)
		args = append(args, *query.IsOnline)
		argIndex++
	}
	if query.Keyword != "" {
		baseQuery += fmt.Sprintf(" AND (device_id ILIKE $%d OR device_name ILIKE $%d OR device_alias ILIKE $%d)", argIndex, argIndex, argIndex)
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

	// 3. 查询列表（关联organizations表获取org_name）
	listQuery := `
		SELECT d.id, d.device_id, d.device_name, d.device_alias, d.device_model, d.device_category,
		       d.location, d.status, d.is_online, d.last_data_at, d.org_id,
		       COALESCE(o.name, '') as org_name, d.created_at
		FROM iot_device_bindings d
		LEFT JOIN system_organizations o ON d.org_id = o.id AND o.deleted_at IS NULL
	`
	// 修改baseQuery的FROM部分
	baseQueryWithoutFrom := `
		WHERE d.tenant_id = $1
	`
	var argsForList []any
	argsForList = append(argsForList, query.TenantID)
	argIndexForList := 2

	// 重新构建过滤条件（使用d.前缀）
	if query.OrgID != "" {
		baseQueryWithoutFrom += fmt.Sprintf(" AND d.org_id = $%d", argIndexForList)
		argsForList = append(argsForList, query.OrgID)
		argIndexForList++
	}
	if query.DeviceModel != "" {
		baseQueryWithoutFrom += fmt.Sprintf(" AND d.device_model = $%d", argIndexForList)
		argsForList = append(argsForList, query.DeviceModel)
		argIndexForList++
	}
	if query.DeviceCategory != "" {
		baseQueryWithoutFrom += fmt.Sprintf(" AND d.device_category = $%d", argIndexForList)
		argsForList = append(argsForList, query.DeviceCategory)
		argIndexForList++
	}
	if query.Status != "" {
		baseQueryWithoutFrom += fmt.Sprintf(" AND d.status = $%d", argIndexForList)
		argsForList = append(argsForList, query.Status)
		argIndexForList++
	}
	if query.IsOnline != nil {
		baseQueryWithoutFrom += fmt.Sprintf(" AND d.is_online = $%d", argIndexForList)
		argsForList = append(argsForList, *query.IsOnline)
		argIndexForList++
	}
	if query.Keyword != "" {
		baseQueryWithoutFrom += fmt.Sprintf(" AND (d.device_id ILIKE $%d OR d.device_name ILIKE $%d OR d.device_alias ILIKE $%d)", argIndexForList, argIndexForList, argIndexForList)
		argsForList = append(argsForList, "%"+query.Keyword+"%")
		argIndexForList++
	}

	listQuery += baseQueryWithoutFrom

	// 添加排序（添加d.前缀）
	orderBy := query.OrderBy
	if orderBy == "" {
		orderBy = "d.created_at"
	} else {
		// 如果orderBy不包含表前缀，添加d.前缀
		if orderBy != "" && orderBy[0] != 'd' {
			orderBy = "d." + orderBy
		}
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

	var results []deviceBindingSummaryDB

	err = m.dbConn.QueryRowsCtx(ctx, &results, listQuery, argsForList...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询列表失败: %w", err)
	}

	// 4. 转换结果
	devices := make([]*core.DeviceBindingSummary, 0, len(results))
	for _, r := range results {
		d := &core.DeviceBindingSummary{
			ID:             r.ID,
			DeviceID:       r.DeviceID,
			DeviceName:     r.DeviceName,
			DeviceAlias:    r.DeviceAlias,
			DeviceModel:    r.DeviceModel,
			DeviceCategory: nullStringToString(r.DeviceCategory),
			Location:       nullStringToString(r.Location),
			Status:         r.Status,
			IsOnline:       r.IsOnline,
			LastDataAt:     nullTimeToString(r.LastDataAt),
			OrgID:          r.OrgID,
			OrgName:        r.OrgName,
			CreatedAt:      nullTimeToString(r.CreatedAt),
		}

		// 查询并填充标签
		tags, err := m.getDeviceTags(ctx, r.DeviceID)
		if err != nil {
			// 标签查询失败不影响主流程
			tags = []core.DeviceTagSummary{}
		}
		d.Tags = tags

		// 从Redis获取实时在线状态（覆盖数据库中的is_online字段）
		d.IsOnline = m.getDeviceOnlineStatus(r.DeviceID)

		devices = append(devices, d)
	}

	return devices, total, nil
}

// Update 更新设备信息
func (m *deviceManager) Update(ctx context.Context, deviceID string, tenantID string, orgIDs []string, update core.DeviceBindingUpdate) error {
	// 1. 检查设备是否存在并验证租户和组织权限
	var existingTenantID string
	checkQuery := "SELECT tenant_id FROM iot_device_bindings WHERE device_id = $1 AND tenant_id = $2 AND org_id = ANY($3)"
	err := m.dbConn.QueryRowCtx(ctx, &existingTenantID, checkQuery, deviceID, tenantID, pq.Array(orgIDs))
	if err == sql.ErrNoRows {
		return core.ErrDeviceNotFound
	}
	if err != nil {
		return fmt.Errorf("查询设备失败: %w", err)
	}

	// 2. 构建更新SQL（只更新非空字段）
	updateQuery := "UPDATE iot_device_bindings SET updated_at = CURRENT_TIMESTAMP"
	var args []any
	argIndex := 1

	if update.DeviceName != "" {
		updateQuery += fmt.Sprintf(", device_name = $%d", argIndex)
		args = append(args, update.DeviceName)
		argIndex++
	}
	if update.DeviceAlias != "" {
		updateQuery += fmt.Sprintf(", device_alias = $%d", argIndex)
		args = append(args, update.DeviceAlias)
		argIndex++
	}
	if update.Description != "" {
		updateQuery += fmt.Sprintf(", description = $%d", argIndex)
		args = append(args, update.Description)
		argIndex++
	}
	if update.Location != "" {
		updateQuery += fmt.Sprintf(", location = $%d", argIndex)
		args = append(args, update.Location)
		argIndex++
	}
	if update.InstallationDate != "" {
		updateQuery += fmt.Sprintf(", installation_date = $%d", argIndex)
		args = append(args, update.InstallationDate)
		argIndex++
	}
	if update.Status != "" {
		updateQuery += fmt.Sprintf(", status = $%d", argIndex)
		args = append(args, update.Status)
		argIndex++
	}

	updateQuery += fmt.Sprintf(" WHERE device_id = $%d AND tenant_id = $%d", argIndex, argIndex+1)
	args = append(args, deviceID, tenantID)

	// 3. 执行更新
	_, err = m.dbConn.ExecCtx(ctx, updateQuery, args...)
	if err != nil {
		return fmt.Errorf("更新设备失败: %w", err)
	}

	// 4. 处理标签关联（如果提供了TagIDs则替换，nil则不修改）
	if update.TagIDs != nil {
		err = m.replaceDeviceTags(ctx, deviceID, update.TagIDs, tenantID)
		if err != nil {
			return fmt.Errorf("更新标签关联失败: %w", err)
		}
	}

	return nil
}

// Delete 删除设备（物理删除，解绑即删除）
func (m *deviceManager) Delete(ctx context.Context, deviceID string, tenantID string, orgIDs []string) error {
	// 1. 检查设备是否存在并验证租户和组织权限
	var existingTenantID string
	checkQuery := "SELECT tenant_id FROM iot_device_bindings WHERE device_id = $1 AND tenant_id = $2 AND org_id = ANY($3)"
	err := m.dbConn.QueryRowCtx(ctx, &existingTenantID, checkQuery, deviceID, tenantID, pq.Array(orgIDs))
	if err == sql.ErrNoRows {
		return core.ErrDeviceNotFound
	}
	if err != nil {
		return fmt.Errorf("查询设备失败: %w", err)
	}

	// 2. 物理删除
	deleteQuery := "DELETE FROM iot_device_bindings WHERE device_id = $1 AND tenant_id = $2"
	_, err = m.dbConn.ExecCtx(ctx, deleteQuery, deviceID, tenantID)
	if err != nil {
		return fmt.Errorf("删除设备失败: %w", err)
	}

	return nil
}

// ListUnboundDevices 获取未绑定设备列表（当前在线但未绑定的设备）
func (m *deviceManager) ListUnboundDevices(ctx context.Context, tenantID string) ([]core.UnboundDevice, error) {
	// 1. 从Redis获取所有在线设备
	allOnlineDevices, err := m.getAllOnlineDevices()
	if err != nil {
		return nil, fmt.Errorf("获取在线设备列表失败: %w", err)
	}

	// 3. 获取该租户所有已绑定的设备ID（不限制组织）
	boundDeviceIDs, err := m.listBoundDeviceIDsForTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("获取已绑定设备ID列表失败: %w", err)
	}

	// 4. 创建已绑定设备ID的map（用于快速查找）
	boundMap := make(map[string]bool, len(boundDeviceIDs))
	for _, deviceID := range boundDeviceIDs {
		boundMap[deviceID] = true
	}

	// 5. 过滤未绑定设备（差集运算）
	unboundDevices := make([]core.UnboundDevice, 0)
	for _, device := range allOnlineDevices {
		// 如果设备未绑定，添加到结果列表
		if !boundMap[device.DeviceID] {
			unboundDevices = append(unboundDevices, device)
		}
	}

	return unboundDevices, nil
}

// GetDeviceInfoForController 获取设备基本信息（用于Controller Manager）
// 只查询设备的category、model、tenant_id等基本信息，不进行组织权限验证
// 注意：权限验证应在Engine层的AI Box控制方法中统一处理
func (m *deviceManager) GetDeviceInfoForController(ctx context.Context, deviceID string) (*core.DeviceBinding, error) {
	var result deviceInfoForControllerDB

	query := `
		SELECT device_id, device_model, device_category, tenant_id
		FROM iot_device_bindings
		WHERE device_id = $1
	`
	err := m.dbConn.QueryRowCtx(ctx, &result, query, deviceID)
	if err == sql.ErrNoRows {
		return nil, core.ErrDeviceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询设备信息失败: %w", err)
	}

	// 查询Redis实时在线状态
	isOnline := m.getDeviceOnlineStatus(result.DeviceID)

	// 转换为 DeviceBinding 结构（包含在线状态）
	return &core.DeviceBinding{
		DeviceID:       result.DeviceID,
		DeviceModel:    result.DeviceModel,
		DeviceCategory: result.DeviceCategory,
		TenantID:       result.TenantID,
		IsOnline:       isOnline,
	}, nil
}
