# 数据库迁移工具 (DB Migrator)

一个智能的Go语言数据库迁移工具，支持MySQL/MariaDB数据库的版本控制、多数据库操作、数据初始化和跨数据库数据复制。

## ✨ 主要特性

### 🔧 智能迁移系统
- **智能存在性检查** - 自动检测表、列、索引、函数、触发器等数据库对象是否存在
- **自动跳过已存在对象** - 避免重复创建，提高迁移的健壮性
- **事务安全** - 支持事务级别的迁移执行，确保数据一致性
- **版本控制** - 完整的迁移历史记录和版本管理
- **回滚支持** - 支持安全的数据库迁移回滚

### 🌐 多数据库支持
- **批量操作** - 同时对多个数据库执行迁移
- **模式匹配** - 支持通配符匹配数据库名称（如 `shop_*`）
- **灵活配置** - 支持目录结构和代码指定两种迁移组织方式
- **并发控制** - 支持多数据库的并发迁移和锁机制

### 📊 数据操作功能
- **数据初始化** - 支持从JSON、YAML文件或其他数据库初始化数据
- **跨数据库复制** - 支持在不同数据库间复制数据
- **多种策略** - 支持覆盖、合并、插入、忽略等多种数据处理策略
- **进度监控** - 实时显示数据操作进度和错误处理

### 🆕 **SQL文件导入功能** (新增)
- **从SQL文件创建数据库** - 支持完整的DDL语句解析和执行
- **智能解析** - 支持表、视图、存储过程、触发器、索引等多种数据库对象
- **依赖关系处理** - 自动分析和排序SQL语句的执行顺序
- **注释处理** - 正确处理SQL文件中的单行和多行注释
- **字符集配置** - 支持指定数据库字符集和排序规则

### 🆕 **数据插入功能** (新增)
- **从SQL文件插入数据** - 解析INSERT语句并向数据库插入数据
- **表存在性验证** - 插入前自动验证目标表是否存在
- **事务安全** - 支持事务级别的数据插入，失败时自动回滚
- **批量处理** - 支持大文件的分批插入，提高性能
- **冲突处理** - 主键冲突时可选择报错停止或忽略继续
- **多数据库支持** - 支持同时向多个数据库插入相同数据

## 🚀 快速开始

### 安装

```bash
# 克隆项目
git clone <repository-url>
cd db-migrator

# 安装依赖
go mod download

# 构建
go build -o db-migrator
```

### 初始化

```bash
# 初始化项目
./db-migrator init

# 编辑配置文件
vim config.yaml
```

### 基本配置

```yaml
database:
  driver: mysql
  host: localhost
  port: 3306
  username: root
  password: your_password
  database: your_database
  charset: utf8mb4

migrator:
  migrations_table: schema_migrations
  lock_table: schema_migrations_lock
  auto_backup: false
  dry_run: false
  migrations_dir: migrations
```

## 📖 使用指南

### 基础迁移操作

```bash
# 创建迁移文件
./db-migrator create add_users_table

# 执行迁移
./db-migrator up

# 查看状态
./db-migrator status

# 回滚迁移
./db-migrator down --steps=1
```

### **🆕 从SQL文件创建数据库**

这是新增的强大功能，可以从完整的SQL文件创建数据库和所有对象：

```bash
# 基本用法 - 从SQL文件创建数据库
./db-migrator create-db --name "my_new_shop" --from-sql "schema.sql"

# 指定字符集和排序规则
./db-migrator create-db \
  --name "my_shop" \
  --from-sql "schema.sql" \
  --charset utf8mb4 \
  --collation utf8mb4_unicode_ci

# 如果数据库已存在则跳过
./db-migrator create-db \
  --name "my_shop" \
  --from-sql "schema.sql" \
  --if-exists skip

# 处理复杂的SQL文件（包含存储过程、触发器等）
./db-migrator create-db \
  --name "complex_db" \
  --from-sql "examples/sql_schema/sample_shop.sql"
```

### **🆕 向数据库插入数据**

支持从SQL文件向已存在的数据库插入数据：

```bash
# 向单个数据库插入数据
./db-migrator insert-data \
  --database "my_shop" \
  --from-sql "data.sql"

# 向多个数据库插入相同数据
./db-migrator insert-data \
  --patterns "shop_*" \
  --from-sql "base_data.sql"

# 指定批量大小和冲突处理
./db-migrator insert-data \
  --database "my_shop" \
  --from-sql "large_data.sql" \
  --batch-size 500 \
  --on-conflict ignore

# 完整工作流：创建数据库 + 插入数据
./db-migrator create-db --name "demo_shop" --from-sql "schema.sql"
./db-migrator insert-data --database "demo_shop" --from-sql "data.sql"
```

