package infoatom

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/monc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/syncx"
)

var (
	// can't use one SingleFlight per conn, because multiple conns may share the same cache key.
	singleFlight = syncx.NewSingleFlight()
	stats        = cache.NewStat("infoAtomRegistry")
)

// InfoAtomRegistry 实现 InfoAtom 接口
type infoAtomRegistry struct {
	config          *Config
	sqlDB           sqlx.SqlConn
	CacheInterface  cache.Cache
	getTagNamesFunc core.GetTagNamesByIDsFunc
}

// newInfoAtomRegistry 创建一个新的信息原子注册表
// 参数：
//   - sqlConn: 已建立的 PostgreSQL 连接（外部注入）
//   - getTagNamesFunc: 根据标签ID获取标签名称的函数（用于标签查询）
func newInfoAtomRegistry(config *Config, sqlConn sqlx.SqlConn, getTagNamesFunc core.GetTagNamesByIDsFunc) InfoAtomRegistry {
	// 初始化缓存接口（使用配置中的 CacheConf，包内创建）
	cacheInterface := cache.New(config.CacheConf, singleFlight, stats, monc.ErrNotFound)

	return &infoAtomRegistry{
		config:          config,
		sqlDB:           sqlConn,
		CacheInterface:  cacheInterface,
		getTagNamesFunc: getTagNamesFunc,
	}
}

// InitTable 初始化数据库表
func (r *infoAtomRegistry) InitInfoAtomTable(ctx context.Context) error {
	// 先检查表是否已存在
	var count int
	err := r.sqlDB.QueryRowCtx(ctx, &count, CheckTableExistsSQL)
	if err != nil {
		return fmt.Errorf("%w: 检查表存在性失败: %v", ErrSQLConnectionFailed, err)
	}

	// 如果表已存在，直接返回
	if count > 0 {
		return nil
	}

	// 表不存在，执行创建
	_, err = r.sqlDB.ExecCtx(ctx, CreateTableSQL)
	if err != nil {
		return fmt.Errorf("%w: 创建表失败: %v", ErrSQLConnectionFailed, err)
	}
	return nil
}

// CheckInfoAtom 检查信息原子类型是否存在
func (r *infoAtomRegistry) CheckInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey) bool {
	query := fmt.Sprintf(`SELECT COUNT(1) FROM %s WHERE id = $1`, TableName)

	var count int
	err := r.sqlDB.QueryRowCtx(ctx, &count, query, key.ID)

	// 如果没有错误且count > 0，表示找到了，返回true
	return err == nil && count > 0
}

// CreateInfoAtomType 创建新的信息原子类型
func (r *infoAtomRegistry) CreateInfoAtomType(ctx context.Context, infoAtomType core.InfoAtomType) error {
	// 检查参数
	if infoAtomType == nil {
		return fmt.Errorf("%w: 信息原子类型不能为空", ErrInvalidInfoAtomTypeSpec)
	}

	if r.config.RunMode != core.Manager {
		return fmt.Errorf("%w: %s", ErrNotManagerMode, r.config.RunMode)
	}

	// 检查必填字段
	if infoAtomType.GetName() == "" {
		return ErrInfoAtomTypeIDEmpty
	}

	// 构造键
	key := core.InfoAtomTypeKey{
		ID: infoAtomType.GetID(),
		// TenantId: infoAtomType.GetTenantId(),
		// Name:      infoAtomType.GetName(),
		// Version:   infoAtomType.GetVersion(),
	}

	// 检查是否已存在
	if r.CheckInfoAtomType(ctx, key) {
		return fmt.Errorf("%w: id=%s",
			ErrInfoAtomTypeAlreadyExists,
			key.ID)
	}

	// 转换为基本类型进行存储
	basicType, ok := infoAtomType.(*core.BasicInfoAtomType)
	if !ok {
		// 如果不是基本类型，则创建一个
		basicType = &core.BasicInfoAtomType{
			TenantId:   infoAtomType.GetTenantId(),
			Name:       infoAtomType.GetName(),
			Version:    infoAtomType.GetVersion(),
			DataFormat: infoAtomType.GetDataFormat(),
			CreatedBy:  infoAtomType.GetCreatedBy(), // 保留创建人信息
		}
	}

	// 序列化数据格式
	dataFormatBytes, err := json.Marshal(basicType.DataFormat)
	if err != nil {
		return fmt.Errorf("%w: 数据格式序列化失败: %v", ErrInvalidInfoAtomTypeSpec, err)
	}

	// 插入数据库（ID使用UUID，created_by和时间戳字段由数据库自动生成）
	var result struct {
		ID        string `db:"id"`
		CreatedAt int64  `db:"created_at"`
		UpdatedAt int64  `db:"updated_at"`
	}
	err = r.sqlDB.QueryRowCtx(ctx, &result,
		fmt.Sprintf(`INSERT INTO %s (id, tenant_id, name, version, tag_ids, data_format, created_by)
		 VALUES (COALESCE(NULLIF($1, ''), gen_random_uuid()::TEXT)::UUID, $2, $3, $4, $5, $6, $7)
		 RETURNING id::TEXT, EXTRACT(epoch FROM created_at)::BIGINT as created_at, EXTRACT(epoch FROM updated_at)::BIGINT as updated_at`, TableName),
		basicType.ID, basicType.TenantId, basicType.Name, basicType.Version, pq.Array(basicType.TagIDs), string(dataFormatBytes), basicType.CreatedBy)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSQLInsertFailed, err)
	}

	// 更新结构体中的ID和时间戳
	basicType.ID = result.ID
	basicType.CreatedAt = result.CreatedAt
	basicType.UpdatedAt = result.UpdatedAt

	return nil
}

