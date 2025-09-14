# 第3章：装饰器与元数据编程 🎭

> *"装饰器是TypeScript的魔法糖衣，让代码更优雅、更强大！"* ✨

## 📚 本章导览

装饰器（Decorators）是TypeScript中一个强大而优雅的特性，它允许我们以声明式的方式修改类、方法、属性和参数的行为。在Mall-Frontend项目中，我们将看到装饰器如何简化代码、增强功能，让开发变得更加高效！

### 🎯 学习目标

通过本章学习，你将掌握：

- **装饰器基础概念** - 理解装饰器的工作原理和应用场景
- **类装饰器应用** - 学会使用类装饰器增强类功能
- **方法装饰器实践** - 掌握方法装饰器的常见用法
- **属性装饰器设计** - 理解属性装饰器的设计模式
- **参数装饰器应用** - 学会参数装饰器的实际应用
- **元数据编程** - 掌握Reflect Metadata的使用
- **实战应用** - 在Mall-Frontend中应用装饰器模式

### 🛠️ 技术栈概览

```typescript
{
  "decorators": "TypeScript 5.0+",
  "metadata": "reflect-metadata",
  "frameworks": ["Angular", "NestJS", "TypeORM"],
  "applications": ["依赖注入", "验证", "缓存", "日志"]
}
```

---

## 🌟 装饰器基础：语法糖的魔力

### 什么是装饰器？

装饰器是一种特殊类型的声明，它能够被附加到类声明、方法、访问符、属性或参数上。装饰器使用 `@expression` 这种形式，其中 `expression` 必须求值为一个函数。

```typescript
// 启用装饰器支持
// tsconfig.json
{
  "compilerOptions": {
    "experimentalDecorators": true,
    "emitDecoratorMetadata": true
  }
}

// 简单的装饰器示例
function log(target: any, propertyKey: string, descriptor: PropertyDescriptor) {
  const originalMethod = descriptor.value;

  descriptor.value = function(...args: any[]) {
    console.log(`调用方法 ${propertyKey}，参数:`, args);
    const result = originalMethod.apply(this, args);
    console.log(`方法 ${propertyKey} 返回:`, result);
    return result;
  };
}

class Calculator {
  @log
  add(a: number, b: number): number {
    return a + b;
  }
}

const calc = new Calculator();
calc.add(2, 3); // 自动记录日志
```

### 🔄 语言对比：装饰器模式实现

```java
// Java - 注解（Annotations）
import java.lang.annotation.*;
import java.lang.reflect.Method;

// 定义注解
@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.METHOD)
public @interface Log {
    String value() default "";
}

// 使用注解
public class Calculator {
    @Log("计算方法")
    public int add(int a, int b) {
        return a + b;
    }
}

// 注解处理器（需要反射）
public class LogProcessor {
    public static void processAnnotations(Object obj) {
        Class<?> clazz = obj.getClass();
        for (Method method : clazz.getDeclaredMethods()) {
            if (method.isAnnotationPresent(Log.class)) {
                Log logAnnotation = method.getAnnotation(Log.class);
                System.out.println("方法 " + method.getName() + " 有日志注解: " + logAnnotation.value());
            }
        }
    }
}

// 使用示例
Calculator calc = new Calculator();
LogProcessor.processAnnotations(calc); // 需要手动处理
```

```python
# Python - 装饰器（原生支持）
import functools
from typing import Any, Callable

# 函数装饰器
def log(func: Callable) -> Callable:
    @functools.wraps(func)
    def wrapper(*args, **kwargs):
        print(f"调用方法 {func.__name__}，参数: {args}, {kwargs}")
        result = func(*args, **kwargs)
        print(f"方法 {func.__name__} 返回: {result}")
        return result
    return wrapper

# 类装饰器
def singleton(cls):
    instances = {}
    def get_instance(*args, **kwargs):
        if cls not in instances:
            instances[cls] = cls(*args, **kwargs)
        return instances[cls]
    return get_instance

# 使用装饰器
class Calculator:
    @log
    def add(self, a: int, b: int) -> int:
        return a + b

@singleton
class DatabaseConnection:
    def __init__(self):
        print("创建数据库连接")

# 使用示例
calc = Calculator()
calc.add(2, 3)  # 自动记录日志

# 单例模式
db1 = DatabaseConnection()
db2 = DatabaseConnection()
print(db1 is db2)  # True
```

