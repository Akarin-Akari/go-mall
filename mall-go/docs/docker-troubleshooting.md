# Docker常见问题排查指南

## 快速诊断

### 第一步：检查容器状态

```bash
# 查看所有容器
docker-compose ps

# 预期输出：所有容器状态应为 "Up"
# NAME                COMMAND             STATUS
# mall-go-mysql       "docker-entrypoint" Up (healthy)
# mall-go-redis       "redis-server"      Up (healthy)
# mall-go-app         "./mall-server"     Up (healthy)
```

### 第二步：查看错误日志

```bash
# 查看应用日志
docker-compose logs mall-go

# 查看MySQL日志
docker-compose logs mysql

# 查看Redis日志
docker-compose logs redis
```

### 第三步：测试连接

```bash
# 测试API
curl http://localhost:8081/health

# 测试MySQL
docker exec mall-go-mysql mysql -u root -p -e "SELECT 1"

# 测试Redis
docker exec mall-go-redis redis-cli ping
```

---

## 常见问题及解决方案

### 问题1: 容器无法启动

**症状**:
```
ERROR: for mall-go Cannot start service mall-go: OCI runtime create failed
```

**原因**: 镜像构建失败或依赖服务未启动

**解决方案**:

```bash
# 1. 查看详细错误
docker-compose logs mall-go

# 2. 检查镜像是否存在
docker images | grep mall-go

# 3. 重新构建镜像
docker-compose build --no-cache

# 4. 清理并重新启动
docker-compose down -v
docker-compose up -d

# 5. 查看启动日志
docker-compose logs -f mall-go
```

---

### 问题2: 数据库连接失败

**症状**:
```
Error: connect ECONNREFUSED 127.0.0.1:3306
或
Error: dial tcp: lookup mysql on 127.0.0.1:53: no such host
```

**原因**: MySQL容器未启动或网络连接问题

**解决方案**:

```bash
# 1. 检查MySQL容器状态
docker-compose ps mysql

# 2. 查看MySQL日志
docker-compose logs mysql

# 3. 检查MySQL是否健康
docker-compose ps | grep mysql
# 应该显示 "Up (healthy)"

# 4. 测试MySQL连接
docker exec mall-go-mysql mysql -u root -p -e "SELECT 1"

# 5. 检查网络连接
docker exec mall-go-app ping mysql

# 6. 如果仍然失败，重启MySQL
docker-compose restart mysql

# 7. 等待MySQL完全启动（约30秒）
sleep 30
docker-compose restart mall-go
```

---

### 问题3: Redis连接失败

**症状**:
```
Error: connect ECONNREFUSED 127.0.0.1:6379
或
Error: WRONGPASS invalid username-password pair
```

**原因**: Redis容器未启动或密码错误

**解决方案**:

```bash
# 1. 检查Redis容器状态
docker-compose ps redis

# 2. 查看Redis日志
docker-compose logs redis

# 3. 测试Redis连接
docker exec mall-go-redis redis-cli ping

# 4. 检查Redis密码
docker exec mall-go-redis redis-cli -a redis123456 ping

# 5. 如果密码错误，检查docker-compose.yml中的配置
grep -A 5 "redis:" docker-compose.yml

# 6. 重启Redis
docker-compose restart redis
```

---

### 问题4: 端口被占用

**症状**:
```
Error response from daemon: Ports are not available: exposing port TCP 0.0.0.0:8081 -> 0.0.0.0:0: listen tcp 0.0.0.0:8081: bind: An attempt was made to reuse a socket address that cannot be reused.
```

**原因**: 端口已被其他进程占用

**解决方案**:

```bash
# Windows
netstat -ano | findstr :8081
taskkill /PID <PID> /F

# macOS/Linux
lsof -i :8081
kill -9 <PID>

# 或者修改docker-compose.yml中的端口映射
# 将 "8081:8081" 改为 "8082:8081"
```

---

### 问题5: 镜像构建失败

**症状**:
```
ERROR: failed to solve with frontend dockerfile.v0
```

**原因**: Dockerfile语法错误或依赖下载失败

**解决方案**:

```bash
# 1. 查看构建日志
docker-compose build --no-cache 2>&1 | tail -50

# 2. 检查Dockerfile语法
docker build --no-cache -t mall-go:test .

# 3. 清理Docker缓存
docker builder prune -a

# 4. 重新构建
docker-compose build --no-cache

# 5. 如果仍然失败，检查网络连接
# 确保能访问Go模块仓库
```

---

### 问题6: 磁盘空间不足

**症状**:
```
Error response from daemon: no space left on device
```

