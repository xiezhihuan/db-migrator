# DB Migrator 使用场景手册

本文档展示了在各种实际开发场景下如何使用 DB Migrator 工具。

## 📋 目录

1. [基础使用场景](#基础使用场景)
2. [电商系统场景](#电商系统场景)
3. [权限管理场景](#权限管理场景)
4. [日志系统场景](#日志系统场景)
5. [数据迁移场景](#数据迁移场景)
6. [高级使用技巧](#高级使用技巧)

## 🚀 基础使用场景

### 场景1：创建简单的用户表

```go
package migrations

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
    checker := checker.NewMySQLChecker(db, "your_database")
    advancedBuilder := builder.NewAdvancedBuilder(checker, db)

    return advancedBuilder.Table("users").
        ID().
        String("email", 255).NotNull().Unique().Comment("邮箱").End().
        String("password", 255).NotNull().Comment("密码").End().
        String("name", 100).NotNull().Comment("姓名").End().
        Timestamps().
        Engine("InnoDB").
        Comment("用户表").
        Create(ctx)
}

func (m *CreateUsersTableMigration) Down(ctx context.Context, db types.DB) error {
    _, err := db.Exec("DROP TABLE IF EXISTS users")
    return err
}
```

### 场景2：为现有表添加列

```go
func (m *AddUserPhoneMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "your_database")
    advancedBuilder := builder.NewAdvancedBuilder(checker, db)

    // 添加手机号列
    return advancedBuilder.ModifyTable("users").
        AddColumn("phone", builder.TypeVarchar, 20).
        Comment("手机号").
        After("email").
        Execute(ctx)
}
```

### 场景3：创建索引

```go
func (m *AddUserIndexesMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "your_database")
    advancedBuilder := builder.NewAdvancedBuilder(checker, db)

    tableModifier := advancedBuilder.ModifyTable("users")
    
    // 添加邮箱索引
    err := tableModifier.AddIndex(ctx, "idx_users_email", []string{"email"}, false)
    if err != nil {
        return err
    }

    // 添加电话唯一索引
    return tableModifier.AddIndex(ctx, "uk_users_phone", []string{"phone"}, true)
}
```

## 🛒 电商系统场景

### 场景4：创建完整的产品管理系统

```go
func (m *CreateProductSystemMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "ecommerce_db")
    advancedBuilder := builder.NewAdvancedBuilder(checker, db)

    // 创建分类表
    err := advancedBuilder.Table("categories").
        ID().
        String("name", 100).NotNull().Comment("分类名称").End().
        String("slug", 100).NotNull().Unique().Comment("URL友好名称").End().
        Text("description").Nullable().Comment("描述").End().
        Integer("parent_id").Nullable().Comment("父分类").End().
        Integer("sort_order").Default(0).Comment("排序").End().
        Boolean("is_active").Default(true).Comment("是否激活").End().
        Timestamps().
        ForeignKey("parent_id").References("categories", "id").OnDelete(builder.ActionSetNull).End().
        Index("parent_id").End().
        Index("sort_order").End().
        Create(ctx)
    if err != nil {
        return err
    }

    // 创建产品表
    return advancedBuilder.Table("products").
        ID().
        String("sku", 50).NotNull().Unique().Comment("SKU").End().
        String("name", 200).NotNull().Comment("产品名称").End().
        Text("description").Nullable().Comment("描述").End().
        Integer("category_id").NotNull().Comment("分类ID").End().
        Decimal("price", 10, 2).NotNull().Comment("价格").End().
        Integer("stock").Default(0).Comment("库存").End().
        Enum("status", []string{"draft", "active", "inactive"}).Default("draft").End().
        Json("attributes").Nullable().Comment("产品属性").End().
        Timestamps().
        ForeignKey("category_id").References("categories", "id").OnDelete(builder.ActionRestrict).End().
        Index("category_id").End().
        Index("status").End().
        Index("price").End().
        Engine("InnoDB").
        Comment("产品表").
        Create(ctx)
}
```

### 场景5：创建订单系统

```go
func (m *CreateOrderSystemMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "ecommerce_db")
    advancedBuilder := builder.NewAdvancedBuilder(checker, db)

    // 创建订单表
    err := advancedBuilder.Table("orders").
        ID().
        String("order_number", 50).NotNull().Unique().Comment("订单号").End().
        Integer("user_id").NotNull().Comment("用户ID").End().
        Enum("status", []string{"pending", "confirmed", "shipped", "delivered", "cancelled"}).
            Default("pending").Comment("订单状态").End().
        Decimal("total_amount", 10, 2).NotNull().Comment("总金额").End().
        Json("shipping_address").NotNull().Comment("配送地址").End().
        Timestamps().
        ForeignKey("user_id").References("users", "id").OnDelete(builder.ActionRestrict).End().
        Index("user_id").End().
        Index("status").End().
        Create(ctx)
    if err != nil {
        return err
    }

    // 创建订单项表
    return advancedBuilder.Table("order_items").
        ID().
        Integer("order_id").NotNull().Comment("订单ID").End().
        Integer("product_id").NotNull().Comment("产品ID").End().
        Integer("quantity").NotNull().Comment("数量").End().
        Decimal("unit_price", 10, 2).NotNull().Comment("单价").End().
        Decimal("total_price", 10, 2).NotNull().Comment("总价").End().
        Timestamps().
        ForeignKey("order_id").References("orders", "id").OnDelete(builder.ActionCascade).End().
        ForeignKey("product_id").References("products", "id").OnDelete(builder.ActionRestrict).End().
        Index("order_id").End().
        Index("product_id").End().
        Create(ctx)
}
```

## 🔐 权限管理场景

### 场景6：创建RBAC权限系统

```go
func (m *CreateRBACMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "app_db")
    advancedBuilder := builder.NewAdvancedBuilder(checker, db)

    // 创建角色表
    err := advancedBuilder.Table("roles").
        ID().
        String("name", 100).NotNull().Unique().Comment("角色名称").End().
        String("display_name", 200).Nullable().Comment("显示名称").End().
        Text("description").Nullable().Comment("描述").End().
        Boolean("is_system").Default(false).Comment("是否系统角色").End().
        Timestamps().
        Index("name").End().
        Create(ctx)
    if err != nil {
        return err
    }

    // 创建权限表
    err = advancedBuilder.Table("permissions").
        ID().
        String("name", 100).NotNull().Unique().Comment("权限名称").End().
        String("resource", 100).NotNull().Comment("资源").End().
        String("action", 100).NotNull().Comment("动作").End().
        Text("description").Nullable().Comment("描述").End().
        Timestamps().
        Index("name").End().
        Index("resource").End().
        Create(ctx)
    if err != nil {
        return err
    }

    // 创建用户角色关联表
    return advancedBuilder.Table("user_roles").
        ID().
        Integer("user_id").NotNull().Comment("用户ID").End().
        Integer("role_id").NotNull().Comment("角色ID").End().
        Timestamp("expires_at").Nullable().Comment("过期时间").End().
        Timestamps().
        ForeignKey("user_id").References("users", "id").OnDelete(builder.ActionCascade).End().
        ForeignKey("role_id").References("roles", "id").OnDelete(builder.ActionCascade).End().
        Index("user_id").End().
        Index("role_id").End().
        Unique("user_id", "role_id").End().
        Create(ctx)
}
```

## 📊 日志系统场景

### 场景7：创建应用日志系统

```go
func (m *CreateLoggingMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "app_db")
    advancedBuilder := builder.NewAdvancedBuilder(checker, db)

    // 创建应用日志表
    err := advancedBuilder.Table("app_logs").
        ID().
        Enum("level", []string{"debug", "info", "warning", "error", "critical"}).
            NotNull().Comment("日志级别").End().
        String("message", 1000).NotNull().Comment("日志消息").End().
        String("logger", 100).NotNull().Comment("记录器").End().
        Json("context").Nullable().Comment("上下文").End().
        String("user_id", 36).Nullable().Comment("用户ID").End().
        String("ip_address", 45).Nullable().Comment("IP地址").End().
        String("user_agent", 500).Nullable().Comment("用户代理").End().
        Timestamp("logged_at").Default("CURRENT_TIMESTAMP").Comment("记录时间").End().
        Index("level").End().
        Index("logger").End().
        Index("user_id").End().
        Index("logged_at").End().
        Create(ctx)
    if err != nil {
        return err
    }

    // 创建错误日志表
    return advancedBuilder.Table("error_logs").
        ID().
        String("error_type", 100).NotNull().Comment("错误类型").End().
        String("message", 1000).NotNull().Comment("错误消息").End().
        Text("stack_trace").Nullable().Comment("堆栈跟踪").End().
        String("file", 500).Nullable().Comment("文件").End().
        Integer("line").Nullable().Comment("行号").End().
        Json("context").Nullable().Comment("错误上下文").End().
        Boolean("is_resolved").Default(false).Comment("是否已解决").End().
        Timestamps().
        Index("error_type").End().
        Index("is_resolved").End().
        Create(ctx)
}
```

## 🔄 数据迁移场景

### 场景8：从旧系统迁移数据

```go
func (m *MigrateUserDataMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "app_db")
    advancedBuilder := builder.NewAdvancedBuilder(checker, db)

    // 1. 创建临时映射表
    err := advancedBuilder.Table("user_migration_mapping").
        ID().
        String("old_user_id", 100).NotNull().Comment("旧用户ID").End().
        Integer("new_user_id").NotNull().Comment("新用户ID").End().
        Boolean("migrated").Default(false).Comment("是否已迁移").End().
        Timestamps().
        Index("old_user_id").End().
        Index("new_user_id").End().
        Create(ctx)
    if err != nil {
        return err
    }

    // 2. 批量插入用户数据
    userData := [][]interface{}{
        {"john@example.com", "hashed_password", "John Doe", "2023-01-01"},
        {"jane@example.com", "hashed_password", "Jane Smith", "2023-01-02"},
        {"bob@example.com", "hashed_password", "Bob Wilson", "2023-01-03"},
    }

    columns := []string{"email", "password", "name", "migrated_at"}
    return advancedBuilder.BulkInsert(ctx, "users", columns, userData)
}
```

### 场景9：数据结构重构

```go
func (m *RefactorUserTableMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "app_db")
    advancedBuilder := builder.NewAdvancedBuilder(checker, db)

    // 1. 创建新的用户表结构
    err := advancedBuilder.Table("users_new").
        ID().
        String("email", 255).NotNull().Unique().Comment("邮箱").End().
        String("password_hash", 255).NotNull().Comment("密码哈希").End().
        String("first_name", 100).Nullable().Comment("名").End().
        String("last_name", 100).Nullable().Comment("姓").End().
        String("phone", 20).Nullable().Comment("电话").End().
        Enum("status", []string{"active", "inactive", "suspended"}).Default("active").End().
        Json("profile").Nullable().Comment("用户档案").End().
        Timestamps().
        Index("email").End().
        Index("status").End().
        Create(ctx)
    if err != nil {
        return err
    }

    // 2. 迁移数据（分割name为first_name和last_name）
    _, err = db.Exec(`
        INSERT INTO users_new (email, password_hash, first_name, last_name, phone, status, created_at, updated_at)
        SELECT 
            email,
            password,
            SUBSTRING_INDEX(name, ' ', 1) as first_name,
            SUBSTRING_INDEX(name, ' ', -1) as last_name,
            NULL as phone,
            'active' as status,
            created_at,
            updated_at
        FROM users
    `)
    if err != nil {
        return err
    }

    // 3. 重命名表
    err = advancedBuilder.RenameTable(ctx, "users", "users_old")
    if err != nil {
        return err
    }

    return advancedBuilder.RenameTable(ctx, "users_new", "users")
}
```

## 🔧 高级使用技巧

### 场景10：使用结构体驱动的表创建

```go
// 定义用户结构体
type User struct {
    ID        int       `db:"id,primary_key,auto_increment"`
    Email     string    `db:"email,not_null,unique,size:255" comment:"用户邮箱"`
    Username  string    `db:"username,not_null,unique,size:100" comment:"用户名"`
    Password  string    `db:"password_hash,not_null,size:255" comment:"密码哈希"`
    FirstName string    `db:"first_name,size:100" comment:"名"`
    LastName  string    `db:"last_name,size:100" comment:"姓"`
    IsActive  bool      `db:"is_active,default:true" comment:"是否激活"`
    CreatedAt time.Time `db:"created_at,default:CURRENT_TIMESTAMP"`
    UpdatedAt time.Time `db:"updated_at,default:CURRENT_TIMESTAMP"`
}

func (m *CreateUserFromStructMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "app_db")
    sqlBuilder := builder.NewSQLBuilder(checker, db)

    // 从结构体创建表
    tableBuilder := builder.CreateTableFromStruct(sqlBuilder, "users", User{})
    
    return tableBuilder.
        Engine("InnoDB").
        Comment("用户表").
        Create(ctx)
}
```

### 场景11：创建函数和存储过程

```go
func (m *CreateUserFunctionsMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "app_db")
    advancedBuilder := builder.NewAdvancedBuilder(checker, db)

    // 创建计算用户年龄的函数
    err := advancedBuilder.CreateFunction(ctx, "calculate_user_age", `
        CREATE FUNCTION calculate_user_age(birth_date DATE) 
        RETURNS INT
        READS SQL DATA
        DETERMINISTIC
        BEGIN
            DECLARE age INT;
            IF birth_date IS NULL THEN
                RETURN NULL;
            END IF;
            SET age = TIMESTAMPDIFF(YEAR, birth_date, CURDATE());
            RETURN age;
        END
    `)
    if err != nil {
        return err
    }

    // 创建用户统计存储过程
    return advancedBuilder.CreateStoredProcedure(ctx, "get_user_stats", `
        CREATE PROCEDURE get_user_stats()
        BEGIN
            SELECT 
                COUNT(*) as total_users,
                COUNT(CASE WHEN status = 'active' THEN 1 END) as active_users,
                COUNT(CASE WHEN status = 'inactive' THEN 1 END) as inactive_users,
                AVG(calculate_user_age(birth_date)) as avg_age
            FROM users;
        END
    `)
}
```

### 场景12：创建视图

```go
func (m *CreateUserViewsMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "app_db")
    advancedBuilder := builder.NewAdvancedBuilder(checker, db)

    // 创建活跃用户视图
    return advancedBuilder.CreateView(ctx, "active_users", `
        SELECT 
            id,
            email,
            username,
            CONCAT(first_name, ' ', last_name) as full_name,
            created_at
        FROM users 
        WHERE status = 'active' AND is_active = true
    `)
}
```

### 场景13：复杂的条件迁移

```go
func (m *ConditionalMigrationMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "app_db")
    advancedBuilder := builder.NewAdvancedBuilder(checker, db)

    // 检查是否需要添加新列
    exists, err := checker.ColumnExists(ctx, "users", "last_login_at")
    if err != nil {
        return err
    }

    if !exists {
        // 添加最后登录时间列
        err = advancedBuilder.ModifyTable("users").
            AddColumn("last_login_at", builder.TypeTimestamp, 0).
            Nullable().
            Comment("最后登录时间").
            Execute(ctx)
        if err != nil {
            return err
        }
    }

    // 检查数据并更新
    var userCount int
    err = db.QueryRow("SELECT COUNT(*) FROM users WHERE status = 'pending'").Scan(&userCount)
    if err != nil {
        return err
    }

    if userCount > 0 {
        // 将待处理用户设为激活状态
        _, err = db.Exec("UPDATE users SET status = 'active' WHERE status = 'pending'")
        if err != nil {
            return err
        }
    }

    return nil
}
```

## 📝 最佳实践

### 1. 使用事务
```go
func (m *TransactionalMigration) Up(ctx context.Context, db types.DB) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 执行多个操作
    // ... 迁移逻辑 ...

    return tx.Commit()
}
```

### 2. 错误处理
```go
func (m *ErrorHandlingMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "app_db")
    advancedBuilder := builder.NewAdvancedBuilder(checker, db)

    err := advancedBuilder.Table("example").
        ID().
        String("name", 100).NotNull().End().
        Create(ctx)
    if err != nil {
        return fmt.Errorf("创建示例表失败: %w", err)
    }

    return nil
}
```

### 3. 数据验证
```go
func (m *DataValidationMigration) Up(ctx context.Context, db types.DB) error {
    // 验证数据完整性
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email IS NULL OR email = ''").Scan(&count)
    if err != nil {
        return err
    }

    if count > 0 {
        return fmt.Errorf("发现 %d 个用户的邮箱为空，请先清理数据", count)
    }

    // 继续迁移...
    return nil
}
```

## 🎯 小结

通过这些使用场景，你可以看到 DB Migrator 提供了：

1. **简洁的API**：链式调用，易于阅读和编写
2. **智能检查**：自动跳过已存在的对象
3. **类型安全**：Go代码定义，编译时检查
4. **灵活性**：支持复杂的业务场景
5. **安全性**：事务支持，错误处理完善

使用这个工具，你几乎不需要编写原生SQL，就能完成各种复杂的数据库迁移任务。 