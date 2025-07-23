# 快速入门指南

本指南将帮助你在5分钟内上手使用db-migrator的核心功能。

## 🚀 前置条件

1. **安装MySQL/MariaDB**
   ```bash
   # macOS
   brew install mysql
   brew services start mysql
   
   # Ubuntu/Debian
   sudo apt update
   sudo apt install mysql-server
   sudo systemctl start mysql
   
   # Windows
   # 下载并安装MySQL官方安装包
   ```

2. **编译db-migrator**
   ```bash
   git clone <repository-url>
   cd db-migrator
   go build -o db-migrator
   ```

## ⚡ 5分钟快速体验

### 步骤1：创建配置文件

```bash
# 创建配置文件
cat > config.yaml << EOF
database:
  driver: mysql
  host: localhost
  port: 3306
  username: root
  password: your_password  # 替换为你的密码
  database: ""
  charset: utf8mb4

migrator:
  migrations_table: schema_migrations
  lock_table: schema_migrations_lock
  auto_backup: false
  dry_run: false
  migrations_dir: migrations
EOF
```

### 步骤2：从SQL文件创建数据库

```bash
# 使用示例SQL文件创建完整的商店数据库
./db-migrator create-db \
  --name "demo_shop" \
  --from-sql "examples/sql_schema/sample_shop.sql"
```

预期输出：
```
开始创建数据库...
成功创建数据库: demo_shop (字符集: utf8mb4, 排序规则: utf8mb4_unicode_ci)
开始执行 XX 个DDL语句...
[1/XX] 执行 CREATE_TABLE: users
[2/XX] 执行 CREATE_TABLE: categories
...
🎉 数据库创建完成!
```

### 步骤3：插入示例数据

```bash
# 向数据库插入示例数据
./db-migrator insert-data \
  --database "demo_shop" \
  --from-sql "examples/sql_data/sample_data.sql"
```

预期输出：
```
开始执行数据插入...
✅ 所有表存在性验证通过
开始执行 XX 个INSERT语句...
[1/XX] 插入数据到表: categories (7行)
[2/XX] 插入数据到表: products (6行)
...
🎉 数据插入完成!
```

### 步骤4：验证结果

```bash
# 连接数据库验证
mysql -u root -p -e "
USE demo_shop;
SELECT 'categories' as table_name, COUNT(*) as row_count FROM categories
UNION ALL SELECT 'products', COUNT(*) FROM products
UNION ALL SELECT 'users', COUNT(*) FROM users;
"
```

## 🎯 实际场景示例

### 场景1：多商店系统

```bash
# 1. 创建总部数据库
./db-migrator create-db \
  --name "headquarters" \
  --from-sql "examples/sql_schema/sample_shop.sql"

# 2. 创建多个分店数据库
for i in {001..003}; do
  ./db-migrator create-db \
    --name "shop_$i" \
    --from-sql "examples/sql_schema/sample_shop.sql"
done

# 3. 为总部插入完整数据
./db-migrator insert-data \
  --database "headquarters" \
  --from-sql "examples/sql_data/sample_data.sql"

# 4. 为所有分店插入基础数据
./db-migrator insert-data \
  --patterns "shop_*" \
  --from-sql "examples/sql_data/base_data.sql"
```

### 场景2：开发环境快速搭建

```bash
# 1. 从生产环境导出结构（模拟）
# mysqldump -u root -p --no-data production_db > schema.sql

# 2. 从SQL文件快速创建开发数据库
./db-migrator create-db \
  --name "dev_database" \
  --from-sql "schema.sql"

# 3. 插入测试数据
./db-migrator insert-data \
  --database "dev_database" \
  --from-sql "test_data.sql" \
  --batch-size 500
```

### 场景3：SaaS多租户

```bash
# 1. 创建租户模板
./db-migrator create-db \
  --name "tenant_template" \
  --from-sql "saas_schema.sql"

# 2. 插入默认配置
./db-migrator insert-data \
  --database "tenant_template" \
  --from-sql "default_settings.sql"

# 3. 为新租户复制结构
for tenant in "tenant_001" "tenant_002" "tenant_003"; do
  ./db-migrator create-db \
    --name "$tenant" \
    --from-sql "saas_schema.sql"
  
  ./db-migrator insert-data \
    --database "$tenant" \
    --from-sql "default_settings.sql"
done
```

## 🔧 高级用法

### 批量处理大文件

```bash
# 处理大数据文件
./db-migrator insert-data \
  --database "large_db" \
  --from-sql "large_dataset.sql" \
  --batch-size 100 \
  --use-transaction=false \
  --stop-on-error=false
```

### 处理数据冲突

```bash
# 忽略重复数据
./db-migrator insert-data \
  --database "existing_db" \
  --from-sql "additional_data.sql" \
  --on-conflict ignore \
  --validate-tables=false
```

### 并行处理多数据库

```bash
# 同时处理多个数据库
./db-migrator insert-data \
  --patterns "prod_*,staging_*,dev_*" \
  --from-sql "update_data.sql" \
  --batch-size 200
```

## 🛠️ 故障排除

### 常见问题

1. **连接失败**
   ```bash
   # 测试数据库连接
   mysql -u root -p -e "SELECT 1"
   ```

2. **权限问题**
   ```bash
   # 确保用户有创建数据库权限
   mysql -u root -p -e "GRANT ALL PRIVILEGES ON *.* TO 'your_user'@'localhost'"
   ```

3. **文件路径错误**
   ```bash
   # 使用绝对路径
   ./db-migrator create-db --name "test" --from-sql "/full/path/to/schema.sql"
   ```

4. **字符编码问题**
   ```bash
   # 确保SQL文件使用UTF-8编码
   file -bi schema.sql
   ```

### 日志调试

```bash
# 启用详细输出
./db-migrator create-db --name "test" --from-sql "schema.sql" --verbose

# 或设置环境变量
DB_MIGRATOR_DEBUG=true ./db-migrator insert-data --database "test" --from-sql "data.sql"
```

## 📚 下一步

1. **阅读完整文档**: [README.md](../README.md)
2. **查看使用示例**: [examples/README.md](README.md)
3. **了解高级功能**: [ADVANCED_FEATURES.md](ADVANCED_FEATURES.md)
4. **多数据库操作**: [MULTI_DATABASE_GUIDE.md](MULTI_DATABASE_GUIDE.md)

## 💡 最佳实践

1. **始终备份生产数据**
2. **在测试环境先验证SQL文件**
3. **大文件使用适当的批量大小**
4. **重要操作保持事务开启**
5. **使用版本控制管理SQL文件** 