```csharp
// C# - 特性（Attributes）
using System;
using System.Reflection;

// 定义特性
[AttributeUsage(AttributeTargets.Method)]
public class LogAttribute : Attribute
{
    public string Message { get; }

    public LogAttribute(string message = "")
    {
        Message = message;
    }
}

// 使用特性
public class Calculator
{
    [Log("计算方法")]
    public int Add(int a, int b)
    {
        return a + b;
    }
}

// 特性处理（需要反射）
public class LogProcessor
{
    public static void ProcessAttributes(object obj)
    {
        Type type = obj.GetType();
        foreach (MethodInfo method in type.GetMethods())
        {
            LogAttribute logAttr = method.GetCustomAttribute<LogAttribute>();
            if (logAttr != null)
            {
                Console.WriteLine($"方法 {method.Name} 有日志特性: {logAttr.Message}");
            }
        }
    }
}
```

```go
// Go - 没有原生装饰器，使用函数包装
package main

import (
    "fmt"
    "reflect"
    "runtime"
)

// 函数装饰器模式
func LogDecorator(fn interface{}) interface{} {
    fnValue := reflect.ValueOf(fn)
    fnType := fnValue.Type()

    return reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
        fnName := runtime.FuncForPC(fnValue.Pointer()).Name()
        fmt.Printf("调用方法 %s，参数: %v\n", fnName, args)

        results := fnValue.Call(args)

        fmt.Printf("方法 %s 返回: %v\n", fnName, results)
        return results
    }).Interface()
}

// 使用示例
func add(a, b int) int {
    return a + b
}

func main() {
    // 包装函数
    loggedAdd := LogDecorator(add).(func(int, int) int)

    result := loggedAdd(2, 3)
    fmt.Println("结果:", result)
}
```

**💡 装饰器/注解对比：**

| 特性 | TypeScript | Java | Python | C# | Go |
|------|------------|------|--------|----|----|
| **原生支持** | 实验性 | 注解 | 原生装饰器 | 特性 | 无（函数包装） |
| **运行时处理** | 编译时 | 反射处理 | 自动处理 | 反射处理 | 手动包装 |
| **元数据** | reflect-metadata | 注解属性 | 函数属性 | 特性属性 | 反射 |
| **性能影响** | 编译时 | 运行时反射 | 函数调用 | 运行时反射 | 函数调用 |
| **应用场景** | 框架增强 | 配置标记 | 通用装饰 | 配置标记 | 函数增强 |

### 装饰器的执行顺序

```typescript
function classDecorator(constructor: Function) {
  console.log('类装饰器执行');
}

function methodDecorator(target: any, propertyKey: string, descriptor: PropertyDescriptor) {
  console.log('方法装饰器执行:', propertyKey);
}

function propertyDecorator(target: any, propertyKey: string) {
  console.log('属性装饰器执行:', propertyKey);
}

function parameterDecorator(target: any, propertyKey: string, parameterIndex: number) {
  console.log('参数装饰器执行:', propertyKey, parameterIndex);
}

@classDecorator
class Example {
  @propertyDecorator
  name: string = 'example';

  @methodDecorator
  greet(@parameterDecorator message: string) {
    return `Hello, ${message}`;
  }
}

// 执行顺序：
// 1. 属性装饰器执行: name
// 2. 参数装饰器执行: greet 0
// 3. 方法装饰器执行: greet
// 4. 类装饰器执行
```

---

## 🏗️ 类装饰器：增强类的能力

### 基础类装饰器

```typescript
// 类装饰器工厂
function Entity(tableName: string) {
  return function<T extends { new(...args: any[]): {} }>(constructor: T) {
    return class extends constructor {
      tableName = tableName;
      
      save() {
        console.log(`保存到表 ${tableName}`);
      }
      
      static findAll() {
        console.log(`从表 ${tableName} 查询所有记录`);
      }
    };
  };
}

// 应用到Mall-Frontend的用户实体
@Entity('users')
class User {
  constructor(
    public id: number,
    public username: string,
    public email: string
  ) {}
}

const user = new User(1, 'john', 'john@example.com');
user.save(); // 保存到表 users
User.findAll(); // 从表 users 查询所有记录
```

