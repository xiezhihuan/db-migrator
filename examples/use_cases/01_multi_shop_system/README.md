# 案例1：多店铺连锁系统

## 📋 业务背景

你经营一个连锁店品牌，有以下需求：
- **总部数据库**：管理所有店铺信息、供应商、总体报表
- **各店铺数据库**：每个店铺独立的商品、订单、客户数据
- **新店快速上线**：新开店铺需要快速初始化数据库结构
- **统一功能更新**：新功能需要同时部署到所有店铺

## 🏪 系统架构

```
总部系统 (headquarters_db)
├── 店铺管理
├── 供应商管理  
├── 财务汇总
└── 运营报表

店铺系统 (shop_*)
├── shop_001_db (北京旗舰店)
├── shop_002_db (上海分店)  
├── shop_003_db (广州分店)
├── shop_004_db (深圳分店)
└── shop_new_001_db (即将开业)
```

## ⚙️ 配置文件

```yaml
# config.yaml
database:
  driver: mysql
  host: localhost
  port: 3306
  username: shop_admin
  password: secure_password_2024
  charset: utf8mb4

databases:
  # 总部数据库
  headquarters:
    driver: mysql
    host: localhost
    port: 3306
    username: shop_admin
    password: secure_password_2024
    database: headquarters_db
    charset: utf8mb4
    
  # 各店铺数据库
  shop_001:
    driver: mysql
    host: localhost
    port: 3306
    username: shop_admin
    password: secure_password_2024
    database: shop_001_db  # 北京旗舰店
    charset: utf8mb4
    
  shop_002:
    driver: mysql
    host: localhost
    port: 3306
    username: shop_admin
    password: secure_password_2024
    database: shop_002_db  # 上海分店
    charset: utf8mb4
    
  shop_003:
    driver: mysql
    host: localhost
    port: 3306
    username: shop_admin
    password: secure_password_2024
    database: shop_003_db  # 广州分店
    charset: utf8mb4
    
  shop_004:
    driver: mysql
    host: localhost
    port: 3306
    username: shop_admin
    password: secure_password_2024
    database: shop_004_db  # 深圳分店
    charset: utf8mb4
    
  # 新店铺（即将开业）
  shop_new_001:
    driver: mysql
    host: localhost
    port: 3306
    username: shop_admin
    password: secure_password_2024
    database: shop_new_001_db
    charset: utf8mb4

migrator:
  migrations_table: schema_migrations
  lock_table: schema_migrations_lock
  auto_backup: false
  dry_run: false
  default_database: headquarters
  migrations_dir: migrations
  database_patterns:
    - "shop_*"           # 匹配所有店铺数据库
    - "shop_new_*"       # 匹配新店铺数据库
```

## 🗂️ 迁移文件结构

```
migrations/
├── headquarters/              # 总部专用迁移
│   ├── 001_create_shops.go
│   ├── 002_create_suppliers.go
│   └── 003_create_reports.go
├── shop_common/               # 所有店铺通用迁移
│   ├── 001_create_products.go
│   ├── 002_create_orders.go
│   ├── 003_create_customers.go
│   └── 004_create_inventory.go
├── shop_specific/             # 特定店铺迁移
│   └── 001_beijing_special_features.go
└── new_features/              # 新功能更新
    └── 001_add_loyalty_program.go
```

## 📝 迁移文件示例

### 总部数据库迁移

```go
// migrations/headquarters/001_create_shops.go
package headquarters

import (
    "context"
    "db-migrator/internal/builder"
    "db-migrator/internal/checker"
    "db-migrator/internal/types"
)

type CreateShopsTableMigration struct{}

func (m *CreateShopsTableMigration) Version() string {
    return "001"
}

func (m *CreateShopsTableMigration) Description() string {
    return "创建店铺管理表（总部数据库）"
}

// 指定只在总部数据库执行
func (m *CreateShopsTableMigration) Database() string {
    return "headquarters"
}

func (m *CreateShopsTableMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "headquarters_db")
    builder := builder.NewAdvancedBuilder(checker, db)

    return builder.Table("shops").
        ID().
        String("shop_code", 10).NotNull().Unique().Comment("店铺编码").End().
        String("name", 100).NotNull().Comment("店铺名称").End().
        String("address", 200).NotNull().Comment("店铺地址").End().
        String("city", 50).NotNull().Comment("所在城市").End().
        String("province", 50).NotNull().Comment("所在省份").End().
        String("manager_name", 50).NotNull().Comment("店长姓名").End().
        String("manager_phone", 20).NotNull().Comment("店长电话").End().
        String("database_name", 50).NotNull().Comment("对应数据库名").End().
        Enum("status", []string{"active", "inactive", "preparing"}).Default("preparing").End().
        Decimal("area_sqm", 10, 2).Nullable().Comment("营业面积(平方米)").End().
        Date("opening_date").Nullable().Comment("开业日期").End().
        Json("settings").Nullable().Comment("店铺配置").End().
        Timestamps().
        Index("shop_code").End().
        Index("city").End().
        Index("status").End().
        Engine("InnoDB").
        Comment("店铺信息表").
        Create(ctx)
}

func (m *CreateShopsTableMigration) Down(ctx context.Context, db types.DB) error {
    _, err := db.Exec("DROP TABLE IF EXISTS shops")
    return err
}
```

