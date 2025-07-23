# 案例3：SaaS多租户系统

## 📋 业务背景

你开发了一个SaaS项目管理平台，采用每个租户独立数据库的架构：
- **平台管理数据库**：管理租户信息、订阅计划、计费等
- **租户数据库**：每个客户公司独立的项目、任务、用户数据
- **快速开通**：新客户注册后自动创建专属数据库
- **差异化功能**：不同套餐的客户有不同的功能模块

## 🏢 系统架构

```
平台管理 (platform_db)
├── 租户管理
├── 订阅计划
├── 计费账单
└── 系统监控

租户数据库 (tenant_*)
├── tenant_001_db (ABC科技公司 - 企业版)
├── tenant_002_db (XYZ工作室 - 标准版)
├── tenant_003_db (某某集团 - 企业版)
├── tenant_trial_001_db (试用客户)
└── tenant_new_005_db (即将激活)
```

## ⚙️ 配置文件

```yaml
# config.yaml
database:
  driver: mysql
  host: saas-db.company.com
  port: 3306
  username: saas_admin
  password: ${DB_PASSWORD}  # 使用环境变量
  charset: utf8mb4

databases:
  # 平台管理数据库
  platform:
    driver: mysql
    host: saas-db.company.com
    port: 3306
    username: saas_admin
    password: ${DB_PASSWORD}
    database: platform_db
    charset: utf8mb4
    
  # 企业版租户
  tenant_001:
    driver: mysql
    host: saas-db.company.com
    port: 3306
    username: saas_admin
    password: ${DB_PASSWORD}
    database: tenant_001_db  # ABC科技公司
    charset: utf8mb4
    
  tenant_003:
    driver: mysql
    host: saas-db.company.com
    port: 3306
    username: saas_admin
    password: ${DB_PASSWORD}
    database: tenant_003_db  # 某某集团
    charset: utf8mb4
    
  # 标准版租户
  tenant_002:
    driver: mysql
    host: saas-db.company.com
    port: 3306
    username: saas_admin
    password: ${DB_PASSWORD}
    database: tenant_002_db  # XYZ工作室
    charset: utf8mb4
    
  # 试用租户
  tenant_trial_001:
    driver: mysql
    host: saas-db.company.com
    port: 3306
    username: saas_admin
    password: ${DB_PASSWORD}
    database: tenant_trial_001_db
    charset: utf8mb4
    
  # 待激活租户
  tenant_new_005:
    driver: mysql
    host: saas-db.company.com
    port: 3306
    username: saas_admin
    password: ${DB_PASSWORD}
    database: tenant_new_005_db
    charset: utf8mb4

migrator:
  migrations_table: schema_migrations
  lock_table: schema_migrations_lock
  auto_backup: true  # SaaS系统建议开启备份
  dry_run: false
  default_database: platform
  migrations_dir: migrations
  database_patterns:
    - "tenant_*"        # 所有租户数据库
    - "tenant_trial_*"  # 试用租户
    - "tenant_new_*"    # 新注册租户
```

## 🗂️ 迁移文件结构

```
migrations/
├── platform/                  # 平台管理数据库
│   ├── 001_create_tenants.go
│   ├── 002_create_plans.go
│   ├── 003_create_billing.go
│   └── 004_create_analytics.go
├── tenant_base/               # 所有租户基础功能
│   ├── 001_create_users.go
│   ├── 002_create_projects.go
│   ├── 003_create_tasks.go
│   └── 004_create_files.go
├── tenant_standard/           # 标准版功能
│   ├── 001_add_time_tracking.go
│   └── 002_add_basic_reports.go
├── tenant_enterprise/         # 企业版功能
│   ├── 001_add_advanced_reports.go
│   ├── 002_add_custom_fields.go
│   └── 003_add_api_access.go
└── maintenance/               # 系统维护
    ├── 001_optimize_indexes.go
    └── 002_cleanup_old_data.go
```

## 📝 迁移文件示例

### 平台管理数据库迁移

