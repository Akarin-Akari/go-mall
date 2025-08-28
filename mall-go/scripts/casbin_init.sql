-- Casbin权限管理系统数据库初始化脚本
-- 适用于Mall-Go商城系统

-- ========== 创建Casbin规则表 ==========
-- 注意：这个表会由gorm-adapter自动创建，这里提供参考结构

CREATE TABLE IF NOT EXISTS `casbin_rule` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `ptype` varchar(100) NOT NULL DEFAULT '' COMMENT '策略类型(p=policy, g=grouping)',
  `v0` varchar(100) NOT NULL DEFAULT '' COMMENT '主体(subject)/用户/角色',
  `v1` varchar(100) NOT NULL DEFAULT '' COMMENT '对象(object)/资源',
  `v2` varchar(100) NOT NULL DEFAULT '' COMMENT '动作(action)/操作',
  `v3` varchar(100) NOT NULL DEFAULT '' COMMENT '扩展字段1',
  `v4` varchar(100) NOT NULL DEFAULT '' COMMENT '扩展字段2',
  `v5` varchar(100) NOT NULL DEFAULT '' COMMENT '扩展字段3',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_casbin_rule` (`ptype`,`v0`,`v1`,`v2`,`v3`,`v4`,`v5`),
  KEY `idx_casbin_rule_ptype` (`ptype`),
  KEY `idx_casbin_rule_v0` (`v0`),
  KEY `idx_casbin_rule_v1` (`v1`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Casbin权限规则表';

-- ========== 插入默认权限策略 ==========

-- 清空现有数据（可选，首次运行时使用）
-- DELETE FROM casbin_rule;

-- 插入角色权限策略 (p策略)

-- 普通用户权限
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
('p', 'user', 'user', 'read'),
('p', 'user', 'user', 'write'),
('p', 'user', 'product', 'read'),
('p', 'user', 'category', 'read'),
('p', 'user', 'order', 'read'),
('p', 'user', 'order', 'create'),
('p', 'user', 'file', 'create');

-- 商家权限
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
('p', 'merchant', 'product', 'read'),
('p', 'merchant', 'product', 'write'),
('p', 'merchant', 'product', 'create'),
('p', 'merchant', 'product', 'delete'),
('p', 'merchant', 'category', 'read'),
('p', 'merchant', 'category', 'write'),
('p', 'merchant', 'category', 'create'),
('p', 'merchant', 'order', 'read'),
('p', 'merchant', 'order', 'write'),
('p', 'merchant', 'store', 'read'),
('p', 'merchant', 'store', 'write'),
('p', 'merchant', 'file', 'create'),
('p', 'merchant', 'file', 'read'),
('p', 'merchant', 'report', 'read');

-- 管理员权限
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
('p', 'admin', 'user', 'manage'),
('p', 'admin', 'product', 'manage'),
('p', 'admin', 'category', 'manage'),
('p', 'admin', 'order', 'manage'),
('p', 'admin', 'store', 'manage'),
('p', 'admin', 'system', 'manage'),
('p', 'admin', 'config', 'manage'),
('p', 'admin', 'file', 'manage'),
('p', 'admin', 'report', 'manage');

-- ========== 插入用户角色关系 (g策略) ==========

-- 示例用户角色分配（实际使用时根据具体用户ID调整）
-- 格式：INSERT INTO casbin_rule (ptype, v0, v1) VALUES ('g', 'user:用户ID', '角色');

-- 示例：用户ID为1的用户是管理员
INSERT INTO casbin_rule (ptype, v0, v1) VALUES
('g', 'user:1', 'admin');

-- 示例：用户ID为2的用户是商家
INSERT INTO casbin_rule (ptype, v0, v1) VALUES
('g', 'user:2', 'merchant');

-- 示例：用户ID为3的用户是普通用户
INSERT INTO casbin_rule (ptype, v0, v1) VALUES
('g', 'user:3', 'user');

-- ========== 权限验证查询示例 ==========

-- 查看所有权限策略
-- SELECT * FROM casbin_rule WHERE ptype = 'p' ORDER BY v0, v1, v2;

-- 查看所有用户角色关系
-- SELECT * FROM casbin_rule WHERE ptype = 'g' ORDER BY v0;

-- 查看特定角色的权限
-- SELECT * FROM casbin_rule WHERE ptype = 'p' AND v0 = 'admin';

-- 查看特定用户的角色
-- SELECT * FROM casbin_rule WHERE ptype = 'g' AND v0 = 'user:1';

-- ========== 权限管理常用操作 ==========

-- 为用户添加角色
-- INSERT INTO casbin_rule (ptype, v0, v1) VALUES ('g', 'user:用户ID', '角色名');

-- 移除用户角色
-- DELETE FROM casbin_rule WHERE ptype = 'g' AND v0 = 'user:用户ID' AND v1 = '角色名';

-- 添加权限策略
-- INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', '角色名', '资源', '操作');

-- 移除权限策略
-- DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = '角色名' AND v1 = '资源' AND v2 = '操作';

-- ========== 索引优化建议 ==========

-- 如果数据量大，可以考虑添加以下复合索引
-- CREATE INDEX idx_casbin_rule_subject_object ON casbin_rule(v0, v1);
-- CREATE INDEX idx_casbin_rule_ptype_subject ON casbin_rule(ptype, v0);

-- ========== 数据完整性检查 ==========

-- 检查权限策略数量
-- SELECT ptype, COUNT(*) as count FROM casbin_rule GROUP BY ptype;

-- 检查各角色权限数量
-- SELECT v0 as role, COUNT(*) as permission_count 
-- FROM casbin_rule 
-- WHERE ptype = 'p' 
-- GROUP BY v0 
-- ORDER BY permission_count DESC;

-- 检查用户角色分配
-- SELECT v1 as role, COUNT(*) as user_count 
-- FROM casbin_rule 
-- WHERE ptype = 'g' 
-- GROUP BY v1 
-- ORDER BY user_count DESC;
