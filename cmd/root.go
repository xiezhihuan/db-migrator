package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"db-migrator/internal/checker"
	"db-migrator/internal/database"
	"db-migrator/internal/migrator"
	"db-migrator/internal/types"
)

var (
	cfgFile string
	config  types.Config
	verbose bool

	// 多数据库支持的参数
	targetDatabase   string   // 目标数据库
	targetDatabases  []string // 多个目标数据库
	databasePatterns []string // 数据库匹配模式
	allDatabases     bool     // 是否操作所有数据库
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "db-migrator",
	Short: "智能数据库迁移工具",
	Long: `db-migrator 是一个智能的数据库迁移工具，支持：

• 智能存在性检查（表、列、索引、函数等）
• 自动跳过已存在的对象
• 事务安全的迁移执行
• 迁移版本控制和历史记录
• 支持回滚操作
• MySQL/MariaDB 支持`,
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// 全局参数
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认: ./config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")

	// 添加子命令
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(versionCmd)
	// rootCmd.AddCommand(discoverCmd) // 新增的数据库发现命令 - 稍后实现

	// 为支持多数据库的命令添加数据库参数
	addDatabaseFlags(upCmd)
	addDatabaseFlags(downCmd)
	addDatabaseFlags(statusCmd)

	// 为down命令添加特定参数
	downCmd.Flags().IntP("steps", "s", 1, "回滚步数")

	// 为create命令添加数据库参数
	createCmd.Flags().StringVarP(&targetDatabase, "database", "d", "", "为指定数据库创建迁移文件")
}

// initConfig 初始化配置
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// 环境变量前缀
	viper.SetEnvPrefix("DB_MIGRATOR")
	viper.AutomaticEnv()

	// 设置默认值
	viper.SetDefault("database.driver", "mysql")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.charset", "utf8mb4")
	viper.SetDefault("migrator.migrations_table", "schema_migrations")
	viper.SetDefault("migrator.lock_table", "schema_migrations_lock")
	viper.SetDefault("migrator.auto_backup", false)
	viper.SetDefault("migrator.dry_run", false)
	viper.SetDefault("migrator.migrations_dir", "migrations")
	viper.SetDefault("migrator.default_database", "")

	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Printf("使用配置文件: %s\n", viper.ConfigFileUsed())
		}
	}

	// 解析配置到结构体
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}
}

// createMigrator 创建迁移器实例
func createMigrator() (*migrator.Migrator, error) {
	// 创建数据库连接
	db, err := database.NewMySQLDB(config.Database)
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接失败: %v", err)
	}

	// 创建检查器
	checker := checker.NewMySQLChecker(db, config.Database.Database)

	// 创建迁移器
	m := migrator.NewMigrator(db, checker, config.Migrator)

	return m, nil
}

// initCmd 初始化命令
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化迁移器",
	Long:  "创建必要的系统表和配置文件",
	Run: func(cmd *cobra.Command, args []string) {
		// 创建默认配置文件
		if err := createDefaultConfig(); err != nil {
			log.Fatalf("创建配置文件失败: %v", err)
		}

		// 创建迁移器
		m, err := createMigrator()
		if err != nil {
			log.Fatalf("创建迁移器失败: %v", err)
		}

		// 初始化
		ctx := context.Background()
		if err := m.Init(ctx); err != nil {
			log.Fatalf("初始化迁移器失败: %v", err)
		}

		// 创建迁移目录
		if err := createMigrationsDir(); err != nil {
			log.Fatalf("创建迁移目录失败: %v", err)
		}

		fmt.Println("✅ 初始化完成！")
		fmt.Println("📁 配置文件: config.yaml")
		fmt.Println("📁 迁移目录: migrations/")
		fmt.Println("🚀 可以开始使用 'db-migrator create' 创建迁移了")
	},
}

