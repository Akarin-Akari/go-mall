-- 清理初始化数据

DELETE FROM `products` WHERE `name` IN ('iPhone 15 Pro', 'MacBook Pro 14', 'AirPods Pro 2');
DELETE FROM `categories` WHERE `slug` IN ('mobile', 'computer', 'accessories');
DELETE FROM `categories` WHERE `slug` IN ('electronics', 'fashion', 'food', 'home', 'books');
DELETE FROM `users` WHERE `username` = 'admin';

