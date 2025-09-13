# 第4章：Next.js框架基础与SSR/SSG应用 🚀

> *"Next.js不仅仅是一个React框架，它是现代Web开发的完整解决方案！"* 🎯

## 📚 本章导览

欢迎来到Next.js的精彩世界！在前面的章节中，我们已经掌握了TypeScript和React的核心技能，现在是时候学习如何用Next.js构建生产级的现代Web应用了。

### 🎯 学习目标

通过本章学习，你将掌握：

- **Next.js 15.5.2核心概念** - 理解现代React框架的设计哲学
- **App Router vs Pages Router** - 掌握新旧路由系统的区别和选择
- **服务端渲染(SSR)** - 实现更好的SEO和首屏性能
- **静态站点生成(SSG)** - 构建超快的静态网站
- **增量静态再生(ISR)** - 平衡性能与动态内容
- **API Routes设计** - 全栈开发的完整方案
- **性能优化策略** - 企业级应用的最佳实践

### 🛠️ 技术栈概览

基于Mall-Frontend项目，我们将学习：

```json
{
  "framework": "Next.js 15.5.2",
  "runtime": "React 19.1.0",
  "language": "TypeScript 5",
  "styling": "Ant Design 5.27.1 + CSS",
  "state": "Redux Toolkit + Zustand",
  "data": "React Query + Axios",
  "deployment": "Vercel/Docker"
}
```

---

## 🌟 Next.js简介：为什么选择Next.js？

### 传统React SPA的痛点

还记得我们在传统React应用中遇到的问题吗？🤔

```typescript
// 传统React SPA的问题
const TraditionalReactApp = () => {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // 客户端数据获取 - SEO不友好
    fetchData().then(result => {
      setData(result);
      setLoading(false);
    });
  }, []);

  // 首屏白屏时间长
  if (loading) return <div>Loading...</div>;

  return <div>{data?.content}</div>;
};
```

**主要问题**：
1. **SEO困难** - 搜索引擎看到的是空白页面
2. **首屏性能差** - 需要等待JS加载和执行
3. **路由复杂** - 需要手动配置React Router
4. **代码分割麻烦** - 手动实现懒加载
5. **API开发分离** - 前后端完全分离增加复杂度

### Next.js的解决方案

Next.js就像是React的"超级英雄版本"！🦸‍♂️

```typescript
// Next.js的优雅解决方案
// app/products/[id]/page.tsx
export default async function ProductPage({ params }: { params: { id: string } }) {
  // 服务端数据获取 - SEO友好
  const product = await fetchProduct(params.id);

  // 直接返回渲染好的HTML
  return (
    <div>
      <h1>{product.name}</h1>
      <p>{product.description}</p>
    </div>
  );
}

// 自动代码分割，无需配置
// 自动路由，基于文件系统
// 内置API Routes
```

### Next.js核心优势

#### 1. 🚀 多种渲染模式

```typescript
// SSR - 服务端渲染
export default async function SSRPage() {
  const data = await fetch('https://api.example.com/data');
  return <div>{data}</div>;
}

// SSG - 静态站点生成
export async function generateStaticParams() {
  return [{ id: '1' }, { id: '2' }];
}

// ISR - 增量静态再生
export const revalidate = 60; // 60秒后重新生成
```

#### 2. 🎯 零配置开发体验

```bash
# 创建Next.js项目
npx create-next-app@latest my-app --typescript --tailwind --eslint

# 启动开发服务器
npm run dev

# 自动获得：
# ✅ TypeScript支持
# ✅ 热重载
# ✅ 代码分割
# ✅ 图片优化
# ✅ 字体优化
# ✅ SEO优化
```

#### 3. 🔧 内置性能优化

```typescript
// 自动图片优化
import Image from 'next/image';

export default function ProductCard({ product }) {
  return (
    <div>
      <Image
        src={product.image}
        alt={product.name}
        width={300}
        height={200}
        priority // 优先加载
        placeholder="blur" // 模糊占位符
      />
    </div>
  );
}

// 自动字体优化
import { Inter } from 'next/font/google';

const inter = Inter({ subsets: ['latin'] });

export default function Layout({ children }) {
  return (
    <html className={inter.className}>
      <body>{children}</body>
    </html>
  );
}
```

---

## 🆚 App Router vs Pages Router：新时代的选择

Next.js 13引入了革命性的App Router，让我们看看它与传统Pages Router的区别：

### Pages Router（传统方式）

```typescript
// pages/products/[id].tsx
import { GetServerSideProps } from 'next';

interface Props {
  product: Product;
}

export default function ProductPage({ product }: Props) {
  return <div>{product.name}</div>;
}

export const getServerSideProps: GetServerSideProps = async ({ params }) => {
  const product = await fetchProduct(params?.id as string);

  return {
    props: {
      product,
    },
  };
};
```

### App Router（现代方式）

```typescript
// app/products/[id]/page.tsx
interface Props {
  params: { id: string };
}

export default async function ProductPage({ params }: Props) {
  // 直接在组件中获取数据
  const product = await fetchProduct(params.id);

  return <div>{product.name}</div>;
}

// 生成元数据
export async function generateMetadata({ params }: Props) {
  const product = await fetchProduct(params.id);

  return {
    title: product.name,
    description: product.description,
  };
}
```

### 对比分析

| 特性 | Pages Router | App Router |
|------|-------------|------------|
| **文件位置** | `pages/` | `app/` |
| **数据获取** | `getServerSideProps` | `async/await` |
| **布局** | `_app.tsx` + `_document.tsx` | `layout.tsx` |
| **元数据** | `Head` 组件 | `generateMetadata` |
| **错误处理** | `_error.tsx` | `error.tsx` |
| **加载状态** | 手动实现 | `loading.tsx` |
| **嵌套布局** | 复杂 | 原生支持 |
| **流式渲染** | 不支持 | 支持 |

### 为什么选择App Router？

1. **更简洁的API** - 减少样板代码
2. **更好的TypeScript支持** - 类型推断更准确
3. **原生嵌套布局** - 复杂UI结构更容易管理
4. **流式渲染** - 更好的用户体验
5. **未来趋势** - Next.js团队重点发展方向

---

## 🏗️ Mall-Frontend项目结构深度解析

让我们深入分析Mall-Frontend项目的Next.js应用结构：

### 项目目录结构

```
mall-frontend/
├── src/
│   ├── app/                    # App Router目录
│   │   ├── layout.tsx         # 根布局
│   │   ├── page.tsx           # 首页
│   │   ├── globals.css        # 全局样式
│   │   ├── products/          # 商品相关页面
│   │   │   ├── page.tsx       # 商品列表页
│   │   │   └── [id]/          # 动态路由
│   │   │       └── page.tsx   # 商品详情页
│   │   ├── cart/              # 购物车页面
│   │   ├── checkout/          # 结算页面
│   │   ├── login/             # 登录页面
│   │   └── register/          # 注册页面
│   ├── components/            # 组件库
│   ├── store/                 # 状态管理
│   ├── types/                 # 类型定义
│   └── utils/                 # 工具函数
├── next.config.ts             # Next.js配置
├── package.json               # 依赖管理
└── tsconfig.json             # TypeScript配置
```

### 根布局分析

<augment_code_snippet path="mall-frontend/src/app/layout.tsx" mode="EXCERPT">
````typescript
import type { Metadata } from "next";
import { AntdRegistry } from '@ant-design/nextjs-registry';
import AppProviders from '@/components/providers/AppProviders';
import "./globals.css";

