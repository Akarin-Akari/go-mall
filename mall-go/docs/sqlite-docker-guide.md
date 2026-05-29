# SQLite Docker支持指南

## 概述

GoMall项目现在支持在Docker中使用SQLite数据库，提供了快速本地开发和测试的能力。

### 支持的数据库驱动

- ✅ **MySQL** - 生产环境推荐，支持高并发
- ✅ **SQLite** - 开发环境推荐，快速启动

---

## 快速开始

### 使用SQLite启动（推荐用于快速测试）

```bash
cd mall-go

# 构建镜像
docker-compose -f docker-compose.sqlite.yml build

# 启动服务（仅需Redis，无需MySQL）
docker-compose -f docker-compose.sqlite.yml up -d

# 验证服务
curl http://localhost:8081/health

# 查看日志
docker-compose -f docker-compose.sqlite.yml logs -f mall-go
```

### 使用MySQL启动（推荐用于完整测试）

```bash
cd mall-go

# 构建镜像
docker-compose build

# 启动服务（MySQL + Redis + GoMall）
docker-compose up -d

# 验证服务
curl http://localhost:8081/health

# 查看日志
docker-compose logs -f mall-go
```

---

## 配置文件说明

### docker-compose.yml（MySQL模式）

```yaml
environment:
  DB_DRIVER: mysql          # 使用MySQL驱动
  DB_HOST: mysql            # MySQL主机
  DB_PORT: 3306             # MySQL端口
  DB_USERNAME: mall_user    # MySQL用户名
  DB_PASSWORD: mall_password # MySQL密码
  DB_NAME: mall_go          # 数据库名称
```

### docker-compose.sqlite.yml（SQLite模式）

```yaml
environment:
  DB_DRIVER: sqlite         # 使用SQLite驱动
  DB_NAME: /app/data/mall_go.db  # SQLite数据库文件路径
```

---

## 数据库驱动切换

### 方式1：使用不同的docker-compose文件

```bash
# 切换到SQLite
docker-compose down
docker-compose -f docker-compose.sqlite.yml up -d

# 切换回MySQL
docker-compose -f docker-compose.sqlite.yml down
docker-compose up -d
```

### 方式2：使用环境变量覆盖

```bash
# 在MySQL配置中使用SQLite
docker-compose up -d -e DB_DRIVER=sqlite -e DB_NAME=/app/data/mall_go.db

# 在SQLite配置中使用MySQL
docker-compose -f docker-compose.sqlite.yml up -d \
  -e DB_DRIVER=mysql \
  -e DB_HOST=mysql \
  -e DB_PORT=3306 \
  -e DB_USERNAME=mall_user \
  -e DB_PASSWORD=mall_password \
  -e DB_NAME=mall_go
```

---

## 使用场景

### 何时使用SQLite

✅ **快速本地开发**
- 启动速度快（< 1秒）
- 无需MySQL服务
- 适合单人开发

✅ **功能演示**
- 快速搭建演示环境
- 无需复杂配置
- 数据自动持久化

✅ **单元测试**
- 每个测试使用独立数据库
- 快速清理和重置
- 无需数据库服务

✅ **CI/CD流水线**
- 快速构建和测试
- 减少资源占用
- 加快流水线速度

### 何时使用MySQL

✅ **生产环境**
- 支持高并发
- 数据安全性高
- 支持复杂查询

✅ **多用户开发**
- 支持多个开发者共享数据库
- 数据一致性保证
- 支持事务

✅ **性能测试**
- 真实生产环境模拟
- 准确的性能指标
- 支持复杂业务场景

---

## 数据持久化

### SQLite数据卷

SQLite数据库文件存储在Docker数据卷中，确保容器重启后数据不丢失。

