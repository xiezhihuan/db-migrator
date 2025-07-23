package ecommerce

import (
	"context"
	"db-migrator/internal/builder"
	"db-migrator/internal/checker"
	"db-migrator/internal/types"
)

// CreateOrderSystemMigration 创建订单系统迁移
type CreateOrderSystemMigration struct{}

func (m *CreateOrderSystemMigration) Version() string {
	return "002"
}

func (m *CreateOrderSystemMigration) Description() string {
	return "创建电商订单系统 - 购物车、订单、支付、配送"
}

func (m *CreateOrderSystemMigration) Up(ctx context.Context, db types.DB) error {
	checker := checker.NewMySQLChecker(db, "ecommerce_db")
	advancedBuilder := builder.NewAdvancedBuilder(checker, db)

	// 1. 创建用户表（简化版）
	err := advancedBuilder.Table("users").
		ID().
		String("email", 255).NotNull().Unique().Comment("邮箱").End().
		String("password_hash", 255).NotNull().Comment("密码哈希").End().
		String("first_name", 100).Nullable().Comment("名").End().
		String("last_name", 100).Nullable().Comment("姓").End().
		String("phone", 20).Nullable().Comment("电话").End().
		Date("birth_date").Nullable().Comment("生日").End().
		Enum("gender", []string{"male", "female", "other"}).Nullable().Comment("性别").End().
		Enum("status", []string{"active", "inactive", "suspended"}).Default("active").Comment("状态").End().
		Timestamp("email_verified_at").Nullable().Comment("邮箱验证时间").End().
		Timestamp("last_login_at").Nullable().Comment("最后登录时间").End().
		Timestamps().
		Index("email").End().
		Index("status").End().
		Engine("InnoDB").
		Comment("用户表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 2. 创建用户地址表
	err = advancedBuilder.Table("user_addresses").
		ID().
		Integer("user_id").NotNull().Comment("用户ID").End().
		String("type", 20).Default("shipping").Comment("地址类型").End().
		String("first_name", 100).NotNull().Comment("收件人名").End().
		String("last_name", 100).NotNull().Comment("收件人姓").End().
		String("company", 200).Nullable().Comment("公司").End().
		String("address_line_1", 255).NotNull().Comment("地址行1").End().
		String("address_line_2", 255).Nullable().Comment("地址行2").End().
		String("city", 100).NotNull().Comment("城市").End().
		String("state", 100).Nullable().Comment("州/省").End().
		String("postal_code", 20).Nullable().Comment("邮政编码").End().
		String("country", 100).NotNull().Comment("国家").End().
		String("phone", 20).Nullable().Comment("电话").End().
		Boolean("is_default").Default(false).Comment("是否默认地址").End().
		Timestamps().
		ForeignKey("user_id").References("users", "id").OnDelete(builder.ActionCascade).End().
		Index("user_id").End().
		Index("type").End().
		Engine("InnoDB").
		Comment("用户地址表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 3. 创建购物车表
	err = advancedBuilder.Table("carts").
		ID().
		Integer("user_id").Nullable().Comment("用户ID(登录用户)").End().
		String("session_id", 255).Nullable().Comment("会话ID(游客)").End().
		Timestamp("expires_at").Nullable().Comment("过期时间").End().
		Timestamps().
		ForeignKey("user_id").References("users", "id").OnDelete(builder.ActionCascade).End().
		Index("user_id").End().
		Index("session_id").End().
		Index("expires_at").End().
		Engine("InnoDB").
		Comment("购物车表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 4. 创建购物车商品表
	err = advancedBuilder.Table("cart_items").
		ID().
		Integer("cart_id").NotNull().Comment("购物车ID").End().
		Integer("product_id").NotNull().Comment("产品ID").End().
		Integer("variant_id").Nullable().Comment("变体ID").End().
		Integer("quantity").NotNull().Comment("数量").End().
		Decimal("unit_price", 10, 2).NotNull().Comment("单价").End().
		Json("product_snapshot").Nullable().Comment("产品快照").End().
		Timestamps().
		ForeignKey("cart_id").References("carts", "id").OnDelete(builder.ActionCascade).End().
		ForeignKey("product_id").References("products", "id").OnDelete(builder.ActionCascade).End().
		ForeignKey("variant_id").References("product_variants", "id").OnDelete(builder.ActionCascade).End().
		Index("cart_id").End().
		Index("product_id").End().
		Engine("InnoDB").
		Comment("购物车商品表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 5. 创建优惠券表
	err = advancedBuilder.Table("coupons").
		ID().
		String("code", 50).NotNull().Unique().Comment("优惠券代码").End().
		String("name", 200).NotNull().Comment("优惠券名称").End().
		Text("description").Nullable().Comment("描述").End().
		Enum("type", []string{"fixed", "percentage"}).NotNull().Comment("类型").End().
		Decimal("value", 10, 2).NotNull().Comment("优惠值").End().
		Decimal("minimum_amount", 10, 2).Nullable().Comment("最小订单金额").End().
		Decimal("maximum_discount", 10, 2).Nullable().Comment("最大优惠金额").End().
		Integer("usage_limit").Nullable().Comment("使用次数限制").End().
		Integer("used_count").Default(0).Comment("已使用次数").End().
		Boolean("is_active").Default(true).Comment("是否激活").End().
		DateTime("starts_at").Nullable().Comment("开始时间").End().
		DateTime("expires_at").Nullable().Comment("过期时间").End().
		Timestamps().
		Index("code").End().
		Index("is_active").End().
		Index("starts_at").End().
		Index("expires_at").End().
		Engine("InnoDB").
		Comment("优惠券表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 6. 创建订单表
	err = advancedBuilder.Table("orders").
		ID().
		String("order_number", 50).NotNull().Unique().Comment("订单号").End().
		Integer("user_id").Nullable().Comment("用户ID").End().
		Enum("status", []string{"pending", "confirmed", "processing", "shipped", "delivered", "cancelled", "refunded"}).
		Default("pending").Comment("订单状态").End().
		Decimal("subtotal", 10, 2).NotNull().Comment("小计").End().
		Decimal("tax_amount", 10, 2).Default(0).Comment("税费").End().
		Decimal("shipping_amount", 10, 2).Default(0).Comment("运费").End().
		Decimal("discount_amount", 10, 2).Default(0).Comment("优惠金额").End().
		Decimal("total_amount", 10, 2).NotNull().Comment("总金额").End().
		String("currency", 3).Default("USD").Comment("货币").End().
		Integer("coupon_id").Nullable().Comment("优惠券ID").End().
		Json("billing_address").NotNull().Comment("账单地址").End().
		Json("shipping_address").NotNull().Comment("配送地址").End().
		String("customer_email", 255).NotNull().Comment("客户邮箱").End().
		String("customer_phone", 20).Nullable().Comment("客户电话").End().
		Text("notes").Nullable().Comment("订单备注").End().
		Timestamp("confirmed_at").Nullable().Comment("确认时间").End().
		Timestamp("shipped_at").Nullable().Comment("发货时间").End().
		Timestamp("delivered_at").Nullable().Comment("送达时间").End().
		Timestamps().
		ForeignKey("user_id").References("users", "id").OnDelete(builder.ActionSetNull).End().
		ForeignKey("coupon_id").References("coupons", "id").OnDelete(builder.ActionSetNull).End().
		Index("user_id").End().
		Index("status").End().
		Index("order_number").End().
		Index("customer_email").End().
		Engine("InnoDB").
		Comment("订单表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 7. 创建订单商品表
	err = advancedBuilder.Table("order_items").
		ID().
		Integer("order_id").NotNull().Comment("订单ID").End().
		Integer("product_id").NotNull().Comment("产品ID").End().
		Integer("variant_id").Nullable().Comment("变体ID").End().
		String("product_sku", 50).NotNull().Comment("产品SKU").End().
		String("product_name", 200).NotNull().Comment("产品名称").End().
		Json("product_snapshot").Nullable().Comment("产品快照").End().
		Integer("quantity").NotNull().Comment("数量").End().
		Decimal("unit_price", 10, 2).NotNull().Comment("单价").End().
		Decimal("total_price", 10, 2).NotNull().Comment("总价").End().
		Timestamps().
		ForeignKey("order_id").References("orders", "id").OnDelete(builder.ActionCascade).End().
		ForeignKey("product_id").References("products", "id").OnDelete(builder.ActionRestrict).End().
		ForeignKey("variant_id").References("product_variants", "id").OnDelete(builder.ActionRestrict).End().
		Index("order_id").End().
		Index("product_id").End().
		Engine("InnoDB").
		Comment("订单商品表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 8. 创建支付表
	err = advancedBuilder.Table("payments").
		ID().
		Integer("order_id").NotNull().Comment("订单ID").End().
		String("payment_method", 50).NotNull().Comment("支付方式").End().
		String("gateway", 50).NotNull().Comment("支付网关").End().
		String("transaction_id", 255).Nullable().Comment("交易ID").End().
		String("gateway_transaction_id", 255).Nullable().Comment("网关交易ID").End().
		Enum("status", []string{"pending", "processing", "completed", "failed", "cancelled", "refunded"}).
		Default("pending").Comment("支付状态").End().
		Decimal("amount", 10, 2).NotNull().Comment("支付金额").End().
		String("currency", 3).Default("USD").Comment("货币").End().
		Json("gateway_response").Nullable().Comment("网关响应").End().
		Text("failure_reason").Nullable().Comment("失败原因").End().
		Timestamp("processed_at").Nullable().Comment("处理时间").End().
		Timestamps().
		ForeignKey("order_id").References("orders", "id").OnDelete(builder.ActionCascade).End().
		Index("order_id").End().
		Index("status").End().
		Index("payment_method").End().
		Index("transaction_id").End().
		Engine("InnoDB").
		Comment("支付表").
		Create(ctx)
	if err != nil {
		return err
	}

	// 9. 创建配送表
	err = advancedBuilder.Table("shipments").
		ID().
		Integer("order_id").NotNull().Comment("订单ID").End().
		String("tracking_number", 100).Nullable().Comment("快递单号").End().
		String("carrier", 100).Nullable().Comment("承运商").End().
		String("method", 100).NotNull().Comment("配送方式").End().
		Enum("status", []string{"pending", "processing", "shipped", "in_transit", "delivered", "failed"}).
		Default("pending").Comment("配送状态").End().
		Json("shipping_address").NotNull().Comment("配送地址").End().
		Decimal("cost", 10, 2).Default(0).Comment("配送费用").End().
		Timestamp("shipped_at").Nullable().Comment("发货时间").End().
		Timestamp("estimated_delivery").Nullable().Comment("预计送达时间").End().
		Timestamp("delivered_at").Nullable().Comment("实际送达时间").End().
		Text("notes").Nullable().Comment("配送备注").End().
		Timestamps().
		ForeignKey("order_id").References("orders", "id").OnDelete(builder.ActionCascade).End().
		Index("order_id").End().
		Index("tracking_number").End().
		Index("status").End().
		Engine("InnoDB").
		Comment("配送表").
		Create(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m *CreateOrderSystemMigration) Down(ctx context.Context, db types.DB) error {
	tables := []string{
		"shipments",
		"payments",
		"order_items",
		"orders",
		"coupons",
		"cart_items",
		"carts",
		"user_addresses",
		"users",
	}

	for _, table := range tables {
		_, err := db.Exec("DROP TABLE IF EXISTS " + table)
		if err != nil {
			return err
		}
	}

	return nil
}
