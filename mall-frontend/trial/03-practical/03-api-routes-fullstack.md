# 第3章：API Routes与全栈开发实践 🌐

> _"Next.js让前端开发者也能轻松构建全栈应用，一套代码搞定前后端！"_ 🚀

## 📚 本章导览

Next.js的API Routes功能让我们能够在同一个项目中构建前端和后端，实现真正的全栈开发。通过本章学习，我们将深入探讨API Routes的设计模式、最佳实践，以及如何在Mall-Frontend项目中构建高质量的API服务。

### 🎯 学习目标

通过本章学习，你将掌握：

- **API Routes基础** - 理解Next.js API Routes的工作原理和路由系统
- **RESTful API设计** - 掌握现代API设计的最佳实践和规范
- **数据库集成** - 学会集成各种数据库和ORM框架
- **身份认证授权** - 实现JWT、OAuth等认证机制
- **API安全防护** - 掌握API安全的各种防护措施
- **性能优化** - API缓存、限流、监控等优化技巧
- **全栈框架对比** - 对比Next.js与其他全栈解决方案
- **企业级实践** - 大型项目中的API架构设计

### 🛠️ 技术栈概览

```typescript
{
  "framework": "Next.js 15.5.2 API Routes",
  "database": ["PostgreSQL", "MongoDB", "Redis"],
  "orm": ["Prisma", "TypeORM", "Mongoose"],
  "auth": ["NextAuth.js", "JWT", "OAuth 2.0"],
  "validation": ["Zod", "Joi", "Yup"],
  "testing": ["Jest", "Supertest", "MSW"],
  "deployment": ["Vercel", "AWS Lambda", "Docker"],
  "monitoring": ["Sentry", "DataDog", "New Relic"]
}
```

### 📖 本章目录

