# Docker使用指南

## 概述

GoMall项目提供了完整的Docker容器化支持，包括：
- ✅ 多阶段Dockerfile构建
- ✅ 开发环境docker-compose编排
- ✅ 生产环境docker-compose编排
- ✅ 自动健康检查
- ✅ 数据卷持久化
- ✅ 网络隔离

---

## 1. 前置要求

### 安装Docker和Docker Compose

```bash
# Windows (使用Docker Desktop)
# 下载: https://www.docker.com/products/docker-desktop

# macOS
brew install docker docker-compose

# Linux (Ubuntu/Debian)
sudo apt-get install docker.io docker-compose
sudo usermod -aG docker $USER
```

### 验证安装

```bash
docker --version
docker-compose --version
```

---

## 2. 开发环境使用

### 2.1 MySQL开发环境（推荐）

```bash
# 进入项目目录
cd mall-go

# 构建镜像
docker-compose build

# 启动所有服务（MySQL + Redis + GoMall）
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f mall-go
```

### 2.2 SQLite开发环境（快速测试）

SQLite是一个轻量级的文件数据库，适合快速本地开发和测试，无需MySQL服务。

```bash
# 进入项目目录
cd mall-go

# 使用SQLite配置启动（仅需Redis）
docker-compose -f docker-compose.sqlite.yml build

# 启动SQLite开发环境
docker-compose -f docker-compose.sqlite.yml up -d

# 查看服务状态
docker-compose -f docker-compose.sqlite.yml ps

# 查看日志
docker-compose -f docker-compose.sqlite.yml logs -f mall-go
```

### 2.3 在MySQL和SQLite之间切换

**切换到SQLite**:
```bash
# 停止MySQL环境
docker-compose down

# 启动SQLite环境
docker-compose -f docker-compose.sqlite.yml up -d
```

**切换到MySQL**:
```bash
# 停止SQLite环境
docker-compose -f docker-compose.sqlite.yml down

# 启动MySQL环境
docker-compose up -d
```

### 2.4 使用环境变量切换数据库驱动

如果需要在同一个docker-compose配置中切换，可以使用环境变量：

```bash
# 使用MySQL
docker-compose up -d -e DB_DRIVER=mysql

# 使用SQLite
docker-compose up -d -e DB_DRIVER=sqlite -e DB_NAME=/app/data/mall_go.db
```

### 2.5 SQLite vs MySQL对比

| 特性 | SQLite | MySQL |
|------|--------|-------|
| 启动速度 | ⚡ 极快（<1秒） | 🐢 较慢（10-30秒） |
| 资源占用 | 💚 极低 | 📊 中等 |
| 并发支持 | ⚠️ 有限 | ✅ 优秀 |
| 数据持久化 | ✅ 文件存储 | ✅ 数据库存储 |
| 适用场景 | 快速测试、演示、单机 | 生产环境、多用户 |
| 数据卷大小 | 📦 < 10MB | 📦 100MB+ |

### 2.6 访问服务

```bash
# API服务
curl http://localhost:8081/health

# MySQL（仅MySQL模式）
mysql -h 127.0.0.1 -u mall_user -p mall_password -D mall_go

# Redis（两种模式都有）
redis-cli -h 127.0.0.1 -p 6379 -a redis123456

# SQLite数据库文件（SQLite模式）
# 位置: docker volume inspect mall-go_sqlite_data_sqlite
```

### 2.7 停止开发环境

```bash
# 停止MySQL环境
docker-compose down

# 停止MySQL环境并删除数据卷
docker-compose down -v

# 停止SQLite环境
docker-compose -f docker-compose.sqlite.yml down

# 停止SQLite环境并删除数据卷
docker-compose -f docker-compose.sqlite.yml down -v
```

### 2.8 查看容器日志

```bash
# MySQL环境 - 查看所有服务日志
docker-compose logs

# MySQL环境 - 查看特定服务日志
docker-compose logs mysql
docker-compose logs redis
docker-compose logs mall-go

# SQLite环境 - 查看所有服务日志
docker-compose -f docker-compose.sqlite.yml logs

# SQLite环境 - 查看特定服务日志
docker-compose -f docker-compose.sqlite.yml logs redis
docker-compose -f docker-compose.sqlite.yml logs mall-go

# 实时查看日志
docker-compose logs -f mall-go
docker-compose -f docker-compose.sqlite.yml logs -f mall-go

# 查看最后100行日志
docker-compose logs --tail=100 mall-go
docker-compose -f docker-compose.sqlite.yml logs --tail=100 mall-go
```

---

## 3. 生产环境使用

### 环境变量配置

创建 `.env.prod` 文件：

```bash
# 数据库配置
DB_ROOT_PASSWORD=your-secure-root-password
DB_USERNAME=mall_user
DB_PASSWORD=your-secure-db-password
DB_NAME=mall_go

# Redis配置
REDIS_PASSWORD=your-secure-redis-password

# JWT配置
JWT_SECRET=your-secure-jwt-secret-at-least-32-chars

# CORS配置
CORS_ALLOW_ORIGINS=https://www.example.com,https://admin.example.com
```

### 启动生产环境

```bash
# 使用生产环境配置启动
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d

# 查看服务状态
docker-compose -f docker-compose.prod.yml ps

# 查看日志
docker-compose -f docker-compose.prod.yml logs -f mall-go
```

### 停止生产环境

