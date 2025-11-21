package graph

import "errors"

// 预定义错误
var (
	ErrConfigNil               = errors.New("配置不能为空")
	ErrGraphAlreadyRegistered  = errors.New("逻辑图已注册")                       // ErrGraphAlreadyRegistered 表示逻辑图已经注册
	ErrGraphNotFound           = errors.New("逻辑图未找到")                       // ErrGraphNotFound 表示逻辑图未找到
	ErrNodeNotFound            = errors.New("节点未找到")                        // ErrNodeNotFound 表示节点未找到
	ErrInvalidGraphSpec        = errors.New("无效的图规格")                       // ErrInvalidGraphSpec 表示无效的图规格
	ErrInvalidRegistry         = errors.New("无效的注册表")                       // ErrInvalidRegistry 表示无效的注册表
	ErrGraphIDEmpty            = errors.New("图ID不能为空")                      // ErrGraphIDEmpty 表示图ID为空
	ErrBlockCreatorNotProvided = errors.New("未提供createBlockFunc函数，无法创建逻辑块") // ErrBlockCreatorNotProvided 表示未提供createBlockFunc
	ErrNodeConfigNotFound      = errors.New("节点配置未找到")                      // ErrNodeConfigNotFound 表示节点配置未找到
	ErrEdgeSourceNotFound      = errors.New("边的源节点不存在")                     // ErrEdgeSourceNotFound 表示边的源节点不存在
	ErrEdgeTargetNotFound      = errors.New("边的目标节点不存在")                    // ErrEdgeTargetNotFound 表示边的目标节点不存在
	ErrGraphValidateFailed     = errors.New("图验证失败")                        // ErrGraphValidateFailed 表示图验证失败
	ErrBlockCreateFailed       = errors.New("创建节点逻辑块失败")                    // ErrBlockCreateFailed 表示创建节点逻辑块失败
	ErrLoadFromDBFailed        = errors.New("从数据库加载图配置失败")                  // ErrLoadFromDBFailed 表示从数据库加载图配置失败
	ErrNotEngineMode           = errors.New("当前模式不是EngineMode")             // ErrNotEngineMode 表示当前模式不是EngineMode

	// CreateGraph 相关错误
	ErrNotManagerMode      = errors.New("当前模式不是ManagerMode") // ErrNotManagerMode 表示当前模式不是ManagerMode
	ErrGraphConfigInvalid  = errors.New("图配置无效")             // ErrGraphConfigInvalid 表示图配置无效
	ErrGraphAlreadyExists  = errors.New("图已存在")              // ErrGraphAlreadyExists 表示图已存在
	ErrMongoDBInsertFailed = errors.New("MongoDB插入失败")       // ErrMongoDBInsertFailed 表示MongoDB插入失败
	ErrMongoDBQueryFailed  = errors.New("MongoDB查询失败")       // ErrMongoDBQueryFailed 表示MongoDB查询失败
	ErrMarshalFailed       = errors.New("序列化失败")             // ErrMarshalFailed 表示序列化失败
	ErrEtcdNotifyFailed    = errors.New("etcd通知失败")          // ErrEtcdNotifyFailed 表示etcd通知失败

	// UpdateGraph 相关错误
	ErrMongoDBUpdateFailed = errors.New("MongoDB更新失败") // ErrMongoDBUpdateFailed 表示MongoDB更新失败

	// DeleteGraph 相关错误
	ErrMongoDBDeleteFailed = errors.New("MongoDB删除失败") // ErrMongoDBDeleteFailed 表示MongoDB删除失败

	// SQL相关错误
	ErrSQLConnectionFailed = errors.New("SQL连接失败") // ErrSQLConnectionFailed 表示SQL连接失败
	ErrSQLQueryFailed      = errors.New("SQL查询失败") // ErrSQLQueryFailed 表示SQL查询失败
	ErrSQLInsertFailed     = errors.New("SQL插入失败") // ErrSQLInsertFailed 表示SQL插入失败
	ErrSQLUpdateFailed     = errors.New("SQL更新失败") // ErrSQLUpdateFailed 表示SQL更新失败
	ErrSQLDeleteFailed     = errors.New("SQL删除失败") // ErrSQLDeleteFailed 表示SQL删除失败
)
