# GoMall  - 企业级Go语言商城全栈项目 🚀

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Gin Framework](https://img.shields.io/badge/Gin-1.9+-green.svg)](https://github.com/gin-gonic/gin)
[![React](https://img.shields.io/badge/React-18+-61DAFB.svg)](https://reactjs.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5+-3178C6.svg)](https://www.typescriptlang.org)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## 📖 项目简介

**GoMall** 是一个企业级的全栈电商项目，包含完整的前后端实现和系统性的Go语言学习文档。项目不仅是一个功能完整的商城系统，更是一个从入门到精通的Go语言学习平台。

### 🌟 项目特色

- **🏗️ 企业级架构**：采用微服务架构思想，代码结构清晰，易于维护和扩展
- **📚 完整学习体系**：包含14章系统性Go语言学习文档，从基础到高级全覆盖
- **🎯 实战导向**：结合真实商城项目，理论与实践完美结合
- **🔧 现代化技术栈**：使用最新的Go生态和前端技术
- **🧪 完善测试**：包含单元测试、集成测试、前端测试框架
- **📦 容器化部署**：支持Docker和Kubernetes部署
- **📊 监控运维**：集成监控、日志、性能分析等生产级功能

### 🎓 适合人群

- **Go语言初学者**：通过系统性文档和实战项目快速入门
- **有经验的开发者**：学习企业级Go项目架构和最佳实践
- **全栈开发者**：了解现代化前后端分离架构
- **面试准备者**：包含大量面试常考点和标准答案

## 🏗️ 技术栈

### 后端技术栈

#### 核心框架
- **Web 框架**: [Gin](https://github.com/gin-gonic/gin) - 高性能 HTTP Web 框架
- **数据库**: MySQL 8.0 - 关系型数据库
- **ORM**: [GORM](https://gorm.io/) - Go 语言 ORM 库
- **缓存**: Redis - 高性能内存数据库

#### 工具库
- **配置管理**: [Viper](https://github.com/spf13/viper) - 配置解决方案
- **日志**: [Zap](https://github.com/uber-go/zap) - 高性能日志库
- **认证**: [JWT](https://github.com/golang-jwt/jwt) - JSON Web Token
- **权限**: [Casbin](https://github.com/casbin/casbin) - 权限管理
- **验证**: [Validator](https://github.com/go-playground/validator) - 数据验证
- **API 文档**: [Swagger](https://github.com/swaggo/swag) - API 文档生成
- **数值计算**: [Decimal](https://github.com/shopspring/decimal) - 精确数值计算
- **消息队列**: RabbitMQ/Kafka - 异步消息处理

#### 监控运维
- **监控**: Prometheus + Grafana - 指标监控和可视化
- **链路追踪**: Jaeger/OpenTelemetry - 分布式链路追踪
- **容器化**: Docker + Kubernetes - 容器化部署
- **CI/CD**: GitHub Actions - 自动化构建部署

### 前端技术栈

- **框架**: React 18 + TypeScript - 现代化前端框架
- **构建工具**: Next.js 14 - 全栈React框架
- **状态管理**: Redux Toolkit - 状态管理
- **UI组件**: Tailwind CSS - 原子化CSS框架
- **HTTP客户端**: Axios - HTTP请求库
- **测试框架**: Jest + React Testing Library - 前端测试

### 开发工具

- **版本控制**: Git + GitHub - 代码版本管理
- **API测试**: Postman/Insomnia - API接口测试
- **数据库管理**: MySQL Workbench/DBeaver - 数据库管理工具
- **代码编辑器**: VS Code/GoLand - 开发环境

## 📁 项目结构

```
go-mall/
├── 📚 trial/                          # Go语言学习文档系列
│   ├── 01-basics/                     # 基础篇（4章）
│   │   ├── 01-variables-and-types.md      # 变量类型与基本语法
│   │   ├── 02-control-structures.md       # 控制结构与流程控制
│   │   ├── 03-functions-and-methods.md    # 函数方法与包管理
│   │   └── 04-packages-and-imports.md     # 包管理与模块化
│   ├── 02-advanced/                   # 进阶篇（4章）
│   │   ├── 01-structs-and-interfaces.md   # 结构体与接口
│   │   ├── 02-error-handling.md           # 错误处理与异常管理
│   │   ├── 03-concurrency-basics.md       # 并发编程与goroutine
│   │   └── 04-interface-design-patterns.md # 接口设计与多态
│   ├── 03-practical/                 # 实战篇（4章）
│   │   ├── 01-gin-framework-basics.md     # Gin框架入门与实践
│   │   ├── 02-gorm-database-operations.md # GORM数据库操作
│   │   ├── 03-redis-cache-applications.md # Redis缓存应用
│   │   └── 04-message-queue-integration.md # 消息队列集成
│   ├── 04-architecture/              # 架构篇（1章）
│   │   └── 01-microservices-design.md     # 微服务架构设计
│   ├── 05-advanced/                  # 高级篇（1章）
│   │   └── 01-production-practices.md     # 生产实践与运维
│   └── README.md                      # 学习路径指南
├── 🖥️ mall-go/                        # Go后端项目
│   ├── cmd/                           # 应用入口
│   │   └── server/main.go
│   ├── internal/                      # 内部包
│   │   ├── config/                    # 配置管理
│   │   ├── handler/                   # HTTP处理器
│   │   │   ├── user/                  # 用户相关接口
│   │   │   ├── product/               # 商品相关接口
│   │   │   ├── order/                 # 订单相关接口
│   │   │   └── cart/                  # 购物车相关接口
│   │   ├── service/                   # 业务逻辑层
│   │   ├── repository/                # 数据访问层
│   │   ├── model/                     # 数据模型
│   │   ├── middleware/                # 中间件
│   │   └── utils/                     # 工具函数
│   ├── pkg/                           # 可导出的包
│   │   ├── auth/                      # 认证相关
│   │   ├── database/                  # 数据库连接
│   │   ├── cache/                     # 缓存管理
│   │   ├── payment/                   # 支付集成
│   │   └── response/                  # 响应封装
│   ├── configs/                       # 配置文件
│   ├── tests/                         # 测试文件
│   │   ├── handler/                   # 接口测试
│   │   ├── integration/               # 集成测试
│   │   └── helpers/                   # 测试辅助
│   ├── scripts/                       # 脚本文件
│   ├── go.mod
│   └── go.sum
├── 🌐 mall-frontend/                  # React前端项目
│   ├── src/
│   │   ├── app/                       # Next.js页面
│   │   │   ├── login/                 # 登录页面
│   │   │   ├── register/              # 注册页面
│   │   │   ├── products/              # 商品页面
│   │   │   ├── cart/                  # 购物车页面
│   │   │   └── orders/                # 订单页面
│   │   ├── components/                # React组件
│   │   │   ├── business/              # 业务组件
│   │   │   ├── layout/                # 布局组件
│   │   │   └── __tests__/             # 组件测试
│   │   ├── store/                     # Redux状态管理
│   │   ├── utils/                     # 工具函数
│   │   └── types/                     # TypeScript类型定义
│   ├── public/                        # 静态资源
│   ├── package.json
│   └── next.config.js
├── 📋 docs/                           # 项目文档
│   ├── api/                           # API文档
│   ├── deployment/                    # 部署文档
│   └── testing/                       # 测试文档
├── 🔧 scripts/                        # 项目脚本
│   ├── quick-start.sh                 # 快速启动脚本
│   ├── run-tests.sh                   # 测试运行脚本
│   └── deploy.sh                      # 部署脚本
├── 🐳 docker-compose.yml              # Docker编排文件
├── 📄 README.md                       # 项目说明文档
└── 📜 LICENSE                         # 开源许可证
```

## 📚 Go语言学习文档系列

本项目包含完整的Go语言学习文档系列，共14章节，从基础到高级全面覆盖：

### 📖 学习路径

#### 🎯 基础篇（4章）
1. **[变量类型与基本语法](trial/01-basics/01-variables-and-types.md)** - Go语言基础语法入门
2. **[控制结构与流程控制](trial/01-basics/02-control-structures.md)** - 条件判断、循环控制
3. **[函数方法与包管理](trial/01-basics/03-functions-and-methods.md)** - 函数定义、方法调用
4. **[包管理与模块化](trial/01-basics/04-packages-and-imports.md)** - Go模块系统详解

#### 🚀 进阶篇（4章）
5. **[结构体与接口](trial/02-advanced/01-structs-and-interfaces.md)** - 面向对象编程
6. **[错误处理与异常管理](trial/02-advanced/02-error-handling.md)** - Go错误处理最佳实践
7. **[并发编程与goroutine](trial/02-advanced/03-concurrency-basics.md)** - Go并发编程核心
8. **[接口设计与多态](trial/02-advanced/04-interface-design-patterns.md)** - 高级接口设计模式

#### 💼 实战篇（4章）
9. **[Gin框架入门与实践](trial/03-practical/01-gin-framework-basics.md)** - Web开发框架
10. **[GORM数据库操作](trial/03-practical/02-gorm-database-operations.md)** - 数据库ORM实践
11. **[Redis缓存应用](trial/03-practical/03-redis-cache-applications.md)** - 缓存系统集成
12. **[消息队列集成](trial/03-practical/04-message-queue-integration.md)** - 异步消息处理

#### 🏗️ 架构篇（1章）
13. **[微服务架构设计](trial/04-architecture/01-microservices-design.md)** - 分布式系统架构

#### 🎯 高级篇（1章）
14. **[生产实践与运维](trial/05-advanced/01-production-practices.md)** - 容器化部署、监控运维

### 📋 学习特色

- **📖 系统性**：从基础到高级，循序渐进的学习路径
- **🎯 实战性**：结合mall-go项目的真实应用场景
- **🔍 对比性**：与Java、Python等语言的详细对比
- **💡 面试导向**：包含面试常考点和标准答案
- **⚠️ 踩坑指南**：真实的生产环境问题和解决方案
- **🏋️ 练习巩固**：每章包含2-3道实战练习题

## 🚀 快速开始

### 🔧 环境要求

#### 后端环境
- **Go**: 1.21+
- **MySQL**: 8.0+
- **Redis**: 6.0+ (可选)
- **Docker**: 20.0+ (可选)

#### 前端环境
- **Node.js**: 18+
- **npm/yarn**: 最新版本

### 📦 一键启动（推荐）

使用我们提供的快速启动脚本：

```bash
# 克隆项目
git clone https://github.com/Akarin-Akari/go-mall.git
cd go-mall

# 一键启动（包含数据库初始化）
chmod +x quick-start.sh
./quick-start.sh
```

### 🔨 手动安装

#### 1. 后端安装

```bash
# 进入后端目录
cd mall-go

# 安装Go依赖
go mod tidy

# 创建MySQL数据库
mysql -u root -p
CREATE DATABASE mall_go CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 修改配置文件
cp configs/config.yaml.example configs/config.yaml
# 编辑configs/config.yaml，配置数据库连接信息

# 运行数据库迁移
go run cmd/migrate/main.go

# 启动后端服务
go run cmd/server/main.go
```

#### 2. 前端安装

```bash
# 进入前端目录
cd mall-frontend

# 安装依赖
npm install
# 或使用yarn
yarn install

# 启动开发服务器
npm run dev
# 或使用yarn
yarn dev
```

### 🐳 Docker部署

```bash
# 使用Docker Compose一键部署
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

## 🌐 访问地址

### 前端应用
- **开发环境**: http://localhost:3000
- **生产环境**: https://your-domain.com

### 后端API
- **开发环境**: http://localhost:8080
- **API文档**: http://localhost:8080/swagger/index.html
- **健康检查**: http://localhost:8080/health

### 数据库管理
- **phpMyAdmin**: http://localhost:8081 (Docker部署时)
- **Redis Commander**: http://localhost:8082 (Docker部署时)

## 📚 API 文档

### Swagger文档
启动后端服务后，访问 Swagger 在线文档：
- **地址**: http://localhost:8080/swagger/index.html
- **功能**: 在线API测试、参数说明、响应示例

### API概览

#### 用户相关
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `GET /api/v1/user/profile` - 获取用户信息
- `PUT /api/v1/user/profile` - 更新用户信息

#### 商品相关
- `GET /api/v1/products` - 获取商品列表
- `GET /api/v1/products/:id` - 获取商品详情
- `POST /api/v1/products` - 创建商品（管理员）
- `PUT /api/v1/products/:id` - 更新商品（管理员）

#### 购物车相关
- `GET /api/v1/cart` - 获取购物车
- `POST /api/v1/cart/items` - 添加商品到购物车
- `PUT /api/v1/cart/items/:id` - 更新购物车商品
- `DELETE /api/v1/cart/items/:id` - 删除购物车商品

#### 订单相关
- `POST /api/v1/orders` - 创建订单
- `GET /api/v1/orders` - 获取订单列表
- `GET /api/v1/orders/:id` - 获取订单详情
- `PUT /api/v1/orders/:id/status` - 更新订单状态

## 🔧 开发指南

### 🏗️ 后端开发

#### 添加新的API接口

1. **定义数据模型** - 在 `mall-go/internal/model/` 中定义数据结构
```go
type Product struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"not null"`
    Price       decimal.Decimal `json:"price" gorm:"type:decimal(10,2)"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

2. **实现数据访问层** - 在 `mall-go/internal/repository/` 中实现数据操作
```go
type ProductRepository interface {
    Create(product *model.Product) error
    GetByID(id uint) (*model.Product, error)
    List(offset, limit int) ([]*model.Product, error)
}
```

3. **实现业务逻辑层** - 在 `mall-go/internal/service/` 中实现业务逻辑
```go
type ProductService struct {
    repo repository.ProductRepository
}

func (s *ProductService) CreateProduct(req *CreateProductRequest) error {
    // 业务逻辑处理
}
```

4. **实现HTTP处理器** - 在 `mall-go/internal/handler/` 中实现API接口
```go
func (h *ProductHandler) CreateProduct(c *gin.Context) {
    // HTTP请求处理
}
```

5. **注册路由** - 在 `mall-go/internal/handler/routes.go` 中注册路由
```go
v1.POST("/products", productHandler.CreateProduct)
```

#### 数据库操作

```bash
# 数据库迁移
go run mall-go/cmd/migrate/main.go

# 初始化测试数据
go run mall-go/scripts/init_test_data.go

# 数据库连接测试
go run mall-go/debug_db_connection.go
```

### 🌐 前端开发

#### 添加新页面

1. **创建页面组件** - 在 `mall-frontend/src/app/` 中创建页面
```typescript
// mall-frontend/src/app/products/page.tsx
export default function ProductsPage() {
  return <div>Products Page</div>
}
```

2. **创建业务组件** - 在 `mall-frontend/src/components/business/` 中创建组件
```typescript
// mall-frontend/src/components/business/ProductCard.tsx
interface ProductCardProps {
  product: Product;
}

export const ProductCard: React.FC<ProductCardProps> = ({ product }) => {
  return <div>{product.name}</div>
}
```

3. **状态管理** - 在 `mall-frontend/src/store/slices/` 中管理状态
```typescript
// mall-frontend/src/store/slices/productSlice.ts
export const productSlice = createSlice({
  name: 'product',
  initialState,
  reducers: {
    setProducts: (state, action) => {
      state.products = action.payload
    }
  }
})
```

4. **API调用** - 在 `mall-frontend/src/utils/` 中封装API
```typescript
// mall-frontend/src/utils/api.ts
export const productAPI = {
  getProducts: () => httpClient.get('/api/v1/products'),
  getProduct: (id: number) => httpClient.get(`/api/v1/products/${id}`)
}
```

## 🧪 测试指南

### 后端测试

```bash
# 运行所有测试
cd mall-go
go test ./...

# 运行特定模块测试
go test ./internal/handler/...

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 运行集成测试
go test ./tests/integration/...
```

### 前端测试

```bash
# 运行单元测试
cd mall-frontend
npm test

# 运行测试并生成覆盖率报告
npm run test:coverage

# 运行E2E测试
npm run test:e2e
```

### API测试

```bash
# 使用提供的测试脚本
cd scripts
./run-all-tests.sh

# 手动API测试
cd mall-go
go run test_api.go
```

## 📝 功能模块

### 🔐 用户模块
- [x] 用户注册（邮箱验证）
- [x] 用户登录（JWT认证）
- [x] 用户信息管理
- [x] 权限控制（RBAC）
- [x] 密码重置
- [x] 用户头像上传
- [x] 登录日志记录

### 🛍️ 商品模块
- [x] 商品管理（CRUD）
- [x] 商品分类管理
- [x] 商品图片上传
- [x] 库存管理
- [x] 商品搜索（全文检索）
- [x] 商品推荐算法
- [x] 价格历史记录
- [x] 商品评价系统

### 🛒 购物车模块
- [x] 购物车管理
- [x] 商品数量调整
- [x] 购物车同步（登录用户）
- [x] 购物车持久化
- [x] 批量操作
- [x] 购物车推荐

### 📦 订单模块
- [x] 订单创建
- [x] 订单状态管理
- [x] 订单查询（多条件筛选）
- [x] 订单详情
- [x] 订单取消
- [x] 订单退款
- [ ] 支付集成（支付宝/微信）
- [ ] 物流跟踪

### 💰 支付模块
- [x] 支付接口设计
- [x] 支付方式管理
- [ ] 支付宝集成
- [ ] 微信支付集成
- [ ] 支付回调处理
- [ ] 支付安全验证

### 🏪 系统功能
- [x] JWT 认证
- [x] RBAC权限管理
- [x] 结构化日志记录
- [x] 配置管理（多环境）
- [x] 文件上传（本地/云存储）
- [x] Redis缓存管理
- [x] 数据库连接池
- [x] API限流
- [x] 跨域处理
- [x] 健康检查接口

### 📊 监控运维
- [x] Prometheus指标监控
- [x] 链路追踪集成
- [x] 性能分析工具
- [x] 错误日志收集
- [x] 数据库监控
- [x] 缓存监控
- [ ] 告警系统
- [ ] 自动化部署

### 🌐 前端功能
- [x] 响应式设计
- [x] 用户认证界面
- [x] 商品展示页面
- [x] 购物车功能
- [x] 订单管理
- [x] 用户中心
- [x] 搜索功能
- [x] 商品筛选
- [x] 图片懒加载
- [x] 无限滚动
- [x] PWA支持
- [ ] 移动端适配优化

## 🚀 部署指南

### 🐳 Docker部署（推荐）

```bash
# 1. 克隆项目
git clone https://github.com/Akarin-Akari/go-mall.git
cd go-mall

# 2. 配置环境变量
cp .env.example .env
# 编辑.env文件，配置数据库等信息

# 3. 启动所有服务
docker-compose up -d

# 4. 查看服务状态
docker-compose ps

# 5. 查看日志
docker-compose logs -f mall-go
```

### ☸️ Kubernetes部署

```bash
# 1. 创建命名空间
kubectl create namespace mall-go

# 2. 应用配置文件
kubectl apply -f k8s/

# 3. 查看部署状态
kubectl get pods -n mall-go

# 4. 查看服务
kubectl get svc -n mall-go
```

### 🌐 传统部署

详细的传统部署指南请参考：[部署文档](docs/deployment/)

## 📈 性能优化

### 后端优化
- **数据库优化**：索引优化、查询优化、连接池配置
- **缓存策略**：Redis缓存、本地缓存、CDN缓存
- **并发优化**：Goroutine池、连接复用
- **内存优化**：对象池、内存复用

### 前端优化
- **代码分割**：路由懒加载、组件懒加载
- **资源优化**：图片压缩、静态资源CDN
- **缓存策略**：浏览器缓存、Service Worker
- **性能监控**：Web Vitals、性能分析

## 🔒 安全特性

- **认证安全**：JWT Token、密码加密、登录限制
- **权限控制**：RBAC权限模型、API权限验证
- **数据安全**：SQL注入防护、XSS防护、CSRF防护
- **传输安全**：HTTPS加密、API签名验证
- **审计日志**：操作日志记录、安全事件监控

## 🤝 贡献指南

我们欢迎所有形式的贡献！无论是bug修复、功能增强、文档改进还是问题反馈。

### 🔧 开发贡献

1. **Fork项目** - 点击右上角Fork按钮
2. **创建分支** - `git checkout -b feature/AmazingFeature`
3. **提交代码** - `git commit -m 'Add some AmazingFeature'`
4. **推送分支** - `git push origin feature/AmazingFeature`
5. **创建PR** - 打开Pull Request

### 📝 文档贡献

- 改进现有文档
- 添加使用示例
- 翻译文档到其他语言
- 修复文档中的错误

### 🐛 问题反馈

- 使用[Issue模板](../../issues/new/choose)报告bug
- 提供详细的复现步骤
- 包含相关的日志信息
- 说明你的环境信息

### 💡 功能建议

- 在[Discussions](../../discussions)中讨论新功能
- 提供详细的功能描述
- 说明使用场景和价值

## 📊 项目统计

- **代码行数**: 70,000+ 行
- **文档字数**: 100,000+ 字
- **测试覆盖率**: 85%+
- **API接口**: 50+ 个
- **学习章节**: 14 章
- **支持语言**: 中文、英文

## 🏆 致谢

感谢所有为这个项目做出贡献的开发者！

### 核心贡献者
- [@Akarin-Akari](https://github.com/Akarin-Akari) - 项目创建者和主要维护者

### 技术支持
- [Gin Framework](https://github.com/gin-gonic/gin) - 高性能Web框架
- [GORM](https://gorm.io/) - 优秀的Go ORM库
- [React](https://reactjs.org/) - 现代化前端框架

## 📄 许可证

本项目采用 [MIT 许可证](LICENSE) - 查看LICENSE文件了解详情

## 🆘 获取帮助

如果你在使用过程中遇到任何问题，可以通过以下方式获取帮助：

### 📚 文档资源
- **[Go学习文档](trial/README.md)** - 系统性Go语言学习资料
- **[API文档](http://localhost:8080/swagger/index.html)** - 在线API文档
- **[部署指南](docs/deployment/)** - 详细部署说明

### 💬 社区支持
- **[GitHub Issues](../../issues)** - 问题反馈和bug报告
- **[GitHub Discussions](../../discussions)** - 功能讨论和经验分享
- **[项目Wiki](../../wiki)** - 详细的使用指南

### 📧 联系方式
- **项目维护者**: [@Akarin-Akari](https://github.com/Akarin-Akari)
- **邮箱**: akarinzhang@foxmail.com

---

## 🌟 Star History

如果这个项目对你有帮助，请给我们一个⭐️！

[![Star History Chart](https://api.star-history.com/svg?repos=Akarin-Akari/go-mall&type=Date)](https://star-history.com/#Akarin-Akari/go-mall&Date)

---

<div align="center">

**🎉 Happy Coding! 让我们一起用Go语言构建更美好的世界！ 🚀**

Made with ❤️ by [Akarin-Akari](https://github.com/Akarin-Akari)

</div>
