package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/xiezhihuan/db-migrator/internal/database"
	"github.com/xiezhihuan/db-migrator/internal/types"

	"github.com/spf13/cobra"
)

var createDBCmd = &cobra.Command{
	Use:   "create-db",
	Short: "从SQL文件创建新数据库",
	Long: `从SQL文件创建新数据库，支持：
- 指定数据库名称和字符集
- 检查数据库是否已存在
- 解析SQL文件中的DDL语句
- 按依赖关系执行创建语句
- 支持表、视图、存储过程、触发器、索引等`,
	Example: `  # 从SQL文件创建数据库
  db-migrator create-db --name "my_new_shop" --from-sql "schema.sql"
  
  # 指定字符集和排序规则
  db-migrator create-db --name "my_shop" --from-sql "schema.sql" --charset utf8mb4 --collation utf8mb4_unicode_ci
  
  # 如果数据库已存在则跳过
  db-migrator create-db --name "my_shop" --from-sql "schema.sql" --if-exists skip`,
	RunE: runCreateDB,
}

var (
	createDBName      string
	createDBSQLFile   string
	createDBCharset   string
	createDBCollation string
	createDBIfExists  string
)

func init() {
	rootCmd.AddCommand(createDBCmd)

	createDBCmd.Flags().StringVar(&createDBName, "name", "", "数据库名称 (必填)")
	createDBCmd.Flags().StringVar(&createDBSQLFile, "from-sql", "", "SQL文件路径 (必填)")
	createDBCmd.Flags().StringVar(&createDBCharset, "charset", "utf8mb4", "数据库字符集")
	createDBCmd.Flags().StringVar(&createDBCollation, "collation", "utf8mb4_unicode_ci", "数据库排序规则")
	createDBCmd.Flags().StringVar(&createDBIfExists, "if-exists", "error", "数据库已存在时的处理方式: error, skip, prompt")

	createDBCmd.MarkFlagRequired("name")
	createDBCmd.MarkFlagRequired("from-sql")
}

func runCreateDB(cmd *cobra.Command, args []string) error {
	// 验证参数
	if createDBName == "" {
		return fmt.Errorf("必须指定数据库名称 (--name)")
	}

	if createDBSQLFile == "" {
		return fmt.Errorf("必须指定SQL文件路径 (--from-sql)")
	}

	// 验证SQL文件是否存在
	if _, err := os.Stat(createDBSQLFile); os.IsNotExist(err) {
		return fmt.Errorf("SQL文件不存在: %s", createDBSQLFile)
	}

	// 验证if-exists参数
	validIfExists := []string{"error", "skip", "prompt"}
	if !contains(validIfExists, createDBIfExists) {
		return fmt.Errorf("无效的if-exists值: %s，支持的值: %s", createDBIfExists, strings.Join(validIfExists, ", "))
	}

	// 获取绝对路径
	absPath, err := filepath.Abs(createDBSQLFile)
	if err != nil {
		return fmt.Errorf("获取SQL文件绝对路径失败: %v", err)
	}

	// 验证数据库名称
	if err := validateDatabaseName(createDBName); err != nil {
		return fmt.Errorf("数据库名称无效: %v", err)
	}

	// 打印执行信息
	log.Printf("开始创建数据库...")
	log.Printf("  数据库名称: %s", createDBName)
	log.Printf("  SQL文件: %s", absPath)
	log.Printf("  字符集: %s", createDBCharset)
	log.Printf("  排序规则: %s", createDBCollation)
	log.Printf("  已存在处理: %s", createDBIfExists)

	// 使用全局配置变量（已在根命令中初始化）

	// 创建根连接（不指定数据库）
	rootConn, err := createRootConnection(&config.Database)
	if err != nil {
		return fmt.Errorf("连接数据库服务器失败: %v", err)
	}
	defer rootConn.Close()

	// 创建数据库创建器
	creator := database.NewCreator(rootConn, &config.Database)

	// 准备数据库创建配置
	dbConfig := types.DatabaseCreateConfig{
		Name:      createDBName,
		Charset:   createDBCharset,
		Collation: createDBCollation,
		IfExists:  createDBIfExists,
	}

	// 执行创建
	ctx := context.Background()
	result, err := creator.CreateFromSQLFile(ctx, dbConfig, absPath)
	if err != nil {
		if len(result.Errors) > 0 {
			log.Printf("执行过程中发生错误:")
			for _, errMsg := range result.Errors {
				log.Printf("  ❌ %s", errMsg)
			}
		}
		return fmt.Errorf("创建数据库失败: %v", err)
	}

	// 打印结果
	printCreateDBResult(result)

	return nil
}

// createRootConnection 创建根连接（不指定数据库）
func createRootConnection(config *types.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
	)

	return sql.Open("mysql", dsn)
}

// validateDatabaseName 验证数据库名称
func validateDatabaseName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("数据库名称不能为空")
	}

	if len(name) > 64 {
		return fmt.Errorf("数据库名称长度不能超过64个字符")
	}

	// MySQL数据库名称规则：只能包含字母、数字、下划线
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return fmt.Errorf("数据库名称只能包含字母、数字和下划线")
		}
	}

	return nil
}

// printCreateDBResult 打印创建结果
func printCreateDBResult(result *types.CreateFromSQLResult) {
	log.Printf("\n🎉 数据库创建完成!")
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("📊 执行摘要:")
	log.Printf("  数据库名称: %s", result.DatabaseName)
	log.Printf("  数据库已创建: %v", result.DatabaseCreated)
	log.Printf("  SQL语句总数: %d", result.StatementsTotal)
	log.Printf("  成功执行: %d", result.StatementsSuccess)
	log.Printf("  执行失败: %d", result.StatementsFailed)
	log.Printf("  执行时间: %s", result.ExecutionTime)

	if len(result.CreatedObjects) > 0 {
		log.Printf("\n📋 创建的数据库对象:")

		// 按类型分组统计
		typeCount := make(map[string][]string)
		for _, obj := range result.CreatedObjects {
			typeCount[obj.Type] = append(typeCount[obj.Type], obj.Name)
		}

		// 按类型顺序打印
		typeOrder := []string{"CREATE_TABLE", "CREATE_VIEW", "CREATE_PROCEDURE", "CREATE_FUNCTION", "CREATE_TRIGGER", "CREATE_INDEX", "CREATE_OTHER"}
		for _, objType := range typeOrder {
			if objects, exists := typeCount[objType]; exists {
				displayName := getTypeDisplayName(objType)
				log.Printf("  %s (%d个):", displayName, len(objects))
				for _, name := range objects {
					log.Printf("    ✓ %s", name)
				}
			}
		}
	}

	if len(result.Errors) > 0 {
		log.Printf("\n⚠️  警告/错误:")
		for _, errMsg := range result.Errors {
			log.Printf("  • %s", errMsg)
		}
	}

	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// getTypeDisplayName 获取类型的中文显示名称
func getTypeDisplayName(objType string) string {
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

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
