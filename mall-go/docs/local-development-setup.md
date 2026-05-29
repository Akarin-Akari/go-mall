# 本地开发环境搭建指南

## 概述

本指南介绍如何使用Docker Compose快速搭建GoMall项目的本地开发环境。

---

## 1. 快速开始（5分钟）

### 第一步：克隆项目

```bash
git clone https://github.com/your-org/mall-go.git
cd mall-go
```

### 第二步：启动开发环境

```bash
# 构建镜像并启动所有服务
docker-compose up -d

# 等待服务启动（约30秒）
docker-compose ps

# 查看应用日志
docker-compose logs -f mall-go
```

### 第三步：验证服务

```bash
# 检查API服务
curl http://localhost:8081/health

# 检查MySQL
mysql -h 127.0.0.1 -u mall_user -p mall_password -D mall_go -e "SELECT 1"

# 检查Redis
redis-cli -h 127.0.0.1 -p 6379 -a redis123456 ping
```

---

## 2. 详细配置

### 服务端口

| 服务 | 端口 | 用途 |
|------|------|------|
| mall-go | 8081 | API服务 |
| MySQL | 3306 | 数据库 |
| Redis | 6379 | 缓存 |

### 数据库凭证

```
MySQL:
  Host: localhost
  Port: 3306
  Username: mall_user
  Password: mall_password
  Database: mall_go

Redis:
  Host: localhost
  Port: 6379
  Password: redis123456
  DB: 0
```

### 应用配置

```
API基础URL: http://localhost:8081
健康检查: http://localhost:8081/health
Metrics: http://localhost:8081/metrics
```

---

## 3. 常见开发任务

### 查看日志

```bash
# 查看所有服务日志
docker-compose logs

# 实时查看应用日志
docker-compose logs -f mall-go

# 查看MySQL日志
docker-compose logs -f mysql

# 查看Redis日志
docker-compose logs -f redis

# 查看最后100行日志
docker-compose logs --tail=100 mall-go
```

### 进入容器

```bash
# 进入应用容器
docker exec -it mall-go-app sh

# 进入MySQL容器
docker exec -it mall-go-mysql mysql -u root -p

# 进入Redis容器
docker exec -it mall-go-redis redis-cli
```

### 重启服务

```bash
# 重启所有服务
docker-compose restart

# 重启特定服务
docker-compose restart mall-go
docker-compose restart mysql
docker-compose restart redis
```

### 查看数据库

```bash
# 连接MySQL
mysql -h 127.0.0.1 -u mall_user -p mall_password -D mall_go

# 常用命令
SHOW TABLES;
SELECT * FROM users LIMIT 10;
DESC users;
```

### 查看缓存

```bash
# 连接Redis
redis-cli -h 127.0.0.1 -p 6379 -a redis123456

# 常用命令
KEYS *
GET key_name
DEL key_name
FLUSHDB
```

---

## 4. 开发工作流

### 修改代码后重新构建

```bash
# 方式1：重新构建镜像并启动
docker-compose build --no-cache
docker-compose up -d

# 方式2：只重启应用容器（如果只修改了配置）
docker-compose restart mall-go
```

### 查看实时日志

```bash
# 在一个终端中查看日志
docker-compose logs -f mall-go

# 在另一个终端中进行开发和测试
# 修改代码 → 构建 → 测试
```

### 调试应用

```bash
# 进入容器进行调试
docker exec -it mall-go-app sh

# 查看应用进程
ps aux

# 查看网络连接
netstat -an

# 查看环境变量
env | grep -E "DB_|REDIS_|JWT_"
```

---

## 5. 数据管理

### 备份数据

```bash
# 备份MySQL数据
docker exec mall-go-mysql mysqldump -u root -p mall_go > backup.sql

# 备份Redis数据
docker exec mall-go-redis redis-cli BGSAVE
docker cp mall-go-redis:/data/dump.rdb ./redis_backup.rdb
```

### 恢复数据

```bash
# 恢复MySQL数据
mysql -h 127.0.0.1 -u mall_user -p mall_password -D mall_go < backup.sql

# 恢复Redis数据
docker cp redis_backup.rdb mall-go-redis:/data/dump.rdb
docker exec mall-go-redis redis-cli SHUTDOWN
docker-compose restart redis
```

### 清空数据

```bash
# 清空所有数据卷
docker-compose down -v

# 重新启动（会重新初始化数据库）
docker-compose up -d
```

---

## 6. 性能优化

### 增加资源限制

编辑 `docker-compose.yml`，在 `mall-go` 服务中添加：

```yaml
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 2G
    reservations:
      cpus: '1'
      memory: 1G
```

### 优化MySQL性能

```bash
# 进入MySQL容器
docker exec -it mall-go-mysql mysql -u root -p

# 查看当前配置
SHOW VARIABLES LIKE 'max_connections';
SHOW VARIABLES LIKE 'innodb_buffer_pool_size';

# 修改配置（需要重启）
# 编辑docker-compose.yml中的MySQL command参数
```

### 优化Redis性能

```bash
# 进入Redis容器
docker exec -it mall-go-redis redis-cli

# 查看内存使用
INFO memory

# 查看连接数
INFO clients

# 查看性能统计
INFO stats
```

---

## 7. 故障排查

### 服务无法启动

```bash
# 查看详细错误
docker-compose logs mall-go

# 检查镜像是否存在
docker images | grep mall-go

# 重新构建
docker-compose build --no-cache
```

### 数据库连接失败

```bash
# 检查MySQL是否运行
docker-compose ps mysql

# 检查MySQL日志
docker-compose logs mysql

# 测试连接
docker exec mall-go-app ping mysql
```

### Redis连接失败

```bash
# 检查Redis是否运行
docker-compose ps redis

# 检查Redis日志
docker-compose logs redis

# 测试连接
docker exec mall-go-app ping redis
```

### 端口被占用

```bash
# 查看占用的端口
netstat -ano | findstr :8081  # Windows
lsof -i :8081                 # macOS/Linux

# 修改docker-compose.yml中的端口
# 例如: "8082:8081" 代替 "8081:8081"
```

---

## 8. 开发工具集成

### VS Code集成

安装Docker扩展：
1. 打开VS Code
2. 搜索"Docker"
3. 安装官方Docker扩展
4. 在Docker视图中管理容器

### IDE调试

```bash
# 进入容器进行调试
docker exec -it mall-go-app sh

# 查看应用日志
docker logs -f mall-go-app
```

---

## 9. 清理环境

### 停止所有服务

```bash
docker-compose down
```

### 删除所有数据

```bash
docker-compose down -v
```

### 清理镜像

```bash
docker rmi mall-go:latest
```

### 完全清理

```bash
# 停止所有容器
docker-compose down -v

# 删除镜像
docker rmi mall-go:latest

# 清理未使用的资源
docker system prune -a
```

---

## 10. 常见问题

### Q: 如何修改数据库密码？

A: 编辑 `docker-compose.yml`，修改 `MYSQL_PASSWORD` 和 `MYSQL_ROOT_PASSWORD` 环境变量，然后重新启动。

### Q: 如何访问MySQL数据库？

A: 使用MySQL客户端连接到 `localhost:3306`，用户名 `mall_user`，密码 `mall_password`。

### Q: 如何查看应用日志？

A: 运行 `docker-compose logs -f mall-go`。

### Q: 如何重新初始化数据库？

A: 运行 `docker-compose down -v && docker-compose up -d`。

### Q: 如何在容器中运行命令？

A: 使用 `docker exec -it mall-go-app <command>`。

---

**最后更新**: 2025-01-31