func (r *infoAtomRegistry) GetInfoAtomTypeList(ctx context.Context, ListParams core.InfoAtomList) ([]core.InfoAtomType, int64, int64, int64, error) {
	// 设置默认值
	page := ListParams.Page
	if page <= 0 {
		page = 1
	}
	pageSize := ListParams.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // 限制最大页大小
	}

	// 构建查询条件
	var conditions []string
	var args []interface{}

	if ListParams.TenantId != "" {
		conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", len(args)+1))
		args = append(args, ListParams.TenantId)

	}

	if ListParams.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name LIKE $%d", len(args)+1))
		args = append(args, "%"+ListParams.Name+"%")

	}

	// 添加tags筛选逻辑
	if len(ListParams.Tags) > 0 {
		conditions = append(conditions, fmt.Sprintf("tags && $%d", len(args)+1))
		args = append(args, pq.Array(ListParams.Tags))

	}

	var whereClause string
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", TableName, whereClause)
	var total int64
	err := r.sqlDB.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("%w: %v", ErrSQLQueryFailed, err)
	}

	// 计算总页数
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	// 查询数据
	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf(
		"SELECT id::TEXT, tenant_id::TEXT, name, version, tag_ids, data_format, created_by, updated_by, EXTRACT(epoch FROM created_at)::BIGINT as created_at, EXTRACT(epoch FROM updated_at)::BIGINT as updated_at FROM %s %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		TableName, whereClause, len(args)+1, len(args)+2)
	args = append(args, pageSize, offset)

	// 使用 model.go 中定义的结构体接收查询结果
	var dbResults []infoAtomTypeSummaryDB

	err = r.sqlDB.QueryRowsCtx(ctx, &dbResults, dataQuery, args...)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("%w: %v", ErrSQLQueryFailed, err)
	}

	// 转换为InfoAtomType接口切片
	var results []core.InfoAtomType
	for _, dbResult := range dbResults {
		result := &core.BasicInfoAtomType{
			ID:        dbResult.ID,
			TenantId:  dbResult.TenantId,
			Name:      dbResult.Name,
			Version:   dbResult.Version,
			TagIDs:    []string(dbResult.TagIDs),
			CreatedBy: nullStringToString(dbResult.CreatedBy),
			UpdatedBy: nullStringToString(dbResult.UpdatedBy),
			CreatedAt: dbResult.CreatedAt,
			UpdatedAt: dbResult.UpdatedAt,
		}

		// 反序列化数据格式
		if dbResult.DataFormat != "" {
			err = json.Unmarshal([]byte(dbResult.DataFormat), &result.DataFormat)
			if err != nil {
				return nil, 0, 0, 0, fmt.Errorf("%w: 数据格式反序列化失败: %v", ErrSQLQueryFailed, err)
			}
		}

		results = append(results, result)
	}

	// 批量填充标签名称
	if len(results) > 0 {
		// 收集所有唯一的标签ID
		allTagIDsMap := make(map[string]bool)
		for _, atomType := range results {
			basicType := atomType.(*core.BasicInfoAtomType)
			for _, tagID := range basicType.TagIDs {
				if tagID != "" {
					allTagIDsMap[tagID] = true
				}
			}
		}

		// 注意：批量填充标签名称的逻辑已删除（Tags 字段已从 Core 层移除）
		// Phase 2 重构后，Manager 层的 DTO 会负责填充标签名称
		// getTagNamesFunc 函数保留供 Phase 2 使用
	}

	return results, total, totalPages, int64(page), nil
}