```go
// migrations/platform/001_create_tenants.go
package platform

import (
    "context"
    "db-migrator/internal/builder"
    "db-migrator/internal/checker"
    "db-migrator/internal/types"
)

type CreateTenantsTableMigration struct{}

func (m *CreateTenantsTableMigration) Version() string {
    return "001"
}

func (m *CreateTenantsTableMigration) Description() string {
    return "创建租户管理表（平台数据库）"
}

func (m *CreateTenantsTableMigration) Database() string {
    return "platform"
}

func (m *CreateTenantsTableMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "platform_db")
    builder := builder.NewAdvancedBuilder(checker, db)

    // 创建租户表
    err := builder.Table("tenants").
        ID().
        String("tenant_code", 20).NotNull().Unique().Comment("租户编码").End().
        String("company_name", 200).NotNull().Comment("公司名称").End().
        String("contact_name", 100).NotNull().Comment("联系人姓名").End().
        String("contact_email", 150).NotNull().Comment("联系人邮箱").End().
        String("contact_phone", 20).Nullable().Comment("联系人电话").End().
        String("database_name", 50).NotNull().Comment("数据库名称").End().
        Enum("plan_type", []string{"trial", "standard", "enterprise"}).Default("trial").Comment("套餐类型").End().
        Enum("status", []string{"active", "suspended", "cancelled", "pending"}).Default("pending").Comment("状态").End().
        Integer("max_users").Default(5).Comment("最大用户数").End().
        Integer("max_projects").Default(10).Comment("最大项目数").End().
        BigInteger("storage_limit_mb").Default(1024).Comment("存储限制(MB)").End().
        Date("trial_ends_at").Nullable().Comment("试用结束日期").End().
        Date("subscription_ends_at").Nullable().Comment("订阅结束日期").End().
        Json("features").Nullable().Comment("可用功能列表").End().
        Json("settings").Nullable().Comment("租户设置").End().
        Timestamps().
        Index("tenant_code").End().
        Index("plan_type").End().
        Index("status").End().
        Index("contact_email").End().
        Engine("InnoDB").
        Comment("租户信息表").
        Create(ctx)
    if err != nil {
        return err
    }

    // 创建租户使用统计表
    return builder.Table("tenant_usage_stats").
        ID().
        String("tenant_code", 20).NotNull().Comment("租户编码").End().
        Date("stat_date").NotNull().Comment("统计日期").End().
        Integer("active_users").Default(0).Comment("活跃用户数").End().
        Integer("total_projects").Default(0).Comment("项目总数").End().
        Integer("total_tasks").Default(0).Comment("任务总数").End().
        BigInteger("storage_used_mb").Default(0).Comment("已使用存储(MB)").End().
        Integer("api_calls").Default(0).Comment("API调用次数").End().
        Timestamp("last_activity_at").Nullable().Comment("最后活动时间").End().
        Timestamps().
        ForeignKey("tenant_code").References("tenants", "tenant_code").OnDelete(builder.ActionCascade).End().
        Index("tenant_code", "stat_date").End().
        Index("stat_date").End().
        Engine("InnoDB").
        Comment("租户使用统计表").
        Create(ctx)
}

func (m *CreateTenantsTableMigration) Down(ctx context.Context, db types.DB) error {
    tables := []string{"tenant_usage_stats", "tenants"}
    for _, table := range tables {
        _, err := db.Exec("DROP TABLE IF EXISTS " + table)
        if err != nil {
            return err
        }
    }
    return nil
}
```

### 租户基础功能迁移

```go
// migrations/tenant_base/001_create_users.go
package tenant_base

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
    return "创建用户表（所有租户基础功能）"
}

// 通过命令行 --patterns=tenant_* 来应用到所有租户数据库

func (m *CreateUsersTableMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "")
    builder := builder.NewAdvancedBuilder(checker, db)

    // 创建用户表
    err := builder.Table("users").
        ID().
        String("email", 150).NotNull().Unique().Comment("邮箱").End().
        String("password_hash", 255).NotNull().Comment("密码哈希").End().
        String("first_name", 50).NotNull().Comment("名").End().
        String("last_name", 50).NotNull().Comment("姓").End().
        String("avatar_url", 500).Nullable().Comment("头像URL").End().
        String("phone", 20).Nullable().Comment("电话").End().
        String("department", 100).Nullable().Comment("部门").End().
        String("position", 100).Nullable().Comment("职位").End().
        Enum("role", []string{"admin", "manager", "member", "viewer"}).Default("member").Comment("角色").End().
        Enum("status", []string{"active", "inactive", "invited"}).Default("invited").Comment("状态").End().
        Json("permissions").Nullable().Comment("权限配置").End().
        Json("preferences").Nullable().Comment("个人偏好").End().
        String("timezone", 50).Default("UTC").Comment("时区").End().
        String("language", 10).Default("en").Comment("语言").End().
        Timestamp("last_login_at").Nullable().Comment("最后登录时间").End().
        String("last_login_ip", 45).Nullable().Comment("最后登录IP").End().
        Timestamp("email_verified_at").Nullable().Comment("邮箱验证时间").End().
        String("invitation_token", 100).Nullable().Comment("邀请令牌").End().
        Timestamp("invitation_sent_at").Nullable().Comment("邀请发送时间").End().
        Timestamps().
        Index("email").End().
        Index("role").End().
        Index("status").End().
        Index("invitation_token").End().
        Engine("InnoDB").
        Comment("用户表").
        Create(ctx)
    if err != nil {
        return err
    }

    // 创建用户会话表
    return builder.Table("user_sessions").
        ID().
        Integer("user_id").NotNull().Comment("用户ID").End().
        String("session_token", 255).NotNull().Unique().Comment("会话令牌").End().
        String("device_info", 500).Nullable().Comment("设备信息").End().
        String("ip_address", 45).Nullable().Comment("IP地址").End().
        String("user_agent", 1000).Nullable().Comment("用户代理").End().
        Timestamp("last_activity_at").NotNull().Comment("最后活动时间").End().
        Timestamp("expires_at").NotNull().Comment("过期时间").End().
        Timestamps().
        ForeignKey("user_id").References("users", "id").OnDelete(builder.ActionCascade).End().
        Index("user_id").End().
        Index("session_token").End().
        Index("expires_at").End().
        Engine("InnoDB").
        Comment("用户会话表").
        Create(ctx)
}

func (m *CreateUsersTableMigration) Down(ctx context.Context, db types.DB) error {
    tables := []string{"user_sessions", "users"}
    for _, table := range tables {
        _, err := db.Exec("DROP TABLE IF EXISTS " + table)
        if err != nil {
            return err
        }
    }
    return nil
}
```

