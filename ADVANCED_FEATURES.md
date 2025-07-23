# DB Migrator 高级功能详解

本文档详细介绍了 DB Migrator 工具的所有高级功能和 API。

## 🏗️ 核心架构

### 高级表构建器 (TableBuilder)

基于链式调用的表构建器，提供完全无SQL的数据库对象创建体验：

```go
advancedBuilder.Table("products").
    ID().                                                    // 自动主键
    String("name", 200).NotNull().Comment("产品名称").End().     // 字符串列
    Decimal("price", 10, 2).NotNull().Comment("价格").End().    // 小数列
    Enum("status", []string{"draft", "active"}).Default("draft").End().  // 枚举列
    Json("metadata").Nullable().End().                       // JSON列
    Timestamps().                                            // 自动时间戳
    Index("name").End().                                     // 普通索引
    Unique("name").End().                                    // 唯一索引
    ForeignKey("category_id").References("categories", "id").OnDelete(ActionCascade).End().  // 外键
    Engine("InnoDB").Comment("产品表").                       // 表选项
    Create(ctx)                                              // 执行创建
```

### 智能存在性检查

所有操作都会自动进行存在性检查，避免重复创建：

- `TableExists()` - 检查表是否存在
- `ColumnExists()` - 检查列是否存在
- `IndexExists()` - 检查索引是否存在
- `FunctionExists()` - 检查函数是否存在
- `ConstraintExists()` - 检查约束是否存在
- `TriggerExists()` - 检查触发器是否存在

## 📋 完整数据类型支持

### 数值类型
```go
.Integer("count")                    // INT
.BigInteger("large_number")          // BIGINT
.SmallInt("small_number")           // SMALLINT
.TinyInt("tiny_number")             // TINYINT
.Decimal("price", 10, 2)            // DECIMAL(10,2)
.Float("rating")                    // FLOAT
.Double("precise_value")            // DOUBLE
```

### 字符串类型
```go
.String("name", 100)                // VARCHAR(100)
.Text("description")                // TEXT
.Char("code", 5)                    // CHAR(5)
```

### 日期时间类型
```go
.Date("birth_date")                 // DATE
.DateTime("event_time")             // DATETIME
.Timestamp("created_at")            // TIMESTAMP
.Time("duration")                   // TIME
```

### 特殊类型
```go
.Boolean("is_active")               // BOOLEAN
.Json("metadata")                   // JSON
.Enum("status", []string{"a", "b"}) // ENUM('a','b')
.Blob("binary_data")                // BLOB
```

## 🔗 关系管理

### 外键关系
```go
.ForeignKey("user_id").
    References("users", "id").
    OnDelete(ActionCascade).        // 级联删除
    OnUpdate(ActionRestrict).       // 限制更新
    End()

// 可选的引用动作：
// ActionCascade    - CASCADE
// ActionSetNull    - SET NULL
// ActionRestrict   - RESTRICT
// ActionNoAction   - NO ACTION
// ActionSetDefault - SET DEFAULT
```

### 索引管理
```go
// 普通索引
.Index("column1", "column2").Name("custom_index_name").End()

// 唯一索引
.Unique("email").End()

// 复合唯一索引
.Unique("user_id", "product_id").End()
```

## 🔧 高级表修改

### 表结构修改器 (TableModifier)
```go
modifier := advancedBuilder.ModifyTable("users")

// 添加列
modifier.AddColumn("phone", TypeVarchar, 20).
    NotNull().
    Comment("电话号码").
    After("email").
    Execute(ctx)

// 修改列
modifier.ModifyColumn("name", TypeVarchar, 200).
    NotNull().
    Execute(ctx)

// 删除列
modifier.DropColumn(ctx, "old_column")

// 重命名列
modifier.RenameColumn(ctx, "old_name", "new_name", "VARCHAR(100) NOT NULL")

// 添加索引
modifier.AddIndex(ctx, "idx_phone", []string{"phone"}, false)
```

