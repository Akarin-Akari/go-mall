-- 初始化数据库表结构
-- 注意：不使用外键约束，符合 RuoYi 开发规范

-- 用户表
CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    `email` VARCHAR(100) NOT NULL UNIQUE COMMENT '邮箱',
    `password` VARCHAR(255) NOT NULL COMMENT '密码哈希',
    `nickname` VARCHAR(50) COMMENT '昵称',
    `avatar` VARCHAR(255) COMMENT '头像URL',
    `phone` VARCHAR(20) COMMENT '手机号',
    `role` VARCHAR(20) DEFAULT 'user' COMMENT '角色',
    `status` VARCHAR(20) DEFAULT 'active' COMMENT '状态',
    `email_verified` BOOLEAN DEFAULT FALSE COMMENT '邮箱是否验证',
    `phone_verified` BOOLEAN DEFAULT FALSE COMMENT '手机是否验证',
    `balance` DECIMAL(10,2) DEFAULT 0.00 COMMENT '账户余额',
    `points` INT DEFAULT 0 COMMENT '积分',
    `level` INT DEFAULT 1 COMMENT '用户等级',
    `login_count` INT DEFAULT 0 COMMENT '登录次数',
    `last_login_at` TIMESTAMP NULL COMMENT '最后登录时间',
    `last_login_ip` VARCHAR(45) COMMENT '最后登录IP',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL,
    INDEX `idx_username` (`username`),
    INDEX `idx_email` (`email`),
    INDEX `idx_phone` (`phone`),
    INDEX `idx_status` (`status`),
    INDEX `idx_role` (`role`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 分类表
CREATE TABLE IF NOT EXISTS `categories` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL COMMENT '分类名称',
    `slug` VARCHAR(100) COMMENT 'URL别名',
    `parent_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '父分类ID',
    `level` INT DEFAULT 1 COMMENT '层级',
    `sort_order` INT DEFAULT 0 COMMENT '排序',
    `icon` VARCHAR(255) COMMENT '图标',
    `image` VARCHAR(255) COMMENT '图片',
    `description` TEXT COMMENT '描述',
    `status` VARCHAR(20) DEFAULT 'active' COMMENT '状态',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL,
    INDEX `idx_parent_id` (`parent_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_sort_order` (`sort_order`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分类表';

-- 商品表
CREATE TABLE IF NOT EXISTS `products` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(255) NOT NULL COMMENT '商品名称',
    `sub_title` VARCHAR(500) COMMENT '副标题',
    `description` TEXT COMMENT '简介',
    `detail` LONGTEXT COMMENT '详情',
    `category_id` BIGINT UNSIGNED NOT NULL COMMENT '分类ID',
    `brand_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '品牌ID',
    `merchant_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '商家ID',
    -- 冗余字段（性能优化）
    `category_name` VARCHAR(100) COMMENT '分类名称',
    `brand_name` VARCHAR(100) COMMENT '品牌名称',
    `merchant_name` VARCHAR(100) COMMENT '商家名称',
    -- 价格信息
    `price` DECIMAL(10,2) NOT NULL COMMENT '售价',
    `origin_price` DECIMAL(10,2) COMMENT '原价',
    `cost_price` DECIMAL(10,2) COMMENT '成本价',
    -- 库存信息
    `stock` INT NOT NULL DEFAULT 0 COMMENT '库存',
    `min_stock` INT DEFAULT 0 COMMENT '最小库存',
    `max_stock` INT DEFAULT 0 COMMENT '最大库存',
    `sold_count` INT DEFAULT 0 COMMENT '销量',
    `version` INT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号',
    -- 商品属性
    `weight` DECIMAL(8,3) COMMENT '重量',
    `volume` DECIMAL(8,3) COMMENT '体积',
    `unit` VARCHAR(20) COMMENT '单位',
    `images` JSON COMMENT '图片列表',
    `status` VARCHAR(20) DEFAULT 'active' COMMENT '状态',
    `is_hot` BOOLEAN DEFAULT FALSE COMMENT '是否热门',
    `is_new` BOOLEAN DEFAULT FALSE COMMENT '是否新品',
    `is_recommend` BOOLEAN DEFAULT FALSE COMMENT '是否推荐',
    `sort_order` INT DEFAULT 0 COMMENT '排序',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL,
    INDEX `idx_name` (`name`),
    INDEX `idx_category_id` (`category_id`),
    INDEX `idx_category_name` (`category_name`),
    INDEX `idx_brand_id` (`brand_id`),
    INDEX `idx_merchant_id` (`merchant_id`),
    INDEX `idx_price` (`price`),
    INDEX `idx_stock` (`stock`),
    INDEX `idx_sold_count` (`sold_count`),
    INDEX `idx_status` (`status`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品表';

-- 商品图片表
CREATE TABLE IF NOT EXISTS `product_images` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `product_id` BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
    `url` VARCHAR(500) NOT NULL COMMENT '图片URL',
    `type` VARCHAR(20) DEFAULT 'gallery' COMMENT '类型: main/gallery/detail',
    `sort_order` INT DEFAULT 0 COMMENT '排序',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_product_id` (`product_id`),
    INDEX `idx_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品图片表';

-- 购物车表
CREATE TABLE IF NOT EXISTS `carts` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '用户ID，0表示游客',
    `session_id` VARCHAR(100) COMMENT '会话ID',
    `status` VARCHAR(20) DEFAULT 'active' COMMENT '状态',
    `item_count` INT DEFAULT 0 COMMENT '商品种类数',
    `total_qty` INT DEFAULT 0 COMMENT '商品总数量',
    `total_amount` DECIMAL(10,2) DEFAULT 0.00 COMMENT '总金额',
    `version` INT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL,
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_session_id` (`session_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='购物车表';

-- 购物车项表
CREATE TABLE IF NOT EXISTS `cart_items` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `cart_id` BIGINT UNSIGNED NOT NULL COMMENT '购物车ID',
    `product_id` BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
    `sku_id` BIGINT UNSIGNED DEFAULT 0 COMMENT 'SKU ID',
    `quantity` INT NOT NULL DEFAULT 1 COMMENT '数量',
    `price` DECIMAL(10,2) NOT NULL COMMENT '单价',
    `total_amount` DECIMAL(10,2) NOT NULL COMMENT '小计',
    `selected` BOOLEAN DEFAULT TRUE COMMENT '是否选中',
    `version` INT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL,
    INDEX `idx_cart_id` (`cart_id`),
    INDEX `idx_product_id` (`product_id`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='购物车项表';

-- 订单表
CREATE TABLE IF NOT EXISTS `orders` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `order_no` VARCHAR(32) NOT NULL UNIQUE COMMENT '订单号',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `status` VARCHAR(20) NOT NULL COMMENT '订单状态',
    `payment_status` VARCHAR(20) NOT NULL COMMENT '支付状态',
    `shipping_status` VARCHAR(20) DEFAULT 'pending' COMMENT '物流状态',
    `order_type` VARCHAR(20) DEFAULT 'normal' COMMENT '订单类型',
    `payment_type` VARCHAR(20) COMMENT '支付方式',
    -- 金额信息
    `total_amount` DECIMAL(10,2) NOT NULL COMMENT '订单总额',
    `discount_amount` DECIMAL(10,2) DEFAULT 0.00 COMMENT '优惠金额',
    `shipping_fee` DECIMAL(10,2) DEFAULT 0.00 COMMENT '运费',
    `payable_amount` DECIMAL(10,2) NOT NULL COMMENT '应付金额',
    `paid_amount` DECIMAL(10,2) DEFAULT 0.00 COMMENT '实付金额',
    -- 收货信息
    `receiver_name` VARCHAR(50) COMMENT '收货人',
    `receiver_phone` VARCHAR(20) COMMENT '收货电话',
    `receiver_address` VARCHAR(500) COMMENT '收货地址',
    `receiver_province` VARCHAR(50) COMMENT '省',
    `receiver_city` VARCHAR(50) COMMENT '市',
    `receiver_district` VARCHAR(50) COMMENT '区',
    `receiver_zip` VARCHAR(10) COMMENT '邮编',
    -- 其他信息
    `remark` TEXT COMMENT '备注',
    `paid_at` TIMESTAMP NULL COMMENT '支付时间',
    `shipped_at` TIMESTAMP NULL COMMENT '发货时间',
    `delivered_at` TIMESTAMP NULL COMMENT '收货时间',
    `completed_at` TIMESTAMP NULL COMMENT '完成时间',
    `cancelled_at` TIMESTAMP NULL COMMENT '取消时间',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL,
    INDEX `idx_order_no` (`order_no`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_payment_status` (`payment_status`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';

-- 订单项表
CREATE TABLE IF NOT EXISTS `order_items` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
    `product_id` BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
    `sku_id` BIGINT UNSIGNED DEFAULT 0 COMMENT 'SKU ID',
    `product_name` VARCHAR(255) NOT NULL COMMENT '商品名称',
    `sku_name` VARCHAR(255) COMMENT 'SKU名称',
    `image` VARCHAR(500) COMMENT '商品图片',
    `price` DECIMAL(10,2) NOT NULL COMMENT '单价',
    `quantity` INT NOT NULL COMMENT '数量',
    `total_amount` DECIMAL(10,2) NOT NULL COMMENT '小计',
    `refund_status` VARCHAR(20) DEFAULT 'none' COMMENT '退款状态',
    `refund_quantity` INT DEFAULT 0 COMMENT '退款数量',
    `refund_amount` DECIMAL(10,2) DEFAULT 0.00 COMMENT '退款金额',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL,
    INDEX `idx_order_id` (`order_id`),
    INDEX `idx_product_id` (`product_id`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单项表';

-- 支付记录表
CREATE TABLE IF NOT EXISTS `payments` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `payment_no` VARCHAR(32) NOT NULL UNIQUE COMMENT '支付单号',
    `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `payment_method` VARCHAR(20) NOT NULL COMMENT '支付方式',
    `amount` DECIMAL(10,2) NOT NULL COMMENT '支付金额',
    `status` VARCHAR(20) NOT NULL COMMENT '支付状态',
    `transaction_id` VARCHAR(100) COMMENT '第三方交易号',
    `paid_at` TIMESTAMP NULL COMMENT '支付时间',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_payment_no` (`payment_no`),
    INDEX `idx_order_id` (`order_id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付记录表';

-- 地址表
CREATE TABLE IF NOT EXISTS `addresses` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `receiver_name` VARCHAR(50) NOT NULL COMMENT '收货人',
    `receiver_phone` VARCHAR(20) NOT NULL COMMENT '收货电话',
    `province` VARCHAR(50) NOT NULL COMMENT '省',
    `city` VARCHAR(50) NOT NULL COMMENT '市',
    `district` VARCHAR(50) NOT NULL COMMENT '区',
    `detail` VARCHAR(500) NOT NULL COMMENT '详细地址',
    `zip_code` VARCHAR(10) COMMENT '邮编',
    `is_default` BOOLEAN DEFAULT FALSE COMMENT '是否默认',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL,
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_is_default` (`is_default`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='地址表';

-- 文件表
CREATE TABLE IF NOT EXISTS `files` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `upload_user_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '上传用户ID',
    `filename` VARCHAR(255) NOT NULL COMMENT '文件名',
    `original_name` VARCHAR(255) NOT NULL COMMENT '原始文件名',
    `file_path` VARCHAR(500) NOT NULL COMMENT '文件路径',
    `file_url` VARCHAR(500) NOT NULL COMMENT '文件URL',
    `file_size` BIGINT NOT NULL COMMENT '文件大小(字节)',
    `mime_type` VARCHAR(100) COMMENT 'MIME类型',
    `file_type` VARCHAR(20) COMMENT '文件类型',
    `storage_type` VARCHAR(20) DEFAULT 'local' COMMENT '存储类型',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL,
    INDEX `idx_upload_user_id` (`upload_user_id`),
    INDEX `idx_file_type` (`file_type`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件表';

