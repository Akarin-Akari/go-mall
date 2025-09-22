# Mall-Go 电商系统部署与启动指南

## 📋 文档概述

**目标用户**: 测试人员、运维人员、新开发者  
**文档目的**: 确保任何人都能从零开始成功部署并启动完整的前后端服务，进行端到端的黑盒功能测试  
**系统架构**: Go 后端 API + React 前端 + SQLite 数据库  
**完成度**: 91%，核心功能完全可用

## 🎯 系统概览

### 核心服务

- **后端 API 服务**: Go + Gin + GORM，端口 8081
- **前端 Web 应用**: Next.js + TypeScript + Ant Design，端口 3001
- **数据库**: SQLite，包含 15 个测试商品数据
- **认证系统**: JWT Token 认证

### 主要功能

- ✅ 用户注册/登录
- ✅ 商品浏览/搜索/详情
- ✅ 购物车管理
- ✅ 订单创建/管理
- ✅ 完整的电商购物流程

---

## 1. 环境准备要求

### 1.1 操作系统支持

- ✅ **Windows 10/11** (推荐)
- ✅ **Linux** (Ubuntu 20.04+, CentOS 8+)
- ✅ **macOS** (10.15+)

### 1.2 必需软件版本

#### Go 环境

```bash
# 要求版本: Go 1.19+
go version
# 预期输出: go version go1.19+ windows/amd64
```

#### Node.js 环境

```bash
# 要求版本: Node.js 18+
node --version
# 预期输出: v18.0.0+

npm --version
# 预期输出: 8.0.0+
```

#### Git 版本控制

```bash
git --version
# 预期输出: git version 2.30.0+
```

### 1.3 端口要求

- **8081**: 后端 API 服务端口
- **3001**: 前端 Web 服务端口
- **确保这两个端口未被占用**

### 1.4 硬件要求

- **内存**: 最低 4GB，推荐 8GB+
- **存储**: 最低 2GB 可用空间
- **网络**: 需要互联网连接下载依赖

---

## 2. 后端服务部署步骤

### 2.1 Go 环境配置验证

```bash
# 1. 检查Go版本
go version

# 2. 检查Go环境变量
go env GOPATH
go env GOROOT

# 3. 验证Go模块支持
go mod help
```

### 2.2 项目克隆和依赖安装

```bash
# 1. 克隆项目（替换为实际的Git仓库地址）
git clone https://github.com/your-repo/mall-go.git
cd mall-go

# 2. 进入后端目录
cd mall-go

# 3. 下载Go依赖
go mod tidy
go mod download

# 4. 验证依赖安装
go mod verify
```

### 2.3 数据库初始化

```bash
# SQLite数据库会在首次启动时自动创建
# 数据库文件位置: mall-go/mall_go.db
# 包含15个测试商品数据和3个商品分类
```

### 2.4 配置文件设置

检查配置文件 `config/config.yaml`:

```yaml
server:
  port: 8081
  mode: debug

database:
  type: sqlite
  dsn: mall_go.db

jwt:
  secret: your-jwt-secret-key
  expire: 24h

redis:
  addr: localhost:6379
  password: ""
  db: 0
```

### 2.5 启动后端服务

```bash
# 方法1: 直接运行
go run cmd/server/main.go

# 方法2: 编译后运行
go build -o mall-go cmd/server/main.go
./mall-go

# 预期输出:
# [GIN-debug] Listening and serving HTTP on :8081
# Database connected successfully
# 15 products loaded
```

### 2.6 后端服务验证

#### 健康检查 API 测试

```bash
# 使用curl测试
curl http://localhost:8081/health

# 预期响应:
{
  "message": "Mall Go API is running",
  "status": "ok",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

#### 核心 API 端点验证

```bash
# 1. 获取商品列表
curl http://localhost:8081/api/v1/products

# 2. 获取商品分类
curl http://localhost:8081/api/v1/categories

# 3. 用户注册测试
curl -X POST http://localhost:8081/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"123456789"}'
```

---

## 3. 前端服务部署步骤

### 3.1 Node.js 环境验证

```bash
# 1. 检查Node.js版本
node --version

