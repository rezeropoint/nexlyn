package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"

	"github.com/rezeropoint/etcdtrigger"

	"github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/stores/monc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"go.mongodb.org/mongo-driver/bson"
)

// graphRegistry 实现 GraphRegistry 接口
type graphRegistry struct {
	config *Config

	mongoDB    *monc.Model
	sqlDB      sqlx.SqlConn
	etcdClient etcdtrigger.EtcdClient

	createBlockFunc core.CreateBlockFunc
	getTagNamesFunc core.GetTagNamesByIDsFunc

	graphs map[core.GraphKey]core.LogicGraph
	mu     sync.RWMutex

	// 信息原子类型索引，用于快速查找订阅某种信息原子类型的图和节点
	// 结构: map[core.InfoAtomTypeKey]map[core.GraphKey][]nodeID
	infoAtomTypeIndex map[core.InfoAtomTypeKey]map[core.GraphKey][]string

	ctx    context.Context
	cancel context.CancelFunc
}

// newGraphRegistry 创建一个新的图注册表
// 参数：
//   - sqlConn: 已建立的 PostgreSQL 连接（外部注入）
//   - mongoDB: 已建立的 MongoDB Model（外部注入）
//   - getTagNamesFunc: 根据标签ID获取标签名称的函数（用于标签查询）
func newGraphRegistry(
	ctx context.Context,
	cancel context.CancelFunc,
	config *Config,
	sqlConn sqlx.SqlConn,
	mongoDB *monc.Model,
	createBlockFunc core.CreateBlockFunc,
	getTagNamesFunc core.GetTagNamesByIDsFunc) (*graphRegistry, error) {

	// 创建 Etcd 客户端（包内建立连接）
	etcdClient, err := etcdtrigger.NewEtcdClient(ctx, cancel, &config.EtcdConfig)
	if err != nil {
		return nil, err
	}

	registry := &graphRegistry{
		config:            config,
		graphs:            make(map[core.GraphKey]core.LogicGraph),
		mongoDB:           mongoDB,
		sqlDB:             sqlConn,
		createBlockFunc:   createBlockFunc,
		getTagNamesFunc:   getTagNamesFunc,
		infoAtomTypeIndex: make(map[core.InfoAtomTypeKey]map[core.GraphKey][]string),
		etcdClient:        etcdClient,
		ctx:               ctx,
		cancel:            cancel,
	}

	if config.RunMode == core.Engine {
		if err := registry.watchConfigChanges(); err != nil {
			return nil, err
		}
	}

	return registry, nil
}