```bash
# 查看SQLite数据卷
docker volume ls | grep sqlite

# 查看数据卷详情
docker volume inspect mall-go_sqlite_data_sqlite

# 备份SQLite数据库
docker run --rm -v mall-go_sqlite_data_sqlite:/data \
  -v $(pwd):/backup \
  alpine cp /data/mall_go.db /backup/mall_go.db.backup

# 恢复SQLite数据库
docker run --rm -v mall-go_sqlite_data_sqlite:/data \
  -v $(pwd):/backup \
  alpine cp /backup/mall_go.db.backup /data/mall_go.db
```

### MySQL数据卷

MySQL数据存储在mysql_data数据卷中。

```bash
# 查看MySQL数据卷
docker volume ls | grep mysql

# 备份MySQL数据库
docker exec mall-go-mysql mysqldump -u root -p mall_go > backup.sql

# 恢复MySQL数据库
docker exec -i mall-go-mysql mysql -u root -p mall_go < backup.sql
```

---

## 常见问题

### Q: SQLite和MySQL哪个更快？

**A**: SQLite启动更快（< 1秒 vs 10-30秒），但MySQL在并发查询时性能更好。

### Q: SQLite支持多用户并发吗？

**A**: SQLite支持有限的并发。对于高并发场景，建议使用MySQL。

### Q: 如何在SQLite和MySQL之间迁移数据？

**A**: 
1. 从SQLite导出数据：`sqlite3 mall_go.db ".dump" > dump.sql`
2. 导入到MySQL：`mysql -u root -p mall_go < dump.sql`

### Q: SQLite数据库文件在哪里？

**A**: 在Docker数据卷中，可以通过以下命令查看：
```bash
docker volume inspect mall-go_sqlite_data_sqlite
```

### Q: 如何清理SQLite数据？

**A**: 
```bash
# 删除数据卷（会删除所有数据）
docker-compose -f docker-compose.sqlite.yml down -v

# 重新启动（会创建新的空数据库）
docker-compose -f docker-compose.sqlite.yml up -d
```

---

## 性能对比

| 操作 | SQLite | MySQL |
|------|--------|-------|
| 启动时间 | < 1秒 | 10-30秒 |
| 单条查询 | 1-5ms | 2-10ms |
| 批量插入(1000条) | 50-100ms | 100-200ms |
| 并发查询(10个) | 10-50ms | 5-20ms |
| 内存占用 | < 50MB | 200-500MB |
| 磁盘占用 | < 10MB | 100MB+ |

---

## 最佳实践

### 开发工作流

```bash
# 1. 使用SQLite快速开发
docker-compose -f docker-compose.sqlite.yml up -d

# 2. 开发和测试
# ... 修改代码 ...

# 3. 定期切换到MySQL进行完整测试
docker-compose -f docker-compose.sqlite.yml down
docker-compose up -d

# 4. 运行完整测试套件
docker-compose exec mall-go go test ./...

# 5. 提交前清理
docker-compose down -v
```

### CI/CD流水线

```yaml
# GitHub Actions示例
- name: Run tests with SQLite
  run: |
    docker-compose -f docker-compose.sqlite.yml up -d
    docker-compose -f docker-compose.sqlite.yml exec -T mall-go go test ./...
    docker-compose -f docker-compose.sqlite.yml down -v
```

---

## 故障排查

### SQLite数据库被锁定

**症状**: `database is locked`

**解决方案**:
```bash
# 重启容器
docker-compose -f docker-compose.sqlite.yml restart mall-go

# 或删除数据卷重新开始
docker-compose -f docker-compose.sqlite.yml down -v
docker-compose -f docker-compose.sqlite.yml up -d
```

### 无法连接到SQLite

**症状**: `cannot open database file`

**解决方案**:
```bash
# 检查数据卷是否存在
docker volume ls | grep sqlite

# 检查容器日志
docker-compose -f docker-compose.sqlite.yml logs mall-go

# 确保/app/data目录存在
docker exec mall-go-app-sqlite ls -la /app/data
```

---

## 相关文档

- [Docker使用指南](./docker-usage-guide.md)
- [本地开发环境搭建](./local-development-setup.md)
- [Docker常见问题排查](./docker-troubleshooting.md)

---

**最后更新**: 2025-01-31

