# 案例5：SaaS新租户数据初始化

## 📋 业务场景

你运营一个SaaS平台，为每个新注册的租户（客户）创建独立的数据库实例，并初始化必要的基础数据，让租户能够立即开始使用系统。

### 系统架构
```
模板数据库 (tenant_template)
├── 用户角色模板 (roles)
├── 权限配置 (permissions)
├── 系统配置 (system_configs)
├── 业务流程模板 (workflow_templates)
├── 报表模板 (report_templates)
└── 示例数据 (sample_data)

租户数据库 (tenant_*)
├── tenant_company_a (A公司)
├── tenant_company_b (B公司) 
├── tenant_startup_c (初创公司C)
└── tenant_enterprise_d (企业D)
```

### 初始化需求
- **基础配置**：系统参数、功能开关
- **用户权限**：角色模板、权限分配
- **业务模板**：工作流程、审批流程
- **示例数据**：演示数据、测试数据
- **定制配置**：根据租户类型定制功能

## ⚙️ 初始化策略

### 按租户类型分类
```yaml
# tenant-types.yaml
tenant_types:
  startup:
    name: "初创企业版"
    features: ["basic_crm", "simple_workflow", "basic_reports"]
    user_limit: 10
    storage_limit: "1GB"
    templates: ["basic_roles", "simple_workflow"]
    
  enterprise:
    name: "企业版"  
    features: ["advanced_crm", "complex_workflow", "advanced_reports", "api_access"]
    user_limit: 1000
    storage_limit: "100GB"
    templates: ["enterprise_roles", "advanced_workflow", "compliance_templates"]
    
  trial:
    name: "试用版"
    features: ["basic_crm", "demo_data"]
    user_limit: 3
    storage_limit: "100MB"
    trial_days: 30
    templates: ["trial_roles", "demo_data"]
```

## 📝 迁移文件示例

