# DB Migrator - 智能数据库迁移工具

一个功能强大、智能化的Go数据库迁移工具，支持MySQL/MariaDB，提供智能存在性检查、多数据库管理和丰富的构建器API。

## 🚀 功能特性

- ✅ **多数据库支持** - MySQL、MariaDB，同一服务器多数据库实例
- ✅ **智能迁移** - 自动检查表/列/索引存在性
- ✅ **Go代码定义** - 使用Go结构体定义迁移，而非SQL文件
- ✅ **链式API** - 流畅的表定义和数据操作API
- ✅ **事务支持** - 确保迁移原子性
- ✅ **并发控制** - 防止同时执行迁移
- ✅ **CLI工具** - 完整的命令行界面 
- ✅ **配置灵活** - YAML配置文件 + 环境变量
- ✅ **模式匹配** - 支持 `shop_*` 等通配符模式
- 🆕 **数据初始化** - 支持JSON/YAML/数据库源的数据初始化
- 🆕 **跨数据库复制** - 智能数据复制，支持字段映射和数据转换
- 🆕 **进度显示** - 实时显示数据操作进度和错误处理

## 🚀 快速开始

### 安装

```bash
go install github.com/xiezhihuan/db-migrator
```

### 初始化项目

```bash
# 创建配置文件和迁移目录
db-migrator init
```

### 配置数据库

```yaml
# config.yaml - 单数据库配置（向后兼容）
database:
  driver: mysql
  host: localhost
  port: 3306
  username: root
  password: secret
  database: myapp_db
  charset: utf8mb4

# 多数据库配置（可选）
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

migrator:
  migrations_table: schema_migrations
  lock_table: schema_migrations_lock
  auto_backup: false
  dry_run: false
  default_database: main
  migrations_dir: migrations
  database_patterns:
    - "app_*"
    - "shop_*"
```

### 创建迁移

```bash
# 为默认数据库创建迁移
db-migrator create create_users_table

# 为指定数据库创建迁移
db-migrator create create_orders_table -d orders
```

### 执行迁移

```bash
# 单数据库迁移
db-migrator up                    # 默认数据库
db-migrator up -d main           # 指定数据库

# 多数据库迁移
db-migrator up --databases=main,users    # 多个数据库
db-migrator up --patterns=shop*          # 模式匹配
db-migrator up --all                     # 所有数据库
```

## 📝 迁移文件示例

### 基础迁移

```go
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
    return "创建用户表"
}

func (m *CreateUsersTableMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "myapp_db")
    builder := builder.NewAdvancedBuilder(checker, db)

    return builder.Table("users").
        ID().
        String("email", 255).NotNull().Unique().Comment("邮箱").End().
        String("password_hash", 255).NotNull().Comment("密码哈希").End().
        String("name", 100).NotNull().Comment("姓名").End().
        Enum("status", []string{"active", "inactive"}).Default("active").End().
        Json("profile").Nullable().Comment("用户档案").End().
        Timestamps().
        Index("email").End().
        Engine("InnoDB").
        Comment("用户表").
        Create(ctx)
}

func (m *CreateUsersTableMigration) Down(ctx context.Context, db types.DB) error {
    _, err := db.Exec("DROP TABLE IF EXISTS users")
    return err
}
```

### 多数据库迁移

```go
package shared

// 创建设置表（应用到多个数据库）
type CreateSettingsTableMigration struct{}

func (m *CreateSettingsTableMigration) Version() string {
    return "001"
}

func (m *CreateSettingsTableMigration) Description() string {
    return "创建系统设置表（多数据库共享）"
}

// 实现MultiDatabaseMigration接口
func (m *CreateSettingsTableMigration) Database() string {
    return "" // 空字符串表示不指定单个数据库
}

func (m *CreateSettingsTableMigration) Databases() []string {
    return []string{"main", "users", "orders"} // 应用到多个数据库
}

func (m *CreateSettingsTableMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "")
    builder := builder.NewAdvancedBuilder(checker, db)

    return builder.Table("system_settings").
        ID().
        String("key", 100).NotNull().Unique().Comment("设置键").End().
        Text("value").Nullable().Comment("设置值").End().
        String("category", 50).Default("general").Comment("分类").End().
        Timestamps().
        Engine("InnoDB").
        Comment("系统设置表").
        Create(ctx)
}

func (m *CreateSettingsTableMigration) Down(ctx context.Context, db types.DB) error {
    _, err := db.Exec("DROP TABLE IF EXISTS system_settings")
    return err
}
```

## 🎯 使用场景

### 微服务架构
```bash
# 为各个微服务数据库分别迁移
db-migrator up -d user_service
db-migrator up -d order_service
db-migrator up -d product_service

# 或批量迁移所有微服务数据库
db-migrator up --patterns=microservice_*
```

