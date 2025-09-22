# 第4章：模块系统与命名空间 📦

> _"良好的模块化设计是大型应用的基石，让代码组织井然有序！"_ 🏗️

## 📚 本章导览

在现代TypeScript开发中，模块系统是组织代码的核心机制。通过本章学习，我们将深入了解TypeScript的模块系统、命名空间，以及如何在Mall-Frontend项目中应用这些概念来构建可维护的大型应用。

### 🎯 学习目标

通过本章学习，你将掌握：

- **ES6模块系统** - 现代JavaScript的模块化标准
- **CommonJS与AMD** - 不同模块系统的对比和应用
- **TypeScript命名空间** - 代码组织的传统方式
- **模块解析策略** - TypeScript如何查找和加载模块
- **声明文件** - 为第三方库提供类型支持
- **模块增强** - 扩展现有模块的功能
- **实战应用** - 在Mall-Frontend中的模块化实践

### 🛠️ 技术栈概览

```typescript
{
  "modules": "ES6 Modules",
  "bundlers": ["Webpack", "Vite", "Rollup"],
  "tools": ["TypeScript", "Node.js", "npm"],
  "patterns": ["Barrel Exports", "Re-exports", "Dynamic Imports"]
}
```

---

## 🌟 ES6模块系统：现代化的代码组织

### 基础导入导出

```typescript
// types/index.ts - 类型定义
export interface User {
  id: number;
  username: string;
  email: string;
  role: 'admin' | 'user';
}

export interface Product {
  id: number;
  name: string;
  price: string;
  category_id: number;
  description: string;
  images: string[];
  stock: number;
}

export interface CartItem {
  id: number;
  product: Product;
  quantity: number;
  selected: boolean;
}

// 默认导出
export default interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}
```

```typescript
// utils/api.ts - API工具函数
import ApiResponse, { User, Product } from '../types';

// 命名导出
export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';

export class ApiClient {
  private baseURL: string;

  constructor(baseURL: string = API_BASE_URL) {
    this.baseURL = baseURL;
  }

  async get<T>(endpoint: string): Promise<ApiResponse<T>> {
    const response = await fetch(`${this.baseURL}${endpoint}`);
    return response.json();
  }

  async post<T>(endpoint: string, data: any): Promise<ApiResponse<T>> {
    const response = await fetch(`${this.baseURL}${endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return response.json();
  }
}
```

### 🔄 语言对比：模块系统实现

```java
// Java - 包系统（Package System）
// com/mall/types/User.java
package com.mall.types;

public class User {
    private int id;
    private String username;
    private String email;
    private String role;

    // 构造函数、getter、setter
    public User(int id, String username, String email, String role) {
        this.id = id;
        this.username = username;
        this.email = email;
        this.role = role;
    }

    // getters and setters...
}

// com/mall/utils/ApiClient.java
package com.mall.utils;

import com.mall.types.User;
import com.mall.types.Product;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;

public class ApiClient {
    private static final String API_BASE_URL = "http://localhost:8080";
    private final HttpClient client;

    public ApiClient() {
        this.client = HttpClient.newHttpClient();
    }

    public <T> ApiResponse<T> get(String endpoint, Class<T> responseType) {
        // HTTP GET 实现
        return null; // 简化示例
    }
}
```

```python
# Python - 模块系统（Module System）
# types/__init__.py
from dataclasses import dataclass
from typing import List, Literal

@dataclass
class User:
    id: int
    username: str
    email: str
    role: Literal['admin', 'user']

@dataclass
class Product:
    id: int
    name: str
    price: str
    category_id: int
    description: str
    images: List[str]
    stock: int

@dataclass
class CartItem:
    id: int
    product: Product
    quantity: int
    selected: bool

# utils/api.py
import os
import requests
from typing import TypeVar, Generic
from types import User, Product

T = TypeVar('T')

class ApiResponse(Generic[T]):
    def __init__(self, code: int, message: str, data: T):
        self.code = code
        self.message = message
        self.data = data

API_BASE_URL = os.getenv('API_BASE_URL', 'http://localhost:8080')

class ApiClient:
    def __init__(self, base_url: str = API_BASE_URL):
        self.base_url = base_url

    def get(self, endpoint: str) -> ApiResponse:
        response = requests.get(f"{self.base_url}{endpoint}")
        return ApiResponse(
            code=response.status_code,
            message="success",
            data=response.json()
        )
