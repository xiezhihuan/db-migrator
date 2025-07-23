package types

import (
	"context"
)

// DataInserter 数据插入器接口
type DataInserter interface {
	InsertFromSQLFile(ctx context.Context, dbName, filePath string, config DataInsertConfig) (*DataInsertResult, error)
	ValidateTablesExist(ctx context.Context, dbName string, tables []string) error
	ExecuteInsertStatements(ctx context.Context, dbName string, statements []InsertStatement, config DataInsertConfig) error
}

// DataInsertConfig 数据插入配置
type DataInsertConfig struct {
	BatchSize        int              `yaml:"batch_size" json:"batch_size"`           // 批量插入大小
	OnConflict       string           `yaml:"on_conflict" json:"on_conflict"`         // 冲突处理: error, ignore, replace
	StopOnError      bool             `yaml:"stop_on_error" json:"stop_on_error"`     // 遇错停止
	ValidateTables   bool             `yaml:"validate_tables" json:"validate_tables"` // 验证表存在
	UseTransaction   bool             `yaml:"use_transaction" json:"use_transaction"` // 使用事务
	ProgressCallback ProgressCallback `yaml:"-" json:"-"`                             // 进度回调
}

// InsertStatement INSERT语句结构
type InsertStatement struct {
	TableName  string          `json:"table_name"`  // 目标表名
	Columns    []string        `json:"columns"`     // 列名列表
	Values     [][]interface{} `json:"values"`      // 值列表
	Statement  string          `json:"statement"`   // 原始SQL语句
	LineNumber int             `json:"line_number"` // 源文件行号
}

// DataInsertResult 数据插入结果
type DataInsertResult struct {
	DatabaseName         string              `json:"database_name"`
	TotalStatements      int                 `json:"total_statements"`
	SuccessfulStatements int                 `json:"successful_statements"`
	FailedStatements     int                 `json:"failed_statements"`
	TotalRowsInserted    int64               `json:"total_rows_inserted"`
	ExecutionTime        string              `json:"execution_time"`
	TableResults         []TableInsertResult `json:"table_results"`
	Errors               []InsertError       `json:"errors,omitempty"`
}

// TableInsertResult 表插入结果
type TableInsertResult struct {
	TableName          string `json:"table_name"`
	RowsInserted       int64  `json:"rows_inserted"`
	StatementsExecuted int    `json:"statements_executed"`
}

// InsertError 插入错误
type InsertError struct {
	TableName    string `json:"table_name"`
	Statement    string `json:"statement"`
	LineNumber   int    `json:"line_number"`
	ErrorMessage string `json:"error_message"`
}

// ProgressCallback 进度回调函数
type ProgressCallback func(stage string, current, total int64, err error)

// InsertSQLParser INSERT SQL解析器接口
type InsertSQLParser interface {
	ParseInsertFile(filePath string) ([]InsertStatement, error)
	ValidateInsertStatements(statements []InsertStatement) error
	ExtractTableNames(statements []InsertStatement) []string
}

// MultiDatabaseInsertResult 多数据库插入结果
type MultiDatabaseInsertResult struct {
	TotalDatabases      int                `json:"total_databases"`
	SuccessfulDatabases int                `json:"successful_databases"`
	FailedDatabases     int                `json:"failed_databases"`
	DatabaseResults     []DataInsertResult `json:"database_results"`
	ExecutionTime       string             `json:"execution_time"`
	Errors              []string           `json:"errors,omitempty"`
}
