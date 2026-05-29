-- 初始化基础数据

-- 插入管理员用户（密码: admin123，需要在应用层使用 bcrypt 加密）
-- 注意：这里使用明文密码仅用于示例，生产环境应该使用加密后的密码
INSERT INTO `users` (`username`, `email`, `password`, `nickname`, `role`, `status`, `email_verified`) 
VALUES ('admin', 'admin@mall.com', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z5EHsM8lE9lBOsl7iAt6Z5EH', '系统管理员', 'admin', 'active', TRUE)
ON DUPLICATE KEY UPDATE `username` = `username`;

-- 插入默认分类
INSERT INTO `categories` (`name`, `slug`, `parent_id`, `level`, `sort_order`, `status`) VALUES
('电子产品', 'electronics', 0, 1, 1, 'active'),
('服装鞋包', 'fashion', 0, 1, 2, 'active'),
('食品饮料', 'food', 0, 1, 3, 'active'),
('家居生活', 'home', 0, 1, 4, 'active'),
('图书音像', 'books', 0, 1, 5, 'active')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 插入二级分类（电子产品）
INSERT INTO `categories` (`name`, `slug`, `parent_id`, `level`, `sort_order`, `status`) VALUES
('手机通讯', 'mobile', 1, 2, 1, 'active'),
('电脑办公', 'computer', 1, 2, 2, 'active'),
('数码配件', 'accessories', 1, 2, 3, 'active')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 插入示例商品
INSERT INTO `products` (`name`, `sub_title`, `description`, `category_id`, `category_name`, `price`, `origin_price`, `stock`, `status`, `is_hot`, `is_new`) VALUES
('iPhone 15 Pro', '全新A17 Pro芯片', '苹果最新旗舰手机，性能强劲', 6, '手机通讯', 7999.00, 8999.00, 100, 'active', TRUE, TRUE),
('MacBook Pro 14', 'M3 Pro芯片', '专业级笔记本电脑', 7, '电脑办公', 14999.00, 16999.00, 50, 'active', TRUE, FALSE),
('AirPods Pro 2', '主动降噪', '无线蓝牙耳机', 8, '数码配件', 1899.00, 1999.00, 200, 'active', FALSE, FALSE)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