export const metadata: Metadata = {
  title: "Mall Frontend - Go商城前端应用",
  description: "基于React + Next.js + TypeScript构建的现代化商城前端应用",
  keywords: "商城,电商,React,Next.js,TypeScript",
  authors: [{ name: "Mall Team" }],
  viewport: "width=device-width, initial-scale=1",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-CN">
      <body>
        <AntdRegistry>
          <AppProviders>
            {children}
          </AppProviders>
        </AntdRegistry>
      </body>
    </html>
  );
}
````
</augment_code_snippet>

**设计亮点分析**：

1. **元数据管理** - 使用`Metadata` API统一管理SEO信息
2. **样式隔离** - Ant Design的SSR支持
3. **Provider模式** - 统一的状态管理和主题提供
4. **国际化准备** - `lang="zh-CN"`属性

### 首页实现分析

<augment_code_snippet path="mall-frontend/src/app/page.tsx" mode="EXCERPT">
````typescript
'use client';

import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import MainLayout from '@/components/layout/MainLayout';
import ProductCard from '@/components/business/ProductCard';
import { useAppDispatch, useAppSelector } from '@/store';

export default function Home() {
  const router = useRouter();
  const dispatch = useAppDispatch();

  const products = useAppSelector(selectProducts) || [];
  const categories = useAppSelector(selectCategories) || [];

  // 获取数据
  useEffect(() => {
    dispatch(fetchProductsAsync({ page: 1, page_size: 20 }));
    dispatch(fetchCategoriesAsync());
  }, [dispatch]);

  return (
    <MainLayout>
      {/* 轮播图、商品展示等 */}
    </MainLayout>
  );
}
````
</augment_code_snippet>

**关键特性**：

1. **'use client'指令** - 标识客户端组件
2. **Next.js导航** - 使用`useRouter`进行路由跳转
3. **状态管理集成** - Redux Toolkit的完美集成
4. **组件化设计** - 高度可复用的组件架构

### 动态路由实现

<augment_code_snippet path="mall-frontend/src/app/products/[id]/page.tsx" mode="EXCERPT">
````typescript
'use client';

interface Props {
  params: { id: string };
}

const ProductDetailPage: React.FC = () => {
  const params = useParams();
  const dispatch = useAppDispatch();
  const { currentProduct, productLoading } = useAppSelector(selectProduct);

  const productId = parseInt(params?.id as string);

  // 加载商品详情
  useEffect(() => {
    if (productId) {
      dispatch(fetchProductDetailAsync(productId));
    }
  }, [dispatch, productId]);

  if (productLoading) {
    return (
      <MainLayout>
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
          <Spin size="large" />
        </div>
      </MainLayout>
    );
  }

  return (
    <MainLayout>
      {/* 商品详情内容 */}
    </MainLayout>
  );
};
````
</augment_code_snippet>

**动态路由特点**：

1. **文件系统路由** - `[id]`文件夹自动创建动态路由
2. **参数获取** - 通过`useParams`获取路由参数
3. **类型安全** - TypeScript确保参数类型正确
4. **加载状态** - 优雅的加载和错误处理

---

## ⚙️ Next.js配置深度解析

让我们深入分析Mall-Frontend的Next.js配置：

<augment_code_snippet path="mall-frontend/next.config.ts" mode="EXCERPT">
````typescript
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // API代理配置
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8080/api/:path*', // Go后端API地址
      },
    ];
  },

  // 安全头配置
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          {
            key: 'Content-Security-Policy',
            value: [
              "default-src 'self'",
              "script-src 'self' 'unsafe-eval' 'unsafe-inline'",
              "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
              "font-src 'self' https://fonts.gstatic.com",
              "img-src 'self' data: https: blob:",
              "connect-src 'self' http://localhost:8080 ws://localhost:3000",
            ].join('; ')
          },
        ]
      }
    ];
  },

  // 图片优化配置
  images: {
    domains: ['localhost', '127.0.0.1'],
    formats: ['image/webp', 'image/avif'],
  },

  // 实验性功能
  experimental: {
    turbo: {
      rules: {
        '*.svg': {
          loaders: ['@svgr/webpack'],
          as: '*.js',
        },
      },
    },
  },

  // 编译配置
  compiler: {
    removeConsole: process.env.NODE_ENV === 'production',
  },

  // 输出配置
  output: 'standalone',
};
````
</augment_code_snippet>

### 配置详解

#### 1. API代理配置

```typescript
async rewrites() {
  return [
    {
      source: '/api/:path*',
      destination: 'http://localhost:8080/api/:path*',
    },
  ];
}
```

**作用**：
- 解决跨域问题
- 统一API入口
- 开发环境代理到Go后端

**与传统方案对比**：
```typescript
// 传统方式：需要配置webpack dev server
module.exports = {
  devServer: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
};

// Next.js方式：内置支持，配置简单
```

#### 2. 安全头配置

```typescript
async headers() {
  return [
    {
      source: '/(.*)',
      headers: [
        {
          key: 'Content-Security-Policy',
          value: "default-src 'self'; script-src 'self' 'unsafe-eval'..."
        },
        {
          key: 'X-Frame-Options',
          value: 'DENY'
        },
      ]
    }
  ];
}
```

**安全策略**：
- **CSP** - 防止XSS攻击
- **X-Frame-Options** - 防止点击劫持
- **X-Content-Type-Options** - 防止MIME类型嗅探
- **HSTS** - 强制HTTPS连接

#### 3. 图片优化配置

```typescript
images: {
  domains: ['localhost', '127.0.0.1'],
  formats: ['image/webp', 'image/avif'],
}
```

**优化特性**：
- 自动格式转换（WebP、AVIF）
- 响应式图片
- 懒加载
- 占位符支持

#### 4. 实验性功能

```typescript
experimental: {
  turbo: {
    rules: {
      '*.svg': {
        loaders: ['@svgr/webpack'],
        as: '*.js',
      },
    },
  },
}
```

**Turbopack优势**：
- 比Webpack快700倍的构建速度
- 增量编译
- 更好的缓存策略

---

## 📊 数据获取策略：SSR、SSG、ISR、CSR全解析

Next.js提供了四种主要的数据获取策略，每种都有其适用场景。让我们通过Mall-Frontend的实际案例来深入理解：

### 1. 🔄 SSR (Server-Side Rendering) - 服务端渲染

**适用场景**：需要实时数据、SEO要求高、个性化内容

```typescript
// app/products/page.tsx - 商品列表页
interface SearchParams {
  category?: string;
  page?: string;
  sort?: string;
}

export default async function ProductsPage({
  searchParams,
}: {
  searchParams: SearchParams;
}) {
  // 服务端获取数据
  const products = await fetchProducts({
    category: searchParams.category,
    page: parseInt(searchParams.page || '1'),
    sort: searchParams.sort || 'created_at',
  });

  const categories = await fetchCategories();

  return (
    <div>
      <ProductFilter categories={categories} />
      <ProductGrid products={products} />
      <Pagination total={products.total} />
    </div>
  );
}

// 数据获取函数
async function fetchProducts(params: {
  category?: string;
  page: number;
  sort: string;
}) {
  const searchParams = new URLSearchParams({
    page: params.page.toString(),
    page_size: '20',
    sort: params.sort,
    ...(params.category && { category: params.category }),
  });

  const response = await fetch(
    `${process.env.API_BASE_URL}/api/products?${searchParams}`,
    {
      // 禁用缓存，确保数据实时性
      cache: 'no-store',
      headers: {
        'Content-Type': 'application/json',
      },
    }
  );

  if (!response.ok) {
    throw new Error('Failed to fetch products');
  }

  return response.json();
}
```