# 2. 检查npm版本
npm --version

# 3. 检查npm配置
npm config list
```

### 3.2 依赖包安装

```bash
# 1. 进入前端目录
cd mall-frontend

# 2. 安装依赖包
npm install

# 预期输出:
# added 1019 packages in 30s
# 267 packages are looking for funding

# 3. 验证关键依赖
npm list next
npm list antd
npm list @reduxjs/toolkit
```

### 3.3 环境变量配置

创建 `.env.local` 文件:

```bash
# API配置
NEXT_PUBLIC_API_BASE_URL=http://localhost:8081
NEXT_PUBLIC_API_TIMEOUT=10000

# 应用配置
NEXT_PUBLIC_APP_NAME=Mall-Go
NEXT_PUBLIC_APP_VERSION=1.0.0

# 调试配置
NEXT_PUBLIC_DEBUG=true
```

### 3.4 启动前端服务

```bash
# 启动开发服务器
npm run dev

# 预期输出:
# ▲ Next.js 15.5.2 (Turbopack)
# - Local:        http://localhost:3001
# - Network:      http://192.168.1.14:3001
# ✓ Ready in 4.4s
```

### 3.5 前端服务验证

#### 访问首页

```bash
# 浏览器访问
http://localhost:3001

# 预期结果:
# - 页面正常加载
# - 显示轮播图广告
# - 显示商品分类
# - 显示推荐商品列表
```

---

## 4. 服务验证清单

### 4.1 后端 API 端点可用性检查

| API 端点                 | 方法 | 功能     | 验证命令                                       | 预期状态码 |
| ------------------------ | ---- | -------- | ---------------------------------------------- | ---------- |
| `/health`                | GET  | 健康检查 | `curl http://localhost:8081/health`            | 200        |
| `/api/v1/products`       | GET  | 商品列表 | `curl http://localhost:8081/api/v1/products`   | 200        |
| `/api/v1/categories`     | GET  | 商品分类 | `curl http://localhost:8081/api/v1/categories` | 200        |
| `/api/v1/users/register` | POST | 用户注册 | 见上方示例                                     | 200        |
| `/api/v1/users/login`    | POST | 用户登录 | 见下方示例                                     | 200        |

#### 用户登录 API 测试

```bash
curl -X POST http://localhost:8081/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"username":"newuser2024","password":"123456789"}'

# 预期响应包含token字段
```

### 4.2 前端页面访问验证

| 页面路径    | 功能     | 验证方法   | 预期结果         |
| ----------- | -------- | ---------- | ---------------- |
| `/`         | 首页     | 浏览器访问 | 显示轮播图和商品 |
| `/login`    | 登录页   | 浏览器访问 | 显示登录表单     |
| `/register` | 注册页   | 浏览器访问 | 显示注册表单     |
| `/products` | 商品列表 | 浏览器访问 | 显示商品网格     |
| `/cart`     | 购物车   | 浏览器访问 | 显示购物车页面   |
| `/orders`   | 订单列表 | 需要登录   | 显示订单管理     |
| `/checkout` | 结算页   | 需要登录   | 显示订单确认     |

### 4.3 前后端数据交互测试

#### 测试步骤

1. **前端访问商品列表** → **后端返回商品数据** → **前端正确显示**
2. **前端用户登录** → **后端验证并返回 Token** → **前端保存认证状态**
3. **前端添加购物车** → **后端保存购物车数据** → **前端更新购物车状态**

---

## 5. 黑盒测试指南

### 5.1 测试用户账号信息

#### 预置测试账号

```
用户名: newuser2024
密码: 123456789
邮箱: newuser@example.com
```

#### 新注册测试账号

```
用户名: testuser_[时间戳]
密码: 123456789
邮箱: test_[时间戳]@example.com
```

### 5.2 完整用户购物流程测试

#### 测试流程 1: 新用户注册购物流程