### 企业版功能迁移

```go
// migrations/tenant_enterprise/001_add_advanced_reports.go
package tenant_enterprise

import (
    "context"
    "db-migrator/internal/builder"
    "db-migrator/internal/checker"
    "db-migrator/internal/types"
)

type AddAdvancedReportsMigration struct{}

func (m *AddAdvancedReportsMigration) Version() string {
    return "001"
}

func (m *AddAdvancedReportsMigration) Description() string {
    return "添加高级报表功能（企业版专用）"
}

// 只应用到企业版租户
func (m *AddAdvancedReportsMigration) Databases() []string {
    return []string{"tenant_001", "tenant_003"} // 企业版租户
}

func (m *AddAdvancedReportsMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "")
    builder := builder.NewAdvancedBuilder(checker, db)

    // 创建自定义报表模板表
    err := builder.Table("report_templates").
        ID().
        String("name", 200).NotNull().Comment("报表名称").End().
        Text("description").Nullable().Comment("报表描述").End().
        String("report_type", 50).NotNull().Comment("报表类型").End().
        Json("config").NotNull().Comment("报表配置").End().
        Json("filters").Nullable().Comment("默认筛选条件").End().
        Json("columns").NotNull().Comment("显示列配置").End().
        String("chart_type", 50).Nullable().Comment("图表类型").End().
        Integer("created_by").NotNull().Comment("创建人").End().
        Boolean("is_public").Default(false).Comment("是否公开").End().
        Boolean("is_system").Default(false).Comment("是否系统模板").End().
        Timestamps().
        ForeignKey("created_by").References("users", "id").OnDelete(builder.ActionCascade).End().
        Index("report_type").End().
        Index("created_by").End().
        Index("is_public").End().
        Engine("InnoDB").
        Comment("报表模板表").
        Create(ctx)
    if err != nil {
        return err
    }

    // 创建报表生成历史表
    return builder.Table("report_generations").
        ID().
        Integer("template_id").NotNull().Comment("模板ID").End().
        Integer("generated_by").NotNull().Comment("生成人").End().
        Json("parameters").Nullable().Comment("生成参数").End().
        String("file_path", 500).Nullable().Comment("文件路径").End().
        String("file_format", 20).NotNull().Comment("文件格式").End().
        BigInteger("file_size").Nullable().Comment("文件大小").End().
        Enum("status", []string{"pending", "generating", "completed", "failed"}).Default("pending").End().
        Text("error_message").Nullable().Comment("错误信息").End().
        Timestamp("started_at").Nullable().Comment("开始时间").End().
        Timestamp("completed_at").Nullable().Comment("完成时间").End().
        Timestamps().
        ForeignKey("template_id").References("report_templates", "id").OnDelete(builder.ActionCascade).End().
        ForeignKey("generated_by").References("users", "id").OnDelete(builder.ActionCascade).End().
        Index("template_id").End().
        Index("generated_by").End().
        Index("status").End().
        Index("created_at").End().
        Engine("InnoDB").
        Comment("报表生成历史表").
        Create(ctx)
}

func (m *AddAdvancedReportsMigration) Down(ctx context.Context, db types.DB) error {
    tables := []string{"report_generations", "report_templates"}
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

### 1. 系统初始化

```bash
# 初始化平台管理数据库
db-migrator up -d platform