### 多租户系统
```bash
# 为所有租户数据库应用相同的迁移
db-migrator up --patterns=tenant_*

# 查看所有租户数据库状态
db-migrator status --patterns=tenant_*
```

### 多环境部署
```bash
# 开发环境
db-migrator up --patterns=dev_*

# 测试环境  
db-migrator up --patterns=test_*

# 生产环境
db-migrator up --patterns=prod_*
```

## 📚 文档

- [**多数据库功能指南**](MULTI_DATABASE_GUIDE.md) - 详细的多数据库使用说明
- [**高级功能文档**](ADVANCED_FEATURES.md) - 高级构建器API和功能
- [**使用场景示例**](examples/usage_scenarios.md) - 各种实际使用场景

## 🎯 具体使用案例

根据你的项目类型，选择合适的案例快速上手：

| 项目类型 | 案例文档 | 匹配模式 | 适用场景 |
|---------|---------|---------|---------|
| 🛒 **连锁商店** | [多店铺系统](examples/use_cases/01_multi_shop_system/) | `shop_*` | 连锁店、加盟店、多分店管理 |
| 🏢 **SaaS平台** | [多租户系统](examples/use_cases/03_saas_multi_tenant/) | `tenant_*` | SaaS产品、多客户独立数据库 |
| 🔧 **微服务** | [微服务架构](examples/use_cases/06_microservices/) | `*_service` | 微服务、服务拆分、独立部署 |
| ⚡ **快速参考** | [快速参考指南](examples/use_cases/QUICK_REFERENCE.md) | - | 命令速查、故障排除、最佳实践 |

### 快速命令示例

```bash
# 连锁店管理 - 为所有店铺执行迁移
db-migrator up --patterns=shop*

# SaaS多租户 - 为所有租户添加新功能  
db-migrator up --patterns=tenant_*

# 微服务架构 - 部署所有服务数据库
db-migrator up --patterns=*_service

# 多环境部署 - 生产环境发布
db-migrator up --patterns=*_prod
```

## 🔧 高级功能

- **外键关系**：自动处理外键约束和级联操作
- **复杂索引**：支持复合索引、唯一索引、全文索引
- **存储过程/函数**：创建和管理数据库函数
- **视图管理**：创建和更新数据库视图
- **触发器支持**：数据变更触发器
- **数据迁移助手**：批量数据处理和ID映射

## 🛡️ 安全特性

- **事务回滚**：失败时自动回滚
- **并发锁**：防止同时迁移冲突
- **存在性检查**：避免重复操作
- **错误处理**：详细的错误信息和堆栈跟踪

## 📊 数据操作功能 🆕

### 数据初始化
为新数据库快速初始化基础数据：

```bash
# 从模板数据库初始化
db-migrator init-data -d new_tenant --from-db=template_db

# 从JSON文件初始化
db-migrator init-data --patterns=shop_* --data-file=base-data.json

# 批量初始化多个数据库
db-migrator init-data --patterns=tenant_* --data-type=system_configs
```

### 跨数据库数据复制
在数据库之间复制数据，支持多种策略：

```bash
# 从总部复制商品到所有店铺
db-migrator copy-data --source=headquarters --patterns=shop_* --tables=products,categories

# 智能合并数据
db-migrator copy-data --source=source_db --target=target_db --strategy=merge --tables=orders

# 条件复制
db-migrator copy-data --source=main --target=archive --conditions="orders:created_at<'2023-01-01'"
```

### 支持的数据源
- **数据库复制** - 从其他数据库复制结构化数据
- **JSON文件** - 从JSON文件导入数据  
- **YAML文件** - 从YAML文件导入数据
- **Go结构体** - 直接在迁移代码中定义数据
- **内置数据** - 预定义的系统基础数据

### 复制策略
- **overwrite** - 完全覆盖（清空后插入）
- **merge** - 智能合并（插入或更新）
- **insert** - 仅插入新数据
- **ignore** - 忽略重复数据

### 进度监控
所有数据操作都支持实时进度显示和错误处理：
- ⏳ 实时进度百分比
- 📊 处理行数统计  
- ❌ 详细错误信息
- 🔄 事务保护
- ⏱️ 超时控制

## 具体使用案例

| 场景 | 案例 | 描述 |
|------|------|------|
| 多店铺系统 | [总部到分店数据复制](examples/data_operations/01_headquarters_to_shops/) | 从总部同步商品目录到各店铺 |
| SaaS平台 | [新租户数据初始化](examples/data_operations/05_new_tenant/) | 为新租户快速初始化基础数据 |
| 微服务架构 | [跨服务数据共享](examples/data_operations/04_cross_service/) | 在微服务间共享基础配置数据 |
| 开发测试 | [测试环境准备](examples/data_operations/08_dev_environment/) | 快速准备开发测试数据 |

👉 **查看完整案例**: [数据操作案例大全](examples/data_operations/)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

---

**DB Migrator** - 让数据库迁移变得简单、安全、高效！ 🚀 