**SSR优势**：
- ✅ SEO友好 - 搜索引擎能看到完整内容
- ✅ 首屏快速 - 服务端渲染完整HTML
- ✅ 实时数据 - 每次请求都获取最新数据
- ✅ 社交分享 - 完整的meta标签

**与传统SPA对比**：
```typescript
// 传统SPA方式
const ProductsPage = () => {
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // 客户端获取数据 - SEO不友好
    fetchProducts().then(data => {
      setProducts(data);
      setLoading(false);
    });
  }, []);

  if (loading) return <div>Loading...</div>; // 白屏时间

  return <ProductGrid products={products} />;
};

// Next.js SSR方式
export default async function ProductsPage() {
  // 服务端获取数据 - SEO友好
  const products = await fetchProducts();

  // 直接返回渲染好的HTML
  return <ProductGrid products={products} />;
}
```

### 2. 🏗️ SSG (Static Site Generation) - 静态站点生成

**适用场景**：内容相对静态、性能要求极高、CDN分发

```typescript
// app/blog/[slug]/page.tsx - 博客文章页
interface Props {
  params: { slug: string };
}

// 生成静态参数
export async function generateStaticParams() {
  const posts = await fetch(`${process.env.API_BASE_URL}/api/blog/posts`).then(
    res => res.json()
  );

  return posts.map((post: { slug: string }) => ({
    slug: post.slug,
  }));
}

// 生成元数据
export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const post = await fetchPost(params.slug);

  return {
    title: post.title,
    description: post.excerpt,
    openGraph: {
      title: post.title,
      description: post.excerpt,
      images: [post.featured_image],
    },
  };
}

// 静态页面组件
export default async function BlogPost({ params }: Props) {
  const post = await fetchPost(params.slug);

  return (
    <article>
      <header>
        <h1>{post.title}</h1>
        <time>{post.published_at}</time>
      </header>
      <div dangerouslySetInnerHTML={{ __html: post.content }} />
    </article>
  );
}

async function fetchPost(slug: string) {
  const response = await fetch(
    `${process.env.API_BASE_URL}/api/blog/posts/${slug}`,
    {
      // 构建时缓存
      cache: 'force-cache',
    }
  );

  if (!response.ok) {
    throw new Error('Post not found');
  }

  return response.json();
}
```

**SSG优势**：
- ⚡ 极快加载 - 预构建的静态HTML
- 💰 成本低 - CDN缓存，服务器压力小
- 🔒 安全性高 - 没有服务端运行时
- 📈 SEO完美 - 静态HTML对搜索引擎最友好

### 3. 🔄 ISR (Incremental Static Regeneration) - 增量静态再生

**适用场景**：内容更新不频繁、需要平衡性能与实时性

```typescript
// app/products/[id]/page.tsx - 商品详情页
interface Props {
  params: { id: string };
}

// 设置重新验证时间
export const revalidate = 3600; // 1小时后重新生成

export async function generateStaticParams() {
  // 只预生成热门商品
  const hotProducts = await fetch(
    `${process.env.API_BASE_URL}/api/products/hot`
  ).then(res => res.json());

  return hotProducts.slice(0, 100).map((product: { id: number }) => ({
    id: product.id.toString(),
  }));
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const product = await fetchProduct(params.id);

  return {
    title: `${product.name} - Mall商城`,
    description: product.description,
    openGraph: {
      title: product.name,
      description: product.description,
      images: product.images,
    },
  };
}

export default async function ProductDetail({ params }: Props) {
  const product = await fetchProduct(params.id);
  const relatedProducts = await fetchRelatedProducts(product.category_id);

  return (
    <div>
      <ProductInfo product={product} />
      <ProductSpecs specs={product.specs} />
      <ProductReviews productId={product.id} />
      <RelatedProducts products={relatedProducts} />
    </div>
  );
}

async function fetchProduct(id: string) {
  const response = await fetch(
    `${process.env.API_BASE_URL}/api/products/${id}`,
    {
      // ISR缓存策略
      next: { revalidate: 3600 }, // 1小时后重新验证
    }
  );

  if (!response.ok) {
    throw new Error('Product not found');
  }

  return response.json();
}
```

**ISR工作原理**：
1. **首次访问** - 返回静态页面（如果已生成）
2. **后台重新生成** - 到达revalidate时间后，后台重新获取数据
3. **更新缓存** - 新页面生成后替换旧缓存
4. **按需生成** - 未预生成的页面在首次访问时生成

### 4. 🖥️ CSR (Client-Side Rendering) - 客户端渲染

**适用场景**：用户交互频繁、个人数据、实时更新

```typescript
// app/dashboard/page.tsx - 用户仪表板
'use client';

import { useEffect, useState } from 'react';
import { useUser } from '@/hooks/useUser';

export default function Dashboard() {
  const { user, loading: userLoading } = useUser();
  const [orders, setOrders] = useState([]);
  const [favorites, setFavorites] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (user) {
      // 客户端获取用户相关数据
      Promise.all([
        fetchUserOrders(user.id),
        fetchUserFavorites(user.id),
      ]).then(([ordersData, favoritesData]) => {
        setOrders(ordersData);
        setFavorites(favoritesData);
        setLoading(false);
      });
    }
  }, [user]);

  if (userLoading || loading) {
    return <DashboardSkeleton />;
  }

  return (
    <div>
      <UserProfile user={user} />
      <OrderHistory orders={orders} />
      <FavoriteProducts products={favorites} />
    </div>
  );
}

// 使用React Query优化CSR
import { useQuery } from '@tanstack/react-query';

export default function OptimizedDashboard() {
  const { user } = useUser();

  const { data: orders, isLoading: ordersLoading } = useQuery({
    queryKey: ['orders', user?.id],
    queryFn: () => fetchUserOrders(user!.id),
    enabled: !!user,
    staleTime: 5 * 60 * 1000, // 5分钟内不重新获取
  });

  const { data: favorites, isLoading: favoritesLoading } = useQuery({
    queryKey: ['favorites', user?.id],
    queryFn: () => fetchUserFavorites(user!.id),
    enabled: !!user,
  });

  return (
    <div>
      {ordersLoading ? <OrdersSkeleton /> : <OrderHistory orders={orders} />}
      {favoritesLoading ? <FavoritesSkeleton /> : <FavoriteProducts products={favorites} />}
    </div>
  );
}
```

### 数据获取策略选择指南

| 场景 | 推荐策略 | 原因 |
|------|----------|------|
| 商品列表页 | SSR | 需要SEO，数据实时性要求高 |
| 商品详情页 | ISR | 平衡SEO和性能，内容相对稳定 |
| 用户仪表板 | CSR | 个人数据，需要认证 |
| 博客文章 | SSG | 内容静态，性能要求高 |
| 搜索结果 | SSR | 需要SEO，查询参数动态 |
| 购物车 | CSR | 用户交互频繁，实时更新 |

---

## 🛣️ 路由系统深度解析

Next.js的文件系统路由是其最强大的特性之一。让我们深入了解Mall-Frontend的路由设计：

