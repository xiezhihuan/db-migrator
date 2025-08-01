// Package examples 展示如何在迁移中使用数据初始化功能
package examples

import (
	"context"
	"github.com/xiezhihuan/db-migrator/internal/builder"
	"github.com/xiezhihuan/db-migrator/internal/checker"
	"github.com/xiezhihuan/db-migrator/internal/types"
	"fmt"
)

// DataInitializationMigration 演示数据初始化的迁移示例
type DataInitializationMigration struct{}

func (m *DataInitializationMigration) Version() string {
	return "20240101_001"
}

func (m *DataInitializationMigration) Description() string {
	return "演示数据初始化功能"
}

func (m *DataInitializationMigration) Up(ctx context.Context, db types.DB) error {
	checker := checker.NewMySQLChecker(db, "")

	// 1. 创建表结构
	advancedBuilder := builder.NewAdvancedBuilder(checker, db)

	// 创建用户角色表
	err := advancedBuilder.Table("user_roles").
		ID().
		String("name", 50).NotNull().Unique().Comment("角色名称").End().
		String("code", 20).NotNull().Unique().Comment("角色代码").End().
		Text("description").Nullable().Comment("角色描述").End().
		Boolean("is_active").Default(true).Comment("是否激活").End().
		Timestamps().
		Create(ctx)
	if err != nil {
		return err
	}

	// 创建系统配置表
	err = advancedBuilder.Table("system_configs").
		ID().
		String("category", 50).NotNull().Comment("配置分类").End().
		String("key", 100).NotNull().Comment("配置键").End().
		Text("value").NotNull().Comment("配置值").End().
		String("type", 20).Default("string").Comment("值类型").End().
		Boolean("editable").Default(true).Comment("是否可编辑").End().
		Text("description").Nullable().Comment("配置描述").End().
		Timestamps().
		Index("category").End().
		Index("key").End().
		Unique("category", "key").End().
		Create(ctx)
	if err != nil {
		return err
	}

	// 2. 初始化数据
	dataBuilder := builder.NewDataBuilder(checker, db)

	// 初始化用户角色数据
	err = m.initUserRoles(ctx, dataBuilder)
	if err != nil {
		return err
	}

	// 初始化系统配置数据
	err = m.initSystemConfigs(ctx, dataBuilder)
	if err != nil {
		return err
	}

	// 从JSON文件初始化基础数据（如果文件存在）
	err = m.initDataFromFiles(ctx, dataBuilder)
	if err != nil {
		return err
	}

	return nil
}

func (m *DataInitializationMigration) Down(ctx context.Context, db types.DB) error {
	// 删除表（按依赖关系顺序）
	tables := []string{"system_configs", "user_roles"}
	for _, table := range tables {
		sql := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
		_, err := db.Exec(sql)
		if err != nil {
			return err
		}
	}

	return nil
}

// initUserRoles 初始化用户角色数据
func (m *DataInitializationMigration) initUserRoles(ctx context.Context, dataBuilder *builder.DataBuilder) error {
	roles := []map[string]interface{}{
		{
			"id":          1,
			"name":        "超级管理员",
			"code":        "super_admin",
			"description": "系统超级管理员，拥有所有权限",
			"is_active":   true,
		},
		{
			"id":          2,
			"name":        "管理员",
			"code":        "admin",
			"description": "系统管理员，拥有管理权限",
			"is_active":   true,
		},
		{
			"id":          3,
			"name":        "编辑者",
			"code":        "editor",
			"description": "内容编辑者，可以编辑内容",
			"is_active":   true,
		},
		{
			"id":          4,
			"name":        "查看者",
			"code":        "viewer",
			"description": "只读用户，只能查看内容",
			"is_active":   true,
		},
	}

	// 使用智能插入策略，存在则更新，不存在则插入
	return dataBuilder.Table("user_roles").
		Strategy(builder.StrategyInsertOrUpdate).
		BatchSize(100).
		InsertData(ctx, roles)
}

// initSystemConfigs 初始化系统配置数据
func (m *DataInitializationMigration) initSystemConfigs(ctx context.Context, dataBuilder *builder.DataBuilder) error {
	configs := []map[string]interface{}{
		// 系统基础配置
		{
			"category":    "system",
			"key":         "app_name",
			"value":       "数据库迁移工具",
			"type":        "string",
			"editable":    true,
			"description": "应用程序名称",
		},
		{
			"category":    "system",
			"key":         "app_version",
			"value":       "1.0.0",
			"type":        "string",
			"editable":    false,
			"description": "应用程序版本",
		},
		{
			"category":    "system",
			"key":         "timezone",
			"value":       "Asia/Shanghai",
			"type":        "string",
			"editable":    true,
			"description": "系统时区",
		},

		// 功能开关配置
		{
			"category":    "features",
			"key":         "enable_registration",
			"value":       "true",
			"type":        "boolean",
			"editable":    true,
			"description": "是否开放用户注册",
		},
		{
			"category":    "features",
			"key":         "enable_email_verification",
			"value":       "false",
			"type":        "boolean",
			"editable":    true,
			"description": "是否开启邮箱验证",
		},
		{
			"category":    "features",
			"key":         "maintenance_mode",
			"value":       "false",
			"type":        "boolean",
			"editable":    true,
			"description": "维护模式开关",
		},

		// 限制配置
		{
			"category":    "limits",
			"key":         "max_login_attempts",
			"value":       "5",
			"type":        "integer",
			"editable":    true,
			"description": "最大登录尝试次数",
		},
		{
			"category":    "limits",
			"key":         "session_timeout",
			"value":       "3600",
			"type":        "integer",
			"editable":    true,
			"description": "会话超时时间（秒）",
		},
		{
			"category":    "limits",
			"key":         "max_file_size",
			"value":       "10485760",
			"type":        "integer",
			"editable":    true,
			"description": "最大文件上传大小（字节）",
		},

		// 邮件配置
		{
			"category":    "email",
			"key":         "smtp_host",
			"value":       "localhost",
			"type":        "string",
			"editable":    true,
			"description": "SMTP服务器地址",
		},
		{
			"category":    "email",
			"key":         "smtp_port",
			"value":       "587",
			"type":        "integer",
			"editable":    true,
			"description": "SMTP服务器端口",
		},
		{
			"category":    "email",
			"key":         "smtp_encryption",
			"value":       "tls",
			"type":        "string",
			"editable":    true,
			"description": "SMTP加密方式",
		},
	}

	// 使用智能插入策略
	return dataBuilder.Table("system_configs").
		Strategy(builder.StrategyInsertOrUpdate).
		BatchSize(50).
		InsertData(ctx, configs)
}