#### SQL文件支持的对象类型

- ✅ **表 (CREATE TABLE)** - 包括外键约束和依赖关系
- ✅ **视图 (CREATE VIEW)** - 自动处理表依赖关系
- ✅ **存储过程 (CREATE PROCEDURE)** - 支持复杂的存储过程定义
- ✅ **函数 (CREATE FUNCTION)** - 支持用户定义函数
- ✅ **触发器 (CREATE TRIGGER)** - 自动处理表依赖关系
- ✅ **索引 (CREATE INDEX)** - 包括唯一索引和复合索引
- ✅ **注释处理** - 正确处理 `--` 和 `/* */` 注释
- ✅ **分隔符处理** - 支持 `DELIMITER` 语句

#### 智能特性

1. **依赖关系自动排序** - 自动分析表间的外键依赖，按正确顺序创建
2. **存在性检查** - 创建前检查数据库是否已存在
3. **事务安全** - 所有DDL操作在事务中执行，失败时自动回滚
4. **详细报告** - 显示创建的对象统计和执行时间

### 多数据库操作

```bash
# 操作单个数据库
./db-migrator up --database main_db

# 操作多个数据库
./db-migrator up --databases main_db,log_db,user_db

# 使用模式匹配
./db-migrator up --patterns shop_*

# 操作所有数据库
./db-migrator up --all
```

### 数据初始化

```bash
# 从模板数据库初始化新租户
./db-migrator init-data \
  --database tenant_new_001 \
  --from-db tenant_template

# 从JSON文件初始化数据
./db-migrator init-data \
  --patterns shop_* \
  --data-file shop-init-data.json

# 为微服务初始化配置数据
./db-migrator init-data \
  --patterns *_service \
  --data-type system_configs
```

### 跨数据库数据复制

```bash
# 从总部复制商品数据到所有店铺
./db-migrator copy-data \
  --source headquarters \
  --patterns shop_* \
  --tables products,categories

# 复制指定条件的数据
./db-migrator copy-data \
  --source main_db \
  --target backup_db \
  --tables orders \
  --conditions "orders:status='completed'"

# 使用配置文件复制
./db-migrator copy-data --config copy-config.json
```

## 🗂️ 项目结构

```
db-migrator/
├── cmd/                    # CLI命令实现
│   ├── root.go            # 根命令和全局配置
│   ├── create_db.go       # 🆕 SQL文件导入命令
│   └── data.go            # 数据操作命令
├── internal/
│   ├── types/             # 类型定义
│   │   ├── migration.go   # 迁移相关类型
│   │   └── database.go    # 🆕 数据库创建相关类型
│   ├── database/          # 数据库操作
│   │   ├── manager.go     # 多数据库管理器
│   │   └── creator.go     # 🆕 数据库创建器
│   ├── sqlparser/         # 🆕 SQL解析器
│   │   └── parser.go      # SQL文件解析实现
│   ├── migrator/          # 迁移器实现
│   ├── builder/           # SQL构建器
│   └── checker/           # 存在性检查器
├── examples/
│   ├── sql_schema/        # 🆕 SQL示例文件
│   │   └── sample_shop.sql # 完整的商店数据库结构
│   ├── use_cases/         # 使用场景示例
│   └── data_operations/   # 数据操作示例
├── migrations/            # 迁移文件目录
├── config.yaml           # 配置文件
└── README.md
```

## 📝 示例SQL文件

项目包含了一个完整的示例SQL文件 `examples/sql_schema/sample_shop.sql`，展示了支持的所有对象类型：

```sql
-- 表结构
CREATE TABLE `users` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `username` varchar(50) NOT NULL,
    -- ... 更多字段
    FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`)
);

-- 视图
CREATE VIEW `product_sales_stats` AS
SELECT p.`id`, p.`name`, SUM(oi.`quantity`) AS `total_sold`
FROM `products` p
LEFT JOIN `order_items` oi ON p.`id` = oi.`product_id`;

-- 存储过程
DELIMITER $$
CREATE PROCEDURE `GetUserCartTotal`(IN p_user_id BIGINT)
BEGIN
    SELECT COUNT(*) AS item_count FROM cart_items WHERE user_id = p_user_id;
END$$
DELIMITER ;

-- 触发器
CREATE TRIGGER `trg_order_item_stock_decrease` 
AFTER INSERT ON `order_items`
FOR EACH ROW
BEGIN
    UPDATE `products` SET `stock` = `stock` - NEW.`quantity`;
