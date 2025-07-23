package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"db-migrator/internal/database"
	"db-migrator/internal/datacopy"
)

var (
	// 数据复制相关参数
	copySourceDB   string
	copyTargetDB   string
	copyTargetDBs  []string
	copyTables     []string
	copyStrategy   string
	copyScope      string
	copyConditions []string
	copyMappings   []string
	copyBatchSize  int
	copyTimeout    string
	copyOnError    string
	copyConfigFile string

	// 数据初始化相关参数
	initDataType string
	initDataFile string
	initDataDir  string
	initFromDB   string
	initTables   []string
	initStrategy string
)

// copyDataCmd 数据复制命令
var copyDataCmd = &cobra.Command{
	Use:   "copy-data",
	Short: "在数据库之间复制数据",
	Long: `在数据库之间复制数据，支持多种策略和范围。

支持的复制策略：
• overwrite  - 完全覆盖（清空后插入）
• merge      - 智能合并（插入或更新）  
• insert     - 仅插入新数据
• ignore     - 忽略重复数据

支持的复制范围：
• full       - 整表复制
• condition  - 条件复制
• mapping    - 字段映射
• transform  - 数据转换

示例：
  # 从总部复制商品数据到所有店铺
  db-migrator copy-data --source=headquarters --patterns=shop_* --tables=products,categories
  
  # 复制指定条件的数据
  db-migrator copy-data --source=main_db --target=backup_db --tables=orders --conditions="orders:status='completed'"
  
  # 使用配置文件复制
  db-migrator copy-data --config=copy-config.json`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := validateCopyFlags(); err != nil {
			log.Fatalf("参数错误: %v", err)
		}

		// 解析目标数据库
		var targetDBs []string
		if copyTargetDB != "" {
			targetDBs = []string{copyTargetDB}
		} else if len(copyTargetDBs) > 0 {
			targetDBs = copyTargetDBs
		} else {
			// 使用模式匹配解析数据库
			multiMigrator, err := createMultiMigrator()
			if err != nil {
				log.Fatalf("创建迁移器失败: %v", err)
			}
			defer multiMigrator.Close()

			patterns := databasePatterns
			if len(patterns) == 0 {
				log.Fatalf("必须指定目标数据库或数据库模式")
			}

			targetDBs, err = multiMigrator.GetMatchedDatabases(context.Background(), patterns)
			if err != nil {
				log.Fatalf("解析目标数据库失败: %v", err)
			}
		}

		fmt.Printf("📊 源数据库: %s\n", copySourceDB)
		fmt.Printf("📊 目标数据库 (%d个): %s\n", len(targetDBs), strings.Join(targetDBs, ", "))
		fmt.Printf("📊 复制表: %s\n", strings.Join(copyTables, ", "))

		// 创建复制配置
		copyConfig, err := createCopyConfig()
		if err != nil {
			log.Fatalf("创建复制配置失败: %v", err)
		}

		// 创建数据库管理器
		dbManager := database.NewManager(config)

		// 创建跨数据库复制器
		copier := datacopy.NewCrossDatabaseCopier(dbManager)

		// 设置进度回调
		progressCallback := func(table string, current, total int64, err error) {
			if err != nil {
				fmt.Printf("❌ 表 %s 复制失败: %v\n", table, err)
			} else if total > 0 {
				progress := float64(current) / float64(total) * 100
				fmt.Printf("⏳ 表 %s: %d/%d (%.1f%%)\n", table, current, total, progress)
			}
		}

		// 执行复制
		ctx := context.Background()
		if len(targetDBs) == 1 {
			err = copier.CopyBetweenDatabases(ctx, copySourceDB, targetDBs[0], *copyConfig, progressCallback)
		} else {
			err = copier.CopyToMultipleDatabases(ctx, copySourceDB, targetDBs, *copyConfig, progressCallback)
		}

		if err != nil {
			log.Fatalf("数据复制失败: %v", err)
		}

		fmt.Println("\n🎉 数据复制完成")
	},
}

