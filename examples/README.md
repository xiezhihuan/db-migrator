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

## 故障排除

### 常见错误

1. **数据库连接失败**
   ```
   错误: 连接数据库服务器失败: dial tcp 127.0.0.1:3306: connect: connection refused
   ```
   解决：检查MySQL服务是否启动，配置是否正确

2. **数据库已存在**
   ```
   错误: 数据库 'demo_shop' 已存在
   ```
   解决：使用 `--if-exists skip` 参数或选择其他数据库名称

3. **SQL解析错误**
   ```
   错误: 第X行解析错误: 无法解析表名
   ```
   解决：检查SQL文件格式，确保语法正确

### 验证结果

创建成功后，可以连接数据库验证：

```bash
mysql -u root -p
```

```sql
USE demo_shop;
SHOW TABLES;
DESCRIBE users;
SELECT * FROM INFORMATION_SCHEMA.VIEWS WHERE TABLE_SCHEMA = 'demo_shop';
``` 