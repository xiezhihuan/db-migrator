# 多数据库迁移 - 快速参考指南

## 🚀 一分钟快速上手

### 基本命令模板

```bash
# 单数据库操作
db-migrator <command> -d <database_name>

# 多数据库操作  
db-migrator <command> --databases=<db1,db2,db3>

# 模式匹配操作
db-migrator <command> --patterns=<pattern>

# 全部数据库操作
db-migrator <command> --all
```

## 📋 常用命令速查

| 场景 | 命令 | 说明 |
|------|------|------|
| **单数据库** | `db-migrator up -d main` | 迁移单个数据库 |
| **多数据库** | `db-migrator up --databases=main,users` | 迁移指定多个数据库 |
| **模式匹配** | `db-migrator up --patterns=shop*` | 迁移匹配的数据库 |
| **全部数据库** | `db-migrator up --all` | 迁移所有配置的数据库 |
| **查看状态** | `db-migrator status --patterns=shop*` | 查看匹配数据库状态 |
| **回滚操作** | `db-migrator down --patterns=shop* --steps=1` | 批量回滚 |

## 🎯 业务场景快速匹配

### 1. 连锁店管理
```bash
# 配置模式
database_patterns: ["shop_*"]

# 常用命令
db-migrator up --patterns=shop*              # 所有店铺
db-migrator up --patterns=shop_new_*         # 新店铺
db-migrator status --patterns=shop*          # 查看状态
```

### 2. SaaS多租户
```bash
# 配置模式  
database_patterns: ["tenant_*", "tenant_trial_*"]

# 常用命令
db-migrator up --patterns=tenant_*           # 所有租户
db-migrator up --patterns=tenant_trial_*     # 试用租户
db-migrator up --databases=tenant_001,tenant_002  # 指定租户
```

### 3. 微服务架构
```bash
# 配置模式
database_patterns: ["*_service"]

# 常用命令  
db-migrator up --patterns=*_service          # 所有服务
db-migrator up -d user_service               # 单个服务
db-migrator up --databases=user_service,order_service  # 核心服务
```

### 4. 多环境部署
```bash
# 配置模式
database_patterns: ["*_prod", "*_test", "*_dev"]

# 常用命令
db-migrator up --patterns=*_prod             # 生产环境
db-migrator up --patterns=*_test             # 测试环境
db-migrator up --patterns=*_dev              # 开发环境
```

## ⚙️ 配置文件模板

### 基础多数据库配置
```yaml
# config.yaml
database:
  driver: mysql
  host: localhost
  port: 3306
  username: admin
  password: password
  charset: utf8mb4

databases:
  main:
    database: app_main_db
  users:
    database: app_users_db
  orders:
    database: app_orders_db

migrator:
  default_database: main
  database_patterns:
    - "app_*"
    - "shop_*"
```

### 环境变量配置
```yaml
database:
  host: ${DB_HOST}
  username: ${DB_USER}
  password: ${DB_PASSWORD}

databases:
  prod_main:
    database: ${PROD_MAIN_DB}
  prod_users:
    database: ${PROD_USERS_DB}
```

## 📁 迁移文件组织

### 方式1：按目录组织
```
migrations/
├── main/001_create_users.go      # 指定数据库
├── users/001_create_profiles.go  # 指定数据库
└── shared/001_settings.go        # 多数据库共享
```

### 方式2：代码指定
```go
// 单数据库迁移
func (m *Migration) Database() string {
    return "main"
}

// 多数据库迁移
func (m *Migration) Databases() []string {
    return []string{"main", "users", "orders"}
}
```

## 🔄 实际操作流程

### 新项目初始化
```bash
# 1. 初始化项目
db-migrator init

# 2. 编辑配置文件 config.yaml

# 3. 创建迁移文件
db-migrator create init_tables

# 4. 执行迁移
db-migrator up --all
```

### 日常功能开发
```bash
# 1. 创建新功能迁移
db-migrator create add_new_feature -d target_db

# 2. 编辑迁移文件

# 3. 测试迁移
db-migrator up -d test_db

# 4. 生产发布
db-migrator up --patterns=prod_*
```

### 紧急回滚
```bash
# 1. 查看当前状态
db-migrator status --all

# 2. 回滚所有相关数据库
db-migrator down --patterns=shop* --steps=1

# 3. 验证回滚结果
db-migrator status --patterns=shop*
```

## 🛠️ 故障排除速查

### 常见问题

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| 找不到数据库 | 配置文件中未定义 | 检查 `databases` 配置 |
| 模式匹配失败 | 数据库名不匹配模式 | 检查 `database_patterns` |
| 迁移失败 | 数据库连接问题 | 检查连接参数和权限 |
| 版本冲突 | 多个实例同时迁移 | 检查锁表状态 |

### 调试命令
```bash
# 检查配置
db-migrator status --all

# 测试连接
mysql -h $DB_HOST -u $DB_USER -p -e "SHOW DATABASES"

# 查看锁状态
mysql -e "SELECT * FROM schema_migrations_lock"

# 检查迁移记录
mysql -e "SELECT * FROM schema_migrations ORDER BY applied_at DESC LIMIT 10"
```

## 📊 性能优化建议

### 大量数据库操作
```bash
# 分批处理，避免超时
db-migrator up --databases=shop_001,shop_002,shop_003
db-migrator up --databases=shop_004,shop_005,shop_006

# 并行处理（需要应用程序支持）
db-migrator up --patterns=shop_00[1-3]* &
db-migrator up --patterns=shop_00[4-6]* &
wait
```

### 监控和日志
```bash
# 记录操作日志
db-migrator up --patterns=shop* 2>&1 | tee migration.log

# 监控进度
watch "db-migrator status --patterns=shop* | grep -E '(✅|❌)'"
```

## 🎯 最佳实践清单

### ✅ 推荐做法
- [ ] 使用清晰的数据库命名规范
- [ ] 配置文件使用环境变量
- [ ] 重要操作前备份数据库
- [ ] 分阶段发布（测试→生产）
- [ ] 设置迁移失败告警
- [ ] 定期检查迁移状态

### ❌ 避免做法
- [ ] 在生产环境直接测试新迁移
- [ ] 同时操作过多数据库
- [ ] 忽略迁移失败的错误
- [ ] 硬编码数据库连接信息
- [ ] 跳过备份步骤

## 🚨 应急处理

### 迁移卡住
```bash
# 1. 检查锁状态
mysql -e "SELECT * FROM schema_migrations_lock"

# 2. 手动释放锁（谨慎操作）
mysql -e "DELETE FROM schema_migrations_lock WHERE locked_at < NOW() - INTERVAL 1 HOUR"

# 3. 重新执行迁移
db-migrator up --patterns=affected_pattern
```

### 数据损坏
```bash
# 1. 立即停止所有迁移操作

# 2. 从备份恢复
mysql -e "DROP DATABASE damaged_db"
mysql -e "CREATE DATABASE damaged_db"
mysql damaged_db < backup.sql

# 3. 重新执行迁移
db-migrator up -d damaged_db
```

## 📚 更多资源

- [多数据库功能指南](../MULTI_DATABASE_GUIDE.md)
- [案例1: 多店铺系统](01_multi_shop_system/)
- [案例3: SaaS多租户](03_saas_multi_tenant/)
- [案例6: 微服务架构](06_microservices/)

---

**记住：先在测试环境验证，再在生产环境操作！** 🛡️ 