package builder

import (
	"context"
	"fmt"
	"strings"

	"github.com/xiezhihuan/db-migrator/internal/types"
)

// AdvancedBuilder 高级构建器，提供更多封装方法
type AdvancedBuilder struct {
	sqlBuilder *SQLBuilder
	checker    types.Checker
	db         types.DB
}

// NewAdvancedBuilder 创建高级构建器
func NewAdvancedBuilder(checker types.Checker, db types.DB) *AdvancedBuilder {
	return &AdvancedBuilder{
		sqlBuilder: NewSQLBuilder(checker, db),
		checker:    checker,
		db:         db,
	}
}

// Table 创建表构建器
func (ab *AdvancedBuilder) Table(name string) *TableBuilder {
	return NewTableBuilder(ab.sqlBuilder, name)
}

// ModifyTable 修改表结构
func (ab *AdvancedBuilder) ModifyTable(name string) *TableModifier {
	return &TableModifier{
		advancedBuilder: ab,
		tableName:       name,
	}
}

// CreateView 创建视图
func (ab *AdvancedBuilder) CreateView(ctx context.Context, viewName, query string) error {
	// 检查视图是否存在
	exists, err := ab.checkViewExists(ctx, viewName)
	if err != nil {
		return fmt.Errorf("检查视图 %s 是否存在失败: %v", viewName, err)
	}

	if exists {
		fmt.Printf("视图 %s 已存在，跳过创建\n", viewName)
		return nil
	}

	sql := fmt.Sprintf("CREATE VIEW %s AS %s", viewName, query)
	_, err = ab.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("创建视图 %s 失败: %v", viewName, err)
	}

	fmt.Printf("成功创建视图: %s\n", viewName)
	return nil
}

// DropView 删除视图
func (ab *AdvancedBuilder) DropView(ctx context.Context, viewName string) error {
	exists, err := ab.checkViewExists(ctx, viewName)
	if err != nil {
		return fmt.Errorf("检查视图 %s 是否存在失败: %v", viewName, err)
	}

	if !exists {
		fmt.Printf("视图 %s 不存在，跳过删除\n", viewName)
		return nil
	}

	sql := fmt.Sprintf("DROP VIEW %s", viewName)
	_, err = ab.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("删除视图 %s 失败: %v", viewName, err)
	}

	fmt.Printf("成功删除视图: %s\n", viewName)
	return nil
}

// RenameTable 重命名表
func (ab *AdvancedBuilder) RenameTable(ctx context.Context, oldName, newName string) error {
	exists, err := ab.checker.TableExists(ctx, oldName)
	if err != nil {
		return fmt.Errorf("检查表 %s 是否存在失败: %v", oldName, err)
	}

	if !exists {
		return fmt.Errorf("表 %s 不存在，无法重命名", oldName)
	}

	// 检查新表名是否已存在
	newExists, err := ab.checker.TableExists(ctx, newName)
	if err != nil {
		return fmt.Errorf("检查表 %s 是否存在失败: %v", newName, err)
	}

	if newExists {
		return fmt.Errorf("表 %s 已存在，无法重命名", newName)
	}

	sql := fmt.Sprintf("RENAME TABLE %s TO %s", oldName, newName)
	_, err = ab.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("重命名表失败: %v", err)
	}

	fmt.Printf("成功将表 %s 重命名为 %s\n", oldName, newName)
	return nil
}

// CopyTable 复制表结构（可选择是否复制数据）
func (ab *AdvancedBuilder) CopyTable(ctx context.Context, srcTable, destTable string, copyData bool) error {
	exists, err := ab.checker.TableExists(ctx, srcTable)
	if err != nil {
		return fmt.Errorf("检查源表 %s 是否存在失败: %v", srcTable, err)
	}

	if !exists {
		return fmt.Errorf("源表 %s 不存在", srcTable)
	}

	// 检查目标表是否已存在
	destExists, err := ab.checker.TableExists(ctx, destTable)
	if err != nil {
		return fmt.Errorf("检查目标表 %s 是否存在失败: %v", destTable, err)
	}

	if destExists {
		fmt.Printf("目标表 %s 已存在，跳过复制\n", destTable)
		return nil
	}

	// 复制表结构
	var sql string
	if copyData {
		sql = fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM %s", destTable, srcTable)
	} else {
		sql = fmt.Sprintf("CREATE TABLE %s LIKE %s", destTable, srcTable)
	}

	_, err = ab.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("复制表失败: %v", err)
	}

	action := "结构"
	if copyData {
		action = "结构和数据"
	}
	fmt.Printf("成功复制表 %s 的%s到 %s\n", srcTable, action, destTable)
	return nil
}

