package pattern_example

import (
	"context"
	"db-migrator/internal/builder"
	"db-migrator/internal/checker"
	"db-migrator/internal/types"
)

// ShopDatabasesMigration 演示shop*模式匹配的迁移
// 这个迁移会应用到所有以"shop"开头的数据库
type ShopDatabasesMigration struct{}

func (m *ShopDatabasesMigration) Version() string {
	return "001"
}

func (m *ShopDatabasesMigration) Description() string {
	return "为所有shop开头的数据库创建商品表（模式匹配示例）"
}

// 注意：这个迁移不实现MultiDatabaseMigration接口
// 而是通过命令行参数 --patterns=shop* 来指定要应用的数据库

func (m *ShopDatabasesMigration) Up(ctx context.Context, db types.DB) error {
	// 由于我们不知道具体的数据库名，所以checker使用空字符串
	checker := checker.NewMySQLChecker(db, "")
	advancedBuilder := builder.NewAdvancedBuilder(checker, db)

	// 创建商品表（适用于所有商店数据库）
	err := advancedBuilder.Table("products").
		ID().
		String("sku", 50).NotNull().Unique().Comment("商品SKU").End().
		String("name", 200).NotNull().Comment("商品名称").End().
		Text("description").Nullable().Comment("商品描述").End().
		Decimal("price", 10, 2).NotNull().Comment("价格").End().
		Decimal("cost", 10, 2).Nullable().Comment("成本价").End().
		Integer("stock").Default(0).Comment("库存数量").End().
		String("category", 100).Nullable().Comment("分类").End().
		String("brand", 100).Nullable().Comment("品牌").End().
		Enum("status", []string{"active", "inactive", "draft"}).Default("draft").Comment("状态").End().
		Json("attributes").Nullable().Comment("商品属性").End().
		String("image_url", 500).Nullable().Comment("主图URL").End().
		Decimal("weight", 8, 2).Nullable().Comment("重量(kg)").End().
		Boolean("is_featured").Default(false).Comment("是否推荐").End().
		Integer("sort_order").Default(0).Comment("排序").End().
		Timestamps().
		Index("sku").End().
		Index("name").End().
		Index("category").End().
		Index("brand").End().
		Index("status").End().
		Index("is_featured").End().
		Index("sort_order").End().
		Engine("InnoDB").
		Comment("商品表（shop*数据库通用）").
		Create(ctx)
	if err != nil {
		return err
	}

	// 创建商品分类表
	err = advancedBuilder.Table("product_categories").
		ID().
		String("name", 100).NotNull().Comment("分类名称").End().
		String("slug", 100).NotNull().Unique().Comment("URL友好名称").End().
		Text("description").Nullable().Comment("分类描述").End().
		Integer("parent_id").Nullable().Comment("父分类ID").End().
		String("image_url", 500).Nullable().Comment("分类图片").End().
		Integer("sort_order").Default(0).Comment("排序").End().
		Boolean("is_active").Default(true).Comment("是否激活").End().
		Json("metadata").Nullable().Comment("扩展数据").End().
		Timestamps().
		ForeignKey("parent_id").References("product_categories", "id").OnDelete(builder.ActionSetNull).End().
		Index("parent_id").End().
		Index("slug").End().
		Index("sort_order").End().
		Index("is_active").End().
		Engine("InnoDB").
		Comment("商品分类表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 创建库存记录表
	return advancedBuilder.Table("inventory_logs").
		ID().
		Integer("product_id").NotNull().Comment("商品ID").End().
		Enum("type", []string{"in", "out", "adjust"}).NotNull().Comment("操作类型").End().
		Integer("quantity").NotNull().Comment("数量变化").End().
		Integer("before_stock").NotNull().Comment("操作前库存").End().
		Integer("after_stock").NotNull().Comment("操作后库存").End().
		String("reason", 200).Nullable().Comment("操作原因").End().
		String("operator", 100).Nullable().Comment("操作人").End().
		String("reference_type", 50).Nullable().Comment("关联类型").End().
		String("reference_id", 100).Nullable().Comment("关联ID").End().
		Json("metadata").Nullable().Comment("操作元数据").End().
		Timestamp("created_at").Default("CURRENT_TIMESTAMP").Comment("创建时间").End().
		ForeignKey("product_id").References("products", "id").OnDelete(builder.ActionCascade).End().
		Index("product_id").End().
		Index("type").End().
		Index("created_at").End().
		Index("reference_type", "reference_id").End().
		Engine("InnoDB").
		Comment("库存操作日志表").
		Create(ctx)
}

func (m *ShopDatabasesMigration) Down(ctx context.Context, db types.DB) error {
	// 按依赖关系逆序删除表
	tables := []string{
		"inventory_logs",
		"product_categories",
		"products",
	}

	for _, table := range tables {
		_, err := db.Exec("DROP TABLE IF EXISTS " + table)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
使用示例：

1. 应用到所有shop开头的数据库：
   db-migrator up --patterns=shop*

2. 应用到特定的shop数据库：
   db-migrator up --databases=shop_main,shop_branch1,shop_branch2

3. 查看所有shop数据库的状态：
   db-migrator status --patterns=shop*

4. 回滚所有shop数据库：
   db-migrator down --patterns=shop* --steps=1

这种模式特别适用于：
- 多租户系统（每个租户一个数据库）
- 多分店系统（每个分店一个数据库）
- 微服务架构（每个服务可能有多个数据库实例）
*/
