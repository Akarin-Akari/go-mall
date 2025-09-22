# 第4章：数据获取与缓存策略优化 ⚡

> _"高效的数据获取和智能的缓存策略，是构建高性能Web应用的核心！"_ 🚀

## 📚 本章导览

在现代Web应用中，数据获取和缓存策略直接影响用户体验和应用性能。Next.js提供了多种数据获取方式和缓存机制，结合React Query、SWR等状态管理库，我们可以构建出响应迅速、用户体验优秀的应用。本章将深入探讨各种数据获取模式、缓存策略，以及在Mall-Frontend项目中的最佳实践。

### 🎯 学习目标

通过本章学习，你将掌握：

- **数据获取模式** - 理解SSR、SSG、CSR、ISR的数据获取策略
- **React Query深度应用** - 掌握服务端状态管理的最佳实践
- **缓存策略设计** - 学会多层缓存架构和缓存失效策略
- **性能优化技巧** - 掌握数据预加载、懒加载、虚拟滚动等技术
- **错误处理机制** - 实现健壮的错误处理和重试策略
- **实时数据同步** - 学会WebSocket、SSE等实时通信技术
- **框架对比分析** - 对比不同框架的数据获取方案
- **企业级实践** - 大型项目中的数据架构设计

### 🛠️ 技术栈概览

```typescript
{
  "dataFetching": {
    "server": ["fetch", "axios", "ky"],
    "client": ["React Query", "SWR", "Apollo Client"],
    "realtime": ["WebSocket", "SSE", "Socket.io"]
  },
  "caching": {
    "browser": ["HTTP Cache", "Service Worker", "IndexedDB"],
    "server": ["Redis", "Memcached", "CDN"],
    "application": ["React Query Cache", "SWR Cache"]
  },
  "optimization": {
    "techniques": ["Prefetching", "Lazy Loading", "Virtual Scrolling"],
    "patterns": ["Stale-While-Revalidate", "Cache-First", "Network-First"]
  },
  "monitoring": ["Performance API", "Core Web Vitals", "Custom Metrics"]
}
```

### 📖 本章目录