### 基础数据初始化迁移
```go
// migrations/tenant_init/001_init_basic_data.go
package tenant_init

import (
    "context"
    "encoding/json"
    "fmt"
    "db-migrator/internal/builder"
    "db-migrator/internal/checker"
    "db-migrator/internal/types"
)

type InitBasicDataMigration struct{}

func (m *InitBasicDataMigration) Version() string {
    return "001"
}

func (m *InitBasicDataMigration) Description() string {
    return "初始化租户基础数据"
}

func (m *InitBasicDataMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "")
    dataBuilder := builder.NewDataBuilder(checker, db)

    // 1. 初始化系统配置
    if err := m.initSystemConfigs(ctx, dataBuilder); err != nil {
        return fmt.Errorf("初始化系统配置失败: %v", err)
    }

    // 2. 初始化用户角色
    if err := m.initUserRoles(ctx, dataBuilder); err != nil {
        return fmt.Errorf("初始化用户角色失败: %v", err)
    }

    // 3. 初始化权限配置
    if err := m.initPermissions(ctx, dataBuilder); err != nil {
        return fmt.Errorf("初始化权限配置失败: %v", err)
    }

    // 4. 初始化业务模板
    if err := m.initBusinessTemplates(ctx, dataBuilder); err != nil {
        return fmt.Errorf("初始化业务模板失败: %v", err)
    }

    return nil
}

func (m *InitBasicDataMigration) initSystemConfigs(ctx context.Context, dataBuilder *builder.DataBuilder) error {
    configs := []map[string]interface{}{
        {
            "category":    "system",
            "key":         "tenant_name",
            "value":       "新租户",
            "type":        "string",
            "editable":    true,
            "description": "租户名称",
        },
        {
            "category":    "system",
            "key":         "timezone",
            "value":       "Asia/Shanghai",
            "type":        "string",
            "editable":    true,
            "description": "时区设置",
        },
        {
            "category":    "system",
            "key":         "date_format",
            "value":       "YYYY-MM-DD",
            "type":        "string",
            "editable":    true,
            "description": "日期格式",
        },
        {
            "category":    "system",
            "key":         "currency",
            "value":       "CNY",
            "type":        "string",
            "editable":    true,
            "description": "货币单位",
        },
        {
            "category":    "features",
            "key":         "enable_api",
            "value":       "false",
            "type":        "boolean",
            "editable":    false,
            "description": "API访问开关",
        },
        {
            "category":    "limits",
            "key":         "max_users",
            "value":       "10",
            "type":        "integer",
            "editable":    false,
            "description": "最大用户数",
        },
    }

    return dataBuilder.Table("system_configs").
        Strategy(builder.StrategyInsertOrUpdate).
        InsertData(ctx, configs)
}

func (m *InitBasicDataMigration) initUserRoles(ctx context.Context, dataBuilder *builder.DataBuilder) error {
    roles := []map[string]interface{}{
        {
            "id":          1,
            "name":        "管理员",
            "code":        "admin",
            "description": "系统管理员，拥有所有权限",
            "is_system":   true,
            "created_at":  "NOW()",
        },
        {
            "id":          2,
            "name":        "经理",
            "code":        "manager",
            "description": "部门经理，拥有管理权限",
            "is_system":   true,
            "created_at":  "NOW()",
        },
        {
            "id":          3,
            "name":        "员工",
            "code":        "employee",
            "description": "普通员工，基础权限",
            "is_system":   true,
            "created_at":  "NOW()",
        },
        {
            "id":          4,
            "name":        "访客",
            "code":        "guest",
            "description": "访客用户，只读权限",
            "is_system":   true,
            "created_at":  "NOW()",
        },
    }

    return dataBuilder.Table("roles").
        Strategy(builder.StrategyInsertOrUpdate).
        InsertData(ctx, roles)
}

func (m *InitBasicDataMigration) initPermissions(ctx context.Context, dataBuilder *builder.DataBuilder) error {
    permissions := []map[string]interface{}{
        // 用户管理权限
        {"module": "users", "action": "create", "name": "创建用户", "description": "创建新用户"},
        {"module": "users", "action": "read", "name": "查看用户", "description": "查看用户信息"},
        {"module": "users", "action": "update", "name": "编辑用户", "description": "编辑用户信息"},
        {"module": "users", "action": "delete", "name": "删除用户", "description": "删除用户"},
        
        // 角色权限管理
        {"module": "roles", "action": "create", "name": "创建角色", "description": "创建新角色"},
        {"module": "roles", "action": "read", "name": "查看角色", "description": "查看角色信息"},
        {"module": "roles", "action": "update", "name": "编辑角色", "description": "编辑角色信息"},
        {"module": "roles", "action": "delete", "name": "删除角色", "description": "删除角色"},
        
        // 系统设置权限
        {"module": "settings", "action": "read", "name": "查看设置", "description": "查看系统设置"},
        {"module": "settings", "action": "update", "name": "修改设置", "description": "修改系统设置"},
        
        // 报表权限
        {"module": "reports", "action": "read", "name": "查看报表", "description": "查看各类报表"},
        {"module": "reports", "action": "export", "name": "导出报表", "description": "导出报表数据"},
    }

    return dataBuilder.Table("permissions").
        Strategy(builder.StrategyInsertOrUpdate).
        InsertData(ctx, permissions)
}

func (m *InitBasicDataMigration) initBusinessTemplates(ctx context.Context, dataBuilder *builder.DataBuilder) error {
    templates := []map[string]interface{}{
        {
            "type":        "workflow",
            "name":        "请假审批流程",
            "code":        "leave_approval",
            "description": "员工请假审批流程模板",
            "config":      `{"steps": [{"name": "提交申请", "role": "employee"}, {"name": "主管审批", "role": "manager"}, {"name": "HR确认", "role": "admin"}]}`,
            "is_active":   true,
        },
        {
            "type":        "workflow", 
            "name":        "采购申请流程",
            "code":        "purchase_approval",
            "description": "采购申请审批流程模板",
            "config":      `{"steps": [{"name": "提交申请", "role": "employee"}, {"name": "预算审核", "role": "manager"}, {"name": "最终审批", "role": "admin"}]}`,
            "is_active":   true,
        },
        {
            "type":        "report",
            "name":        "用户活跃度报表",
            "code":        "user_activity",
            "description": "用户活跃度统计报表",
            "config":      `{"fields": ["login_count", "last_login", "active_days"], "period": "monthly"}`,
            "is_active":   true,
        },
    }

    return dataBuilder.Table("business_templates").
        Strategy(builder.StrategyInsertOrUpdate).
        InsertData(ctx, templates)
}

func (m *InitBasicDataMigration) Down(ctx context.Context, db types.DB) error {
    // 通常不需要回滚初始化数据
    return nil
}
```