### 单例模式装饰器

```typescript
// 单例装饰器
function Singleton<T extends { new(...args: any[]): {} }>(constructor: T) {
  let instance: T;
  
  return class extends constructor {
    constructor(...args: any[]) {
      if (instance) {
        return instance;
      }
      super(...args);
      instance = this as any;
      return instance;
    }
  } as T;
}

// Mall-Frontend中的API客户端单例
@Singleton
class ApiClient {
  private baseURL: string;
  
  constructor(baseURL: string = 'http://localhost:8080') {
    this.baseURL = baseURL;
    console.log('ApiClient 实例创建');
  }
  
  async get(endpoint: string) {
    return fetch(`${this.baseURL}${endpoint}`);
  }
  
  async post(endpoint: string, data: any) {
    return fetch(`${this.baseURL}${endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
  }
}

// 无论创建多少次，都是同一个实例
const client1 = new ApiClient();
const client2 = new ApiClient();
console.log(client1 === client2); // true
```

### 配置注入装饰器

```typescript
// 配置注入装饰器
function Injectable(config?: any) {
  return function<T extends { new(...args: any[]): {} }>(constructor: T) {
    // 将配置注入到类的原型中
    if (config) {
      Object.assign(constructor.prototype, { config });
    }
    
    // 添加依赖注入标记
    Reflect.defineMetadata('injectable', true, constructor);
    
    return constructor;
  };
}

// Mall-Frontend中的服务类
@Injectable({
  timeout: 5000,
  retries: 3,
  baseURL: process.env.NEXT_PUBLIC_API_BASE_URL
})
class ProductService {
  config: any;
  
  async getProducts(params: any) {
    console.log('使用配置:', this.config);
    // 实际的API调用逻辑
    return fetch(`${this.config.baseURL}/api/products`, {
      signal: AbortSignal.timeout(this.config.timeout),
    });
  }
  
  async getProduct(id: number) {
    return fetch(`${this.config.baseURL}/api/products/${id}`);
  }
}
```

---

## 🔧 方法装饰器：增强方法功能

### 缓存装饰器

```typescript
// 缓存装饰器
function Cache(ttl: number = 60000) { // 默认缓存1分钟
  const cache = new Map<string, { value: any; expiry: number }>();
  
  return function(target: any, propertyKey: string, descriptor: PropertyDescriptor) {
    const originalMethod = descriptor.value;
    
    descriptor.value = async function(...args: any[]) {
      const cacheKey = `${propertyKey}_${JSON.stringify(args)}`;
      const cached = cache.get(cacheKey);
      
      // 检查缓存是否有效
      if (cached && Date.now() < cached.expiry) {
        console.log(`缓存命中: ${cacheKey}`);
        return cached.value;
      }
      
      // 执行原方法
      const result = await originalMethod.apply(this, args);
      
      // 存储到缓存
      cache.set(cacheKey, {
        value: result,
        expiry: Date.now() + ttl
      });
      
      console.log(`缓存存储: ${cacheKey}`);
      return result;
    };
  };
}

// Mall-Frontend中的商品服务
class ProductService {
  @Cache(300000) // 缓存5分钟
  async getProducts(category?: string, page: number = 1) {
    console.log('从API获取商品数据...');
    const response = await fetch(`/api/products?category=${category}&page=${page}`);
    return response.json();
  }
  