- [数据获取模式深度解析](#数据获取模式深度解析)
- [React Query状态管理](#react-query状态管理)
- [多层缓存架构设计](#多层缓存架构设计)
- [性能优化策略](#性能优化策略)
- [错误处理与重试机制](#错误处理与重试机制)
- [实时数据同步](#实时数据同步)
- [框架数据获取对比](#框架数据获取对比)
- [企业级数据架构](#企业级数据架构)
- [Mall-Frontend实战案例](#mall-frontend实战案例)
- [面试常考知识点](#面试常考知识点)
- [实战练习](#实战练习)

---

## 🔄 数据获取模式深度解析

### Next.js数据获取策略

Next.js提供了多种数据获取方式，每种都有其特定的使用场景：

```typescript
// 1. 服务端组件数据获取 (Server Components)
// app/products/page.tsx
import { getProducts, getCategories } from '@/lib/api';

export default async function ProductsPage() {
  // 服务端并行数据获取
  const [products, categories] = await Promise.all([
    getProducts(),
    getCategories()
  ]);

  return (
    <div>
      <CategoryFilter categories={categories} />
      <ProductGrid products={products} />
    </div>
  );
}

// 2. 客户端数据获取 (Client Components)
'use client';

import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';

function ProductList() {
  const { data: products, isLoading, error } = useQuery({
    queryKey: ['products'],
    queryFn: () => fetch('/api/products').then(res => res.json()),
    staleTime: 5 * 60 * 1000, // 5分钟内数据被认为是新鲜的
    cacheTime: 10 * 60 * 1000, // 10分钟后从缓存中移除
  });

  if (isLoading) return <ProductSkeleton />;
  if (error) return <ErrorMessage error={error} />;

  return <ProductGrid products={products} />;
}

// 3. 混合数据获取策略
// app/products/[id]/page.tsx
import { Suspense } from 'react';
import { getProduct } from '@/lib/api';
import { ProductReviews } from '@/components/ProductReviews';

interface ProductPageProps {
  params: { id: string };
}

// 服务端获取核心数据
export default async function ProductPage({ params }: ProductPageProps) {
  const product = await getProduct(params.id);

  return (
    <div>
      {/* 服务端渲染的产品信息 */}
      <ProductDetails product={product} />

      {/* 客户端懒加载的评论 */}
      <Suspense fallback={<ReviewsSkeleton />}>
        <ProductReviews productId={params.id} />
      </Suspense>
    </div>
  );
}

// 4. 增量静态再生 (ISR)
// app/blog/[slug]/page.tsx
export const revalidate = 3600; // 1小时重新验证

export async function generateStaticParams() {
  const posts = await getBlogPosts();
  return posts.map(post => ({ slug: post.slug }));
}

export default async function BlogPost({ params }: { params: { slug: string } }) {
  const post = await getBlogPost(params.slug);

  return (
    <article>
      <h1>{post.title}</h1>
      <div dangerouslySetInnerHTML={{ __html: post.content }} />
    </article>
  );
}
```

### 数据获取性能优化

```typescript
// lib/api-client.ts - 优化的API客户端
class ApiClient {
  private baseURL: string;
  private cache = new Map<
    string,
    { data: any; timestamp: number; ttl: number }
  >();

  constructor(baseURL: string) {
    this.baseURL = baseURL;
  }

  // 带缓存的GET请求
  async get<T>(
    endpoint: string,
    options: {
      cache?: boolean;
      ttl?: number;
      revalidate?: boolean;
    } = {}
  ): Promise<T> {
    const { cache = true, ttl = 5 * 60 * 1000, revalidate = false } = options;
    const cacheKey = `${this.baseURL}${endpoint}`;

    // 检查缓存
    if (cache && !revalidate) {
      const cached = this.cache.get(cacheKey);
      if (cached && Date.now() - cached.timestamp < cached.ttl) {
        return cached.data;
      }
    }

    try {
      const response = await fetch(`${this.baseURL}${endpoint}`, {
        next: { revalidate: ttl / 1000 }, // Next.js缓存配置
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();

      // 更新缓存
      if (cache) {
        this.cache.set(cacheKey, {
          data,
          timestamp: Date.now(),
          ttl,
        });
      }

      return data;
    } catch (error) {
      // 如果网络请求失败，尝试返回过期的缓存数据
      const cached = this.cache.get(cacheKey);
      if (cached) {
        console.warn('Network request failed, returning stale data');
        return cached.data;
      }
      throw error;
    }
  }

  // 批量请求优化
  async batchGet<T>(endpoints: string[]): Promise<T[]> {
    const requests = endpoints.map(endpoint => this.get<T>(endpoint));
    return Promise.all(requests);
  }

  // 预加载数据
  async prefetch(endpoint: string): Promise<void> {
    try {
      await this.get(endpoint, { cache: true });
    } catch (error) {
      console.warn('Prefetch failed:', error);
    }
  }

  // 清除缓存
  clearCache(pattern?: string): void {
    if (pattern) {
      for (const key of this.cache.keys()) {
        if (key.includes(pattern)) {
          this.cache.delete(key);
        }
      }
    } else {
      this.cache.clear();
    }
  }
}

export const apiClient = new ApiClient(
  process.env.NEXT_PUBLIC_API_BASE_URL || ''
);
```

### 数据获取Hook封装

```typescript
// hooks/useApi.ts - 通用数据获取Hook
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api-client';

interface UseApiOptions<T> {
  enabled?: boolean;
  staleTime?: number;
  cacheTime?: number;
  retry?: number;
  onSuccess?: (data: T) => void;
  onError?: (error: Error) => void;
}

export function useApi<T>(endpoint: string, options: UseApiOptions<T> = {}) {
  const {
    enabled = true,
    staleTime = 5 * 60 * 1000,
    cacheTime = 10 * 60 * 1000,
    retry = 3,
    onSuccess,
    onError,
  } = options;

  return useQuery({
    queryKey: [endpoint],
    queryFn: () => apiClient.get<T>(endpoint),
    enabled,
    staleTime,
    cacheTime,
    retry,
    onSuccess,
    onError,
  });
}

// 分页数据获取Hook
export function usePaginatedApi<T>(
  endpoint: string,
  page: number = 1,
  limit: number = 10
) {
  const queryKey = [endpoint, 'paginated', page, limit];

  return useQuery({
    queryKey,
    queryFn: () =>
      apiClient.get<{
        data: T[];
        meta: {
          page: number;
          limit: number;
          total: number;
          totalPages: number;
        };
      }>(`${endpoint}?page=${page}&limit=${limit}`),
    keepPreviousData: true, // 保持上一页数据，避免闪烁
    staleTime: 2 * 60 * 1000, // 2分钟
  });
}

// 无限滚动数据获取Hook
export function useInfiniteApi<T>(endpoint: string, limit: number = 10) {
  return useInfiniteQuery({
    queryKey: [endpoint, 'infinite'],
    queryFn: ({ pageParam = 1 }) =>
      apiClient.get<{
        data: T[];
        meta: { page: number; hasMore: boolean };
      }>(`${endpoint}?page=${pageParam}&limit=${limit}`),
    getNextPageParam: lastPage =>
      lastPage.meta.hasMore ? lastPage.meta.page + 1 : undefined,
    staleTime: 5 * 60 * 1000,
  });
}

// 数据变更Hook
export function useApiMutation<TData, TVariables>(
  endpoint: string,
  method: 'POST' | 'PUT' | 'DELETE' = 'POST'
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (variables: TVariables) => {
      switch (method) {
        case 'POST':
          return fetch(endpoint, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(variables),
          }).then(res => res.json());
        case 'PUT':
          return fetch(endpoint, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(variables),
          }).then(res => res.json());
        case 'DELETE':
          return fetch(endpoint, { method: 'DELETE' }).then(res => res.json());
        default:
          throw new Error(`Unsupported method: ${method}`);
      }
    },
    onSuccess: () => {
      // 使相关查询失效，触发重新获取
      queryClient.invalidateQueries({ queryKey: [endpoint] });
    },
  });
}
```

### 实际应用示例

```typescript
// components/ProductList.tsx - 商品列表组件
'use client';

import { useState } from 'react';
import { useApi, usePaginatedApi } from '@/hooks/useApi';
import { Product, Category } from '@/types';

interface ProductListProps {
  initialProducts?: Product[];
  initialCategories?: Category[];
}

export function ProductList({ initialProducts, initialCategories }: ProductListProps) {
  const [selectedCategory, setSelectedCategory] = useState<number | null>(null);
  const [currentPage, setCurrentPage] = useState(1);

  // 获取分类数据
  const { data: categories } = useApi<Category[]>('/api/categories', {
    // 如果有初始数据，禁用自动获取
    enabled: !initialCategories,
    staleTime: 30 * 60 * 1000, // 分类数据30分钟内有效
  });

  // 获取分页商品数据
  const { data: productsResponse, isLoading, error } = usePaginatedApi<Product>(
    selectedCategory
      ? `/api/products?category=${selectedCategory}`
      : '/api/products',
    currentPage,
    12
  );

  const displayCategories = categories || initialCategories || [];
  const products = productsResponse?.data || initialProducts || [];
  const pagination = productsResponse?.meta;

  if (error) {
    return <ErrorMessage error={error} />;
  }

  return (
    <div className="product-list">
      {/* 分类筛选 */}
      <div className="category-filter">
        <button
          onClick={() => setSelectedCategory(null)}
          className={selectedCategory === null ? 'active' : ''}
        >
          全部
        </button>
        {displayCategories.map(category => (
          <button
            key={category.id}
            onClick={() => setSelectedCategory(category.id)}
            className={selectedCategory === category.id ? 'active' : ''}
          >
            {category.name}
          </button>
        ))}
      </div>

      {/* 商品网格 */}
      {isLoading ? (
        <ProductGridSkeleton />
      ) : (
        <div className="product-grid">
          {products.map(product => (
            <ProductCard key={product.id} product={product} />
          ))}
        </div>
      )}

      {/* 分页控件 */}
      {pagination && (
        <Pagination
          current={pagination.page}
          total={pagination.totalPages}
          onChange={setCurrentPage}
        />
      )}
    </div>
  );
}
```

---

## 🔄 React Query状态管理

### React Query配置和优化

```typescript
// lib/react-query.ts - React Query配置
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';

// 创建QueryClient实例
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // 数据被认为是新鲜的时间（5分钟）
      staleTime: 5 * 60 * 1000,
      // 数据在缓存中保留的时间（10分钟）
      cacheTime: 10 * 60 * 1000,
      // 重试次数
      retry: (failureCount, error: any) => {
        // 4xx错误不重试
        if (error?.status >= 400 && error?.status < 500) {
          return false;
        }
        // 最多重试3次
        return failureCount < 3;
      },
      // 重试延迟（指数退避）
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      // 窗口重新获得焦点时重新获取数据
      refetchOnWindowFocus: false,
      // 网络重连时重新获取数据
      refetchOnReconnect: true,
    },
    mutations: {
      // 变更重试次数
      retry: 1,
      // 变更重试延迟
      retryDelay: 1000,
    },
  },
});

// React Query Provider组件
export function ReactQueryProvider({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
      {process.env.NODE_ENV === 'development' && (
        <ReactQueryDevtools initialIsOpen={false} />
      )}
    </QueryClientProvider>
  );
}
```

### 高级查询模式

```typescript
// hooks/useProducts.ts - 商品相关查询Hook
import {
  useQuery,
  useInfiniteQuery,
  useMutation,
  useQueryClient,
} from '@tanstack/react-query';
import { Product, ProductFilters } from '@/types';
import { apiClient } from '@/lib/api-client';

// 商品查询键工厂
export const productKeys = {
  all: ['products'] as const,
  lists: () => [...productKeys.all, 'list'] as const,
  list: (filters: ProductFilters) => [...productKeys.lists(), filters] as const,
  details: () => [...productKeys.all, 'detail'] as const,
  detail: (id: number) => [...productKeys.details(), id] as const,
  search: (query: string) => [...productKeys.all, 'search', query] as const,
};

// 商品列表查询
export function useProducts(filters: ProductFilters = {}) {
  return useQuery({
    queryKey: productKeys.list(filters),
    queryFn: () =>
      apiClient.get<{
        data: Product[];
        meta: PaginationMeta;
      }>('/api/products', { params: filters }),
    staleTime: 2 * 60 * 1000, // 商品列表2分钟内有效
    select: data => ({
      products: data.data,
      pagination: data.meta,
    }),
  });
}

// 商品详情查询
export function useProduct(id: number) {
  return useQuery({
    queryKey: productKeys.detail(id),
    queryFn: () => apiClient.get<Product>(`/api/products/${id}`),
    staleTime: 10 * 60 * 1000, // 商品详情10分钟内有效
    enabled: !!id, // 只有当id存在时才执行查询
  });
}

// 商品搜索查询（带防抖）
export function useProductSearch(query: string, delay: number = 300) {
  const [debouncedQuery, setDebouncedQuery] = useState(query);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedQuery(query);
    }, delay);

    return () => clearTimeout(timer);
  }, [query, delay]);

  return useQuery({
    queryKey: productKeys.search(debouncedQuery),
    queryFn: () =>
      apiClient.get<Product[]>(`/api/products/search?q=${debouncedQuery}`),
    enabled: debouncedQuery.length >= 2, // 至少2个字符才搜索
    staleTime: 5 * 60 * 1000,
  });
}

// 无限滚动商品查询
export function useInfiniteProducts(filters: ProductFilters = {}) {
  return useInfiniteQuery({
    queryKey: [...productKeys.list(filters), 'infinite'],
    queryFn: ({ pageParam = 1 }) =>
      apiClient.get<{
        data: Product[];
        meta: PaginationMeta & { hasMore: boolean };
      }>('/api/products', {
        params: { ...filters, page: pageParam, limit: 20 },
      }),
    getNextPageParam: lastPage =>
      lastPage.meta.hasMore ? lastPage.meta.page + 1 : undefined,
    staleTime: 5 * 60 * 1000,
    select: data => ({
      pages: data.pages,
      products: data.pages.flatMap(page => page.data),
      hasNextPage: data.pages[data.pages.length - 1]?.meta.hasMore ?? false,
    }),
  });
}

// 商品变更操作
export function useProductMutations() {
  const queryClient = useQueryClient();

  // 创建商品
  const createProduct = useMutation({
    mutationFn: (productData: Omit<Product, 'id'>) =>
      apiClient.post<Product>('/api/products', productData),
    onSuccess: newProduct => {
      // 更新商品列表缓存
      queryClient.setQueryData(productKeys.lists(), (oldData: any) => {
        if (!oldData) return oldData;
        return {
          ...oldData,
          data: [newProduct, ...oldData.data],
        };
      });

      // 使相关查询失效
      queryClient.invalidateQueries({ queryKey: productKeys.lists() });
    },
  });

  // 更新商品
  const updateProduct = useMutation({
    mutationFn: ({ id, ...productData }: Partial<Product> & { id: number }) =>
      apiClient.put<Product>(`/api/products/${id}`, productData),
    onSuccess: updatedProduct => {
      // 更新商品详情缓存
      queryClient.setQueryData(
        productKeys.detail(updatedProduct.id),
        updatedProduct
      );

      // 更新商品列表缓存
      queryClient.setQueryData(productKeys.lists(), (oldData: any) => {
        if (!oldData) return oldData;
        return {
          ...oldData,
          data: oldData.data.map((product: Product) =>
            product.id === updatedProduct.id ? updatedProduct : product
          ),
        };
      });
    },
  });

  return {
    createProduct,
    updateProduct,
  };
}
```

### 乐观更新和错误回滚

```typescript
// hooks/useOptimisticUpdates.ts - 乐观更新Hook
export function useOptimisticProductUpdate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      id,
      updates,
    }: {
      id: number;
      updates: Partial<Product>;
    }) => {
      // 模拟网络延迟
      await new Promise(resolve => setTimeout(resolve, 1000));

      // 模拟可能的错误
      if (Math.random() < 0.3) {
        throw new Error('Update failed');
      }

      return apiClient.put<Product>(`/api/products/${id}`, updates);
    },

    // 乐观更新
    onMutate: async ({ id, updates }) => {
      // 取消相关的查询，避免覆盖乐观更新
      await queryClient.cancelQueries({ queryKey: productKeys.detail(id) });

      // 获取当前数据的快照
      const previousProduct = queryClient.getQueryData(productKeys.detail(id));

      // 乐观更新数据
      queryClient.setQueryData(
        productKeys.detail(id),
        (old: Product | undefined) => {
          if (!old) return old;
          return { ...old, ...updates };
        }
      );

      // 返回上下文对象，包含回滚数据
      return { previousProduct, id };
    },

    // 成功时的处理
    onSuccess: (updatedProduct, { id }) => {
      // 用服务器返回的真实数据替换乐观更新的数据
      queryClient.setQueryData(productKeys.detail(id), updatedProduct);
    },

    // 错误时的回滚
    onError: (error, { id }, context) => {
      // 回滚到之前的数据
      if (context?.previousProduct) {
        queryClient.setQueryData(
          productKeys.detail(id),
          context.previousProduct
        );
      }

      // 显示错误提示
      toast.error('更新失败，请重试');
    },

    // 无论成功还是失败都会执行
    onSettled: (data, error, { id }) => {
      // 重新获取数据以确保一致性
      queryClient.invalidateQueries({ queryKey: productKeys.detail(id) });
    },
  });
}
```

### 数据同步和实时更新

```typescript
// hooks/useRealtimeData.ts - 实时数据同步
import { useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';

export function useRealtimeProductUpdates() {
  const queryClient = useQueryClient();

  useEffect(() => {
    // WebSocket连接
    const ws = new WebSocket(
      process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:3001'
    );

    ws.onmessage = event => {
      const message = JSON.parse(event.data);

      switch (message.type) {
        case 'PRODUCT_UPDATED':
          // 更新商品详情缓存
          queryClient.setQueryData(
            productKeys.detail(message.data.id),
            message.data
          );

          // 使商品列表查询失效
          queryClient.invalidateQueries({ queryKey: productKeys.lists() });
          break;

        case 'PRODUCT_DELETED':
          // 移除商品详情缓存
          queryClient.removeQueries({
            queryKey: productKeys.detail(message.data.id),
          });

          // 使商品列表查询失效
          queryClient.invalidateQueries({ queryKey: productKeys.lists() });
          break;

        case 'PRODUCT_CREATED':
          // 使商品列表查询失效，触发重新获取
          queryClient.invalidateQueries({ queryKey: productKeys.lists() });
          break;
      }
    };

    ws.onerror = error => {
      console.error('WebSocket error:', error);
    };

    return () => {
      ws.close();
    };
  }, [queryClient]);
}

// Server-Sent Events (SSE) 实现
export function useSSEUpdates() {
  const queryClient = useQueryClient();

  useEffect(() => {
    const eventSource = new EventSource('/api/sse/products');

    eventSource.onmessage = event => {
      const data = JSON.parse(event.data);

      // 根据事件类型更新缓存
      switch (data.type) {
        case 'stock_update':
          queryClient.setQueryData(
            productKeys.detail(data.productId),
            (old: Product | undefined) => {
              if (!old) return old;
              return { ...old, stock: data.stock };
            }
          );
          break;

        case 'price_update':
          queryClient.setQueryData(
            productKeys.detail(data.productId),
            (old: Product | undefined) => {
              if (!old) return old;
              return { ...old, price: data.price };
            }
          );
          break;
      }
    };

    eventSource.onerror = error => {
      console.error('SSE error:', error);
    };

    return () => {
      eventSource.close();
    };
  }, [queryClient]);
}
```

---

## 🏗️ 多层缓存架构设计

### 浏览器缓存策略

```typescript
// lib/cache-manager.ts - 缓存管理器
interface CacheConfig {
  name: string;
  version: number;
  stores: {
    memory: boolean;
    localStorage: boolean;
    sessionStorage: boolean;
    indexedDB: boolean;
    serviceWorker: boolean;
  };
}

class CacheManager {
  private config: CacheConfig;
  private memoryCache = new Map<string, { data: any; expires: number }>();

  constructor(config: CacheConfig) {
    this.config = config;
  }

  // 内存缓存
  async setMemory(
    key: string,
    data: any,
    ttl: number = 5 * 60 * 1000
  ): Promise<void> {
    this.memoryCache.set(key, {
      data,
      expires: Date.now() + ttl,
    });
  }

  async getMemory<T>(key: string): Promise<T | null> {
    const cached = this.memoryCache.get(key);
    if (!cached) return null;

    if (Date.now() > cached.expires) {
      this.memoryCache.delete(key);
      return null;
    }

    return cached.data;
  }

  // localStorage缓存
  async setLocalStorage(key: string, data: any, ttl?: number): Promise<void> {
    if (!this.config.stores.localStorage) return;

    const item = {
      data,
      expires: ttl ? Date.now() + ttl : null,
    };

    try {
      localStorage.setItem(key, JSON.stringify(item));
    } catch (error) {
      console.warn('localStorage write failed:', error);
    }
  }

  async getLocalStorage<T>(key: string): Promise<T | null> {
    if (!this.config.stores.localStorage) return null;

    try {
      const item = localStorage.getItem(key);
      if (!item) return null;

      const parsed = JSON.parse(item);

      if (parsed.expires && Date.now() > parsed.expires) {
        localStorage.removeItem(key);
        return null;
      }

      return parsed.data;
    } catch (error) {
      console.warn('localStorage read failed:', error);
      return null;
    }
  }
}

// 创建缓存管理器实例
export const cacheManager = new CacheManager({
  name: 'mall-frontend-cache',
  version: 1,
  stores: {
    memory: true,
    localStorage: true,
    sessionStorage: false,
    indexedDB: true,
    serviceWorker: false,
  },
});
```

---

## 🎯 面试常考知识点

### 1. 数据获取策略对比

**Q: Next.js中SSR、SSG、CSR、ISR的区别和适用场景？**

**A: 四种渲染策略详细对比：**

```typescript
// 渲染策略对比表
interface RenderingStrategy {
  name: string;
  renderTime: 'build' | 'request' | 'client';
  dataFreshness: 'static' | 'dynamic' | 'stale-while-revalidate';
  performance: 'excellent' | 'good' | 'average';
  seo: 'excellent' | 'good' | 'poor';
  serverLoad: 'none' | 'low' | 'medium' | 'high';
  useCase: string[];
}

const renderingStrategies: RenderingStrategy[] = [
  {
    name: 'SSG (Static Site Generation)',
    renderTime: 'build',
    dataFreshness: 'static',
    performance: 'excellent',
    seo: 'excellent',
    serverLoad: 'none',
    useCase: ['博客文章', '产品页面', '营销页面', '文档站点']
  },
  {
    name: 'SSR (Server-Side Rendering)',
    renderTime: 'request',
    dataFreshness: 'dynamic',
    performance: 'good',
    seo: 'excellent',
    serverLoad: 'high',
    useCase: ['个人仪表板', '实时数据', '个性化内容', '用户特定页面']
  },
  {
    name: 'CSR (Client-Side Rendering)',
    renderTime: 'client',
    dataFreshness: 'dynamic',
    performance: 'average',
    seo: 'poor',
    serverLoad: 'low',
    serverLoad: 'low',
    useCase: ['管理后台', '复杂交互', '实时应用', 'SPA应用']
  },
  {
    name: 'ISR (Incremental Static Regeneration)',
    renderTime: 'build',
    dataFreshness: 'stale-while-revalidate',
    performance: 'excellent',
    seo: 'excellent',
    serverLoad: 'low',
    useCase: ['电商产品', '新闻网站', '内容管理', '大型网站']
  }
];

// 实际代码示例
// 1. SSG - 静态生成
export async function generateStaticParams() {
  const products = await getProducts();
  return products.map(product => ({ id: product.id.toString() }));
}

export default async function ProductPage({ params }: { params: { id: string } }) {
  const product = await getProduct(params.id);
  return <ProductDetails product={product} />;
}

// 2. SSR - 服务端渲染
export default async function DashboardPage() {
  const userData = await getUserData(); // 每次请求都执行
  return <Dashboard data={userData} />;
}

// 3. CSR - 客户端渲染
'use client';
export default function AdminPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['admin-data'],
    queryFn: fetchAdminData
  });

  if (isLoading) return <Loading />;
  return <AdminDashboard data={data} />;
}

// 4. ISR - 增量静态再生
export const revalidate = 3600; // 1小时重新验证

export default async function ProductListPage() {
  const products = await getProducts();
  return <ProductList products={products} />;
}
```

### 2. React Query vs SWR vs Apollo Client

**Q: 如何选择合适的数据获取库？**

**A: 三大数据获取库对比：**

```typescript
// 功能对比矩阵
interface DataFetchingLibrary {
  name: string;
  bundleSize: string;
  typescript: 'excellent' | 'good' | 'basic';
  caching: 'advanced' | 'good' | 'basic';
  devtools: boolean;
  offline: boolean;
  optimisticUpdates: boolean;
  infiniteQueries: boolean;
  suspense: boolean;
  ecosystem: 'rich' | 'growing' | 'limited';
}

const libraries: DataFetchingLibrary[] = [
  {
    name: 'React Query',
    bundleSize: '~13kb',
    typescript: 'excellent',
    caching: 'advanced',
    devtools: true,
    offline: true,
    optimisticUpdates: true,
    infiniteQueries: true,
    suspense: true,
    ecosystem: 'rich',
  },
  {
    name: 'SWR',
    bundleSize: '~4kb',
    typescript: 'good',
    caching: 'good',
    devtools: false,
    offline: true,
    optimisticUpdates: true,
    infiniteQueries: true,
    suspense: true,
    ecosystem: 'growing',
  },
  {
    name: 'Apollo Client',
    bundleSize: '~33kb',
    typescript: 'excellent',
    caching: 'advanced',
    devtools: true,
    offline: true,
    optimisticUpdates: true,
    infiniteQueries: false,
    suspense: true,
    ecosystem: 'rich',
  },
];

// 使用场景对比
const useCaseComparison = {
  'React Query': {
    bestFor: 'REST API, 复杂缓存需求, 企业级应用',
    pros: ['强大的缓存机制', '丰富的配置选项', '优秀的开发体验'],
    cons: ['包体积较大', '学习曲线陡峭'],
  },

  SWR: {
    bestFor: '简单应用, 快速原型, 包体积敏感',
    pros: ['轻量级', '简单易用', 'Vercel官方支持'],
    cons: ['功能相对简单', '生态系统较小'],
  },

  'Apollo Client': {
    bestFor: 'GraphQL应用, 复杂状态管理',
    pros: ['GraphQL原生支持', '强大的缓存', '丰富的功能'],
    cons: ['包体积最大', '仅适用于GraphQL'],
  },
};
```

### 3. 缓存策略设计

**Q: 如何设计多层缓存架构？**

**A: 缓存层次和策略：**

```typescript
// 缓存层次架构
const cacheArchitecture = {
  // L1: 内存缓存 (最快)
  memory: {
    ttl: '5分钟',
    size: '50MB',
    hitRate: '90%',
    useCase: '热点数据, 计算结果',
  },

  // L2: 浏览器缓存
  browser: {
    localStorage: {
      ttl: '1天',
      size: '5-10MB',
      useCase: '用户偏好, 配置信息',
    },
    sessionStorage: {
      ttl: '会话期间',
      size: '5-10MB',
      useCase: '临时数据, 表单状态',
    },
    indexedDB: {
      ttl: '1周',
      size: '50MB+',
      useCase: '大量数据, 离线支持',
    },
  },

  // L3: HTTP缓存
  http: {
    browserCache: {
      ttl: '1小时',
      useCase: '静态资源, API响应',
    },
    cdn: {
      ttl: '1天',
      useCase: '图片, CSS, JS文件',
    },
  },

  // L4: 服务端缓存
  server: {
    redis: {
      ttl: '1小时',
      useCase: '数据库查询结果, 会话数据',
    },
    memcached: {
      ttl: '30分钟',
      useCase: '计算密集型结果',
    },
  },
};

// 缓存策略模式
const cacheStrategies = {
  'Cache-First': {
    description: '优先使用缓存，缓存未命中时请求网络',
    useCase: '静态内容, 不经常变化的数据',
    implementation: `
      const data = await cache.get(key) || await fetch(url);
    `,
  },

  'Network-First': {
    description: '优先请求网络，网络失败时使用缓存',
    useCase: '实时数据, 经常变化的内容',
    implementation: `
      try {
        const data = await fetch(url);
        cache.set(key, data);
        return data;
      } catch {
        return cache.get(key);
      }
    `,
  },

  'Stale-While-Revalidate': {
    description: '返回缓存数据，同时在后台更新',
    useCase: '用户体验优先, 可接受短暂过期数据',
    implementation: `
      const cachedData = cache.get(key);
      if (cachedData) {
        fetch(url).then(data => cache.set(key, data)); // 后台更新
        return cachedData;
      }
      return fetch(url);
    `,
  },
};
```

### 4. 性能优化技巧

**Q: 如何优化数据获取性能？**

**A: 性能优化策略：**

```typescript
// 性能优化技巧清单
const performanceOptimizations = {
  // 1. 数据预加载
  prefetching: {
    linkPrefetch: '<link rel="prefetch" href="/api/products">',
    routePrefetch: 'router.prefetch("/products")',
    dataPrefetch: 'queryClient.prefetchQuery(productKeys.list())'
  },

  // 2. 懒加载
  lazyLoading: {
    componentLazy: 'const LazyComponent = lazy(() => import("./Component"))',
    imageLazy: '<img loading="lazy" src="image.jpg">',
    dataLazy: 'enabled: isVisible // 只有可见时才加载'
  },

  // 3. 虚拟滚动
  virtualScrolling: {
    library: 'react-window, react-virtualized',
    useCase: '大量列表数据, 表格',
    benefit: '只渲染可见项目，减少DOM节点'
  },

  // 4. 请求合并
  requestBatching: {
    dataloader: '合并多个数据库查询',
    graphql: '单个请求获取多个资源',
    restBatch: '批量API端点设计'
  },

  // 5. 响应压缩
  compression: {
    gzip: '文本压缩率70-90%',
    brotli: '比gzip更好的压缩率',
    imageOptimization: 'WebP, AVIF格式'
  }
};

// 实际优化示例
// 1. 数据预加载Hook
export function usePrefetchOnHover() {
  const queryClient = useQueryClient();

  return useCallback((productId: number) => {
    queryClient.prefetchQuery({
      queryKey: productKeys.detail(productId),
      queryFn: () => apiClient.get(`/api/products/${productId}`),
      staleTime: 10 * 60 * 1000,
    });
  }, [queryClient]);
}

// 2. 虚拟滚动实现
import { FixedSizeList as List } from 'react-window';

function VirtualizedProductList({ products }: { products: Product[] }) {
  const Row = ({ index, style }: { index: number; style: React.CSSProperties }) => (
    <div style={style}>
      <ProductCard product={products[index]} />
    </div>
  );

  return (
    <List
      height={600}
      itemCount={products.length}
      itemSize={200}
      width="100%"
    >
      {Row}
    </List>
  );
}

// 3. 请求防抖Hook
export function useDebouncedQuery<T>(
  queryKey: any[],
  queryFn: () => Promise<T>,
  searchTerm: string,
  delay: number = 300
) {
  const [debouncedTerm, setDebouncedTerm] = useState(searchTerm);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedTerm(searchTerm), delay);
    return () => clearTimeout(timer);
  }, [searchTerm, delay]);

  return useQuery({
    queryKey: [...queryKey, debouncedTerm],
    queryFn,
    enabled: debouncedTerm.length >= 2,
  });
}
```

---

## 📚 实战练习

### 练习1：实现智能缓存系统

**任务**: 为Mall-Frontend实现一个智能的多层缓存系统，包括内存、localStorage、IndexedDB三层缓存。

**要求**:

- 实现缓存优先级和回退机制
- 支持缓存过期和自动清理
- 提供缓存命中率统计
- 实现缓存预热功能

### 练习2：优化商品列表性能

**任务**: 优化商品列表页面的加载性能，实现虚拟滚动、图片懒加载、数据预加载等功能。

**要求**:

- 使用react-window实现虚拟滚动
- 实现图片懒加载和渐进式加载
- 添加骨架屏和加载状态
- 实现无限滚动分页

### 练习3：构建实时数据同步

**任务**: 实现商品库存的实时同步，当库存发生变化时自动更新所有相关页面。

**要求**:

- 使用WebSocket或SSE实现实时通信
- 实现乐观更新和错误回滚
- 添加离线状态检测和数据同步
- 实现冲突解决机制

---

## 📚 本章总结

通过本章学习，我们全面掌握了数据获取与缓存策略的核心技术：

### 🎯 核心收获

1. **数据获取精通** 🔄
   - 掌握了Next.js四种渲染模式的特点和应用场景
   - 学会了React Query的高级用法和最佳实践
   - 理解了数据获取的性能优化策略

2. **缓存架构设计** 🏗️
   - 掌握了多层缓存架构的设计原理
   - 学会了不同缓存策略的选择和实现
   - 理解了缓存失效和数据一致性问题

3. **性能优化技巧** ⚡
   - 掌握了虚拟滚动、懒加载等优化技术
   - 学会了请求合并、防抖等性能优化方法
   - 理解了Core Web Vitals和性能监控

4. **实时数据同步** 🔄
   - 掌握了WebSocket和SSE的实现方式
   - 学会了乐观更新和错误回滚机制
   - 理解了实时数据的冲突解决策略

5. **企业级实践** 🚀
   - 学会了大型项目的数据架构设计
   - 掌握了缓存监控和性能分析
   - 理解了数据获取的最佳实践模式

### 🚀 技术进阶

- **下一步学习**: 前端架构设计原则
- **实践建议**: 在项目中应用多层缓存架构
- **深入方向**: 微前端数据共享和状态管理

高效的数据获取和智能的缓存策略是构建高性能Web应用的基石！ 🎉

---

_下一章我们将学习《前端架构设计原则》，探索大型项目的架构设计！_ 🚀
