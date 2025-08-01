package migrations

import (
	"context"
	"github.com/xiezhihuan/db-migrator/internal/builder"
	"github.com/xiezhihuan/db-migrator/internal/checker"
	"github.com/xiezhihuan/db-migrator/internal/types"
	"fmt"
)

// AddUserProfileMigration 添加用户资料功能
type AddUserProfileMigration struct{}

// Version 返回迁移版本
func (m *AddUserProfileMigration) Version() string {
	return "002"
}

// Description 返回迁移描述
func (m *AddUserProfileMigration) Description() string {
	return "添加用户资料相关字段和函数"
}

// Up 执行向上迁移
func (m *AddUserProfileMigration) Up(ctx context.Context, db types.DB) error {
	// 创建检查器
	checker := checker.NewMySQLChecker(db, "your_database") // 替换为实际数据库名
	builder := builder.NewSQLBuilder(checker, db)

	// 添加头像字段
	err := builder.AddColumnIfNotExists(ctx, "users", "avatar", "VARCHAR(255)")
	if err != nil {
		return fmt.Errorf("添加头像字段失败: %v", err)
	}

	// 添加生日字段
	err = builder.AddColumnIfNotExists(ctx, "users", "birthday", "DATE")
	if err != nil {
		return fmt.Errorf("添加生日字段失败: %v", err)
	}

	// 添加个人介绍字段
	err = builder.AddColumnIfNotExists(ctx, "users", "bio", "TEXT")
	if err != nil {
		return fmt.Errorf("添加个人介绍字段失败: %v", err)
	}

	// 添加最后登录时间字段
	err = builder.AddColumnIfNotExists(ctx, "users", "last_login_at", "TIMESTAMP NULL")
	if err != nil {
		return fmt.Errorf("添加最后登录时间字段失败: %v", err)
	}

	// 创建用户资料表
	err = builder.CreateTableIfNotExists(ctx, "user_profiles", `
		CREATE TABLE user_profiles (
			id INT PRIMARY KEY AUTO_INCREMENT,
			user_id INT NOT NULL,
			address TEXT,
			city VARCHAR(50),
			country VARCHAR(50),
			postal_code VARCHAR(20),
			website VARCHAR(255),
			social_links JSON,
			preferences JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE KEY uk_user_profiles_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户资料表'
	`)
	if err != nil {
		return fmt.Errorf("创建用户资料表失败: %v", err)
	}

	// 创建计算用户年龄的函数
	err = builder.CreateFunctionIfNotExists(ctx, "calculate_user_age", `
		CREATE FUNCTION calculate_user_age(birth_date DATE) 
		RETURNS INT
		READS SQL DATA
		DETERMINISTIC
		BEGIN
			DECLARE age INT;
			IF birth_date IS NULL THEN
				RETURN NULL;
			END IF;
			SET age = TIMESTAMPDIFF(YEAR, birth_date, CURDATE());
			RETURN age;
		END
	`)
	if err != nil {
		return fmt.Errorf("创建年龄计算函数失败: %v", err)
	}

	// 创建格式化用户显示名的函数
	err = builder.CreateFunctionIfNotExists(ctx, "format_user_display_name", `
		CREATE FUNCTION format_user_display_name(full_name VARCHAR(100), username VARCHAR(50)) 
		RETURNS VARCHAR(100)
		READS SQL DATA
		DETERMINISTIC
		BEGIN
			IF full_name IS NOT NULL AND TRIM(full_name) != '' THEN
				RETURN full_name;
			ELSE
				RETURN username;
			END IF;
		END
	`)
	if err != nil {
		return fmt.Errorf("创建显示名格式化函数失败: %v", err)
	}

	return nil
}

// Down 执行向下迁移（回滚）
func (m *AddUserProfileMigration) Down(ctx context.Context, db types.DB) error {
	// 创建检查器
	checker := checker.NewMySQLChecker(db, "your_database") // 替换为实际数据库名
	builder := builder.NewSQLBuilder(checker, db)

	// 删除函数
	err := builder.DropFunctionIfExists(ctx, "format_user_display_name")
	if err != nil {
		return fmt.Errorf("删除显示名格式化函数失败: %v", err)
	}

	err = builder.DropFunctionIfExists(ctx, "calculate_user_age")
	if err != nil {
		return fmt.Errorf("删除年龄计算函数失败: %v", err)
	}

	// 删除用户资料表
	_, err = db.Exec("DROP TABLE IF EXISTS user_profiles")
	if err != nil {
		return fmt.Errorf("删除用户资料表失败: %v", err)
	}

	// 删除添加的列
	err = builder.DropColumnIfExists(ctx, "users", "last_login_at")
	if err != nil {
		return fmt.Errorf("删除最后登录时间字段失败: %v", err)
	}

	err = builder.DropColumnIfExists(ctx, "users", "bio")
	if err != nil {
		return fmt.Errorf("删除个人介绍字段失败: %v", err)
	}

	err = builder.DropColumnIfExists(ctx, "users", "birthday")
	if err != nil {
		return fmt.Errorf("删除生日字段失败: %v", err)
	}

	err = builder.DropColumnIfExists(ctx, "users", "avatar")
	if err != nil {
		return fmt.Errorf("删除头像字段失败: %v", err)
	}

	return nil
}
