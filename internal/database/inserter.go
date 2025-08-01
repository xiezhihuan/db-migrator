package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/xiezhihuan/db-migrator/internal/sqlparser"
	"github.com/xiezhihuan/db-migrator/internal/types"
)

// Inserter 数据插入器实现
type Inserter struct {
	rootConn *sql.DB
	config   *types.DatabaseConfig
	parser   types.InsertSQLParser
}

// NewInserter 创建新的数据插入器
func NewInserter(rootConn *sql.DB, config *types.DatabaseConfig) *Inserter {
	return &Inserter{
		rootConn: rootConn,
		config:   config,
		parser:   sqlparser.NewInsertParser(),
	}
}

// InsertFromSQLFile 从SQL文件插入数据
func (i *Inserter) InsertFromSQLFile(ctx context.Context, dbName, filePath string, config types.DataInsertConfig) (*types.DataInsertResult, error) {
	startTime := time.Now()

	result := &types.DataInsertResult{
		DatabaseName:         dbName,
		TotalStatements:      0,
		SuccessfulStatements: 0,
		FailedStatements:     0,
		TotalRowsInserted:    0,
		TableResults:         []types.TableInsertResult{},
		Errors:               []types.InsertError{},
	}

	// 1. 解析SQL文件
	statements, err := i.parser.ParseInsertFile(filePath)
	if err != nil {
		result.ExecutionTime = time.Since(startTime).String()
		result.Errors = append(result.Errors, types.InsertError{
			ErrorMessage: fmt.Sprintf("解析SQL文件失败: %v", err),
		})
		return result, err
	}

	result.TotalStatements = len(statements)
	if len(statements) == 0 {
		result.ExecutionTime = time.Since(startTime).String()
		return result, fmt.Errorf("SQL文件中没有找到有效的INSERT语句")
	}

	// 2. 验证INSERT语句
	if err := i.parser.ValidateInsertStatements(statements); err != nil {
		result.ExecutionTime = time.Since(startTime).String()
		result.Errors = append(result.Errors, types.InsertError{
			ErrorMessage: fmt.Sprintf("验证INSERT语句失败: %v", err),
		})
		return result, err
	}

	// 3. 验证表存在性
	if config.ValidateTables {
		tables := i.parser.ExtractTableNames(statements)
		if err := i.ValidateTablesExist(ctx, dbName, tables); err != nil {
			result.ExecutionTime = time.Since(startTime).String()
			result.Errors = append(result.Errors, types.InsertError{
				ErrorMessage: fmt.Sprintf("表存在性验证失败: %v", err),
			})
			return result, err
		}
	}

	// 4. 执行插入操作
	if err := i.ExecuteInsertStatements(ctx, dbName, statements, config); err != nil {
		result.ExecutionTime = time.Since(startTime).String()
		result.FailedStatements = result.TotalStatements
		result.Errors = append(result.Errors, types.InsertError{
			ErrorMessage: fmt.Sprintf("执行插入操作失败: %v", err),
		})
		return result, err
	}

	// 5. 统计结果
	result.SuccessfulStatements = result.TotalStatements
	result.ExecutionTime = time.Since(startTime).String()

	// 统计每个表的插入结果
	tableStats := make(map[string]*types.TableInsertResult)
	for _, stmt := range statements {
		if tableStats[stmt.TableName] == nil {
			tableStats[stmt.TableName] = &types.TableInsertResult{
				TableName:          stmt.TableName,
				RowsInserted:       0,
				StatementsExecuted: 0,
			}
		}
		tableStats[stmt.TableName].StatementsExecuted++
		tableStats[stmt.TableName].RowsInserted += int64(len(stmt.Values))
	}

	for _, tableResult := range tableStats {
		result.TableResults = append(result.TableResults, *tableResult)
		result.TotalRowsInserted += tableResult.RowsInserted
	}

	return result, nil
}

// ValidateTablesExist 验证表是否存在
func (i *Inserter) ValidateTablesExist(ctx context.Context, dbName string, tables []string) error {
	dbConn, err := i.connectToDatabase(dbName)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}
	defer dbConn.Close()

	for _, tableName := range tables {
		exists, err := i.checkTableExists(ctx, dbConn, dbName, tableName)
		if err != nil {
			return fmt.Errorf("检查表 %s 存在性失败: %v", tableName, err)
		}
		if !exists {
			return fmt.Errorf("表 %s 不存在", tableName)
		}
	}

	log.Printf("✅ 所有表存在性验证通过: %v", tables)
	return nil
}

// ExecuteInsertStatements 执行INSERT语句
func (i *Inserter) ExecuteInsertStatements(ctx context.Context, dbName string, statements []types.InsertStatement, config types.DataInsertConfig) error {
	dbConn, err := i.connectToDatabase(dbName)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}
	defer dbConn.Close()

	// 开始事务（如果配置要求）
	var tx *sql.Tx
	if config.UseTransaction {
		tx, err = dbConn.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("开始事务失败: %v", err)
		}
		defer tx.Rollback() // 如果没有提交，自动回滚
	}

	log.Printf("开始执行 %d 个INSERT语句...", len(statements))

	// 执行所有INSERT语句
	totalRows := int64(0)
	for idx, stmt := range statements {
		if config.ProgressCallback != nil {
			config.ProgressCallback("执行INSERT", int64(idx), int64(len(statements)), nil)
		}

		log.Printf("[%d/%d] 插入数据到表: %s (%d行)",
			idx+1, len(statements), stmt.TableName, len(stmt.Values))

		var execErr error
		if config.UseTransaction {
			execErr = i.executeInsertStatement(ctx, tx, stmt, config)
		} else {
			execErr = i.executeInsertStatementDB(ctx, dbConn, stmt, config)
		}

		if execErr != nil {
			errMsg := fmt.Sprintf("执行INSERT语句失败 (表: %s, 行: %d): %v",
				stmt.TableName, stmt.LineNumber, execErr)
			log.Printf("❌ %s", errMsg)

			if config.StopOnError {
				return fmt.Errorf(errMsg)
			}
		} else {
			totalRows += int64(len(stmt.Values))
		}
	}

	// 提交事务
	if config.UseTransaction && tx != nil {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("提交事务失败: %v", err)
		}
	}

	log.Printf("✅ 成功插入 %d 行数据", totalRows)

	if config.ProgressCallback != nil {
		config.ProgressCallback("完成", int64(len(statements)), int64(len(statements)), nil)
	}

	return nil
}

