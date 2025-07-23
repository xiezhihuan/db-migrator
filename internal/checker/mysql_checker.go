package checker

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"db-migrator/internal/types"
)

// MySQLChecker MySQL存在性检查器
type MySQLChecker struct {
	db       types.DB
	database string
}

// NewMySQLChecker 创建MySQL检查器
func NewMySQLChecker(db types.DB, database string) *MySQLChecker {
	return &MySQLChecker{
		db:       db,
		database: database,
	}
}

// TableExists 检查表是否存在
func (c *MySQLChecker) TableExists(ctx context.Context, tableName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
	`

	var count int
	err := c.db.QueryRow(query, c.database, tableName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查表 %s 是否存在失败: %v", tableName, err)
	}

	return count > 0, nil
}

// ColumnExists 检查列是否存在
func (c *MySQLChecker) ColumnExists(ctx context.Context, tableName, columnName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND COLUMN_NAME = ?
	`

	var count int
	err := c.db.QueryRow(query, c.database, tableName, columnName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查列 %s.%s 是否存在失败: %v", tableName, columnName, err)
	}

	return count > 0, nil
}

// IndexExists 检查索引是否存在
func (c *MySQLChecker) IndexExists(ctx context.Context, tableName, indexName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.STATISTICS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND INDEX_NAME = ?
	`

	var count int
	err := c.db.QueryRow(query, c.database, tableName, indexName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查索引 %s 是否存在失败: %v", indexName, err)
	}

	return count > 0, nil
}

// FunctionExists 检查函数是否存在
func (c *MySQLChecker) FunctionExists(ctx context.Context, functionName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.ROUTINES 
		WHERE ROUTINE_SCHEMA = ? AND ROUTINE_NAME = ? AND ROUTINE_TYPE = 'FUNCTION'
	`

	var count int
	err := c.db.QueryRow(query, c.database, functionName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查函数 %s 是否存在失败: %v", functionName, err)
	}

	return count > 0, nil
}

// ProcedureExists 检查存储过程是否存在
func (c *MySQLChecker) ProcedureExists(ctx context.Context, procedureName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.ROUTINES 
		WHERE ROUTINE_SCHEMA = ? AND ROUTINE_NAME = ? AND ROUTINE_TYPE = 'PROCEDURE'
	`

	var count int
	err := c.db.QueryRow(query, c.database, procedureName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查存储过程 %s 是否存在失败: %v", procedureName, err)
	}

	return count > 0, nil
}

// ConstraintExists 检查约束是否存在
func (c *MySQLChecker) ConstraintExists(ctx context.Context, tableName, constraintName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.TABLE_CONSTRAINTS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND CONSTRAINT_NAME = ?
	`

	var count int
	err := c.db.QueryRow(query, c.database, tableName, constraintName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查约束 %s 是否存在失败: %v", constraintName, err)
	}

	return count > 0, nil
}

// TriggerExists 检查触发器是否存在
func (c *MySQLChecker) TriggerExists(ctx context.Context, triggerName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.TRIGGERS 
		WHERE TRIGGER_SCHEMA = ? AND TRIGGER_NAME = ?
	`

	var count int
	err := c.db.QueryRow(query, c.database, triggerName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查触发器 %s 是否存在失败: %v", triggerName, err)
	}

	return count > 0, nil
}

// GetTableColumns 获取表的所有列信息
func (c *MySQLChecker) GetTableColumns(ctx context.Context, tableName string) ([]ColumnInfo, error) {
	query := `
		SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, EXTRA, COLUMN_COMMENT
		FROM information_schema.COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := c.db.Query(query, c.database, tableName)
	if err != nil {
		return nil, fmt.Errorf("获取表 %s 列信息失败: %v", tableName, err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var nullable, defaultValue, extra, comment sql.NullString

		err := rows.Scan(&col.Name, &col.Type, &nullable, &defaultValue, &extra, &comment)
		if err != nil {
			return nil, fmt.Errorf("扫描列信息失败: %v", err)
		}

		col.Nullable = nullable.String == "YES"
		col.Default = defaultValue.String
		col.Extra = extra.String
		col.Comment = comment.String

		columns = append(columns, col)
	}

	return columns, nil
}

// ColumnInfo 列信息结构
type ColumnInfo struct {
	Name     string
	Type     string
	Nullable bool
	Default  string
	Extra    string
	Comment  string
}

// CompareColumns 比较两个列定义是否相同
func (c *MySQLChecker) CompareColumns(expected, actual ColumnInfo) bool {
	// 标准化类型名称
	expectedType := strings.ToLower(strings.TrimSpace(expected.Type))
	actualType := strings.ToLower(strings.TrimSpace(actual.Type))

	return expectedType == actualType &&
		expected.Nullable == actual.Nullable &&
		strings.TrimSpace(expected.Default) == strings.TrimSpace(actual.Default)
}