// upCmd 执行迁移命令
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "执行所有待处理的迁移",
	Long: `执行所有待处理的迁移到指定的数据库。

支持多种数据库选择方式：
• --database=name        指定单个数据库
• --databases=db1,db2    指定多个数据库
• --patterns=shop*       使用通配符匹配数据库
• --all                  操作所有配置的数据库

示例：
  db-migrator up                    # 默认数据库
  db-migrator up -d main            # 指定数据库
  db-migrator up --databases=main,logs  # 多个数据库
  db-migrator up --patterns=shop*   # 匹配模式
  db-migrator up --all              # 所有数据库`,
	Run: func(cmd *cobra.Command, args []string) {
		// 验证数据库参数
		if err := validateDatabaseFlags(); err != nil {
			log.Fatalf("参数错误: %v", err)
		}

		// 解析目标数据库
		databases, err := resolveDatabases()
		if err != nil {
			log.Fatalf("解析数据库失败: %v", err)
		}

		// 打印操作信息
		printDatabaseInfo(databases)

		// 创建多数据库迁移器
		multiMigrator, err := createMultiMigrator()
		if err != nil {
			log.Fatalf("创建迁移器失败: %v", err)
		}
		defer multiMigrator.Close()

		// 加载迁移（这里需要实现具体的加载逻辑）
		// TODO: 实现从目录加载迁移文件的逻辑

		// 执行迁移
		ctx := context.Background()
		if err := multiMigrator.Up(ctx, databases); err != nil {
			log.Fatalf("执行迁移失败: %v", err)
		}

		fmt.Println("\n🎉 所有数据库迁移执行完成")
	},
}

// downCmd 回滚迁移命令
var downCmd = &cobra.Command{
	Use:   "down [steps]",
	Short: "回滚指定数量的迁移",
	Long: `回滚指定数量的迁移。

支持多种数据库选择方式：
• --database=name        指定单个数据库
• --databases=db1,db2    指定多个数据库  
• --patterns=shop*       使用通配符匹配数据库
• --all                  操作所有配置的数据库

示例：
  db-migrator down                    # 回滚默认数据库1步
  db-migrator down --steps=3          # 回滚默认数据库3步
  db-migrator down -d main --steps=2  # 回滚指定数据库2步
  db-migrator down --all --steps=1    # 回滚所有数据库1步`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 验证数据库参数
		if err := validateDatabaseFlags(); err != nil {
			log.Fatalf("参数错误: %v", err)
		}

		// 获取回滚步数
		steps, _ := cmd.Flags().GetInt("steps")
		if steps <= 0 {
			steps = 1
		}

		// 解析目标数据库
		databases, err := resolveDatabases()
		if err != nil {
			log.Fatalf("解析数据库失败: %v", err)
		}

		// 打印操作信息
		printDatabaseInfo(databases)
		fmt.Printf("📊 回滚步数: %d\n", steps)

		// 创建多数据库迁移器
		multiMigrator, err := createMultiMigrator()
		if err != nil {
			log.Fatalf("创建迁移器失败: %v", err)
		}
		defer multiMigrator.Close()

		// 执行回滚
		ctx := context.Background()
		if err := multiMigrator.Down(ctx, databases, steps); err != nil {
			log.Fatalf("回滚迁移失败: %v", err)
		}

		fmt.Printf("\n🎉 成功回滚 %d 个迁移\n", steps)
	},
}