  @Cache(600000) // 缓存10分钟
  async getProductDetail(id: number) {
    console.log('从API获取商品详情...');
    const response = await fetch(`/api/products/${id}`);
    return response.json();
  }
}
```

### 重试装饰器

```typescript
// 重试装饰器
function Retry(maxAttempts: number = 3, delay: number = 1000) {
  return function(target: any, propertyKey: string, descriptor: PropertyDescriptor) {
    const originalMethod = descriptor.value;
    
    descriptor.value = async function(...args: any[]) {
      let lastError: Error;
      
      for (let attempt = 1; attempt <= maxAttempts; attempt++) {
        try {
          return await originalMethod.apply(this, args);
        } catch (error) {
          lastError = error as Error;
          console.log(`方法 ${propertyKey} 第 ${attempt} 次尝试失败:`, error);
          
          if (attempt < maxAttempts) {
            console.log(`等待 ${delay}ms 后重试...`);
            await new Promise(resolve => setTimeout(resolve, delay));
          }
        }
      }
      
      throw new Error(`方法 ${propertyKey} 在 ${maxAttempts} 次尝试后仍然失败: ${lastError.message}`);
    };
  };
}

// Mall-Frontend中的支付服务
class PaymentService {
  @Retry(3, 2000) // 最多重试3次，间隔2秒
  async processPayment(orderId: number, amount: number) {
    console.log(`处理订单 ${orderId} 的支付，金额: ${amount}`);
    
    // 模拟可能失败的支付API调用
    const response = await fetch('/api/payments', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ orderId, amount }),
    });
    
    if (!response.ok) {
      throw new Error(`支付失败: ${response.status}`);
    }
    
    return response.json();
  }
}
```

### 性能监控装饰器

```typescript
// 性能监控装饰器
function Performance(threshold: number = 1000) {
  return function(target: any, propertyKey: string, descriptor: PropertyDescriptor) {
    const originalMethod = descriptor.value;
    
    descriptor.value = async function(...args: any[]) {
      const startTime = performance.now();
      
      try {
        const result = await originalMethod.apply(this, args);
        const endTime = performance.now();
        const duration = endTime - startTime;
        
        if (duration > threshold) {
          console.warn(`⚠️ 方法 ${propertyKey} 执行时间过长: ${duration.toFixed(2)}ms`);
        } else {
          console.log(`✅ 方法 ${propertyKey} 执行时间: ${duration.toFixed(2)}ms`);
        }
        
        return result;
      } catch (error) {
        const endTime = performance.now();
        const duration = endTime - startTime;
        console.error(`❌ 方法 ${propertyKey} 执行失败 (${duration.toFixed(2)}ms):`, error);
        throw error;
      }
    };
  };
}

// Mall-Frontend中的搜索服务
class SearchService {
  @Performance(500) // 超过500ms就警告
  async searchProducts(query: string, filters: any) {
    console.log(`搜索商品: ${query}`);
    
    // 模拟复杂的搜索逻辑
    const response = await fetch('/api/search', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ query, filters }),
    });
    
    return response.json();
  }
  
  @Performance(200) // 超过200ms就警告
  async getSearchSuggestions(query: string) {
    const response = await fetch(`/api/search/suggestions?q=${query}`);
    return response.json();
  }
}
```

---

## 🏷️ 属性装饰器：属性的魔法增强

### 验证装饰器

```typescript
// 验证装饰器
function Required(target: any, propertyKey: string) {
  // 存储需要验证的属性
  const requiredProperties = Reflect.getMetadata('required', target) || [];
  requiredProperties.push(propertyKey);
  Reflect.defineMetadata('required', requiredProperties, target);
}

function MinLength(length: number) {
  return function(target: any, propertyKey: string) {
    const minLengthProperties = Reflect.getMetadata('minLength', target) || {};
    minLengthProperties[propertyKey] = length;
    Reflect.defineMetadata('minLength', minLengthProperties, target);
  };
}

function Email(target: any, propertyKey: string) {
  const emailProperties = Reflect.getMetadata('email', target) || [];
  emailProperties.push(propertyKey);
  Reflect.defineMetadata('email', emailProperties, target);
}

