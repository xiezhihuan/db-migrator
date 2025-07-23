package database

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"db-migrator/internal/types"
)

// Manager 数据库管理器
type Manager struct {
	config      types.Config
	connections map[string]types.DB
	baseConfig  types.DatabaseConfig
}

// NewManager 创建数据库管理器
func NewManager(config types.Config) *Manager {
	return &Manager{
		config:      config,
		connections: make(map[string]types.DB),
		baseConfig:  config.Database,
	}
}

// GetDatabase 获取指定数据库连接
func (m *Manager) GetDatabase(name string) (types.DB, error) {
	// 如果已有连接，直接返回
	if db, exists := m.connections[name]; exists {
		return db, nil
	}

	// 获取数据库配置
	dbConfig, err := m.getDatabaseConfig(name)
	if err != nil {
		return nil, err
	}

	// 创建连接
	db, err := NewMySQLDB(*dbConfig)
	if err != nil {
		return nil, fmt.Errorf("连接数据库 %s 失败: %v", name, err)
	}

	// 缓存连接
	m.connections[name] = db
	return db, nil
}

// GetDefaultDatabase 获取默认数据库连接
func (m *Manager) GetDefaultDatabase() (types.DB, string, error) {
	defaultName := m.getDefaultDatabaseName()
	db, err := m.GetDatabase(defaultName)
	return db, defaultName, err
}

// DiscoverDatabases 发现数据库（基于模式匹配）
func (m *Manager) DiscoverDatabases(ctx context.Context, patterns []string) ([]types.DatabaseInfo, error) {
	var databases []types.DatabaseInfo

	// 添加预配置的数据库
	for key, config := range m.config.Databases {
		databases = append(databases, types.DatabaseInfo{
			Name:        config.Database,
			ConfigKey:   key,
			Description: fmt.Sprintf("预配置数据库: %s", key),
			Matched:     m.matchesPatterns(config.Database, patterns),
		})
	}

	// 如果有默认数据库且不在预配置中，也添加进去
	defaultName := m.getDefaultDatabaseName()
	if defaultName != "" && !m.isDatabaseInList(defaultName, databases) {
		databases = append(databases, types.DatabaseInfo{
			Name:        defaultName,
			ConfigKey:   "default",
			Description: "默认数据库",
			Matched:     m.matchesPatterns(defaultName, patterns),
		})
	}

	// 如果有模式，尝试从数据库服务器发现更多数据库
	if len(patterns) > 0 {
		discoveredDbs, err := m.discoverFromServer(ctx, patterns)
		if err != nil {
			// 发现失败不影响整体流程，只记录警告
			fmt.Printf("警告: 从服务器发现数据库失败: %v\n", err)
		} else {
			databases = append(databases, discoveredDbs...)
		}
	}

	return databases, nil
}

// GetMatchedDatabases 获取匹配模式的数据库列表
func (m *Manager) GetMatchedDatabases(ctx context.Context, patterns []string) ([]string, error) {
	databases, err := m.DiscoverDatabases(ctx, patterns)
	if err != nil {
		return nil, err
	}

	var matched []string
	for _, db := range databases {
		if db.Matched {
			matched = append(matched, db.Name)
		}
	}

	return matched, nil
}

// CloseAll 关闭所有数据库连接
func (m *Manager) CloseAll() error {
	var errors []string
	for name, db := range m.connections {
		if err := db.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("关闭数据库 %s 失败: %v", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}

	return nil
}

// getDatabaseConfig 获取数据库配置
func (m *Manager) getDatabaseConfig(name string) (*types.DatabaseConfig, error) {
	// 首先检查预配置的数据库
	if config, exists := m.config.Databases[name]; exists {
		return &config, nil
	}

	// 检查是否是默认数据库
	defaultName := m.getDefaultDatabaseName()
	if name == defaultName || name == "default" || name == "" {
		return &m.baseConfig, nil
	}

	// 如果没有多数据库配置，且请求的是基础配置中的数据库名
	if len(m.config.Databases) == 0 && name == m.baseConfig.Database {
		return &m.baseConfig, nil
	}

	// 动态创建配置（基于基础配置）
	config := m.baseConfig
	config.Database = name
	return &config, nil
}

// getDefaultDatabaseName 获取默认数据库名
func (m *Manager) getDefaultDatabaseName() string {
	// 如果配置了默认数据库
	if m.config.Migrator.DefaultDatabase != "" {
		return m.config.Migrator.DefaultDatabase
	}

	// 如果有多数据库配置，但没有指定默认数据库，返回第一个
	if len(m.config.Databases) > 0 {
		for key := range m.config.Databases {
			return key // 返回第一个数据库配置的键名
		}
	}

	// 向后兼容：返回基础配置中的数据库名
	return m.baseConfig.Database
}

// matchesPatterns 检查数据库名是否匹配模式
func (m *Manager) matchesPatterns(dbName string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}

	for _, pattern := range patterns {
		if m.matchPattern(dbName, pattern) {
			return true
		}
	}
	return false
}

// matchPattern 匹配单个模式
func (m *Manager) matchPattern(dbName, pattern string) bool {
	// 支持通配符模式
	if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
		// 转换为正则表达式
		regexPattern := strings.ReplaceAll(pattern, "*", ".*")
		regexPattern = strings.ReplaceAll(regexPattern, "?", ".")
		regexPattern = "^" + regexPattern + "$"

		matched, err := regexp.MatchString(regexPattern, dbName)
		if err != nil {
			return false
		}
		return matched
	}

	// 完全匹配
	return dbName == pattern
}

// isDatabaseInList 检查数据库是否已在列表中
func (m *Manager) isDatabaseInList(dbName string, databases []types.DatabaseInfo) bool {
	for _, db := range databases {
		if db.Name == dbName {
			return true
		}
	}
	return false
}

// discoverFromServer 从数据库服务器发现数据库
func (m *Manager) discoverFromServer(ctx context.Context, patterns []string) ([]types.DatabaseInfo, error) {
	// 使用默认连接查询所有数据库
	db, err := NewMySQLDB(m.baseConfig)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// 查询所有数据库
	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []types.DatabaseInfo
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			continue
		}

		// 跳过系统数据库
		if m.isSystemDatabase(dbName) {
			continue
		}

		// 检查是否匹配模式
		if m.matchesPatterns(dbName, patterns) {
			databases = append(databases, types.DatabaseInfo{
				Name:        dbName,
				ConfigKey:   "discovered",
				Description: "从服务器发现的数据库",
				Matched:     true,
			})
		}
	}

	return databases, nil
}

// isSystemDatabase 检查是否是系统数据库
func (m *Manager) isSystemDatabase(dbName string) bool {
	systemDbs := []string{"information_schema", "mysql", "performance_schema", "sys"}
	for _, sysDb := range systemDbs {
		if dbName == sysDb {
			return true
		}
	}
	return false
}
