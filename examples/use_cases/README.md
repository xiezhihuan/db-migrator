# 多数据库迁移 - 实际使用案例

本目录包含了各种真实业务场景的多数据库迁移案例，帮助你根据自己的项目需求快速上手。

## 📁 案例分类

### 🛒 电商相关
- [**多店铺系统**](01_multi_shop_system/) - 连锁店、加盟店管理
- [**跨境电商**](02_cross_border_ecommerce/) - 不同国家/地区数据库

### 🏢 企业应用
- [**SaaS多租户**](03_saas_multi_tenant/) - 每个客户独立数据库
- [**企业多分公司**](04_enterprise_branches/) - 总部+分公司架构
- [**集团公司**](05_conglomerate/) - 不同业务线数据库

### 🔧 技术架构
- [**微服务架构**](06_microservices/) - 服务拆分数据库
- [**分库分表**](07_database_sharding/) - 大数据量分库
- [**多环境部署**](08_multi_environment/) - 开发/测试/生产

### 🎮 特殊场景
- [**游戏多服务器**](09_game_servers/) - 游戏分服数据库
- [**IoT设备管理**](10_iot_devices/) - 按地区/类型分库
- [**教育系统**](11_education_system/) - 多校区管理

## 🚀 快速选择指南

| 你的项目类型 | 推荐案例 | 适用模式 |
|-------------|---------|---------|
| 连锁商店管理 | [多店铺系统](01_multi_shop_system/) | `shop_*` |
| SaaS产品 | [多租户系统](03_saas_multi_tenant/) | `tenant_*` |
| 微服务项目 | [微服务架构](06_microservices/) | `service_*` |
| 大型企业 | [企业多分公司](04_enterprise_branches/) | `branch_*` |
| 游戏开发 | [游戏多服务器](09_game_servers/) | `server_*` |
| 物联网 | [IoT设备管理](10_iot_devices/) | `region_*` |

## 📖 如何使用这些案例

1. **找到匹配的场景**：根据你的业务需求选择最相似的案例
2. **复制配置文件**：使用案例中的 `config.yaml` 作为模板
3. **参考迁移文件**：学习如何组织和编写迁移
4. **运行示例命令**：按照案例中的命令进行操作
5. **根据需要调整**：修改数据库名、表结构等细节

## 💡 通用最佳实践

### 数据库命名规范
```bash
# 按业务+环境
shop_main_prod, shop_main_dev, shop_main_test

# 按地区+业务  
shop_us_main, shop_eu_main, shop_asia_main

# 按编号+类型
tenant_001_db, tenant_002_db, tenant_003_db

# 按服务+环境
user_service_prod, order_service_prod, payment_service_prod
```

### 配置组织建议
```yaml
# 生产环境
databases:
  main_prod:
    database: company_main_prod
  shop_001_prod:
    database: shop_001_prod
  shop_002_prod:
    database: shop_002_prod

# 开发环境  
databases:
  main_dev:
    database: company_main_dev
  shop_001_dev:
    database: shop_001_dev
```

### 迁移组织方式
```
migrations/
├── core/              # 核心业务表（所有数据库）
├── shop_specific/     # 店铺特有功能
├── tenant_specific/   # 租户特有功能
└── maintenance/       # 维护相关迁移
```

每个案例都包含完整的：
- 📋 **业务背景说明**
- ⚙️ **完整配置文件**  
- 🗂️ **迁移文件示例**
- 💻 **具体操作命令**
- 🔧 **故障排除指南**

选择一个案例开始你的多数据库迁移之旅吧！🚀 