### 根据租户类型定制初始化
```go
// migrations/tenant_init/002_init_tenant_specific_data.go
package tenant_init

import (
    "context"
    "os"
    "db-migrator/internal/builder"
    "db-migrator/internal/checker"
    "db-migrator/internal/types"
)

type InitTenantSpecificDataMigration struct{}

func (m *InitTenantSpecificDataMigration) Version() string {
    return "002"
}

func (m *InitTenantSpecificDataMigration) Description() string {
    return "根据租户类型初始化定制数据"
}

func (m *InitTenantSpecificDataMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "")
    dataBuilder := builder.NewDataBuilder(checker, db)

    // 从环境变量获取租户类型
    tenantType := os.Getenv("TENANT_TYPE")
    if tenantType == "" {
        tenantType = "startup" // 默认类型
    }

    switch tenantType {
    case "enterprise":
        return m.initEnterpriseData(ctx, dataBuilder)
    case "trial":
        return m.initTrialData(ctx, dataBuilder)
    default:
        return m.initStartupData(ctx, dataBuilder)
    }
}

func (m *InitTenantSpecificDataMigration) initEnterpriseData(ctx context.Context, dataBuilder *builder.DataBuilder) error {
    // 企业版特性配置
    enterpriseConfigs := []map[string]interface{}{
        {"key": "max_users", "value": "1000"},
        {"key": "storage_limit", "value": "100GB"},
        {"key": "enable_api", "value": "true"},
        {"key": "enable_sso", "value": "true"},
        {"key": "enable_audit_log", "value": "true"},
        {"key": "backup_retention", "value": "90"},
    }

    for _, config := range enterpriseConfigs {
        err := dataBuilder.Table("system_configs").
            Strategy(builder.StrategyInsertOrUpdate).
            InsertData(ctx, []map[string]interface{}{
                {
                    "category": "limits",
                    "key":      config["key"],
                    "value":    config["value"],
                    "type":     "string",
                    "editable": false,
                },
            })
        if err != nil {
            return err
        }
    }

    // 企业版额外角色
    enterpriseRoles := []map[string]interface{}{
        {
            "name":        "审计员",
            "code":        "auditor",
            "description": "审计人员，查看审计日志",
            "is_system":   true,
        },
        {
            "name":        "系统集成员",
            "code":        "integrator",
            "description": "系统集成人员，API权限",
            "is_system":   true,
        },
    }

    return dataBuilder.Table("roles").
        Strategy(builder.StrategyInsertOrUpdate).
        InsertData(ctx, enterpriseRoles)
}

func (m *InitTenantSpecificDataMigration) initTrialData(ctx context.Context, dataBuilder *builder.DataBuilder) error {
    // 试用版限制配置
    trialConfigs := []map[string]interface{}{
        {"key": "max_users", "value": "3"},
        {"key": "storage_limit", "value": "100MB"},
        {"key": "trial_days", "value": "30"},
        {"key": "enable_api", "value": "false"},
        {"key": "enable_export", "value": "false"},
    }

    for _, config := range trialConfigs {
        err := dataBuilder.Table("system_configs").
            Strategy(builder.StrategyInsertOrUpdate).
            InsertData(ctx, []map[string]interface{}{
                {
                    "category": "limits",
                    "key":      config["key"],
                    "value":    config["value"],
                    "type":     "string",
                    "editable": false,
                },
            })
        if err != nil {
            return err
        }
    }

    // 试用版示例数据
    return m.insertDemoData(ctx, dataBuilder)
}

func (m *InitTenantSpecificDataMigration) initStartupData(ctx context.Context, dataBuilder *builder.DataBuilder) error {
    // 初创版配置
    startupConfigs := []map[string]interface{}{
        {"key": "max_users", "value": "10"},
        {"key": "storage_limit", "value": "1GB"},
        {"key": "enable_api", "value": "false"},
        {"key": "backup_retention", "value": "30"},
    }

    for _, config := range startupConfigs {
        err := dataBuilder.Table("system_configs").
            Strategy(builder.StrategyInsertOrUpdate).
            InsertData(ctx, []map[string]interface{}{
                {
                    "category": "limits",
                    "key":      config["key"],
                    "value":    config["value"],
                    "type":     "string",
                    "editable": false,
                },
            })
        if err != nil {
            return err
        }
    }

    return nil
}

func (m *InitTenantSpecificDataMigration) insertDemoData(ctx context.Context, dataBuilder *builder.DataBuilder) error {
    // 插入演示用户
    demoUsers := []map[string]interface{}{
        {
            "username":    "demo_admin",
            "email":       "admin@demo.com",
            "name":        "演示管理员",
            "role_id":     1,
            "is_demo":     true,
            "password":    "$2a$10$demo_password_hash",
            "created_at":  "NOW()",
        },
        {
            "username":    "demo_user",
            "email":       "user@demo.com",
            "name":        "演示用户",
            "role_id":     3,
            "is_demo":     true,
            "password":    "$2a$10$demo_password_hash",
            "created_at":  "NOW()",
        },
    }

    err := dataBuilder.Table("users").
        Strategy(builder.StrategyInsertOrUpdate).
        InsertData(ctx, demoUsers)
    if err != nil {
        return err
    }

    // 插入演示项目
    demoProjects := []map[string]interface{}{
        {
            "name":        "演示项目",
            "description": "这是一个演示项目，用于体验系统功能",
            "status":      "active",
            "owner_id":    1,
            "is_demo":     true,
            "created_at":  "NOW()",
        },
    }

    return dataBuilder.Table("projects").
        Strategy(builder.StrategyInsertOrUpdate).
        InsertData(ctx, demoProjects)
}

func (m *InitTenantSpecificDataMigration) Down(ctx context.Context, db types.DB) error {
    return nil
}
```