// TruncateTable 清空表数据
func (ab *AdvancedBuilder) TruncateTable(ctx context.Context, tableName string) error {
	exists, err := ab.checker.TableExists(ctx, tableName)
	if err != nil {
		return fmt.Errorf("检查表 %s 是否存在失败: %v", tableName, err)
	}

	if !exists {
		return fmt.Errorf("表 %s 不存在", tableName)
	}

	sql := fmt.Sprintf("TRUNCATE TABLE %s", tableName)
	_, err = ab.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("清空表 %s 失败: %v", tableName, err)
	}

	fmt.Printf("成功清空表: %s\n", tableName)
	return nil
}

// CreateStoredProcedure 创建存储过程
func (ab *AdvancedBuilder) CreateStoredProcedure(ctx context.Context, name, body string) error {
	// 检查存储过程是否存在
	exists, err := ab.checkProcedureExists(ctx, name)
	if err != nil {
		return fmt.Errorf("检查存储过程 %s 是否存在失败: %v", name, err)
	}

	if exists {
		// 先删除再创建
		dropSQL := fmt.Sprintf("DROP PROCEDURE IF EXISTS %s", name)
		_, err = ab.db.Exec(dropSQL)
		if err != nil {
			return fmt.Errorf("删除存储过程 %s 失败: %v", name, err)
		}
		fmt.Printf("删除已存在的存储过程: %s\n", name)
	}

	_, err = ab.db.Exec(body)
	if err != nil {
		return fmt.Errorf("创建存储过程 %s 失败: %v", name, err)
	}

	fmt.Printf("成功创建存储过程: %s\n", name)
	return nil
}

// CreateTrigger 创建触发器
func (ab *AdvancedBuilder) CreateTrigger(ctx context.Context, name, body string) error {
	// 检查触发器是否存在
	exists, err := ab.checkTriggerExists(ctx, name)
	if err != nil {
		return fmt.Errorf("检查触发器 %s 是否存在失败: %v", name, err)
	}

	if exists {
		// 先删除再创建
		dropSQL := fmt.Sprintf("DROP TRIGGER IF EXISTS %s", name)
		_, err = ab.db.Exec(dropSQL)
		if err != nil {
			return fmt.Errorf("删除触发器 %s 失败: %v", name, err)
		}
		fmt.Printf("删除已存在的触发器: %s\n", name)
	}

	_, err = ab.db.Exec(body)
	if err != nil {
		return fmt.Errorf("创建触发器 %s 失败: %v", name, err)
	}

	fmt.Printf("成功创建触发器: %s\n", name)
	return nil
}

// BulkInsert 批量插入数据
func (ab *AdvancedBuilder) BulkInsert(ctx context.Context, tableName string, columns []string, data [][]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	exists, err := ab.checker.TableExists(ctx, tableName)
	if err != nil {
		return fmt.Errorf("检查表 %s 是否存在失败: %v", tableName, err)
	}

	if !exists {
		return fmt.Errorf("表 %s 不存在", tableName)
	}

	// 构建批量插入SQL
	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = "?"
	}

	valueTemplate := fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))

	// 分批插入以避免SQL过长
	batchSize := 1000
	for i := 0; i < len(data); i += batchSize {
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}

		batch := data[i:end]
		values := make([]string, len(batch))
		var args []interface{}

		for j, row := range batch {
			values[j] = valueTemplate
			args = append(args, row...)
		}

		sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
			tableName,
			strings.Join(columns, ", "),
			strings.Join(values, ", "))

		_, err = ab.db.Exec(sql, args...)
		if err != nil {
			return fmt.Errorf("批量插入数据失败: %v", err)
		}
	}

	fmt.Printf("成功批量插入 %d 条记录到表 %s\n", len(data), tableName)
	return nil
}

