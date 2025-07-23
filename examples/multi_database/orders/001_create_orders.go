package orders

import (
	"context"
	"db-migrator/internal/builder"
	"db-migrator/internal/checker"
	"db-migrator/internal/types"
)

// CreateOrdersTableMigration 创建订单表迁移（订单数据库）
type CreateOrdersTableMigration struct{}

func (m *CreateOrdersTableMigration) Version() string {
	return "001"
}

func (m *CreateOrdersTableMigration) Description() string {
	return "在订单数据库创建订单表"
}

func (m *CreateOrdersTableMigration) Up(ctx context.Context, db types.DB) error {
	checker := checker.NewMySQLChecker(db, "app_orders_db")
	advancedBuilder := builder.NewAdvancedBuilder(checker, db)

	// 创建订单表
	err := advancedBuilder.Table("orders").
		ID().
		String("order_number", 50).NotNull().Unique().Comment("订单号").End().
		String("user_id", 36).NotNull().Comment("用户ID（来自主数据库）").End().
		Enum("status", []string{"pending", "confirmed", "processing", "shipped", "delivered", "cancelled", "refunded"}).
		Default("pending").Comment("订单状态").End().
		Decimal("subtotal", 10, 2).NotNull().Comment("小计").End().
		Decimal("tax_amount", 10, 2).Default(0).Comment("税费").End().
		Decimal("shipping_amount", 10, 2).Default(0).Comment("运费").End().
		Decimal("discount_amount", 10, 2).Default(0).Comment("优惠金额").End().
		Decimal("total_amount", 10, 2).NotNull().Comment("总金额").End().
		String("currency", 3).Default("USD").Comment("货币").End().
		Json("billing_address").NotNull().Comment("账单地址").End().
		Json("shipping_address").NotNull().Comment("配送地址").End().
		String("customer_email", 255).NotNull().Comment("客户邮箱").End().
		String("customer_phone", 20).Nullable().Comment("客户电话").End().
		Text("notes").Nullable().Comment("订单备注").End().
		Json("metadata").Nullable().Comment("订单元数据").End().
		Timestamp("confirmed_at").Nullable().Comment("确认时间").End().
		Timestamp("shipped_at").Nullable().Comment("发货时间").End().
		Timestamp("delivered_at").Nullable().Comment("送达时间").End().
		Timestamps().
		Index("user_id").End().
		Index("status").End().
		Index("order_number").End().
		Index("customer_email").End().
		Index("created_at").End().
		Engine("InnoDB").
		Comment("订单表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 创建订单项表
	err = advancedBuilder.Table("order_items").
		ID().
		Integer("order_id").NotNull().Comment("订单ID").End().
		String("product_id", 36).NotNull().Comment("产品ID").End().
		String("product_sku", 50).NotNull().Comment("产品SKU").End().
		String("product_name", 200).NotNull().Comment("产品名称").End().
		Json("product_snapshot").Nullable().Comment("产品快照").End().
		Integer("quantity").NotNull().Comment("数量").End().
		Decimal("unit_price", 10, 2).NotNull().Comment("单价").End().
		Decimal("total_price", 10, 2).NotNull().Comment("总价").End().
		Json("attributes").Nullable().Comment("商品属性").End().
		Timestamps().
		ForeignKey("order_id").References("orders", "id").OnDelete(builder.ActionCascade).End().
		Index("order_id").End().
		Index("product_id").End().
		Index("product_sku").End().
		Engine("InnoDB").
		Comment("订单项表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 创建订单状态历史表
	return advancedBuilder.Table("order_status_history").
		ID().
		Integer("order_id").NotNull().Comment("订单ID").End().
		String("from_status", 50).Nullable().Comment("原状态").End().
		String("to_status", 50).NotNull().Comment("新状态").End().
		String("changed_by", 36).Nullable().Comment("操作人").End().
		Text("reason").Nullable().Comment("变更原因").End().
		Json("metadata").Nullable().Comment("变更元数据").End().
		Timestamp("changed_at").Default("CURRENT_TIMESTAMP").Comment("变更时间").End().
		ForeignKey("order_id").References("orders", "id").OnDelete(builder.ActionCascade).End().
		Index("order_id").End().
		Index("to_status").End().
		Index("changed_at").End().
		Engine("InnoDB").
		Comment("订单状态历史表").
		Create(ctx)
}

func (m *CreateOrdersTableMigration) Down(ctx context.Context, db types.DB) error {
	// 按依赖关系逆序删除表
	tables := []string{
		"order_status_history",
		"order_items",
		"orders",
	}

	for _, table := range tables {
		_, err := db.Exec("DROP TABLE IF EXISTS " + table)
		if err != nil {
			return err
		}
	}

	return nil
}
