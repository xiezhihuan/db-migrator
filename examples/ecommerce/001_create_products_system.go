package ecommerce

import (
	"context"
	"github.com/xiezhihuan/db-migrator/internal/builder"
	"github.com/xiezhihuan/db-migrator/internal/checker"
	"github.com/xiezhihuan/db-migrator/internal/types"
)

// CreateProductsSystemMigration 创建电商产品系统迁移
type CreateProductsSystemMigration struct{}

func (m *CreateProductsSystemMigration) Version() string {
	return "001"
}

func (m *CreateProductsSystemMigration) Description() string {
	return "创建电商产品系统 - 分类、产品、库存、价格表"
}

func (m *CreateProductsSystemMigration) Up(ctx context.Context, db types.DB) error {
	// 创建高级构建器
	checker := checker.NewMySQLChecker(db, "ecommerce_db") // 替换为实际数据库名
	advancedBuilder := builder.NewAdvancedBuilder(checker, db)

	// 1. 创建产品分类表
	err := advancedBuilder.Table("categories").
		ID().
		String("name", 100).NotNull().Unique().Comment("分类名称").End().
		String("slug", 100).NotNull().Unique().Comment("URL友好名称").End().
		Text("description").Nullable().Comment("分类描述").End().
		String("image_url", 255).Nullable().Comment("分类图片").End().
		Integer("parent_id").Nullable().Comment("父分类ID").End().
		Integer("sort_order").Default(0).Comment("排序").End().
		Boolean("is_active").Default(true).Comment("是否激活").End().
		Timestamps().
		ForeignKey("parent_id").References("categories", "id").OnDelete(builder.ActionSetNull).End().
		Index("parent_id").End().
		Index("sort_order").End().
		Unique("slug").End().
		Engine("InnoDB").
		Comment("产品分类表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 2. 创建品牌表
	err = advancedBuilder.Table("brands").
		ID().
		String("name", 100).NotNull().Unique().Comment("品牌名称").End().
		String("slug", 100).NotNull().Unique().Comment("URL友好名称").End().
		Text("description").Nullable().Comment("品牌描述").End().
		String("logo_url", 255).Nullable().Comment("品牌Logo").End().
		String("website", 255).Nullable().Comment("官方网站").End().
		Boolean("is_active").Default(true).Comment("是否激活").End().
		Timestamps().
		Engine("InnoDB").
		Comment("品牌表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 3. 创建产品表
	err = advancedBuilder.Table("products").
		ID().
		String("sku", 50).NotNull().Unique().Comment("产品SKU").End().
		String("name", 200).NotNull().Comment("产品名称").End().
		String("slug", 200).NotNull().Unique().Comment("URL友好名称").End().
		Text("description").Nullable().Comment("产品描述").End().
		Text("short_description").Nullable().Comment("简短描述").End().
		Integer("category_id").NotNull().Comment("分类ID").End().
		Integer("brand_id").Nullable().Comment("品牌ID").End().
		Decimal("price", 10, 2).NotNull().Comment("价格").End().
		Decimal("compare_price", 10, 2).Nullable().Comment("对比价格").End().
		Decimal("cost_price", 10, 2).Nullable().Comment("成本价格").End().
		Boolean("track_inventory").Default(true).Comment("是否跟踪库存").End().
		Integer("weight").Default(0).Comment("重量(克)").End().
		Enum("status", []string{"draft", "active", "inactive", "archived"}).Default("draft").Comment("状态").End().
		Boolean("is_featured").Default(false).Comment("是否推荐").End().
		Json("meta_data").Nullable().Comment("元数据").End().
		String("seo_title", 200).Nullable().Comment("SEO标题").End().
		Text("seo_description").Nullable().Comment("SEO描述").End().
		Timestamps().
		ForeignKey("category_id").References("categories", "id").OnDelete(builder.ActionRestrict).End().
		ForeignKey("brand_id").References("brands", "id").OnDelete(builder.ActionSetNull).End().
		Index("category_id").End().
		Index("brand_id").End().
		Index("status").End().
		Index("is_featured").End().
		Index("price").End().
		Unique("sku").End().
		Engine("InnoDB").
		Comment("产品表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 4. 创建产品变体表（不同规格的产品）
	err = advancedBuilder.Table("product_variants").
		ID().
		Integer("product_id").NotNull().Comment("产品ID").End().
		String("sku", 50).NotNull().Unique().Comment("变体SKU").End().
		String("title", 200).NotNull().Comment("变体标题").End().
		Decimal("price", 10, 2).Nullable().Comment("变体价格").End().
		Decimal("compare_price", 10, 2).Nullable().Comment("变体对比价格").End().
		Integer("weight").Default(0).Comment("重量(克)").End().
		String("barcode", 50).Nullable().Comment("条形码").End().
		Integer("position").Default(1).Comment("排序位置").End().
		Boolean("is_default").Default(false).Comment("是否默认变体").End().
		Json("attributes").Nullable().Comment("变体属性(颜色、尺寸等)").End().
		Timestamps().
		ForeignKey("product_id").References("products", "id").OnDelete(builder.ActionCascade).End().
		Index("product_id").End().
		Index("position").End().
		Unique("sku").End().
		Engine("InnoDB").
		Comment("产品变体表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 5. 创建库存表
	err = advancedBuilder.Table("inventories").
		ID().
		Integer("product_id").Nullable().Comment("产品ID").End().
		Integer("variant_id").Nullable().Comment("变体ID").End().
		Integer("quantity").Default(0).Comment("库存数量").End().
		Integer("reserved_quantity").Default(0).Comment("预留数量").End().
		Integer("incoming_quantity").Default(0).Comment("待入库数量").End().
		String("location", 100).Nullable().Comment("仓库位置").End().
		Boolean("track_quantity").Default(true).Comment("是否跟踪数量").End().
		Boolean("continue_selling").Default(false).Comment("缺货时是否继续销售").End().
		Timestamps().
		ForeignKey("product_id").References("products", "id").OnDelete(builder.ActionCascade).End().
		ForeignKey("variant_id").References("product_variants", "id").OnDelete(builder.ActionCascade).End().
		Index("product_id").End().
		Index("variant_id").End().
		Index("location").End().
		Engine("InnoDB").
		Comment("库存表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 6. 创建产品图片表
	err = advancedBuilder.Table("product_images").
		ID().
		Integer("product_id").NotNull().Comment("产品ID").End().
		Integer("variant_id").Nullable().Comment("变体ID").End().
		String("url", 500).NotNull().Comment("图片URL").End().
		String("alt_text", 200).Nullable().Comment("替代文本").End().
		Integer("position").Default(1).Comment("排序位置").End().
		Boolean("is_primary").Default(false).Comment("是否主图").End().
		Timestamps().
		ForeignKey("product_id").References("products", "id").OnDelete(builder.ActionCascade).End().
		ForeignKey("variant_id").References("product_variants", "id").OnDelete(builder.ActionCascade).End().
		Index("product_id").End().
		Index("variant_id").End().
		Index("position").End().
		Engine("InnoDB").
		Comment("产品图片表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 7. 创建产品属性表
	err = advancedBuilder.Table("product_attributes").
		ID().
		String("name", 100).NotNull().Comment("属性名称").End().
		String("type", 50).Default("text").Comment("属性类型").End().
		Json("options").Nullable().Comment("属性选项").End().
		Boolean("is_required").Default(false).Comment("是否必填").End().
		Boolean("is_filterable").Default(false).Comment("是否可筛选").End().
		Integer("sort_order").Default(0).Comment("排序").End().
		Timestamps().
		Index("name").End().
		Index("type").End().
		Index("sort_order").End().
		Engine("InnoDB").
		Comment("产品属性表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 8. 创建产品属性值表
	err = advancedBuilder.Table("product_attribute_values").
		ID().
		Integer("product_id").NotNull().Comment("产品ID").End().
		Integer("attribute_id").NotNull().Comment("属性ID").End().
		Text("value").NotNull().Comment("属性值").End().
		Timestamps().
		ForeignKey("product_id").References("products", "id").OnDelete(builder.ActionCascade).End().
		ForeignKey("attribute_id").References("product_attributes", "id").OnDelete(builder.ActionCascade).End().
		Index("product_id").End().
		Index("attribute_id").End().
		Unique("product_id", "attribute_id").End().
		Engine("InnoDB").
		Comment("产品属性值表").
		Create(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m *CreateProductsSystemMigration) Down(ctx context.Context, db types.DB) error {
	// 按依赖关系逆序删除表
	tables := []string{
		"product_attribute_values",
		"product_attributes",
		"product_images",
		"inventories",
		"product_variants",
		"products",
		"brands",
		"categories",
	}

	for _, table := range tables {
		_, err := db.Exec("DROP TABLE IF EXISTS " + table)
		if err != nil {
			return err
		}
	}

	return nil
}