### 基础路由

```
app/
├── page.tsx                    # / (首页)
├── products/
│   ├── page.tsx               # /products (商品列表)
│   └── [id]/
│       └── page.tsx           # /products/[id] (商品详情)
├── cart/
│   └── page.tsx               # /cart (购物车)
├── checkout/
│   └── page.tsx               # /checkout (结算)
├── login/
│   └── page.tsx               # /login (登录)
└── register/
    └── page.tsx               # /register (注册)
```

### 动态路由实现

```typescript
// app/products/[id]/page.tsx
interface Props {
  params: { id: string };
  searchParams: { [key: string]: string | string[] | undefined };
}

export default async function ProductDetail({ params, searchParams }: Props) {
  const productId = parseInt(params.id);
  const variant = searchParams.variant as string;

  // 根据路由参数获取数据
  const product = await fetchProduct(productId);

  return (
    <div>
      <h1>{product.name}</h1>
      {variant && <p>选中规格: {variant}</p>}
    </div>
  );
}
```

### 路由组和布局

```typescript
// app/(shop)/layout.tsx - 商店布局
export default function ShopLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div>
      <ShopHeader />
      <ShopSidebar />
      <main>{children}</main>
      <ShopFooter />
    </div>
  );
}

// app/(shop)/products/page.tsx - 继承商店布局
// app/(shop)/cart/page.tsx - 继承商店布局

// app/(auth)/layout.tsx - 认证布局
export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="auth-container">
      <AuthHeader />
      <div className="auth-content">{children}</div>
    </div>
  );
}

// app/(auth)/login/page.tsx - 继承认证布局
// app/(auth)/register/page.tsx - 继承认证布局
```

### 路由拦截和重写

```typescript
// next.config.ts
const nextConfig = {
  async rewrites() {
    return [
      // API代理
      {
        source: '/api/:path*',
        destination: 'http://localhost:8080/api/:path*',
      },
      // 旧路径重定向
      {
        source: '/old-products/:id',
        destination: '/products/:id',
      },
    ];
  },

  async redirects() {
    return [
      // 永久重定向
      {
        source: '/home',
        destination: '/',
        permanent: true,
      },
      // 条件重定向
      {
        source: '/admin/:path*',
        destination: '/login?redirect=/admin/:path*',
        permanent: false,
        has: [
          {
            type: 'cookie',
            key: 'auth-token',
            value: undefined,
          },
        ],
      },
    ];
  },
};
```

### 程序化导航

```typescript
// 使用useRouter进行导航
'use client';

import { useRouter, useSearchParams } from 'next/navigation';

export default function ProductFilter() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const handleCategoryChange = (categoryId: string) => {
    const params = new URLSearchParams(searchParams);
    params.set('category', categoryId);
    params.delete('page'); // 重置页码

    // 更新URL
    router.push(`/products?${params.toString()}`);
  };

  const handleSortChange = (sortBy: string) => {
    const params = new URLSearchParams(searchParams);
    params.set('sort', sortBy);

    // 替换当前历史记录
    router.replace(`/products?${params.toString()}`);
  };

  return (
    <div>
      <CategorySelect onChange={handleCategoryChange} />
      <SortSelect onChange={handleSortChange} />
    </div>
  );
}
```

---

## 🔌 API Routes：全栈开发的完整方案

Next.js的API Routes让我们可以在同一个项目中构建前后端，虽然Mall-Frontend主要使用Go后端，但我们也可以用API Routes处理一些前端特定的逻辑：

### 基础API Routes

```typescript
// app/api/health/route.ts - 健康检查接口
export async function GET() {
  return Response.json({
    status: 'ok',
    timestamp: new Date().toISOString(),
    version: process.env.npm_package_version,
  });
}

// app/api/config/route.ts - 前端配置接口
export async function GET() {
  return Response.json({
    apiBaseUrl: process.env.NEXT_PUBLIC_API_BASE_URL,
    appName: process.env.NEXT_PUBLIC_APP_NAME,
    features: {
      enablePayment: process.env.ENABLE_PAYMENT === 'true',
      enableChat: process.env.ENABLE_CHAT === 'true',
    },
  });
}
```

### 代理API实现

```typescript
// app/api/proxy/[...path]/route.ts - 通用API代理
import { NextRequest } from 'next/server';

const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:8080';

export async function GET(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  return proxyRequest(request, params.path, 'GET');
}

export async function POST(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  return proxyRequest(request, params.path, 'POST');
}

export async function PUT(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  return proxyRequest(request, params.path, 'PUT');
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  return proxyRequest(request, params.path, 'DELETE');
}

async function proxyRequest(
  request: NextRequest,
  pathSegments: string[],
  method: string
) {
  const path = pathSegments.join('/');
  const url = `${API_BASE_URL}/api/${path}`;

  // 获取查询参数
  const searchParams = request.nextUrl.searchParams;
  const queryString = searchParams.toString();
  const finalUrl = queryString ? `${url}?${queryString}` : url;

  try {
    // 转发请求头
    const headers = new Headers();
    request.headers.forEach((value, key) => {
      // 过滤掉一些不需要的头
      if (!['host', 'connection', 'content-length'].includes(key.toLowerCase())) {
        headers.set(key, value);
      }
    });

    // 获取请求体
    const body = method !== 'GET' ? await request.text() : undefined;

    // 发送请求到后端
    const response = await fetch(finalUrl, {
      method,
      headers,
      body,
    });

    // 转发响应
    const responseData = await response.text();

    return new Response(responseData, {
      status: response.status,
      statusText: response.statusText,
      headers: {
        'Content-Type': response.headers.get('Content-Type') || 'application/json',
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
        'Access-Control-Allow-Headers': 'Content-Type, Authorization',
      },
    });
  } catch (error) {
    console.error('Proxy error:', error);
    return Response.json(
      { error: 'Internal Server Error' },
      { status: 500 }
    );
  }
}
```

### 文件上传API

```typescript
// app/api/upload/route.ts - 文件上传接口
import { NextRequest } from 'next/server';
import { writeFile } from 'fs/promises';
import path from 'path';

export async function POST(request: NextRequest) {
  try {
    const formData = await request.formData();
    const file = formData.get('file') as File;

    if (!file) {
      return Response.json({ error: 'No file uploaded' }, { status: 400 });
    }

    // 验证文件类型
    const allowedTypes = ['image/jpeg', 'image/png', 'image/webp'];
    if (!allowedTypes.includes(file.type)) {
      return Response.json({ error: 'Invalid file type' }, { status: 400 });
    }

    // 验证文件大小 (5MB)
    if (file.size > 5 * 1024 * 1024) {
      return Response.json({ error: 'File too large' }, { status: 400 });
    }

    // 生成唯一文件名
    const timestamp = Date.now();
    const extension = path.extname(file.name);
    const filename = `${timestamp}${extension}`;

    // 保存文件
    const bytes = await file.arrayBuffer();
    const buffer = Buffer.from(bytes);
    const uploadPath = path.join(process.cwd(), 'public/uploads', filename);

    await writeFile(uploadPath, buffer);

    return Response.json({
      success: true,
      filename,
      url: `/uploads/${filename}`,
      size: file.size,
      type: file.type,
    });
  } catch (error) {
    console.error('Upload error:', error);
    return Response.json({ error: 'Upload failed' }, { status: 500 });
  }
}
```

