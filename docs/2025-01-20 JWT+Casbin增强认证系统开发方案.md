# JWT + Casbin 增强认证系统开发方案

**项目**: Mall-Go 商城后端系统  
**版本**: v1.0  
**创建时间**: 2025-01-20  
**作者**: 系统架构师  

## 📋 目录

- [1. 技术方案设计](#1-技术方案设计)
- [2. API接口设计](#2-api接口设计)
- [3. 数据库变更](#3-数据库变更)
- [4. 安全考虑](#4-安全考虑)
- [5. 测试策略](#5-测试策略)
- [6. 实施计划](#6-实施计划)

## 1. 技术方案设计

### 1.1 整体架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端应用      │    │   API网关       │    │   认证服务      │
│                 │────│                 │────│                 │
│ - Web/Mobile    │    │ - 路由转发      │    │ - JWT生成/验证  │
│ - Token存储     │    │ - 限流/熔断     │    │ - 密码加密      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                       ┌─────────────────┐             │
                       │   权限服务      │             │
                       │                 │─────────────┘
                       │ - Casbin引擎    │
                       │ - 权限缓存      │
                       │ - 角色管理      │
                       └─────────────────┘
                                │
                       ┌─────────────────┐
                       │   数据存储      │
                       │                 │
                       │ - MySQL主库     │
                       │ - Redis缓存     │
                       └─────────────────┘
```

### 1.2 核心组件设计

#### 1.2.1 JWT Token 结构
```go
type MallClaims struct {
    UserID      uint      `json:"user_id"`
    Username    string    `json:"username"`
    Role        string    `json:"role"`        // user, merchant, admin
    StoreID     *uint     `json:"store_id"`    // 商家店铺ID（可选）
    Permissions []string  `json:"permissions"` // 缓存的权限列表
    DeviceID    string    `json:"device_id"`   // 设备标识
    LoginTime   time.Time `json:"login_time"`  // 登录时间
    jwt.RegisteredClaims
}
```

#### 1.2.2 权限模型设计
```ini
# Casbin RBAC模型
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

#### 1.2.3 多角色权限设计
```go
// 角色定义
const (
    RoleUser     = "user"     // 普通用户
    RoleMerchant = "merchant" // 商家
    RoleAdmin    = "admin"    // 管理员
)

// 权限资源定义
const (
    ResourceUser    = "user"    // 用户管理
    ResourceProduct = "product" // 商品管理
    ResourceOrder   = "order"   // 订单管理
    ResourceStore   = "store"   // 店铺管理
    ResourceSystem  = "system"  // 系统管理
)

// 操作定义
const (
    ActionRead   = "read"   // 查看
    ActionWrite  = "write"  // 编辑
    ActionDelete = "delete" // 删除
    ActionManage = "manage" // 管理
)
```

### 1.3 安全增强特性

#### 1.3.1 密码安全
- 使用 bcrypt 加密存储
- 密码强度验证
- 防暴力破解机制

#### 1.3.2 Token 安全
- JWT 签名验证
- Token 过期时间控制
- 自动续签机制
- 多设备登录管理

#### 1.3.3 权限安全
- 基于角色的访问控制（RBAC）
- 权限缓存机制
- 动态权限更新

## 2. API接口设计

### 2.1 认证相关接口

#### 2.1.1 用户注册
```http
POST /api/v1/auth/register
Content-Type: application/json

{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePass123!",
    "nickname": "测试用户",
    "role": "user"
}

Response:
{
    "code": 200,
    "message": "注册成功",
    "data": {
        "user": {
            "id": 1,
            "username": "testuser",
            "email": "test@example.com",
            "nickname": "测试用户",
            "role": "user",
            "created_at": "2025-01-20T10:00:00Z"
        }
    }
}
```

#### 2.1.2 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
    "username": "testuser",
    "password": "SecurePass123!",
    "device_id": "web_chrome_001"
}

Response:
{
    "code": 200,
    "message": "登录成功",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "expires_in": 86400,
        "user": {
            "id": 1,
            "username": "testuser",
            "role": "user",
            "permissions": ["user:read", "order:read"]
        }
    }
}
```

#### 2.1.3 Token刷新
```http
POST /api/v1/auth/refresh
Authorization: Bearer <refresh_token>

Response:
{
    "code": 200,
    "message": "Token刷新成功",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "expires_in": 86400
    }
}
```

### 2.2 权限管理接口

#### 2.2.1 获取用户权限
```http
GET /api/v1/auth/permissions
Authorization: Bearer <token>

Response:
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "permissions": [
            "user:read",
            "user:write",
            "order:read",
            "product:read"
        ],
        "roles": ["user"]
    }
}
```

## 3. 数据库变更

### 3.1 用户表增强
```sql
-- 修改用户表，增加安全字段
ALTER TABLE users 
ADD COLUMN password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
ADD COLUMN salt VARCHAR(32) COMMENT '密码盐值',
ADD COLUMN last_login_at TIMESTAMP NULL COMMENT '最后登录时间',
ADD COLUMN login_attempts INT DEFAULT 0 COMMENT '登录尝试次数',
ADD COLUMN locked_until TIMESTAMP NULL COMMENT '账户锁定到期时间',
ADD COLUMN two_factor_enabled BOOLEAN DEFAULT FALSE COMMENT '是否启用双因子认证',
ADD INDEX idx_username_status (username, status),
ADD INDEX idx_email_status (email, status);
```

### 3.2 权限相关表
```sql
-- 创建角色表
CREATE TABLE roles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE COMMENT '角色名称',
    display_name VARCHAR(100) NOT NULL COMMENT '显示名称',
    description TEXT COMMENT '角色描述',
    status VARCHAR(20) DEFAULT 'active' COMMENT '状态',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name_status (name, status)
);