### 表操作
```go
// 重命名表
advancedBuilder.RenameTable(ctx, "old_table", "new_table")

// 复制表结构
advancedBuilder.CopyTable(ctx, "source_table", "target_table", false)

// 复制表结构和数据
advancedBuilder.CopyTable(ctx, "source_table", "target_table", true)

// 清空表数据
advancedBuilder.TruncateTable(ctx, "table_name")
```

## 🎯 结构体驱动开发

### 从Go结构体生成表
```go
type User struct {
    ID        int       `db:"id,primary_key,auto_increment"`
    Email     string    `db:"email,not_null,unique,size:255"`
    Username  string    `db:"username,not_null,size:100"`
    FirstName string    `db:"first_name,size:100"`
    IsActive  bool      `db:"is_active,default:true"`
    CreatedAt time.Time `db:"created_at,default:CURRENT_TIMESTAMP"`
}

// 自动生成表结构
tableBuilder := builder.CreateTableFromStruct(sqlBuilder, "users", User{})
tableBuilder.Engine("InnoDB").Comment("用户表").Create(ctx)
```

### 支持的标签
- `primary_key` - 主键
- `auto_increment` - 自增
- `not_null` - 非空
- `unique` - 唯一
- `size:N` - 字段长度
- `default:value` - 默认值
- `comment:text` - 注释

## 🗃️ 数据库对象管理

### 视图管理
```go
// 创建视图
advancedBuilder.CreateView(ctx, "active_users", `
    SELECT id, email, name 
    FROM users 
    WHERE status = 'active'
`)

// 删除视图
advancedBuilder.DropView(ctx, "active_users")
```

### 函数管理
```go
// 创建函数
advancedBuilder.CreateFunction(ctx, "calculate_age", `
    CREATE FUNCTION calculate_age(birth_date DATE) 
    RETURNS INT
    DETERMINISTIC
    BEGIN
        RETURN TIMESTAMPDIFF(YEAR, birth_date, CURDATE());
    END
`)

// 删除函数
advancedBuilder.DropFunction(ctx, "calculate_age")
```

### 存储过程管理
```go
// 创建存储过程
advancedBuilder.CreateStoredProcedure(ctx, "get_user_stats", `
    CREATE PROCEDURE get_user_stats()
    BEGIN
        SELECT COUNT(*) as total_users FROM users;
    END
`)
```

### 触发器管理
```go
// 创建触发器
advancedBuilder.CreateTrigger(ctx, "user_updated", `
    CREATE TRIGGER user_updated 
    BEFORE UPDATE ON users
    FOR EACH ROW
    BEGIN
        SET NEW.updated_at = NOW();
    END
`)
```

## 📊 数据操作

### 批量数据插入
```go
// 准备数据
data := [][]interface{}{
    {"john@example.com", "John Doe", "active"},
    {"jane@example.com", "Jane Smith", "active"},
    {"bob@example.com", "Bob Wilson", "inactive"},
}

columns := []string{"email", "name", "status"}

// 批量插入（自动分批处理）
advancedBuilder.BulkInsert(ctx, "users", columns, data)
```

### 智能数据操作
```go
// 智能插入（检查后插入）
builder.InsertIfNotExists(ctx, "settings", 
    "key = 'theme'", 
    "INSERT INTO settings (key, value) VALUES ('theme', 'dark')")

// 智能更新（检查后更新）
builder.UpdateIfExists(ctx, "users", 
    "email = 'john@example.com'",
    "UPDATE users SET last_login = NOW() WHERE email = 'john@example.com'")
```

## 🔄 数据迁移工具

### 迁移任务管理
```go
// 创建迁移任务追踪
advancedBuilder.Table("migration_tasks").
    ID().
    String("name", 200).NotNull().End().
    String("source_table", 100).NotNull().End().
    String("target_table", 100).NotNull().End().
    Enum("status", []string{"pending", "running", "completed", "failed"}).End().
    Integer("total_records").Default(0).End().
    Integer("processed_records").Default(0).End().
    Json("mapping_config").Nullable().End().
    Create(ctx)
```