# 为所有现有租户初始化基础功能
db-migrator up --patterns=tenant_* --directory=tenant_base
```

### 2. 新租户开通流程

```bash
# 1. 应用程序自动创建租户数据库配置

# 2. 为新租户执行基础迁移
db-migrator up -d tenant_new_005 --directory=tenant_base

# 3. 根据套餐类型添加功能
# 标准版
db-migrator up -d tenant_new_005 --directory=tenant_standard

# 企业版
db-migrator up -d tenant_new_005 --directory=tenant_enterprise

# 4. 在平台数据库中更新租户状态
# （通过应用程序API完成）
```

### 3. 功能发布管理

```bash
# 为所有租户发布基础功能更新
db-migrator up --patterns=tenant_* --directory=tenant_base

# 只为企业版租户发布高级功能
db-migrator up --databases=tenant_001,tenant_003 --directory=tenant_enterprise

# 为标准版租户发布功能
db-migrator up --databases=tenant_002 --directory=tenant_standard
```

### 4. 套餐升级/降级

```bash
# 客户从标准版升级到企业版
# 1. 添加企业版功能
db-migrator up -d tenant_002 --directory=tenant_enterprise

# 2. 在平台数据库中更新套餐信息
# （通过应用程序API完成）

# 客户降级（需要移除某些功能表）
# 通常需要自定义迁移来处理数据保留/清理
```

### 5. 系统维护

```bash
# 查看所有租户状态
db-migrator status --patterns=tenant_*

# 为所有租户执行性能优化
db-migrator up --patterns=tenant_* --directory=maintenance

# 清理试用期过期的租户
db-migrator status --patterns=tenant_trial_*
```

## 📊 日常运维场景

### 场景1：批量租户升级

```bash
# 1. 查看当前所有租户状态
db-migrator status --patterns=tenant_*

# 2. 先在一个测试租户验证
db-migrator up -d tenant_trial_001

# 3. 分批升级生产租户
db-migrator up --databases=tenant_001,tenant_002
db-migrator up --databases=tenant_003,tenant_004

# 4. 检查升级结果
db-migrator status --patterns=tenant_*
```

### 场景2：新功能灰度发布

```bash
# 1. 先在试用租户测试
db-migrator up --patterns=tenant_trial_*

# 2. 再在一个标准版租户测试
db-migrator up -d tenant_002

# 3. 全面发布
db-migrator up --patterns=tenant_*
```

### 场景3：数据合规性检查

```bash
# 为所有租户添加GDPR合规字段
db-migrator create add_gdpr_fields
# 编辑迁移文件...
db-migrator up --patterns=tenant_*
```

## 🔧 故障排除

### 问题1：某个租户迁移失败

```bash
# 查看具体错误
db-migrator status -d tenant_002

# 检查数据库连接
mysql -u saas_admin -p -h saas-db.company.com -e "USE tenant_002_db; SHOW TABLES;"

# 重试迁移
db-migrator up -d tenant_002
```

### 问题2：企业版功能应用错误

```bash
# 检查哪些租户应该有企业版功能
mysql -u saas_admin -p -h saas-db.company.com platform_db -e "
SELECT tenant_code, plan_type FROM tenants WHERE plan_type = 'enterprise';"

# 为企业版租户重新应用功能
db-migrator up --databases=tenant_001,tenant_003 --directory=tenant_enterprise
```

### 问题3：新租户开通失败

```bash
# 检查数据库是否创建成功
mysql -u saas_admin -p -h saas-db.company.com -e "SHOW DATABASES LIKE 'tenant_new_%'"

# 检查配置文件中是否添加了租户配置

# 重新执行初始化
db-migrator up -d tenant_new_005 --directory=tenant_base
```

## 🎯 最佳实践

1. **租户隔离**：严格按套餐类型控制功能迁移
2. **分阶段部署**：先试用租户→标准版→企业版
3. **监控告警**：设置迁移失败和性能监控
4. **数据备份**：关键迁移前自动备份
5. **套餐管理**：通过平台数据库统一管理租户状态
6. **自动化流程**：集成到租户开通/升级的自动化流程中

这个案例展示了如何使用 `tenant_*` 模式管理SaaS多租户系统，支持不同套餐的差异化功能！ 