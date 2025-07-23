# 案例1：总部到分店数据复制

## 📋 业务场景

你运营一个连锁店系统，总部负责维护商品目录、价格策略等核心数据，需要定期将这些数据同步到各个分店。

### 系统架构
```
总部数据库 (headquarters_db)
├── 商品目录 (products)
├── 商品分类 (categories)  
├── 供应商信息 (suppliers)
├── 价格策略 (pricing_rules)
└── 促销活动 (promotions)

分店数据库 (shop_*)
├── shop_001_db (北京旗舰店)
├── shop_002_db (上海分店)
├── shop_003_db (广州分店)
└── shop_004_db (深圳分店)
```

### 数据同步需求
- **商品目录**：新品上架，产品信息更新
- **价格策略**：统一定价，促销价格
- **分类管理**：商品分类结构调整
- **供应商信息**：供应商资料更新

## ⚙️ 复制配置

### 基础复制配置
```json
{
  "strategy": "merge",
  "scope": "full",
  "tables": ["products", "categories", "suppliers", "pricing_rules"],
  "batch_size": 1000,
  "timeout": "30m",
  "on_error": "continue"
}
```

### 高级复制配置
```json
{
  "strategy": "merge",
  "scope": "condition",
  "tables": ["products", "categories", "suppliers", "pricing_rules", "promotions"],
  "conditions": {
    "products": "status = 'active' AND updated_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)",
    "promotions": "start_date <= NOW() AND end_date >= NOW()"
  },
  "field_mappings": {
    "products": [
      {
        "source_field": "headquarters_price",
        "target_field": "base_price"
      },
      {
        "source_field": "created_at",
        "target_field": "sync_time",
        "transform": "NOW()"
      }
    ]
  },
  "batch_size": 500,
  "timeout": "45m",
  "on_error": "stop"
}
```

## 📝 迁移文件示例

### 数据初始化迁移
```go
// migrations/data_init/001_sync_headquarters_data.go
package data_init

import (
    "context"
    "db-migrator/internal/builder"
    "db-migrator/internal/checker"
    "db-migrator/internal/types"
)

type SyncHeadquartersDataMigration struct{}

func (m *SyncHeadquartersDataMigration) Version() string {
    return "001"
}

func (m *SyncHeadquartersDataMigration) Description() string {
    return "从总部同步基础数据到店铺"
}

// 只应用到店铺数据库
func (m *SyncHeadquartersDataMigration) Databases() []string {
    return []string{} // 通过命令行 --patterns=shop_* 指定
}

func (m *SyncHeadquartersDataMigration) Up(ctx context.Context, db types.DB) error {
    checker := checker.NewMySQLChecker(db, "")
    dataBuilder := builder.NewDataBuilder(checker, db)

    // 示例：插入基础分类数据
    categories := []map[string]interface{}{
        {
            "id":          1,
            "name":        "电子产品",
            "code":        "electronics",
            "parent_id":   nil,
            "sort_order":  1,
            "is_active":   true,
            "description": "电子产品分类",
        },
        {
            "id":          2,
            "name":        "服装",
            "code":        "clothing",
            "parent_id":   nil,
            "sort_order":  2,
            "is_active":   true,
            "description": "服装分类",
        },
        {
            "id":          3,
            "name":        "家居用品",
            "code":        "home",
            "parent_id":   nil,
            "sort_order":  3,
            "is_active":   true,
            "description": "家居用品分类",
        },
    }

    // 使用智能插入策略
    err := dataBuilder.Table("categories").
        Strategy(builder.StrategyInsertOrUpdate).
        InsertData(ctx, categories)
    if err != nil {
        return err
    }

    // 插入基础配置数据
    configs := []map[string]interface{}{
        {
            "key":         "shop_currency",
            "value":       "CNY",
            "type":        "string",
            "description": "店铺货币单位",
        },
        {
            "key":         "tax_rate",
            "value":       "0.06",
            "type":        "decimal",
            "description": "税率",
        },
        {
            "key":         "max_discount",
            "value":       "0.50",
            "type":        "decimal",
            "description": "最大折扣率",
        },
    }

    return dataBuilder.Table("shop_configs").
        Strategy(builder.StrategyInsertOrUpdate).
        InsertData(ctx, configs)
}

func (m *SyncHeadquartersDataMigration) Down(ctx context.Context, db types.DB) error {
    // 通常数据初始化迁移不需要回滚
    return nil
}
```

