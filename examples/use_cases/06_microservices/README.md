# 案例6：微服务架构

## 📋 业务背景

你的团队正在构建一个电商平台，采用微服务架构：
- **用户服务**：用户注册、登录、个人信息管理
- **商品服务**：商品信息、分类、库存管理
- **订单服务**：订单创建、支付、状态跟踪
- **消息服务**：通知、邮件、短信发送
- **分析服务**：数据分析、报表生成

每个服务都有自己独立的数据库，需要独立部署和迁移。

## 🏗️ 系统架构

```
微服务数据库架构
├── user_service_db      # 用户服务数据库
├── product_service_db   # 商品服务数据库
├── order_service_db     # 订单服务数据库
├── message_service_db   # 消息服务数据库
├── analytics_service_db # 分析服务数据库
└── shared_service_db    # 共享服务数据库（配置、字典等）
```

## ⚙️ 配置文件

```yaml
# config.yaml
database:
  driver: mysql
  host: microservices-db.internal
  port: 3306
  username: microservice_admin
  password: ${MICROSERVICE_DB_PASSWORD}
  charset: utf8mb4

databases:
  # 用户服务
  user_service:
    driver: mysql
    host: microservices-db.internal
    port: 3306
    username: microservice_admin
    password: ${MICROSERVICE_DB_PASSWORD}
    database: user_service_db
    charset: utf8mb4
    
  # 商品服务
  product_service:
    driver: mysql
    host: microservices-db.internal
    port: 3306
    username: microservice_admin
    password: ${MICROSERVICE_DB_PASSWORD}
    database: product_service_db
    charset: utf8mb4
    
  # 订单服务
  order_service:
    driver: mysql
    host: microservices-db.internal
    port: 3306
    username: microservice_admin
    password: ${MICROSERVICE_DB_PASSWORD}
    database: order_service_db
    charset: utf8mb4
    
  # 消息服务
  message_service:
    driver: mysql
    host: microservices-db.internal
    port: 3306
    username: microservice_admin
    password: ${MICROSERVICE_DB_PASSWORD}
    database: message_service_db
    charset: utf8mb4
    
  # 分析服务
  analytics_service:
    driver: mysql
    host: microservices-db.internal
    port: 3306
    username: microservice_admin
    password: ${MICROSERVICE_DB_PASSWORD}
    database: analytics_service_db
    charset: utf8mb4
    
  # 共享服务
  shared_service:
    driver: mysql
    host: microservices-db.internal
    port: 3306
    username: microservice_admin
    password: ${MICROSERVICE_DB_PASSWORD}
    database: shared_service_db
    charset: utf8mb4

migrator:
  migrations_table: schema_migrations
  lock_table: schema_migrations_lock
  auto_backup: true
  dry_run: false
  default_database: shared_service
  migrations_dir: migrations
  database_patterns:
    - "*_service"      # 所有服务数据库
    - "user_*"         # 用户相关服务
    - "order_*"        # 订单相关服务
```

## 🗂️ 迁移文件结构

```
migrations/
├── shared_service/            # 共享服务迁移
│   ├── 001_create_system_configs.go
│   ├── 002_create_dictionaries.go
│   └── 003_create_audit_logs.go
├── user_service/             # 用户服务迁移
│   ├── 001_create_users.go
│   ├── 002_create_profiles.go
│   ├── 003_create_auth_tokens.go
│   └── 004_add_oauth_providers.go
├── product_service/          # 商品服务迁移
│   ├── 001_create_categories.go
│   ├── 002_create_products.go
│   ├── 003_create_inventory.go
│   └── 004_add_product_variants.go
├── order_service/            # 订单服务迁移
│   ├── 001_create_orders.go
│   ├── 002_create_order_items.go
│   ├── 003_create_payments.go
│   └── 004_add_shipping_info.go
├── message_service/          # 消息服务迁移
│   ├── 001_create_templates.go
│   ├── 002_create_messages.go
│   └── 003_create_subscriptions.go
├── analytics_service/        # 分析服务迁移
│   ├── 001_create_events.go
│   ├── 002_create_metrics.go
│   └── 003_create_reports.go
└── cross_service/            # 跨服务功能
    ├── 001_add_distributed_locks.go
    └── 002_add_event_sourcing.go
```

## 📝 迁移文件示例

### 用户服务迁移

