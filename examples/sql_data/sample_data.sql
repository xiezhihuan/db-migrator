-- 示例商店数据
-- 用于测试 insert-data 命令

-- ================================
-- 基础数据插入
-- ================================

-- 商品分类数据
INSERT INTO `categories` (`id`, `parent_id`, `name`, `description`, `sort_order`, `status`) VALUES
(1, 0, '电子产品', '各类电子设备和配件', 1, 1),
(2, 0, '服装鞋帽', '男女装、童装、鞋类', 2, 1),
(3, 0, '家居用品', '家具、装饰、生活用品', 3, 1),
(4, 1, '手机数码', '智能手机、数码相机等', 1, 1),
(5, 1, '电脑办公', '笔记本、台式机、办公用品', 2, 1),
(6, 2, '男装', '男士服装', 1, 1),
(7, 2, '女装', '女士服装', 2, 1);

-- 商品数据
INSERT INTO `products` (`id`, `category_id`, `name`, `description`, `sku`, `price`, `original_price`, `stock`, `status`, `sort_order`) VALUES
(1, 4, 'iPhone 15 Pro', '苹果最新旗舰手机', 'IP15P-256-BK', 8999.00, 9999.00, 50, 1, 1),
(2, 4, 'Samsung Galaxy S24', '三星旗舰智能手机', 'SGS24-128-WH', 6999.00, 7999.00, 30, 1, 2),
(3, 5, 'MacBook Pro 14寸', '苹果专业级笔记本', 'MBP14-512-SG', 15999.00, 17999.00, 20, 1, 1),
(4, 5, 'Dell XPS 13', '戴尔超薄笔记本', 'DXPS13-256-SL', 9999.00, 11999.00, 25, 1, 2),
(5, 6, '男士休闲T恤', '100%纯棉舒适T恤', 'MT001-L-BL', 89.00, 129.00, 100, 1, 1),
(6, 7, '女士连衣裙', '优雅修身连衣裙', 'WD001-M-RD', 299.00, 399.00, 80, 1, 1);

-- 用户数据
INSERT INTO `users` (`id`, `username`, `email`, `password_hash`, `phone`, `status`) VALUES
(1, 'admin', 'admin@example.com', 'hash_admin_123', '13800138000', 1),
(2, 'user1', 'user1@example.com', 'hash_user1_456', '13800138001', 1),
(3, 'user2', 'user2@example.com', 'hash_user2_789', '13800138002', 1),
(4, 'testuser', 'test@example.com', 'hash_test_000', '13800138003', 1);

-- 用户地址数据
INSERT INTO `user_addresses` (`id`, `user_id`, `name`, `phone`, `province`, `city`, `district`, `address`, `is_default`) VALUES
(1, 2, '张三', '13800138001', '北京市', '北京市', '朝阳区', '朝阳区三里屯路123号', 1),
(2, 2, '张三', '13800138001', '上海市', '上海市', '黄浦区', '黄浦区南京路456号', 0),
(3, 3, '李四', '13800138002', '广东省', '深圳市', '南山区', '南山区科技园路789号', 1),
(4, 4, '王五', '13800138003', '浙江省', '杭州市', '西湖区', '西湖区文三路101号', 1);

-- 订单数据
INSERT INTO `orders` (`id`, `order_no`, `user_id`, `status`, `total_amount`, `payment_amount`, `shipping_name`, `shipping_phone`, `shipping_address`, `payment_status`) VALUES
(1, 'ORD20240101001', 2, 4, 9088.00, 9088.00, '张三', '13800138001', '北京市朝阳区三里屯路123号', 1),
(2, 'ORD20240101002', 3, 2, 15999.00, 15999.00, '李四', '13800138002', '广东省深圳市南山区科技园路789号', 1),
(3, 'ORD20240101003', 4, 1, 388.00, 388.00, '王五', '13800138003', '浙江省杭州市西湖区文三路101号', 0);

