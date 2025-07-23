# 使用示例

## create-db 命令示例

### 基本用法

从SQL文件创建新数据库：

```bash
# 使用示例SQL文件创建商店数据库
./db-migrator create-db \
  --name "demo_shop" \
  --from-sql "examples/sql_schema/sample_shop.sql"
```

### 高级用法

```bash
# 指定字符集和排序规则
./db-migrator create-db \
  --name "my_shop_utf8" \
  --from-sql "examples/sql_schema/sample_shop.sql" \
  --charset utf8mb4 \
  --collation utf8mb4_unicode_ci

# 如果数据库已存在则跳过
./db-migrator create-db \
  --name "existing_shop" \
  --from-sql "examples/sql_schema/sample_shop.sql" \
  --if-exists skip
```

## 🆕 insert-data 命令示例

### 基本用法

向已存在的数据库插入数据：

```bash
# 向单个数据库插入数据
./db-migrator insert-data \
  --database "demo_shop" \
  --from-sql "examples/sql_data/sample_data.sql"
```

### 多数据库操作

```bash
# 向所有匹配的数据库插入相同数据
./db-migrator insert-data \
  --patterns "shop_*" \
  --from-sql "examples/sql_data/sample_data.sql"

# 向多个指定数据库插入数据
./db-migrator insert-data \
  --databases "shop_001,shop_002,shop_003" \
  --from-sql "examples/sql_data/sample_data.sql"

# 向所有数据库插入数据
./db-migrator insert-data \
  --all \
  --from-sql "examples/sql_data/sample_data.sql"
```

### 高级配置

```bash
# 指定批量大小和验证表存在性
./db-migrator insert-data \
  --database "my_shop" \
  --from-sql "examples/sql_data/sample_data.sql" \
  --batch-size 500 \
  --validate-tables

# 禁用事务（不推荐）
./db-migrator insert-data \
  --database "my_shop" \
  --from-sql "examples/sql_data/sample_data.sql" \
  --use-transaction=false

# 忽略主键冲突（仅用于测试）
./db-migrator insert-data \
  --database "my_shop" \
  --from-sql "examples/sql_data/sample_data.sql" \
  --on-conflict ignore \
  --stop-on-error=false
```

### 完整工作流示例

```bash
# 1. 创建数据库结构
./db-migrator create-db \
  --name "demo_shop" \
  --from-sql "examples/sql_schema/sample_shop.sql"

# 2. 插入基础数据
./db-migrator insert-data \
  --database "demo_shop" \
  --from-sql "examples/sql_data/sample_data.sql"

# 3. 验证数据
mysql -u root -p -e "USE demo_shop; SELECT COUNT(*) FROM products; SELECT COUNT(*) FROM users;"
```

## 配置文件示例

创建 `config.yaml` 文件：

```yaml
database:
  driver: mysql
  host: localhost
  port: 3306
  username: root
  password: your_password
  database: your_default_db
  charset: utf8mb4

migrator:
  migrations_table: schema_migrations
  lock_table: schema_migrations_lock
  auto_backup: false
  dry_run: false
  migrations_dir: migrations
```

## 预期输出

### create-db 成功输出

成功执行后，你会看到类似以下的输出：

```
开始创建数据库...
  数据库名称: demo_shop
  SQL文件: /path/to/examples/sql_schema/sample_shop.sql
  字符集: utf8mb4
  排序规则: utf8mb4_unicode_ci
  已存在处理: error

成功创建数据库: demo_shop (字符集: utf8mb4, 排序规则: utf8mb4_unicode_ci)
开始执行 XX 个DDL语句...
[1/XX] 执行 CREATE_TABLE: users
[2/XX] 执行 CREATE_TABLE: categories
...

🎉 数据库创建完成!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 执行摘要:
  数据库名称: demo_shop
  数据库已创建: true
  SQL语句总数: XX
  成功执行: XX
  执行失败: 0
  执行时间: X.XXXs

📋 创建的数据库对象:
  表 (XX个):
    ✓ users
    ✓ user_addresses
    ✓ categories
    ✓ products
    ✓ orders
    ✓ order_items
    ✓ cart_items
    ✓ settings
  视图 (2个):
    ✓ product_sales_stats
    ✓ user_order_stats
  索引 (2个):
    ✓ idx_orders_user_status_time
    ✓ idx_products_category_status_sort
  触发器 (2个):
    ✓ trg_order_item_stock_decrease
    ✓ trg_order_item_stock_increase
  存储过程 (2个):
    ✓ GetUserCartTotal
    ✓ CleanExpiredCartItems
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### insert-data 成功输出

```
开始执行数据插入...
  SQL文件: /path/to/examples/sql_data/sample_data.sql
  目标数据库 (1个): demo_shop
  批量大小: 1000
  冲突策略: error
  验证表: true
  使用事务: true
  遇错停止: true

