package shared

import (
	"context"
	"github.com/xiezhihuan/db-migrator/internal/builder"
	"github.com/xiezhihuan/db-migrator/internal/checker"
	"github.com/xiezhihuan/db-migrator/internal/types"
)

// CreateSettingsTableMigration 创建设置表迁移（多数据库共享）
type CreateSettingsTableMigration struct{}

func (m *CreateSettingsTableMigration) Version() string {
	return "001"
}

func (m *CreateSettingsTableMigration) Description() string {
	return "创建设置表（应用到多个数据库）"
}

// 实现MultiDatabaseMigration接口
func (m *CreateSettingsTableMigration) Database() string {
	// 返回空字符串表示不指定单个数据库
	return ""
}

func (m *CreateSettingsTableMigration) Databases() []string {
	// 指定要应用到的数据库列表
	return []string{"main", "users", "orders"}
}

func (m *CreateSettingsTableMigration) Up(ctx context.Context, db types.DB) error {
	// 注意：这里我们不能硬编码数据库名，因为会在多个数据库中执行
	// 可以通过其他方式获取当前数据库名
	checker := checker.NewMySQLChecker(db, "")
	advancedBuilder := builder.NewAdvancedBuilder(checker, db)

	// 创建系统设置表
	return advancedBuilder.Table("system_settings").
		ID().
		String("key", 100).NotNull().Unique().Comment("设置键").End().
		Text("value").Nullable().Comment("设置值").End().
		String("type", 20).Default("string").Comment("值类型").End().
		String("category", 50).Default("general").Comment("设置分类").End().
		Text("description").Nullable().Comment("设置描述").End().
		Boolean("is_public").Default(false).Comment("是否公开").End().
		Boolean("is_encrypted").Default(false).Comment("是否加密").End().
		Json("metadata").Nullable().Comment("元数据").End().
		Timestamps().
		Index("key").End().
		Index("category").End().
		Index("type").End().
		Engine("InnoDB").
		Comment("系统设置表").
		Create(ctx)
}

func (m *CreateSettingsTableMigration) Down(ctx context.Context, db types.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS system_settings")
	return err
}