// statusCmd 查看状态命令
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看迁移状态",
	Long: `查看数据库迁移状态。

支持多种数据库选择方式：
• --database=name        指定单个数据库
• --databases=db1,db2    指定多个数据库
• --patterns=shop*       使用通配符匹配数据库
• --all                  查看所有配置的数据库

示例：
  db-migrator status                   # 默认数据库状态
  db-migrator status -d main           # 指定数据库状态
  db-migrator status --all             # 所有数据库状态
  db-migrator status --patterns=shop*  # 匹配模式的数据库`,
	Run: func(cmd *cobra.Command, args []string) {
		// 验证数据库参数
		if err := validateDatabaseFlags(); err != nil {
			log.Fatalf("参数错误: %v", err)
		}

		// 解析目标数据库
		databases, err := resolveDatabases()
		if err != nil {
			log.Fatalf("解析数据库失败: %v", err)
		}

		// 创建多数据库迁移器
		multiMigrator, err := createMultiMigrator()
		if err != nil {
			log.Fatalf("创建迁移器失败: %v", err)
		}
		defer multiMigrator.Close()

		// 获取状态
		ctx := context.Background()
		multiStatuses, err := multiMigrator.Status(ctx, databases)
		if err != nil {
			log.Fatalf("获取迁移状态失败: %v", err)
		}

		// 显示状态
		for _, dbStatus := range multiStatuses {
			fmt.Printf("\n📊 数据库: %s\n", dbStatus.Database)
			fmt.Println("---------------------------------------------------------------")

			if len(dbStatus.Statuses) == 0 {
				fmt.Println("  📭 暂无迁移记录")
				continue
			}

			for _, status := range dbStatus.Statuses {
				statusIcon := "⏳"
				statusText := "待执行"
				if status.Applied {
					statusIcon = "✅"
					statusText = "已执行"
				}

				appliedTime := ""
				if status.AppliedAt != nil {
					appliedTime = status.AppliedAt.Format("2006-01-02 15:04:05")
				}

				fmt.Printf("  %s %s - %s (%s) %s\n",
					statusIcon, status.Version, status.Description, statusText, appliedTime)
			}
		}

		fmt.Println("\n🎯 状态查看完成")
	},
}

// createCmd 创建迁移命令
var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "创建新的迁移文件",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if err := createMigrationFile(name); err != nil {
			log.Fatalf("创建迁移文件失败: %v", err)
		}
	},
}

// versionCmd 版本命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("db-migrator v1.0.0")
		fmt.Println("智能数据库迁移工具")
	},
}

// 辅助函数

func createDefaultConfig() error {
	if _, err := os.Stat("config.yaml"); err == nil {
		if verbose {
			fmt.Println("配置文件已存在，跳过创建")
		}
		return nil
	}

	configContent := `# 数据库迁移工具配置文件
database:
  driver: mysql
  host: localhost
  port: 3306
  username: root
  password: ""
  database: your_database
  charset: utf8mb4

migrator:
  migrations_table: schema_migrations
  lock_table: schema_migrations_lock
  auto_backup: false
  dry_run: false
`

	return os.WriteFile("config.yaml", []byte(configContent), 0644)
}

func createMigrationsDir() error {
	dir := "migrations"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

func loadMigrations(m *migrator.Migrator) error {
	// 这里应该加载 migrations 目录下的迁移文件
	// 由于我们使用 Go 代码定义迁移，这里需要动态加载
	// 暂时返回 nil，实际使用时需要根据具体需求实现
	return nil
}

func createMigrationFile(name string) error {
	// 生成时间戳
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	filename := fmt.Sprintf("migrations/%s_%s.go", timestamp, name)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	// 模板内容
	template := `package migrations

import (
	"context"
	"db-migrator/internal/builder"
	"db-migrator/internal/types"
)

// %sMigration %s迁移
type %sMigration struct{}

// Version 返回迁移版本
func (m *%sMigration) Version() string {
	return "%s"
}

// Description 返回迁移描述
func (m *%sMigration) Description() string {
	return "%s"
}

// Up 执行向上迁移
func (m *%sMigration) Up(ctx context.Context, db types.DB) error {
	builder := builder.NewSQLBuilder(nil, db) // 注意：这里需要传入 checker
	
	// 示例：创建表
	// err := builder.CreateTableIfNotExists(ctx, "users", ` + "`" + `
	//     CREATE TABLE users (
	//         id INT PRIMARY KEY AUTO_INCREMENT,
	//         username VARCHAR(50) NOT NULL UNIQUE,
	//         email VARCHAR(100) NOT NULL UNIQUE,
	//         created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	//     ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	// ` + "`" + `)
	// if err != nil {
	//     return err
	// }

	// 示例：添加列
	// err = builder.AddColumnIfNotExists(ctx, "users", "phone", "VARCHAR(20)")
	// if err != nil {
	//     return err
	// }

	// 示例：创建索引
	// err = builder.CreateIndexIfNotExists(ctx, "users", "idx_username", 
	//     "CREATE INDEX idx_username ON users(username)")
	// if err != nil {
	//     return err
	// }

	return nil
}

// Down 执行向下迁移（回滚）
func (m *%sMigration) Down(ctx context.Context, db types.DB) error {
	builder := builder.NewSQLBuilder(nil, db) // 注意：这里需要传入 checker

	// 示例：删除索引
	// err := builder.DropIndexIfExists(ctx, "users", "idx_username")
	// if err != nil {
	//     return err
	// }

	// 示例：删除列
	// err = builder.DropColumnIfExists(ctx, "users", "phone")
	// if err != nil {
	//     return err
	// }

	// 示例：删除表
	// _, err = db.Exec("DROP TABLE IF EXISTS users")
	// if err != nil {
	//     return err
	// }

	return nil
}
`

	content := fmt.Sprintf(template,
		name, name, name, name, timestamp, name, name, name, name)

	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return err
	}

	fmt.Printf("✅ 创建迁移文件: %s\n", filename)
	fmt.Println("🚀 请编辑文件并实现 Up() 和 Down() 方法")
	return nil
}