// GetInfoAtomType 获取信息原子类型
func (r *infoAtomRegistry) GetInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey) (core.InfoAtomType, error) {
	// 生成缓存键
	cacheKey := generateCacheKey(key)

	// 先尝试从缓存获取
	var cachedResult core.BasicInfoAtomType
	err := r.CacheInterface.GetCtx(ctx, cacheKey, &cachedResult)
	if err == nil {
		// Phase 1 重构：返回深拷贝，避免外部修改影响缓存
		// 注意：由于 BasicInfoAtomType 的字段都是值类型或切片，
		// 需要确保切片字段（TagIDs、DataFormat.Fields）也被拷贝
		copiedResult := cachedResult
		if len(cachedResult.TagIDs) > 0 {
			copiedResult.TagIDs = make([]string, len(cachedResult.TagIDs))
			copy(copiedResult.TagIDs, cachedResult.TagIDs)
		}
		if len(cachedResult.DataFormat.Fields) > 0 {
			copiedResult.DataFormat.Fields = make([]core.FieldConfig, len(cachedResult.DataFormat.Fields))
			copy(copiedResult.DataFormat.Fields, cachedResult.DataFormat.Fields)
		}
		return &copiedResult, nil
	}

	// 缓存未命中，从数据库查询
	query := fmt.Sprintf(`SELECT id::TEXT, tenant_id::TEXT, name, version, tag_ids, data_format, created_by, updated_by, EXTRACT(epoch FROM created_at)::BIGINT as created_at, EXTRACT(epoch FROM updated_at)::BIGINT as updated_at FROM %s WHERE id = $1::UUID`, TableName)

	// 使用 model.go 中定义的结构体接收查询结果
	var dbResult infoAtomTypeDB

	err = r.sqlDB.QueryRowCtx(ctx, &dbResult, query, key.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInfoAtomTypeNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrSQLQueryFailed, err)
	}

	// 转换为BasicInfoAtomType
	var result core.BasicInfoAtomType
	result.ID = dbResult.ID
	result.TenantId = dbResult.TenantId
	result.Name = dbResult.Name
	result.Version = dbResult.Version
	result.TagIDs = []string(dbResult.TagIDs)
	result.CreatedBy = nullStringToString(dbResult.CreatedBy)
	result.UpdatedBy = nullStringToString(dbResult.UpdatedBy)
	result.CreatedAt = dbResult.CreatedAt
	result.UpdatedAt = dbResult.UpdatedAt

	// 反序列化数据格式
	if dbResult.DataFormat != "" {
		err = json.Unmarshal([]byte(dbResult.DataFormat), &result.DataFormat)
		if err != nil {
			return nil, fmt.Errorf("%w: 数据格式反序列化失败: %v", ErrSQLQueryFailed, err)
		}
	}

	// 注意：填充标签名称的逻辑已删除（Tags 字段已从 Core 层移除）
	// Phase 2 重构后，Manager 层的 DTO 会负责填充标签名称
	// getTagNamesFunc 函数保留供 Phase 2 使用

	// 保存到缓存
	r.CacheInterface.SetCtx(ctx, cacheKey, &result)

	return &result, nil
}

// UpdateInfoAtomType 更新信息原子类型
func (r *infoAtomRegistry) UpdateInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey, infoAtomType core.InfoAtomType) error {
	// 检查参数
	if infoAtomType == nil {
		return fmt.Errorf("%w: 信息原子类型不能为空", ErrInvalidInfoAtomTypeSpec)
	}

	if r.config.RunMode != core.Manager {
		return fmt.Errorf("%w: %s", ErrNotManagerMode, r.config.RunMode)
	}

	// 检查必填字段
	if key.ID == "" {
		return ErrInfoAtomTypeIDEmpty
	}

	// 检查信息原子类型是否存在
	if !r.CheckInfoAtomType(ctx, key) {
		return fmt.Errorf("%w: id=%s",
			ErrInfoAtomTypeNotFound,
			key.ID)
	}

	// 转换为基本类型进行存储
	basicType, ok := infoAtomType.(*core.BasicInfoAtomType)
	if !ok {
		// 如果不是基本类型，则创建一个
		basicType = &core.BasicInfoAtomType{
			TenantId:   infoAtomType.GetTenantId(),
			Name:       infoAtomType.GetName(),
			Version:    infoAtomType.GetVersion(),
			DataFormat: infoAtomType.GetDataFormat(),
			UpdatedBy:  infoAtomType.GetUpdatedBy(), // 保留更新人信息
		}
	}

	// 序列化数据格式
	dataFormatBytes, err := json.Marshal(basicType.DataFormat)
	if err != nil {
		return fmt.Errorf("%w: 数据格式序列化失败: %v", ErrInvalidInfoAtomTypeSpec, err)
	}

	// 更新数据库（updated_at由触发器自动更新）
	_, err = r.sqlDB.ExecCtx(ctx,
		fmt.Sprintf(`UPDATE %s SET tag_ids = $1, data_format = $2, updated_by = $3 WHERE id = $4::UUID`, TableName),
		pq.Array(basicType.TagIDs), string(dataFormatBytes), basicType.UpdatedBy, key.ID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSQLUpdateFailed, err)
	}

	// 清除缓存
	cacheKey := generateCacheKey(key)
	r.CacheInterface.DelCtx(ctx, cacheKey)

	return nil
}