// executeInsertStatement 使用事务执行INSERT语句
func (i *Inserter) executeInsertStatement(ctx context.Context, tx *sql.Tx, stmt types.InsertStatement, config types.DataInsertConfig) error {
	// 批量处理
	return i.executeBatchInsert(ctx, tx, stmt, config)
}

// executeInsertStatementDB 使用数据库连接执行INSERT语句
func (i *Inserter) executeInsertStatementDB(ctx context.Context, db *sql.DB, stmt types.InsertStatement, config types.DataInsertConfig) error {
	// 批量处理
	return i.executeBatchInsertDB(ctx, db, stmt, config)
}

// executeBatchInsert 批量执行INSERT（事务版本）
func (i *Inserter) executeBatchInsert(ctx context.Context, tx *sql.Tx, stmt types.InsertStatement, config types.DataInsertConfig) error {
	batchSize := config.BatchSize
	totalBatches := (len(stmt.Values) + batchSize - 1) / batchSize

	for batchIndex := 0; batchIndex < totalBatches; batchIndex++ {
		start := batchIndex * batchSize
		end := start + batchSize
		if end > len(stmt.Values) {
			end = len(stmt.Values)
		}

		batchValues := stmt.Values[start:end]
		batchSQL := i.buildBatchInsertSQL(stmt.TableName, stmt.Columns, batchValues)

		log.Printf("  批次 %d/%d: 插入 %d 行", batchIndex+1, totalBatches, len(batchValues))

		_, err := tx.ExecContext(ctx, batchSQL)
		if err != nil {
			if i.isDuplicateKeyError(err) && config.OnConflict == "ignore" {
				log.Printf("⚠️ 批次 %d 忽略重复键错误", batchIndex+1)
				continue
			}
			return fmt.Errorf("批次 %d 执行失败: %v", batchIndex+1, err)
		}
	}

	return nil
}

// executeBatchInsertDB 批量执行INSERT（数据库连接版本）
func (i *Inserter) executeBatchInsertDB(ctx context.Context, db *sql.DB, stmt types.InsertStatement, config types.DataInsertConfig) error {
	batchSize := config.BatchSize
	totalBatches := (len(stmt.Values) + batchSize - 1) / batchSize

	for batchIndex := 0; batchIndex < totalBatches; batchIndex++ {
		start := batchIndex * batchSize
		end := start + batchSize
		if end > len(stmt.Values) {
			end = len(stmt.Values)
		}

		batchValues := stmt.Values[start:end]
		batchSQL := i.buildBatchInsertSQL(stmt.TableName, stmt.Columns, batchValues)

		log.Printf("  批次 %d/%d: 插入 %d 行", batchIndex+1, totalBatches, len(batchValues))

		_, err := db.ExecContext(ctx, batchSQL)
		if err != nil {
			if i.isDuplicateKeyError(err) && config.OnConflict == "ignore" {
				log.Printf("⚠️ 批次 %d 忽略重复键错误", batchIndex+1)
				continue
			}
			return fmt.Errorf("批次 %d 执行失败: %v", batchIndex+1, err)
		}
	}

	return nil
}

// buildBatchInsertSQL 构建批量INSERT SQL
func (i *Inserter) buildBatchInsertSQL(tableName string, columns []string, values [][]interface{}) string {
	var sql strings.Builder

	sql.WriteString("INSERT INTO `")
	sql.WriteString(tableName)
	sql.WriteString("`")

	// 添加列名（如果指定）
	if len(columns) > 0 {
		sql.WriteString(" (")
		for idx, col := range columns {
			if idx > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString("`")
			sql.WriteString(col)
			sql.WriteString("`")
		}
		sql.WriteString(")")
	}

	sql.WriteString(" VALUES ")

	// 添加值
	for idx, valueRow := range values {
		if idx > 0 {
			sql.WriteString(", ")
		}
		sql.WriteString("(")
		for j, value := range valueRow {
			if j > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString(i.formatValue(value))
		}
		sql.WriteString(")")
	}

	return sql.String()
}

// formatValue 格式化值
func (i *Inserter) formatValue(value interface{}) string {
	if value == nil {
		return "NULL"
	}

	switch v := value.(type) {
	case string:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
	case int64, int, int32:
		return fmt.Sprintf("%v", v)
	case float64, float32:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

// checkTableExists 检查表是否存在
func (i *Inserter) checkTableExists(ctx context.Context, db *sql.DB, dbName, tableName string) (bool, error) {
	query := `SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES 
			  WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?`

	var name string
	err := db.QueryRowContext(ctx, query, dbName, tableName).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// connectToDatabase 连接到指定数据库
func (i *Inserter) connectToDatabase(dbName string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		i.config.Username,
		i.config.Password,
		i.config.Host,
		i.config.Port,
		dbName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %v", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	return db, nil
}

// isDuplicateKeyError 判断是否是重复键错误
func (i *Inserter) isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "duplicate entry") ||
		strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "unique constraint")
}