// InitTable 初始化数据库表
func (r *graphRegistry) InitGraphTable(ctx context.Context) error {
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

func (r *graphRegistry) GetGraphConfig(ctx context.Context, key core.GraphKey) (core.GraphConfig, error) {
	// 先从PostgreSQL查询基本信息
	querySQL := fmt.Sprintf(`SELECT id::TEXT, tenant_id::TEXT, org_id::TEXT, name, version, description, tag_ids, enable, icon, icon_color, created_by, updated_by,
	              EXTRACT(epoch FROM created_at)::BIGINT as created_at, EXTRACT(epoch FROM updated_at)::BIGINT as updated_at
	              FROM %s WHERE id = $1::UUID`, TableName)
	var basic graphConfigBasicDB
	err := r.sqlDB.QueryRowCtx(ctx, &basic, querySQL, key.ID)
	if err != nil {
		return core.GraphConfig{}, fmt.Errorf("%w: id=%s", ErrGraphNotFound, key.ID)
	}

	// 从MongoDB查询详细信息
	filter := bson.M{
		"id": key.ID,
	}

	var detail graphConfigDetail
	cacheKey := generateCacheKey(key.ID)
	err = r.mongoDB.FindOne(ctx, cacheKey, &detail, filter)
	if err != nil {
		// 如果MongoDB中没有详细信息，创建一个空的详细信息
		detail = graphConfigDetail{
			ID:       key.ID,
			TenantId: basic.TenantId,
			OrgID:    basic.OrgID,
			Name:     basic.Name,
			Version:  basic.Version,
			Nodes:    []graphNodeConfigMongo{}, // 使用 MongoDB 专用类型
			Edges:    []graphEdgeConfigMongo{}, // 使用 MongoDB 专用类型
		}
	}

	// 转换 MongoDB 结构为 Core 层结构
	nodes := make([]core.NodeConfig, len(detail.Nodes))
	for i, mongoNode := range detail.Nodes {
		nodes[i] = mongoToNodeConfig(mongoNode)
	}

	edges := make([]core.EdgeConfig, len(detail.Edges))
	for i, mongoEdge := range detail.Edges {
		edges[i] = mongoToEdgeConfig(mongoEdge)
	}

	// 组装完整配置
	config := core.GraphConfig{
		ID:          basic.ID,
		TenantId:    basic.TenantId,
		OrgID:       basic.OrgID,
		Name:        basic.Name,
		Version:     basic.Version,
		Description: basic.Description,
		TagIDs:      []string(basic.TagIDs),
		Enable:      basic.Enable,
		Icon:        nullStringToString(basic.Icon),
		IconColor:   nullStringToString(basic.IconColor),
		Nodes:       nodes, // 使用转换后的 Core 类型
		Edges:       edges, // 使用转换后的 Core 类型
		CreatedBy:   nullStringToString(basic.CreatedBy),
		UpdatedBy:   nullStringToString(basic.UpdatedBy),
		CreatedAt:   basic.CreatedAt,
		UpdatedAt:   basic.UpdatedAt,
	}

	// 注意：Tags 字段已从 Core 层移除
	// Phase 2 重构后，Manager 层的 DTO 会负责填充标签名称
	// getTagNamesFunc 函数保留供 Phase 2 使用

	return config, nil
}

func (r *graphRegistry) CreateGraphConfig(ctx context.Context, config core.GraphConfig) error {
	// 检查运行模式，只有 Manager 模式才能创建图
	if r.config.RunMode != core.Manager {
		return fmt.Errorf("%w: %s", ErrNotManagerMode, r.config.RunMode)
	}

	// 处理节点和边的ID生成（在验证之前）
	if err := processNodeAndEdgeIDs(&config); err != nil {
		return fmt.Errorf("处理节点和边ID失败: %w", err)
	}

	// 验证配置
	if err := core.ScanConfig(&config); err != nil {
		return fmt.Errorf("%w: %v", ErrGraphConfigInvalid, err)
	}

	// 使用事务确保数据一致性
	return r.sqlDB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 检查PostgreSQL中是否已存在（仅当ID非空时检查）
		if config.ID != "" {
			checkSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id = $1::UUID", TableName)
			var countResult struct {
				Count int64 `db:"count"`
			}
			err := session.QueryRowCtx(ctx, &countResult, checkSQL, config.ID)
			if err != nil {
				return fmt.Errorf("%w: 检查图是否存在失败: %v", ErrSQLQueryFailed, err)
			}
			if countResult.Count > 0 {
				return fmt.Errorf("%w: id=%s", ErrGraphAlreadyExists, config.ID)
			}
		}

		// 1. 将基本信息插入PostgreSQL（ID和时间戳由数据库自动生成）
		var result struct {
			ID        string `db:"id"`
			CreatedAt int64  `db:"created_at"`
			UpdatedAt int64  `db:"updated_at"`
		}
		insertSQL := fmt.Sprintf(`
			INSERT INTO %s (id, tenant_id, org_id, name, version, description, tag_ids, enable, icon, icon_color, created_by)
			VALUES (COALESCE(NULLIF($1, ''), gen_random_uuid()::TEXT)::UUID, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, ''), NULLIF($10, ''), $11)
			RETURNING id::TEXT, EXTRACT(epoch FROM created_at)::BIGINT as created_at, EXTRACT(epoch FROM updated_at)::BIGINT as updated_at
		`, TableName)
		err := session.QueryRowCtx(ctx, &result, insertSQL,
			config.ID, config.TenantId, config.OrgID, config.Name, config.Version,
			config.Description, pq.Array(config.TagIDs), config.Enable, config.Icon, config.IconColor, config.CreatedBy)
		if err != nil {
			return fmt.Errorf("%w: PostgreSQL插入失败: %v", ErrSQLInsertFailed, err)
		}

		// 更新配置中的ID和时间戳
		config.ID = result.ID
		config.CreatedAt = result.CreatedAt
		config.UpdatedAt = result.UpdatedAt

		// 2. 将详细信息（Nodes和Edges）存储到MongoDB
		// 转换 Core 层结构为 MongoDB 存储结构
		mongoNodes := make([]graphNodeConfigMongo, len(config.Nodes))
		for i, node := range config.Nodes {
			mongoNodes[i] = nodeConfigToMongo(node)
		}

		mongoEdges := make([]graphEdgeConfigMongo, len(config.Edges))
		for i, edge := range config.Edges {
			mongoEdges[i] = edgeConfigToMongo(edge)
		}

		detail := graphConfigDetail{
			ID:       config.ID,
			TenantId: config.TenantId,
			OrgID:    config.OrgID,
			Name:     config.Name,
			Version:  config.Version,
			Nodes:    mongoNodes, // 使用转换后的 MongoDB 类型
			Edges:    mongoEdges, // 使用转换后的 MongoDB 类型
		}

		// 生成缓存key
		cacheKey := generateCacheKey(config.ID)

		// 将详细配置存储到 MongoDB
		_, err = r.mongoDB.InsertOne(ctx, cacheKey, &detail)
		if err != nil {
			// MongoDB插入失败会导致事务回滚
			return fmt.Errorf("%w: %v", ErrMongoDBInsertFailed, err)
		}

		// 3. 通过 etcd 发出通知（在事务成功后）
		etcdKey := generateEtcdKey(r.config.EtcdConfig.Key, "create", config.ID)

		// 创建简单的通知消息，包含基本标识信息
		notifyMsg := generateNotifyMessage("create", config.ID)

		// 序列化通知消息
		notifyBytes, err := json.Marshal(notifyMsg)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrMarshalFailed, err)
		}

		// 向 etcd 发送通知
		if err := r.etcdClient.Put(etcdKey, notifyBytes); err != nil {
			return fmt.Errorf("%w: %v", ErrEtcdNotifyFailed, err)
		}

		return nil
	})
}

