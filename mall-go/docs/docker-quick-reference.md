# Docker快速参考卡片

## 🚀 快速启动

### 开发环境（5分钟）

```bash
# 进入项目目录
cd mall-go

# 启动所有服务
docker-compose up -d

# 验证服务
curl http://localhost:8081/health

# 查看日志
docker-compose logs -f mall-go
```

### 生产环境

```bash
# 创建环境文件
cat > .env.prod << EOF
DB_ROOT_PASSWORD=your-password
DB_USERNAME=mall_user
DB_PASSWORD=your-password
REDIS_PASSWORD=your-password
JWT_SECRET=your-secret-key
CORS_ALLOW_ORIGINS=https://www.example.com
EOF

# 启动生产环境
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d
```

---

## 📋 常用命令

### 容器管理

```bash
# 查看容器状态
docker-compose ps

# 查看日志
docker-compose logs -f mall-go

# 进入容器
docker exec -it mall-go-app sh

# 重启容器
docker-compose restart mall-go

# 停止服务
docker-compose down

# 停止并删除数据
docker-compose down -v
```

### 镜像管理

```bash
# 构建镜像
docker-compose build

# 查看镜像
docker images | grep mall-go

# 删除镜像
docker rmi mall-go:latest
```

### 数据库操作

```bash
# 连接MySQL
mysql -h 127.0.0.1 -u mall_user -p mall_password -D mall_go

# 连接Redis
redis-cli -h 127.0.0.1 -p 6379 -a redis123456

# 备份MySQL
docker exec mall-go-mysql mysqldump -u root -p mall_go > backup.sql

# 备份Redis
docker exec mall-go-redis redis-cli BGSAVE
```

---

## 🌐 访问地址

| 服务 | 地址 | 用途 |
|------|------|------|
| API | http://localhost:8081 | 应用服务 |
| 健康检查 | http://localhost:8081/health | 服务健康状态 |
| Metrics | http://localhost:8081/metrics | Prometheus指标 |
| MySQL | localhost:3306 | 数据库 |
| Redis | localhost:6379 | 缓存 |

---

## 💾 数据库凭证

### MySQL

```
Host: localhost
Port: 3306
Username: mall_user
Password: mall_password
Database: mall_go
```

### Redis

```
Host: localhost
Port: 6379
Password: redis123456
DB: 0
```

---

## 📊 服务配置

### 开发环境

| 服务 | 镜像 | 端口 | 重启策略 |
|------|------|------|---------|
| MySQL | mysql:8.0 | 3306 | unless-stopped |
| Redis | redis:7-alpine | 6379 | unless-stopped |
| GoMall | mall-go:latest | 8081 | unless-stopped |

### 生产环境

| 服务 | 镜像 | 端口 | CPU限制 | 内存限制 |
|------|------|------|---------|---------|
| MySQL | mysql:8.0 | 3306 | 2 | 2G |
| Redis | redis:7-alpine | 6379 | 1 | 1G |
| GoMall | mall-go:latest | 8081 | 2 | 2G |

---

## 🔍 故障排查

### 容器无法启动

```bash
# 查看错误日志
docker-compose logs mall-go

# 重新构建镜像
docker-compose build --no-cache

# 清理并重新启动
docker-compose down -v
docker-compose up -d
```

### 数据库连接失败

```bash
# 检查MySQL状态
docker-compose ps mysql

# 查看MySQL日志
docker-compose logs mysql

# 测试连接
docker exec mall-go-app ping mysql
```

### Redis连接失败

```bash
# 检查Redis状态
docker-compose ps redis

# 查看Redis日志
docker-compose logs redis

# 测试连接
docker exec mall-go-app ping redis
```

### 端口被占用

```bash
# Windows
netstat -ano | findstr :8081

# macOS/Linux
lsof -i :8081

# 修改docker-compose.yml中的端口映射
```

---

## 📁 文件结构

```
mall-go/
├── Dockerfile                    # 多阶段构建
├── .dockerignore                 # 构建优化
├── docker-compose.yml            # 开发环境编排
├── docker-compose.prod.yml       # 生产环境编排
├── scripts/
│   ├── docker-start.sh          # Linux/macOS启动脚本
│   └── docker-start.bat         # Windows启动脚本
├── docs/
│   ├── docker-usage-guide.md    # 使用指南
│   ├── local-development-setup.md # 开发环境搭建
│   ├── docker-troubleshooting.md # 问题排查
│   └── docker-quick-reference.md # 快速参考
└── configs/
    ├── config.yaml              # 开发配置
    └── config-prod.yaml         # 生产配置
```

---

## 🛠️ 启动脚本

### Linux/macOS

```bash
# 开发环境
./scripts/docker-start.sh dev

# 生产环境
./scripts/docker-start.sh prod
```

### Windows

```batch
# 开发环境
scripts\docker-start.bat dev

# 生产环境
scripts\docker-start.bat prod
```

---

## 📚 相关文档

- [Docker使用指南](./docker-usage-guide.md)
- [本地开发环境搭建](./local-development-setup.md)
- [Docker常见问题排查](./docker-troubleshooting.md)

---

## ⚡ 性能优化

### 镜像优化

- ✅ 多阶段构建减小体积
- ✅ 使用alpine基础镜像
- ✅ 删除调试符号

### 容器优化

- ✅ 资源限制（CPU、内存）
- ✅ 日志大小限制
- ✅ 非root用户运行

### 网络优化

- ✅ 自定义网络隔离
- ✅ 服务间直接通信
- ✅ 健康检查配置

---

## 🔐 安全最佳实践

- ✅ 使用非root用户
- ✅ 敏感信息通过环境变量
- ✅ 使用.env文件管理配置
- ✅ 定期扫描镜像漏洞
- ✅ 网络隔离

---

**最后更新**: 2025-01-31

