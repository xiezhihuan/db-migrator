package migrator

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"db-migrator/internal/checker"
	"db-migrator/internal/database"
	"db-migrator/internal/types"
)

// MultiMigrator 多数据库迁移器
type MultiMigrator struct {
	config     types.Config
	dbManager  *database.Manager
	migrators  map[string]*Migrator
	migrations []types.Migration
}

// NewMultiMigrator 创建多数据库迁移器
func NewMultiMigrator(config types.Config) *MultiMigrator {
	return &MultiMigrator{
		config:     config,
		dbManager:  database.NewManager(config),
		migrators:  make(map[string]*Migrator),
		migrations: make([]types.Migration, 0),
	}
}

// RegisterMigration 注册迁移
func (mm *MultiMigrator) RegisterMigration(migration types.Migration) {
	mm.migrations = append(mm.migrations, migration)
}

// GetMigrator 获取指定数据库的迁移器
func (mm *MultiMigrator) GetMigrator(dbName string) (*Migrator, error) {
	// 如果已存在，直接返回
	if migrator, exists := mm.migrators[dbName]; exists {
		return migrator, nil
	}

	// 获取数据库连接
	db, err := mm.dbManager.GetDatabase(dbName)
	if err != nil {
		return nil, err
	}

	// 创建检查器
	checker := checker.NewMySQLChecker(db, dbName)

	// 创建迁移器
	migrator := NewMigrator(db, checker, mm.config.Migrator)

	// 注册所有迁移到这个迁移器
	for _, migration := range mm.migrations {
		if mm.shouldApplyToDatabase(migration, dbName) {
			migrator.RegisterMigration(migration)
		}
	}

	// 缓存迁移器
	mm.migrators[dbName] = migrator
	return migrator, nil
}

