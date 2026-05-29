-- 添加性能优化索引

-- 用户表复合索引
CREATE INDEX `idx_users_status_role` ON `users` (`status`, `role`);
CREATE INDEX `idx_users_created_at` ON `users` (`created_at`);

-- 商品表复合索引
CREATE INDEX `idx_products_category_status` ON `products` (`category_id`, `status`);
CREATE INDEX `idx_products_price_status` ON `products` (`price`, `status`);
CREATE INDEX `idx_products_sold_count_status` ON `products` (`sold_count` DESC, `status`);
CREATE INDEX `idx_products_created_at` ON `products` (`created_at` DESC);

-- 订单表复合索引
CREATE INDEX `idx_orders_user_status` ON `orders` (`user_id`, `status`);
CREATE INDEX `idx_orders_user_created` ON `orders` (`user_id`, `created_at` DESC);
CREATE INDEX `idx_orders_payment_status` ON `orders` (`payment_status`, `status`);

-- 购物车项复合索引
CREATE INDEX `idx_cart_items_cart_selected` ON `cart_items` (`cart_id`, `selected`);