## 💻 操作命令

### 1. 从模板数据库初始化新租户
```bash
# 基础初始化（初创版）
db-migrator init-data \
  -d tenant_new_company \
  --from-db=tenant_template \
  --strategy=merge

# 指定租户类型初始化
TENANT_TYPE=enterprise db-migrator init-data \
  -d tenant_big_corp \
  --from-db=tenant_template \
  --strategy=merge

# 试用版初始化（包含演示数据）
TENANT_TYPE=trial db-migrator init-data \
  -d tenant_trial_user \
  --from-db=tenant_template \
  --strategy=merge
```

### 2. 使用数据文件初始化
```bash
# 从JSON文件初始化基础配置
db-migrator init-data \
  -d tenant_new_company \
  --data-file=configs/startup-config.json

# 从YAML文件初始化用户角色
db-migrator init-data \
  -d tenant_new_company \
  --data-file=configs/enterprise-roles.yaml
```

### 3. 批量初始化多个租户
```bash
# 为所有新租户执行初始化迁移
db-migrator up --patterns=tenant_new_* --directory=tenant_init

# 为特定租户批量初始化
db-migrator init-data \
  --databases=tenant_company_a,tenant_company_b,tenant_company_c \
  --from-db=tenant_template
```

### 4. 分步骤初始化
```bash
# 步骤1：基础数据初始化
db-migrator init-data \
  -d tenant_new_company \
  --data-type=basic_configs

# 步骤2：用户权限初始化  
db-migrator init-data \
  -d tenant_new_company \
  --data-type=user_permissions

# 步骤3：业务模板初始化
db-migrator init-data \
  -d tenant_new_company \
  --data-type=business_templates

# 步骤4：示例数据初始化（可选）
db-migrator init-data \
  -d tenant_new_company \
  --data-type=demo_data
```

## 📊 数据文件示例