### 认证中间件

```typescript
// app/api/auth/middleware.ts - 认证中间件
import { NextRequest } from 'next/server';
import jwt from 'jsonwebtoken';

export interface AuthenticatedRequest extends NextRequest {
  user?: {
    id: number;
    email: string;
    role: string;
  };
}

export function withAuth(handler: (req: AuthenticatedRequest) => Promise<Response>) {
  return async (request: NextRequest) => {
    const token = request.headers.get('Authorization')?.replace('Bearer ', '');

    if (!token) {
      return Response.json({ error: 'Unauthorized' }, { status: 401 });
    }

    try {
      const decoded = jwt.verify(token, process.env.JWT_SECRET!) as any;
      (request as AuthenticatedRequest).user = decoded;

      return handler(request as AuthenticatedRequest);
    } catch (error) {
      return Response.json({ error: 'Invalid token' }, { status: 401 });
    }
  };
}

// 使用认证中间件
// app/api/user/profile/route.ts
export const GET = withAuth(async (request: AuthenticatedRequest) => {
  const user = request.user!;

  // 获取用户资料
  const profile = await getUserProfile(user.id);

  return Response.json(profile);
});
```

---

## 🛡️ 中间件(Middleware)应用

Next.js中间件在请求到达页面之前运行，可以用于认证、重定向、国际化等：

### 认证中间件

```typescript
// middleware.ts
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // 需要认证的路径
  const protectedPaths = ['/dashboard', '/orders', '/profile', '/checkout'];
  const isProtectedPath = protectedPaths.some(path => pathname.startsWith(path));

  if (isProtectedPath) {
    const token = request.cookies.get('auth-token')?.value;

    if (!token) {
      // 重定向到登录页，并保存原始URL
      const loginUrl = new URL('/login', request.url);
      loginUrl.searchParams.set('redirect', pathname);
      return NextResponse.redirect(loginUrl);
    }

    // 验证token（简化版本）
    try {
      // 这里应该验证JWT token
      // const decoded = jwt.verify(token, process.env.JWT_SECRET!);
    } catch (error) {
      // token无效，清除cookie并重定向
      const response = NextResponse.redirect(new URL('/login', request.url));
      response.cookies.delete('auth-token');
      return response;
    }
  }

  // 管理员路径检查
  if (pathname.startsWith('/admin')) {
    const userRole = request.cookies.get('user-role')?.value;

    if (userRole !== 'admin') {
      return NextResponse.redirect(new URL('/unauthorized', request.url));
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    // 匹配所有路径，除了静态文件和API路由
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
};
```

### 国际化中间件

```typescript
// middleware.ts - 国际化支持
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const locales = ['zh-CN', 'en-US', 'ja-JP'];
const defaultLocale = 'zh-CN';

function getLocale(request: NextRequest): string {
  // 1. 检查URL路径中的语言
  const pathname = request.nextUrl.pathname;
  const pathnameLocale = locales.find(
    locale => pathname.startsWith(`/${locale}/`) || pathname === `/${locale}`
  );

  if (pathnameLocale) return pathnameLocale;

  // 2. 检查cookie中的语言设置
  const cookieLocale = request.cookies.get('locale')?.value;
  if (cookieLocale && locales.includes(cookieLocale)) {
    return cookieLocale;
  }

  // 3. 检查Accept-Language头
  const acceptLanguage = request.headers.get('Accept-Language');
  if (acceptLanguage) {
    const browserLocale = acceptLanguage
      .split(',')[0]
      .split('-')[0];

    const matchedLocale = locales.find(locale =>
      locale.startsWith(browserLocale)
    );

    if (matchedLocale) return matchedLocale;
  }

  return defaultLocale;
}

export function middleware(request: NextRequest) {
  const pathname = request.nextUrl.pathname;

  // 检查路径是否已经包含语言前缀
  const pathnameIsMissingLocale = locales.every(
    locale => !pathname.startsWith(`/${locale}/`) && pathname !== `/${locale}`
  );

  if (pathnameIsMissingLocale) {
    const locale = getLocale(request);

    // 重定向到带语言前缀的URL
    const newUrl = new URL(`/${locale}${pathname}`, request.url);
    const response = NextResponse.redirect(newUrl);

    // 设置语言cookie
    response.cookies.set('locale', locale, {
      maxAge: 365 * 24 * 60 * 60, // 1年
      path: '/',
    });

    return response;
  }

  return NextResponse.next();
}
```

### 性能监控中间件

```typescript
// middleware.ts - 性能监控
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  const start = Date.now();

  // 创建响应
  const response = NextResponse.next();

  // 添加性能头
  response.headers.set('X-Response-Time', `${Date.now() - start}ms`);
  response.headers.set('X-Request-ID', crypto.randomUUID());

  // 记录慢请求
  const duration = Date.now() - start;
  if (duration > 1000) {
    console.warn(`Slow request: ${request.url} took ${duration}ms`);
  }

  return response;
}
```

---

## 🎨 布局组件和元数据管理

### 嵌套布局设计

```typescript
// app/layout.tsx - 根布局
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: {
    template: '%s | Mall商城',
    default: 'Mall商城 - 优质商品，优惠价格',
  },
  description: '专业的在线购物平台，提供优质商品和服务',
  keywords: ['商城', '购物', '电商', '优惠'],
  authors: [{ name: 'Mall Team', url: 'https://mall.com' }],
  creator: 'Mall Team',
  publisher: 'Mall Inc.',
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
  openGraph: {
    type: 'website',
    locale: 'zh_CN',
    url: 'https://mall.com',
    siteName: 'Mall商城',
    title: 'Mall商城 - 优质商品，优惠价格',
    description: '专业的在线购物平台，提供优质商品和服务',
    images: [
      {
        url: '/og-image.jpg',
        width: 1200,
        height: 630,
        alt: 'Mall商城',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Mall商城',
    description: '专业的在线购物平台',
    images: ['/twitter-image.jpg'],
  },
  verification: {
    google: 'google-verification-code',
    yandex: 'yandex-verification-code',
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-CN">
      <body>
        <div id="root">
          {children}
        </div>
      </body>
    </html>
  );
}
```

### 商品页面布局

```typescript
// app/(shop)/layout.tsx - 商店布局
import { Suspense } from 'react';
import Header from '@/components/layout/Header';
import Footer from '@/components/layout/Footer';
import Sidebar from '@/components/layout/Sidebar';
import LoadingSpinner from '@/components/ui/LoadingSpinner';

export default function ShopLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="shop-layout">
      <Header />
      <div className="shop-content">
        <aside className="shop-sidebar">
          <Suspense fallback={<LoadingSpinner />}>
            <Sidebar />
          </Suspense>
        </aside>
        <main className="shop-main">
          {children}
        </main>
      </div>
      <Footer />
    </div>
  );
}
```

### 动态元数据生成