### ID映射管理
```go
// 创建新旧ID映射表
advancedBuilder.Table("migration_mappings").
    ID().
    String("entity_type", 100).NotNull().End().
    String("source_id", 100).NotNull().End().
    String("target_id", 100).NotNull().End().
    Json("metadata").Nullable().End().
    Unique("entity_type", "source_id").End().
    Create(ctx)
```

## ⚡ 性能优化

### 索引策略
```go
// 复合索引
.Index("user_id", "created_at").Name("idx_user_timeline").End()

// 部分索引（通过条件）
.Index("email").Name("idx_active_users_email").End()  // 配合WHERE条件使用

// 全文索引
.Index("title", "content").Type("FULLTEXT").End()
```

### 分区表支持
```go
// 创建分区表（通过原生SQL）
builder.ExecuteRawSQL(ctx, `
    CREATE TABLE logs (
        id INT AUTO_INCREMENT,
        message TEXT,
        created_at TIMESTAMP,
        PRIMARY KEY (id, created_at)
    ) PARTITION BY RANGE (YEAR(created_at)) (
        PARTITION p2023 VALUES LESS THAN (2024),
        PARTITION p2024 VALUES LESS THAN (2025)
    )
`, "创建分区日志表")
```

## 🔐 安全特性

### 事务支持
```go
func (m *SafeMigration) Up(ctx context.Context, db types.DB) error {
    // 开始事务
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()  // 自动回滚

    // 执行迁移操作...
    
    // 提交事务
    return tx.Commit()
}
```

### 锁机制
```go
// 工具自动处理并发锁，确保同时只有一个迁移在执行
// 锁信息存储在 schema_migrations_lock 表中
```

### 错误处理
```go
// 统一错误类型
type Error struct {
    Code    string
    Message string
    Cause   error
}

// 常见错误码
const (
    ErrCodeMigrationFailed     = "MIGRATION_FAILED"
    ErrCodeDatabaseConnection  = "DATABASE_CONNECTION"
    ErrCodeConfigInvalid       = "CONFIG_INVALID"
    ErrCodeMigrationNotFound   = "MIGRATION_NOT_FOUND"
    ErrCodeVersionConflict     = "VERSION_CONFLICT"
)
```

## 📈 监控和日志

### 执行状态跟踪
- 每个迁移的执行时间
- 成功/失败状态
- 错误信息记录
- 操作审计日志

### 进度监控
```go
// 迁移状态查看
statuses, err := migrator.Status(ctx)
for _, status := range statuses {
    fmt.Printf("迁移 %s: %s (执行时间: %v)\n", 
        status.Version, 
        status.Applied ? "已执行" : "未执行",
        status.AppliedAt)
}
```

## 🔧 配置选项

### 完整配置示例
```yaml
database:
  driver: mysql
  host: localhost
  port: 3306
  username: root
  password: secret
  database: myapp_db
  charset: utf8mb4

migrator:
  migrations_table: schema_migrations      # 迁移记录表名
  lock_table: schema_migrations_lock       # 锁表名
  auto_backup: true                        # 自动备份
  dry_run: false                          # 干运行模式
```

### 环境变量支持
```bash
export DB_MIGRATOR_DATABASE_HOST=prod-db.example.com
export DB_MIGRATOR_DATABASE_PASSWORD=prod_password
export DB_MIGRATOR_MIGRATOR_DRY_RUN=true
```

## 🎯 最佳实践总结

1. **使用高级API**：优先使用TableBuilder而非原生SQL
2. **利用智能检查**：让工具自动处理存在性检查
3. **结构化迁移**：使用Go代码定义，获得类型安全
4. **事务包装**：重要操作使用事务确保一致性
5. **错误处理**：妥善处理迁移错误，提供详细信息
6. **版本管理**：使用有意义的版本号和描述
7. **测试优先**：在开发环境充分测试后再部署生产

通过这些高级功能，DB Migrator 提供了一个完整的、类型安全的、智能化的数据库迁移解决方案，让开发者能够专注于业务逻辑而不是底层SQL操作。 