package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"db-migrator/internal/database"
	"db-migrator/internal/types"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)

var insertDataCmd = &cobra.Command{
	Use:   "insert-data",
	Short: "从SQL文件向数据库插入数据",
	Long: `从SQL文件向数据库插入数据，支持：
- 解析INSERT语句并验证表存在性
- 支持批量插入和事务管理
- 主键冲突时报错停止并回滚
- 支持多数据库和模式匹配
- 详细的进度显示和错误报告`,
	Example: `  # 向单个数据库插入数据
  db-migrator insert-data --database "my_shop" --from-sql "data.sql"
  
  # 向所有匹配的数据库插入相同数据
  db-migrator insert-data --patterns "shop_*" --from-sql "base_data.sql"
  
  # 指定批量大小和验证表存在性
  db-migrator insert-data --database "my_shop" --from-sql "large_data.sql" --batch-size 500 --validate-tables
  
  # 向多个指定数据库插入数据
  db-migrator insert-data --databases "shop_001,shop_002,shop_003" --from-sql "promotion_data.sql"`,
	RunE: runInsertData,
}

var (
	insertDataSQLFile        string
	insertDataBatchSize      int
	insertDataOnConflict     string
	insertDataValidateTables bool
	insertDataUseTransaction bool
	insertDataStopOnError    bool
)

func init() {
	rootCmd.AddCommand(insertDataCmd)

	insertDataCmd.Flags().StringVar(&insertDataSQLFile, "from-sql", "", "包含INSERT语句的SQL文件路径 (必填)")
	insertDataCmd.Flags().IntVar(&insertDataBatchSize, "batch-size", 1000, "批量插入大小")
	insertDataCmd.Flags().StringVar(&insertDataOnConflict, "on-conflict", "error", "主键冲突处理: error, ignore")
	insertDataCmd.Flags().BoolVar(&insertDataValidateTables, "validate-tables", true, "验证表是否存在")
	insertDataCmd.Flags().BoolVar(&insertDataUseTransaction, "use-transaction", true, "使用事务保证一致性")
	insertDataCmd.Flags().BoolVar(&insertDataStopOnError, "stop-on-error", true, "遇到错误时停止执行")

	insertDataCmd.MarkFlagRequired("from-sql")

	// 添加数据库选择参数
	addDatabaseFlags(insertDataCmd)
}

func runInsertData(cmd *cobra.Command, args []string) error {
	// 验证参数
	if insertDataSQLFile == "" {
		return fmt.Errorf("必须指定SQL文件路径 (--from-sql)")
	}

	// 验证SQL文件是否存在
	if _, err := os.Stat(insertDataSQLFile); os.IsNotExist(err) {
		return fmt.Errorf("SQL文件不存在: %s", insertDataSQLFile)
	}

	// 验证on-conflict参数
	validConflictStrategies := []string{"error", "ignore"}
	if !contains(validConflictStrategies, insertDataOnConflict) {
		return fmt.Errorf("无效的on-conflict值: %s，支持的值: %s",
			insertDataOnConflict, strings.Join(validConflictStrategies, ", "))
	}

	// 验证数据库参数
	if err := validateDatabaseFlags(); err != nil {
		return fmt.Errorf("参数错误: %v", err)
	}

	// 获取绝对路径
	absPath, err := filepath.Abs(insertDataSQLFile)
	if err != nil {
		return fmt.Errorf("获取SQL文件绝对路径失败: %v", err)
	}

	// 解析目标数据库
	databases, err := resolveDatabases()
	if err != nil {
		return fmt.Errorf("解析数据库失败: %v", err)
	}

	// 如果没有指定数据库，使用默认数据库
	if len(databases) == 0 {
		if config.Database.Database == "" {
			return fmt.Errorf("必须指定目标数据库或在配置中设置默认数据库")
		}
		databases = []string{config.Database.Database}
	}

	// 打印执行信息
	log.Printf("开始执行数据插入...")
	log.Printf("  SQL文件: %s", absPath)
	log.Printf("  目标数据库 (%d个): %s", len(databases), strings.Join(databases, ", "))
	log.Printf("  批量大小: %d", insertDataBatchSize)
	log.Printf("  冲突策略: %s", insertDataOnConflict)
	log.Printf("  验证表: %v", insertDataValidateTables)
	log.Printf("  使用事务: %v", insertDataUseTransaction)
	log.Printf("  遇错停止: %v", insertDataStopOnError)

	// 创建根连接（不指定数据库）
	rootConn, err := createRootConnection(&config.Database)
	if err != nil {
		return fmt.Errorf("连接数据库服务器失败: %v", err)
	}
	defer rootConn.Close()

	// 创建数据插入器
	inserter := database.NewInserter(rootConn, &config.Database)

	// 准备插入配置
	insertConfig := types.DataInsertConfig{
		BatchSize:        insertDataBatchSize,
		OnConflict:       insertDataOnConflict,
		StopOnError:      insertDataStopOnError,
		ValidateTables:   insertDataValidateTables,
		UseTransaction:   insertDataUseTransaction,
		ProgressCallback: createProgressCallback(),
	}

	// 执行插入
	ctx := context.Background()
	if len(databases) == 1 {
		// 单数据库插入
		result, err := inserter.InsertFromSQLFile(ctx, databases[0], absPath, insertConfig)
		if err != nil {
			if result != nil && len(result.Errors) > 0 {
				log.Printf("执行过程中发生错误:")
				for _, errInfo := range result.Errors {
					log.Printf("  ❌ %s", errInfo.ErrorMessage)
				}
			}
			return fmt.Errorf("数据插入失败: %v", err)
		}
		printInsertResult(result)
	} else {
		// 多数据库插入
		multiResult, err := executeMultiDatabaseInsert(ctx, inserter, databases, absPath, insertConfig)
		if err != nil {
			return fmt.Errorf("多数据库插入失败: %v", err)
		}
		printMultiInsertResult(multiResult)
	}

	return nil
}