-- 创建权限表
CREATE TABLE permissions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE COMMENT '权限名称',
    resource VARCHAR(50) NOT NULL COMMENT '资源',
    action VARCHAR(50) NOT NULL COMMENT '操作',
    description TEXT COMMENT '权限描述',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_resource_action (resource, action)
);

-- 创建用户角色关联表
CREATE TABLE user_roles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    UNIQUE KEY uk_user_role (user_id, role_id)
);

-- 创建角色权限关联表
CREATE TABLE role_permissions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    role_id BIGINT UNSIGNED NOT NULL,
    permission_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
    UNIQUE KEY uk_role_permission (role_id, permission_id)
);
```

### 3.3 设备登录记录表
```sql
-- 创建设备登录记录表
CREATE TABLE user_devices (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    device_id VARCHAR(100) NOT NULL COMMENT '设备标识',
    device_type VARCHAR(50) COMMENT '设备类型',
    device_name VARCHAR(100) COMMENT '设备名称',
    ip_address VARCHAR(45) COMMENT 'IP地址',
    user_agent TEXT COMMENT '用户代理',
    last_active_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '最后活跃时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uk_user_device (user_id, device_id),
    INDEX idx_user_active (user_id, last_active_at)
);
```

## 4. 安全考虑

### 4.1 密码安全策略
- **加密算法**: bcrypt (cost=12)
- **密码复杂度**: 最少8位，包含大小写字母、数字、特殊字符
- **防暴力破解**: 5次失败后锁定账户30分钟
- **密码历史**: 记录最近5次密码，防止重复使用

### 4.2 Token安全策略
- **签名算法**: HS256 (生产环境建议RS256)
- **过期时间**: Access Token 24小时，Refresh Token 7天
- **密钥管理**: 使用环境变量存储，定期轮换
- **Token撤销**: 支持单设备和全设备登出

### 4.3 权限安全策略
- **最小权限原则**: 用户只获得必需的最小权限
- **权限缓存**: Redis缓存5分钟，减少数据库查询
- **动态权限**: 支持实时权限变更和撤销
- **审计日志**: 记录所有权限变更操作

## 5. 测试策略

### 5.1 单元测试
- JWT生成和验证功能
- 密码加密和验证功能
- 权限检查逻辑
- 中间件功能测试

### 5.2 集成测试
- 完整的登录流程测试
- 权限验证流程测试
- Token刷新流程测试
- 多设备登录测试

### 5.3 安全测试
- 密码暴力破解测试
- Token伪造测试
- 权限绕过测试
- SQL注入防护测试

### 5.4 性能测试
- 高并发登录测试
- 权限验证性能测试
- 缓存命中率测试
- 数据库查询优化测试

## 6. 实施计划

### 阶段一：基础认证实现（预计3-4天）
1. 密码加密存储实现
2. JWT生成和验证实现
3. 基础认证中间件完善
4. 用户注册登录API完善

### 阶段二：权限管理集成（预计2-3天）
1. Casbin权限模型配置
2. 权限数据库表创建
3. 权限检查中间件实现
4. 权限管理API实现

### 阶段三：安全增强特性（预计2-3天）
1. 多设备登录管理
2. Token自动续签机制
3. 权限缓存优化
4. 安全日志记录

### 阶段四：测试和优化（预计2天）
1. 单元测试编写
2. 集成测试验证
3. 性能优化调整
4. 文档完善

---

**总预计开发时间**: 9-12天  
**关键里程碑**: 
- Day 4: 基础认证功能完成
- Day 7: 权限管理功能完成  
- Day 10: 安全增强功能完成
- Day 12: 测试和优化完成

**风险评估**: 
- 🟢 技术风险: 低（基于成熟技术栈）
- 🟡 时间风险: 中（功能较多，需要仔细测试）
- 🟢 质量风险: 低（有完整的测试策略）