```bash
# 停止所有服务
docker-compose -f docker-compose.prod.yml down

# 停止并删除数据卷（谨慎操作！）
docker-compose -f docker-compose.prod.yml down -v
```

---

## 4. 常用命令

### 容器管理

```bash
# 查看所有容器
docker ps -a

# 进入容器
docker exec -it mall-go-app sh

# 查看容器日志
docker logs -f mall-go-app

# 重启容器
docker restart mall-go-app

# 停止容器
docker stop mall-go-app

# 启动容器
docker start mall-go-app
```

### 镜像管理

```bash
# 查看镜像
docker images

# 删除镜像
docker rmi mall-go:latest

# 构建镜像
docker build -t mall-go:latest .

# 标记镜像
docker tag mall-go:latest myregistry/mall-go:v1.0.0

# 推送镜像
docker push myregistry/mall-go:v1.0.0
```

### 数据卷管理

```bash
# 查看数据卷
docker volume ls

# 查看数据卷详情
docker volume inspect mall-go_mysql_data

# 删除数据卷
docker volume rm mall-go_mysql_data

# 备份数据卷
docker run --rm -v mall-go_mysql_data:/data -v $(pwd):/backup \
  alpine tar czf /backup/mysql_backup.tar.gz -C /data .
```

### 网络管理

```bash
# 查看网络
docker network ls

# 查看网络详情
docker network inspect mall-go_mall-network

# 连接容器到网络
docker network connect mall-go_mall-network container_name
```

---

## 5. 故障排查

### 问题1: 容器无法启动

**症状**: `docker-compose up` 失败

**解决方案**:
```bash
# 查看详细错误日志
docker-compose logs mall-go

# 检查镜像是否存在
docker images | grep mall-go

# 重新构建镜像
docker-compose build --no-cache

# 清理并重新启动
docker-compose down -v
docker-compose up -d
```

### 问题2: 数据库连接失败

**症状**: 应用无法连接MySQL

**解决方案**:
```bash
# 检查MySQL容器状态
docker-compose ps mysql

# 查看MySQL日志
docker-compose logs mysql

# 进入MySQL容器测试
docker exec -it mall-go-mysql mysql -u root -p -e "SELECT 1"

# 检查网络连接
docker exec mall-go-app ping mysql
```

### 问题3: Redis连接失败

**症状**: 应用无法连接Redis

**解决方案**:
```bash
# 检查Redis容器状态
docker-compose ps redis

# 查看Redis日志
docker-compose logs redis

# 进入Redis容器测试
docker exec -it mall-go-redis redis-cli ping

# 检查网络连接
docker exec mall-go-app ping redis
```

### 问题4: 端口被占用

**症状**: `Error response from daemon: Ports are not available`

**解决方案**:
```bash
# 查看占用的端口
netstat -ano | findstr :8081  # Windows
lsof -i :8081                 # macOS/Linux

# 修改docker-compose.yml中的端口映射
# 例如: "8082:8081" 代替 "8081:8081"

# 或者停止占用端口的进程
```

### 问题5: 磁盘空间不足

**症状**: `no space left on device`

**解决方案**:
```bash
# 清理未使用的镜像
docker image prune -a

# 清理未使用的容器
docker container prune

# 清理未使用的数据卷
docker volume prune

# 清理所有未使用的资源
docker system prune -a
```

---

## 6. 性能优化

### 镜像优化

```bash
# 查看镜像大小
docker images mall-go

# 使用多阶段构建减小镜像体积
# Dockerfile已使用多阶段构建

# 使用alpine基础镜像
# Dockerfile已使用alpine:latest
```

### 容器优化

```bash
# 限制容器资源使用
docker run -m 512m --cpus="1" mall-go:latest

# 在docker-compose中配置资源限制
# 已在docker-compose.prod.yml中配置
```

### 日志优化

```bash
# 配置日志驱动和大小限制
# 已在docker-compose.yml中配置
# max-size: "10m"
# max-file: "3"
```

---

## 7. 安全最佳实践

### 1. 使用非root用户

```dockerfile
# Dockerfile已配置非root用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser
USER appuser
```

### 2. 敏感信息管理

```bash
# 使用环境变量而不是硬编码
# 使用.env文件管理敏感信息
# 不要将.env文件提交到Git

# .gitignore中添加
echo ".env*" >> .gitignore
```

### 3. 镜像扫描

```bash
# 扫描镜像漏洞
docker scan mall-go:latest

# 使用Trivy扫描
trivy image mall-go:latest
```

### 4. 网络隔离

```bash
# 使用自定义网络而不是host网络
# docker-compose.yml已配置自定义网络
```

---

## 8. 监控和日志

### 查看容器资源使用

```bash
# 实时查看资源使用
docker stats

# 查看特定容器
docker stats mall-go-app
```

### 收集日志

```bash
# 导出日志
docker logs mall-go-app > app.log 2>&1

# 查看特定时间范围的日志
docker logs --since 2025-01-31T10:00:00 mall-go-app
docker logs --until 2025-01-31T11:00:00 mall-go-app
```

---

## 9. 参考资源

- [Docker官方文档](https://docs.docker.com/)
- [Docker Compose文档](https://docs.docker.com/compose/)
- [Dockerfile最佳实践](https://docs.docker.com/develop/dev-best-practices/)
- [Docker安全最佳实践](https://docs.docker.com/engine/security/)

---

**最后更新**: 2025-01-31