// UpdateGraph 更新逻辑图配置
func (r *graphRegistry) UpdateGraphConfig(ctx context.Context, config core.GraphConfig) error {
	// 检查运行模式，只有 Manager 模式才能更新图
	if r.config.RunMode != core.Manager {
		return fmt.Errorf("%w: %s", ErrNotManagerMode, r.config.RunMode)
	}

	// 处理节点和边的ID生成（在验证之前）
	if err := processNodeAndEdgeIDs(&config); err != nil {
		return fmt.Errorf("处理节点和边ID失败: %w", err)
	}

	// 验证配置
	if err := core.ScanConfig(&config); err != nil {
		return fmt.Errorf("%w: %v", ErrGraphConfigInvalid, err)
	}

	// 使用事务确保数据一致性
	return r.sqlDB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 先从PostgreSQL查询现有配置以获取创建时间
		querySQL := fmt.Sprintf("SELECT EXTRACT(epoch FROM created_at)::BIGINT as created_at FROM %s WHERE id = $1::UUID", TableName)
		var existingCreatedAt int64
		err := session.QueryRowCtx(ctx, &existingCreatedAt, querySQL, config.ID)
		if err != nil {
			return fmt.Errorf("%w: id=%s", ErrGraphNotFound, config.ID)
		}

		// 保留原创建时间
		config.CreatedAt = existingCreatedAt

		// 1. 更新PostgreSQL中的基本信息（updated_at由触发器自动更新）
		var updatedAt int64
		updateSQL := fmt.Sprintf(`
			UPDATE %s
			SET tenant_id = $2, name = $3, version = $4, description = $5, tag_ids = $6, enable = $7, icon = NULLIF($8, ''), icon_color = NULLIF($9, ''), updated_by = $10
			WHERE id = $1::UUID
			RETURNING EXTRACT(epoch FROM updated_at)::BIGINT as updated_at
		`, TableName)
		err = session.QueryRowCtx(ctx, &updatedAt, updateSQL,
			config.ID, config.TenantId, config.Name, config.Version,
			config.Description, pq.Array(config.TagIDs), config.Enable, config.Icon, config.IconColor, config.UpdatedBy)
		if err != nil {
			return fmt.Errorf("%w: PostgreSQL更新失败: %v", ErrSQLUpdateFailed, err)
		}
		config.UpdatedAt = updatedAt

		// 2. 更新MongoDB中的详细信息
		// 转换 Core 层结构为 MongoDB 存储结构
		mongoNodes := make([]graphNodeConfigMongo, len(config.Nodes))
		for i, node := range config.Nodes {
			mongoNodes[i] = nodeConfigToMongo(node)
		}

		mongoEdges := make([]graphEdgeConfigMongo, len(config.Edges))
		for i, edge := range config.Edges {
			mongoEdges[i] = edgeConfigToMongo(edge)
		}

		detail := graphConfigDetail{
			ID:       config.ID,
			TenantId: config.TenantId,
			OrgID:    config.OrgID,
			Name:     config.Name,
			Version:  config.Version,
			Nodes:    mongoNodes, // 使用转换后的 MongoDB 类型
			Edges:    mongoEdges, // 使用转换后的 MongoDB 类型
		}

		// 生成缓存key
		cacheKey := generateCacheKey(config.ID)

		// 构造更新条件
		filter := bson.M{
			"id": config.ID,
		}

		// 更新MongoDB中的配置
		_, err = r.mongoDB.ReplaceOne(ctx, cacheKey, filter, &detail)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrMongoDBUpdateFailed, err)
		}

		// 3. 通过 etcd 发出通知（在事务成功后）
		etcdKey := generateEtcdKey(r.config.EtcdConfig.Key, "update", config.ID)

		// 创建简单的通知消息，包含基本标识信息
		notifyMsg := generateNotifyMessage("update", config.ID)

		// 序列化通知消息
		notifyBytes, err := json.Marshal(notifyMsg)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrMarshalFailed, err)
		}

		// 向 etcd 发送通知
		if err := r.etcdClient.Put(etcdKey, notifyBytes); err != nil {
			return fmt.Errorf("%w: %v", ErrEtcdNotifyFailed, err)
		}

		return nil
	})
}