// 验证函数
function validate(obj: any): string[] {
  const errors: string[] = [];
  
  // 检查必填字段
  const requiredProperties = Reflect.getMetadata('required', obj) || [];
  for (const prop of requiredProperties) {
    if (!obj[prop]) {
      errors.push(`${prop} 是必填字段`);
    }
  }
  
  // 检查最小长度
  const minLengthProperties = Reflect.getMetadata('minLength', obj) || {};
  for (const [prop, minLength] of Object.entries(minLengthProperties)) {
    if (obj[prop] && obj[prop].length < (minLength as number)) {
      errors.push(`${prop} 最少需要 ${minLength} 个字符`);
    }
  }
  
  // 检查邮箱格式
  const emailProperties = Reflect.getMetadata('email', obj) || [];
  for (const prop of emailProperties) {
    if (obj[prop] && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(obj[prop])) {
      errors.push(`${prop} 邮箱格式不正确`);
    }
  }
  
  return errors;
}

// Mall-Frontend中的用户注册表单
class UserRegistrationForm {
  @Required
  @MinLength(3)
  username: string = '';
  
  @Required
  @Email
  email: string = '';
  
  @Required
  @MinLength(6)
  password: string = '';
  
  constructor(data: Partial<UserRegistrationForm>) {
    Object.assign(this, data);
  }
  
  validate(): string[] {
    return validate(this);
  }
}

// 使用示例
const form = new UserRegistrationForm({
  username: 'jo',
  email: 'invalid-email',
  password: '123'
});

const errors = form.validate();
console.log(errors);
// ['username 最少需要 3 个字符', 'email 邮箱格式不正确', 'password 最少需要 6 个字符']
```

### 格式化装饰器

```typescript
// 格式化装饰器
function Format(formatter: (value: any) => any) {
  return function(target: any, propertyKey: string) {
    let value = target[propertyKey];
    
    Object.defineProperty(target, propertyKey, {
      get() {
        return value;
      },
      set(newValue: any) {
        value = formatter(newValue);
      },
      enumerable: true,
      configurable: true
    });
  };
}

// 常用格式化函数
const formatters = {
  currency: (value: number) => `¥${value.toFixed(2)}`,
  phone: (value: string) => value.replace(/(\d{3})(\d{4})(\d{4})/, '$1-$2-$3'),
  date: (value: Date) => value.toLocaleDateString('zh-CN'),
  uppercase: (value: string) => value.toUpperCase(),
  trim: (value: string) => value.trim()
};

// Mall-Frontend中的商品模型
class Product {
  @Format(formatters.trim)
  @Format(formatters.uppercase)
  name: string = '';
  
  @Format(formatters.currency)
  price: number = 0;
  
  @Format(formatters.date)
  createdAt: Date = new Date();
  
  constructor(data: Partial<Product>) {
    Object.assign(this, data);
  }
}

const product = new Product({
  name: '  iphone 15 pro  ',
  price: 7999,
  createdAt: new Date()
});

console.log(product.name); // "IPHONE 15 PRO"
console.log(product.price); // "¥7999.00"
```

---

## 📋 参数装饰器：参数的智能处理

### 参数验证装饰器

```typescript
// 参数验证装饰器
function ValidateParam(validator: (value: any) => boolean, message: string) {
  return function(target: any, propertyKey: string, parameterIndex: number) {
    const existingValidators = Reflect.getMetadata('paramValidators', target, propertyKey) || {};
    existingValidators[parameterIndex] = { validator, message };
    Reflect.defineMetadata('paramValidators', existingValidators, target, propertyKey);
  };
}

// 方法装饰器：执行参数验证
function ValidateParams(target: any, propertyKey: string, descriptor: PropertyDescriptor) {
  const originalMethod = descriptor.value;
  
  descriptor.value = function(...args: any[]) {
    const validators = Reflect.getMetadata('paramValidators', target, propertyKey) || {};
    
    for (const [index, { validator, message }] of Object.entries(validators)) {
      const paramIndex = parseInt(index);
      const paramValue = args[paramIndex];
      
      if (!validator(paramValue)) {
        throw new Error(`参数 ${paramIndex} 验证失败: ${message}`);
      }
    }
    
    return originalMethod.apply(this, args);
  };
}

// 常用验证器
const validators = {
  isPositive: (value: number) => value > 0,
  isEmail: (value: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value),
  isNotEmpty: (value: string) => value.trim().length > 0,
  isValidId: (value: number) => Number.isInteger(value) && value > 0
};

