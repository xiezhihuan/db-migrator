# DB Migrator 多数据库功能指南

## 🎯 功能概述

DB Migrator 现在支持强大的多数据库迁移功能，允许你：

- **同时管理多个数据库**：在一台服务器上管理多个独立的数据库实例
- **模式匹配**：使用通配符（如 `shop*`）批量操作匹配的数据库
- **灵活的迁移组织**：支持按目录和代码两种方式组织迁移
- **向后兼容**：现有的单数据库配置和迁移无需修改

## 📋 配置方式

### 1. 多数据库配置

```yaml
# config.yaml

# 默认数据库配置（向后兼容）
database:
  driver: mysql
  host: localhost
  port: 3306
  username: root
  password: secret
  database: default_db
  charset: utf8mb4

# 多数据库配置
databases:
  main:
    driver: mysql
    host: localhost
    port: 3306
    username: root
    password: secret
    database: app_main_db
    charset: utf8mb4
    
  users:
    driver: mysql
    host: localhost
    port: 3306
    username: root
    password: secret
    database: app_users_db
    charset: utf8mb4
    
  orders:
    driver: mysql
    host: localhost
    port: 3306
    username: root
    password: secret
    database: app_orders_db
    charset: utf8mb4

# 迁移器配置
migrator:
  migrations_table: schema_migrations
  lock_table: schema_migrations_lock
  auto_backup: false
  dry_run: false
  default_database: main              # 默认操作的数据库
  migrations_dir: migrations          # 迁移文件目录
  database_patterns:                  # 数据库匹配模式
    - "app_*"                        # 匹配以app_开头的数据库
    - "shop_*"                       # 匹配以shop_开头的数据库
```

## 🏗️ 迁移文件组织方式

### 方式1：按目录组织

```
migrations/
├── main/                    # 主数据库迁移
│   ├── 001_create_users.go
│   └── 002_add_profiles.go
├── users/                   # 用户数据库迁移
│   ├── 001_create_auth.go
│   └── 002_add_sessions.go
├── orders/                  # 订单数据库迁移
│   ├── 001_create_orders.go
│   └── 002_add_payments.go
└── shared/                  # 共享迁移（多数据库）
    └── 001_create_settings.go
```

### 方式2：代码指定数据库

```go
package migrations

import (
    "context"
    "db-migrator/internal/types"
)

// 单数据库迁移
type CreateUsersTableMigration struct{}

func (m *CreateUsersTableMigration) Database() string {
    return "main"  // 指定目标数据库
}

// 多数据库迁移
type CreateSettingsTableMigration struct{}

func (m *CreateSettingsTableMigration) Databases() []string {
    return []string{"main", "users", "orders"}  // 应用到多个数据库
}
```

## 💻 命令行使用

### 基本命令格式

```bash
db-migrator <command> [database-flags] [other-flags]
```

### 数据库选择参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `--database, -d` | 指定单个数据库 | `-d main` |
| `--databases` | 指定多个数据库 | `--databases=main,users,orders` |
| `--patterns` | 使用模式匹配 | `--patterns=shop*,app_*` |
| `--all` | 操作所有配置的数据库 | `--all` |

### 迁移操作

```bash
# 默认数据库迁移
db-migrator up

# 指定单个数据库
db-migrator up -d main
db-migrator up --database=users

# 指定多个数据库
db-migrator up --databases=main,users,orders

# 模式匹配（所有shop开头的数据库）
db-migrator up --patterns=shop*

# 多个模式
db-migrator up --patterns=shop*,app_*

# 所有配置的数据库
db-migrator up --all
```

### 回滚操作

```bash
# 回滚默认数据库1步
db-migrator down

# 回滚指定数据库3步
db-migrator down -d main --steps=3

# 回滚所有shop数据库1步
db-migrator down --patterns=shop* --steps=1

# 回滚多个数据库2步
db-migrator down --databases=main,users --steps=2
```

### 状态查看

```bash
# 查看默认数据库状态
db-migrator status

# 查看指定数据库状态
db-migrator status -d main

# 查看所有数据库状态
db-migrator status --all

# 查看匹配模式的数据库状态
db-migrator status --patterns=shop*
```

### 创建迁移文件

```bash
# 为默认数据库创建迁移
db-migrator create add_user_profile

# 为指定数据库创建迁移
db-migrator create add_order_items -d orders
```

## 📝 迁移文件示例

### 1. 单数据库迁移（目录组织）

```go
// migrations/main/001_create_users.go
package main

import (
    "context"
    "db-migrator/internal/builder"
    "db-migrator/internal/checker"
    "db-migrator/internal/types"
)

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

    return advancedBuilder.Table("users").
        ID().
        String("email", 255).NotNull().Unique().End().
        String("password_hash", 255).NotNull().End().
        Enum("status", []string{"active", "inactive"}).Default("active").End().
        Timestamps().
        Engine("InnoDB").
        Create(ctx)
}

func (m *CreateUsersTableMigration) Down(ctx context.Context, db types.DB) error {
    _, err := db.Exec("DROP TABLE IF EXISTS users")
    return err
}
```