// DeleteGraph 删除逻辑图
func (r *graphRegistry) DeleteGraphConfig(ctx context.Context, key core.GraphKey) error {
	// 检查运行模式，只有 Manager 模式才能删除图
	if r.config.RunMode != core.Manager {
		return fmt.Errorf("%w: %s", ErrNotManagerMode, r.config.RunMode)
	}

	// Phase 1 重构：检查图是否正在运行中（防止删除运行中的图）
	r.mu.RLock()
	_, isRunning := r.graphs[key]
	r.mu.RUnlock()

	if isRunning {
		return fmt.Errorf("%w: 无法删除正在运行的逻辑图 (id=%s)", core.ErrGraphInUse, key.ID)
	}

	// 使用事务确保数据一致性
	return r.sqlDB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 先查询配置以验证图是否存在（不需要获取完整信息）
		checkSQL := fmt.Sprintf("SELECT 1 FROM %s WHERE id = $1", TableName)
		var exists int
		err := session.QueryRowCtx(ctx, &exists, checkSQL, key.ID)
		if err != nil {
			return fmt.Errorf("%w: id=%s", ErrGraphNotFound, key.ID)
		}

		// 1. 从PostgreSQL删除基本信息（在事务中）
		deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE id = $1", TableName)
		_, err = session.ExecCtx(ctx, deleteSQL, key.ID)
		if err != nil {
			return fmt.Errorf("%w: PostgreSQL删除失败: %v", ErrSQLDeleteFailed, err)
		}

		// 2. 从MongoDB删除详细信息
		filter := bson.M{
			"id": key.ID,
		}

		cacheKey := generateCacheKey(key.ID)
		_, err = r.mongoDB.DeleteOne(ctx, cacheKey, filter)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrMongoDBDeleteFailed, err)
		}

		// 3. 通过 etcd 发出通知（在事务成功后）
		etcdKey := generateEtcdKey(r.config.EtcdConfig.Key, "delete", key.ID)

		// 创建简单的通知消息，包含基本标识信息
		notifyMsg := generateNotifyMessage("delete", key.ID)

		// 序列化通知消息
		notifyBytes, err := json.Marshal(notifyMsg)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrMarshalFailed, err)
		}

		// 向 etcd 发送通知
		if err := r.etcdClient.Put(etcdKey, notifyBytes); err != nil {
			return fmt.Errorf("%w: %v", ErrEtcdNotifyFailed, err)
		}

		return nil
	})
}

