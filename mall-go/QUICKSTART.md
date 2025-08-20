# 🚀 Mall Go 快速启动指南

## ⚡ 5 分钟快速启动

### 1. 环境检查

确保你已经安装了：

- ✅ Go 1.21+
- ✅ MySQL 8.0+

### 2. 配置 Go 环境变量（现代方式）

**Windows 系统：**

1. 右键"此电脑" → "属性" → "高级系统设置" → "环境变量"
2. 添加系统变量：
   ```
   GOROOT = C:\Go                    # Go安装目录
   GOPATH = C:\Users\YourName\go     # 全局工作空间（可选）
   ```
3. 在 Path 中添加：
   ```
   %GOROOT%\bin
   %GOPATH%\bin
   ```

**重要说明：**

- GOPATH 现在主要用于全局工具安装
- 项目可以在任意目录下，使用 Go Modules 管理
- 不需要为每个项目修改 GOPATH！

### 3. 验证 Go 环境

```bash
go version
go env GOPATH
go env GOROOT
```

### 4. 数据库准备

```sql
-- 登录MySQL
mysql -u root -p

-- 创建数据库
CREATE DATABASE mall_go CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 执行初始化脚本
mysql -u root -p mall_go < scripts/init-db.sql
```

### 5. 修改配置文件

编辑 `configs/config.yaml`：

```yaml
database:
  username: root
  password: your_mysql_password # 修改为你的MySQL密码
```

### 6. 启动项目

```bash
# 进入项目目录
cd mall-go

# 安装依赖
go mod tidy

# 启动项目
go run cmd/server/main.go
```

### 7. 测试 API

```bash
# 健康检查
curl http://localhost:8080/api/v1/health

# 获取商品列表
curl http://localhost:8080/api/v1/products
```

## 🎯 多项目开发示例

使用 Go Modules，你可以在任意目录下创建项目：

```bash
# 项目1：商城项目
E:\Workspace_Draft\GoLang\goProject\mall-go\
├── go.mod
├── cmd/
└── ...

# 项目2：博客API
E:\Workspace_Draft\GoLang\goProject\blog-api\
├── go.mod
├── cmd/
└── ...

# 项目3：微服务
E:\Workspace_Draft\GoLang\goProject\microservice\
├── go.mod
├── cmd/
└── ...
```

每个项目独立管理，不需要修改 GOPATH！

## 🔍 常见问题

### Q: Go 命令未找到？

A: 检查环境变量 PATH 是否包含 Go 的 bin 目录，重启命令行

### Q: 数据库连接失败？

A: 检查 MySQL 服务是否启动，配置文件中的连接信息是否正确

### Q: 依赖下载失败？

A: 配置 Go 代理：

```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

### Q: 为什么不需要修改 GOPATH？

A: Go Modules 让每个项目独立管理依赖，GOPATH 只用于全局工具

## 📚 详细文档

- [Go 环境配置最佳实践](docs/Go环境配置最佳实践.md)
- [项目设置指南](docs/项目设置指南.md)
- [技术选型推荐](docs/技术选型推荐.md)
- [API 文档](http://localhost:8080/swagger/index.html) (启动后访问)

---

**开始你的 Go 语言学习之旅吧！🎉**