```typescript
// app/products/[id]/page.tsx - 商品详情页元数据
export async function generateMetadata({
  params,
}: {
  params: { id: string };
}): Promise<Metadata> {
  const product = await fetchProduct(params.id);

  if (!product) {
    return {
      title: '商品不存在',
      description: '您访问的商品不存在或已下架',
    };
  }

  const price = product.discount_price || product.price;
  const images = product.images?.map(img => ({
    url: img,
    width: 800,
    height: 600,
    alt: product.name,
  })) || [];

  return {
    title: product.name,
    description: product.description,
    keywords: [product.name, product.category_name, '商城', '购物'].filter(Boolean),
    openGraph: {
      title: product.name,
      description: product.description,
      type: 'product',
      images,
      siteName: 'Mall商城',
      locale: 'zh_CN',
    },
    twitter: {
      card: 'summary_large_image',
      title: product.name,
      description: product.description,
      images: images.map(img => img.url),
    },
    other: {
      'product:price:amount': price,
      'product:price:currency': 'CNY',
      'product:availability': product.stock > 0 ? 'in stock' : 'out of stock',
      'product:condition': 'new',
      'product:retailer_item_id': product.id.toString(),
    },
  };
}
```

---

## ⚡ 性能优化策略

### 1. 图片优化

```typescript
// 使用Next.js Image组件
import Image from 'next/image';

export default function ProductCard({ product }: { product: Product }) {
  return (
    <div className="product-card">
      <Image
        src={product.image}
        alt={product.name}
        width={300}
        height={200}
        priority={product.featured} // 首屏图片优先加载
        placeholder="blur" // 模糊占位符
        blurDataURL="data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQ..." // 自定义模糊图片
        sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw" // 响应式尺寸
        style={{
          objectFit: 'cover',
          borderRadius: '8px',
        }}
      />
    </div>
  );
}

// 图片优化配置
// next.config.ts
const nextConfig = {
  images: {
    domains: ['example.com', 'cdn.example.com'],
    formats: ['image/webp', 'image/avif'], // 现代格式
    deviceSizes: [640, 750, 828, 1080, 1200, 1920, 2048, 3840],
    imageSizes: [16, 32, 48, 64, 96, 128, 256, 384],
    minimumCacheTTL: 60 * 60 * 24 * 365, // 1年缓存
  },
};
```

### 2. 字体优化

```typescript
// app/layout.tsx - 字体优化
import { Inter, Noto_Sans_SC } from 'next/font/google';

// 英文字体
const inter = Inter({
  subsets: ['latin'],
  display: 'swap', // 字体交换策略
  variable: '--font-inter',
});

// 中文字体
const notoSansSC = Noto_Sans_SC({
  subsets: ['latin'],
  weight: ['400', '500', '700'],
  display: 'swap',
  variable: '--font-noto-sans-sc',
});

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-CN" className={`${inter.variable} ${notoSansSC.variable}`}>
      <body>{children}</body>
    </html>
  );
}

// CSS中使用
/* globals.css */
body {
  font-family: var(--font-noto-sans-sc), var(--font-inter), sans-serif;
}
```

### 3. 代码分割和懒加载

```typescript
// 动态导入组件
import { lazy, Suspense } from 'react';
import LoadingSpinner from '@/components/ui/LoadingSpinner';

// 懒加载重型组件
const ProductReviews = lazy(() => import('@/components/business/ProductReviews'));
const ProductRecommendations = lazy(() => import('@/components/business/ProductRecommendations'));

export default function ProductDetail({ product }: { product: Product }) {
  return (
    <div>
      <ProductInfo product={product} />

      {/* 懒加载评论组件 */}
      <Suspense fallback={<LoadingSpinner />}>
        <ProductReviews productId={product.id} />
      </Suspense>

      {/* 懒加载推荐组件 */}
      <Suspense fallback={<LoadingSpinner />}>
        <ProductRecommendations categoryId={product.category_id} />
      </Suspense>
    </div>
  );
}

// 路由级别的代码分割
// Next.js自动为每个页面创建单独的bundle
// app/products/page.tsx -> products.js
// app/cart/page.tsx -> cart.js
```

### 4. 缓存策略

```typescript
// 数据缓存配置
export default async function ProductsPage() {
  // 强缓存 - 适用于静态数据
  const categories = await fetch('/api/categories', {
    cache: 'force-cache',
  });

  // 无缓存 - 适用于实时数据
  const products = await fetch('/api/products', {
    cache: 'no-store',
  });

  // 时间缓存 - 适用于半静态数据
  const recommendations = await fetch('/api/recommendations', {
    next: { revalidate: 3600 }, // 1小时后重新验证
  });

  return <ProductGrid products={products} />;
}

// 客户端缓存 - React Query
import { useQuery } from '@tanstack/react-query';

export default function ProductList() {
  const { data: products, isLoading } = useQuery({
    queryKey: ['products'],
    queryFn: fetchProducts,
    staleTime: 5 * 60 * 1000, // 5分钟内不重新获取
    cacheTime: 10 * 60 * 1000, // 10分钟缓存时间
  });

  return isLoading ? <Loading /> : <ProductGrid products={products} />;
}
```

### 5. Bundle分析和优化

```typescript
// next.config.ts - Bundle分析
const nextConfig = {
  // 启用Bundle分析
  webpack: (config, { isServer }) => {
    if (!isServer) {
      // 客户端Bundle分析
      config.resolve.fallback = {
        ...config.resolve.fallback,
        fs: false,
      };
    }

    return config;
  },

  // 实验性功能
  experimental: {
    optimizeCss: true, // CSS优化
    optimizePackageImports: ['antd', 'lodash'], // 包导入优化
  },
};

// 使用Bundle Analyzer
npm install --save-dev @next/bundle-analyzer

// package.json
{
  "scripts": {
    "analyze": "ANALYZE=true next build"
  }
}
```

---

## 🆚 Next.js vs 传统React SPA对比

### 开发体验对比

| 特性 | 传统React SPA | Next.js |
|------|---------------|---------|
| **项目初始化** | 复杂配置 | 零配置启动 |
| **路由配置** | 手动配置React Router | 文件系统自动路由 |
| **代码分割** | 手动配置 | 自动代码分割 |
| **SEO优化** | 需要额外工具 | 内置SSR/SSG |
| **性能优化** | 手动配置 | 内置优化 |
| **部署** | 需要配置服务器 | 一键部署 |

### 性能对比

```typescript
// 传统SPA - 客户端渲染
const TraditionalApp = () => {
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // 客户端获取数据
    fetchProducts().then(data => {
      setProducts(data);
      setLoading(false);
    });
  }, []);

  // 首屏白屏，SEO不友好
  if (loading) return <div>Loading...</div>;

  return <ProductList products={products} />;
};

// Next.js - 服务端渲染
export default async function NextJSApp() {
  // 服务端获取数据
  const products = await fetchProducts();

  // 直接返回渲染好的HTML，SEO友好
  return <ProductList products={products} />;
}
```

### SEO对比

```html
<!-- 传统SPA的HTML -->
<!DOCTYPE html>
<html>
<head>
  <title>React App</title>
</head>
<body>
  <div id="root"></div>
  <script src="/static/js/bundle.js"></script>
</body>
</html>

<!-- Next.js的HTML -->
<!DOCTYPE html>
<html>
<head>
  <title>iPhone 15 Pro - Mall商城</title>
  <meta name="description" content="最新iPhone 15 Pro，A17 Pro芯片，钛金属设计">
  <meta property="og:title" content="iPhone 15 Pro">
  <meta property="og:description" content="最新iPhone 15 Pro，A17 Pro芯片">
  <meta property="og:image" content="/iphone-15-pro.jpg">
</head>
<body>
  <div id="__next">
    <h1>iPhone 15 Pro</h1>
    <p>最新iPhone 15 Pro，A17 Pro芯片，钛金属设计</p>
    <!-- 完整的HTML内容 -->
  </div>
  <script src="/_next/static/chunks/pages/products/[id].js"></script>
</body>
</html>
```

