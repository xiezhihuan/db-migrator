package builder

import (
	"context"
	"fmt"
	"strings"

	"db-migrator/internal/types"
)

// SQLBuilder SQL构建器
type SQLBuilder struct {
	checker types.Checker
	db      types.DB
}

// NewSQLBuilder 创建SQL构建器
func NewSQLBuilder(checker types.Checker, db types.DB) *SQLBuilder {
	return &SQLBuilder{
		checker: checker,
		db:      db,
	}
}

// CreateTableIfNotExists 智能创建表
func (b *SQLBuilder) CreateTableIfNotExists(ctx context.Context, tableName, tableSQL string) error {
	exists, err := b.checker.TableExists(ctx, tableName)
	if err != nil {
		return fmt.Errorf("检查表 %s 是否存在失败: %v", tableName, err)
	}

	if exists {
		fmt.Printf("表 %s 已存在，跳过创建\n", tableName)
		return nil
	}

	_, err = b.db.Exec(tableSQL)
	if err != nil {
		return fmt.Errorf("创建表 %s 失败: %v", tableName, err)
	}

	fmt.Printf("成功创建表: %s\n", tableName)
	return nil
}

// AddColumnIfNotExists 智能添加列
func (b *SQLBuilder) AddColumnIfNotExists(ctx context.Context, tableName, columnName, columnDef string) error {
	exists, err := b.checker.ColumnExists(ctx, tableName, columnName)
	if err != nil {
		return fmt.Errorf("检查列 %s.%s 是否存在失败: %v", tableName, columnName, err)
	}

	if exists {
		fmt.Printf("列 %s.%s 已存在，跳过添加\n", tableName, columnName)
		return nil
	}

	sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnDef)
	_, err = b.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("添加列 %s.%s 失败: %v", tableName, columnName, err)
	}

	fmt.Printf("成功添加列: %s.%s\n", tableName, columnName)
	return nil
}

// DropColumnIfExists 智能删除列
func (b *SQLBuilder) DropColumnIfExists(ctx context.Context, tableName, columnName string) error {
	exists, err := b.checker.ColumnExists(ctx, tableName, columnName)
	if err != nil {
		return fmt.Errorf("检查列 %s.%s 是否存在失败: %v", tableName, columnName, err)
	}

	if !exists {
		fmt.Printf("列 %s.%s 不存在，跳过删除\n", tableName, columnName)
		return nil
	}

	sql := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, columnName)
	_, err = b.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("删除列 %s.%s 失败: %v", tableName, columnName, err)
	}

	fmt.Printf("成功删除列: %s.%s\n", tableName, columnName)
	return nil
}

// CreateIndexIfNotExists 智能创建索引
func (b *SQLBuilder) CreateIndexIfNotExists(ctx context.Context, tableName, indexName, indexSQL string) error {
	exists, err := b.checker.IndexExists(ctx, tableName, indexName)
	if err != nil {
		return fmt.Errorf("检查索引 %s 是否存在失败: %v", indexName, err)
	}

	if exists {
		fmt.Printf("索引 %s 已存在，跳过创建\n", indexName)
		return nil
	}

	_, err = b.db.Exec(indexSQL)
	if err != nil {
		return fmt.Errorf("创建索引 %s 失败: %v", indexName, err)
	}

	fmt.Printf("成功创建索引: %s\n", indexName)
	return nil
}

// DropIndexIfExists 智能删除索引
func (b *SQLBuilder) DropIndexIfExists(ctx context.Context, tableName, indexName string) error {
	exists, err := b.checker.IndexExists(ctx, tableName, indexName)
	if err != nil {
		return fmt.Errorf("检查索引 %s 是否存在失败: %v", indexName, err)
	}

	if !exists {
		fmt.Printf("索引 %s 不存在，跳过删除\n", indexName)
		return nil
	}

	sql := fmt.Sprintf("DROP INDEX %s ON %s", indexName, tableName)
	_, err = b.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("删除索引 %s 失败: %v", indexName, err)
	}

	fmt.Printf("成功删除索引: %s\n", indexName)
	return nil
}

// CreateFunctionIfNotExists 智能创建函数
func (b *SQLBuilder) CreateFunctionIfNotExists(ctx context.Context, functionName, functionSQL string) error {
	exists, err := b.checker.FunctionExists(ctx, functionName)
	if err != nil {
		return fmt.Errorf("检查函数 %s 是否存在失败: %v", functionName, err)
	}

	if exists {
		// 函数存在，先删除再创建（用于更新函数）
		dropSQL := fmt.Sprintf("DROP FUNCTION IF EXISTS %s", functionName)
		_, err = b.db.Exec(dropSQL)
		if err != nil {
			return fmt.Errorf("删除函数 %s 失败: %v", functionName, err)
		}
		fmt.Printf("删除已存在的函数: %s\n", functionName)
	}

	_, err = b.db.Exec(functionSQL)
	if err != nil {
		return fmt.Errorf("创建函数 %s 失败: %v", functionName, err)
	}

	fmt.Printf("成功创建函数: %s\n", functionName)
	return nil
}