// LoadGraph 从配置创建并加载一个逻辑图
func (r *graphRegistry) LoadGraph(config core.GraphConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.config.RunMode != core.Engine {
		return fmt.Errorf("%w: %s", ErrNotEngineMode, r.config.RunMode)
	}

	// 检查是否已经存在，如果存在则先移除
	key := core.GraphKey{ID: config.ID}
	if _, exists := r.graphs[key]; exists {
		// 先从索引中移除
		err := r.UnregisterGraph(key)
		if err != nil {
			return err
		}
		// 从图映射中删除
		delete(r.graphs, key)
	}

	// 创建新的逻辑图
	graph := core.NewBasicLogicGraph(
		config.TenantId,
		config.Name,
		config.Version,
		config.Description,
		config.Enable,
	)

	// 创建并添加所有节点
	nodeMap := make(map[string]core.Node)
	for _, nodeConfig := range config.Nodes {
		// 使用提供的创建函数创建逻辑块
		block, err := r.createBlockFunc(
			nodeConfig.ID,
			core.BlockKey{BlockType: nodeConfig.BlockType, Version: nodeConfig.BlockVersion},
			nodeConfig.BlockConfig,
		)
		if err != nil {
			return fmt.Errorf("%w: 节点ID=%s, %v", ErrBlockCreateFailed, nodeConfig.ID, err)
		}

		// 创建节点
		node := &core.BasicNode{
			ID:                        nodeConfig.ID,
			Type:                      nodeConfig.Type,
			Block:                     block,
			EntryPoint:                nodeConfig.IsEntryPoint,
			SubscribedInfoAtomTypeIDs: nodeConfig.SubscribedInfoAtomTypeIDs,
			SubscribedSource:          nodeConfig.SubscribedSource,
			SubscribedLabels:          nodeConfig.SubscribedLabels,
		}

		// 添加到图和本地缓存
		graph.AddNode(node)
		nodeMap[nodeConfig.ID] = node
	}

	// 创建并添加所有边
	for _, edgeConfig := range config.Edges {
		// 验证源节点和目标节点存在
		if _, exists := nodeMap[edgeConfig.SourceID]; !exists {
			return fmt.Errorf("%w: %s", ErrEdgeSourceNotFound, edgeConfig.SourceID)
		}
		if _, exists := nodeMap[edgeConfig.TargetID]; !exists {
			return fmt.Errorf("%w: %s", ErrEdgeTargetNotFound, edgeConfig.TargetID)
		}

		// 创建边
		edge := &core.BasicEdge{
			ID:        edgeConfig.ID,
			SourceID:  edgeConfig.SourceID,
			TargetID:  edgeConfig.TargetID,
			Condition: edgeConfig.Condition,
		}

		// 添加到图
		graph.AddEdge(edge)
	}

	// 验证图结构
	enable, err := graph.Validate()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrGraphValidateFailed, err)
	}
	// 如果校验结果认为图不完全成功，则强制设置启用状态为关闭
	if !enable {
		graph.SetEnable(false)
	}

	// 存储到注册表
	r.graphs[key] = graph

	// 更新信息原子类型索引
	r.updateInfoAtomTypeIndex(key, graph)

	return nil
}

// UnregisterGraph 注销逻辑图
func (r *graphRegistry) UnregisterGraph(key core.GraphKey) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.config.RunMode != core.Engine {
		return fmt.Errorf("%w: %s", ErrNotEngineMode, r.config.RunMode)
	}

	if _, exists := r.graphs[key]; !exists {
		return fmt.Errorf("%w: id=%s", ErrGraphNotFound, key.ID)
	}

	// 遍历所有信息原子类型索引
	for _, graphsMap := range r.infoAtomTypeIndex {
		// 从每个信息原子类型的图映射中删除该图
		delete(graphsMap, key)
	}

	// 清理空映射
	for atomKey, graphsMap := range r.infoAtomTypeIndex {
		if len(graphsMap) == 0 {
			delete(r.infoAtomTypeIndex, atomKey)
		}
	}

	// 从图映射中删除
	delete(r.graphs, key)
	return nil
}