// TableModifier 表修改器
type TableModifier struct {
	advancedBuilder *AdvancedBuilder
	tableName       string
}

// AddColumn 添加列
func (tm *TableModifier) AddColumn(name string, columnType ColumnType, size int) *ColumnModifier {
	return &ColumnModifier{
		tableModifier: tm,
		operation:     "ADD",
		column: AdvancedColumn{
			Name: name,
			Type: columnType,
			Size: size,
		},
	}
}

// ModifyColumn 修改列
func (tm *TableModifier) ModifyColumn(name string, columnType ColumnType, size int) *ColumnModifier {
	return &ColumnModifier{
		tableModifier: tm,
		operation:     "MODIFY",
		column: AdvancedColumn{
			Name: name,
			Type: columnType,
			Size: size,
		},
	}
}

// DropColumn 删除列
func (tm *TableModifier) DropColumn(ctx context.Context, columnName string) error {
	return tm.advancedBuilder.sqlBuilder.DropColumnIfExists(ctx, tm.tableName, columnName)
}

// RenameColumn 重命名列
func (tm *TableModifier) RenameColumn(ctx context.Context, oldName, newName, columnDef string) error {
	exists, err := tm.advancedBuilder.checker.ColumnExists(ctx, tm.tableName, oldName)
	if err != nil {
		return fmt.Errorf("检查列 %s.%s 是否存在失败: %v", tm.tableName, oldName, err)
	}

	if !exists {
		fmt.Printf("列 %s.%s 不存在，跳过重命名\n", tm.tableName, oldName)
		return nil
	}

	sql := fmt.Sprintf("ALTER TABLE %s CHANGE %s %s %s",
		tm.tableName, oldName, newName, columnDef)
	_, err = tm.advancedBuilder.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("重命名列失败: %v", err)
	}

	fmt.Printf("成功将列 %s.%s 重命名为 %s\n", tm.tableName, oldName, newName)
	return nil
}

// AddIndex 添加索引
func (tm *TableModifier) AddIndex(ctx context.Context, indexName string, columns []string, unique bool) error {
	exists, err := tm.advancedBuilder.checker.IndexExists(ctx, tm.tableName, indexName)
	if err != nil {
		return fmt.Errorf("检查索引 %s 是否存在失败: %v", indexName, err)
	}

	if exists {
		fmt.Printf("索引 %s 已存在，跳过创建\n", indexName)
		return nil
	}

	var sql string
	if unique {
		sql = fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)",
			indexName, tm.tableName, strings.Join(columns, ", "))
	} else {
		sql = fmt.Sprintf("CREATE INDEX %s ON %s (%s)",
			indexName, tm.tableName, strings.Join(columns, ", "))
	}

	_, err = tm.advancedBuilder.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("添加索引 %s 失败: %v", indexName, err)
	}

	fmt.Printf("成功添加索引: %s\n", indexName)
	return nil
}

// ColumnModifier 列修改器
type ColumnModifier struct {
	tableModifier *TableModifier
	operation     string
	column        AdvancedColumn
}

// NotNull 设置为非空
func (cm *ColumnModifier) NotNull() *ColumnModifier {
	cm.column.NotNull = true
	return cm
}

// Default 设置默认值
func (cm *ColumnModifier) Default(value interface{}) *ColumnModifier {
	cm.column.Default = value
	return cm
}

// Comment 设置注释
func (cm *ColumnModifier) Comment(comment string) *ColumnModifier {
	cm.column.Comment = comment
	return cm
}

// After 在指定列之后添加
func (cm *ColumnModifier) After(columnName string) *ColumnModifier {
	cm.column.After = columnName
	return cm
}

