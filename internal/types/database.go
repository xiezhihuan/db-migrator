package types

import (
	"context"
)

// DatabaseCreator 数据库创建器接口
type DatabaseCreator interface {
	CreateDatabase(ctx context.Context, config DatabaseCreateConfig) error
	DatabaseExists(ctx context.Context, name string) (bool, error)
	ExecuteSQLFile(ctx context.Context, dbName, filePath string) error
}

// DatabaseCreateConfig 数据库创建配置
type DatabaseCreateConfig struct {
	Name      string `yaml:"name" json:"name"`
	Charset   string `yaml:"charset" json:"charset"`
	Collation string `yaml:"collation" json:"collation"`
	IfExists  string `yaml:"if_exists" json:"if_exists"` // "error", "skip", "prompt"
}

// SQLStatement SQL语句结构
type SQLStatement struct {
	Type         string   // CREATE_TABLE, CREATE_VIEW, CREATE_PROCEDURE, etc.
	Name         string   // 对象名称
	Statement    string   // 完整SQL语句
	Dependencies []string // 依赖的对象名称
}

// SQLParser SQL解析器接口
type SQLParser interface {
	ParseFile(filePath string) ([]SQLStatement, error)
	ValidateStatements(statements []SQLStatement) error
	SortByDependencies(statements []SQLStatement) ([]SQLStatement, error)
}

// CreateFromSQLResult 从SQL创建的结果
type CreateFromSQLResult struct {
	DatabaseName      string       `json:"database_name"`
	DatabaseCreated   bool         `json:"database_created"`
	StatementsTotal   int          `json:"statements_total"`
	StatementsSuccess int          `json:"statements_success"`
	StatementsFailed  int          `json:"statements_failed"`
	ExecutionTime     string       `json:"execution_time"`
	Errors            []string     `json:"errors,omitempty"`
	CreatedObjects    []ObjectInfo `json:"created_objects"`
}

// ObjectInfo 创建的对象信息
type ObjectInfo struct {
	Type string `json:"type"`
	Name string `json:"name"`
}
