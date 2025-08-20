# Mall Go - Go 语言商城后端项目

## 📖 项目简介

Mall Go 是一个基于 Go 语言和 Gin 框架开发的现代化商城后端系统，专为学习和练手而设计。项目采用微服务架构思想，包含完整的用户管理、商品管理、订单管理等核心功能。

## 🏗️ 技术栈

### 核心框架

- **Web 框架**: [Gin](https://github.com/gin-gonic/gin) - 高性能 HTTP Web 框架
- **数据库**: MySQL 8.0 - 关系型数据库
- **ORM**: [GORM](https://gorm.io/) - Go 语言 ORM 库

### 工具库

- **配置管理**: [Viper](https://github.com/spf13/viper) - 配置解决方案
- **日志**: [Zap](https://github.com/uber-go/zap) - 高性能日志库
- **认证**: [JWT](https://github.com/golang-jwt/jwt) - JSON Web Token
- **权限**: [Casbin](https://github.com/casbin/casbin) - 权限管理
- **验证**: [Validator](https://github.com/go-playground/validator) - 数据验证
- **API 文档**: [Swagger](https://github.com/swaggo/swag) - API 文档生成
- **数值计算**: [Decimal](https://github.com/shopspring/decimal) - 精确数值计算

## 📁 项目结构

```
mall-go/
├── cmd/                    # 应用入口
│   └── server/
│       └── main.go
├── internal/              # 内部包
│   ├── config/           # 配置管理
│   ├── handler/          # HTTP处理器
│   ├── service/          # 业务逻辑层
│   ├── repository/       # 数据访问层
│   ├── model/            # 数据模型
│   ├── middleware/       # 中间件
│   └── utils/            # 工具函数
├── pkg/                  # 可导出的包
│   ├── auth/             # 认证相关
│   ├── database/         # 数据库连接
│   └── response/         # 响应封装
├── configs/              # 配置文件
├── scripts/              # 脚本文件
├── go.mod
├── go.sum
└── README.md
```

## 🚀 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+
- Redis (可选)

### 安装依赖

```bash
# 克隆项目
git clone <repository-url>
cd mall-go

# 安装依赖
go mod tidy
```

### 配置数据库

1. 创建 MySQL 数据库

```sql
CREATE DATABASE mall_go CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

2. 修改配置文件 `configs/config.yaml`

```yaml
database:
  host: localhost
  port: 3306
  username: your_username
  password: your_password
  dbname: mall_go
```

### 运行项目

```bash
# 开发模式
go run cmd/server/main.go

# 或者构建后运行
go build -o mall-go cmd/server/main.go
./mall-go
```

## 📚 API 文档

启动服务后，访问 Swagger 文档：

- 地址: http://localhost:8080/swagger/index.html

## 🔧 开发指南

### 添加新的 API

1. 在 `internal/model/` 中定义数据模型
2. 在 `internal/handler/` 中实现 HTTP 处理器
3. 在 `internal/service/` 中实现业务逻辑
4. 在 `internal/repository/` 中实现数据访问
5. 在 `internal/handler/routes.go` 中注册路由

### 数据库迁移

```bash
# 自动迁移（开发环境）
go run cmd/server/main.go
```

## 📝 功能模块

### 用户模块

- [x] 用户注册
- [x] 用户登录
- [x] 用户信息管理
- [x] 权限控制

### 商品模块

- [x] 商品管理
- [x] 商品分类
- [x] 商品图片
- [x] 库存管理

### 订单模块

- [x] 订单创建
- [x] 订单状态管理
- [x] 订单查询
- [ ] 支付集成

### 系统功能

- [x] JWT 认证
- [x] 权限管理
- [x] 日志记录
- [x] 配置管理
- [ ] 文件上传
- [ ] 缓存管理

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🆘 支持

如果你遇到任何问题，请：

1. 查看 [Issues](../../issues) 页面
2. 创建新的 Issue 描述问题
3. 联系项目维护者

---

**Happy Coding! 🎉**