✅ 所有表存在性验证通过: [categories products users user_addresses orders order_items cart_items settings]
开始执行 XX 个INSERT语句...
[1/XX] 插入数据到表: categories (7行)
[2/XX] 插入数据到表: products (6行)
[3/XX] 插入数据到表: users (4行)
...
✅ 成功插入 XXX 行数据

🎉 数据插入完成!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 执行摘要:
  数据库名称: demo_shop
  SQL语句总数: XX
  成功执行: XX
  执行失败: 0
  插入总行数: XXX
  执行时间: X.XXXs

📋 各表插入统计:
  categories: XX行 (X个语句)
  products: XX行 (X个语句)
  users: XX行 (X个语句)
  user_addresses: XX行 (X个语句)
  orders: XX行 (X个语句)
  order_items: XX行 (X个语句)
  cart_items: XX行 (X个语句)
  settings: XX行 (X个语句)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## 故障排除

### 常见错误

1. **数据库连接失败**
   ```
   错误: 连接数据库服务器失败: dial tcp 127.0.0.1:3306: connect: connection refused
   ```
   解决：检查MySQL服务是否启动，配置是否正确

2. **数据库已存在 (create-db)**
   ```
   错误: 数据库 'demo_shop' 已存在
   ```
   解决：使用 `--if-exists skip` 参数或选择其他数据库名称

3. **表不存在 (insert-data)**
   ```
   错误: 表存在性验证失败: 表 products 不存在
   ```
   解决：先使用 `create-db` 命令创建表结构，或使用 `--validate-tables=false` 跳过验证

4. **主键冲突 (insert-data)**
   ```
   错误: 执行INSERT语句失败: Error 1062: Duplicate entry '1' for key 'PRIMARY'
   ```
   解决：
   - 使用 `--on-conflict ignore` 忽略冲突
   - 或清空相关表数据
   - 或修改SQL文件中的主键值

5. **SQL解析错误**
   ```
   错误: 第X行解析错误: 无法解析表名
   ```
   解决：检查SQL文件格式，确保语法正确

### 验证结果

插入成功后，可以连接数据库验证：

```bash
mysql -u root -p
```

```sql
USE demo_shop;

-- 检查各表数据量
SELECT 'categories' as table_name, COUNT(*) as row_count FROM categories
UNION ALL
SELECT 'products', COUNT(*) FROM products
UNION ALL
SELECT 'users', COUNT(*) FROM users
UNION ALL
SELECT 'settings', COUNT(*) FROM settings;

-- 查看具体数据
SELECT * FROM categories LIMIT 5;
SELECT * FROM products LIMIT 5;
SELECT * FROM users LIMIT 5;

-- 检查关联数据
SELECT 
    u.username,
    COUNT(o.id) as order_count,
    SUM(o.payment_amount) as total_spent
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.id, u.username;
```

### 性能建议

1. **批量大小调整**
   - 大文件建议使用 `--batch-size 500` 或更小值
   - 小文件可以使用默认的 1000

2. **事务控制**
   - 重要数据建议保持 `--use-transaction=true`（默认）
   - 测试数据可以考虑 `--use-transaction=false` 提高性能

3. **错误处理**
   - 生产环境建议使用 `--stop-on-error=true`（默认）
   - 批量导入可以考虑 `--stop-on-error=false` 继续处理

4. **表验证**
   - 首次导入建议使用 `--validate-tables=true`（默认）
   - 重复导入可以使用 `--validate-tables=false` 节省时间 