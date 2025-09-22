-- 初始化商品数据
-- 清理现有数据
DELETE FROM products;
DELETE FROM categories;
DELETE FROM brands;

-- 插入分类数据
INSERT OR REPLACE INTO categories (id, name, description, parent_id, sort, status, created_at, updated_at) VALUES
(1, '电子产品', '各种电子设备和数码产品', 0, 1, 'active', datetime('now'), datetime('now')),
(2, '服装鞋帽', '时尚服装和配饰', 0, 2, 'active', datetime('now'), datetime('now')),
(3, '家居用品', '家庭生活用品', 0, 3, 'active', datetime('now'), datetime('now')),
(4, '图书文具', '书籍和办公用品', 0, 4, 'active', datetime('now'), datetime('now'));

-- 插入品牌数据
INSERT OR REPLACE INTO brands (id, name, description, status, sort, created_at, updated_at) VALUES
(1, 'Apple', '苹果公司产品', 'active', 1, datetime('now'), datetime('now')),
(2, 'Nike', '耐克运动品牌', 'active', 2, datetime('now'), datetime('now')),
(3, 'IKEA', '宜家家居', 'active', 3, datetime('now'), datetime('now')),
(4, '小米', '小米科技产品', 'active', 4, datetime('now'), datetime('now'));

-- 插入商品数据
INSERT OR REPLACE INTO products (id, name, description, price, stock, status, category_id, brand_id, images, attributes, weight, sold_count, view_count, created_at, updated_at) VALUES
(1, 'iPhone 15 Pro', '苹果最新款智能手机，配备A17 Pro芯片', 8999.00, 100, 'active', 1, 1, 
 '["https://example.com/iphone15pro.jpg"]', 
 '{"color": "深空黑", "storage": "256GB", "screen": "6.1英寸"}', 
 0.187, 0, 0, datetime('now'), datetime('now')),

(2, 'MacBook Pro 14', '专业级笔记本电脑，搭载M3芯片', 14999.00, 50, 'active', 1, 1,
 '["https://example.com/macbookpro14.jpg"]',
 '{"color": "深空灰", "memory": "16GB", "storage": "512GB SSD"}',
 1.6, 0, 0, datetime('now'), datetime('now')),

(3, 'Nike Air Max 270', '经典运动鞋，舒适透气', 899.00, 200, 'active', 2, 2,
 '["https://example.com/airmax270.jpg"]',
 '{"color": "黑白", "size": "42", "material": "网布+合成革"}',
 0.8, 0, 0, datetime('now'), datetime('now')),

(4, 'IKEA BILLY书架', '简约现代书架，多种颜色可选', 199.00, 150, 'active', 3, 3,
 '["https://example.com/billy-bookshelf.jpg"]',
 '{"color": "白色", "material": "刨花板", "dimensions": "80x28x202cm"}',
 25.5, 0, 0, datetime('now'), datetime('now')),

(5, 'AirPods Pro 2', '主动降噪无线耳机', 1899.00, 80, 'active', 1, 1,
 '["https://example.com/airpods-pro2.jpg"]',
 '{"color": "白色", "battery": "6小时+24小时", "features": "主动降噪"}',
 0.056, 0, 0, datetime('now'), datetime('now')),

(6, '小米13 Pro', '小米旗舰手机，徕卡影像', 3999.00, 120, 'active', 1, 4,
 '["https://example.com/mi13pro.jpg"]',
 '{"color": "陶瓷白", "storage": "256GB", "camera": "徕卡三摄"}',
 0.210, 0, 0, datetime('now'), datetime('now')),

(7, 'Nike Dunk Low', '经典板鞋，街头时尚', 699.00, 180, 'active', 2, 2,
 '["https://example.com/dunk-low.jpg"]',
 '{"color": "熊猫配色", "size": "41", "material": "真皮"}',
 0.9, 0, 0, datetime('now'), datetime('now')),

(8, 'IKEA POÄNG扶手椅', '舒适休闲椅，北欧设计', 599.00, 60, 'active', 3, 3,
 '["https://example.com/poang-chair.jpg"]',
 '{"color": "桦木色", "material": "桦木+棉质", "dimensions": "68x82x100cm"}',
 12.0, 0, 0, datetime('now'), datetime('now')),

(9, '编程珠玑', '经典编程思维训练书籍', 89.00, 200, 'active', 4, 4,
 '["https://example.com/programming-pearls.jpg"]',
 '{"author": "Jon Bentley", "pages": "256", "language": "中文"}',
 0.3, 0, 0, datetime('now'), datetime('now')),

(10, '无线充电器', '支持快充的无线充电板', 199.00, 300, 'active', 1, 4,
 '["https://example.com/wireless-charger.jpg"]',
 '{"power": "15W", "compatibility": "iPhone/Android", "color": "黑色"}',
 0.2, 0, 0, datetime('now'), datetime('now'));

-- 验证数据插入
SELECT 'Categories:' as table_name, count(*) as count FROM categories
UNION ALL
SELECT 'Brands:', count(*) FROM brands  
UNION ALL
SELECT 'Products:', count(*) FROM products;