**原因**: Docker镜像、容器或数据卷占用过多空间

**解决方案**:

```bash
# 1. 查看Docker磁盘使用
docker system df

# 2. 清理未使用的镜像
docker image prune -a

# 3. 清理未使用的容器
docker container prune

# 4. 清理未使用的数据卷
docker volume prune

# 5. 清理所有未使用的资源
docker system prune -a

# 6. 查看数据卷大小
docker volume ls
du -sh /var/lib/docker/volumes/mall-go_mysql_data/_data
```

---

### 问题7: 应用启动但无法访问

**症状**:
```
curl: (7) Failed to connect to localhost port 8081: Connection refused
```

**原因**: 应用未正确启动或端口映射错误

**解决方案**:

```bash
# 1. 检查容器是否运行
docker-compose ps mall-go

# 2. 查看应用日志
docker-compose logs mall-go

# 3. 进入容器检查应用状态
docker exec -it mall-go-app sh
ps aux | grep mall-server

# 4. 检查端口映射
docker port mall-go-app

# 5. 测试容器内部连接
docker exec mall-go-app wget -O- http://localhost:8081/health

# 6. 检查防火墙
# Windows: 检查Windows防火墙设置
# macOS: 检查系统偏好设置 > 安全性与隐私
# Linux: sudo ufw allow 8081
```

---

### 问题8: 数据卷权限问题

**症状**:
```
Error: permission denied
或
Error: cannot open directory: Permission denied
```

**原因**: 数据卷权限不正确

**解决方案**:

```bash
# 1. 检查数据卷权限
docker exec mall-go-mysql ls -la /var/lib/mysql

# 2. 修复权限
docker exec mall-go-mysql chown -R mysql:mysql /var/lib/mysql

# 3. 重启容器
docker-compose restart mysql

# 4. 如果仍然失败，删除数据卷并重新创建
docker-compose down -v
docker-compose up -d
```

---

### 问题9: 内存不足

**症状**:
```
OOMKilled
或
Killed
```

**原因**: 容器内存限制不足

**解决方案**:

```bash
# 1. 查看容器内存使用
docker stats

# 2. 增加内存限制
# 编辑docker-compose.yml，在deploy.resources.limits中增加memory

# 3. 重启容器
docker-compose down
docker-compose up -d

# 4. 监控内存使用
docker stats --no-stream
```

---

### 问题10: 网络连接问题

**症状**:
```
Error: getaddrinfo ENOTFOUND mysql
或
Error: connect ETIMEDOUT
```

**原因**: 容器网络配置问题

**解决方案**:

```bash
# 1. 检查网络
docker network ls
docker network inspect mall-go_mall-network

# 2. 检查容器是否连接到网络
docker inspect mall-go-app | grep -A 10 "Networks"

# 3. 测试容器间连接
docker exec mall-go-app ping mysql
docker exec mall-go-app ping redis

# 4. 重新创建网络
docker-compose down
docker-compose up -d

# 5. 检查DNS解析
docker exec mall-go-app nslookup mysql
```

---

## 诊断工具

### 收集诊断信息

```bash
#!/bin/bash
# 创建诊断脚本

echo "=== Docker版本 ==="
docker --version
docker-compose --version

echo "=== 容器状态 ==="
docker-compose ps

echo "=== 容器日志 ==="
docker-compose logs --tail=50

echo "=== 镜像列表 ==="
docker images | grep mall-go

echo "=== 网络信息 ==="
docker network ls
docker network inspect mall-go_mall-network

echo "=== 磁盘使用 ==="
docker system df

echo "=== 资源使用 ==="
docker stats --no-stream
```

### 运行诊断

```bash
# 保存为 diagnose.sh
chmod +x diagnose.sh
./diagnose.sh > diagnosis.log 2>&1

# 查看诊断结果
cat diagnosis.log
```

---

## 性能诊断

### 查看容器资源使用

```bash
# 实时监控
docker stats

# 查看历史数据
docker stats --no-stream

# 查看特定容器
docker stats mall-go-app
```

### 查看日志性能

```bash
# 查看日志大小
du -sh /var/lib/docker/containers/*/

# 清理日志
docker exec mall-go-app truncate -s 0 /var/log/app.log
```

---

## 恢复步骤

### 完全重置

```bash
# 1. 停止所有服务
docker-compose down

# 2. 删除所有数据
docker-compose down -v

# 3. 删除镜像
docker rmi mall-go:latest

# 4. 清理Docker系统
docker system prune -a

# 5. 重新构建和启动
docker-compose build --no-cache
docker-compose up -d
```

---

**最后更新**: 2025-01-31