// Mall-Frontend中的订单服务
class OrderService {
  @ValidateParams
  async createOrder(
    @ValidateParam(validators.isValidId, '用户ID必须是正整数')
    userId: number,
    
    @ValidateParam(validators.isPositive, '总金额必须大于0')
    totalAmount: number,
    
    @ValidateParam(validators.isNotEmpty, '收货地址不能为空')
    shippingAddress: string
  ) {
    console.log('创建订单:', { userId, totalAmount, shippingAddress });
    
    // 实际的订单创建逻辑
    return {
      id: Date.now(),
      userId,
      totalAmount,
      shippingAddress,
      status: 'pending'
    };
  }
}

// 使用示例
const orderService = new OrderService();

try {
  await orderService.createOrder(1, 99.99, '北京市朝阳区xxx街道');
  console.log('订单创建成功');
} catch (error) {
  console.error(error.message);
}

try {
  await orderService.createOrder(-1, 99.99, ''); // 参数验证失败
} catch (error) {
  console.error(error.message); // "参数 0 验证失败: 用户ID必须是正整数"
}
```

---

## 🔍 元数据编程：Reflect Metadata

### 安装和配置

```bash
npm install reflect-metadata
```

```typescript
// 在应用入口导入
import 'reflect-metadata';

// 或在tsconfig.json中配置
{
  "compilerOptions": {
    "experimentalDecorators": true,
    "emitDecoratorMetadata": true
  }
}
```

### 依赖注入系统

```typescript
import 'reflect-metadata';

// 依赖注入容器
class Container {
  private services = new Map<string, any>();
  private singletons = new Map<string, any>();
  
  register<T>(token: string, implementation: new (...args: any[]) => T, singleton = false) {
    this.services.set(token, { implementation, singleton });
  }
  
  resolve<T>(token: string): T {
    const service = this.services.get(token);
    if (!service) {
      throw new Error(`服务 ${token} 未注册`);
    }
    
    if (service.singleton) {
      if (!this.singletons.has(token)) {
        const instance = this.createInstance(service.implementation);
        this.singletons.set(token, instance);
      }
      return this.singletons.get(token);
    }
    
    return this.createInstance(service.implementation);
  }
  
  private createInstance<T>(constructor: new (...args: any[]) => T): T {
    // 获取构造函数的参数类型
    const paramTypes = Reflect.getMetadata('design:paramtypes', constructor) || [];
    
    // 解析依赖
    const dependencies = paramTypes.map((type: any) => {
      const token = Reflect.getMetadata('inject:token', type) || type.name;
      return this.resolve(token);
    });
    
    return new constructor(...dependencies);
  }
}

// 装饰器
function Injectable(token?: string) {
  return function<T extends { new(...args: any[]): {} }>(constructor: T) {
    const actualToken = token || constructor.name;
    Reflect.defineMetadata('inject:token', actualToken, constructor);
    return constructor;
  };
}

function Inject(token: string) {
  return function(target: any, propertyKey: string | symbol | undefined, parameterIndex: number) {
    const existingTokens = Reflect.getMetadata('inject:tokens', target) || [];
    existingTokens[parameterIndex] = token;
    Reflect.defineMetadata('inject:tokens', existingTokens, target);
  };
}

// Mall-Frontend中的服务
@Injectable('ApiClient')
class ApiClient {
  constructor(private baseURL: string = 'http://localhost:8080') {}
  
  async get(endpoint: string) {
    return fetch(`${this.baseURL}${endpoint}`);
  }
}

@Injectable('ProductService')
class ProductService {
  constructor(private apiClient: ApiClient) {}
  
  async getProducts() {
    return this.apiClient.get('/api/products');
  }
}

@Injectable('CartService')
class CartService {
  constructor(
    private apiClient: ApiClient,
    private productService: ProductService
  ) {}
  
  async addToCart(productId: number, quantity: number) {
    const product = await this.productService.getProducts();
    return this.apiClient.get(`/api/cart/add/${productId}/${quantity}`);
  }
}

// 使用容器
const container = new Container();