### 部署对比

```bash
# 传统SPA部署
npm run build
# 需要配置Nginx/Apache
# 需要处理路由回退
# 需要配置HTTPS
# 需要配置缓存策略

# Next.js部署
npm run build
npx next start
# 或者一键部署到Vercel
vercel --prod
```

---

## 🎯 面试常考知识点

### 1. Next.js核心概念

**Q: 什么是Next.js？它解决了什么问题？**

**A: Next.js核心价值：**

1. **React的生产级框架** - 提供开箱即用的最佳实践
2. **多种渲染模式** - SSR、SSG、ISR、CSR灵活选择
3. **零配置开发** - 内置TypeScript、ESLint、代码分割等
4. **性能优化** - 自动图片优化、字体优化、Bundle优化
5. **全栈能力** - API Routes支持后端开发

**解决的核心问题**：
- SEO困难 → SSR/SSG解决
- 首屏性能差 → 服务端渲染解决
- 配置复杂 → 零配置解决
- 性能优化难 → 内置优化解决

### 2. 渲染模式选择

**Q: 如何选择SSR、SSG、ISR、CSR？**

**A: 选择策略：**

```typescript
// SSR - 实时数据，SEO要求高
export default async function ProductsPage() {
  const products = await fetchProducts(); // 每次请求都获取最新数据
  return <ProductList products={products} />;
}

// SSG - 静态内容，性能要求高
export async function generateStaticParams() {
  const posts = await fetchPosts();
  return posts.map(post => ({ slug: post.slug }));
}

// ISR - 半静态内容，平衡性能和实时性
export const revalidate = 3600; // 1小时后重新生成

// CSR - 用户相关数据，交互频繁
'use client';
export default function Dashboard() {
  const { data } = useQuery(['user'], fetchUserData);
  return <UserProfile data={data} />;
}
```

### 3. App Router vs Pages Router

**Q: App Router相比Pages Router有什么优势？**

**A: App Router优势：**

1. **更简洁的API** - 直接在组件中使用async/await
2. **原生嵌套布局** - 支持复杂的布局结构
3. **更好的TypeScript支持** - 类型推断更准确
4. **流式渲染** - 支持Suspense和流式传输
5. **更灵活的数据获取** - 组件级别的数据获取

```typescript
// Pages Router
export const getServerSideProps = async () => {
  const data = await fetchData();
  return { props: { data } };
};

// App Router
export default async function Page() {
  const data = await fetchData(); // 更直观
  return <Component data={data} />;
}
```

### 4. 性能优化策略

**Q: Next.js有哪些性能优化策略？**

**A: 性能优化全方位：**

1. **自动代码分割** - 每个页面独立bundle
2. **图片优化** - WebP/AVIF格式，响应式加载
3. **字体优化** - 自动字体优化和预加载
4. **预取策略** - Link组件自动预取
5. **缓存策略** - 多层缓存机制
6. **Bundle优化** - Tree shaking和压缩

### 5. 数据获取最佳实践

**Q: Next.js中如何优化数据获取？**

**A: 数据获取优化：**

```typescript
// 1. 并行数据获取
export default async function Page() {
  const [products, categories] = await Promise.all([
    fetchProducts(),
    fetchCategories(),
  ]);

  return <ProductPage products={products} categories={categories} />;
}

// 2. 流式渲染
export default function Page() {
  return (
    <div>
      <Suspense fallback={<ProductsSkeleton />}>
        <Products />
      </Suspense>
      <Suspense fallback={<CategoriesSkeleton />}>
        <Categories />
      </Suspense>
    </div>
  );
}

// 3. 缓存策略
const products = await fetch('/api/products', {
  next: { revalidate: 3600 }, // ISR
});
```

---

## 🏋️ 实战练习

### 练习1: 构建商品搜索页面

**题目**: 使用Next.js App Router构建一个商品搜索页面，支持SSR和动态路由

**要求**:
1. 支持搜索关键词、分类筛选、价格排序
2. URL参数同步，支持分享和书签
3. SEO友好的元数据生成
4. 加载状态和错误处理

**解决方案**:

```typescript
// app/search/page.tsx
interface SearchParams {
  q?: string;
  category?: string;
  sort?: string;
  page?: string;
}

export async function generateMetadata({
  searchParams,
}: {
  searchParams: SearchParams;
}): Promise<Metadata> {
  const query = searchParams.q || '';
  const category = searchParams.category || '';

  return {
    title: query
      ? `搜索"${query}"的结果 - Mall商城`
      : '商品搜索 - Mall商城',
    description: `在Mall商城搜索${query ? `"${query}"` : '商品'}${category ? `，分类：${category}` : ''}`,
  };
}

export default async function SearchPage({
  searchParams,
}: {
  searchParams: SearchParams;
}) {
  const query = searchParams.q || '';
  const category = searchParams.category || '';
  const sort = searchParams.sort || 'relevance';
  const page = parseInt(searchParams.page || '1');

  // 并行获取数据
  const [searchResults, categories] = await Promise.all([
    searchProducts({
      query,
      category,
      sort,
      page,
      pageSize: 20,
    }),
    fetchCategories(),
  ]);

  return (
    <div className="search-page">
      <SearchHeader query={query} resultCount={searchResults.total} />

      <div className="search-content">
        <aside className="search-filters">
          <CategoryFilter
            categories={categories}
            selected={category}
          />
          <PriceFilter />
          <BrandFilter />
        </aside>

        <main className="search-results">
          <SearchToolbar
            sort={sort}
            total={searchResults.total}
          />

          {searchResults.products.length > 0 ? (
            <>
              <ProductGrid products={searchResults.products} />
              <SearchPagination
                current={page}
                total={searchResults.total}
                pageSize={20}
              />
            </>
          ) : (
            <EmptySearchResults query={query} />
          )}
        </main>
      </div>
    </div>
  );
}

// 搜索工具栏组件
'use client';

import { useRouter, useSearchParams } from 'next/navigation';

interface SearchToolbarProps {
  sort: string;
  total: number;
}

export default function SearchToolbar({ sort, total }: SearchToolbarProps) {
  const router = useRouter();
  const searchParams = useSearchParams();

  const handleSortChange = (newSort: string) => {
    const params = new URLSearchParams(searchParams);
    params.set('sort', newSort);
    params.delete('page'); // 重置页码

    router.push(`/search?${params.toString()}`);
  };

  return (
    <div className="search-toolbar">
      <div className="result-count">
        找到 {total.toLocaleString()} 个商品
      </div>

      <div className="sort-options">
        <Select value={sort} onChange={handleSortChange}>
          <Option value="relevance">相关度</Option>
          <Option value="price_asc">价格从低到高</Option>
          <Option value="price_desc">价格从高到低</Option>
          <Option value="sales">销量</Option>
          <Option value="rating">评分</Option>
          <Option value="newest">最新</Option>
        </Select>
      </div>
    </div>
  );
}

// 数据获取函数
async function searchProducts(params: {
  query: string;
  category: string;
  sort: string;
  page: number;
  pageSize: number;
}) {
  const searchParams = new URLSearchParams({
    q: params.query,
    page: params.page.toString(),
    page_size: params.pageSize.toString(),
    sort: params.sort,
    ...(params.category && { category: params.category }),
  });

  const response = await fetch(
    `${process.env.API_BASE_URL}/api/search?${searchParams}`,
    {
      cache: 'no-store', // 搜索结果需要实时性
    }
  );

  if (!response.ok) {
    throw new Error('Search failed');
  }

  return response.json();
}
```