- [API Routes基础架构](#api-routes基础架构)
- [RESTful API设计规范](#restful-api设计规范)
- [数据库集成与ORM](#数据库集成与orm)
- [身份认证与授权](#身份认证与授权)
- [API安全与防护](#api安全与防护)
- [性能优化策略](#性能优化策略)
- [全栈框架对比](#全栈框架对比)
- [企业级API架构](#企业级api架构)
- [Mall-Frontend实战案例](#mall-frontend实战案例)
- [面试常考知识点](#面试常考知识点)
- [实战练习](#实战练习)

---

## 🏗️ API Routes基础架构

### Next.js API Routes工作原理

Next.js API Routes基于文件系统路由，每个API文件对应一个端点：

```typescript
// app/api/hello/route.ts - 基础API路由
import { NextRequest, NextResponse } from 'next/server';

// GET请求处理
export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url);
  const name = searchParams.get('name') || 'World';

  return NextResponse.json({
    message: `Hello, ${name}!`,
    timestamp: new Date().toISOString(),
  });
}

// POST请求处理
export async function POST(request: NextRequest) {
  try {
    const body = await request.json();

    // 请求验证
    if (!body.name) {
      return NextResponse.json({ error: 'Name is required' }, { status: 400 });
    }

    // 业务逻辑处理
    const result = {
      id: Date.now(),
      name: body.name,
      createdAt: new Date().toISOString(),
    };

    return NextResponse.json(result, { status: 201 });
  } catch (error) {
    return NextResponse.json({ error: 'Invalid JSON' }, { status: 400 });
  }
}

// PUT请求处理
export async function PUT(request: NextRequest) {
  // 更新逻辑
  return NextResponse.json({ message: 'Updated successfully' });
}

// DELETE请求处理
export async function DELETE(request: NextRequest) {
  // 删除逻辑
  return NextResponse.json({ message: 'Deleted successfully' });
}
```

### 动态路由和参数处理

```typescript
// app/api/products/[id]/route.ts - 动态路由
import { NextRequest, NextResponse } from 'next/server';
import { getProduct, updateProduct, deleteProduct } from '@/lib/products';

interface RouteParams {
  params: { id: string };
}

// 获取单个商品
export async function GET(request: NextRequest, { params }: RouteParams) {
  try {
    const productId = parseInt(params.id);

    if (isNaN(productId)) {
      return NextResponse.json(
        { error: 'Invalid product ID' },
        { status: 400 }
      );
    }

    const product = await getProduct(productId);

    if (!product) {
      return NextResponse.json({ error: 'Product not found' }, { status: 404 });
    }

    return NextResponse.json(product);
  } catch (error) {
    console.error('Error fetching product:', error);
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    );
  }
}

// 更新商品
export async function PUT(request: NextRequest, { params }: RouteParams) {
  try {
    const productId = parseInt(params.id);
    const updateData = await request.json();

    const updatedProduct = await updateProduct(productId, updateData);

    return NextResponse.json(updatedProduct);
  } catch (error) {
    return NextResponse.json(
      { error: 'Failed to update product' },
      { status: 500 }
    );
  }
}

// 删除商品
export async function DELETE(request: NextRequest, { params }: RouteParams) {
  try {
    const productId = parseInt(params.id);

    await deleteProduct(productId);

    return NextResponse.json(
      { message: 'Product deleted successfully' },
      { status: 200 }
    );
  } catch (error) {
    return NextResponse.json(
      { error: 'Failed to delete product' },
      { status: 500 }
    );
  }
}
```

### 中间件和请求处理

```typescript
// middleware.ts - 全局中间件
import { NextRequest, NextResponse } from 'next/server';
import { verifyJWT } from '@/lib/auth';

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // API路由中间件
  if (pathname.startsWith('/api/')) {
    // CORS处理
    const response = NextResponse.next();
    response.headers.set('Access-Control-Allow-Origin', '*');
    response.headers.set(
      'Access-Control-Allow-Methods',
      'GET, POST, PUT, DELETE, OPTIONS'
    );
    response.headers.set(
      'Access-Control-Allow-Headers',
      'Content-Type, Authorization'
    );

    // OPTIONS请求处理
    if (request.method === 'OPTIONS') {
      return new Response(null, { status: 200, headers: response.headers });
    }

    // 受保护的API路由
    if (pathname.startsWith('/api/protected/')) {
      const token = request.headers
        .get('Authorization')
        ?.replace('Bearer ', '');

      if (!token) {
        return NextResponse.json(
          { error: 'Authorization token required' },
          { status: 401 }
        );
      }

      try {
        const payload = await verifyJWT(token);
        // 将用户信息添加到请求头
        response.headers.set('X-User-ID', payload.userId.toString());
        response.headers.set('X-User-Role', payload.role);
      } catch (error) {
        return NextResponse.json({ error: 'Invalid token' }, { status: 401 });
      }
    }

    return response;
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/api/:path*', '/protected/:path*'],
};
```

### 错误处理和响应格式

```typescript
// lib/api-response.ts - 统一响应格式
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: any;
  };
  meta?: {
    timestamp: string;
    requestId: string;
    pagination?: {
      page: number;
      limit: number;
      total: number;
      totalPages: number;
    };
  };
}

export class ApiResponseBuilder {
  static success<T>(data: T, meta?: any): ApiResponse<T> {
    return {
      success: true,
      data,
      meta: {
        timestamp: new Date().toISOString(),
        requestId: crypto.randomUUID(),
        ...meta,
      },
    };
  }

  static error(
    code: string,
    message: string,
    details?: any,
    statusCode: number = 500
  ): NextResponse {
    const response: ApiResponse = {
      success: false,
      error: {
        code,
        message,
        details,
      },
      meta: {
        timestamp: new Date().toISOString(),
        requestId: crypto.randomUUID(),
      },
    };

    return NextResponse.json(response, { status: statusCode });
  }

  static paginated<T>(
    data: T[],
    page: number,
    limit: number,
    total: number
  ): ApiResponse<T[]> {
    return {
      success: true,
      data,
      meta: {
        timestamp: new Date().toISOString(),
        requestId: crypto.randomUUID(),
        pagination: {
          page,
          limit,
          total,
          totalPages: Math.ceil(total / limit),
        },
      },
    };
  }
}

// 使用示例
// app/api/products/route.ts
export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const page = parseInt(searchParams.get('page') || '1');
    const limit = parseInt(searchParams.get('limit') || '10');

    const { products, total } = await getProducts({ page, limit });

    const response = ApiResponseBuilder.paginated(products, page, limit, total);
    return NextResponse.json(response);
  } catch (error) {
    return ApiResponseBuilder.error(
      'PRODUCTS_FETCH_ERROR',
      'Failed to fetch products',
      error,
      500
    );
  }
}
```

### 请求验证和数据校验

```typescript
// lib/validation.ts - 数据验证
import { z } from 'zod';

// 商品创建验证模式
export const createProductSchema = z.object({
  name: z.string().min(1, 'Product name is required').max(100),
  description: z.string().min(10, 'Description must be at least 10 characters'),
  price: z.string().regex(/^\d+(\.\d{2})?$/, 'Invalid price format'),
  category_id: z.number().int().positive(),
  stock: z.number().int().min(0),
  images: z.array(z.string().url()).min(1, 'At least one image is required'),
  status: z.enum(['active', 'inactive', 'draft']).default('draft'),
});

// 商品更新验证模式
export const updateProductSchema = createProductSchema.partial();

// 用户注册验证模式
export const registerUserSchema = z.object({
  username: z
    .string()
    .min(3)
    .max(20)
    .regex(/^[a-zA-Z0-9_]+$/),
  email: z.string().email(),
  password: z
    .string()
    .min(8)
    .regex(
      /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]/,
      'Password must contain uppercase, lowercase, number and special character'
    ),
  nickname: z.string().min(1).max(50),
});

// 验证中间件
export function validateRequest<T>(schema: z.ZodSchema<T>) {
  return async (
    request: NextRequest
  ): Promise<
    { data: T; error?: never } | { data?: never; error: NextResponse }
  > => {
    try {
      const body = await request.json();
      const data = schema.parse(body);
      return { data };
    } catch (error) {
      if (error instanceof z.ZodError) {
        return {
          error: ApiResponseBuilder.error(
            'VALIDATION_ERROR',
            'Invalid request data',
            error.errors,
            400
          ),
        };
      }

      return {
        error: ApiResponseBuilder.error(
          'PARSE_ERROR',
          'Invalid JSON format',
          null,
          400
        ),
      };
    }
  };
}

// 使用验证中间件
// app/api/products/route.ts
export async function POST(request: NextRequest) {
  const validation = await validateRequest(createProductSchema)(request);

  if (validation.error) {
    return validation.error;
  }

  try {
    const product = await createProduct(validation.data);
    const response = ApiResponseBuilder.success(product);
    return NextResponse.json(response, { status: 201 });
  } catch (error) {
    return ApiResponseBuilder.error(
      'PRODUCT_CREATE_ERROR',
      'Failed to create product',
      error,
      500
    );
  }
}
```

---

## 🗄️ 数据库集成与ORM

### Prisma集成实践

```typescript
// prisma/schema.prisma - 数据库模式定义
generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

model User {
  id        Int      @id @default(autoincrement())
  username  String   @unique
  email     String   @unique
  password  String
  nickname  String
  avatar    String?
  phone     String?
  role      String   @default("user")
  status    String   @default("active")
  createdAt DateTime @default(now()) @map("created_at")
  updatedAt DateTime @updatedAt @map("updated_at")

  // 关联关系
  orders    Order[]
  cartItems CartItem[]
  reviews   Review[]

  @@map("users")
}

model Product {
  id            Int      @id @default(autoincrement())
  name          String
  description   String
  price         String
  discountPrice String?  @map("discount_price")
  stock         Int
  categoryId    Int      @map("category_id")
  images        String[]
  status        String   @default("active")
  createdAt     DateTime @default(now()) @map("created_at")
  updatedAt     DateTime @updatedAt @map("updated_at")

  // 关联关系
  category    Category   @relation(fields: [categoryId], references: [id])
  cartItems   CartItem[]
  orderItems  OrderItem[]
  reviews     Review[]

  @@map("products")
}

model Category {
  id          Int       @id @default(autoincrement())
  name        String
  description String?
  parentId    Int?      @map("parent_id")
  createdAt   DateTime  @default(now()) @map("created_at")
  updatedAt   DateTime  @updatedAt @map("updated_at")

  // 关联关系
  parent   Category?  @relation("CategoryHierarchy", fields: [parentId], references: [id])
  children Category[] @relation("CategoryHierarchy")
  products Product[]

  @@map("categories")
}

model Order {
  id          Int      @id @default(autoincrement())
  userId      Int      @map("user_id")
  totalAmount String   @map("total_amount")
  status      String   @default("pending")
  createdAt   DateTime @default(now()) @map("created_at")
  updatedAt   DateTime @updatedAt @map("updated_at")

  // 关联关系
  user       User        @relation(fields: [userId], references: [id])
  orderItems OrderItem[]

  @@map("orders")
}

model OrderItem {
  id        Int    @id @default(autoincrement())
  orderId   Int    @map("order_id")
  productId Int    @map("product_id")
  quantity  Int
  price     String

  // 关联关系
  order   Order   @relation(fields: [orderId], references: [id])
  product Product @relation(fields: [productId], references: [id])

  @@map("order_items")
}

model CartItem {
  id        Int     @id @default(autoincrement())
  userId    Int     @map("user_id")
  productId Int     @map("product_id")
  quantity  Int
  selected  Boolean @default(true)

  // 关联关系
  user    User    @relation(fields: [userId], references: [id])
  product Product @relation(fields: [productId], references: [id])

  @@unique([userId, productId])
  @@map("cart_items")
}

model Review {
  id        Int      @id @default(autoincrement())
  userId    Int      @map("user_id")
  productId Int      @map("product_id")
  rating    Int
  comment   String?
  createdAt DateTime @default(now()) @map("created_at")

  // 关联关系
  user    User    @relation(fields: [userId], references: [id])
  product Product @relation(fields: [productId], references: [id])

  @@map("reviews")
}
```

### 数据库操作封装

```typescript
// lib/database.ts - 数据库连接和操作
import { PrismaClient } from '@prisma/client';

// 全局Prisma客户端
declare global {
  var prisma: PrismaClient | undefined;
}

export const prisma =
  globalThis.prisma ||
  new PrismaClient({
    log:
      process.env.NODE_ENV === 'development'
        ? ['query', 'error', 'warn']
        : ['error'],
  });

if (process.env.NODE_ENV !== 'production') {
  globalThis.prisma = prisma;
}

// 数据库操作基类
export abstract class BaseRepository<T> {
  protected abstract model: any;

  async findById(id: number): Promise<T | null> {
    return await this.model.findUnique({
      where: { id },
    });
  }

  async findMany(
    options: {
      where?: any;
      orderBy?: any;
      skip?: number;
      take?: number;
      include?: any;
    } = {}
  ): Promise<T[]> {
    return await this.model.findMany(options);
  }

  async create(data: any): Promise<T> {
    return await this.model.create({
      data,
    });
  }

  async update(id: number, data: any): Promise<T> {
    return await this.model.update({
      where: { id },
      data,
    });
  }

  async delete(id: number): Promise<T> {
    return await this.model.delete({
      where: { id },
    });
  }

  async count(where?: any): Promise<number> {
    return await this.model.count({ where });
  }
}

// 商品仓储实现
export class ProductRepository extends BaseRepository<Product> {
  protected model = prisma.product;

  async findByCategory(
    categoryId: number,
    options: {
      page?: number;
      limit?: number;
      orderBy?: string;
    } = {}
  ): Promise<{ products: Product[]; total: number }> {
    const { page = 1, limit = 10, orderBy = 'createdAt' } = options;
    const skip = (page - 1) * limit;

    const [products, total] = await Promise.all([
      this.model.findMany({
        where: { categoryId, status: 'active' },
        include: {
          category: true,
          reviews: {
            select: {
              rating: true,
            },
          },
        },
        orderBy: { [orderBy]: 'desc' },
        skip,
        take: limit,
      }),
      this.model.count({
        where: { categoryId, status: 'active' },
      }),
    ]);

    return { products, total };
  }

  async search(
    query: string,
    options: {
      page?: number;
      limit?: number;
      categoryId?: number;
      minPrice?: number;
      maxPrice?: number;
    } = {}
  ): Promise<{ products: Product[]; total: number }> {
    const { page = 1, limit = 10, categoryId, minPrice, maxPrice } = options;
    const skip = (page - 1) * limit;

    const where: any = {
      status: 'active',
      OR: [
        { name: { contains: query, mode: 'insensitive' } },
        { description: { contains: query, mode: 'insensitive' } },
      ],
    };

    if (categoryId) {
      where.categoryId = categoryId;
    }

    if (minPrice !== undefined || maxPrice !== undefined) {
      where.price = {};
      if (minPrice !== undefined) {
        where.price.gte = minPrice.toString();
      }
      if (maxPrice !== undefined) {
        where.price.lte = maxPrice.toString();
      }
    }

    const [products, total] = await Promise.all([
      this.model.findMany({
        where,
        include: {
          category: true,
          reviews: {
            select: {
              rating: true,
            },
          },
        },
        orderBy: { createdAt: 'desc' },
        skip,
        take: limit,
      }),
      this.model.count({ where }),
    ]);

    return { products, total };
  }

  async updateStock(id: number, quantity: number): Promise<Product> {
    return await this.model.update({
      where: { id },
      data: {
        stock: {
          decrement: quantity,
        },
      },
    });
  }
}

// 用户仓储实现
export class UserRepository extends BaseRepository<User> {
  protected model = prisma.user;

  async findByEmail(email: string): Promise<User | null> {
    return await this.model.findUnique({
      where: { email },
    });
  }

  async findByUsername(username: string): Promise<User | null> {
    return await this.model.findUnique({
      where: { username },
    });
  }

  async createWithHashedPassword(userData: {
    username: string;
    email: string;
    password: string;
    nickname: string;
  }): Promise<User> {
    const { hashPassword } = await import('@/lib/auth');
    const hashedPassword = await hashPassword(userData.password);

    return await this.model.create({
      data: {
        ...userData,
        password: hashedPassword,
      },
    });
  }
}

// 仓储工厂
export class RepositoryFactory {
  private static productRepo: ProductRepository;
  private static userRepo: UserRepository;

  static getProductRepository(): ProductRepository {
    if (!this.productRepo) {
      this.productRepo = new ProductRepository();
    }
    return this.productRepo;
  }

  static getUserRepository(): UserRepository {
    if (!this.userRepo) {
      this.userRepo = new UserRepository();
    }
    return this.userRepo;
  }
}
```

---

## 🔄 全栈框架对比

### Next.js vs 其他全栈框架

不同全栈框架在API开发上有着各自的特色和优势：

```typescript
// Next.js API Routes - 文件系统路由
// app/api/products/route.ts
export async function GET(request: NextRequest) {
  const products = await getProducts();
  return NextResponse.json(products);
}

export async function POST(request: NextRequest) {
  const data = await request.json();
  const product = await createProduct(data);
  return NextResponse.json(product, { status: 201 });
}
```

```typescript
// Nuxt.js Server API - 文件系统路由
// server/api/products.get.ts
export default defineEventHandler(async event => {
  const products = await getProducts();
  return products;
});

// server/api/products.post.ts
export default defineEventHandler(async event => {
  const data = await readBody(event);
  const product = await createProduct(data);
  setResponseStatus(event, 201);
  return product;
});

// server/api/products/[id].get.ts
export default defineEventHandler(async event => {
  const id = getRouterParam(event, 'id');
  const product = await getProduct(parseInt(id));

  if (!product) {
    throw createError({
      statusCode: 404,
      statusMessage: 'Product not found',
    });
  }

  return product;
});
```

```typescript
// SvelteKit API Routes - 文件系统路由
// src/routes/api/products/+server.ts
import type { RequestHandler } from './$types';
import { json } from '@sveltejs/kit';

export const GET: RequestHandler = async ({ url }) => {
  const products = await getProducts();
  return json(products);
};

export const POST: RequestHandler = async ({ request }) => {
  const data = await request.json();
  const product = await createProduct(data);
  return json(product, { status: 201 });
};

// src/routes/api/products/[id]/+server.ts
export const GET: RequestHandler = async ({ params }) => {
  const product = await getProduct(parseInt(params.id));

  if (!product) {
    return json({ error: 'Product not found' }, { status: 404 });
  }

  return json(product);
};
```

```typescript
// Remix - Loader/Action模式
// app/routes/api.products.tsx
import type { LoaderFunctionArgs, ActionFunctionArgs } from '@remix-run/node';
import { json } from '@remix-run/node';

export const loader = async ({ request }: LoaderFunctionArgs) => {
  const products = await getProducts();
  return json(products);
};

export const action = async ({ request }: ActionFunctionArgs) => {
  const data = await request.json();

  switch (request.method) {
    case 'POST':
      const product = await createProduct(data);
      return json(product, { status: 201 });

    case 'PUT':
      const updatedProduct = await updateProduct(data.id, data);
      return json(updatedProduct);

    case 'DELETE':
      await deleteProduct(data.id);
      return json({ success: true });

    default:
      return json({ error: 'Method not allowed' }, { status: 405 });
  }
};

// app/routes/api.products.$id.tsx
export const loader = async ({ params }: LoaderFunctionArgs) => {
  const product = await getProduct(parseInt(params.id!));

  if (!product) {
    throw new Response('Product not found', { status: 404 });
  }

  return json(product);
};
```

```go
// Go + Gin - 传统后端API
package main

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
)

type ProductController struct {
    productService *ProductService
}

func (pc *ProductController) GetProducts(c *gin.Context) {
    products, err := pc.productService.GetProducts()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": products})
}

func (pc *ProductController) CreateProduct(c *gin.Context) {
    var productData ProductCreateRequest
    if err := c.ShouldBindJSON(&productData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    product, err := pc.productService.CreateProduct(productData)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"data": product})
}

func (pc *ProductController) GetProduct(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }

    product, err := pc.productService.GetProduct(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": product})
}

// 路由设置
func setupRoutes() *gin.Engine {
    r := gin.Default()

    api := r.Group("/api")
    {
        products := api.Group("/products")
        {
            products.GET("", productController.GetProducts)
            products.POST("", productController.CreateProduct)
            products.GET("/:id", productController.GetProduct)
            products.PUT("/:id", productController.UpdateProduct)
            products.DELETE("/:id", productController.DeleteProduct)
        }
    }

    return r
}
```

### 框架特性对比表

```typescript
// 全栈框架API开发特性对比
interface FrameworkFeatures {
  framework: string;
  routingStyle: 'file-based' | 'code-based';
  typeScript: 'native' | 'excellent' | 'good' | 'basic';
  middleware: 'built-in' | 'third-party' | 'custom';
  validation: 'manual' | 'built-in' | 'ecosystem';
  database: 'agnostic' | 'opinionated' | 'flexible';
  deployment: 'easy' | 'medium' | 'complex';
  performance: 'excellent' | 'good' | 'average';
  ecosystem: 'rich' | 'growing' | 'limited';
  learningCurve: 'easy' | 'medium' | 'steep';
}

const frameworkComparison: FrameworkFeatures[] = [
  {
    framework: 'Next.js',
    routingStyle: 'file-based',
    typeScript: 'native',
    middleware: 'built-in',
    validation: 'ecosystem',
    database: 'agnostic',
    deployment: 'easy',
    performance: 'excellent',
    ecosystem: 'rich',
    learningCurve: 'medium',
  },
  {
    framework: 'Nuxt.js',
    routingStyle: 'file-based',
    typeScript: 'excellent',
    middleware: 'built-in',
    validation: 'ecosystem',
    database: 'agnostic',
    deployment: 'easy',
    performance: 'excellent',
    ecosystem: 'rich',
    learningCurve: 'easy',
  },
  {
    framework: 'SvelteKit',
    routingStyle: 'file-based',
    typeScript: 'excellent',
    middleware: 'built-in',
    validation: 'manual',
    database: 'agnostic',
    deployment: 'easy',
    performance: 'excellent',
    ecosystem: 'growing',
    learningCurve: 'easy',
  },
  {
    framework: 'Remix',
    routingStyle: 'file-based',
    typeScript: 'good',
    middleware: 'custom',
    validation: 'manual',
    database: 'agnostic',
    deployment: 'medium',
    performance: 'good',
    ecosystem: 'growing',
    learningCurve: 'steep',
  },
  {
    framework: 'Express.js',
    routingStyle: 'code-based',
    typeScript: 'good',
    middleware: 'third-party',
    validation: 'ecosystem',
    database: 'agnostic',
    deployment: 'complex',
    performance: 'good',
    ecosystem: 'rich',
    learningCurve: 'medium',
  },
  {
    framework: 'Fastify',
    routingStyle: 'code-based',
    typeScript: 'excellent',
    middleware: 'built-in',
    validation: 'built-in',
    database: 'agnostic',
    deployment: 'complex',
    performance: 'excellent',
    ecosystem: 'growing',
    learningCurve: 'medium',
  },
];
```

---

## 🎯 面试常考知识点

### 1. API设计最佳实践

**Q: 设计RESTful API时需要遵循哪些原则？**

**A: RESTful API设计核心原则：**

```typescript
// 1. 资源导向的URL设计
const apiEndpoints = {
  // ✅ 正确的资源导向设计
  products: {
    list: 'GET /api/products',
    create: 'POST /api/products',
    detail: 'GET /api/products/:id',
    update: 'PUT /api/products/:id',
    delete: 'DELETE /api/products/:id',
  },

  // ✅ 嵌套资源
  productReviews: {
    list: 'GET /api/products/:id/reviews',
    create: 'POST /api/products/:id/reviews',
  },

  // ❌ 错误的动词导向设计
  wrongDesign: {
    getProducts: 'GET /api/getProducts', // 应该是 GET /api/products
    createProduct: 'POST /api/createProduct', // 应该是 POST /api/products
    deleteProduct: 'POST /api/deleteProduct', // 应该是 DELETE /api/products/:id
  },
};

// 2. HTTP状态码的正确使用
const statusCodes = {
  success: {
    200: 'OK - 请求成功',
    201: 'Created - 资源创建成功',
    204: 'No Content - 删除成功，无返回内容',
  },
  clientError: {
    400: 'Bad Request - 请求参数错误',
    401: 'Unauthorized - 未认证',
    403: 'Forbidden - 无权限',
    404: 'Not Found - 资源不存在',
    409: 'Conflict - 资源冲突',
    422: 'Unprocessable Entity - 验证失败',
  },
  serverError: {
    500: 'Internal Server Error - 服务器内部错误',
    502: 'Bad Gateway - 网关错误',
    503: 'Service Unavailable - 服务不可用',
  },
};

// 3. 统一的响应格式
interface StandardApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: any;
  };
  meta?: {
    timestamp: string;
    requestId: string;
    pagination?: PaginationMeta;
  };
}

// 4. 版本控制策略
const versioningStrategies = {
  urlPath: '/api/v1/products', // 推荐
  queryParam: '/api/products?version=1', // 可选
  header: 'Accept: application/vnd.api+json;version=1', // 高级
};
```

### 2. Next.js API Routes深度理解

**Q: Next.js API Routes相比传统Express.js有什么优势？**

**A: 核心优势对比：**

```typescript
// Next.js API Routes优势
const nextjsAdvantages = {
  fileSystemRouting: {
    description: '基于文件系统的自动路由',
    example: 'app/api/products/[id]/route.ts 自动映射到 /api/products/:id',
    benefit: '减少路由配置，提高开发效率',
  },

  typeScriptFirst: {
    description: '原生TypeScript支持',
    example: 'NextRequest, NextResponse 提供完整类型支持',
    benefit: '更好的开发体验和类型安全',
  },

  serverlessReady: {
    description: '天然支持Serverless部署',
    example: 'Vercel Functions, AWS Lambda',
    benefit: '自动扩缩容，按需付费',
  },

  edgeRuntime: {
    description: '支持Edge Runtime',
    example: 'export const runtime = "edge"',
    benefit: '全球边缘计算，降低延迟',
  },

  builtInOptimizations: {
    description: '内置性能优化',
    example: '自动代码分割，Tree Shaking',
    benefit: '更小的包体积，更快的启动时间',
  },
};

// Express.js对比
const expressComparison = {
  routing: '需要手动配置路由',
  typeScript: '需要额外配置和类型定义',
  deployment: '需要服务器管理',
  scaling: '需要手动配置负载均衡',
  optimization: '需要手动配置各种优化',
};
```

### 3. 数据库集成和ORM选择

**Q: 在Next.js项目中如何选择合适的数据库和ORM？**

**A: 数据库和ORM选择指南：**

```typescript
// 数据库选择矩阵
interface DatabaseChoice {
  database: string;
  useCase: string[];
  pros: string[];
  cons: string[];
  scalability: 'low' | 'medium' | 'high';
  complexity: 'low' | 'medium' | 'high';
}

const databaseOptions: DatabaseChoice[] = [
  {
    database: 'PostgreSQL',
    useCase: ['复杂查询', '事务处理', '数据一致性'],
    pros: ['ACID特性', '丰富的数据类型', '强大的查询能力'],
    cons: ['配置复杂', '资源消耗较高'],
    scalability: 'high',
    complexity: 'medium',
  },
  {
    database: 'MongoDB',
    useCase: ['文档存储', '快速原型', '灵活schema'],
    pros: ['Schema灵活', '水平扩展', 'JSON原生支持'],
    cons: ['缺乏事务', '内存消耗大'],
    scalability: 'high',
    complexity: 'low',
  },
  {
    database: 'SQLite',
    useCase: ['小型应用', '开发测试', '嵌入式'],
    pros: ['零配置', '轻量级', '文件数据库'],
    cons: ['并发限制', '功能有限'],
    scalability: 'low',
    complexity: 'low',
  },
];

// ORM选择对比
const ormComparison = {
  Prisma: {
    pros: ['类型安全', '自动生成客户端', '迁移管理', '查询优化'],
    cons: ['学习曲线', '包体积较大'],
    bestFor: '新项目，TypeScript优先',
  },

  TypeORM: {
    pros: ['装饰器语法', '活跃记录模式', '多数据库支持'],
    cons: ['性能问题', '复杂配置'],
    bestFor: '传统ORM用户，复杂关系',
  },

  Mongoose: {
    pros: ['MongoDB专用', '中间件支持', '验证内置'],
    cons: ['仅支持MongoDB', '回调地狱'],
    bestFor: 'MongoDB项目，文档数据库',
  },
};
```

### 4. API安全最佳实践

**Q: 如何保证API的安全性？**

**A: 多层次安全防护：**

```typescript
// API安全检查清单
const securityChecklist = {
  authentication: [
    '✅ 实现JWT认证',
    '✅ 设置合理的token过期时间',
    '✅ 支持token刷新机制',
    '✅ 实现OAuth2.0集成',
  ],

  authorization: [
    '✅ 基于角色的访问控制(RBAC)',
    '✅ 资源级权限检查',
    '✅ API端点权限验证',
    '✅ 最小权限原则',
  ],

  inputValidation: [
    '✅ 使用schema验证(Zod/Joi)',
    '✅ SQL注入防护',
    '✅ XSS攻击防护',
    '✅ 文件上传安全检查',
  ],

  rateLimit: [
    '✅ API调用频率限制',
    '✅ IP白名单/黑名单',
    '✅ DDoS攻击防护',
    '✅ 用户级别限流',
  ],

  dataProtection: [
    '✅ HTTPS强制使用',
    '✅ 敏感数据加密',
    '✅ 数据库连接加密',
    '✅ 日志脱敏处理',
  ],
};

// 安全中间件实现示例
export function securityMiddleware() {
  return async (request: NextRequest) => {
    // 1. CORS检查
    const origin = request.headers.get('origin');
    const allowedOrigins = process.env.ALLOWED_ORIGINS?.split(',') || [];

    if (origin && !allowedOrigins.includes(origin)) {
      return new Response('CORS policy violation', { status: 403 });
    }

    // 2. Rate Limiting
    const ip = request.ip || 'unknown';
    const rateLimitKey = `rate_limit:${ip}`;
    const currentRequests = (await redis.get(rateLimitKey)) || 0;

    if (currentRequests > 100) {
      // 每分钟100次请求
      return new Response('Rate limit exceeded', { status: 429 });
    }

    await redis.setex(rateLimitKey, 60, currentRequests + 1);

    // 3. 安全头设置
    const response = NextResponse.next();
    response.headers.set('X-Content-Type-Options', 'nosniff');
    response.headers.set('X-Frame-Options', 'DENY');
    response.headers.set('X-XSS-Protection', '1; mode=block');
    response.headers.set('Strict-Transport-Security', 'max-age=31536000');

    return response;
  };
}
```

---

## 📚 实战练习

### 练习1：设计电商API

**任务**: 为Mall-Frontend设计完整的商品管理API，包括CRUD操作、搜索、分类筛选等功能。

**要求**:

- 使用TypeScript和Zod进行类型定义和验证
- 实现分页、排序、筛选功能
- 添加适当的错误处理和响应格式
- 考虑性能优化和缓存策略

**参考实现**:

```typescript
// app/api/products/route.ts
import { NextRequest, NextResponse } from 'next/server';
import { z } from 'zod';
import { RepositoryFactory } from '@/lib/database';
import { ApiResponseBuilder } from '@/lib/api-response';

const productQuerySchema = z.object({
  page: z.string().transform(Number).default('1'),
  limit: z.string().transform(Number).default('10'),
  category: z.string().optional(),
  search: z.string().optional(),
  minPrice: z.string().transform(Number).optional(),
  maxPrice: z.string().transform(Number).optional(),
  sortBy: z.enum(['name', 'price', 'createdAt']).default('createdAt'),
  sortOrder: z.enum(['asc', 'desc']).default('desc'),
});

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const query = productQuerySchema.parse(Object.fromEntries(searchParams));

    const productRepo = RepositoryFactory.getProductRepository();

    let result;
    if (query.search) {
      result = await productRepo.search(query.search, {
        page: query.page,
        limit: query.limit,
        categoryId: query.category ? parseInt(query.category) : undefined,
        minPrice: query.minPrice,
        maxPrice: query.maxPrice,
      });
    } else if (query.category) {
      result = await productRepo.findByCategory(parseInt(query.category), {
        page: query.page,
        limit: query.limit,
        orderBy: query.sortBy,
      });
    } else {
      result = await productRepo.findMany({
        skip: (query.page - 1) * query.limit,
        take: query.limit,
        orderBy: { [query.sortBy]: query.sortOrder },
      });

      const total = await productRepo.count();
      result = { products: result, total };
    }

    const response = ApiResponseBuilder.paginated(
      result.products,
      query.page,
      query.limit,
      result.total
    );

    return NextResponse.json(response);
  } catch (error) {
    return ApiResponseBuilder.error(
      'PRODUCTS_FETCH_ERROR',
      'Failed to fetch products',
      error,
      500
    );
  }
}
```

### 练习2：实现用户认证系统

**任务**: 实现完整的用户认证系统，包括注册、登录、JWT验证、密码重置等功能。

**要求**:

- 使用bcrypt进行密码哈希
- 实现JWT token生成和验证
- 添加邮箱验证功能
- 实现密码重置流程
- 添加登录尝试限制

### 练习3：API性能优化

**任务**: 对现有API进行性能优化，包括缓存、数据库查询优化、响应压缩等。

**要求**:

- 实现Redis缓存
- 优化数据库查询（避免N+1问题）
- 添加响应压缩
- 实现API响应时间监控

---

## 📚 本章总结

通过本章学习，我们全面掌握了Next.js API Routes的开发实践：

### 🎯 核心收获

1. **API Routes精通** 🏗️
   - 掌握了Next.js API Routes的工作原理和路由系统
   - 学会了动态路由、中间件、错误处理的实现
   - 理解了文件系统路由的优势和最佳实践

2. **数据库集成** 🗄️
   - 掌握了Prisma ORM的使用和数据库设计
   - 学会了仓储模式和数据访问层的封装
   - 理解了数据库事务和性能优化

3. **身份认证授权** 🔐
   - 掌握了JWT认证和OAuth集成
   - 学会了密码安全和权限控制
   - 理解了认证中间件和安全防护

4. **全栈框架对比** 🔄
   - 深入对比了Next.js与其他全栈框架
   - 理解了不同框架的设计理念和适用场景
   - 掌握了技术选型的决策依据

5. **企业级实践** 🚀
   - 学会了API安全防护和性能优化
   - 掌握了错误处理和监控策略
   - 理解了大型项目的API架构设计

### 🚀 技术进阶

- **下一步学习**: 数据获取与缓存策略优化
- **实践建议**: 在项目中应用API设计最佳实践
- **深入方向**: 微服务架构和API网关设计

API Routes让前端开发者也能轻松构建全栈应用，是现代Web开发的重要技能！ 🎉

---

_下一章我们将学习《数据获取与缓存策略优化》，探索高性能数据处理技术！_ 🚀