// 注册服务
container.register('ApiClient', ApiClient, true); // 单例
container.register('ProductService', ProductService);
container.register('CartService', CartService);

// 解析服务
const cartService = container.resolve<CartService>('CartService');
cartService.addToCart(1, 2);
```

---

## 🎯 面试常考知识点

### 1. 装饰器的执行时机

**Q: 装饰器什么时候执行？执行顺序是怎样的？**

**A: 装饰器执行时机和顺序：**

1. **执行时机**: 装饰器在类定义时执行，不是在实例化时
2. **执行顺序**: 
   - 属性装饰器 → 参数装饰器 → 方法装饰器 → 类装饰器
   - 多个同类型装饰器：从下到上执行

```typescript
function first() {
  console.log('first(): factory evaluated');
  return function (target: any, propertyKey: string, descriptor: PropertyDescriptor) {
    console.log('first(): called');
  };
}

function second() {
  console.log('second(): factory evaluated');
  return function (target: any, propertyKey: string, descriptor: PropertyDescriptor) {
    console.log('second(): called');
  };
}

class Example {
  @first()
  @second()
  method() {}
}

// 输出：
// first(): factory evaluated
// second(): factory evaluated
// second(): called
// first(): called
```

### 2. 装饰器的类型和参数

**Q: 不同类型的装饰器接收什么参数？**

**A: 装饰器参数详解：**

```typescript
// 类装饰器
function ClassDecorator(constructor: Function) {
  // constructor: 类的构造函数
}

// 方法装饰器
function MethodDecorator(
  target: any,                    // 类的原型对象
  propertyKey: string,            // 方法名
  descriptor: PropertyDescriptor  // 属性描述符
) {}

// 属性装饰器
function PropertyDecorator(
  target: any,        // 类的原型对象
  propertyKey: string // 属性名
) {}

// 参数装饰器
function ParameterDecorator(
  target: any,        // 类的原型对象
  propertyKey: string, // 方法名
  parameterIndex: number // 参数索引
) {}
```

### 3. 装饰器的实际应用场景

**Q: 装饰器在实际项目中有哪些应用场景？**

**A: 常见应用场景：**

1. **依赖注入** - Angular、NestJS
2. **数据验证** - class-validator
3. **ORM映射** - TypeORM
4. **缓存控制** - 方法级缓存
5. **日志记录** - 自动日志
6. **权限控制** - 路由守卫
7. **性能监控** - 方法执行时间

---

## 🏋️ 实战练习

### 练习1: 实现一个完整的验证系统

**题目**: 为Mall-Frontend创建一个基于装饰器的表单验证系统

**要求**:
1. 支持多种验证规则（必填、长度、格式等）
2. 支持自定义验证器
3. 提供友好的错误信息
4. 支持异步验证

**解决方案**:

```typescript
// 验证装饰器系统
import 'reflect-metadata';

// 验证规则接口
interface ValidationRule {
  validator: (value: any) => boolean | Promise<boolean>;
  message: string;
  async?: boolean;
}

// 验证装饰器
function Validate(rule: ValidationRule) {
  return function(target: any, propertyKey: string) {
    const rules = Reflect.getMetadata('validation:rules', target) || {};
    if (!rules[propertyKey]) {
      rules[propertyKey] = [];
    }
    rules[propertyKey].push(rule);
    Reflect.defineMetadata('validation:rules', rules, target);
  };
}

// 预定义验证规则
const ValidationRules = {
  required: (message = '此字段为必填项') => ({
    validator: (value: any) => value !== null && value !== undefined && value !== '',
    message
  }),
  
  minLength: (length: number, message?: string) => ({
    validator: (value: string) => !value || value.length >= length,
    message: message || `最少需要${length}个字符`
  }),
  
  maxLength: (length: number, message?: string) => ({
    validator: (value: string) => !value || value.length <= length,
    message: message || `最多允许${length}个字符`
  }),
  
  email: (message = '邮箱格式不正确') => ({
    validator: (value: string) => !value || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value),
    message
  }),
  
  phone: (message = '手机号格式不正确') => ({
    validator: (value: string) => !value || /^1[3-9]\d{9}$/.test(value),
    message
  }),
  
  uniqueUsername: (message = '用户名已存在') => ({
    validator: async (value: string) => {
      if (!value) return true;
      // 模拟异步验证
      const response = await fetch(`/api/users/check-username?username=${value}`);
      const result = await response.json();
      return result.available;
    },
    message,
    async: true
  })
};