### 店铺通用迁移

```go
// migrations/shop_common/001_create_products.go
package shop_common

import (
    "context"
    "db-migrator/internal/builder"
    "db-migrator/internal/checker"
    "db-migrator/internal/types"
)

type CreateProductsTableMigration struct{}

func (m *CreateProductsTableMigration) Version() string {
    return "001"
}

func (m *CreateProductsTableMigration) Description() string {
    return "创建商品表（所有店铺通用）"
}

// 注意：不实现Database()或Databases()方法
// 通过命令行 --patterns=shop_* 来指定应用到哪些数据库

func (m *CreateProductsTableMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "") // 空数据库名，因为会在多个数据库运行
    builder := builder.NewAdvancedBuilder(checker, db)

    return builder.Table("products").
        ID().
        String("sku", 50).NotNull().Unique().Comment("商品SKU").End().
        String("barcode", 50).Nullable().Comment("条形码").End().
        String("name", 200).NotNull().Comment("商品名称").End().
        Text("description").Nullable().Comment("商品描述").End().
        String("category", 100).NotNull().Comment("商品分类").End().
        String("brand", 100).Nullable().Comment("品牌").End().
        Decimal("cost_price", 10, 2).NotNull().Comment("成本价").End().
        Decimal("sale_price", 10, 2).NotNull().Comment("销售价").End().
        Decimal("discount_price", 10, 2).Nullable().Comment("促销价").End().
        Integer("stock_quantity").Default(0).Comment("库存数量").End().
        Integer("min_stock").Default(10).Comment("最低库存警戒线").End().
        Enum("status", []string{"active", "inactive", "discontinued"}).Default("active").End().
        String("supplier_code", 50).Nullable().Comment("供应商编码").End().
        Json("specifications").Nullable().Comment("商品规格").End().
        String("image_url", 500).Nullable().Comment("商品图片").End().
        Decimal("weight", 8, 2).Nullable().Comment("重量(kg)").End().
        Timestamps().
        Index("sku").End().
        Index("barcode").End().
        Index("category").End().
        Index("brand").End().
        Index("status").End().
        Index("supplier_code").End().
        Engine("InnoDB").
        Comment("商品信息表").
        Create(ctx)
}

func (m *CreateProductsTableMigration) Down(ctx context.Context, db types.DB) error {
    _, err := db.Exec("DROP TABLE IF EXISTS products")
    return err
}
```

### 新功能迁移（多数据库）

```go
// migrations/new_features/001_add_loyalty_program.go
package new_features

import (
    "context"
    "db-migrator/internal/builder"
    "db-migrator/internal/checker"
    "db-migrator/internal/types"
)

type AddLoyaltyProgramMigration struct{}

func (m *AddLoyaltyProgramMigration) Version() string {
    return "001"
}

func (m *AddLoyaltyProgramMigration) Description() string {
    return "添加会员积分系统（所有店铺+总部）"
}

// 实现MultiDatabaseMigration接口 - 应用到所有数据库
func (m *AddLoyaltyProgramMigration) Database() string {
    return ""
}

func (m *AddLoyaltyProgramMigration) Databases() []string {
    return []string{"headquarters", "shop_001", "shop_002", "shop_003", "shop_004"}
}

func (m *AddLoyaltyProgramMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "")
    builder := builder.NewAdvancedBuilder(checker, db)

    // 创建会员积分表
    err := builder.Table("loyalty_points").
        ID().
        String("member_id", 50).NotNull().Comment("会员ID").End().
        String("transaction_type", 20).NotNull().Comment("交易类型").End().
        Integer("points").NotNull().Comment("积分变动").End().
        Integer("balance").NotNull().Comment("积分余额").End().
        String("order_id", 50).Nullable().Comment("关联订单").End().
        Text("description").Nullable().Comment("积分描述").End().
        Timestamp("expired_at").Nullable().Comment("过期时间").End().
        Timestamps().
        Index("member_id").End().
        Index("transaction_type").End().
        Index("expired_at").End().
        Engine("InnoDB").
        Comment("会员积分明细表").
        Create(ctx)
    if err != nil {
        return err
    }

    // 创建会员等级表
    return builder.Table("loyalty_levels").
        ID().
        String("level_name", 50).NotNull().Unique().Comment("等级名称").End().
        Integer("min_points").NotNull().Comment("最低积分要求").End().
        Decimal("discount_rate", 5, 4).Default(0).Comment("折扣率").End().
        Json("benefits").Nullable().Comment("会员权益").End().
        String("level_color", 7).Default("#000000").Comment("等级颜色").End().
        Timestamps().
        Index("min_points").End().
        Engine("InnoDB").
        Comment("会员等级表").
        Create(ctx)
}

func (m *AddLoyaltyProgramMigration) Down(ctx context.Context, db types.DB) error {
    tables := []string{"loyalty_levels", "loyalty_points"}
    for _, table := range tables {
        _, err := db.Exec("DROP TABLE IF EXISTS " + table)
        if err != nil {
            return err
        }
    }
    return nil
}
```

