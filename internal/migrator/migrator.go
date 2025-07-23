package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"time"

	"db-migrator/internal/types"
)

// Migrator 迁移器
type Migrator struct {
	db              types.DB
	checker         types.Checker
	config          types.MigratorConfig
	migrations      []types.Migration
	migrationsTable string
	lockTable       string
}

// NewMigrator 创建迁移器
func NewMigrator(db types.DB, checker types.Checker, config types.MigratorConfig) *Migrator {
	migrationsTable := config.MigrationsTable
	if migrationsTable == "" {
		migrationsTable = "schema_migrations"
	}

	lockTable := config.LockTable
	if lockTable == "" {
		lockTable = "schema_migrations_lock"
	}

	return &Migrator{
		db:              db,
		checker:         checker,
		config:          config,
		migrations:      make([]types.Migration, 0),
		migrationsTable: migrationsTable,
		lockTable:       lockTable,
	}
}

// RegisterMigration 注册迁移
func (m *Migrator) RegisterMigration(migration types.Migration) {
	m.migrations = append(m.migrations, migration)
}

// RegisterMigrations 批量注册迁移
func (m *Migrator) RegisterMigrations(migrations ...types.Migration) {
	m.migrations = append(m.migrations, migrations...)
}

// Init 初始化迁移器（创建必要的系统表）
func (m *Migrator) Init(ctx context.Context) error {
	log.Println("正在初始化迁移器...")

	// 创建迁移记录表
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("创建迁移记录表失败: %v", err)
	}

	// 创建锁表
	if err := m.createLockTable(ctx); err != nil {
		return fmt.Errorf("创建锁表失败: %v", err)
	}

	log.Println("迁移器初始化完成")
	return nil
}

// Up 执行所有待执行的迁移
func (m *Migrator) Up(ctx context.Context) error {
	// 获取锁
	if err := m.acquireLock(ctx); err != nil {
		return fmt.Errorf("获取迁移锁失败: %v", err)
	}
	defer m.releaseLock(ctx)

	// 排序迁移
	m.sortMigrations()

	// 获取已执行的迁移
	appliedMigrations, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("获取已执行迁移失败: %v", err)
	}

	log.Printf("找到 %d 个迁移，已执行 %d 个", len(m.migrations), len(appliedMigrations))

	// 执行待处理的迁移
	executed := 0
	for _, migration := range m.migrations {
		if _, applied := appliedMigrations[migration.Version()]; applied {
			log.Printf("跳过已执行的迁移: %s - %s", migration.Version(), migration.Description())
			continue
		}

		if err := m.executeMigration(ctx, migration, true); err != nil {
			return fmt.Errorf("执行迁移 %s 失败: %v", migration.Version(), err)
		}
		executed++
	}

	if executed == 0 {
		log.Println("没有需要执行的迁移")
	} else {
		log.Printf("成功执行了 %d 个迁移", executed)
	}

	return nil
}

// Down 回滚指定数量的迁移
func (m *Migrator) Down(ctx context.Context, steps int) error {
	if steps <= 0 {
		return fmt.Errorf("回滚步数必须大于0")
	}

	// 获取锁
	if err := m.acquireLock(ctx); err != nil {
		return fmt.Errorf("获取迁移锁失败: %v", err)
	}
	defer m.releaseLock(ctx)

	// 获取已执行的迁移（按时间倒序）
	appliedMigrations, err := m.getAppliedMigrationsOrdered(ctx, "DESC")
	if err != nil {
		return fmt.Errorf("获取已执行迁移失败: %v", err)
	}

	if len(appliedMigrations) == 0 {
		log.Println("没有可回滚的迁移")
		return nil
	}

	// 限制回滚步数
	if steps > len(appliedMigrations) {
		steps = len(appliedMigrations)
	}

	log.Printf("将回滚 %d 个迁移", steps)

	// 创建版本到迁移的映射
	migrationMap := make(map[string]types.Migration)
	for _, migration := range m.migrations {
		migrationMap[migration.Version()] = migration
	}

	// 执行回滚
	for i := 0; i < steps; i++ {
		record := appliedMigrations[i]
		migration, exists := migrationMap[record.Version]

		if !exists {
			log.Printf("警告: 未找到迁移 %s 的定义，跳过回滚", record.Version)
			continue
		}

		if err := m.executeMigration(ctx, migration, false); err != nil {
			return fmt.Errorf("回滚迁移 %s 失败: %v", migration.Version(), err)
		}
	}

	log.Printf("成功回滚了 %d 个迁移", steps)
	return nil
}