### JSON数据文件示例
```json
// data/shop-base-products.json
[
  {
    "id": 1001,
    "sku": "ELEC-001",
    "name": "智能手机",
    "category_id": 1,
    "supplier_id": 101,
    "cost_price": 2000.00,
    "retail_price": 2999.00,
    "status": "active",
    "description": "高性能智能手机",
    "specifications": {
      "brand": "TechBrand",
      "model": "X1",
      "color": "黑色",
      "storage": "128GB"
    }
  },
  {
    "id": 1002,
    "sku": "CLOTH-001", 
    "name": "商务衬衫",
    "category_id": 2,
    "supplier_id": 102,
    "cost_price": 80.00,
    "retail_price": 199.00,
    "status": "active",
    "description": "经典商务衬衫",
    "specifications": {
      "material": "纯棉",
      "size": "L",
      "color": "白色"
    }
  }
]
```

## 💻 操作命令

### 1. 完整数据复制
```bash
# 从总部复制所有商品数据到所有店铺
db-migrator copy-data \
  --source=headquarters \
  --patterns=shop_* \
  --tables=products,categories,suppliers,pricing_rules \
  --strategy=merge \
  --batch-size=1000
```

### 2. 增量数据同步
```bash
# 只复制最近更新的商品
db-migrator copy-data \
  --source=headquarters \
  --patterns=shop_* \
  --tables=products \
  --conditions="products:updated_at >= DATE_SUB(NOW(), INTERVAL 1 DAY)" \
  --strategy=merge
```

### 3. 使用配置文件复制
```bash
# 使用预定义的复制配置
db-migrator copy-data --config=headquarters-sync-config.json
```

### 4. 新店初始化
```bash
# 为新开店铺初始化基础数据
db-migrator init-data \
  -d shop_new_005 \
  --from-db=headquarters \
  --tables=categories,suppliers,pricing_rules

# 从JSON文件初始化示例商品
db-migrator init-data \
  -d shop_new_005 \
  --data-file=data/shop-base-products.json
```

### 5. 批量店铺同步
```bash
# 为所有店铺执行基础数据迁移
db-migrator up --patterns=shop_* --directory=data_init

# 为特定店铺复制促销数据  
db-migrator copy-data \
  --source=headquarters \
  --databases=shop_001,shop_002 \
  --tables=promotions \
  --conditions="promotions:region IN ('北京', '上海')"
```

## 📊 高级场景

### 场景1：分区域商品同步
```bash
# 为北方区域店铺同步商品
db-migrator copy-data \
  --source=headquarters \
  --databases=shop_beijing,shop_tianjin,shop_shenyang \
  --tables=products \
  --conditions="products:region = 'north' OR region IS NULL" \
  --strategy=merge

# 为南方区域店铺同步商品
db-migrator copy-data \
  --source=headquarters \
  --databases=shop_guangzhou,shop_shenzhen,shop_xiamen \
  --tables=products \
  --conditions="products:region = 'south' OR region IS NULL" \
  --strategy=merge
```

### 场景2：价格策略同步
```bash
# 同步VIP客户价格策略
db-migrator copy-data \
  --source=headquarters \
  --patterns=shop_* \
  --tables=pricing_rules \
  --conditions="pricing_rules:customer_type = 'vip'" \
  --strategy=overwrite
```

### 场景3：季节性商品管理
```bash
# 复制春季商品到所有店铺
db-migrator copy-data \
  --source=headquarters \
  --patterns=shop_* \
  --tables=products \
  --conditions="products:season = 'spring' AND status = 'active'" \
  --mappings="products:season_price=retail_price"
```

## 🔄 定期同步脚本