// GetGraph 根据ID和版本获取图
func (r *graphRegistry) GetGraph(key core.GraphKey) (core.LogicGraph, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.config.RunMode != core.Engine {
		return nil, fmt.Errorf("%w: %s", ErrNotEngineMode, r.config.RunMode)
	}

	graph, exists := r.graphs[key]
	if !exists {
		return nil, fmt.Errorf("%w: id=%s", ErrGraphNotFound, key.ID)
	}

	return graph, nil
}

// FindGraphsByInfoAtom 根据信息原子查找可能处理该信息原子的逻辑图和入口节点
func (r *graphRegistry) FindGraphsByInfoAtom(infoAtom core.InfoAtom) (map[core.GraphKey][]core.Node, error) {
	if infoAtom == nil {
		return nil, fmt.Errorf("信息原子不能为空")
	}

	if r.config.RunMode != core.Engine {
		return nil, fmt.Errorf("%w: %s", ErrNotEngineMode, r.config.RunMode)
	}

	// 获取信息原子类型
	atomType := infoAtom.GetType()
	if atomType == nil {
		return nil, fmt.Errorf("信息原子类型不能为空")
	}

	atomTypeKey := core.InfoAtomTypeKey{
		ID: atomType.GetID(),
		// TenantId: atomType.GetTenantId(),
		// Name:      atomType.GetName(),
		// Version:   atomType.GetVersion(),
	}

	// 先根据类型查找
	result, err := r.findGraphsByInfoAtomType(atomTypeKey)
	if err != nil {
		return nil, err
	}

	// 如果没有找到匹配的图，直接返回空结果
	if len(result) == 0 {
		return result, nil
	}

	// 获取信息原子的标签和来源
	atomLabels := infoAtom.GetLabels()
	atomSource := infoAtom.GetSource()

	// 根据标签和来源进一步过滤
	filteredResult := make(map[core.GraphKey][]core.Node, len(result))

	for gKey, nodes := range result {
		matchedNodes := make([]core.Node, 0, len(nodes))

		for _, node := range nodes {
			// 检查节点是否订阅了信息原子的标签
			if !matchLabels(node.GetSubscribedLabels(), atomLabels) {
				continue
			}

			// 检查节点是否订阅了信息原子的来源
			nodeSubscribedSource := node.GetSubscribedSource()
			if nodeSubscribedSource != "" && nodeSubscribedSource != atomSource {
				continue
			}

			matchedNodes = append(matchedNodes, node)
		}

		if len(matchedNodes) > 0 {
			filteredResult[gKey] = matchedNodes
		}
	}

	return filteredResult, nil
}

