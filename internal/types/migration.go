package types

import (
	"context"
	"database/sql"
	"time"
)

// Migration 定义迁移接口
type Migration interface {
	// Version 返回迁移版本号
	Version() string
	// Description 返回迁移描述
	Description() string
	// Up 执行向上迁移
	Up(ctx context.Context, db DB) error
	// Down 执行向下迁移（回滚）
	Down(ctx context.Context, db DB) error
}

// MultiDatabaseMigration 多数据库迁移接口
type MultiDatabaseMigration interface {
	Migration
	// Database 返回目标数据库名，如果返回空字符串则使用默认数据库
	Database() string
	// Databases 返回目标数据库列表，支持同时操作多个数据库
	Databases() []string
}

// DB 数据库操作接口
type DB interface {
	// Exec 执行SQL语句
	Exec(query string, args ...interface{}) (sql.Result, error)
	// Query 查询数据
	Query(query string, args ...interface{}) (*sql.Rows, error)
	// QueryRow 查询单行数据
	QueryRow(query string, args ...interface{}) *sql.Row
	// Begin 开始事务
	Begin() (*sql.Tx, error)
	// Close 关闭连接
	Close() error
}

// Checker 存在性检查器接口
type Checker interface {
	// TableExists 检查表是否存在
	TableExists(ctx context.Context, tableName string) (bool, error)
	// ColumnExists 检查列是否存在
	ColumnExists(ctx context.Context, tableName, columnName string) (bool, error)
	// IndexExists 检查索引是否存在
	IndexExists(ctx context.Context, tableName, indexName string) (bool, error)
	// FunctionExists 检查函数是否存在
	FunctionExists(ctx context.Context, functionName string) (bool, error)
	// ConstraintExists 检查约束是否存在
	ConstraintExists(ctx context.Context, tableName, constraintName string) (bool, error)
}

// MigrationRecord 迁移记录
type MigrationRecord struct {
	Version     string    `json:"version"`
	Description string    `json:"description"`
	AppliedAt   time.Time `json:"applied_at"`
	Success     bool      `json:"success"`
	ErrorMsg    string    `json:"error_msg,omitempty"`
}

// MigrationStatus 迁移状态
type MigrationStatus struct {
	Version     string     `json:"version"`
	Description string     `json:"description"`
	Applied     bool       `json:"applied"`
	AppliedAt   *time.Time `json:"applied_at,omitempty"`
	Database    string     `json:"database,omitempty"` // 所属数据库
}

// DatabaseInfo 数据库信息
type DatabaseInfo struct {
	Name        string `json:"name"`        // 数据库名
	ConfigKey   string `json:"config_key"`  // 配置键名
	Description string `json:"description"` // 描述
	Matched     bool   `json:"matched"`     // 是否匹配模式
}

// MultiDatabaseStatus 多数据库迁移状态
type MultiDatabaseStatus struct {
	Database string            `json:"database"`
	Statuses []MigrationStatus `json:"statuses"`
}

// Config 配置结构
type Config struct {
	Database  DatabaseConfig            `yaml:"database"`            // 默认数据库配置（向后兼容）
	Databases map[string]DatabaseConfig `yaml:"databases,omitempty"` // 多数据库配置
	Migrator  MigratorConfig            `yaml:"migrator"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Charset  string `yaml:"charset"`
}

// MigratorConfig 迁移器配置
type MigratorConfig struct {
	MigrationsTable  string   `yaml:"migrations_table"`
	LockTable        string   `yaml:"lock_table"`
	AutoBackup       bool     `yaml:"auto_backup"`
	DryRun           bool     `yaml:"dry_run"`
	DefaultDatabase  string   `yaml:"default_database,omitempty"`  // 默认操作的数据库
	MigrationsDir    string   `yaml:"migrations_dir"`              // 迁移文件目录
	DatabasePatterns []string `yaml:"database_patterns,omitempty"` // 数据库名匹配模式
}

// Error 错误类型
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"cause,omitempty"`
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// 常用错误码
const (
	ErrCodeMigrationFailed    = "MIGRATION_FAILED"
	ErrCodeDatabaseConnection = "DATABASE_CONNECTION"
	ErrCodeConfigInvalid      = "CONFIG_INVALID"
	ErrCodeMigrationNotFound  = "MIGRATION_NOT_FOUND"
	ErrCodeVersionConflict    = "VERSION_CONFLICT"
)

// DBManager 数据库管理器接口
type DBManager interface {
	// GetDatabase 获取指定数据库连接
	GetDatabase(name string) (DB, error)
	// GetDefaultDatabase 获取默认数据库连接
	GetDefaultDatabase() (DB, string, error)
	// DiscoverDatabases 发现数据库（基于模式匹配）
	DiscoverDatabases(ctx context.Context, patterns []string) ([]DatabaseInfo, error)
	// GetMatchedDatabases 获取匹配模式的数据库列表
	GetMatchedDatabases(ctx context.Context, patterns []string) ([]string, error)
	// CloseAll 关闭所有数据库连接
	CloseAll() error
}