## 💻 实际操作命令

### 1. 初始化所有数据库

```bash
# 初始化项目
db-migrator init

# 为总部数据库执行迁移
db-migrator up -d headquarters

# 为所有店铺数据库执行基础迁移
db-migrator up --patterns=shop_*
```

### 2. 新店开业流程

```bash
# 假设要开新店 shop_005
# 1. 先在配置文件中添加数据库配置

# 2. 为新店执行所有基础迁移
db-migrator up -d shop_005

# 3. 在总部数据库中添加店铺记录
# （这可以通过应用程序或手动SQL完成）
```

### 3. 全店铺功能更新

```bash
# 查看当前所有店铺状态
db-migrator status --patterns=shop_*

# 为所有店铺添加新功能
db-migrator up --patterns=shop_*

# 检查更新结果
db-migrator status --patterns=shop_*
```

### 4. 特定店铺操作

```bash
# 只更新指定店铺
db-migrator up --databases=shop_001,shop_002

# 回滚特定店铺的迁移
db-migrator down -d shop_003 --steps=1

# 查看单个店铺状态
db-migrator status -d shop_001
```

### 5. 新店批量初始化

```bash
# 为所有新店铺执行初始化
db-migrator up --patterns=shop_new_*

# 查看新店铺状态
db-migrator status --patterns=shop_new_*
```

## 📊 日常运维场景

### 场景1：新功能发布
```bash
# 1. 先在测试店铺验证
db-migrator up -d shop_test

# 2. 确认无误后全店铺发布
db-migrator up --patterns=shop_*

# 3. 检查发布结果
db-migrator status --patterns=shop_*
```

### 场景2：紧急回滚
```bash
# 如果某个迁移有问题，快速回滚所有店铺
db-migrator down --patterns=shop_* --steps=1
```

### 场景3：店铺数据迁移
```bash
# 将店铺从一个数据库迁移到另一个数据库
# （需要应用程序配合，这里只处理结构迁移）
db-migrator up -d shop_new_location
```

## 🔧 故障排除

### 问题1：某个店铺迁移失败
```bash
# 查看具体错误
db-migrator status -d shop_002

# 单独重试该店铺
db-migrator up -d shop_002

# 如果还有问题，检查数据库连接和权限
```

### 问题2：部分店铺版本不一致
```bash
# 查看所有店铺状态，找出版本差异
db-migrator status --patterns=shop_*

# 为落后的店铺执行更新
db-migrator up --databases=shop_003,shop_004
```

### 问题3：新店铺无法创建表
```bash
# 检查数据库是否存在
mysql -u shop_admin -p -e "SHOW DATABASES LIKE 'shop_new_%'"

# 检查用户权限
mysql -u shop_admin -p -e "SHOW GRANTS"

# 手动创建数据库（如果需要）
mysql -u root -p -e "CREATE DATABASE shop_new_002_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
```

## 🎯 最佳实践

1. **命名规范**：严格按照 `shop_编号_db` 格式命名数据库
2. **分阶段发布**：先在少数店铺测试，再全面发布
3. **备份策略**：重要迁移前要备份数据库
4. **监控告警**：设置迁移失败的监控告警
5. **文档记录**：每次迁移都要记录变更内容和影响

这个案例展示了如何使用 `shop_*` 模式管理多店铺系统，非常适合连锁店、加盟店等业务场景！ 