// 验证器类
class Validator {
  static async validate(obj: any): Promise<{ [key: string]: string[] }> {
    const rules = Reflect.getMetadata('validation:rules', obj) || {};
    const errors: { [key: string]: string[] } = {};
    
    for (const [property, propertyRules] of Object.entries(rules)) {
      const value = obj[property];
      const propertyErrors: string[] = [];
      
      for (const rule of propertyRules as ValidationRule[]) {
        try {
          const isValid = rule.async 
            ? await rule.validator(value)
            : rule.validator(value);
            
          if (!isValid) {
            propertyErrors.push(rule.message);
          }
        } catch (error) {
          propertyErrors.push(`验证过程中发生错误: ${error.message}`);
        }
      }
      
      if (propertyErrors.length > 0) {
        errors[property] = propertyErrors;
      }
    }
    
    return errors;
  }
}

// Mall-Frontend用户注册表单
class UserRegistrationForm {
  @Validate(ValidationRules.required())
  @Validate(ValidationRules.minLength(3))
  @Validate(ValidationRules.maxLength(20))
  @Validate(ValidationRules.uniqueUsername())
  username: string = '';
  
  @Validate(ValidationRules.required())
  @Validate(ValidationRules.email())
  email: string = '';
  
  @Validate(ValidationRules.required())
  @Validate(ValidationRules.minLength(6))
  password: string = '';
  
  @Validate(ValidationRules.phone())
  phone: string = '';
  
  constructor(data: Partial<UserRegistrationForm>) {
    Object.assign(this, data);
  }
  
  async validate() {
    return Validator.validate(this);
  }
}

// 使用示例
async function testValidation() {
  const form = new UserRegistrationForm({
    username: 'jo',
    email: 'invalid-email',
    password: '123',
    phone: '123456'
  });
  
  const errors = await form.validate();
  console.log('验证错误:', errors);
  
  // 输出:
  // {
  //   username: ['最少需要3个字符', '用户名已存在'],
  //   email: ['邮箱格式不正确'],
  //   password: ['最少需要6个字符'],
  //   phone: ['手机号格式不正确']
  // }
}
```

这个练习展示了：

1. **元数据存储** - 使用Reflect.metadata存储验证规则
2. **装饰器组合** - 多个验证装饰器可以组合使用
3. **异步验证** - 支持异步验证器（如用户名唯一性检查）
4. **错误收集** - 完整的错误信息收集和展示
5. **实际应用** - 在真实的用户注册场景中使用

---

## 📚 本章总结

通过本章学习，我们深入掌握了TypeScript装饰器和元数据编程：

### 🎯 核心收获

1. **装饰器基础** 🎭
   - 理解了装饰器的工作原理和执行时机
   - 掌握了四种装饰器类型的使用方法
   - 学会了装饰器工厂的设计模式

2. **实际应用** 🛠️
   - 掌握了缓存、重试、性能监控等实用装饰器
   - 学会了验证系统的装饰器实现
   - 理解了依赖注入的装饰器应用

3. **元数据编程** 🔍
   - 掌握了Reflect Metadata的使用
   - 学会了设计类型安全的依赖注入系统
   - 理解了元数据在框架中的应用

4. **最佳实践** 💡
   - 学会了装饰器的组合使用
   - 掌握了错误处理和异步装饰器
   - 理解了装饰器在企业级应用中的价值

### 🚀 技术进阶

- **下一步学习**: 模块系统与命名空间
- **实践建议**: 在项目中应用装饰器简化重复代码
- **深入方向**: 框架级装饰器设计和元编程

装饰器为我们提供了优雅的代码增强方式，让TypeScript代码更加简洁和强大！ 🎉

---

*下一章我们将学习《模块系统与命名空间》，探索TypeScript的代码组织和模块化开发！* 🚀