### 练习2: 实现购物车页面

**题目**: 构建一个功能完整的购物车页面，支持客户端交互和服务端渲染

**要求**:
1. 支持商品数量修改、删除
2. 实时计算总价
3. 优化用户体验（乐观更新）
4. 错误处理和重试机制

**解决方案**:

```typescript
// app/cart/page.tsx
'use client';

import { useEffect, useState } from 'react';
import { useAppDispatch, useAppSelector } from '@/store';
import {
  fetchCartAsync,
  updateCartItemAsync,
  removeCartItemAsync,
  selectCart
} from '@/store/slices/cartSlice';

export default function CartPage() {
  const dispatch = useAppDispatch();
  const { items, total, loading, error } = useAppSelector(selectCart);
  const [selectedItems, setSelectedItems] = useState<number[]>([]);

  useEffect(() => {
    dispatch(fetchCartAsync());
  }, [dispatch]);

  const handleQuantityChange = async (itemId: number, quantity: number) => {
    try {
      await dispatch(updateCartItemAsync({ itemId, quantity })).unwrap();
    } catch (error) {
      message.error('更新失败，请重试');
    }
  };

  const handleRemoveItem = async (itemId: number) => {
    try {
      await dispatch(removeCartItemAsync(itemId)).unwrap();
      message.success('商品已移除');
    } catch (error) {
      message.error('移除失败，请重试');
    }
  };

  const handleSelectItem = (itemId: number, selected: boolean) => {
    setSelectedItems(prev =>
      selected
        ? [...prev, itemId]
        : prev.filter(id => id !== itemId)
    );
  };

  const selectedTotal = items
    .filter(item => selectedItems.includes(item.id))
    .reduce((sum, item) => sum + item.price * item.quantity, 0);

  if (loading) return <CartSkeleton />;
  if (error) return <ErrorMessage error={error} />;

  return (
    <div className="cart-page">
      <div className="cart-header">
        <h1>购物车 ({items.length})</h1>
      </div>

      {items.length > 0 ? (
        <div className="cart-content">
          <div className="cart-items">
            {items.map(item => (
              <CartItem
                key={item.id}
                item={item}
                selected={selectedItems.includes(item.id)}
                onQuantityChange={(quantity) => handleQuantityChange(item.id, quantity)}
                onRemove={() => handleRemoveItem(item.id)}
                onSelect={(selected) => handleSelectItem(item.id, selected)}
              />
            ))}
          </div>

          <div className="cart-summary">
            <div className="summary-content">
              <div className="total-price">
                <span>已选商品 ({selectedItems.length}) 件</span>
                <span className="price">¥{selectedTotal.toFixed(2)}</span>
              </div>
              <Button
                type="primary"
                size="large"
                disabled={selectedItems.length === 0}
                onClick={() => router.push('/checkout')}
              >
                去结算
              </Button>
            </div>
          </div>
        </div>
      ) : (
        <EmptyCart />
      )}
    </div>
  );
}

// 购物车商品组件
interface CartItemProps {
  item: CartItem;
  selected: boolean;
  onQuantityChange: (quantity: number) => void;
  onRemove: () => void;
  onSelect: (selected: boolean) => void;
}

function CartItem({
  item,
  selected,
  onQuantityChange,
  onRemove,
  onSelect
}: CartItemProps) {
  const [quantity, setQuantity] = useState(item.quantity);
  const [updating, setUpdating] = useState(false);

  // 防抖更新数量
  useEffect(() => {
    if (quantity !== item.quantity) {
      const timer = setTimeout(async () => {
        setUpdating(true);
        try {
          await onQuantityChange(quantity);
        } finally {
          setUpdating(false);
        }
      }, 500);

      return () => clearTimeout(timer);
    }
  }, [quantity, item.quantity, onQuantityChange]);

  return (
    <div className={`cart-item ${selected ? 'selected' : ''}`}>
      <Checkbox
        checked={selected}
        onChange={(e) => onSelect(e.target.checked)}
      />

      <div className="item-image">
        <Image
          src={item.product.image}
          alt={item.product.name}
          width={80}
          height={80}
        />
      </div>

      <div className="item-info">
        <h3>{item.product.name}</h3>
        <p>{item.product.description}</p>
        <div className="item-specs">
          {item.specs?.map(spec => (
            <Tag key={spec.name}>{spec.name}: {spec.value}</Tag>
          ))}
        </div>
      </div>

      <div className="item-price">
        ¥{item.price.toFixed(2)}
      </div>

      <div className="item-quantity">
        <InputNumber
          min={1}
          max={item.product.stock}
          value={quantity}
          onChange={(value) => setQuantity(value || 1)}
          loading={updating}
        />
      </div>

      <div className="item-total">
        ¥{(item.price * quantity).toFixed(2)}
      </div>

      <div className="item-actions">
        <Button
          type="text"
          danger
          onClick={onRemove}
          icon={<DeleteOutlined />}
        >
          删除
        </Button>
      </div>
    </div>
  );
}
```

这个练习展示了：

1. **客户端状态管理** - Redux Toolkit的使用
2. **乐观更新** - 立即更新UI，后台同步数据
3. **防抖优化** - 避免频繁的API调用
4. **错误处理** - 完善的错误提示和重试机制
5. **用户体验** - 加载状态、骨架屏、空状态处理

---

## 📚 本章总结

通过本章学习，我们深入掌握了Next.js框架的核心概念和实战应用：

### 🎯 核心收获

1. **Next.js优势** 🚀
   - 理解了Next.js相比传统React SPA的优势
   - 掌握了多种渲染模式的选择和应用
   - 学会了零配置开发的最佳实践

2. **App Router应用** 🛣️
   - 掌握了现代化的文件系统路由
   - 学会了嵌套布局和动态路由设计
   - 理解了与Pages Router的区别和迁移策略

3. **数据获取策略** 📊
   - 深入理解SSR、SSG、ISR、CSR的适用场景
   - 掌握了服务端和客户端数据获取的最佳实践
   - 学会了缓存策略和性能优化

4. **全栈开发** 🔌
   - 掌握了API Routes的设计和实现
   - 学会了中间件的应用场景
   - 理解了前后端一体化开发的优势

5. **性能优化** ⚡
   - 掌握了图片、字体、代码分割等优化技术
   - 学会了缓存策略和Bundle优化
   - 理解了现代Web性能优化的最佳实践

### 🚀 技术进阶

- **下一步学习**: 状态管理与数据流设计
- **实践建议**: 在项目中应用学到的Next.js最佳实践
- **深入方向**: 微前端架构和服务端组件

### 💡 最佳实践

1. **渲染策略**: 根据数据特性选择合适的渲染模式
2. **性能优先**: 充分利用Next.js的内置优化特性
3. **SEO友好**: 合理使用元数据和结构化数据
4. **用户体验**: 注重加载状态、错误处理和交互反馈

Next.js为我们提供了构建现代Web应用的完整解决方案，让开发者能够专注于业务逻辑而不是基础设施！ 🎉

---

*下一章我们将学习《状态管理与数据流设计》，探索复杂应用的状态管理最佳实践！* 🚀
```