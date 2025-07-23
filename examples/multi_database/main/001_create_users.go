package main

import (
	"context"
	"db-migrator/internal/builder"
	"db-migrator/internal/checker"
	"db-migrator/internal/types"
)

// CreateUsersTableMigration 创建用户表迁移（主数据库）
type CreateUsersTableMigration struct{}

func (m *CreateUsersTableMigration) Version() string {
	return "001"
}

func (m *CreateUsersTableMigration) Description() string {
	return "在主数据库创建用户表"
}

func (m *CreateUsersTableMigration) Up(ctx context.Context, db types.DB) error {
	checker := checker.NewMySQLChecker(db, "app_main_db")
	advancedBuilder := builder.NewAdvancedBuilder(checker, db)

	// 创建用户表
	return advancedBuilder.Table("users").
		ID().
		String("email", 255).NotNull().Unique().Comment("邮箱").End().
		String("password_hash", 255).NotNull().Comment("密码哈希").End().
		String("first_name", 100).Nullable().Comment("名").End().
		String("last_name", 100).Nullable().Comment("姓").End().
		String("phone", 20).Nullable().Comment("电话").End().
		Enum("status", []string{"active", "inactive", "suspended"}).Default("active").Comment("状态").End().
		Enum("role", []string{"admin", "user", "manager"}).Default("user").Comment("角色").End().
		Json("profile").Nullable().Comment("用户档案").End().
		Timestamp("email_verified_at").Nullable().Comment("邮箱验证时间").End().
		Timestamp("last_login_at").Nullable().Comment("最后登录时间").End().
		String("last_login_ip", 45).Nullable().Comment("最后登录IP").End().
		Timestamps().
		Index("email").End().
		Index("status").End().
		Index("role").End().
		Engine("InnoDB").
		Comment("用户主表").
		Create(ctx)
}

func (m *CreateUsersTableMigration) Down(ctx context.Context, db types.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS users")
	return err
}
