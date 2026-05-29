-- 删除性能优化索引

DROP INDEX `idx_cart_items_cart_selected` ON `cart_items`;
DROP INDEX `idx_orders_payment_status` ON `orders`;
DROP INDEX `idx_orders_user_created` ON `orders`;
DROP INDEX `idx_orders_user_status` ON `orders`;
DROP INDEX `idx_products_created_at` ON `products`;
DROP INDEX `idx_products_sold_count_status` ON `products`;
DROP INDEX `idx_products_price_status` ON `products`;
DROP INDEX `idx_products_category_status` ON `products`;
DROP INDEX `idx_users_created_at` ON `users`;
DROP INDEX `idx_users_status_role` ON `users`;