```
1. 访问首页 http://localhost:3001
   ✓ 页面正常加载，显示商品推荐

2. 点击"注册"按钮
   ✓ 跳转到注册页面 /register

3. 填写注册信息并提交
   ✓ 注册成功，自动跳转到登录页

4. 使用注册账号登录
   ✓ 登录成功，跳转到首页，显示用户信息

5. 浏览商品列表
   ✓ 点击"浏览商品"，显示15个测试商品

6. 查看商品详情
   ✓ 点击任意商品，显示详细信息

7. 添加商品到购物车
   ✓ 点击"加入购物车"，显示成功提示

8. 查看购物车
   ✓ 点击购物车图标，显示已添加的商品

9. 确认订单
   ✓ 点击"去结算"，跳转到结算页面

10. 填写收货信息并提交订单
    ✓ 订单创建成功，跳转到订单详情页

11. 查看订单列表
    ✓ 在"我的订单"中显示刚创建的订单
```

#### 测试流程 2: 已有用户购物流程

```
1. 直接登录 (用户名: newuser2024, 密码: 123456789)
2. 搜索商品 (搜索关键词: "iPhone")
3. 筛选商品 (按价格区间筛选)
4. 批量添加购物车 (添加多个商品)
5. 购物车管理 (修改数量、删除商品)
6. 创建订单 (选择配送方式、支付方式)
7. 订单管理 (查看订单状态、取消订单)
```

### 5.3 关键功能测试用例

#### 用户认证功能

| 测试用例     | 操作步骤           | 预期结果               |
| ------------ | ------------------ | ---------------------- |
| 用户注册     | 填写有效信息提交   | 注册成功，跳转登录页   |
| 用户登录     | 输入正确用户名密码 | 登录成功，显示用户信息 |
| 登录状态保持 | 刷新页面           | 保持登录状态           |
| 用户登出     | 点击登出按钮       | 清除登录状态，跳转首页 |

#### 商品浏览功能

| 测试用例     | 操作步骤         | 预期结果           |
| ------------ | ---------------- | ------------------ |
| 商品列表加载 | 访问商品页面     | 显示 15 个测试商品 |
| 商品搜索     | 输入"iPhone"搜索 | 显示相关商品       |
| 商品筛选     | 选择价格区间筛选 | 显示符合条件商品   |
| 商品详情     | 点击商品卡片     | 显示详细信息页面   |

#### 购物车功能

| 测试用例   | 操作步骤         | 预期结果                   |
| ---------- | ---------------- | -------------------------- |
| 添加商品   | 点击"加入购物车" | 显示成功提示，购物车数量+1 |
| 查看购物车 | 点击购物车图标   | 显示已添加商品列表         |
| 修改数量   | 调整商品数量     | 实时更新总金额             |
| 删除商品   | 点击删除按钮     | 商品从购物车移除           |

#### 订单功能

| 测试用例 | 操作步骤         | 预期结果             |
| -------- | ---------------- | -------------------- |
| 创建订单 | 填写收货信息提交 | 订单创建成功         |
| 查看订单 | 访问订单列表页   | 显示用户所有订单     |
| 订单详情 | 点击订单查看详情 | 显示完整订单信息     |
| 取消订单 | 点击取消按钮     | 订单状态更新为已取消 |

---

## 6. 故障排除

### 6.1 常见启动错误及解决方案

#### 后端启动错误

**错误 1: 端口 8081 被占用**

```bash
# 错误信息: bind: address already in use
# 解决方案1: 查找占用进程
netstat -ano | findstr :8081  # Windows
lsof -i :8081                 # Linux/macOS

# 解决方案2: 杀死占用进程
taskkill /PID [进程ID] /F     # Windows
kill -9 [进程ID]              # Linux/macOS

# 解决方案3: 修改配置文件端口
# 编辑 config/config.yaml，修改 server.port
```

**错误 2: Go 依赖下载失败**

```bash
# 错误信息: go: module not found
# 解决方案1: 设置Go代理
go env -w GOPROXY=https://goproxy.cn,direct

# 解决方案2: 清理模块缓存
go clean -modcache
go mod download

# 解决方案3: 验证Go版本
go version  # 确保版本 >= 1.19
```

**错误 3: 数据库连接失败**

```bash
# 错误信息: database connection failed
# 解决方案: 检查SQLite文件权限
ls -la mall_go.db
chmod 666 mall_go.db  # Linux/macOS
```