// Status 获取迁移状态
func (m *Migrator) Status(ctx context.Context) ([]types.MigrationStatus, error) {
	// 获取已执行的迁移
	appliedMigrations, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取已执行迁移失败: %v", err)
	}

	// 排序迁移
	m.sortMigrations()

	var statuses []types.MigrationStatus
	for _, migration := range m.migrations {
		status := types.MigrationStatus{
			Version:     migration.Version(),
			Description: migration.Description(),
			Applied:     false,
		}

		if record, applied := appliedMigrations[migration.Version()]; applied {
			status.Applied = true
			status.AppliedAt = &record.AppliedAt
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

// executeMigration 执行单个迁移
func (m *Migrator) executeMigration(ctx context.Context, migration types.Migration, isUp bool) error {
	version := migration.Version()
	description := migration.Description()

	action := "执行"
	if !isUp {
		action = "回滚"
	}

	log.Printf("%s迁移: %s - %s", action, version, description)

	if m.config.DryRun {
		log.Printf("干运行模式: 跳过实际%s", action)
		return nil
	}

	// 开始事务
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 记录开始时间
	startTime := time.Now()
	var migrationErr error

	// 执行迁移
	if isUp {
		migrationErr = migration.Up(ctx, &TxWrapper{tx})
	} else {
		migrationErr = migration.Down(ctx, &TxWrapper{tx})
	}

	// 更新迁移记录
	if isUp {
		if migrationErr == nil {
			// 记录成功的迁移
			err = m.recordMigration(tx, version, description, true, "")
		} else {
			// 记录失败的迁移
			err = m.recordMigration(tx, version, description, false, migrationErr.Error())
		}
	} else {
		// 删除迁移记录
		err = m.removeMigrationRecord(tx, version)
	}

	if err != nil {
		return fmt.Errorf("更新迁移记录失败: %v", err)
	}

	// 如果迁移失败，回滚事务
	if migrationErr != nil {
		return migrationErr
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	duration := time.Since(startTime)
	log.Printf("%s完成，耗时: %v", action, duration)

	return nil
}

// 创建迁移记录表
func (m *Migrator) createMigrationsTable(ctx context.Context) error {
	exists, err := m.checker.TableExists(ctx, m.migrationsTable)
	if err != nil {
		return err
	}

	if exists {
		log.Printf("迁移记录表 %s 已存在", m.migrationsTable)
		return nil
	}

	query := fmt.Sprintf(`
		CREATE TABLE %s (
			version VARCHAR(255) PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			success BOOLEAN NOT NULL DEFAULT TRUE,
			error_msg TEXT
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`, m.migrationsTable)

	_, err = m.db.Exec(query)
	if err != nil {
		return err
	}

	log.Printf("成功创建迁移记录表: %s", m.migrationsTable)
	return nil
}

// 创建锁表
func (m *Migrator) createLockTable(ctx context.Context) error {
	exists, err := m.checker.TableExists(ctx, m.lockTable)
	if err != nil {
		return err
	}

	if exists {
		log.Printf("锁表 %s 已存在", m.lockTable)
		return nil
	}

	query := fmt.Sprintf(`
		CREATE TABLE %s (
			id INT PRIMARY KEY,
			locked BOOLEAN NOT NULL DEFAULT FALSE,
			locked_at TIMESTAMP NULL,
			locked_by VARCHAR(255)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`, m.lockTable)

	_, err = m.db.Exec(query)
	if err != nil {
		return err
	}

	// 插入初始记录
	insertQuery := fmt.Sprintf(`
		INSERT INTO %s (id, locked) VALUES (1, FALSE)
	`, m.lockTable)

	_, err = m.db.Exec(insertQuery)
	if err != nil {
		return err
	}

	log.Printf("成功创建锁表: %s", m.lockTable)
	return nil
}

// 获取锁
func (m *Migrator) acquireLock(ctx context.Context) error {
	query := fmt.Sprintf(`
		UPDATE %s SET locked = TRUE, locked_at = NOW(), locked_by = ?
		WHERE id = 1 AND locked = FALSE
	`, m.lockTable)

	result, err := m.db.Exec(query, "db-migrator")
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return fmt.Errorf("无法获取迁移锁，可能有其他迁移正在进行")
	}

	return nil
}

// 释放锁
func (m *Migrator) releaseLock(ctx context.Context) error {
	query := fmt.Sprintf(`
		UPDATE %s SET locked = FALSE, locked_at = NULL, locked_by = NULL
		WHERE id = 1
	`, m.lockTable)

	_, err := m.db.Exec(query)
	return err
}

// 获取已执行的迁移
func (m *Migrator) getAppliedMigrations(ctx context.Context) (map[string]types.MigrationRecord, error) {
	query := fmt.Sprintf(`
		SELECT version, description, applied_at, success, error_msg
		FROM %s WHERE success = TRUE
		ORDER BY applied_at
	`, m.migrationsTable)

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	migrations := make(map[string]types.MigrationRecord)
	for rows.Next() {
		var record types.MigrationRecord
		var errorMsg sql.NullString

		err := rows.Scan(&record.Version, &record.Description, &record.AppliedAt,
			&record.Success, &errorMsg)
		if err != nil {
			return nil, err
		}

		record.ErrorMsg = errorMsg.String
		migrations[record.Version] = record
	}

	return migrations, nil
}

// 获取已执行的迁移（有序）
func (m *Migrator) getAppliedMigrationsOrdered(ctx context.Context, order string) ([]types.MigrationRecord, error) {
	query := fmt.Sprintf(`
		SELECT version, description, applied_at, success, error_msg
		FROM %s WHERE success = TRUE
		ORDER BY applied_at %s
	`, m.migrationsTable, order)

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []types.MigrationRecord
	for rows.Next() {
		var record types.MigrationRecord
		var errorMsg sql.NullString

		err := rows.Scan(&record.Version, &record.Description, &record.AppliedAt,
			&record.Success, &errorMsg)
		if err != nil {
			return nil, err
		}

		record.ErrorMsg = errorMsg.String
		migrations = append(migrations, record)
	}

	return migrations, nil
}

// 记录迁移
func (m *Migrator) recordMigration(tx *sql.Tx, version, description string, success bool, errorMsg string) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (version, description, success, error_msg)
		VALUES (?, ?, ?, ?)
	`, m.migrationsTable)

	_, err := tx.Exec(query, version, description, success, errorMsg)
	return err
}

// 删除迁移记录
func (m *Migrator) removeMigrationRecord(tx *sql.Tx, version string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE version = ?`, m.migrationsTable)
	_, err := tx.Exec(query, version)
	return err
}

// 排序迁移
func (m *Migrator) sortMigrations() {
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version() < m.migrations[j].Version()
	})
}

// TxWrapper 事务包装器，实现DB接口
type TxWrapper struct {
	tx *sql.Tx
}

func (tw *TxWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tw.tx.Exec(query, args...)
}

func (tw *TxWrapper) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tw.tx.Query(query, args...)
}

func (tw *TxWrapper) QueryRow(query string, args ...interface{}) *sql.Row {
	return tw.tx.QueryRow(query, args...)
}

func (tw *TxWrapper) Begin() (*sql.Tx, error) {
	return nil, fmt.Errorf("在事务中不能开始新事务")
}

func (tw *TxWrapper) Close() error {
	return nil // 事务由外部管理
}