```go
// migrations/user_service/001_create_users.go
package user_service

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
    return "创建用户表（用户服务）"
}

func (m *CreateUsersTableMigration) Database() string {
    return "user_service"
}

func (m *CreateUsersTableMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "user_service_db")
    builder := builder.NewAdvancedBuilder(checker, db)

    // 创建用户表
    err := builder.Table("users").
        ID().
        String("user_uuid", 36).NotNull().Unique().Comment("用户UUID（全局唯一）").End().
        String("email", 255).NotNull().Unique().Comment("邮箱").End().
        String("username", 50).Nullable().Unique().Comment("用户名").End().
        String("phone", 20).Nullable().Comment("手机号").End().
        String("password_hash", 255).NotNull().Comment("密码哈希").End().
        Enum("status", []string{"active", "inactive", "suspended", "deleted"}).Default("active").End().
        Timestamp("email_verified_at").Nullable().Comment("邮箱验证时间").End().
        Timestamp("phone_verified_at").Nullable().Comment("手机验证时间").End().
        String("avatar_url", 500).Nullable().Comment("头像URL").End().
        Json("metadata").Nullable().Comment("用户元数据").End().
        Timestamp("last_login_at").Nullable().Comment("最后登录时间").End().
        String("last_login_ip", 45).Nullable().Comment("最后登录IP").End().
        Timestamps().
        Index("user_uuid").End().
        Index("email").End().
        Index("username").End().
        Index("phone").End().
        Index("status").End().
        Engine("InnoDB").
        Comment("用户主表").
        Create(ctx)
    if err != nil {
        return err
    }

    // 创建用户登录历史表
    return builder.Table("user_login_history").
        ID().
        String("user_uuid", 36).NotNull().Comment("用户UUID").End().
        String("session_id", 64).NotNull().Comment("会话ID").End().
        String("login_type", 20).NotNull().Comment("登录方式").End().
        String("ip_address", 45).NotNull().Comment("IP地址").End().
        String("user_agent", 1000).Nullable().Comment("用户代理").End().
        String("device_fingerprint", 64).Nullable().Comment("设备指纹").End().
        Enum("status", []string{"success", "failed", "logout"}).NotNull().End().
        String("failure_reason", 200).Nullable().Comment("失败原因").End().
        Timestamp("login_at").NotNull().Comment("登录时间").End().
        Timestamp("logout_at").Nullable().Comment("退出时间").End().
        Json("extra_data").Nullable().Comment("额外数据").End().
        Index("user_uuid").End().
        Index("session_id").End().
        Index("login_at").End().
        Index("status").End().
        Engine("InnoDB").
        Comment("用户登录历史表").
        Create(ctx)
}

func (m *CreateUsersTableMigration) Down(ctx context.Context, db types.DB) error {
    tables := []string{"user_login_history", "users"}
    for _, table := range tables {
        _, err := db.Exec("DROP TABLE IF EXISTS " + table)
        if err != nil {
            return err
        }
    }
    return nil
}
```

### 跨服务功能迁移

```go
// migrations/cross_service/001_add_distributed_locks.go
package cross_service

import (
    "context"
    "db-migrator/internal/builder"
    "db-migrator/internal/checker"
    "db-migrator/internal/types"
)

type AddDistributedLocksMigration struct{}

func (m *AddDistributedLocksMigration) Version() string {
    return "001"
}

func (m *AddDistributedLocksMigration) Description() string {
    return "添加分布式锁表（所有服务数据库）"
}

// 应用到所有服务数据库
func (m *AddDistributedLocksMigration) Databases() []string {
    return []string{
        "user_service", "product_service", "order_service", 
        "message_service", "analytics_service", "shared_service",
    }
}

func (m *AddDistributedLocksMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "")
    builder := builder.NewAdvancedBuilder(checker, db)

    // 创建分布式锁表
    return builder.Table("distributed_locks").
        ID().
        String("lock_key", 255).NotNull().Unique().Comment("锁键名").End().
        String("lock_value", 255).NotNull().Comment("锁值").End().
        String("service_name", 50).NotNull().Comment("服务名称").End().
        String("instance_id", 100).NotNull().Comment("实例ID").End().
        Integer("ttl_seconds").NotNull().Comment("过期时间(秒)").End().
        Timestamp("acquired_at").NotNull().Comment("获取时间").End().
        Timestamp("expires_at").NotNull().Comment("过期时间").End().
        String("metadata", 1000).Nullable().Comment("元数据").End().
        Index("lock_key").End().
        Index("service_name").End().
        Index("expires_at").End().
        Engine("InnoDB").
        Comment("分布式锁表").
        Create(ctx)
}

func (m *AddDistributedLocksMigration) Down(ctx context.Context, db types.DB) error {
    _, err := db.Exec("DROP TABLE IF EXISTS distributed_locks")
    return err
}
```

## 💻 实际操作命令

### 1. 初始化所有微服务

```bash
# 初始化共享服务
db-migrator up -d shared_service

# 初始化所有微服务数据库
db-migrator up --patterns=*_service
```

### 2. 单个服务部署

```bash
# 部署用户服务
db-migrator up -d user_service

# 部署商品服务
db-migrator up -d product_service

# 查看服务状态
db-migrator status -d user_service
```

### 3. 服务组合部署

```bash
# 部署核心服务（用户+商品+订单）
db-migrator up --databases=user_service,product_service,order_service

# 部署支撑服务（消息+分析）
db-migrator up --databases=message_service,analytics_service
```