### 自动化同步脚本
```bash
#!/bin/bash
# sync-headquarters-data.sh

# 设置错误时退出
set -e

echo "🔄 开始总部数据同步..."

# 1. 同步商品分类（每天一次）
echo "📂 同步商品分类..."
db-migrator copy-data \
  --source=headquarters \
  --patterns=shop_* \
  --tables=categories \
  --strategy=merge \
  --batch-size=100

# 2. 同步商品信息（增量）
echo "📦 同步商品信息..."
db-migrator copy-data \
  --source=headquarters \
  --patterns=shop_* \
  --tables=products \
  --conditions="products:updated_at >= DATE_SUB(NOW(), INTERVAL 1 DAY)" \
  --strategy=merge \
  --batch-size=500

# 3. 同步价格策略
echo "💰 同步价格策略..."
db-migrator copy-data \
  --source=headquarters \
  --patterns=shop_* \
  --tables=pricing_rules \
  --strategy=merge \
  --batch-size=200

# 4. 同步促销活动
echo "🎉 同步促销活动..."
db-migrator copy-data \
  --source=headquarters \
  --patterns=shop_* \
  --tables=promotions \
  --conditions="promotions:status = 'active' AND start_date <= NOW() AND end_date >= NOW()" \
  --strategy=merge

echo "✅ 总部数据同步完成！"

# 验证同步结果
echo "🔍 验证同步结果..."
db-migrator status --patterns=shop_*

# 发送通知（可选）
# curl -X POST "https://api.notifications.com/webhook" \
#   -d "message=总部数据同步完成"
```

### Cron定时任务
```bash
# 编辑crontab
crontab -e

# 每天凌晨2点执行数据同步
0 2 * * * /path/to/sync-headquarters-data.sh >> /var/log/headquarters-sync.log 2>&1

# 每小时同步促销活动（工作时间）
0 9-18 * * 1-5 db-migrator copy-data --source=headquarters --patterns=shop_* --tables=promotions --conditions="promotions:updated_at >= DATE_SUB(NOW(), INTERVAL 1 HOUR)" --strategy=merge
```

## 🔧 故障排除

### 问题1：某个店铺同步失败
```bash
# 查看具体错误
db-migrator copy-data \
  --source=headquarters \
  --target=shop_003 \
  --tables=products \
  --strategy=merge \
  --on-error=stop

# 单独重试失败的店铺
db-migrator copy-data \
  --source=headquarters \
  --target=shop_003 \
  --tables=products \
  --strategy=overwrite
```

### 问题2：数据不一致
```bash
# 检查数据差异
mysql -e "
SELECT 
  h.id, h.name as hq_name, s.name as shop_name, 
  h.updated_at as hq_time, s.updated_at as shop_time
FROM headquarters_db.products h
LEFT JOIN shop_001_db.products s ON h.id = s.id
WHERE h.updated_at > s.updated_at OR s.id IS NULL
LIMIT 10"

# 强制重新同步
db-migrator copy-data \
  --source=headquarters \
  --target=shop_001 \
  --tables=products \
  --strategy=overwrite
```

### 问题3：同步性能慢
```bash
# 增加批量大小
db-migrator copy-data \
  --source=headquarters \
  --patterns=shop_* \
  --tables=products \
  --batch-size=2000 \
  --timeout=60m

# 分表同步
db-migrator copy-data --source=headquarters --patterns=shop_* --tables=categories &
db-migrator copy-data --source=headquarters --patterns=shop_* --tables=products &
db-migrator copy-data --source=headquarters --patterns=shop_* --tables=suppliers &
wait
```

## 🎯 最佳实践

1. **分时段同步**：避免在业务高峰期进行大量数据同步
2. **增量同步**：优先使用条件复制，减少数据传输量
3. **监控告警**：设置同步失败的告警通知
4. **数据验证**：定期验证总部和分店数据的一致性
5. **备份保护**：重要同步前进行数据备份
6. **性能优化**：根据网络和数据库性能调整批量大小

这个案例展示了如何高效地从总部向多个分店同步数据，确保业务数据的一致性和及时性！ 