// DeleteInfoAtomType 删除信息原子类型
func (r *infoAtomRegistry) DeleteInfoAtomType(ctx context.Context, key core.InfoAtomTypeKey) error {
	if r.config.RunMode != core.Manager {
		return fmt.Errorf("%w: %s", ErrNotManagerMode, r.config.RunMode)
	}

	// 检查必填字段
	if key.ID == "" {
		return ErrInfoAtomTypeIDEmpty
	}

	// 检查信息原子类型是否存在
	if !r.CheckInfoAtomType(ctx, key) {
		return fmt.Errorf("%w: id=%s",
			ErrInfoAtomTypeNotFound,
			key.ID)
	}

	// Phase 1 重构：TODO - 检查信息原子类型是否被引用
	// 理想实现：注入 CheckInfoAtomTypeUsageFunc 函数类型，由 graphRegistry 提供实现
	// 当前限制：infoAtomRegistry 无法访问 graphRegistry 的 infoAtomTypeIndex
	// 临时方案：允许删除，依赖业务层约束（前端提示用户）
	// Phase 2 可以通过以下方式完善：
	//   1. 定义 type CheckInfoAtomTypeUsageFunc func(ctx, key) (bool, []string)
	//   2. 在 newInfoAtomRegistry 构造函数中注入该函数
	//   3. 此处调用：if isUsed, graphIDs := r.checkUsageFunc(ctx, key); isUsed {
	//         return fmt.Errorf("%w: 被以下逻辑图引用: %v", core.ErrInfoAtomTypeInUse, graphIDs)
	//      }

	// 生成缓存键并删除缓存
	cacheKey := generateCacheKey(key)
	r.CacheInterface.DelCtx(ctx, cacheKey)

	// 删除数据库记录
	_, err := r.sqlDB.ExecCtx(ctx,
		fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, TableName),
		key.ID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSQLDeleteFailed, err)
	}

	return nil
}

// validateAndExtractPayload 验证并提取信息原子载荷（内部使用）
//
// 职责：
//  1. 解析原始JSON数据
//  2. 根据InfoAtomType的DataFormat提取配置的字段
//  3. 验证字段类型
//  4. 调用InfoAtomType.Validate()进行业务验证
//  5. 解析标签（key:value格式）
//
// 注意：这是私有方法，仅供 CreateInfoAtom 内部复用
func (r *infoAtomRegistry) validateAndExtractPayload(
	ctx context.Context,
	rawData []byte,
	tags []string,
	infoAtomType core.InfoAtomType,
) (map[string]any, map[string]string, error) {
	// 1. 解析JSON
	var rawPayload map[string]any
	if err := json.Unmarshal(rawData, &rawPayload); err != nil {
		return nil, nil, fmt.Errorf("JSON解析失败: %w", err)
	}

	// 2. 根据DataFormat提取字段
	payload, err := extractFields(rawPayload, infoAtomType.GetDataFormat())
	if err != nil {
		return nil, nil, fmt.Errorf("字段提取失败: %w", err)
	}

	// 3. 调用InfoAtomType的Validate方法验证
	if err := infoAtomType.Validate(payload); err != nil {
		return nil, nil, fmt.Errorf("载荷验证失败: %w", err)
	}

	// 4. 解析标签
	labels := parseTags(tags)

	return payload, labels, nil
}

// CreateInfoAtom 创建信息原子
//
// 职责：
//  1. 验证并提取载荷
//  2. 生成 UUID
//  3. 构造 InfoAtom 对象
//
// 注意：这是从 Engine 层迁移的职责，使得 InfoAtomRegistry 负责领域对象的完整生命周期
func (r *infoAtomRegistry) CreateInfoAtom(
	ctx context.Context,
	req core.InfoAtomRequest,
	infoAtomType core.InfoAtomType,
) (core.InfoAtom, error) {
	// 1. 验证并提取载荷（调用私有方法）
	payload, labels, err := r.validateAndExtractPayload(ctx, req.RawData, req.Tags, infoAtomType)
	if err != nil {
		return nil, err
	}

	// 2. 生成 UUID
	infoAtomID := uuid.New().String()

	// 3. 构造 InfoAtom 对象
	infoAtom := core.NewInfoAtom(
		req.TenantID,
		infoAtomID,
		infoAtomType,
		req.Source,
		req.Timestamp,
		labels,
		payload,
	)

	return infoAtom, nil
}