#### 前端启动错误

**错误 1: Node.js 版本不兼容**

```bash
# 错误信息: Node.js version not supported
# 解决方案: 升级Node.js版本
node --version  # 确保版本 >= 18.0.0

# 使用nvm管理Node.js版本
nvm install 18
nvm use 18
```

**错误 2: npm 依赖安装失败**

```bash
# 错误信息: npm install failed
# 解决方案1: 清理npm缓存
npm cache clean --force

# 解决方案2: 删除node_modules重新安装
rm -rf node_modules package-lock.json
npm install

# 解决方案3: 使用yarn替代npm
npm install -g yarn
yarn install
```

**错误 3: 端口 3001 被占用**

```bash
# 错误信息: Port 3001 is already in use
# 解决方案1: Next.js会自动使用下一个可用端口
# 解决方案2: 手动指定端口
npm run dev -- -p 3002
```

### 6.2 API 连接失败排查

#### 前端无法连接后端 API

**检查清单:**

```bash
# 1. 确认后端服务正在运行
curl http://localhost:8081/health

# 2. 检查前端环境变量
cat .env.local
# 确认 NEXT_PUBLIC_API_BASE_URL=http://localhost:8081

# 3. 检查浏览器网络面板
# F12 → Network → 查看API请求状态

# 4. 检查CORS配置
# 后端应该允许前端域名的跨域请求
```

**常见解决方案:**

```bash
# 1. 重启后端服务
# Ctrl+C 停止服务，然后重新运行

# 2. 清除浏览器缓存
# Ctrl+Shift+R 强制刷新

# 3. 检查防火墙设置
# 确保8081和3000端口未被防火墙阻止
```

### 6.3 数据问题排查

#### 商品数据未加载

**检查步骤:**

```bash
# 1. 检查数据库文件
ls -la mall_go.db

# 2. 检查后端日志
# 查看启动日志中是否有"15 products loaded"

# 3. 直接查询API
curl http://localhost:8081/api/v1/products

# 4. 检查数据库内容（如果需要）
sqlite3 mall_go.db
.tables
SELECT COUNT(*) FROM products;
.quit
```

### 6.4 性能问题排查

#### 页面加载缓慢

**优化建议:**

```bash
# 1. 检查网络连接
ping localhost

# 2. 检查系统资源
# Windows: 任务管理器
# Linux: top 或 htop
# macOS: Activity Monitor

# 3. 清理浏览器缓存
# 4. 关闭不必要的应用程序
# 5. 重启开发服务器
```

---

## 7. 成功部署验证

### 7.1 最终验证清单

完成以下所有检查项，确认系统部署成功:

- [ ] **后端服务**: `curl http://localhost:8081/health` 返回 200 状态码
- [ ] **前端服务**: 浏览器访问 `http://localhost:3000` 正常显示首页
- [ ] **数据库**: 商品列表显示 15 个测试商品
- [ ] **用户认证**: 能够成功注册和登录用户
- [ ] **购物流程**: 能够完成从浏览商品到创建订单的完整流程
- [ ] **API 交互**: 前端能够正常调用后端 API 并显示数据

### 7.2 部署成功标志

当看到以下输出时，表示系统部署成功:

**后端控制台输出:**

```
[GIN-debug] Listening and serving HTTP on :8081
Database connected successfully
15 products loaded
3 categories loaded
```

**前端控制台输出:**

```
▲ Next.js 15.5.2 (Turbopack)
- Local:        http://localhost:3001
✓ Ready in 4.4s
```

**浏览器访问结果:**

- 首页显示轮播图、商品分类、推荐商品
- 用户能够正常注册、登录
- 商品列表显示完整的商品信息
- 购物车和订单功能正常工作

---

## 8. 联系支持

如果按照本指南操作仍遇到问题，请提供以下信息:

1. **操作系统版本**
2. **Go 和 Node.js 版本**
3. **具体错误信息**
4. **操作步骤**
5. **控制台日志输出**

---

**Mall-Go 电商系统部署与启动指南 v1.0**  
**更新时间**: 2024 年 1 月  
**适用版本**: Mall-Go v1.0  
**文档状态**: 生产就绪
