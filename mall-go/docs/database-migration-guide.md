# 数据库迁移指南

## 概述

本项目使用 `golang-migrate` 工具进行数据库版本化迁移管理，支持升级和回滚操作。

## 迁移文件结构

```
migrations/
├── 000001_init_schema.up.sql      # 初始化表结构（升级）
├── 000001_init_schema.down.sql    # 初始化表结构（回滚）
├── 000002_add_indexes.up.sql      # 添加索引（升级）
├── 000002_add_indexes.down.sql    # 添加索引（回滚）
├── 000003_seed_data.up.sql        # 初始化数据（升级）
└── 000003_seed_data.down.sql      # 初始化数据（回滚）
```

## 迁移命令

### 1. 执行所有未应用的迁移（升级）

```bash
go run cmd/db-migrate/main.go -cmd=up
```

**预期输出**：
```
✅ 迁移成功完成，当前版本: 3 (dirty: false)
```

### 2. 回滚最后一次迁移

```bash
go run cmd/db-migrate/main.go -cmd=down -steps=1
```

**预期输出**：
```
✅ 回滚成功，当前版本: 2 (dirty: false)
```

### 3. 回滚所有迁移

```bash
go run cmd/db-migrate/main.go -cmd=down
```

**预期输出**：
```
✅ 已回滚所有迁移
```

### 4. 查看当前迁移版本

```bash
go run cmd/db-migrate/main.go -cmd=version
```

**预期输出**：
```
📌 当前迁移版本: 3 (dirty: false)
```

### 5. 强制设置版本号（用于修复）

```bash
go run cmd/db-migrate/main.go -cmd=force -force=1
```

**预期输出**：
```
✅ 已强制设置版本为: 1
```

## 环境配置

### 开发环境

使用 SQLite 数据库（不支持迁移工具，使用 GORM AutoMigrate）：

```bash
ENV=dev go run cmd/server/main.go
```

### 生产环境

使用 MySQL 数据库（支持迁移工具）：

```bash
# 设置环境变量
export ENV=prod
export DB_PASSWORD=your-db-password
export JWT_SECRET=your-jwt-secret

# 执行迁移
go run cmd/db-migrate/main.go -cmd=up

# 启动服务
go run cmd/server/main.go
```

## 创建新的迁移文件

### 命名规范

迁移文件必须遵循以下命名规范：

```
{version}_{description}.{up|down}.sql
```

例如：
- `000004_add_user_profile.up.sql`
- `000004_add_user_profile.down.sql`

### 创建步骤

1. 创建升级脚本（`.up.sql`）：

```sql
-- 000004_add_user_profile.up.sql
CREATE TABLE IF NOT EXISTS `user_profiles` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT UNSIGNED NOT NULL UNIQUE,
    `bio` TEXT,
    `avatar` VARCHAR(255),
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

2. 创建回滚脚本（`.down.sql`）：

```sql
-- 000004_add_user_profile.down.sql
DROP TABLE IF EXISTS `user_profiles`;
```

3. 执行迁移：

```bash
go run cmd/db-migrate/main.go -cmd=up
```

## 最佳实践

### 1. 幂等性

所有迁移脚本应该是幂等的，即多次执行结果相同：

```sql
-- ✅ 好的做法
CREATE TABLE IF NOT EXISTS `users` (...);
ALTER TABLE `users` ADD COLUMN IF NOT EXISTS `phone` VARCHAR(20);

-- ❌ 不好的做法
CREATE TABLE `users` (...);  -- 第二次执行会报错
ALTER TABLE `users` ADD COLUMN `phone` VARCHAR(20);  -- 第二次执行会报错
```

### 2. 不使用外键

遵循 RuoYi 开发规范，不使用外键约束：

```sql
-- ✅ 好的做法
CREATE TABLE `order_items` (
    `order_id` BIGINT UNSIGNED NOT NULL,
    INDEX `idx_order_id` (`order_id`)
);

-- ❌ 不好的做法
CREATE TABLE `order_items` (
    `order_id` BIGINT UNSIGNED NOT NULL,
    FOREIGN KEY (`order_id`) REFERENCES `orders`(`id`)
);
```

### 3. 向后兼容

新的迁移应该向后兼容，避免破坏性变更：

```sql
-- ✅ 好的做法：添加新列时设置默认值
ALTER TABLE `users` ADD COLUMN `status` VARCHAR(20) DEFAULT 'active';

-- ❌ 不好的做法：添加 NOT NULL 列但不设置默认值
ALTER TABLE `users` ADD COLUMN `status` VARCHAR(20) NOT NULL;
```

### 4. 数据迁移

如果需要迁移数据，应该在单独的迁移文件中进行：

```sql
-- 000005_migrate_user_status.up.sql
UPDATE `users` SET `status` = 'active' WHERE `status` IS NULL;
```

## 故障排查

### 问题1：迁移失败，版本标记为 dirty

**症状**：
```
📌 当前迁移版本: 2 (dirty: true)
```

**解决方案**：

1. 检查数据库中失败的迁移脚本
2. 手动修复数据库问题
3. 强制设置版本号：

```bash
go run cmd/db-migrate/main.go -cmd=force -force=2
```

### 问题2：迁移文件找不到

**症状**：
```
❌ 创建迁移实例失败: file does not exist
```

**解决方案**：

确保在项目根目录下执行命令，且 `migrations/` 目录存在。

### 问题3：数据库连接失败

**症状**：
```
❌ 数据库连接测试失败: dial tcp: connection refused
```

**解决方案**：

1. 检查数据库服务是否启动
2. 检查配置文件中的数据库连接信息
3. 检查环境变量是否正确设置

## 生产环境部署流程

### 1. 备份数据库

```bash
mysqldump -u root -p mall > backup_$(date +%Y%m%d_%H%M%S).sql
```

### 2. 执行迁移

```bash
# 查看当前版本
go run cmd/db-migrate/main.go -cmd=version

# 执行迁移
go run cmd/db-migrate/main.go -cmd=up

# 验证版本
go run cmd/db-migrate/main.go -cmd=version
```

### 3. 验证数据

```bash
# 连接数据库
mysql -u root -p mall

# 检查表结构
SHOW TABLES;
DESC users;

# 检查数据
SELECT COUNT(*) FROM users;
```

### 4. 回滚（如果需要）

```bash
# 回滚最后一次迁移
go run cmd/db-migrate/main.go -cmd=down -steps=1

# 恢复备份
mysql -u root -p mall < backup_20250131_120000.sql
```

## 参考资料

- [golang-migrate 官方文档](https://github.com/golang-migrate/migrate)
- [MySQL 迁移最佳实践](https://github.com/golang-migrate/migrate/blob/master/database/mysql/TUTORIAL.md)