// initDataFromFiles 从文件初始化数据（可选）
func (m *DataInitializationMigration) initDataFromFiles(ctx context.Context, dataBuilder *builder.DataBuilder) error {
	// 尝试从JSON文件初始化用户数据
	err := dataBuilder.QuickInsertFromJSON(ctx, "users", "data/initial_users.json")
	if err != nil {
		// 文件不存在或读取失败时，记录日志但不返回错误
		// log.Printf("JSON文件初始化失败，跳过: %v", err)
	}

	// 尝试从YAML文件初始化其他配置
	err = dataBuilder.QuickInsertFromYAML(ctx, "additional_configs", "data/extra_configs.yaml")
	if err != nil {
		// 文件不存在或读取失败时，记录日志但不返回错误
		// log.Printf("YAML文件初始化失败，跳过: %v", err)
	}

	return nil
}

// 演示结构体数据插入
type User struct {
	ID       int    `db:"id"`
	Username string `db:"username"`
	Email    string `db:"email"`
	RoleID   int    `db:"role_id"`
	IsActive bool   `db:"is_active"`
}

// initUsersFromStruct 演示从Go结构体初始化数据
func (m *DataInitializationMigration) initUsersFromStruct(ctx context.Context, dataBuilder *builder.DataBuilder) error {
	users := []User{
		{
			ID:       1,
			Username: "admin",
			Email:    "admin@example.com",
			RoleID:   1,
			IsActive: true,
		},
		{
			ID:       2,
			Username: "editor",
			Email:    "editor@example.com",
			RoleID:   3,
			IsActive: true,
		},
	}

	// 从结构体插入数据
	return dataBuilder.Table("users").
		Strategy(builder.StrategyInsertOrUpdate).
		InsertFromStruct(ctx, users)
}

// 演示高级数据操作
func (m *DataInitializationMigration) advancedDataOperations(ctx context.Context, dataBuilder *builder.DataBuilder) error {
	// 1. 条件更新配置
	err := dataBuilder.Table("system_configs").
		Strategy(builder.StrategyInsertOrUpdate).
		Where("category = 'system' AND editable = true").
		InsertData(ctx, []map[string]interface{}{
			{
				"category": "system",
				"key":      "updated_at",
				"value":    "NOW()",
				"type":     "datetime",
			},
		})
	if err != nil {
		return err
	}

	// 2. 批量插入大量数据
	largeDataset := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		largeDataset[i] = map[string]interface{}{
			"category":    "batch",
			"key":         fmt.Sprintf("batch_item_%d", i),
			"value":       fmt.Sprintf("value_%d", i),
			"type":        "string",
			"description": fmt.Sprintf("批量数据项 %d", i),
		}
	}

	err = dataBuilder.Table("system_configs").
		Strategy(builder.StrategyInsertOnly).
		BatchSize(200). // 设置批量大小
		InsertData(ctx, largeDataset)
	if err != nil {
		return err
	}

	// 3. 直接执行SQL插入
	err = dataBuilder.Table("system_configs").
		InsertSQL(ctx,
			"INSERT IGNORE INTO system_configs (category, `key`, value, type, description) VALUES (?, ?, ?, ?, ?)",
			"custom", "sql_insert", "true", "boolean", "通过SQL直接插入的配置")
	if err != nil {
		return err
	}

	return nil
}

// 演示数据清理和重置
func (m *DataInitializationMigration) cleanupAndReset(ctx context.Context, dataBuilder *builder.DataBuilder) error {
	// 清空表并重新插入数据
	configs := []map[string]interface{}{
		{
			"category":    "reset",
			"key":         "reset_time",
			"value":       "NOW()",
			"type":        "datetime",
			"description": "重置时间",
		},
	}

	return dataBuilder.Table("system_configs").
		Strategy(builder.StrategyTruncateAndInsert). // 清空表后插入
		InsertData(ctx, configs)
}

// 使用方法：
// 1. 将此文件放在 migrations/ 目录下
// 2. 在 main.go 中注册这个迁移：
//    migrator.RegisterMigration(&DataInitializationMigration{})
// 3. 运行迁移：
//    db-migrator up
