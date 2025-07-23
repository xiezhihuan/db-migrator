package migrations

import (
	"context"
	"db-migrator/internal/builder"
	"db-migrator/internal/checker"
	"db-migrator/internal/types"
	"fmt"
)

// CreateUsersTableMigration 创建用户表迁移
type CreateUsersTableMigration struct{}

// Version 返回迁移版本
func (m *CreateUsersTableMigration) Version() string {
	return "001"
}

// Description 返回迁移描述
func (m *CreateUsersTableMigration) Description() string {
	return "创建用户表和相关索引"
}

// Up 执行向上迁移
func (m *CreateUsersTableMigration) Up(ctx context.Context, db types.DB) error {
	// 创建检查器（实际使用时从外部传入）
	checker := checker.NewMySQLChecker(db, "your_database") // 替换为实际数据库名
	builder := builder.NewSQLBuilder(checker, db)

	// 创建用户表
	err := builder.CreateTableIfNotExists(ctx, "users", `
		CREATE TABLE users (
			id INT PRIMARY KEY AUTO_INCREMENT,
			username VARCHAR(50) NOT NULL UNIQUE,
			email VARCHAR(100) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			full_name VARCHAR(100),
			phone VARCHAR(20),
			status ENUM('active', 'inactive', 'suspended') DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表'
	`)
	if err != nil {
		return fmt.Errorf("创建用户表失败: %v", err)
	}

	// 创建邮箱索引
	err = builder.CreateIndexIfNotExists(ctx, "users", "idx_users_email",
		"CREATE INDEX idx_users_email ON users(email)")
	if err != nil {
		return fmt.Errorf("创建邮箱索引失败: %v", err)
	}

	// 创建用户名索引
	err = builder.CreateIndexIfNotExists(ctx, "users", "idx_users_username",
		"CREATE INDEX idx_users_username ON users(username)")
	if err != nil {
		return fmt.Errorf("创建用户名索引失败: %v", err)
	}

	// 创建状态索引
	err = builder.CreateIndexIfNotExists(ctx, "users", "idx_users_status",
		"CREATE INDEX idx_users_status ON users(status)")
	if err != nil {
		return fmt.Errorf("创建状态索引失败: %v", err)
	}

	// 插入默认管理员用户（如果不存在）
	err = builder.InsertIfNotExists(ctx, "users", "username = 'admin'",
		`INSERT INTO users (username, email, password_hash, full_name, status) 
		 VALUES ('admin', 'admin@example.com', 'hashed_password_here', '系统管理员', 'active')`)
	if err != nil {
		return fmt.Errorf("插入默认管理员失败: %v", err)
	}

	return nil
}

// Down 执行向下迁移（回滚）
func (m *CreateUsersTableMigration) Down(ctx context.Context, db types.DB) error {
	// 创建检查器
	checker := checker.NewMySQLChecker(db, "your_database") // 替换为实际数据库名
	builder := builder.NewSQLBuilder(checker, db)

	// 删除索引
	err := builder.DropIndexIfExists(ctx, "users", "idx_users_status")
	if err != nil {
		return fmt.Errorf("删除状态索引失败: %v", err)
	}

	err = builder.DropIndexIfExists(ctx, "users", "idx_users_username")
	if err != nil {
		return fmt.Errorf("删除用户名索引失败: %v", err)
	}

	err = builder.DropIndexIfExists(ctx, "users", "idx_users_email")
	if err != nil {
		return fmt.Errorf("删除邮箱索引失败: %v", err)
	}

	// 删除用户表
	_, err = db.Exec("DROP TABLE IF EXISTS users")
	if err != nil {
		return fmt.Errorf("删除用户表失败: %v", err)
	}

	return nil
}