// Up 执行向上迁移
func (mm *MultiMigrator) Up(ctx context.Context, databases []string) error {
	if len(databases) == 0 {
		// 使用默认数据库
		_, defaultDB, err := mm.dbManager.GetDefaultDatabase()
		if err != nil {
			return err
		}
		databases = []string{defaultDB}
	}

	var errors []string
	for _, dbName := range databases {
		fmt.Printf("\n🗄️  正在迁移数据库: %s\n", dbName)

		migrator, err := mm.GetMigrator(dbName)
		if err != nil {
			errors = append(errors, fmt.Sprintf("数据库 %s: %v", dbName, err))
			continue
		}

		if err := migrator.Up(ctx); err != nil {
			errors = append(errors, fmt.Sprintf("数据库 %s: %v", dbName, err))
			continue
		}

		fmt.Printf("✅ 数据库 %s 迁移完成\n", dbName)
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分数据库迁移失败:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// Down 执行向下迁移
func (mm *MultiMigrator) Down(ctx context.Context, databases []string, steps int) error {
	if len(databases) == 0 {
		// 使用默认数据库
		_, defaultDB, err := mm.dbManager.GetDefaultDatabase()
		if err != nil {
			return err
		}
		databases = []string{defaultDB}
	}

	var errors []string
	for _, dbName := range databases {
		fmt.Printf("\n🔄 正在回滚数据库: %s\n", dbName)

		migrator, err := mm.GetMigrator(dbName)
		if err != nil {
			errors = append(errors, fmt.Sprintf("数据库 %s: %v", dbName, err))
			continue
		}

		if err := migrator.Down(ctx, steps); err != nil {
			errors = append(errors, fmt.Sprintf("数据库 %s: %v", dbName, err))
			continue
		}

		fmt.Printf("✅ 数据库 %s 回滚完成\n", dbName)
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分数据库回滚失败:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// Status 获取迁移状态
func (mm *MultiMigrator) Status(ctx context.Context, databases []string) ([]types.MultiDatabaseStatus, error) {
	if len(databases) == 0 {
		// 使用默认数据库
		_, defaultDB, err := mm.dbManager.GetDefaultDatabase()
		if err != nil {
			return nil, err
		}
		databases = []string{defaultDB}
	}

	var results []types.MultiDatabaseStatus
	for _, dbName := range databases {
		migrator, err := mm.GetMigrator(dbName)
		if err != nil {
			// 如果无法获取迁移器，返回错误状态
			results = append(results, types.MultiDatabaseStatus{
				Database: dbName,
				Statuses: []types.MigrationStatus{
					{
						Version:     "ERROR",
						Description: fmt.Sprintf("无法连接数据库: %v", err),
						Applied:     false,
						Database:    dbName,
					},
				},
			})
			continue
		}

		statuses, err := migrator.Status(ctx)
		if err != nil {
			results = append(results, types.MultiDatabaseStatus{
				Database: dbName,
				Statuses: []types.MigrationStatus{
					{
						Version:     "ERROR",
						Description: fmt.Sprintf("获取迁移状态失败: %v", err),
						Applied:     false,
						Database:    dbName,
					},
				},
			})
			continue
		}

		// 为每个状态添加数据库信息
		for i := range statuses {
			statuses[i].Database = dbName
		}

		results = append(results, types.MultiDatabaseStatus{
			Database: dbName,
			Statuses: statuses,
		})
	}

	return results, nil
}

// DiscoverDatabases 发现数据库
func (mm *MultiMigrator) DiscoverDatabases(ctx context.Context, patterns []string) ([]types.DatabaseInfo, error) {
	return mm.dbManager.DiscoverDatabases(ctx, patterns)
}

// GetMatchedDatabases 获取匹配的数据库列表
func (mm *MultiMigrator) GetMatchedDatabases(ctx context.Context, patterns []string) ([]string, error) {
	return mm.dbManager.GetMatchedDatabases(ctx, patterns)
}

// Close 关闭所有连接
func (mm *MultiMigrator) Close() error {
	return mm.dbManager.CloseAll()
}

// shouldApplyToDatabase 判断迁移是否应该应用到指定数据库
func (mm *MultiMigrator) shouldApplyToDatabase(migration types.Migration, dbName string) bool {
	// 检查是否实现了多数据库接口
	if multiMigration, ok := migration.(types.MultiDatabaseMigration); ok {
		// 检查 Database() 方法
		if targetDB := multiMigration.Database(); targetDB != "" {
			return targetDB == dbName
		}

		// 检查 Databases() 方法
		if targetDBs := multiMigration.Databases(); len(targetDBs) > 0 {
			for _, target := range targetDBs {
				if target == dbName {
					return true
				}
			}
			return false
		}
	}

	// 检查迁移文件路径（基于目录组织）
	migrationPath := mm.getMigrationPath(migration)
	if migrationPath != "" {
		// 从路径提取数据库名
		pathDB := mm.extractDatabaseFromPath(migrationPath)
		if pathDB != "" {
			return pathDB == dbName
		}
	}

	// 默认应用到默认数据库
	_, defaultDB, err := mm.dbManager.GetDefaultDatabase()
	if err != nil {
		return false
	}

	return dbName == defaultDB
}

// getMigrationPath 获取迁移文件路径（这里简化处理）
func (mm *MultiMigrator) getMigrationPath(migration types.Migration) string {
	// 这里需要根据实际情况实现
	// 可以通过反射获取类型信息，或者添加路径接口
	return ""
}

// extractDatabaseFromPath 从路径提取数据库名
func (mm *MultiMigrator) extractDatabaseFromPath(path string) string {
	// 假设路径格式为: migrations/database_name/xxx.go
	dir := filepath.Dir(path)
	baseName := filepath.Base(dir)

	// 如果是migrations目录下的子目录，则认为是数据库名
	parentDir := filepath.Dir(dir)
	if filepath.Base(parentDir) == mm.config.Migrator.MigrationsDir {
		return baseName
	}

	return ""
}

// LoadMigrationsFromDirectory 从目录加载迁移（支持多数据库目录结构）
func (mm *MultiMigrator) LoadMigrationsFromDirectory(baseDir string) error {
	// 这个方法需要根据实际的迁移文件加载逻辑来实现
	// 这里提供基本框架
	fmt.Printf("从目录 %s 加载迁移文件...\n", baseDir)
	return nil
}
