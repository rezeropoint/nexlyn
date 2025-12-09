package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ExternalDBConfig 外部数据源配置
type ExternalDBConfig struct {
	Name         string
	Driver       string
	Host         string
	Port         int
	Database     string
	Username     string
	Password     string
	MaxOpenConns int
	MaxIdleConns int
}

// ExternalDBService 外部数据库服务，管理多个命名数据源
type ExternalDBService struct {
	mu          sync.RWMutex
	connections map[string]sqlx.SqlConn
}

// NewExternalDBService 创建外部数据库服务
func NewExternalDBService(configs []ExternalDBConfig) (*ExternalDBService, error) {
	svc := &ExternalDBService{
		connections: make(map[string]sqlx.SqlConn),
	}

	for _, cfg := range configs {
		if cfg.Name == "" {
			return nil, fmt.Errorf("数据源名称不能为空")
		}
		if _, exists := svc.connections[cfg.Name]; exists {
			return nil, fmt.Errorf("数据源名称重复: %s", cfg.Name)
		}

		// 设置默认值
		if cfg.MaxOpenConns <= 0 {
			cfg.MaxOpenConns = 10
		}
		if cfg.MaxIdleConns <= 0 {
			cfg.MaxIdleConns = 5
		}

		// 构建 DSN（使用 url.UserPassword 确保特殊字符正确编码）
		userInfo := url.UserPassword(cfg.Username, cfg.Password)
		dsn := fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=disable",
			userInfo.String(), cfg.Host, cfg.Port, cfg.Database)

		conn := sqlx.NewSqlConn(cfg.Driver, dsn)

		// 验证连接并配置连接池
		rawDB, err := conn.RawDB()
		if err != nil {
			return nil, fmt.Errorf("获取数据源 %s 原生连接失败: %w", cfg.Name, err)
		}

		if err := rawDB.Ping(); err != nil {
			return nil, fmt.Errorf("数据源 %s 连接验证失败: %w", cfg.Name, err)
		}

		// 配置连接池
		rawDB.SetMaxOpenConns(cfg.MaxOpenConns)
		rawDB.SetMaxIdleConns(cfg.MaxIdleConns)

		svc.connections[cfg.Name] = conn
		logx.Infof("[ExternalDBService] 数据源 %s 初始化成功 (host=%s:%d, db=%s)",
			cfg.Name, cfg.Host, cfg.Port, cfg.Database)
	}

	return svc, nil
}

// GetConnection 获取指定名称的数据库连接
func (s *ExternalDBService) GetConnection(name string) (sqlx.SqlConn, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conn, ok := s.connections[name]
	if !ok {
		return nil, fmt.Errorf("数据源 '%s' 不存在", name)
	}
	return conn, nil
}

// Query 执行查询（仅支持 SELECT）
// 返回 []map[string]any 格式的结果集，支持任意列结构
func (s *ExternalDBService) Query(ctx context.Context, dataSource, query string, args ...any) ([]map[string]any, error) {
	// 安全检查：只允许 SELECT 语句
	if !isSelectQuery(query) {
		return nil, fmt.Errorf("仅支持 SELECT 查询")
	}

	conn, err := s.GetConnection(dataSource)
	if err != nil {
		return nil, err
	}

	// 获取原生 *sql.DB
	rawDB, err := conn.RawDB()
	if err != nil {
		return nil, fmt.Errorf("获取原生数据库连接失败: %w", err)
	}

	// 执行查询
	rows, err := rawDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询执行失败: %w", err)
	}
	defer rows.Close()

	// 动态解析结果
	return scanRowsToMaps(rows)
}

// scanRowsToMaps 将 sql.Rows 转换为 []map[string]any
func scanRowsToMaps(rows *sql.Rows) ([]map[string]any, error) {
	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("获取列信息失败: %w", err)
	}

	var results []map[string]any

	for rows.Next() {
		// 创建值容器
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("扫描行数据失败: %w", err)
		}

		// 构建 map，处理特殊类型
		row := make(map[string]any)
		for i, col := range columns {
			row[col] = convertValue(values[i])
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历结果集失败: %w", err)
	}

	return results, nil
}

// convertValue 转换数据库值为标准 Go 类型
func convertValue(v any) any {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case []byte:
		// 将 []byte 转换为 string
		return string(val)
	default:
		return val
	}
}

// isSelectQuery 检查是否为 SELECT 查询
func isSelectQuery(query string) bool {
	trimmed := strings.TrimSpace(query)

	// 跳过 SQL 注释
	for strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "/*") {
		if strings.HasPrefix(trimmed, "--") {
			// 单行注释
			idx := strings.Index(trimmed, "\n")
			if idx == -1 {
				return false
			}
			trimmed = strings.TrimSpace(trimmed[idx+1:])
		} else {
			// 多行注释
			idx := strings.Index(trimmed, "*/")
			if idx == -1 {
				return false
			}
			trimmed = strings.TrimSpace(trimmed[idx+2:])
		}
	}

	return strings.HasPrefix(strings.ToUpper(trimmed), "SELECT")
}

// Close 关闭所有连接
func (s *ExternalDBService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for name, conn := range s.connections {
		if rawDB, err := conn.RawDB(); err == nil {
			if err := rawDB.Close(); err != nil {
				logx.Errorf("[ExternalDBService] 关闭数据源 %s 失败: %v", name, err)
			}
		}
		delete(s.connections, name)
	}

	logx.Info("[ExternalDBService] 所有数据源已关闭")
	return nil
}

// ListDataSources 列出所有已配置的数据源名称
func (s *ExternalDBService) ListDataSources() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.connections))
	for name := range s.connections {
		names = append(names, name)
	}
	return names
}