END;
```

## 🔧 高级配置

### 多数据库配置

```yaml
databases:
  # 主数据库
  main:
    driver: mysql
    host: localhost
    port: 3306
    username: root
    password: password
    database: main_db
    charset: utf8mb4

  # SaaS多租户配置
  tenant_template:
    driver: mysql
    host: localhost
    port: 3306
    username: root
    password: password
    database: tenant_template
    charset: utf8mb4

migrator:
  # 多数据库设置
  database_patterns:
    - "shop_*"
    - "tenant_*"
    - "*_service"
  
  # 迁移组织方式
  organization_style: "directory" # or "code"
```

## 📊 命令参考

### create-db 命令

```bash
db-migrator create-db [flags]

Flags:
  --name string         数据库名称 (必填)
  --from-sql string     SQL文件路径 (必填)
  --charset string      数据库字符集 (默认: utf8mb4)
  --collation string    数据库排序规则 (默认: utf8mb4_unicode_ci)
  --if-exists string    数据库已存在时的处理方式: error, skip, prompt (默认: error)

Examples:
  # 从SQL文件创建数据库
  db-migrator create-db --name "my_new_shop" --from-sql "schema.sql"
  
  # 指定字符集和排序规则
  db-migrator create-db --name "my_shop" --from-sql "schema.sql" --charset utf8mb4 --collation utf8mb4_unicode_ci
  
  # 如果数据库已存在则跳过
  db-migrator create-db --name "my_shop" --from-sql "schema.sql" --if-exists skip
```

### insert-data 命令

```bash
db-migrator insert-data [flags]

Flags:
  --from-sql string      包含INSERT语句的SQL文件路径 (必填)
  --batch-size int       批量插入大小 (默认: 1000)
  --on-conflict string   主键冲突处理: error, ignore (默认: error)
  --validate-tables      验证表是否存在 (默认: true)
  --use-transaction      使用事务保证一致性 (默认: true)
  --stop-on-error        遇到错误时停止执行 (默认: true)

Examples:
  # 向单个数据库插入数据
  db-migrator insert-data --database "my_shop" --from-sql "data.sql"
  
  # 向多个数据库插入相同数据
  db-migrator insert-data --patterns "shop_*" --from-sql "base_data.sql"
  
  # 指定批量大小和冲突处理
  db-migrator insert-data --database "my_shop" --from-sql "data.sql" --batch-size 500 --on-conflict ignore
  
  # 向所有数据库插入数据
  db-migrator insert-data --all --from-sql "global_data.sql"
```

### 通用数据库选择参数

所有多数据库命令都支持以下参数：

```bash
  -d, --database string     指定目标数据库
      --databases strings   指定多个目标数据库（逗号分隔）
      --patterns strings    数据库名匹配模式（支持通配符）
      --all                 操作所有配置的数据库
```

## 🛠️ 开发指南

### 迁移文件示例

```go
package migrations

import (
    "context"
    "db-migrator/internal/builder"
    "db-migrator/internal/types"
)

type AddUsersTableMigration struct{}

func (m *AddUsersTableMigration) Version() string {
    return "1703123456"
}

func (m *AddUsersTableMigration) Description() string {
    return "添加用户表"
}

func (m *AddUsersTableMigration) Up(ctx context.Context, db types.DB) error {
    builder := builder.NewAdvancedBuilder(nil, db)
    
    return builder.CreateTable("users").
        AddColumn("id", "INT PRIMARY KEY AUTO_INCREMENT").
        AddColumn("username", "VARCHAR(50) NOT NULL UNIQUE").
        AddColumn("email", "VARCHAR(100) NOT NULL UNIQUE").
        AddColumn("created_at", "TIMESTAMP DEFAULT CURRENT_TIMESTAMP").
        Create(ctx)
}

func (m *AddUsersTableMigration) Down(ctx context.Context, db types.DB) error {
    _, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS users")
    return err
}
```

## 🔍 故障排除

### 常见问题

1. **数据库连接失败**
   ```bash
   # 检查配置文件
   cat config.yaml
   
   # 测试连接
   mysql -h localhost -u root -p
   ```

2. **SQL解析错误**
   ```bash
   # 检查SQL文件格式
   # 确保使用正确的字符编码 (UTF-8)
   # 检查分隔符和注释格式
   ```

3. **依赖关系错误**
   ```bash
   # 检查外键引用的表是否存在
   # 确保表创建顺序正确
   ```

### 日志级别

```bash
# 详细输出
./db-migrator create-db --name test --from-sql schema.sql --verbose

# 调试模式
DB_MIGRATOR_DEBUG=true ./db-migrator create-db --name test --from-sql schema.sql
```

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

**🌟 新功能亮点：`create-db` 命令让您可以轻松地从现有的SQL文件快速创建完整的数据库结构！** 