```

```go
// Go - 包系统（Package System）
// types/user.go
package types

type User struct {
    ID       int    `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Role     string `json:"role"`
}

type Product struct {
    ID          int      `json:"id"`
    Name        string   `json:"name"`
    Price       string   `json:"price"`
    CategoryID  int      `json:"category_id"`
    Description string   `json:"description"`
    Images      []string `json:"images"`
    Stock       int      `json:"stock"`
}

type CartItem struct {
    ID       int     `json:"id"`
    Product  Product `json:"product"`
    Quantity int     `json:"quantity"`
    Selected bool    `json:"selected"`
}

// utils/api.go
package utils

import (
    "encoding/json"
    "net/http"
    "os"
    "mall/types"
)

type ApiResponse[T any] struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    T      `json:"data"`
}

var APIBaseURL = getEnv("API_BASE_URL", "http://localhost:8080")

type ApiClient struct {
    BaseURL string
    Client  *http.Client
}

func NewApiClient() *ApiClient {
    return &ApiClient{
        BaseURL: APIBaseURL,
        Client:  &http.Client{},
    }
}

func (c *ApiClient) Get(endpoint string) (*ApiResponse[interface{}], error) {
    resp, err := c.Client.Get(c.BaseURL + endpoint)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result ApiResponse[interface{}]
    err = json.NewDecoder(resp.Body).Decode(&result)
    return &result, err
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

```csharp
// C# - 命名空间系统（Namespace System）
// Types/User.cs
namespace Mall.Types
{
    public class User
    {
        public int Id { get; set; }
        public string Username { get; set; }
        public string Email { get; set; }
        public string Role { get; set; }
    }

    public class Product
    {
        public int Id { get; set; }
        public string Name { get; set; }
        public string Price { get; set; }
        public int CategoryId { get; set; }
        public string Description { get; set; }
        public string[] Images { get; set; }
        public int Stock { get; set; }
    }
}

// Utils/ApiClient.cs
using System;
using System.Net.Http;
using System.Threading.Tasks;
using Mall.Types;

namespace Mall.Utils
{
    public class ApiResponse<T>
    {
        public int Code { get; set; }
        public string Message { get; set; }
        public T Data { get; set; }
    }

    public class ApiClient
    {
        private static readonly string ApiBaseUrl =
            Environment.GetEnvironmentVariable("API_BASE_URL") ?? "http://localhost:8080";

        private readonly HttpClient _client;

        public ApiClient()
        {
            _client = new HttpClient();
        }

        public async Task<ApiResponse<T>> GetAsync<T>(string endpoint)
        {
            var response = await _client.GetAsync($"{ApiBaseUrl}{endpoint}");
            var content = await response.Content.ReadAsStringAsync();
            return JsonSerializer.Deserialize<ApiResponse<T>>(content);
        }
    }
}
```

**💡 模块系统对比：**

| 特性         | TypeScript      | Java     | Python        | Go       | C#       |
| ------------ | --------------- | -------- | ------------- | -------- | -------- |
| **模块单位** | 文件            | 类文件   | 文件/目录     | 包目录   | 命名空间 |
| **导入语法** | `import/export` | `import` | `import/from` | `import` | `using`  |
| **默认导出** | 支持            | 不支持   | 支持          | 不支持   | 不支持   |
| **重导出**   | 支持            | 不支持   | 支持          | 不支持   | 支持     |
| **动态导入** | 支持            | 反射     | 支持          | 不支持   | 反射     |
| **循环依赖** | 部分支持        | 编译错误 | 支持          | 编译错误 | 编译错误 |

// 默认导出
export default new ApiClient();

````

### 重新导出和Barrel模式

```typescript
// services/index.ts - Barrel导出
export { default as apiClient, ApiClient } from './api';
export { UserService } from './userService';
export { ProductService } from './productService';
export { CartService } from './cartService';
export { OrderService } from './orderService';

// 类型重新导出
export type { User, Product, CartItem } from '../types';

// 默认导出服务集合
export default {
  user: new UserService(),
  product: new ProductService(),
  cart: new CartService(),
  order: new OrderService(),
};
````

```typescript
// services/userService.ts
import apiClient from './api';
import type { User, ApiResponse } from '../types';

export class UserService {
  async getCurrentUser(): Promise<User | null> {
    try {
      const response = await apiClient.get<User>('/api/user/profile');
      return response.data;
    } catch (error) {
      console.error('获取用户信息失败:', error);
      return null;
    }
  }

  async updateProfile(data: Partial<User>): Promise<boolean> {
    try {
      const response = await apiClient.post<User>('/api/user/profile', data);
      return response.code === 200;
    } catch (error) {
      console.error('更新用户信息失败:', error);
      return false;
    }
  }

  async login(username: string, password: string): Promise<string | null> {
    try {
      const response = await apiClient.post<{ token: string }>(
        '/api/auth/login',
        {
          username,
          password,
        }
      );
      return response.data.token;
    } catch (error) {
      console.error('登录失败:', error);
      return null;
    }
  }
}
```

### 动态导入

```typescript
// utils/dynamicImports.ts - 动态导入工具
export class DynamicLoader {
  private static cache = new Map<string, any>();

  // 动态加载组件
  static async loadComponent(componentName: string) {
    if (this.cache.has(componentName)) {
      return this.cache.get(componentName);
    }

    try {
      let component;

      switch (componentName) {
        case 'ProductCard':
          component = await import('../components/business/ProductCard');
          break;
        case 'CartItem':
          component = await import('../components/business/CartItem');
          break;
        case 'OrderHistory':
          component = await import('../components/business/OrderHistory');
          break;
        default:
          throw new Error(`未知组件: ${componentName}`);
      }

      this.cache.set(componentName, component.default);
      return component.default;
    } catch (error) {
      console.error(`加载组件 ${componentName} 失败:`, error);
      return null;
    }
  }

  // 动态加载工具库
  static async loadUtility(utilityName: string) {
    try {
      switch (utilityName) {
        case 'lodash':
          return await import('lodash');
        case 'dayjs':
          return await import('dayjs');
        case 'crypto':
          return await import('crypto-js');
        default:
          throw new Error(`未知工具库: ${utilityName}`);
      }
    } catch (error) {
      console.error(`加载工具库 ${utilityName} 失败:`, error);
      return null;
    }
  }

  // 条件加载
  static async conditionalLoad(condition: boolean, modulePath: string) {
    if (!condition) return null;

    try {
      const module = await import(modulePath);
      return module.default || module;
    } catch (error) {
      console.error(`条件加载模块失败:`, error);
      return null;
    }
  }
}

// 使用示例
export async function loadProductPage() {
  const [ProductCard, ProductFilter, ProductSort] = await Promise.all([
    DynamicLoader.loadComponent('ProductCard'),
    DynamicLoader.loadComponent('ProductFilter'),
    DynamicLoader.loadComponent('ProductSort'),
  ]);

  return { ProductCard, ProductFilter, ProductSort };
}
```

---

## 🏛️ TypeScript命名空间：传统的代码组织

### 基础命名空间

```typescript
// namespaces/Mall.ts - 商城命名空间
namespace Mall {
  // 嵌套命名空间
  export namespace Types {
    export interface User {
      id: number;
      username: string;
      email: string;
    }

    export interface Product {
      id: number;
      name: string;
      price: number;
    }

    export interface Order {
      id: number;
      userId: number;
      products: Product[];
      total: number;
    }
  }

  export namespace Utils {
    export function formatPrice(price: number): string {
      return `¥${price.toFixed(2)}`;
    }

    export function formatDate(date: Date): string {
      return date.toLocaleDateString('zh-CN');
    }

    export class Validator {
      static isEmail(email: string): boolean {
        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
      }

      static isPhone(phone: string): boolean {
        return /^1[3-9]\d{9}$/.test(phone);
      }
    }
  }

  export namespace Services {
    export class UserService {
      static async getUser(id: number): Promise<Types.User | null> {
        // 实现用户服务逻辑
        return null;
      }
    }

    export class ProductService {
      static async getProducts(): Promise<Types.Product[]> {
        // 实现商品服务逻辑
        return [];
      }
    }
  }

  // 命名空间常量
  export const CONFIG = {
    API_BASE_URL: 'http://localhost:8080',
    VERSION: '1.0.0',
    TIMEOUT: 5000,
  };
}

// 使用命名空间
const user: Mall.Types.User = {
  id: 1,
  username: 'john',
  email: 'john@example.com',
};

const formattedPrice = Mall.Utils.formatPrice(99.99);
const isValidEmail = Mall.Utils.Validator.isEmail(user.email);
```

### 命名空间合并

```typescript
// 第一个文件中的命名空间
namespace Mall.API {
  export interface RequestConfig {
    timeout: number;
    retries: number;
  }

  export class HttpClient {
    constructor(private config: RequestConfig) {}
  }
}

// 第二个文件中扩展同一个命名空间
namespace Mall.API {
  export interface ResponseFormat {
    code: number;
    message: string;
    data: any;
  }

  export class ResponseHandler {
    static parse(response: ResponseFormat) {
      return response.data;
    }
  }
}

// 现在Mall.API命名空间包含了两个文件中的所有内容
const client = new Mall.API.HttpClient({ timeout: 5000, retries: 3 });
const handler = Mall.API.ResponseHandler;
```

---

## 🔍 模块解析策略

### TypeScript模块解析

```typescript
// tsconfig.json - 模块解析配置
{
  "compilerOptions": {
    "moduleResolution": "node",           // Node.js风格的模块解析
    "baseUrl": "./src",                   // 基础路径
    "paths": {                            // 路径映射
      "@/*": ["*"],                       // @/ 映射到 src/
      "@/components/*": ["components/*"], // 组件路径
      "@/utils/*": ["utils/*"],           // 工具路径
      "@/types/*": ["types/*"],           // 类型路径
      "@/services/*": ["services/*"],     // 服务路径
      "@/hooks/*": ["hooks/*"],           // Hooks路径
      "@/store/*": ["store/*"]            // 状态管理路径
    },
    "typeRoots": ["./node_modules/@types", "./src/types"], // 类型根目录
    "types": ["node", "react", "react-dom"]                // 包含的类型包
  }
}
```

```typescript
// 使用路径映射的导入
import { Button } from '@/components/ui/Button';
import { formatPrice } from '@/utils/format';
import { User } from '@/types/user';
import { useAuth } from '@/hooks/useAuth';
import { userSlice } from '@/store/slices/userSlice';

// 相对路径导入
import { ProductCard } from '../components/ProductCard';
import { apiClient } from './api';

// 绝对路径导入
import React from 'react';
import { NextPage } from 'next';
import { Button } from 'antd';
```

### 模块增强

```typescript
// types/global.d.ts - 全局类型增强
declare global {
  interface Window {
    __MALL_CONFIG__: {
      apiBaseUrl: string;
      version: string;
      debug: boolean;
    };
  }

  namespace NodeJS {
    interface ProcessEnv {
      NEXT_PUBLIC_API_BASE_URL: string;
      NEXT_PUBLIC_APP_NAME: string;
      DATABASE_URL: string;
      JWT_SECRET: string;
    }
  }
}

// 扩展第三方库
declare module 'antd' {
  interface ButtonProps {
    loading?: boolean;
    danger?: boolean;
  }
}

// 扩展Next.js类型
declare module 'next' {
  interface NextPageContext {
    user?: User;
  }
}

export {}; // 确保这是一个模块
```

```typescript
// types/modules.d.ts - 模块声明
declare module '*.svg' {
  const content: React.FunctionComponent<React.SVGAttributes<SVGElement>>;
  export default content;
}

declare module '*.png' {
  const content: string;
  export default content;
}

declare module '*.jpg' {
  const content: string;
  export default content;
}

declare module '*.css' {
  const classes: { [key: string]: string };
  export default classes;
}

declare module '*.module.css' {
  const classes: { [key: string]: string };
  export default classes;
}

// 为没有类型定义的第三方库声明类型
declare module 'some-library' {
  export function someFunction(param: string): number;
  export interface SomeInterface {
    prop: string;
  }
}
```

---

## 📋 声明文件：类型定义的艺术

### 创建声明文件

```typescript
// types/api.d.ts - API相关类型声明
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
  timestamp: number;
}

export interface PaginationParams {
  page: number;
  pageSize: number;
  total?: number;
}

export interface PaginatedResponse<T> extends ApiResponse<T[]> {
  pagination: {
    current: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
}

// 用户相关类型
export interface User {
  id: number;
  username: string;
  email: string;
  avatar?: string;
  role: UserRole;
  status: UserStatus;
  createdAt: string;
  updatedAt: string;
}

export type UserRole = 'admin' | 'user' | 'guest';
export type UserStatus = 'active' | 'inactive' | 'banned';

export interface LoginRequest {
  username: string;
  password: string;
  remember?: boolean;
}

export interface LoginResponse {
  token: string;
  user: User;
  expiresIn: number;
}

// 商品相关类型
export interface Product {
  id: number;
  name: string;
  description: string;
  price: string;
  discountPrice?: string;
  categoryId: number;
  categoryName?: string;
  images: string[];
  stock: number;
  sales: number;
  rating?: number;
  status: ProductStatus;
  createdAt: string;
  updatedAt: string;
}

export type ProductStatus = 'active' | 'inactive' | 'out_of_stock';

export interface ProductSearchParams {
  keyword?: string;
  categoryId?: number;
  minPrice?: number;
  maxPrice?: number;
  sortBy?: 'price' | 'sales' | 'rating' | 'created_at';
  sortOrder?: 'asc' | 'desc';
}

// 订单相关类型
export interface Order {
  id: number;
  userId: number;
  orderNo: string;
  status: OrderStatus;
  totalAmount: string;
  paymentMethod: PaymentMethod;
  shippingAddress: Address;
  items: OrderItem[];
  createdAt: string;
  updatedAt: string;
}

export type OrderStatus =
  | 'pending'
  | 'paid'
  | 'shipped'
  | 'delivered'
  | 'cancelled';
export type PaymentMethod = 'alipay' | 'wechat' | 'credit_card' | 'cash';

export interface OrderItem {
  id: number;
  productId: number;
  productName: string;
  productImage: string;
  price: string;
  quantity: number;
  subtotal: string;
}

export interface Address {
  id?: number;
  name: string;
  phone: string;
  province: string;
  city: string;
  district: string;
  detail: string;
  isDefault?: boolean;
}
```

### 环境变量类型声明

```typescript
// types/env.d.ts - 环境变量类型
declare namespace NodeJS {
  interface ProcessEnv {
    // Next.js环境变量
    NODE_ENV: 'development' | 'production' | 'test';

    // 公共环境变量（客户端可访问）
    NEXT_PUBLIC_API_BASE_URL: string;
    NEXT_PUBLIC_APP_NAME: string;
    NEXT_PUBLIC_APP_VERSION: string;
    NEXT_PUBLIC_ENABLE_ANALYTICS: string;

    // 服务端环境变量
    DATABASE_URL: string;
    REDIS_URL: string;
    JWT_SECRET: string;
    JWT_EXPIRES_IN: string;

    // 第三方服务
    ALIPAY_APP_ID: string;
    ALIPAY_PRIVATE_KEY: string;
    WECHAT_APP_ID: string;
    WECHAT_APP_SECRET: string;

    // 文件存储
    OSS_ACCESS_KEY_ID: string;
    OSS_ACCESS_KEY_SECRET: string;
    OSS_BUCKET: string;
    OSS_REGION: string;

    // 邮件服务
    SMTP_HOST: string;
    SMTP_PORT: string;
    SMTP_USER: string;
    SMTP_PASS: string;
  }
}
```

---

## 🔧 实战应用：Mall-Frontend模块化架构

### 项目结构设计

```
src/
├── components/          # 组件模块
│   ├── ui/             # 基础UI组件
│   ├── business/       # 业务组件
│   ├── layout/         # 布局组件
│   └── index.ts        # 组件导出
├── hooks/              # 自定义Hooks
│   ├── useAuth.ts
│   ├── useCart.ts
│   └── index.ts
├── services/           # 服务模块
│   ├── api/           # API服务
│   ├── auth/          # 认证服务
│   ├── storage/       # 存储服务
│   └── index.ts
├── store/              # 状态管理
│   ├── slices/        # Redux切片
│   ├── middleware/    # 中间件
│   └── index.ts
├── utils/              # 工具函数
│   ├── format/        # 格式化工具
│   ├── validation/    # 验证工具
│   ├── constants/     # 常量定义
│   └── index.ts
├── types/              # 类型定义
│   ├── api.d.ts       # API类型
│   ├── user.d.ts      # 用户类型
│   ├── product.d.ts   # 商品类型
│   └── index.ts
└── app/                # Next.js应用
    ├── layout.tsx
    ├── page.tsx
    └── ...
```

### 模块导出策略

```typescript
// components/index.ts - 组件统一导出
// UI组件
export { Button } from './ui/Button';
export { Input } from './ui/Input';
export { Modal } from './ui/Modal';
export { Loading } from './ui/Loading';

// 业务组件
export { ProductCard } from './business/ProductCard';
export { CartItem } from './business/CartItem';
export { OrderHistory } from './business/OrderHistory';
export { UserProfile } from './business/UserProfile';

// 布局组件
export { Header } from './layout/Header';
export { Footer } from './layout/Footer';
export { Sidebar } from './layout/Sidebar';
export { MainLayout } from './layout/MainLayout';

// 类型导出
export type { ButtonProps } from './ui/Button';
export type { InputProps } from './ui/Input';
export type { ProductCardProps } from './business/ProductCard';
```

```typescript
// services/index.ts - 服务统一导出
import { ApiClient } from './api/client';
import { AuthService } from './auth/authService';
import { UserService } from './api/userService';
import { ProductService } from './api/productService';
import { CartService } from './api/cartService';
import { OrderService } from './api/orderService';

// 创建服务实例
const apiClient = new ApiClient();
const authService = new AuthService(apiClient);
const userService = new UserService(apiClient);
const productService = new ProductService(apiClient);
const cartService = new CartService(apiClient);
const orderService = new OrderService(apiClient);

// 服务集合
export const services = {
  auth: authService,
  user: userService,
  product: productService,
  cart: cartService,
  order: orderService,
};

// 单独导出
export {
  ApiClient,
  AuthService,
  UserService,
  ProductService,
  CartService,
  OrderService,
};

// 默认导出
export default services;
```

### 懒加载模块设计

```typescript
// utils/lazyLoader.ts - 懒加载工具
import { lazy, ComponentType } from 'react';

interface LazyComponentOptions {
  fallback?: ComponentType;
  delay?: number;
  retry?: number;
}

export class LazyLoader {
  private static cache = new Map<string, ComponentType<any>>();

  static component<T = {}>(
    importFn: () => Promise<{ default: ComponentType<T> }>,
    options: LazyComponentOptions = {}
  ): ComponentType<T> {
    const cacheKey = importFn.toString();

    if (this.cache.has(cacheKey)) {
      return this.cache.get(cacheKey)!;
    }

    const LazyComponent = lazy(async () => {
      try {
        if (options.delay) {
          await new Promise(resolve => setTimeout(resolve, options.delay));
        }

        return await importFn();
      } catch (error) {
        console.error('懒加载组件失败:', error);

        if (options.retry && options.retry > 0) {
          // 重试逻辑
          await new Promise(resolve => setTimeout(resolve, 1000));
          return this.component(importFn, {
            ...options,
            retry: options.retry - 1,
          });
        }

        throw error;
      }
    });

    this.cache.set(cacheKey, LazyComponent);
    return LazyComponent;
  }

  static async preload(importFn: () => Promise<any>): Promise<void> {
    try {
      await importFn();
    } catch (error) {
      console.error('预加载失败:', error);
    }
  }
}

// 使用示例
export const LazyProductDetail = LazyLoader.component(
  () => import('../components/business/ProductDetail'),
  { delay: 100, retry: 2 }
);

export const LazyCheckout = LazyLoader.component(
  () => import('../components/business/Checkout')
);

// 预加载关键组件
export async function preloadCriticalComponents() {
  await Promise.all([
    LazyLoader.preload(() => import('../components/business/ProductCard')),
    LazyLoader.preload(() => import('../components/business/CartItem')),
  ]);
}
```

---

## 🎯 面试常考知识点

### 1. ES6模块 vs CommonJS

**Q: ES6模块和CommonJS模块有什么区别？**

**A: 主要区别：**

| 特性         | ES6 Modules     | CommonJS                 |
| ------------ | --------------- | ------------------------ |
| **语法**     | `import/export` | `require/module.exports` |
| **加载时机** | 编译时静态加载  | 运行时动态加载           |
| **树摇优化** | 支持            | 不支持                   |
| **循环依赖** | 更好的处理      | 可能有问题               |
| **顶层this** | undefined       | global对象               |
| **动态导入** | `import()`      | `require()`              |

```typescript
// ES6 Modules
import { useState } from 'react'; // 命名导入
import React from 'react'; // 默认导入
import * as utils from './utils'; // 命名空间导入
export const Component = () => {}; // 命名导出
export default Component; // 默认导出

// CommonJS
const { useState } = require('react');
const React = require('react');
const utils = require('./utils');
module.exports = { Component };
module.exports = Component;
```

### 2. 模块解析策略

**Q: TypeScript的模块解析策略有哪些？**

**A: 两种主要策略：**

1. **Classic策略** (已废弃)
2. **Node策略** (推荐)

```typescript
// Node策略解析顺序
import { helper } from './helper';

// 1. ./helper.ts
// 2. ./helper.tsx
// 3. ./helper.d.ts
// 4. ./helper/package.json (main字段)
// 5. ./helper/index.ts
// 6. ./helper/index.tsx
// 7. ./helper/index.d.ts
```

### 3. 命名空间 vs 模块

**Q: 什么时候使用命名空间，什么时候使用模块？**

**A: 使用建议：**

- **使用模块**: 现代TypeScript开发的首选
- **使用命名空间**: 全局库、内部API组织

```typescript
// 推荐：使用模块
// math.ts
export function add(a: number, b: number) {
  return a + b;
}
export function subtract(a: number, b: number) {
  return a - b;
}

// 不推荐：使用命名空间（除非特殊场景）
namespace Math {
  export function add(a: number, b: number) {
    return a + b;
  }
  export function subtract(a: number, b: number) {
    return a - b;
  }
}
```

---

## 🏋️ 实战练习

### 练习1: 设计一个模块化的主题系统

**题目**: 为Mall-Frontend设计一个可扩展的主题系统

**要求**:

1. 支持多种主题（亮色、暗色、高对比度）
2. 支持动态切换主题
3. 支持自定义主题
4. 类型安全的主题配置

**解决方案**:

```typescript
// types/theme.d.ts - 主题类型定义
export interface ThemeColors {
  primary: string;
  secondary: string;
  success: string;
  warning: string;
  error: string;
  info: string;
  background: string;
  surface: string;
  text: {
    primary: string;
    secondary: string;
    disabled: string;
  };
  border: string;
  shadow: string;
}

export interface ThemeSpacing {
  xs: string;
  sm: string;
  md: string;
  lg: string;
  xl: string;
}

export interface ThemeTypography {
  fontFamily: string;
  fontSize: {
    xs: string;
    sm: string;
    md: string;
    lg: string;
    xl: string;
  };
  fontWeight: {
    light: number;
    normal: number;
    medium: number;
    bold: number;
  };
}

export interface Theme {
  name: string;
  colors: ThemeColors;
  spacing: ThemeSpacing;
  typography: ThemeTypography;
  borderRadius: string;
  shadows: string[];
}

export type ThemeName = 'light' | 'dark' | 'highContrast';
```

```typescript
// themes/index.ts - 主题定义
import type { Theme, ThemeName } from '../types/theme';

// 基础主题配置
const baseTheme = {
  spacing: {
    xs: '4px',
    sm: '8px',
    md: '16px',
    lg: '24px',
    xl: '32px',
  },
  typography: {
    fontFamily:
      '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    fontSize: {
      xs: '12px',
      sm: '14px',
      md: '16px',
      lg: '18px',
      xl: '24px',
    },
    fontWeight: {
      light: 300,
      normal: 400,
      medium: 500,
      bold: 700,
    },
  },
  borderRadius: '8px',
  shadows: [
    '0 1px 3px rgba(0, 0, 0, 0.1)',
    '0 4px 6px rgba(0, 0, 0, 0.1)',
    '0 10px 15px rgba(0, 0, 0, 0.1)',
  ],
};

// 亮色主题
export const lightTheme: Theme = {
  name: 'light',
  ...baseTheme,
  colors: {
    primary: '#1890ff',
    secondary: '#722ed1',
    success: '#52c41a',
    warning: '#faad14',
    error: '#ff4d4f',
    info: '#13c2c2',
    background: '#ffffff',
    surface: '#fafafa',
    text: {
      primary: '#262626',
      secondary: '#595959',
      disabled: '#bfbfbf',
    },
    border: '#d9d9d9',
    shadow: 'rgba(0, 0, 0, 0.1)',
  },
};

// 暗色主题
export const darkTheme: Theme = {
  name: 'dark',
  ...baseTheme,
  colors: {
    primary: '#177ddc',
    secondary: '#642ab5',
    success: '#49aa19',
    warning: '#d89614',
    error: '#dc4446',
    info: '#13a8a8',
    background: '#141414',
    surface: '#1f1f1f',
    text: {
      primary: '#ffffff',
      secondary: '#a6a6a6',
      disabled: '#595959',
    },
    border: '#434343',
    shadow: 'rgba(0, 0, 0, 0.3)',
  },
};

// 高对比度主题
export const highContrastTheme: Theme = {
  name: 'highContrast',
  ...baseTheme,
  colors: {
    primary: '#0000ff',
    secondary: '#800080',
    success: '#008000',
    warning: '#ff8c00',
    error: '#ff0000',
    info: '#008080',
    background: '#ffffff',
    surface: '#f0f0f0',
    text: {
      primary: '#000000',
      secondary: '#333333',
      disabled: '#666666',
    },
    border: '#000000',
    shadow: 'rgba(0, 0, 0, 0.5)',
  },
};

// 主题映射
export const themes: Record<ThemeName, Theme> = {
  light: lightTheme,
  dark: darkTheme,
  highContrast: highContrastTheme,
};

// 默认导出
export default themes;
```

```typescript
// hooks/useTheme.ts - 主题Hook
import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import type { Theme, ThemeName } from '../types/theme';
import themes from '../themes';

interface ThemeContextValue {
  theme: Theme;
  themeName: ThemeName;
  setTheme: (themeName: ThemeName) => void;
  toggleTheme: () => void;
}

const ThemeContext = createContext<ThemeContextValue | undefined>(undefined);

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [themeName, setThemeName] = useState<ThemeName>('light');

  // 从localStorage加载主题
  useEffect(() => {
    const savedTheme = localStorage.getItem('theme') as ThemeName;
    if (savedTheme && themes[savedTheme]) {
      setThemeName(savedTheme);
    }
  }, []);

  // 保存主题到localStorage
  useEffect(() => {
    localStorage.setItem('theme', themeName);

    // 更新CSS变量
    const root = document.documentElement;
    const theme = themes[themeName];

    Object.entries(theme.colors).forEach(([key, value]) => {
      if (typeof value === 'string') {
        root.style.setProperty(`--color-${key}`, value);
      } else {
        Object.entries(value).forEach(([subKey, subValue]) => {
          root.style.setProperty(`--color-${key}-${subKey}`, subValue);
        });
      }
    });
  }, [themeName]);

  const setTheme = (newThemeName: ThemeName) => {
    setThemeName(newThemeName);
  };

  const toggleTheme = () => {
    setThemeName(current => current === 'light' ? 'dark' : 'light');
  };

  const value: ThemeContextValue = {
    theme: themes[themeName],
    themeName,
    setTheme,
    toggleTheme,
  };

  return (
    <ThemeContext.Provider value={value}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme(): ThemeContextValue {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
}
```

这个练习展示了：

1. **类型安全的主题系统** - 完整的TypeScript类型定义
2. **模块化设计** - 清晰的文件组织和导出策略
3. **可扩展性** - 易于添加新主题和自定义配置
4. **实际应用** - 与React Context和CSS变量集成

---

## 📚 本章总结

通过本章学习，我们深入掌握了TypeScript的模块系统和命名空间：

### 🎯 核心收获

1. **ES6模块系统** 📦
   - 掌握了现代JavaScript的模块化标准
   - 学会了导入导出的各种语法和最佳实践
   - 理解了动态导入和代码分割的应用

2. **模块解析策略** 🔍
   - 理解了TypeScript的模块解析机制
   - 掌握了路径映射和类型根目录配置
   - 学会了模块增强和声明文件的使用

3. **实战应用** 🛠️
   - 设计了完整的项目模块化架构
   - 实现了懒加载和性能优化策略
   - 构建了可扩展的主题系统

4. **最佳实践** 💡
   - 学会了Barrel导出模式
   - 掌握了模块的组织和管理策略
   - 理解了大型应用的模块化设计原则

### 🚀 技术进阶

- **下一步学习**: React组件设计模式
- **实践建议**: 在项目中应用模块化最佳实践
- **深入方向**: 微前端架构和模块联邦

良好的模块化设计是构建可维护大型应用的基础！ 🎉

---

_下一章我们将学习《React组件设计模式》，探索现代React开发的核心技能！_ 🚀