### 4. 跨服务功能部署

```bash
# 为所有服务添加分布式锁功能
db-migrator up --patterns=*_service --directory=cross_service

# 为特定服务组添加功能
db-migrator up --databases=user_service,order_service --directory=cross_service
```

### 5. 服务依赖管理

```bash
# 按依赖顺序部署
# 1. 先部署基础服务
db-migrator up -d shared_service

# 2. 再部署核心业务服务  
db-migrator up --databases=user_service,product_service

# 3. 最后部署依赖服务
db-migrator up --databases=order_service,message_service,analytics_service
```

## 📊 CI/CD集成示例

### GitHub Actions工作流

```yaml
# .github/workflows/microservice-deploy.yml
name: Microservice Database Migration

on:
  push:
    branches: [main, develop]
    paths: ['migrations/**']

jobs:
  migrate:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [user_service, product_service, order_service, message_service, analytics_service]
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
        
    - name: Install db-migrator
      run: go install github.com/your-org/db-migrator
      
    - name: Run migrations
      env:
        MICROSERVICE_DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
      run: |
        db-migrator up -d ${{ matrix.service }} --config=config-${{ github.ref_name }}.yaml
        
    - name: Verify migration
      env:
        MICROSERVICE_DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
      run: |
        db-migrator status -d ${{ matrix.service }} --config=config-${{ github.ref_name }}.yaml
```

### Docker Compose集成

```yaml
# docker-compose.yml
version: '3.8'

services:
  db-migrator:
    image: your-org/db-migrator:latest
    environment:
      - MICROSERVICE_DB_PASSWORD=${DB_PASSWORD}
    volumes:
      - ./migrations:/app/migrations
      - ./config.yaml:/app/config.yaml
    command: |
      sh -c "
        db-migrator up -d shared_service &&
        db-migrator up --patterns=*_service
      "
    depends_on:
      - mysql

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_USER: microservice_admin
      MYSQL_PASSWORD: ${DB_PASSWORD}
    ports:
      - "3306:3306"
```

## 🔄 服务生命周期管理

### 新服务添加

```bash
# 1. 在配置文件中添加新服务数据库配置
# 2. 创建服务专用的迁移目录
mkdir migrations/payment_service

# 3. 为新服务执行迁移
db-migrator up -d payment_service

# 4. 为新服务添加跨服务功能
db-migrator up -d payment_service --directory=cross_service
```

### 服务下线

```bash
# 1. 备份服务数据
mysqldump payment_service_db > payment_service_backup.sql

# 2. 停止应用程序

# 3. 从配置文件中移除服务配置

# 4. 清理数据库（谨慎操作）
# mysql -e "DROP DATABASE IF EXISTS payment_service_db;"
```

### 服务拆分

```bash
# 例如：将用户服务拆分为用户服务和认证服务

# 1. 创建新的认证服务数据库配置
# 2. 创建认证服务迁移文件
# 3. 将用户服务中的认证相关表迁移到认证服务
db-migrator up -d auth_service

# 4. 更新用户服务，移除认证相关表
db-migrator up -d user_service
```

## 🔧 故障排除

### 问题1：服务依赖迁移失败

```bash
# 检查服务依赖关系
db-migrator status --patterns=*_service

# 按正确的依赖顺序重新执行
db-migrator up -d shared_service
db-migrator up -d user_service
db-migrator up -d order_service  # 依赖用户服务
```

### 问题2：跨服务功能不一致

```bash
# 检查哪些服务缺少跨服务功能
db-migrator status --patterns=*_service | grep -E "(distributed_locks|event_sourcing)"

# 为缺少功能的服务补充
db-migrator up --databases=missing_service1,missing_service2 --directory=cross_service
```

### 问题3：服务间数据一致性

```bash
# 检查关键业务数据的一致性
# 这通常需要应用程序级别的检查脚本
./scripts/check-data-consistency.sh

# 如果发现不一致，可能需要数据修复迁移
db-migrator create fix_data_consistency --directory=maintenance
```

## 🎯 最佳实践

### 1. 服务边界管理
- 每个服务独立的数据库和迁移
- 避免跨服务的直接数据库访问
- 通过API进行服务间通信

### 2. 迁移策略
- 按服务依赖关系排序迁移
- 跨服务功能使用统一的迁移文件
- 关键迁移前进行服务间数据一致性检查

### 3. 版本管理
- 使用语义化版本控制
- 记录服务间兼容性信息
- 维护服务API版本与数据库版本的对应关系

### 4. 监控告警
- 监控每个服务的迁移状态
- 设置跨服务数据一致性检查
- 迁移失败时的自动回滚机制

### 5. 数据备份
- 按服务独立备份
- 关键迁移前的全量备份
- 定期进行跨服务数据恢复演练

这个案例展示了如何使用 `*_service` 模式管理微服务架构的数据库迁移，支持服务的独立部署和跨服务功能！ 