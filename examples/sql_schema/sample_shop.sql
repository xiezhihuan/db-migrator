-- 示例商店数据库结构
-- 支持用户、商品、订单、支付等完整功能

-- ================================
-- 用户相关表
-- ================================

-- 用户表
CREATE TABLE `users` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `username` varchar(50) NOT NULL COMMENT '用户名',
    `email` varchar(100) NOT NULL COMMENT '邮箱',
    `password_hash` varchar(255) NOT NULL COMMENT '密码哈希',
    `phone` varchar(20) DEFAULT NULL COMMENT '手机号',
    `avatar` varchar(255) DEFAULT NULL COMMENT '头像URL',
    `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态: 1=正常 2=禁用',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    UNIQUE KEY `uk_email` (`email`),
    KEY `idx_phone` (`phone`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 用户地址表
CREATE TABLE `user_addresses` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '地址ID',
    `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
    `name` varchar(50) NOT NULL COMMENT '收货人姓名',
    `phone` varchar(20) NOT NULL COMMENT '收货人手机',
    `province` varchar(50) NOT NULL COMMENT '省份',
    `city` varchar(50) NOT NULL COMMENT '城市',
    `district` varchar(50) NOT NULL COMMENT '区县',
    `address` varchar(200) NOT NULL COMMENT '详细地址',
    `postal_code` varchar(10) DEFAULT NULL COMMENT '邮政编码',
    `is_default` tinyint NOT NULL DEFAULT '0' COMMENT '是否默认地址: 0=否 1=是',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_is_default` (`is_default`),
    FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户地址表';

-- ================================
-- 商品相关表
-- ================================

-- 商品分类表
CREATE TABLE `categories` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '分类ID',
    `parent_id` bigint unsigned DEFAULT '0' COMMENT '父分类ID，0表示顶级分类',
    `name` varchar(100) NOT NULL COMMENT '分类名称',
    `description` text DEFAULT NULL COMMENT '分类描述',
    `image` varchar(255) DEFAULT NULL COMMENT '分类图片',
    `sort_order` int NOT NULL DEFAULT '0' COMMENT '排序权重',
    `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态: 1=启用 2=禁用',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_parent_id` (`parent_id`),
    KEY `idx_sort_order` (`sort_order`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品分类表';

-- 商品表
CREATE TABLE `products` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '商品ID',
    `category_id` bigint unsigned NOT NULL COMMENT '分类ID',
    `name` varchar(200) NOT NULL COMMENT '商品名称',
    `description` text DEFAULT NULL COMMENT '商品描述',
    `short_description` varchar(500) DEFAULT NULL COMMENT '商品简介',
    `sku` varchar(100) NOT NULL COMMENT '商品SKU',
    `price` decimal(10,2) NOT NULL COMMENT '商品价格',
    `original_price` decimal(10,2) DEFAULT NULL COMMENT '原价',
    `cost_price` decimal(10,2) DEFAULT NULL COMMENT '成本价',
    `stock` int NOT NULL DEFAULT '0' COMMENT '库存数量',
    `min_stock` int NOT NULL DEFAULT '0' COMMENT '最小库存预警',
    `weight` decimal(8,2) DEFAULT NULL COMMENT '重量(kg)',
    `images` json DEFAULT NULL COMMENT '商品图片JSON数组',
    `attributes` json DEFAULT NULL COMMENT '商品属性JSON',
    `tags` varchar(500) DEFAULT NULL COMMENT '商品标签，逗号分隔',
    `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态: 1=上架 2=下架 3=缺货',
    `sort_order` int NOT NULL DEFAULT '0' COMMENT '排序权重',
    `view_count` int NOT NULL DEFAULT '0' COMMENT '浏览次数',
    `sale_count` int NOT NULL DEFAULT '0' COMMENT '销售数量',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_sku` (`sku`),
    KEY `idx_category_id` (`category_id`),
    KEY `idx_price` (`price`),
    KEY `idx_stock` (`stock`),
    KEY `idx_status` (`status`),
    KEY `idx_sort_order` (`sort_order`),
    KEY `idx_created_at` (`created_at`),
    FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品表';

-- ================================
-- 订单相关表
-- ================================

-- 订单表
CREATE TABLE `orders` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '订单ID',
    `order_no` varchar(50) NOT NULL COMMENT '订单号',
    `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
    `status` tinyint NOT NULL DEFAULT '1' COMMENT '订单状态: 1=待付款 2=待发货 3=待收货 4=已完成 5=已取消',
    `total_amount` decimal(10,2) NOT NULL COMMENT '订单总金额',
    `discount_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '优惠金额',
    `shipping_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '运费',
    `payment_amount` decimal(10,2) NOT NULL COMMENT '实付金额',
    `payment_method` varchar(50) DEFAULT NULL COMMENT '支付方式',
    `payment_status` tinyint NOT NULL DEFAULT '0' COMMENT '支付状态: 0=未支付 1=已支付 2=部分退款 3=全额退款',
    `shipping_name` varchar(50) DEFAULT NULL COMMENT '收货人姓名',
    `shipping_phone` varchar(20) DEFAULT NULL COMMENT '收货人手机',
    `shipping_address` varchar(500) DEFAULT NULL COMMENT '收货地址',
    `notes` text DEFAULT NULL COMMENT '订单备注',
    `paid_at` timestamp NULL DEFAULT NULL COMMENT '支付时间',
    `shipped_at` timestamp NULL DEFAULT NULL COMMENT '发货时间',
    `completed_at` timestamp NULL DEFAULT NULL COMMENT '完成时间',
    `cancelled_at` timestamp NULL DEFAULT NULL COMMENT '取消时间',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_no` (`order_no`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_status` (`status`),
    KEY `idx_payment_status` (`payment_status`),
    KEY `idx_created_at` (`created_at`),
    FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';

-- 订单商品表
CREATE TABLE `order_items` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '订单商品ID',
    `order_id` bigint unsigned NOT NULL COMMENT '订单ID',
    `product_id` bigint unsigned NOT NULL COMMENT '商品ID',
    `product_name` varchar(200) NOT NULL COMMENT '商品名称(下单时快照)',
    `product_sku` varchar(100) NOT NULL COMMENT '商品SKU',
    `product_image` varchar(255) DEFAULT NULL COMMENT '商品图片',
    `price` decimal(10,2) NOT NULL COMMENT '商品单价',
    `quantity` int NOT NULL COMMENT '购买数量',
    `total_amount` decimal(10,2) NOT NULL COMMENT '小计金额',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_order_id` (`order_id`),
    KEY `idx_product_id` (`product_id`),
    FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE CASCADE,
    FOREIGN KEY (`product_id`) REFERENCES `products` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单商品表';

-- ================================
-- 购物车表
-- ================================

CREATE TABLE `cart_items` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '购物车ID',
    `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
    `product_id` bigint unsigned NOT NULL COMMENT '商品ID',
    `quantity` int NOT NULL DEFAULT '1' COMMENT '商品数量',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_product` (`user_id`, `product_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_product_id` (`product_id`),
    FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='购物车表';

-- ================================
-- 系统设置表
-- ================================

CREATE TABLE `settings` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '设置ID',
    `key` varchar(100) NOT NULL COMMENT '设置键',
    `value` text DEFAULT NULL COMMENT '设置值',
    `description` varchar(255) DEFAULT NULL COMMENT '设置描述',
    `group` varchar(50) NOT NULL DEFAULT 'default' COMMENT '设置分组',
    `type` varchar(20) NOT NULL DEFAULT 'string' COMMENT '值类型: string, int, float, bool, json',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_key` (`key`),
    KEY `idx_group` (`group`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统设置表';

-- ================================
-- 视图
-- ================================

-- 商品销售统计视图
CREATE VIEW `product_sales_stats` AS
SELECT 
    p.`id`,
    p.`name`,
    p.`sku`,
    p.`price`,
    p.`stock`,
    COALESCE(SUM(oi.`quantity`), 0) AS `total_sold`,
    COALESCE(SUM(oi.`total_amount`), 0) AS `total_revenue`,
    p.`created_at`
FROM `products` p
LEFT JOIN `order_items` oi ON p.`id` = oi.`product_id`
LEFT JOIN `orders` o ON oi.`order_id` = o.`id` AND o.`status` IN (3, 4)
GROUP BY p.`id`;

-- 用户订单统计视图  
CREATE VIEW `user_order_stats` AS
SELECT 
    u.`id`,
    u.`username`,
    u.`email`,
    COUNT(o.`id`) AS `total_orders`,
    COALESCE(SUM(o.`payment_amount`), 0) AS `total_spent`,
    MAX(o.`created_at`) AS `last_order_at`
FROM `users` u
LEFT JOIN `orders` o ON u.`id` = o.`user_id` AND o.`payment_status` = 1
GROUP BY u.`id`;

-- ================================
-- 索引（可选的性能优化索引）
-- ================================

-- 订单表复合索引
CREATE INDEX `idx_orders_user_status_time` ON `orders` (`user_id`, `status`, `created_at`);

-- 商品表复合索引
CREATE INDEX `idx_products_category_status_sort` ON `products` (`category_id`, `status`, `sort_order`);

-- ================================
-- 触发器
-- ================================

-- 更新商品库存触发器
DELIMITER $$

CREATE TRIGGER `trg_order_item_stock_decrease` 
AFTER INSERT ON `order_items`
FOR EACH ROW
BEGIN
    UPDATE `products` 
    SET `stock` = `stock` - NEW.`quantity`,
        `sale_count` = `sale_count` + NEW.`quantity`
    WHERE `id` = NEW.`product_id`;
END$$

CREATE TRIGGER `trg_order_item_stock_increase`
AFTER DELETE ON `order_items`
FOR EACH ROW  
BEGIN
    UPDATE `products`
    SET `stock` = `stock` + OLD.`quantity`,
        `sale_count` = `sale_count` - OLD.`quantity`
    WHERE `id` = OLD.`product_id`;
END$$

DELIMITER ;

-- ================================
-- 存储过程示例
-- ================================

DELIMITER $$

-- 获取用户购物车总价的存储过程
CREATE PROCEDURE `GetUserCartTotal`(IN p_user_id BIGINT)
BEGIN
    SELECT 
        COUNT(*) AS item_count,
        COALESCE(SUM(p.price * ci.quantity), 0) AS total_amount
    FROM cart_items ci
    JOIN products p ON ci.product_id = p.id
    WHERE ci.user_id = p_user_id AND p.status = 1;
END$$

-- 清理过期购物车数据的存储过程
CREATE PROCEDURE `CleanExpiredCartItems`()
BEGIN
    DELETE FROM cart_items 
    WHERE updated_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
    
    SELECT ROW_COUNT() AS deleted_count;
END$$

DELIMITER ; 