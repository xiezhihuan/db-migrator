package data_migration

import (
	"context"
	"fmt"
	"strings"
	"time"

	"db-migrator/internal/builder"
	"db-migrator/internal/checker"
	"db-migrator/internal/types"
)

// DataMigrationHelperMigration 数据迁移辅助工具
type DataMigrationHelperMigration struct{}

func (m *DataMigrationHelperMigration) Version() string {
	return "001"
}

func (m *DataMigrationHelperMigration) Description() string {
	return "数据迁移工具 - 从旧系统迁移用户和订单数据"
}

func (m *DataMigrationHelperMigration) Up(ctx context.Context, db types.DB) error {
	checker := checker.NewMySQLChecker(db, "migration_db")
	advancedBuilder := builder.NewAdvancedBuilder(checker, db)

	// 1. 创建数据迁移任务表
	err := advancedBuilder.Table("migration_tasks").
		ID().
		String("name", 200).NotNull().Comment("迁移任务名称").End().
		String("source_table", 100).NotNull().Comment("源表名").End().
		String("target_table", 100).NotNull().Comment("目标表名").End().
		Text("description").Nullable().Comment("任务描述").End().
		Enum("status", []string{"pending", "running", "completed", "failed", "cancelled"}).Default("pending").Comment("状态").End().
		Integer("total_records").Default(0).Comment("总记录数").End().
		Integer("processed_records").Default(0).Comment("已处理记录数").End().
		Integer("success_records").Default(0).Comment("成功记录数").End().
		Integer("failed_records").Default(0).Comment("失败记录数").End().
		Decimal("progress_percentage", 5, 2).Default(0.00).Comment("进度百分比").End().
		Json("mapping_config").Nullable().Comment("字段映射配置").End().
		Json("transform_rules").Nullable().Comment("数据转换规则").End().
		Text("error_log").Nullable().Comment("错误日志").End().
		Timestamp("started_at").Nullable().Comment("开始时间").End().
		Timestamp("completed_at").Nullable().Comment("完成时间").End().
		String("created_by", 36).Nullable().Comment("创建人").End().
		Timestamps().
		Index("status").End().
		Index("source_table").End().
		Index("target_table").End().
		Engine("InnoDB").
		Comment("数据迁移任务表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 2. 创建迁移错误记录表
	err = advancedBuilder.Table("migration_errors").
		ID().
		Integer("task_id").NotNull().Comment("任务ID").End().
		String("source_id", 100).Nullable().Comment("源记录ID").End().
		Json("source_data").Nullable().Comment("源数据").End().
		String("error_type", 100).NotNull().Comment("错误类型").End().
		Text("error_message").NotNull().Comment("错误消息").End().
		Text("stack_trace").Nullable().Comment("错误堆栈").End().
		Json("context").Nullable().Comment("错误上下文").End().
		Boolean("is_resolved").Default(false).Comment("是否已解决").End().
		String("resolved_by", 36).Nullable().Comment("解决人").End().
		Timestamp("resolved_at").Nullable().Comment("解决时间").End().
		Timestamps().
		ForeignKey("task_id").References("migration_tasks", "id").OnDelete(builder.ActionCascade).End().
		Index("task_id").End().
		Index("error_type").End().
		Index("is_resolved").End().
		Engine("InnoDB").
		Comment("迁移错误记录表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 3. 创建迁移映射表（用于记录新旧ID的对应关系）
	err = advancedBuilder.Table("migration_mappings").
		ID().
		Integer("task_id").NotNull().Comment("任务ID").End().
		String("entity_type", 100).NotNull().Comment("实体类型").End().
		String("source_id", 100).NotNull().Comment("源系统ID").End().
		String("target_id", 100).NotNull().Comment("目标系统ID").End().
		Json("metadata").Nullable().Comment("映射元数据").End().
		Timestamps().
		ForeignKey("task_id").References("migration_tasks", "id").OnDelete(builder.ActionCascade).End().
		Index("task_id").End().
		Index("entity_type").End().
		Index("source_id").End().
		Index("target_id").End().
		Unique("task_id", "entity_type", "source_id").End().
		Engine("InnoDB").
		Comment("迁移映射表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 4. 创建临时用户表（用于演示从旧系统迁移用户）
	err = advancedBuilder.Table("legacy_users").
		ID().
		String("old_username", 100).NotNull().Comment("旧用户名").End().
		String("old_email", 255).NotNull().Comment("旧邮箱").End().
		String("full_name", 200).Nullable().Comment("全名").End().
		String("phone_number", 20).Nullable().Comment("电话").End().
		String("address", 500).Nullable().Comment("地址").End().
		Enum("user_type", []string{"customer", "admin", "vendor"}).Default("customer").Comment("用户类型").End().
		String("old_password", 255).Nullable().Comment("旧密码（已加密）").End().
		Date("birth_date").Nullable().Comment("出生日期").End().
		String("gender", 10).Nullable().Comment("性别").End().
		String("country", 100).Nullable().Comment("国家").End().
		String("city", 100).Nullable().Comment("城市").End().
		Decimal("account_balance", 10, 2).Default(0.00).Comment("账户余额").End().
		Integer("loyalty_points").Default(0).Comment("积分").End().
		DateTime("last_login_old").Nullable().Comment("旧系统最后登录时间").End().
		Boolean("is_active_old").Default(true).Comment("旧系统状态").End().
		Text("preferences_json").Nullable().Comment("用户偏好(JSON格式)").End().
		DateTime("created_at_old").Nullable().Comment("旧系统创建时间").End().
		Timestamps().
		Index("old_username").End().
		Index("old_email").End().
		Index("user_type").End().
		Engine("InnoDB").
		Comment("遗留用户表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 5. 创建新用户表（目标表）
	err = advancedBuilder.Table("new_users").
		ID().
		String("username", 100).NotNull().Unique().Comment("用户名").End().
		String("email", 255).NotNull().Unique().Comment("邮箱").End().
		String("password_hash", 255).NotNull().Comment("密码哈希").End().
		String("first_name", 100).Nullable().Comment("名").End().
		String("last_name", 100).Nullable().Comment("姓").End().
		String("phone", 20).Nullable().Comment("电话").End().
		String("address", 500).Nullable().Comment("地址").End().
		Date("birth_date").Nullable().Comment("生日").End().
		Enum("gender", []string{"male", "female", "other", "prefer_not_to_say"}).Nullable().Comment("性别").End().
		String("country", 100).Nullable().Comment("国家").End().
		String("city", 100).Nullable().Comment("城市").End().
		Enum("role", []string{"customer", "admin", "vendor", "moderator"}).Default("customer").Comment("角色").End().
		Enum("status", []string{"active", "inactive", "suspended", "pending"}).Default("pending").Comment("状态").End().
		Decimal("wallet_balance", 10, 2).Default(0.00).Comment("钱包余额").End().
		Integer("reward_points").Default(0).Comment("奖励积分").End().
		Json("preferences").Nullable().Comment("用户偏好").End().
		Timestamp("email_verified_at").Nullable().Comment("邮箱验证时间").End().
		Timestamp("last_login_at").Nullable().Comment("最后登录时间").End().
		String("legacy_user_id", 100).Nullable().Comment("旧系统用户ID").End().
		Boolean("migrated_from_legacy").Default(false).Comment("是否从旧系统迁移").End().
		Timestamps().
		Index("username").End().
		Index("email").End().
		Index("role").End().
		Index("status").End().
		Index("legacy_user_id").End().
		Engine("InnoDB").
		Comment("新用户表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 6. 执行数据迁移逻辑
	return m.migrateUsersData(ctx, advancedBuilder, db)
}

// migrateUsersData 执行用户数据迁移
func (m *DataMigrationHelperMigration) migrateUsersData(ctx context.Context, advancedBuilder *builder.AdvancedBuilder, db types.DB) error {
	// 首先插入一些测试数据到legacy_users表
	testData := [][]interface{}{
		{"john_doe", "john@example.com", "John Doe", "123-456-7890", "123 Main St", "customer", "hashed_password_1", "1990-01-15", "male", "USA", "New York", 100.50, 500, time.Now().Add(-24 * time.Hour), true, `{"theme":"dark","language":"en"}`, time.Now().Add(-365 * 24 * time.Hour)},
		{"jane_smith", "jane@example.com", "Jane Smith", "098-765-4321", "456 Oak Ave", "admin", "hashed_password_2", "1985-03-22", "female", "USA", "Los Angeles", 250.75, 1200, time.Now().Add(-48 * time.Hour), true, `{"theme":"light","language":"en"}`, time.Now().Add(-200 * 24 * time.Hour)},
		{"bob_wilson", "bob@example.com", "Bob Wilson", "555-123-4567", "789 Pine Rd", "vendor", "hashed_password_3", "1992-07-08", "male", "Canada", "Toronto", 75.25, 300, time.Now().Add(-72 * time.Hour), false, `{"theme":"auto","language":"fr"}`, time.Now().Add(-180 * 24 * time.Hour)},
	}

	columns := []string{"old_username", "old_email", "full_name", "phone_number", "address", "user_type", "old_password", "birth_date", "gender", "country", "city", "account_balance", "loyalty_points", "last_login_old", "is_active_old", "preferences_json", "created_at_old"}

	err := advancedBuilder.BulkInsert(ctx, "legacy_users", columns, testData)
	if err != nil {
		return fmt.Errorf("插入测试数据失败: %v", err)
	}

	// 执行数据转换和迁移
	return m.performUserMigration(ctx, db)
}

// performUserMigration 执行用户迁移逻辑
func (m *DataMigrationHelperMigration) performUserMigration(ctx context.Context, db types.DB) error {
	// 查询所有需要迁移的用户
	query := `
		SELECT id, old_username, old_email, full_name, phone_number, address, 
		       user_type, old_password, birth_date, gender, country, city,
		       account_balance, loyalty_points, is_active_old, preferences_json,
		       created_at_old
		FROM legacy_users 
		WHERE id NOT IN (SELECT COALESCE(source_id, '') FROM migration_mappings 
		                 WHERE entity_type = 'user' AND target_id IS NOT NULL)
	`

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("查询遗留用户失败: %v", err)
	}
	defer rows.Close()

	migrated := 0
	for rows.Next() {
		user := LegacyUser{}
		err := rows.Scan(&user.ID, &user.OldUsername, &user.OldEmail, &user.FullName,
			&user.PhoneNumber, &user.Address, &user.UserType, &user.OldPassword,
			&user.BirthDate, &user.Gender, &user.Country, &user.City,
			&user.AccountBalance, &user.LoyaltyPoints, &user.IsActiveOld,
			&user.PreferencesJSON, &user.CreatedAtOld)
		if err != nil {
			continue
		}

		// 转换用户数据
		newUser := m.transformUser(user)

		// 插入新用户
		newUserID, err := m.insertNewUser(db, newUser)
		if err != nil {
			fmt.Printf("迁移用户失败 %s: %v\n", user.OldUsername, err)
			continue
		}

		// 记录映射关系
		err = m.recordMapping(db, "user", fmt.Sprintf("%d", user.ID), fmt.Sprintf("%d", newUserID))
		if err != nil {
			fmt.Printf("记录映射关系失败: %v\n", err)
		}

		migrated++
	}

	fmt.Printf("成功迁移 %d 个用户\n", migrated)
	return nil
}

// transformUser 转换用户数据
func (m *DataMigrationHelperMigration) transformUser(legacy LegacyUser) NewUser {
	newUser := NewUser{
		Username:           legacy.OldUsername,
		Email:              legacy.OldEmail,
		PasswordHash:       legacy.OldPassword, // 在实际场景中需要重新加密
		Phone:              legacy.PhoneNumber,
		Address:            legacy.Address,
		BirthDate:          legacy.BirthDate,
		Country:            legacy.Country,
		City:               legacy.City,
		WalletBalance:      legacy.AccountBalance,
		RewardPoints:       legacy.LoyaltyPoints,
		LegacyUserID:       fmt.Sprintf("%d", legacy.ID),
		MigratedFromLegacy: true,
	}

	// 分割全名为名和姓
	if legacy.FullName != nil {
		names := strings.SplitN(*legacy.FullName, " ", 2)
		if len(names) > 0 {
			newUser.FirstName = &names[0]
		}
		if len(names) > 1 {
			newUser.LastName = &names[1]
		}
	}

	// 转换性别
	if legacy.Gender != nil {
		switch strings.ToLower(*legacy.Gender) {
		case "m", "male":
			gender := "male"
			newUser.Gender = &gender
		case "f", "female":
			gender := "female"
			newUser.Gender = &gender
		default:
			gender := "other"
			newUser.Gender = &gender
		}
	}

	// 转换角色
	switch legacy.UserType {
	case "customer":
		newUser.Role = "customer"
	case "admin":
		newUser.Role = "admin"
	case "vendor":
		newUser.Role = "vendor"
	default:
		newUser.Role = "customer"
	}

	// 转换状态
	if legacy.IsActiveOld {
		newUser.Status = "active"
	} else {
		newUser.Status = "inactive"
	}

	// 转换偏好设置（从JSON字符串到JSON对象）
	if legacy.PreferencesJSON != nil {
		newUser.Preferences = *legacy.PreferencesJSON
	}

	return newUser
}

// insertNewUser 插入新用户
func (m *DataMigrationHelperMigration) insertNewUser(db types.DB, user NewUser) (int64, error) {
	query := `
		INSERT INTO new_users (username, email, password_hash, first_name, last_name, 
		                      phone, address, birth_date, gender, country, city, role, 
		                      status, wallet_balance, reward_points, preferences, 
		                      legacy_user_id, migrated_from_legacy, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	result, err := db.Exec(query, user.Username, user.Email, user.PasswordHash,
		user.FirstName, user.LastName, user.Phone, user.Address, user.BirthDate,
		user.Gender, user.Country, user.City, user.Role, user.Status,
		user.WalletBalance, user.RewardPoints, user.Preferences,
		user.LegacyUserID, user.MigratedFromLegacy)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// recordMapping 记录ID映射关系
func (m *DataMigrationHelperMigration) recordMapping(db types.DB, entityType, sourceID, targetID string) error {
	query := `
		INSERT INTO migration_mappings (task_id, entity_type, source_id, target_id, created_at, updated_at)
		VALUES (1, ?, ?, ?, NOW(), NOW())
	`
	_, err := db.Exec(query, entityType, sourceID, targetID)
	return err
}

func (m *DataMigrationHelperMigration) Down(ctx context.Context, db types.DB) error {
	tables := []string{
		"new_users",
		"legacy_users",
		"migration_mappings",
		"migration_errors",
		"migration_tasks",
	}

	for _, table := range tables {
		_, err := db.Exec("DROP TABLE IF EXISTS " + table)
		if err != nil {
			return err
		}
	}

	return nil
}

// 数据结构定义

type LegacyUser struct {
	ID              int
	OldUsername     string
	OldEmail        string
	FullName        *string
	PhoneNumber     *string
	Address         *string
	UserType        string
	OldPassword     *string
	BirthDate       *time.Time
	Gender          *string
	Country         *string
	City            *string
	AccountBalance  float64
	LoyaltyPoints   int
	IsActiveOld     bool
	PreferencesJSON *string
	CreatedAtOld    *time.Time
}

type NewUser struct {
	Username           string
	Email              string
	PasswordHash       *string
	FirstName          *string
	LastName           *string
	Phone              *string
	Address            *string
	BirthDate          *time.Time
	Gender             *string
	Country            *string
	City               *string
	Role               string
	Status             string
	WalletBalance      float64
	RewardPoints       int
	Preferences        string
	LegacyUserID       string
	MigratedFromLegacy bool
}