// Execute 执行列修改
func (cm *ColumnModifier) Execute(ctx context.Context) error {
	if cm.operation == "ADD" {
		return cm.executeAddColumn(ctx)
	} else if cm.operation == "MODIFY" {
		return cm.executeModifyColumn(ctx)
	}
	return fmt.Errorf("未知的操作类型: %s", cm.operation)
}

// executeAddColumn 执行添加列
func (cm *ColumnModifier) executeAddColumn(ctx context.Context) error {
	exists, err := cm.tableModifier.advancedBuilder.checker.ColumnExists(
		ctx, cm.tableModifier.tableName, cm.column.Name)
	if err != nil {
		return fmt.Errorf("检查列是否存在失败: %v", err)
	}

	if exists {
		fmt.Printf("列 %s.%s 已存在，跳过添加\n",
			cm.tableModifier.tableName, cm.column.Name)
		return nil
	}

	columnDef := cm.buildColumnDefinition()
	sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s",
		cm.tableModifier.tableName, columnDef)

	if cm.column.After != "" {
		sql += fmt.Sprintf(" AFTER %s", cm.column.After)
	}

	_, err = cm.tableModifier.advancedBuilder.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("添加列失败: %v", err)
	}

	fmt.Printf("成功添加列: %s.%s\n",
		cm.tableModifier.tableName, cm.column.Name)
	return nil
}

// executeModifyColumn 执行修改列
func (cm *ColumnModifier) executeModifyColumn(ctx context.Context) error {
	exists, err := cm.tableModifier.advancedBuilder.checker.ColumnExists(
		ctx, cm.tableModifier.tableName, cm.column.Name)
	if err != nil {
		return fmt.Errorf("检查列是否存在失败: %v", err)
	}

	if !exists {
		return fmt.Errorf("列 %s.%s 不存在，无法修改",
			cm.tableModifier.tableName, cm.column.Name)
	}

	columnDef := cm.buildColumnDefinition()
	sql := fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s",
		cm.tableModifier.tableName, columnDef)

	_, err = cm.tableModifier.advancedBuilder.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("修改列失败: %v", err)
	}

	fmt.Printf("成功修改列: %s.%s\n",
		cm.tableModifier.tableName, cm.column.Name)
	return nil
}

// buildColumnDefinition 构建列定义
func (cm *ColumnModifier) buildColumnDefinition() string {
	var parts []string
	parts = append(parts, cm.column.Name)

	// 构建类型定义
	typeStr := string(cm.column.Type)
	if cm.column.Size > 0 {
		typeStr = fmt.Sprintf("%s(%d)", typeStr, cm.column.Size)
	}
	parts = append(parts, typeStr)

	// 添加约束
	if cm.column.NotNull {
		parts = append(parts, "NOT NULL")
	}

	// 添加默认值
	if cm.column.Default != nil {
		switch v := cm.column.Default.(type) {
		case string:
			if v == "CURRENT_TIMESTAMP" || strings.Contains(v, "CURRENT_TIMESTAMP") {
				parts = append(parts, fmt.Sprintf("DEFAULT %s", v))
			} else {
				parts = append(parts, fmt.Sprintf("DEFAULT '%s'", v))
			}
		case int, int64, float64:
			parts = append(parts, fmt.Sprintf("DEFAULT %v", v))
		}
	}

	if cm.column.Comment != "" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", cm.column.Comment))
	}

	return strings.Join(parts, " ")
}

// 辅助方法

func (ab *AdvancedBuilder) checkViewExists(ctx context.Context, viewName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.VIEWS 
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
	`
	var count int
	err := ab.db.QueryRow(query, viewName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (ab *AdvancedBuilder) checkProcedureExists(ctx context.Context, procName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.ROUTINES 
		WHERE ROUTINE_SCHEMA = DATABASE() AND ROUTINE_NAME = ? AND ROUTINE_TYPE = 'PROCEDURE'
	`
	var count int
	err := ab.db.QueryRow(query, procName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (ab *AdvancedBuilder) checkTriggerExists(ctx context.Context, triggerName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.TRIGGERS 
		WHERE TRIGGER_SCHEMA = DATABASE() AND TRIGGER_NAME = ?
	`
	var count int
	err := ab.db.QueryRow(query, triggerName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