// DropFunctionIfExists 智能删除函数
func (b *SQLBuilder) DropFunctionIfExists(ctx context.Context, functionName string) error {
	exists, err := b.checker.FunctionExists(ctx, functionName)
	if err != nil {
		return fmt.Errorf("检查函数 %s 是否存在失败: %v", functionName, err)
	}

	if !exists {
		fmt.Printf("函数 %s 不存在，跳过删除\n", functionName)
		return nil
	}

	sql := fmt.Sprintf("DROP FUNCTION %s", functionName)
	_, err = b.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("删除函数 %s 失败: %v", functionName, err)
	}

	fmt.Printf("成功删除函数: %s\n", functionName)
	return nil
}

// InsertIfNotExists 智能插入数据
func (b *SQLBuilder) InsertIfNotExists(ctx context.Context, tableName string, whereCondition string, insertSQL string) error {
	// 检查数据是否存在
	checkSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", tableName, whereCondition)
	var count int
	err := b.db.QueryRow(checkSQL).Scan(&count)
	if err != nil {
		return fmt.Errorf("检查数据是否存在失败: %v", err)
	}

	if count > 0 {
		fmt.Printf("数据已存在（条件: %s），跳过插入\n", whereCondition)
		return nil
	}

	_, err = b.db.Exec(insertSQL)
	if err != nil {
		return fmt.Errorf("插入数据失败: %v", err)
	}

	fmt.Printf("成功插入数据到表: %s\n", tableName)
	return nil
}

// UpdateIfExists 智能更新数据
func (b *SQLBuilder) UpdateIfExists(ctx context.Context, tableName string, whereCondition string, updateSQL string) error {
	// 检查数据是否存在
	checkSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", tableName, whereCondition)
	var count int
	err := b.db.QueryRow(checkSQL).Scan(&count)
	if err != nil {
		return fmt.Errorf("检查数据是否存在失败: %v", err)
	}

	if count == 0 {
		fmt.Printf("数据不存在（条件: %s），跳过更新\n", whereCondition)
		return nil
	}

	_, err = b.db.Exec(updateSQL)
	if err != nil {
		return fmt.Errorf("更新数据失败: %v", err)
	}

	fmt.Printf("成功更新数据（条件: %s）\n", whereCondition)
	return nil
}

// ExecuteRawSQL 执行原始SQL
func (b *SQLBuilder) ExecuteRawSQL(ctx context.Context, sql string, description string) error {
	if description != "" {
		fmt.Printf("执行: %s\n", description)
	}

	_, err := b.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("执行SQL失败: %v", err)
	}

	if description != "" {
		fmt.Printf("完成: %s\n", description)
	}
	return nil
}

// Helper functions for building common SQL patterns

// BuildCreateTableSQL 构建创建表的SQL
func BuildCreateTableSQL(tableName string, columns []ColumnDef, options ...TableOption) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("CREATE TABLE %s (", tableName))

	// 添加列定义
	columnStrs := make([]string, len(columns))
	for i, col := range columns {
		columnStrs[i] = "  " + col.String()
	}
	parts = append(parts, strings.Join(columnStrs, ",\n"))

	parts = append(parts, ")")

	// 添加表选项
	for _, opt := range options {
		parts = append(parts, opt.String())
	}

	return strings.Join(parts, "\n")
}

// ColumnDef 列定义
type ColumnDef struct {
	Name       string
	Type       string
	NotNull    bool
	Default    string
	AutoIncr   bool
	PrimaryKey bool
	Comment    string
}

func (c ColumnDef) String() string {
	var parts []string
	parts = append(parts, c.Name, c.Type)

	if c.NotNull {
		parts = append(parts, "NOT NULL")
	}

	if c.Default != "" {
		parts = append(parts, "DEFAULT", c.Default)
	}

	if c.AutoIncr {
		parts = append(parts, "AUTO_INCREMENT")
	}

	if c.PrimaryKey {
		parts = append(parts, "PRIMARY KEY")
	}

	if c.Comment != "" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", c.Comment))
	}

	return strings.Join(parts, " ")
}

// TableOption 表选项
type TableOption interface {
	String() string
}

// EngineOption 引擎选项
type EngineOption string

func (e EngineOption) String() string {
	return fmt.Sprintf("ENGINE=%s", string(e))
}

// CharsetOption 字符集选项
type CharsetOption string

func (c CharsetOption) String() string {
	return fmt.Sprintf("DEFAULT CHARSET=%s", string(c))
}

// CommentOption 表注释选项
type CommentOption string

func (c CommentOption) String() string {
	return fmt.Sprintf("COMMENT='%s'", string(c))
}
