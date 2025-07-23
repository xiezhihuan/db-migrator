package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"db-migrator/internal/sqlparser"
	"db-migrator/internal/types"
)

// Creator 数据库创建器实现
type Creator struct {
	rootConn *sql.DB
	config   *types.DatabaseConfig
	parser   types.SQLParser
}

// NewCreator 创建新的数据库创建器
func NewCreator(rootConn *sql.DB, config *types.DatabaseConfig) *Creator {
	return &Creator{
		rootConn: rootConn,
		config:   config,
		parser:   sqlparser.NewParser(),
	}
}

// CreateDatabase 创建数据库
func (c *Creator) CreateDatabase(ctx context.Context, config types.DatabaseCreateConfig) error {
	// 检查数据库是否已存在
	exists, err := c.DatabaseExists(ctx, config.Name)
	if err != nil {
		return fmt.Errorf("检查数据库存在性失败: %v", err)
	}

	if exists {
		switch config.IfExists {
		case "error":
			return fmt.Errorf("数据库 '%s' 已存在", config.Name)
		case "skip":
			log.Printf("数据库 '%s' 已存在，跳过创建", config.Name)
			return nil
		case "prompt":
			return fmt.Errorf("数据库 '%s' 已存在，请手动确认是否继续", config.Name)
		default:
			return fmt.Errorf("数据库 '%s' 已存在，请指定处理方式", config.Name)
		}
	}

	// 设置默认字符集
	charset := config.Charset
	if charset == "" {
		charset = "utf8mb4"
	}

	collation := config.Collation
	if collation == "" {
		collation = "utf8mb4_unicode_ci"
	}

	// 构建创建数据库的SQL
	createSQL := fmt.Sprintf(
		"CREATE DATABASE `%s` CHARACTER SET %s COLLATE %s",
		config.Name, charset, collation,
	)

	// 执行创建数据库
	_, err = c.rootConn.ExecContext(ctx, createSQL)
	if err != nil {
		return fmt.Errorf("创建数据库失败: %v", err)
	}

	log.Printf("成功创建数据库: %s (字符集: %s, 排序规则: %s)", config.Name, charset, collation)
	return nil
}

// DatabaseExists 检查数据库是否存在
func (c *Creator) DatabaseExists(ctx context.Context, name string) (bool, error) {
	query := "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?"

	var schemaName string
	err := c.rootConn.QueryRowContext(ctx, query, name).Scan(&schemaName)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("查询数据库存在性失败: %v", err)
	}

	return true, nil
}

// ExecuteSQLFile 执行SQL文件
func (c *Creator) ExecuteSQLFile(ctx context.Context, dbName, filePath string) error {
	// 解析SQL文件
	statements, err := c.parser.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("解析SQL文件失败: %v", err)
	}

	if len(statements) == 0 {
		log.Printf("警告: SQL文件 '%s' 中没有找到有效的DDL语句", filePath)
		return nil
	}

	// 验证语句
	if err := c.parser.ValidateStatements(statements); err != nil {
		return fmt.Errorf("验证SQL语句失败: %v", err)
	}

	// 按依赖关系排序
	sortedStatements, err := c.parser.SortByDependencies(statements)
	if err != nil {
		return fmt.Errorf("排序SQL语句失败: %v", err)
	}

	// 连接到目标数据库
	dbConn, err := c.connectToDatabase(dbName)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}
	defer dbConn.Close()

	// 开始事务
	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 执行所有语句
	var createdObjects []types.ObjectInfo
	var errors []string

	log.Printf("开始执行 %d 个DDL语句...", len(sortedStatements))

	for i, stmt := range sortedStatements {
		log.Printf("[%d/%d] 执行 %s: %s", i+1, len(sortedStatements), stmt.Type, stmt.Name)

		_, err := tx.ExecContext(ctx, stmt.Statement)
		if err != nil {
			errMsg := fmt.Sprintf("执行语句失败 (%s: %s): %v", stmt.Type, stmt.Name, err)
			errors = append(errors, errMsg)
			log.Printf("错误: %s", errMsg)
			return fmt.Errorf("执行SQL语句失败: %v", err)
		}

		createdObjects = append(createdObjects, types.ObjectInfo{
			Type: stmt.Type,
			Name: stmt.Name,
		})
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	log.Printf("成功执行所有DDL语句，共创建 %d 个数据库对象", len(createdObjects))

	// 打印创建的对象摘要
	typeCount := make(map[string]int)
	for _, obj := range createdObjects {
		typeCount[obj.Type]++
	}

	log.Printf("创建对象摘要:")
	for objType, count := range typeCount {
		log.Printf("  %s: %d 个", c.getTypeDisplayName(objType), count)
	}

	return nil
}

// connectToDatabase 连接到指定数据库
func (c *Creator) connectToDatabase(dbName string) (*sql.DB, error) {
	// 构建连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.config.Username,
		c.config.Password,
		c.config.Host,
		c.config.Port,
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

// getTypeDisplayName 获取类型的显示名称
func (c *Creator) getTypeDisplayName(objType string) string {
	switch objType {
	case "CREATE_TABLE":
		return "表"
	case "CREATE_VIEW":
		return "视图"
	case "CREATE_PROCEDURE":
		return "存储过程"
	case "CREATE_FUNCTION":
		return "函数"
	case "CREATE_TRIGGER":
		return "触发器"
	case "CREATE_INDEX":
		return "索引"
	default:
		return "其他对象"
	}
}

// CreateFromSQLFile 从SQL文件创建数据库和所有对象（便捷方法）
func (c *Creator) CreateFromSQLFile(ctx context.Context, dbConfig types.DatabaseCreateConfig, sqlFilePath string) (*types.CreateFromSQLResult, error) {
	startTime := time.Now()

	result := &types.CreateFromSQLResult{
		DatabaseName:      dbConfig.Name,
		DatabaseCreated:   false,
		StatementsTotal:   0,
		StatementsSuccess: 0,
		StatementsFailed:  0,
		Errors:            []string{},
		CreatedObjects:    []types.ObjectInfo{},
	}

	// 1. 创建数据库
	err := c.CreateDatabase(ctx, dbConfig)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("创建数据库失败: %v", err))
		result.ExecutionTime = time.Since(startTime).String()
		return result, err
	}
	result.DatabaseCreated = true

	// 2. 解析SQL文件
	statements, err := c.parser.ParseFile(sqlFilePath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("解析SQL文件失败: %v", err))
		result.ExecutionTime = time.Since(startTime).String()
		return result, err
	}
	result.StatementsTotal = len(statements)

	if len(statements) == 0 {
		result.ExecutionTime = time.Since(startTime).String()
		return result, nil
	}

	// 3. 执行SQL语句
	err = c.ExecuteSQLFile(ctx, dbConfig.Name, sqlFilePath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("执行SQL文件失败: %v", err))
		result.StatementsFailed = result.StatementsTotal
		result.ExecutionTime = time.Since(startTime).String()
		return result, err
	}

	result.StatementsSuccess = result.StatementsTotal
	result.ExecutionTime = time.Since(startTime).String()

	// 4. 收集创建的对象信息
	for _, stmt := range statements {
		result.CreatedObjects = append(result.CreatedObjects, types.ObjectInfo{
			Type: stmt.Type,
			Name: stmt.Name,
		})
	}

	return result, nil
}