// executeMultiDatabaseInsert 执行多数据库插入
func executeMultiDatabaseInsert(ctx context.Context, inserter *database.Inserter, databases []string, sqlFile string, config types.DataInsertConfig) (*types.MultiDatabaseInsertResult, error) {
	startTime := time.Now()

	result := &types.MultiDatabaseInsertResult{
		TotalDatabases:      len(databases),
		SuccessfulDatabases: 0,
		FailedDatabases:     0,
		DatabaseResults:     []types.DataInsertResult{},
		Errors:              []string{},
	}

	for i, dbName := range databases {
		log.Printf("\n[%d/%d] 正在处理数据库: %s", i+1, len(databases), dbName)

		dbResult, err := inserter.InsertFromSQLFile(ctx, dbName, sqlFile, config)
		if err != nil {
			result.FailedDatabases++
			errMsg := fmt.Sprintf("数据库 %s 插入失败: %v", dbName, err)
			result.Errors = append(result.Errors, errMsg)
			log.Printf("❌ %s", errMsg)

			if config.StopOnError {
				result.ExecutionTime = time.Since(startTime).String()
				return result, fmt.Errorf("在数据库 %s 处停止执行", dbName)
			}
		} else {
			result.SuccessfulDatabases++
			log.Printf("✅ 数据库 %s 插入成功", dbName)
		}

		if dbResult != nil {
			result.DatabaseResults = append(result.DatabaseResults, *dbResult)
		}
	}

	result.ExecutionTime = time.Since(startTime).String()
	return result, nil
}

// createProgressCallback 创建进度回调函数
func createProgressCallback() types.ProgressCallback {
	return func(stage string, current, total int64, err error) {
		if err != nil {
			log.Printf("❌ %s 错误: %v", stage, err)
		} else if total > 0 {
			progress := float64(current) / float64(total) * 100
			log.Printf("⏳ %s: %d/%d (%.1f%%)", stage, current, total, progress)
		}
	}
}

// printInsertResult 打印单数据库插入结果
func printInsertResult(result *types.DataInsertResult) {
	log.Printf("\n🎉 数据插入完成!")
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("📊 执行摘要:")
	log.Printf("  数据库名称: %s", result.DatabaseName)
	log.Printf("  SQL语句总数: %d", result.TotalStatements)
	log.Printf("  成功执行: %d", result.SuccessfulStatements)
	log.Printf("  执行失败: %d", result.FailedStatements)
	log.Printf("  插入总行数: %d", result.TotalRowsInserted)
	log.Printf("  执行时间: %s", result.ExecutionTime)

	if len(result.TableResults) > 0 {
		log.Printf("\n📋 各表插入统计:")
		for _, tableResult := range result.TableResults {
			log.Printf("  %s: %d行 (%d个语句)",
				tableResult.TableName,
				tableResult.RowsInserted,
				tableResult.StatementsExecuted)
		}
	}

	if len(result.Errors) > 0 {
		log.Printf("\n⚠️  错误详情:")
		for _, errInfo := range result.Errors {
			if errInfo.TableName != "" {
				log.Printf("  • 表 %s (第%d行): %s",
					errInfo.TableName, errInfo.LineNumber, errInfo.ErrorMessage)
			} else {
				log.Printf("  • %s", errInfo.ErrorMessage)
			}
		}
	}

	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// printMultiInsertResult 打印多数据库插入结果
func printMultiInsertResult(result *types.MultiDatabaseInsertResult) {
	log.Printf("\n🎉 多数据库插入完成!")
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("📊 总体摘要:")
	log.Printf("  数据库总数: %d", result.TotalDatabases)
	log.Printf("  成功数据库: %d", result.SuccessfulDatabases)
	log.Printf("  失败数据库: %d", result.FailedDatabases)
	log.Printf("  执行时间: %s", result.ExecutionTime)

	if len(result.DatabaseResults) > 0 {
		log.Printf("\n📋 各数据库详情:")
		totalRows := int64(0)
		for _, dbResult := range result.DatabaseResults {
			status := "✅"
			if dbResult.FailedStatements > 0 {
				status = "❌"
			}
			log.Printf("  %s %s: %d行 (%d语句)",
				status, dbResult.DatabaseName,
				dbResult.TotalRowsInserted, dbResult.SuccessfulStatements)
			totalRows += dbResult.TotalRowsInserted
		}
		log.Printf("\n  📈 总计插入行数: %d", totalRows)
	}

	if len(result.Errors) > 0 {
		log.Printf("\n⚠️  错误汇总:")
		for _, errMsg := range result.Errors {
			log.Printf("  • %s", errMsg)
		}
	}

	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}