-- 订单商品数据
INSERT INTO `order_items` (`id`, `order_id`, `product_id`, `product_name`, `product_sku`, `price`, `quantity`, `total_amount`) VALUES
(1, 1, 1, 'iPhone 15 Pro', 'IP15P-256-BK', 8999.00, 1, 8999.00),
(2, 1, 5, '男士休闲T恤', 'MT001-L-BL', 89.00, 1, 89.00),
(3, 2, 3, 'MacBook Pro 14寸', 'MBP14-512-SG', 15999.00, 1, 15999.00),
(4, 3, 5, '男士休闲T恤', 'MT001-L-BL', 89.00, 2, 178.00),
(5, 3, 6, '女士连衣裙', 'WD001-M-RD', 299.00, 1, 299.00);

-- 购物车数据
INSERT INTO `cart_items` (`id`, `user_id`, `product_id`, `quantity`) VALUES
(1, 2, 2, 1),
(2, 2, 4, 1),
(3, 3, 1, 2),
(4, 4, 6, 1),
(5, 4, 5, 3);

-- 系统设置数据
INSERT INTO `settings` (`id`, `key`, `value`, `description`, `group`, `type`) VALUES
(1, 'site_name', 'Demo Shop', '网站名称', 'basic', 'string'),
(2, 'site_description', '演示商店系统', '网站描述', 'basic', 'string'),
(3, 'currency', 'CNY', '货币单位', 'basic', 'string'),
(4, 'tax_rate', '0.13', '税率', 'payment', 'float'),
(5, 'shipping_fee', '10.00', '基础运费', 'shipping', 'float'),
(6, 'free_shipping_threshold', '99.00', '免运费门槛', 'shipping', 'float'),
(7, 'max_cart_items', '50', '购物车最大商品数', 'cart', 'int'),
(8, 'enable_reviews', 'true', '启用商品评价', 'features', 'bool'),
(9, 'enable_notifications', 'true', '启用通知功能', 'features', 'bool'),
(10, 'maintenance_mode', 'false', '维护模式', 'system', 'bool');

-- ================================
-- 多值插入示例
-- ================================

-- 批量插入更多商品（测试批量插入功能）
INSERT INTO `products` (`category_id`, `name`, `sku`, `price`, `stock`, `status`, `sort_order`) VALUES
(4, 'Xiaomi 14', 'XM14-128-BK', 3999.00, 60, 1, 3),
(4, 'OnePlus 12', 'OP12-256-BL', 4999.00, 40, 1, 4),
(4, 'Google Pixel 8', 'GP8-128-WH', 4299.00, 35, 1, 5),
(5, 'Lenovo ThinkPad X1', 'TPX1-512-BK', 12999.00, 15, 1, 3),
(5, 'HP Spectre x360', 'HSX360-256-SL', 8999.00, 20, 1, 4),
(6, '男士商务衬衫', 'MS001-L-WH', 199.00, 120, 1, 2),
(6, '男士牛仔裤', 'MJ001-32-BL', 299.00, 90, 1, 3),
(7, '女士羊毛大衣', 'WC001-M-GR', 899.00, 45, 1, 2),
(7, '女士丝巾', 'WS001-O-MU', 159.00, 200, 1, 3),
(3, '北欧风台灯', 'NL001-WH', 299.00, 80, 1, 1);

-- ================================
-- 注释测试
-- ================================

/*
 * 多行注释测试
 * 这些INSERT语句不应该被执行
 * INSERT INTO products (name) VALUES ('Should not insert');
 */

-- 单行注释测试
-- INSERT INTO users (username) VALUES ('Should not insert');

-- 正常的INSERT语句
INSERT INTO `settings` (`key`, `value`, `description`, `group`) VALUES
('test_setting', 'test_value', '测试设置项', 'test');

-- ================================
-- 特殊值测试
-- ================================

-- 包含NULL值、特殊字符的数据
INSERT INTO `users` (`username`, `email`, `password_hash`, `phone`, `avatar`, `status`) VALUES
('special_user', 'special@test.com', 'hash_special', NULL, NULL, 1),
('user_with_quote', 'quote@test.com', 'password''with''quote', '13800138999', NULL, 1);

-- 包含转义字符的数据
INSERT INTO `settings` (`key`, `value`, `description`) VALUES
('json_config', '{"name": "test", "value": true}', '包含JSON的配置'),
('path_config', 'C:\\Program Files\\App', '包含反斜杠的路径'),
('quote_text', 'He said "Hello World!"', '包含引号的文本'); 