// initDataCmd 数据初始化命令
var initDataCmd = &cobra.Command{
	Use:   "init-data",
	Short: "初始化数据库数据",
	Long: `为数据库初始化基础数据。

支持的数据源：
• JSON文件    - 从JSON文件读取数据
• YAML文件    - 从YAML文件读取数据
• 源数据库    - 从其他数据库复制数据
• 内置数据    - 使用预定义的数据

示例：
  # 为新租户初始化基础数据
  db-migrator init-data -d tenant_new_001 --from-db=tenant_template
  
  # 从JSON文件初始化数据
  db-migrator init-data --patterns=shop_* --data-file=shop-init-data.json
  
  # 为所有微服务初始化配置数据
  db-migrator init-data --patterns=*_service --data-type=system_configs`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := validateInitFlags(); err != nil {
			log.Fatalf("参数错误: %v", err)
		}

		// 解析目标数据库
		databases, err := resolveDatabases()
		if err != nil {
			log.Fatalf("解析数据库失败: %v", err)
		}

		printDatabaseInfo(databases)

		// 创建多数据库迁移器
		multiMigrator, err := createMultiMigrator()
		if err != nil {
			log.Fatalf("创建迁移器失败: %v", err)
		}
		defer multiMigrator.Close()

		// 为每个数据库执行初始化
		for _, dbName := range databases {
			fmt.Printf("\n🔄 正在初始化数据库: %s\n", dbName)

			if err := initializeDatabaseData(dbName); err != nil {
				log.Fatalf("初始化数据库 %s 失败: %v", dbName, err)
			}

			fmt.Printf("✅ 数据库 %s 初始化完成\n", dbName)
		}

		fmt.Println("\n🎉 所有数据库数据初始化完成")
	},
}

// 辅助函数

func validateCopyFlags() error {
	if copySourceDB == "" {
		return fmt.Errorf("必须指定源数据库 --source")
	}

	if copyTargetDB == "" && len(copyTargetDBs) == 0 && len(databasePatterns) == 0 {
		return fmt.Errorf("必须指定目标数据库 --target 或 --targets 或 --patterns")
	}

	if len(copyTables) == 0 && copyConfigFile == "" {
		return fmt.Errorf("必须指定要复制的表 --tables 或配置文件 --config")
	}

	return nil
}

func validateInitFlags() error {
	// 检查数据库选择
	if targetDatabase == "" && len(targetDatabases) == 0 && len(databasePatterns) == 0 && !allDatabases {
		return fmt.Errorf("必须指定目标数据库")
	}

	// 检查数据源
	if initDataFile == "" && initFromDB == "" && initDataType == "" && initDataDir == "" {
		return fmt.Errorf("必须指定数据源")
	}

	return nil
}

func createCopyConfig() (*datacopy.CopyConfig, error) {
	// 如果指定了配置文件，从文件加载
	if copyConfigFile != "" {
		return loadCopyConfigFromFile(copyConfigFile)
	}

	// 解析超时时间
	timeout := 30 * time.Minute
	if copyTimeout != "" {
		var err error
		timeout, err = time.ParseDuration(copyTimeout)
		if err != nil {
			return nil, fmt.Errorf("解析超时时间失败: %v", err)
		}
	}

	// 解析复制策略
	strategy := datacopy.CopyStrategyMerge
	switch copyStrategy {
	case "overwrite":
		strategy = datacopy.CopyStrategyOverwrite
	case "merge":
		strategy = datacopy.CopyStrategyMerge
	case "insert":
		strategy = datacopy.CopyStrategyInsertNew
	case "ignore":
		strategy = datacopy.CopyStrategyIgnore
	}

	// 解析复制范围
	scope := datacopy.CopyScopeFullTable
	switch copyScope {
	case "full":
		scope = datacopy.CopyScopeFullTable
	case "condition":
		scope = datacopy.CopyScopeConditional
	case "mapping":
		scope = datacopy.CopyScopeMapping
	case "transform":
		scope = datacopy.CopyScopeTransform
	}

	// 解析条件
	conditions := make(map[string]string)
	for _, condition := range copyConditions {
		parts := strings.SplitN(condition, ":", 2)
		if len(parts) == 2 {
			conditions[parts[0]] = parts[1]
		}
	}

	// 解析字段映射
	fieldMappings := make(map[string][]datacopy.FieldMapping)
	for _, mapping := range copyMappings {
		// 格式: table:source_field=target_field,source_field2=target_field2
		parts := strings.SplitN(mapping, ":", 2)
		if len(parts) == 2 {
			tableName := parts[0]
			mappingPairs := strings.Split(parts[1], ",")

			var mappings []datacopy.FieldMapping
			for _, pair := range mappingPairs {
				fieldParts := strings.SplitN(pair, "=", 2)
				if len(fieldParts) == 2 {
					mappings = append(mappings, datacopy.FieldMapping{
						SourceField: strings.TrimSpace(fieldParts[0]),
						TargetField: strings.TrimSpace(fieldParts[1]),
					})
				}
			}
			fieldMappings[tableName] = mappings
		}
	}

	config := &datacopy.CopyConfig{
		Strategy:      strategy,
		Scope:         scope,
		Tables:        copyTables,
		Conditions:    conditions,
		FieldMappings: fieldMappings,
		BatchSize:     copyBatchSize,
		Timeout:       timeout,
		OnError:       copyOnError,
	}

	return config, nil
}