### 企业版配置文件
```json
// configs/enterprise-config.json
{
  "system_configs": [
    {
      "category": "features",
      "key": "enable_api",
      "value": "true",
      "type": "boolean",
      "description": "API访问功能"
    },
    {
      "category": "features", 
      "key": "enable_sso",
      "value": "true",
      "type": "boolean",
      "description": "单点登录功能"
    },
    {
      "category": "limits",
      "key": "max_users",
      "value": "1000",
      "type": "integer",
      "description": "最大用户数"
    },
    {
      "category": "limits",
      "key": "storage_limit_gb",
      "value": "100",
      "type": "integer", 
      "description": "存储空间限制（GB）"
    }
  ],
  "additional_roles": [
    {
      "name": "API管理员",
      "code": "api_admin",
      "description": "管理API密钥和权限",
      "permissions": ["api:create", "api:read", "api:update", "api:delete"]
    },
    {
      "name": "数据分析师",
      "code": "analyst",
      "description": "数据分析和报表权限",
      "permissions": ["reports:read", "reports:export", "analytics:read"]
    }
  ]
}
```

### 试用版数据文件
```yaml
# configs/trial-data.yaml
system_configs:
  - category: "limits"
    key: "trial_days"
    value: "30"
    type: "integer"
    description: "试用天数"
  - category: "features"
    key: "watermark_enabled"
    value: "true"
    type: "boolean"
    description: "试用版水印"

demo_users:
  - username: "trial_admin"
    email: "admin@trial.example.com"
    name: "试用管理员"
    role: "admin"
    is_demo: true
  - username: "trial_user"
    email: "user@trial.example.com"
    name: "试用用户"
    role: "employee"
    is_demo: true

demo_projects:
  - name: "示例项目A"
    description: "演示项目管理功能"
    status: "active"
    is_demo: true
  - name: "示例项目B"
    description: "演示协作功能"
    status: "active"
    is_demo: true
```

## 🔄 自动化脚本

### 新租户创建脚本
```bash
#!/bin/bash
# create-new-tenant.sh

set -e

TENANT_NAME="$1"
TENANT_TYPE="$2"
TENANT_EMAIL="$3"

if [ -z "$TENANT_NAME" ] || [ -z "$TENANT_TYPE" ] || [ -z "$TENANT_EMAIL" ]; then
    echo "用法: $0 <租户名> <租户类型> <租户邮箱>"
    echo "租户类型: startup, enterprise, trial"
    exit 1
fi

TENANT_DB="tenant_${TENANT_NAME}"

echo "🚀 开始创建新租户: $TENANT_NAME ($TENANT_TYPE)"

# 1. 创建数据库
echo "📊 创建数据库..."
mysql -e "CREATE DATABASE IF NOT EXISTS \`$TENANT_DB\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"

# 2. 执行表结构迁移
echo "🏗️ 创建表结构..."
db-migrator up -d "$TENANT_DB"

# 3. 初始化基础数据
echo "🗃️ 初始化基础数据..."
TENANT_TYPE="$TENANT_TYPE" db-migrator up \
  -d "$TENANT_DB" \
  --directory=tenant_init

# 4. 根据类型加载特定配置
echo "⚙️ 加载租户类型配置..."
case "$TENANT_TYPE" in
    "enterprise")
        db-migrator init-data \
          -d "$TENANT_DB" \
          --data-file=configs/enterprise-config.json
        ;;
    "trial")
        db-migrator init-data \
          -d "$TENANT_DB" \
          --data-file=configs/trial-data.yaml
        ;;
    "startup")
        db-migrator init-data \
          -d "$TENANT_DB" \
          --data-file=configs/startup-config.json
        ;;
esac

# 5. 创建租户管理员用户
echo "👤 创建租户管理员..."
mysql "$TENANT_DB" -e "
INSERT INTO users (username, email, name, role_id, is_active, created_at)
VALUES ('admin', '$TENANT_EMAIL', '$TENANT_NAME 管理员', 1, 1, NOW())
ON DUPLICATE KEY UPDATE email='$TENANT_EMAIL', name='$TENANT_NAME 管理员'"

# 6. 更新租户配置
echo "🔧 更新租户配置..."
mysql "$TENANT_DB" -e "
UPDATE system_configs 
SET value='$TENANT_NAME' 
WHERE category='system' AND key='tenant_name'"

# 7. 记录创建日志
echo "📝 记录创建日志..."
mysql saas_management -e "
INSERT INTO tenant_creation_log (tenant_name, tenant_type, tenant_db, admin_email, created_at, status)
VALUES ('$TENANT_NAME', '$TENANT_TYPE', '$TENANT_DB', '$TENANT_EMAIL', NOW(), 'completed')"

echo "✅ 租户 $TENANT_NAME 创建完成！"
echo "📍 数据库: $TENANT_DB"
echo "👤 管理员邮箱: $TENANT_EMAIL"
echo "⭐ 租户类型: $TENANT_TYPE"

# 8. 发送欢迎邮件（可选）
# python send_welcome_email.py "$TENANT_EMAIL" "$TENANT_NAME" "$TENANT_TYPE"
```