// GetGraphConfigList 分页获取逻辑图配置列表
func (r *graphRegistry) GetGraphConfigList(ctx context.Context, params core.GraphList) ([]core.GraphConfig, int64, int64, int64, error) {
	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	if params.TenantId != "" {
		conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
		args = append(args, params.TenantId)
		argIndex++
	}

	// 组织权限过滤（关键安全控制）
	if len(params.OrgIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf("org_id = ANY($%d)", argIndex))
		args = append(args, pq.Array(params.OrgIDs))
		argIndex++
	}

	if params.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+params.Name+"%")
		argIndex++
	}

	if len(params.Tags) > 0 {
		conditions = append(conditions, fmt.Sprintf("tag_ids && $%d", argIndex))
		args = append(args, pq.Array(params.Tags))
		argIndex++
	}

	if params.Enable != nil {
		conditions = append(conditions, fmt.Sprintf("enable = $%d", argIndex))
		args = append(args, *params.Enable)
		argIndex++
	}

	// 构建WHERE子句
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}

	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", TableName, whereClause)
	var countResult struct {
		Count int64 `db:"count"`
	}
	err := r.sqlDB.QueryRowCtx(ctx, &countResult, countSQL, args...)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("%w: 查询总数失败: %v", ErrSQLQueryFailed, err)
	}
	total := countResult.Count

	// 计算分页信息
	offset := (params.Page - 1) * params.PageSize
	totalPages := (total + int64(params.PageSize) - 1) / int64(params.PageSize)

	// 查询基本信息列表
	listSQL := fmt.Sprintf(`
		SELECT id::TEXT, tenant_id::TEXT, org_id::TEXT, name, version, description, tag_ids, enable, icon, icon_color, created_by, updated_by,
		EXTRACT(epoch FROM created_at)::BIGINT as created_at, EXTRACT(epoch FROM updated_at)::BIGINT as updated_at
		FROM %s %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, TableName, whereClause, argIndex, argIndex+1)

	listArgs := make([]interface{}, len(args)+2)
	copy(listArgs, args)
	listArgs[len(args)] = params.PageSize
	listArgs[len(args)+1] = offset

	var basicConfigs []graphConfigBasicDB
	err = r.sqlDB.QueryRowsCtx(ctx, &basicConfigs, listSQL, listArgs...)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("%w: 查询列表失败: %v", ErrSQLQueryFailed, err)
	}

	// 从MongoDB获取详细信息并组装完整配置
	var configs []core.GraphConfig
	if len(basicConfigs) == 0 {
		return configs, total, int64(params.Page), totalPages, nil
	}

	// 提取所有ID用于批量查询
	ids := make([]string, len(basicConfigs))
	basicConfigMap := make(map[string]graphConfigBasicDB)
	for i, basic := range basicConfigs {
		ids[i] = basic.ID
		basicConfigMap[basic.ID] = basic
	}

	// 使用$in操作符批量查询MongoDB中的详细信息
	filter := bson.M{
		"id": bson.M{"$in": ids},
	}

	// 批量查询详细信息
	var details []graphConfigDetail
	err = r.mongoDB.Find(ctx, &details, filter)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("%w: MongoDB批量查询失败: %v", ErrMongoDBQueryFailed, err)
	}

	// 创建详细信息映射以便快速查找
	detailMap := make(map[string]graphConfigDetail)
	for _, detail := range details {
		detailMap[detail.ID] = detail
	}

	// 组装完整配置
	for _, basic := range basicConfigs {
		detail, exists := detailMap[basic.ID]
		if !exists {
			// 如果MongoDB中没有详细信息，创建一个空的详细信息
			detail = graphConfigDetail{
				Nodes: []graphNodeConfigMongo{}, // 使用 MongoDB 专用类型
				Edges: []graphEdgeConfigMongo{}, // 使用 MongoDB 专用类型
			}
		}

		// 转换 MongoDB 结构为 Core 层结构
		nodes := make([]core.NodeConfig, len(detail.Nodes))
		for i, mongoNode := range detail.Nodes {
			nodes[i] = mongoToNodeConfig(mongoNode)
		}

		edges := make([]core.EdgeConfig, len(detail.Edges))
		for i, mongoEdge := range detail.Edges {
			edges[i] = mongoToEdgeConfig(mongoEdge)
		}

		// 组装完整配置
		config := core.GraphConfig{
			ID:          basic.ID,
			TenantId:    basic.TenantId,
			OrgID:       basic.OrgID,
			Name:        basic.Name,
			Version:     basic.Version,
			Description: basic.Description,
			TagIDs:      []string(basic.TagIDs),
			Enable:      basic.Enable,
			Icon:        nullStringToString(basic.Icon),
			IconColor:   nullStringToString(basic.IconColor),
			Nodes:       nodes, // 使用转换后的 Core 类型
			Edges:       edges, // 使用转换后的 Core 类型
			CreatedBy:   nullStringToString(basic.CreatedBy),
			UpdatedBy:   nullStringToString(basic.UpdatedBy),
			CreatedAt:   basic.CreatedAt,
			UpdatedAt:   basic.UpdatedAt,
		}

		configs = append(configs, config)
	}

	// 注意：批量填充标签名称的逻辑已删除（Tags 字段已从 Core 层移除）
	// Phase 2 重构后，Manager 层的 DTO 会负责填充标签名称
	// getTagNamesFunc 函数保留供 Phase 2 使用

	return configs, total, int64(params.Page), totalPages, nil
}

// Close 关闭图注册表，清理资源
func (r *graphRegistry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 关闭 etcd 客户端（会自动取消监听的 goroutine）
	if r.etcdClient != nil {
		if err := r.etcdClient.Close(); err != nil {
			return fmt.Errorf("关闭 etcd 客户端失败: %v", err)
		}
		r.etcdClient = nil
	}

	// 清理内存中的图和索引
	r.graphs = make(map[core.GraphKey]core.LogicGraph)
	r.infoAtomTypeIndex = make(map[core.InfoAtomTypeKey]map[core.GraphKey][]string)

	return nil
}
