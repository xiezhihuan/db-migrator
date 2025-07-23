# 数据操作案例 - 初始化与复制

本目录包含数据初始化和跨数据库数据复制的具体使用案例。

## 📁 案例分类

### 🔄 数据复制案例
- [**总部到分店**](01_headquarters_to_shops/) - 从总部复制商品目录到各店铺
- [**模板到新实例**](02_template_to_new/) - 从模板数据库复制到新实例
- [**数据同步**](03_data_sync/) - 定期数据同步和更新
- [**跨服务数据共享**](04_cross_service/) - 微服务间数据共享

### 🗃️ 数据初始化案例  
- [**新租户开通**](05_new_tenant/) - SaaS新租户数据初始化
- [**基础数据导入**](06_base_data/) - 系统配置和字典数据
- [**批量店铺初始化**](07_bulk_shops/) - 批量新店铺数据准备
- [**开发环境准备**](08_dev_environment/) - 开发测试数据准备

## 🚀 快速开始

### 数据复制命令

```bash
# 从总部复制商品到所有店铺
db-migrator copy-data --source=headquarters --patterns=shop_* --tables=products,categories

# 复制指定条件的订单数据
db-migrator copy-data --source=main_db --target=archive_db --tables=orders --conditions="orders:created_at < '2023-01-01'"

# 智能合并数据
db-migrator copy-data --source=template_db --target=new_db --strategy=merge --tables=all
```

### 数据初始化命令

```bash
# 从模板数据库初始化新租户
db-migrator init-data -d tenant_new_001 --from-db=tenant_template

# 从JSON文件批量初始化
db-migrator init-data --patterns=shop_* --data-file=shop-base-data.json

# 初始化系统配置数据
db-migrator init-data --patterns=*_service --data-type=system_configs
```

## 📊 支持的功能

### 复制策略
- **overwrite** - 完全覆盖（清空后插入）
- **merge** - 智能合并（插入或更新）
- **insert** - 仅插入新数据  
- **ignore** - 忽略重复数据

### 复制范围
- **full** - 整表复制
- **condition** - 条件复制（WHERE子句）
- **mapping** - 字段映射复制
- **transform** - 数据转换复制

### 数据源
- **数据库复制** - 从其他数据库复制
- **JSON文件** - 从JSON文件导入
- **YAML文件** - 从YAML文件导入
- **内置数据** - 预定义的基础数据

## 💡 最佳实践

### 1. 数据复制
- 使用事务保护确保数据一致性
- 设置合适的批量大小提高性能
- 使用条件复制减少不必要的数据传输
- 为大批量操作设置超时时间

### 2. 数据初始化
- 先初始化基础配置数据
- 再初始化业务数据
- 使用模板数据库提高效率
- 验证数据完整性

### 3. 错误处理
- 设置适当的错误处理策略
- 记录详细的操作日志
- 提供进度显示和状态反馈
- 准备回滚方案

### 4. 性能优化
- 合理设置批量大小
- 使用索引优化查询性能
- 避免在高峰期进行大量数据操作
- 监控系统资源使用情况

选择适合你业务场景的案例开始数据操作吧！ 🚀 