// 多数据库支持的辅助函数

// createMultiMigrator 创建多数据库迁移器
func createMultiMigrator() (*migrator.MultiMigrator, error) {
	multiMigrator := migrator.NewMultiMigrator(config)

	// 这里可以注册迁移文件
	// TODO: 实现从目录加载迁移文件的逻辑

	return multiMigrator, nil
}

// resolveDatabases 解析要操作的数据库列表
func resolveDatabases() ([]string, error) {
	ctx := context.Background()

	// 如果指定了操作所有数据库
	if allDatabases {
		multiMigrator, err := createMultiMigrator()
		if err != nil {
			return nil, err
		}
		defer multiMigrator.Close()

		// 发现所有可用数据库
		patterns := config.Migrator.DatabasePatterns
		if len(databasePatterns) > 0 {
			patterns = databasePatterns
		}

		return multiMigrator.GetMatchedDatabases(ctx, patterns)
	}

	// 如果指定了数据库模式
	if len(databasePatterns) > 0 {
		multiMigrator, err := createMultiMigrator()
		if err != nil {
			return nil, err
		}
		defer multiMigrator.Close()

		return multiMigrator.GetMatchedDatabases(ctx, databasePatterns)
	}

	// 如果指定了多个数据库
	if len(targetDatabases) > 0 {
		return targetDatabases, nil
	}

	// 如果指定了单个数据库
	if targetDatabase != "" {
		return []string{targetDatabase}, nil
	}

	// 返回空列表，让MultiMigrator使用默认数据库
	return []string{}, nil
}

// addDatabaseFlags 为命令添加数据库相关的参数
func addDatabaseFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&targetDatabase, "database", "d", "", "指定目标数据库")
	cmd.Flags().StringSliceVar(&targetDatabases, "databases", []string{}, "指定多个目标数据库（逗号分隔）")
	cmd.Flags().StringSliceVar(&databasePatterns, "patterns", []string{}, "数据库名匹配模式（支持通配符，如 shop*）")
	cmd.Flags().BoolVar(&allDatabases, "all", false, "操作所有配置的数据库")
}

// validateDatabaseFlags 验证数据库参数
func validateDatabaseFlags() error {
	flagCount := 0

	if targetDatabase != "" {
		flagCount++
	}
	if len(targetDatabases) > 0 {
		flagCount++
	}
	if len(databasePatterns) > 0 {
		flagCount++
	}
	if allDatabases {
		flagCount++
	}

	if flagCount > 1 {
		return fmt.Errorf("只能指定一种数据库选择方式：--database, --databases, --patterns, 或 --all")
	}

	return nil
}

// printDatabaseInfo 打印数据库信息
func printDatabaseInfo(databases []string) {
	if len(databases) == 0 {
		fmt.Println("📊 使用默认数据库")
		return
	}

	if len(databases) == 1 {
		fmt.Printf("📊 目标数据库: %s\n", databases[0])
	} else {
		fmt.Printf("📊 目标数据库 (%d个):\n", len(databases))
		for _, db := range databases {
			fmt.Printf("   • %s\n", db)
		}
	}
}