### 2. 多数据库迁移（代码指定）

```go
// migrations/shared/001_create_settings.go
package shared

import (
    "context"
    "db-migrator/internal/builder"
    "db-migrator/internal/checker"
    "db-migrator/internal/types"
)

type CreateSettingsTableMigration struct{}

func (m *CreateSettingsTableMigration) Version() string {
    return "001"
}

func (m *CreateSettingsTableMigration) Description() string {
    return "创建设置表（应用到多个数据库）"
}

// 实现MultiDatabaseMigration接口
func (m *CreateSettingsTableMigration) Database() string {
    return "" // 空字符串表示不指定单个数据库
}

func (m *CreateSettingsTableMigration) Databases() []string {
    return []string{"main", "users", "orders"} // 指定多个数据库
}

func (m *CreateSettingsTableMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "")
    advancedBuilder := builder.NewAdvancedBuilder(checker, db)

    return advancedBuilder.Table("system_settings").
        ID().
        String("key", 100).NotNull().Unique().End().
        Text("value").Nullable().End().
        String("category", 50).Default("general").End().
        Timestamps().
        Engine("InnoDB").
        Create(ctx)
}

func (m *CreateSettingsTableMigration) Down(ctx context.Context, db types.DB) error {
    _, err := db.Exec("DROP TABLE IF EXISTS system_settings")
    return err
}
```

### 3. 模式匹配迁移

```go
// migrations/pattern_example/001_shop_products.go
package pattern_example

// 这个迁移通过命令行 --patterns=shop* 来应用到所有shop开头的数据库
type ShopProductsMigration struct{}

func (m *ShopProductsMigration) Version() string {
    return "001"
}

func (m *ShopProductsMigration) Description() string {
    return "为所有shop数据库创建商品表"
}

func (m *ShopProductsMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "") // 空字符串，因为会在多个数据库中运行
    advancedBuilder := builder.NewAdvancedBuilder(checker, db)

    return advancedBuilder.Table("products").
        ID().
        String("sku", 50).NotNull().Unique().End().
        String("name", 200).NotNull().End().
        Decimal("price", 10, 2).NotNull().End().
        Integer("stock").Default(0).End().
        Timestamps().
        Engine("InnoDB").
        Create(ctx)
}

func (m *ShopProductsMigration) Down(ctx context.Context, db types.DB) error {
    _, err := db.Exec("DROP TABLE IF EXISTS products")
    return err
}
```

## 🔍 实际使用场景

### 1. 微服务架构

```bash
# 配置文件中定义各服务数据库
databases:
  user_service:
    database: microservice_users_db
  order_service:
    database: microservice_orders_db
  product_service:
    database: microservice_products_db

# 分别迁移各服务数据库
db-migrator up -d user_service
db-migrator up -d order_service
db-migrator up -d product_service

# 或者一次性迁移所有微服务数据库
db-migrator up --patterns=microservice_*
```

### 2. 多租户系统

```bash
# 为所有租户数据库应用相同的迁移
db-migrator up --patterns=tenant_*

# 查看所有租户数据库状态
db-migrator status --patterns=tenant_*

# 为特定租户创建专门的迁移
db-migrator create add_tenant_feature -d tenant_001
```

### 3. 多环境部署

```bash
# 开发环境
db-migrator up --patterns=dev_*

# 测试环境
db-migrator up --patterns=test_*

# 生产环境（谨慎操作）
db-migrator up --patterns=prod_*
```

### 4. 分库分表

```bash
# 为所有分片数据库应用迁移
db-migrator up --patterns=shard_*

# 查看分片数据库状态
db-migrator status --patterns=shard_*
```

## ⚠️ 注意事项

### 1. 数据库连接管理
- 工具会自动管理多个数据库连接
- 连接会被缓存以提高性能
- 操作完成后会自动关闭所有连接

### 2. 事务处理
- 每个数据库的迁移都在独立的事务中执行
- 如果某个数据库迁移失败，不会影响其他数据库
- 工具会报告所有失败的数据库和错误信息

### 3. 并发安全
- 每个数据库都有独立的迁移锁
- 可以同时对不同数据库进行迁移
- 相同数据库的并发迁移会被阻止

### 4. 性能考虑
- 大量数据库的批量操作可能耗时较长
- 建议先在小范围测试模式匹配规则
- 可以使用 `--dry-run` 模式预览操作

### 5. 错误处理
- 部分数据库迁移失败不会中断整个过程
- 详细的错误信息会显示具体的数据库和失败原因
- 建议定期检查迁移状态

## 🚀 最佳实践

1. **配置管理**：使用环境变量管理敏感信息
2. **迁移测试**：先在开发环境测试多数据库迁移
3. **模式规范**：建立清晰的数据库命名规范
4. **状态监控**：定期检查所有数据库的迁移状态
5. **备份策略**：重要操作前进行数据备份
6. **文档记录**：记录迁移的业务逻辑和依赖关系

通过这些强大的多数据库功能，DB Migrator 可以轻松应对各种复杂的数据库管理场景！🎉 