func loadCopyConfigFromFile(filename string) (*datacopy.CopyConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config datacopy.CopyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

func initializeDatabaseData(dbName string) error {
	// 这里实现具体的数据初始化逻辑
	// 可以根据initFromDB, initDataFile, initDataType等参数来决定初始化方式

	if initFromDB != "" {
		return initializeFromDatabase(dbName, initFromDB)
	}

	if initDataFile != "" {
		return initializeFromFile(dbName, initDataFile)
	}

	if initDataType != "" {
		return initializeBuiltinData(dbName, initDataType)
	}

	if initDataDir != "" {
		return initializeFromDirectory(dbName, initDataDir)
	}

	return fmt.Errorf("未指定有效的数据源")
}

func initializeFromDatabase(targetDB, sourceDB string) error {
	// 从源数据库复制数据到目标数据库
	dbManager := database.NewManager(config)
	copier := datacopy.NewCrossDatabaseCopier(dbManager)

	// 创建复制配置
	copyConfig := datacopy.CopyConfig{
		Strategy:  datacopy.CopyStrategyMerge,
		Scope:     datacopy.CopyScopeFullTable,
		Tables:    initTables,
		BatchSize: 1000,
		Timeout:   30 * time.Minute,
		OnError:   "stop",
	}

	// 如果没有指定表，复制所有表
	if len(initTables) == 0 {
		// 这里可以查询源数据库的所有表
		// 为简化，使用常见的基础数据表
		copyConfig.Tables = []string{"system_configs", "user_roles", "permissions", "categories"}
	}

	ctx := context.Background()
	return copier.CopyBetweenDatabases(ctx, sourceDB, targetDB, copyConfig, nil)
}

func initializeFromFile(dbName, filename string) error {
	// 从文件初始化数据
	fmt.Printf("从文件 %s 初始化数据库 %s\n", filename, dbName)
	// TODO: 实现文件数据初始化逻辑
	return nil
}

func initializeBuiltinData(dbName, dataType string) error {
	// 初始化内置数据
	fmt.Printf("为数据库 %s 初始化内置数据类型: %s\n", dbName, dataType)
	// TODO: 实现内置数据初始化逻辑
	return nil
}

func initializeFromDirectory(dbName, dirPath string) error {
	// 从目录初始化数据
	fmt.Printf("从目录 %s 初始化数据库 %s\n", dirPath, dbName)
	// TODO: 实现目录数据初始化逻辑
	return nil
}

func init() {
	// 数据复制命令参数
	copyDataCmd.Flags().StringVar(&copySourceDB, "source", "", "源数据库名称")
	copyDataCmd.Flags().StringVar(&copyTargetDB, "target", "", "目标数据库名称")
	copyDataCmd.Flags().StringSliceVar(&copyTargetDBs, "targets", []string{}, "多个目标数据库")
	copyDataCmd.Flags().StringSliceVar(&copyTables, "tables", []string{}, "要复制的表名")
	copyDataCmd.Flags().StringVar(&copyStrategy, "strategy", "merge", "复制策略: overwrite, merge, insert, ignore")
	copyDataCmd.Flags().StringVar(&copyScope, "scope", "full", "复制范围: full, condition, mapping, transform")
	copyDataCmd.Flags().StringSliceVar(&copyConditions, "conditions", []string{}, "复制条件 table:condition")
	copyDataCmd.Flags().StringSliceVar(&copyMappings, "mappings", []string{}, "字段映射 table:src=dst,src2=dst2")
	copyDataCmd.Flags().IntVar(&copyBatchSize, "batch-size", 1000, "批量大小")
	copyDataCmd.Flags().StringVar(&copyTimeout, "timeout", "30m", "超时时间")
	copyDataCmd.Flags().StringVar(&copyOnError, "on-error", "stop", "错误处理: stop, continue, rollback")
	copyDataCmd.Flags().StringVar(&copyConfigFile, "config", "", "复制配置文件")

	// 添加数据库选择参数
	addDatabaseFlags(copyDataCmd)

	// 数据初始化命令参数
	initDataCmd.Flags().StringVar(&initDataType, "data-type", "", "内置数据类型")
	initDataCmd.Flags().StringVar(&initDataFile, "data-file", "", "数据文件路径")
	initDataCmd.Flags().StringVar(&initDataDir, "data-dir", "", "数据目录路径")
	initDataCmd.Flags().StringVar(&initFromDB, "from-db", "", "源数据库名称")
	initDataCmd.Flags().StringSliceVar(&initTables, "tables", []string{}, "要初始化的表")
	initDataCmd.Flags().StringVar(&initStrategy, "strategy", "merge", "初始化策略")

	// 添加数据库选择参数
	addDatabaseFlags(initDataCmd)

	// 添加到根命令
	rootCmd.AddCommand(copyDataCmd)
	rootCmd.AddCommand(initDataCmd)
}