### 批量租户初始化脚本
```bash
#!/bin/bash
# batch-tenant-init.sh

# 从CSV文件批量创建租户
CSV_FILE="$1"

if [ ! -f "$CSV_FILE" ]; then
    echo "错误: CSV文件不存在"
    exit 1
fi

echo "📋 开始批量创建租户..."

# 跳过标题行，逐行处理
tail -n +2 "$CSV_FILE" | while IFS=',' read -r tenant_name tenant_type admin_email company_name; do
    echo "处理租户: $tenant_name"
    
    if ./create-new-tenant.sh "$tenant_name" "$tenant_type" "$admin_email"; then
        echo "✅ $tenant_name 创建成功"
        
        # 更新公司名称
        mysql "tenant_${tenant_name}" -e "
        UPDATE system_configs 
        SET value='$company_name' 
        WHERE category='system' AND key='company_name'"
        
    else
        echo "❌ $tenant_name 创建失败"
    fi
    
    echo "---"
done

echo "🎉 批量创建完成！"
```

### CSV文件示例
```csv
# tenants.csv
tenant_name,tenant_type,admin_email,company_name
acme_corp,enterprise,admin@acme.com,Acme Corporation
startup_xyz,startup,ceo@startupxyz.com,Startup XYZ
trial_user123,trial,test@example.com,Trial User
tech_solutions,enterprise,it@techsolutions.com,Tech Solutions Ltd
```

## 🔧 故障排除

### 问题1：初始化部分失败
```bash
# 查看初始化状态
db-migrator status -d tenant_failed_init

# 重新执行特定迁移
db-migrator up -d tenant_failed_init --target=002

# 清除失败数据重新初始化
db-migrator down -d tenant_failed_init --steps=2
db-migrator up -d tenant_failed_init
```

### 问题2：数据模板更新
```bash
# 更新模板数据库
db-migrator up -d tenant_template

# 为现有租户应用新的数据更新
db-migrator copy-data \
  --source=tenant_template \
  --patterns=tenant_* \
  --tables=system_configs,business_templates \
  --strategy=merge \
  --conditions="system_configs:updated_at >= '2024-01-01'"
```

### 问题3：大量租户性能问题
```bash
# 并行创建租户（限制并发数）
echo "tenant1 enterprise admin1@example.com
tenant2 startup admin2@example.com  
tenant3 trial admin3@example.com" | \
xargs -n 3 -P 3 ./create-new-tenant.sh

# 使用更大的批量大小
db-migrator init-data \
  --patterns=tenant_batch_* \
  --from-db=tenant_template \
  --batch-size=2000
```

## 🎯 最佳实践

1. **模板维护**：定期更新租户模板数据库
2. **类型分离**：根据租户类型提供不同的功能和限制
3. **自动化流程**：使用脚本自动化租户创建流程
4. **监控告警**：监控租户创建成功率
5. **数据验证**：验证初始化数据的完整性
6. **回滚准备**：准备租户创建失败的回滚方案
7. **性能优化**：大批量创建时使用并行处理
8. **安全考虑**：确保租户间数据隔离

这个案例展示了如何为SaaS平台高效地初始化新租户数据，确保每个租户都能快速开始使用系统！ 