-- Mall-Go 数据模型优化 - 添加冗余字段和索引
-- 执行日期: 2025-01-10
-- 目标: 减少JOIN查询，提升查询性能

-- 1. 为products表添加冗余字段
ALTER TABLE products 
ADD COLUMN category_name VARCHAR(100) DEFAULT '' COMMENT '分类名称冗余字段',
ADD COLUMN brand_name VARCHAR(100) DEFAULT '' COMMENT '品牌名称冗余字段', 
ADD COLUMN merchant_name VARCHAR(100) DEFAULT '' COMMENT '商家名称冗余字段';

-- 2. 创建性能优化索引
-- 商品搜索复合索引
CREATE INDEX idx_products_category_brand ON products(category_name, brand_name);
CREATE INDEX idx_products_price_status ON products(price, status);
CREATE INDEX idx_products_merchant_status ON products(merchant_id, status);
CREATE INDEX idx_products_stock_status ON products(stock, status);
CREATE INDEX idx_products_sold_count_desc ON products(sold_count DESC);

-- 商品筛选索引
CREATE INDEX idx_products_category_price ON products(category_id, price);
CREATE INDEX idx_products_brand_price ON products(brand_id, price);
CREATE INDEX idx_products_status_hot ON products(status, is_hot);
CREATE INDEX idx_products_status_new ON products(status, is_new);
CREATE INDEX idx_products_status_recommend ON products(status, is_recommend);

-- 商品排序索引
CREATE INDEX idx_products_sort_id ON products(sort ASC, id DESC);
CREATE INDEX idx_products_created_desc ON products(created_at DESC);
CREATE INDEX idx_products_view_count_desc ON products(view_count DESC);

-- 3. 购物车相关索引优化
CREATE INDEX idx_cart_items_user_product ON cart_items(user_id, product_id);
CREATE INDEX idx_cart_items_session_product ON cart_items(session_id, product_id);

-- 4. 订单相关索引优化  
CREATE INDEX idx_orders_user_status_created ON orders(user_id, status, created_at DESC);
CREATE INDEX idx_orders_merchant_status ON orders(merchant_id, status);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);

-- 5. 商品图片索引优化
CREATE INDEX idx_product_images_product_main ON product_images(product_id, is_main);
CREATE INDEX idx_product_images_sort ON product_images(product_id, sort ASC);

-- 6. 商品属性索引优化
CREATE INDEX idx_product_attrs_product_name ON product_attrs(product_id, attr_name);

-- 7. 商品SKU索引优化
CREATE INDEX idx_product_skus_product_status ON product_skus(product_id, status);
CREATE INDEX idx_product_skus_code ON product_skus(sku_code);

-- 8. 分类索引优化
CREATE INDEX idx_categories_parent_status ON categories(parent_id, status);
CREATE INDEX idx_categories_level_sort ON categories(level, sort ASC);

-- 9. 品牌索引优化
CREATE INDEX idx_brands_status_sort ON brands(status, sort ASC);

-- 10. 用户索引优化
CREATE INDEX idx_users_role_status ON users(role, status);
CREATE INDEX idx_users_created_desc ON users(created_at DESC);

-- 11. 更新现有数据的冗余字段
-- 更新商品的分类名称冗余字段
UPDATE products p 
SET category_name = (
    SELECT c.name 
    FROM categories c 
    WHERE c.id = p.category_id
)
WHERE p.category_id IS NOT NULL;

-- 更新商品的品牌名称冗余字段
UPDATE products p 
SET brand_name = (
    SELECT b.name 
    FROM brands b 
    WHERE b.id = p.brand_id
)
WHERE p.brand_id IS NOT NULL AND p.brand_id > 0;

-- 更新商品的商家名称冗余字段
UPDATE products p 
SET merchant_name = (
    SELECT u.username 
    FROM users u 
    WHERE u.id = p.merchant_id
)
WHERE p.merchant_id IS NOT NULL;

-- 12. 修改Version字段默认值为1（乐观锁优化）
ALTER TABLE products MODIFY COLUMN version INT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号';
ALTER TABLE product_skus MODIFY COLUMN version INT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号';

-- 13. 创建数据同步触发器（保证冗余字段一致性）
DELIMITER $$

-- 分类名称更新触发器
CREATE TRIGGER tr_categories_update_product_name
AFTER UPDATE ON categories
FOR EACH ROW
BEGIN
    IF OLD.name != NEW.name THEN
        UPDATE products SET category_name = NEW.name WHERE category_id = NEW.id;
    END IF;
END$$

-- 品牌名称更新触发器  
CREATE TRIGGER tr_brands_update_product_name
AFTER UPDATE ON brands
FOR EACH ROW
BEGIN
    IF OLD.name != NEW.name THEN
        UPDATE products SET brand_name = NEW.name WHERE brand_id = NEW.id;
    END IF;
END$$

-- 用户名称更新触发器
CREATE TRIGGER tr_users_update_product_name
AFTER UPDATE ON users
FOR EACH ROW
BEGIN
    IF OLD.username != NEW.username THEN
        UPDATE products SET merchant_name = NEW.username WHERE merchant_id = NEW.id;
    END IF;
END$$

DELIMITER ;

-- 14. 验证索引创建结果
SHOW INDEX FROM products;
SHOW INDEX FROM cart_items;
SHOW INDEX FROM orders;

-- 15. 分析表统计信息（优化查询计划）
ANALYZE TABLE products;
ANALYZE TABLE categories;
ANALYZE TABLE brands;
ANALYZE TABLE users;
ANALYZE TABLE cart_items;
ANALYZE TABLE orders;

-- 迁移完成提示
SELECT 'Mall-Go数据模型优化迁移完成！' AS message,
       '已添加冗余字段和性能索引' AS description,
       NOW() AS completed_at;
