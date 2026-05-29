# Mall Frontend - Go商城前端应用

基于React + Next.js + TypeScript构建的现代化商城前端应用，与Go后端API完美集成，支持完整的电商功能和移动端扩展。

## 🚀 技术栈

### 核心框架

- **React 18** - 现代化React框架，支持并发特性
- **Next.js 15** - 全栈React框架，支持SSR/SSG
- **TypeScript 5.0+** - 类型安全的JavaScript超集
- **Tailwind CSS** - 原子化CSS框架

### UI组件库

- **Ant Design 5.0** - 企业级UI组件库
- **Ant Design Icons** - 丰富的图标库

### 状态管理

- **Redux Toolkit** - 现代化Redux状态管理
- **Redux Persist** - 状态持久化
- **Zustand** - 轻量级状态管理（备选）

### 数据获取

- **TanStack Query** - 强大的数据获取和缓存库
- **Axios** - HTTP客户端

### 开发工具

- **ESLint** - 代码质量检查
- **Prettier** - 代码格式化
- **Husky** - Git钩子管理
- **lint-staged** - 暂存文件检查
- **Commitlint** - 提交信息规范

## 📁 项目结构

```
mall-frontend/
├── src/
│   ├── app/                    # Next.js App Router
│   │   ├── layout.tsx         # 根布局
│   │   ├── page.tsx           # 首页
│   │   └── globals.css        # 全局样式
│   ├── components/            # 组件目录
│   │   ├── common/           # 通用组件
│   │   ├── business/         # 业务组件
│   │   ├── layout/           # 布局组件
│   │   └── providers/        # 提供者组件
│   ├── hooks/                # 自定义Hooks
│   ├── store/                # Redux状态管理
│   │   ├── slices/          # Redux切片
│   │   └── index.ts         # Store配置
│   ├── services/             # API服务
│   │   └── api.ts           # API接口定义
│   ├── utils/                # 工具函数
│   │   ├── request.ts       # HTTP请求工具
│   │   ├── auth.ts          # 认证工具
│   │   ├── upload.ts        # 文件上传工具
│   │   └── index.ts         # 通用工具
│   ├── types/                # TypeScript类型定义
│   ├── constants/            # 常量定义
│   ├── styles/               # 样式文件
│   └── assets/               # 静态资源
├── public/                   # 公共资源
├── .env.local               # 环境变量
├── .env.example             # 环境变量示例
├── next.config.ts           # Next.js配置
├── tailwind.config.ts       # Tailwind配置
├── tsconfig.json            # TypeScript配置
├── package.json             # 项目依赖
└── README.md                # 项目文档
```

## 🛠️ 开发环境配置

### 环境要求

- Node.js >= 18.0.0
- npm >= 9.0.0 或 yarn >= 1.22.0
- Go后端服务运行在 http://localhost:8081

### 安装依赖

```bash
npm install
# 或
yarn install
```

### 环境变量配置

复制环境变量示例文件并配置：

```bash
cp .env.example .env.local
```

编辑 `.env.local` 文件，配置以下变量：

```env
# API配置
NEXT_PUBLIC_API_BASE_URL=http://localhost:8081
NEXT_PUBLIC_API_TIMEOUT=10000

# 应用配置
NEXT_PUBLIC_APP_NAME=Mall Frontend
NEXT_PUBLIC_APP_VERSION=1.0.0

# 认证配置
NEXT_PUBLIC_JWT_STORAGE_KEY=mall_token
NEXT_PUBLIC_REFRESH_TOKEN_KEY=mall_refresh_token
```

### 启动开发服务器

```bash
npm run dev
# 或
yarn dev
```

访问 http://localhost:3000 查看应用

## 📝 可用脚本

```bash
# 开发模式启动（使用Turbopack）
npm run dev

# 构建生产版本
npm run build

# 启动生产服务器
npm run start

# 代码检查
npm run lint

# 代码检查并修复
npm run lint:fix

# 代码格式化
npm run format

# 代码格式化检查
npm run format:check

# TypeScript类型检查
npm run type-check
```


## 🏭 生产环境配置

- 建议在仓库中添加 `.env.production`，示例：

```env
NODE_ENV=production
NEXT_PUBLIC_API_BASE_URL=https://api.example.com
NEXT_PUBLIC_API_TIMEOUT=10000
NEXT_PUBLIC_APP_NAME=Mall Frontend
NEXT_PUBLIC_APP_VERSION=1.0.0
NEXT_PUBLIC_JWT_STORAGE_KEY=mall_token
NEXT_PUBLIC_REFRESH_TOKEN_KEY=mall_refresh_token
NEXT_PUBLIC_DEBUG=false
NEXT_PUBLIC_LOG_LEVEL=warn
```

- 安全策略（CSP）说明：next.config.ts 已使用 `NEXT_PUBLIC_API_BASE_URL` 生成 `connect-src`，生产环境请确保该变量与后端域名一致，否则将被浏览器拦截。
