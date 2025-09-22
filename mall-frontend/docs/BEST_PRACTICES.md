# Mall-Frontend 最佳实践指南

## 📋 概述

本文档提供了使用Mall-Frontend安全和性能优化功能的最佳实践，帮助开发团队构建高质量、可维护的前端应用。

---

## 🔐 安全最佳实践

### 1. 错误处理策略

#### ✅ 推荐做法

```typescript
// 统一错误处理策略
class ApiService {
  private errorHandler = ErrorHandler.getInstance();

  async request<T>(url: string, options?: RequestInit): Promise<T> {
    try {
      const response = await fetch(url, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...options?.headers,
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      // 统一错误处理
      const errorInfo = this.errorHandler.handleError(error, {
        type: this.getErrorType(error),
        context: {
          url,
          method: options?.method || 'GET',
          timestamp: Date.now(),
        },
      });

      // 根据错误类型决定是否重试
      if (this.shouldRetry(errorInfo)) {
        return this.retryRequest(url, options);
      }

      throw error;
    }
  }

  private getErrorType(error: any): ErrorType {
    if (error.name === 'TypeError' && error.message.includes('fetch')) {
      return ErrorType.NETWORK;
    }
    if (error.message.includes('401') || error.message.includes('403')) {
      return ErrorType.AUTHENTICATION;
    }
    if (error.message.includes('400')) {
      return ErrorType.VALIDATION;
    }
    return ErrorType.UNKNOWN;
  }
}
```

#### ❌ 避免的做法

```typescript
// 不要在每个组件中重复错误处理逻辑
function BadComponent() {
  const [data, setData] = useState(null);

  useEffect(() => {
    fetch('/api/data')
      .then(response => {
        if (!response.ok) {
          console.error('Request failed'); // 不统一的错误处理
          return;
        }
        return response.json();
      })
      .then(setData)
      .catch(error => {
        console.error(error); // 简单的console.error不够
        alert('Something went wrong'); // 用户体验差
      });
  }, []);

  return <div>{data ? JSON.stringify(data) : 'Loading...'}</div>;
}
```

### 2. 配置管理最佳实践

#### ✅ 推荐的配置结构

```typescript
// config/index.ts
export const configSchema = {
  // API配置
  api: {
    baseUrl: {
      type: 'string',
      required: true,
      envVar: 'NEXT_PUBLIC_API_BASE_URL',
      validator: (value: string) => value.startsWith('http'),
      description: 'API服务器基础URL',
    },
    timeout: {
      type: 'number',
      default: 10000,
      envVar: 'NEXT_PUBLIC_API_TIMEOUT',
      validator: (value: number) => value > 0 && value <= 60000,
      description: '请求超时时间(毫秒)',
    },
  },

  // 功能开关
  features: {
    enableAnalytics: {
      type: 'boolean',
      default: false,
      envVar: 'NEXT_PUBLIC_ENABLE_ANALYTICS',
      description: '是否启用分析功能',
    },
    enableExperimentalFeatures: {
      type: 'boolean',
      default: false,
      envVar: 'NEXT_PUBLIC_ENABLE_EXPERIMENTAL',
      description: '是否启用实验性功能',
    },
  },

  // 用户偏好设置
  userPreferences: {
    theme: {
      type: 'string',
      default: 'light',
      persistent: true,
      validator: (value: string) => ['light', 'dark', 'auto'].includes(value),
      description: '主题模式',
    },
    language: {
      type: 'string',
      default: 'zh-CN',
      persistent: true,
      envVar: 'NEXT_PUBLIC_DEFAULT_LOCALE',
      description: '界面语言',
    },
  },
};

// 配置初始化
export async function initializeConfig() {
  const configManager = ConfigManager.getInstance();

  // 批量注册配置
  Object.entries(configSchema).forEach(([groupName, configs]) => {
    configManager.registerGroup(groupName);
    configManager.registerConfig(groupName, configs);
  });

  await configManager.initialize();

  // 验证关键配置
  const criticalGroups = ['api'];
  for (const group of criticalGroups) {
    if (!configManager.validateConfig(group)) {
      throw new Error(`关键配置组 ${group} 验证失败`);
    }
  }

  return configManager;
}
```

#### ✅ 环境特定配置

```typescript
// config/environments.ts
export const environmentConfigs = {
  development: {
    api: {
      baseUrl: 'http://localhost:8080',
      timeout: 30000, // 开发环境允许更长的超时
    },
    features: {
      enableAnalytics: false,
      enableExperimentalFeatures: true,
    },
  },

  staging: {
    api: {
      baseUrl: 'https://staging-api.example.com',
      timeout: 15000,
    },
    features: {
      enableAnalytics: true,
      enableExperimentalFeatures: true,
    },
  },

  production: {
    api: {
      baseUrl: 'https://api.example.com',
      timeout: 10000,
    },
    features: {
      enableAnalytics: true,
      enableExperimentalFeatures: false,
    },
  },
};

// 根据环境加载配置
export function loadEnvironmentConfig() {
  const env = process.env.NODE_ENV || 'development';
  return environmentConfigs[env] || environmentConfigs.development;
}
```

### 3. 资源管理最佳实践

#### ✅ 组件级资源管理

```typescript
// hooks/useComponentResources.ts
export function useComponentResources(componentName: string) {
  const resourceManager = ResourceManager.getInstance();
  const groupId = useMemo(() =>
    resourceManager.createGroup(componentName, true), [componentName]
  );

  const registerResource = useCallback((
    type: 'timer' | 'interval' | 'listener' | 'observer',
    ...args: any[]
  ) => {
    let resourceId: string;

    switch (type) {
      case 'timer':
        resourceId = resourceManager.registerTimer(...args);
        break;
      case 'interval':
        resourceId = resourceManager.registerInterval(...args);
        break;
      case 'listener':
        resourceId = resourceManager.registerEventListener(...args);
        break;
      case 'observer':
        resourceId = resourceManager.registerObserver(...args);
        break;
      default:
        throw new Error(`不支持的资源类型: ${type}`);
    }

    resourceManager.addToGroup(groupId, resourceId);
    return resourceId;
  }, [groupId]);

  // 组件卸载时自动清理
  useEffect(() => {
    return () => {
      resourceManager.cleanupGroup(groupId);
    };
  }, [groupId]);

  return { registerResource, groupId };
}

// 在组件中使用
function OptimizedComponent() {
  const { registerResource } = useComponentResources('OptimizedComponent');

  useEffect(() => {
    // 注册定时器
    registerResource('timer', () => {
      console.log('定时任务');
    }, 1000, '组件定时任务');

    // 注册事件监听
    registerResource('listener', window, 'resize', handleResize, { passive: true }, '窗口大小变化');

    // 注册观察器
    const observer = new IntersectionObserver(handleIntersection);
    registerResource('observer', observer, '元素可见性观察');
  }, [registerResource]);

  return <div>优化的组件</div>;
}
```

#### ✅ 页面级资源管理

```typescript
// components/PageWrapper.tsx
export function PageWrapper({
  children,
  pageName
}: {
  children: ReactNode;
  pageName: string;
}) {
  const resourceManager = ResourceManager.getInstance();
  const [groupId] = useState(() =>
    resourceManager.createGroup(`page-${pageName}`, true)
  );

  // 页面性能监控
  useEffect(() => {
    const startTime = performance.now();

    // 注册页面卸载监听
    const handleBeforeUnload = () => {
      const endTime = performance.now();
      const duration = endTime - startTime;

      // 记录页面停留时间
      analytics.track('page_duration', {
        page: pageName,
        duration: Math.round(duration)
      });
    };

    const listenerId = resourceManager.registerEventListener(
      window,
      'beforeunload',
      handleBeforeUnload,
      undefined,
      `${pageName}页面卸载监听`
    );

    resourceManager.addToGroup(groupId, listenerId);

    return () => {
      handleBeforeUnload();
      resourceManager.cleanupGroup(groupId);
    };
  }, [pageName, groupId]);

  return (
    <div className={`page page--${pageName}`}>
      {children}
    </div>
  );
}
```

---

## 🚀 性能优化最佳实践

### 1. 图片优化策略

#### ✅ 响应式图片加载

```typescript
// components/SmartImage.tsx
interface SmartImageProps {
  src: string;
  alt: string;
  priority?: boolean;
  sizes?: string;
  className?: string;
}

export function SmartImage({
  src,
  alt,
  priority = false,
  sizes,
  className
}: SmartImageProps) {
  const [devicePixelRatio, setDevicePixelRatio] = useState(1);

  useEffect(() => {
    setDevicePixelRatio(window.devicePixelRatio || 1);
  }, []);

  // 根据设备像素比调整质量
  const getQuality = () => {
    if (devicePixelRatio >= 2) return 85; // 高分辨率设备
    return 75; // 普通设备
  };

  // 根据sizes属性确定图片尺寸
  const getImageDimensions = () => {
    if (!sizes) return {};

    // 解析sizes属性，例如: "(max-width: 768px) 100vw, 50vw"
    const viewportWidth = window.innerWidth;
    if (viewportWidth <= 768) {
      return { width: viewportWidth };
    }
    return { width: Math.round(viewportWidth * 0.5) };
  };

  const config = {
    ...getImageDimensions(),
    quality: getQuality(),
    format: 'auto' as const,
    lazy: !priority,
    placeholder: '/images/placeholder.svg'
  };

  return (
    <OptimizedImage
      src={src}
      alt={alt}
      config={config}
      className={className}
      loading={priority ? 'eager' : 'lazy'}
    />
  );
}
```

#### ✅ 图片预加载策略

```typescript
// hooks/useImagePreloadStrategy.ts
export function useImagePreloadStrategy() {
  const imageOptimizer = ImageOptimizer.getInstance();

  // 关键图片立即预加载
  const preloadCriticalImages = useCallback(async (urls: string[]) => {
    await imageOptimizer.preload(urls, {
      priority: 'high',
      batchSize: 3,
      delay: 0
    });
  }, []);

  // 非关键图片延迟预加载
  const preloadSecondaryImages = useCallback(async (urls: string[]) => {
    // 等待主要内容加载完成
    await new Promise(resolve => setTimeout(resolve, 1000));

    await imageOptimizer.preload(urls, {
      priority: 'low',
      batchSize: 2,
      delay: 500
    });
  }, []);

  // 基于用户行为的智能预加载
  const preloadOnHover = useCallback((urls: string[]) => {
    return imageOptimizer.preload(urls, {
      priority: 'high',
      batchSize: 1,
      delay: 0
    });
  }, []);

  return {
    preloadCriticalImages,
    preloadSecondaryImages,
    preloadOnHover
  };
}

// 在页面组件中使用
function ProductListPage() {
  const { preloadCriticalImages, preloadOnHover } = useImagePreloadStrategy();
  const [products, setProducts] = useState([]);

  useEffect(() => {
    // 预加载首屏产品图片
    if (products.length > 0) {
      const firstScreenImages = products
        .slice(0, 6)
        .map(product => product.image);

      preloadCriticalImages(firstScreenImages);
    }
  }, [products, preloadCriticalImages]);

  const handleProductHover = (product: Product) => {
    // 用户悬停时预加载相关图片
    const relatedImages = [
      product.detailImage,
      ...product.galleryImages?.slice(0, 2) || []
    ];

    preloadOnHover(relatedImages);
  };

  return (
    <div className="product-list">
      {products.map(product => (
        <ProductCard
          key={product.id}
          product={product}
          onHover={() => handleProductHover(product)}
        />
      ))}
    </div>
  );
}
```

### 2. 组件性能优化

#### ✅ 智能组件记忆化

```typescript
// components/OptimizedProductCard.tsx
interface ProductCardProps {
  product: Product;
  onAddToCart: (productId: number) => void;
  onViewDetail: (productId: number) => void;
  isInCart?: boolean;
  discount?: number;
}

// 自定义比较函数
const arePropsEqual = (
  prevProps: ProductCardProps,
  nextProps: ProductCardProps
): boolean => {
  // 只比较影响渲染的关键属性
  return (
    prevProps.product.id === nextProps.product.id &&
    prevProps.product.name === nextProps.product.name &&
    prevProps.product.price === nextProps.product.price &&
    prevProps.product.image === nextProps.product.image &&
    prevProps.isInCart === nextProps.isInCart &&
    prevProps.discount === nextProps.discount
    // 注意：不比较函数引用，因为它们可能每次都不同
  );
};

const ProductCard = React.memo<ProductCardProps>(({
  product,
  onAddToCart,
  onViewDetail,
  isInCart = false,
  discount = 0
}) => {
  // 使用useCallback稳定化事件处理器
  const handleAddToCart = useCallback(() => {
    onAddToCart(product.id);
  }, [onAddToCart, product.id]);

  const handleViewDetail = useCallback(() => {
    onViewDetail(product.id);
  }, [onViewDetail, product.id]);

  // 计算派生状态
  const finalPrice = useMemo(() => {
    return discount > 0 ? product.price * (1 - discount) : product.price;
  }, [product.price, discount]);

  return (
    <div className="product-card">
      <SmartImage
        src={product.image}
        alt={product.name}
        className="product-image"
      />

      <div className="product-info">
        <h3>{product.name}</h3>
        <div className="price">
          {discount > 0 && (
            <span className="original-price">¥{product.price}</span>
          )}
          <span className="final-price">¥{finalPrice.toFixed(2)}</span>
        </div>

        <div className="actions">
          <button
            onClick={handleViewDetail}
            className="btn btn-secondary"
          >
            查看详情
          </button>
          <button
            onClick={handleAddToCart}
            className={`btn btn-primary ${isInCart ? 'in-cart' : ''}`}
            disabled={isInCart}
          >
            {isInCart ? '已在购物车' : '加入购物车'}
          </button>
        </div>
      </div>
    </div>
  );
}, arePropsEqual);

export default ProductCard;
```

#### ✅ 虚拟滚动优化

```typescript
// components/VirtualizedProductList.tsx
import { FixedSizeList as List } from 'react-window';

interface VirtualizedProductListProps {
  products: Product[];
  onAddToCart: (productId: number) => void;
  onViewDetail: (productId: number) => void;
}

const ITEM_HEIGHT = 300;
const ITEMS_PER_ROW = 3;

export function VirtualizedProductList({
  products,
  onAddToCart,
  onViewDetail
}: VirtualizedProductListProps) {
  const [containerHeight, setContainerHeight] = useState(600);

  // 计算行数
  const rowCount = Math.ceil(products.length / ITEMS_PER_ROW);

  // 渲染单行
  const Row = useCallback(({ index, style }: { index: number; style: any }) => {
    const startIndex = index * ITEMS_PER_ROW;
    const endIndex = Math.min(startIndex + ITEMS_PER_ROW, products.length);
    const rowProducts = products.slice(startIndex, endIndex);

    return (
      <div style={style} className="product-row">
        {rowProducts.map(product => (
          <ProductCard
            key={product.id}
            product={product}
            onAddToCart={onAddToCart}
            onViewDetail={onViewDetail}
          />
        ))}
      </div>
    );
  }, [products, onAddToCart, onViewDetail]);

  // 响应式容器高度
  useEffect(() => {
    const updateHeight = () => {
      const viewportHeight = window.innerHeight;
      const headerHeight = 80;
      const footerHeight = 60;
      setContainerHeight(viewportHeight - headerHeight - footerHeight);
    };

    updateHeight();
    window.addEventListener('resize', updateHeight);
    return () => window.removeEventListener('resize', updateHeight);
  }, []);

  return (
    <div className="virtualized-product-list">
      <List
        height={containerHeight}
        itemCount={rowCount}
        itemSize={ITEM_HEIGHT}
        width="100%"
        overscanCount={2} // 预渲染2行
      >
        {Row}
      </List>
    </div>
  );
}
```

---

## 🏗️ 架构最佳实践

### 1. 依赖注入架构

#### ✅ 分层服务架构

```typescript
// 数据访问层
interface IRepository<T> {
  findById(id: number): Promise<T | null>;
  findAll(): Promise<T[]>;
  create(entity: Omit<T, 'id'>): Promise<T>;
  update(id: number, entity: Partial<T>): Promise<T>;
  delete(id: number): Promise<void>;
}

// 业务逻辑层
interface IService<T> {
  get(id: number): Promise<T>;
  list(filters?: any): Promise<T[]>;
  create(data: any): Promise<T>;
  update(id: number, data: any): Promise<T>;
  delete(id: number): Promise<void>;
}

// 表现层
interface IController<T> {
  handleGet(id: number): Promise<ApiResponse<T>>;
  handleList(query: any): Promise<ApiResponse<T[]>>;
  handleCreate(data: any): Promise<ApiResponse<T>>;
  handleUpdate(id: number, data: any): Promise<ApiResponse<T>>;
  handleDelete(id: number): Promise<ApiResponse<void>>;
}

// 具体实现
export class ProductService implements IService<Product> {
  constructor(
    private repository: IRepository<Product>,
    private cacheService: ICacheService,
    private eventBus: IEventBus
  ) {}

  async get(id: number): Promise<Product> {
    // 先检查缓存
    const cached = await this.cacheService.get(`product:${id}`);
    if (cached) return cached;

    // 从数据库获取
    const product = await this.repository.findById(id);
    if (!product) {
      throw new Error(`Product ${id} not found`);
    }

    // 缓存结果
    await this.cacheService.set(`product:${id}`, product, 300);

    // 发布事件
    this.eventBus.emit('product.viewed', { productId: id });

    return product;
  }

  // 其他方法...
}
```

### 2. 错误处理架构

#### ✅ 分层错误处理

```typescript
// 错误类型定义
export class AppError extends Error {
  constructor(
    message: string,
    public code: string,
    public statusCode: number = 500,
    public isOperational: boolean = true
  ) {
    super(message);
    this.name = this.constructor.name;
    Error.captureStackTrace(this, this.constructor);
  }
}

export class ValidationError extends AppError {
  constructor(
    message: string,
    public field?: string
  ) {
    super(message, 'VALIDATION_ERROR', 400);
  }
}

export class NotFoundError extends AppError {
  constructor(resource: string, id?: string | number) {
    super(
      `${resource}${id ? ` with id ${id}` : ''} not found`,
      'NOT_FOUND',
      404
    );
  }
}

// 错误处理中间件
export class ErrorHandlingService {
  private errorHandler = ErrorHandler.getInstance();

  handleError(error: Error, context?: any): never {
    // 记录错误
    this.errorHandler.handleError(error, {
      context,
      level: this.getErrorLevel(error),
      type: this.getErrorType(error),
    });

    // 根据错误类型决定处理方式
    if (error instanceof AppError) {
      throw error; // 业务错误直接抛出
    }

    // 系统错误转换为通用错误
    throw new AppError('Internal server error', 'INTERNAL_ERROR', 500, false);
  }

  private getErrorLevel(error: Error): ErrorLevel {
    if (error instanceof ValidationError) return ErrorLevel.WARNING;
    if (error instanceof NotFoundError) return ErrorLevel.INFO;
    if (error instanceof AppError && error.isOperational)
      return ErrorLevel.ERROR;
    return ErrorLevel.CRITICAL;
  }

  private getErrorType(error: Error): ErrorType {
    if (error instanceof ValidationError) return ErrorType.VALIDATION;
    if (error instanceof NotFoundError) return ErrorType.NOT_FOUND;
    if (error.message.includes('network')) return ErrorType.NETWORK;
    return ErrorType.UNKNOWN;
  }
}
```

---

## 📊 监控和调试最佳实践

### 1. 性能监控

```typescript
// utils/performanceMonitor.ts
export class PerformanceMonitor {
  private static instance: PerformanceMonitor;
  private metrics: Map<string, number[]> = new Map();

  static getInstance(): PerformanceMonitor {
    if (!this.instance) {
      this.instance = new PerformanceMonitor();
    }
    return this.instance;
  }

  // 测量函数执行时间
  async measureAsync<T>(name: string, fn: () => Promise<T>): Promise<T> {
    const start = performance.now();
    try {
      const result = await fn();
      this.recordMetric(name, performance.now() - start);
      return result;
    } catch (error) {
      this.recordMetric(`${name}_error`, performance.now() - start);
      throw error;
    }
  }

  // 记录指标
  recordMetric(name: string, value: number): void {
    if (!this.metrics.has(name)) {
      this.metrics.set(name, []);
    }

    const values = this.metrics.get(name)!;
    values.push(value);

    // 保持最近100个值
    if (values.length > 100) {
      values.shift();
    }
  }

  // 获取统计信息
  getStats(name: string) {
    const values = this.metrics.get(name) || [];
    if (values.length === 0) return null;

    const sorted = [...values].sort((a, b) => a - b);
    return {
      count: values.length,
      min: sorted[0],
      max: sorted[sorted.length - 1],
      avg: values.reduce((a, b) => a + b, 0) / values.length,
      p50: sorted[Math.floor(sorted.length * 0.5)],
      p95: sorted[Math.floor(sorted.length * 0.95)],
      p99: sorted[Math.floor(sorted.length * 0.99)],
    };
  }

  // 导出所有指标
  exportMetrics() {
    const result: Record<string, any> = {};
    for (const [name] of this.metrics) {
      result[name] = this.getStats(name);
    }
    return result;
  }
}

// 使用示例
const monitor = PerformanceMonitor.getInstance();

// 在服务中使用
export class ProductService {
  async getProducts(): Promise<Product[]> {
    return monitor.measureAsync('product_service_get_products', async () => {
      const products = await this.repository.findAll();
      return products;
    });
  }
}
```

### 2. 调试工具

```typescript
// utils/debugTools.ts
export class DebugTools {
  private static enabled = process.env.NODE_ENV === 'development';

  static log(category: string, message: string, data?: any): void {
    if (!this.enabled) return;

    const timestamp = new Date().toISOString();
    const style = this.getCategoryStyle(category);

    console.group(`%c[${category}] ${timestamp}`, style);
    console.log(message);
    if (data) {
      console.log('Data:', data);
    }
    console.groupEnd();
  }

  static performance(name: string, fn: () => any): any {
    if (!this.enabled) return fn();

    const start = performance.now();
    const result = fn();
    const duration = performance.now() - start;

    this.log('PERFORMANCE', `${name} took ${duration.toFixed(2)}ms`);
    return result;
  }

  static async performanceAsync<T>(
    name: string,
    fn: () => Promise<T>
  ): Promise<T> {
    if (!this.enabled) return fn();

    const start = performance.now();
    const result = await fn();
    const duration = performance.now() - start;

    this.log('PERFORMANCE', `${name} took ${duration.toFixed(2)}ms`);
    return result;
  }

  private static getCategoryStyle(category: string): string {
    const styles = {
      ERROR: 'color: #ff4444; font-weight: bold;',
      WARNING: 'color: #ffaa00; font-weight: bold;',
      INFO: 'color: #4444ff; font-weight: bold;',
      SUCCESS: 'color: #44ff44; font-weight: bold;',
      PERFORMANCE: 'color: #ff44ff; font-weight: bold;',
      DEBUG: 'color: #888888;',
    };
    return styles[category] || styles.DEBUG;
  }
}

// 在开发环境中启用全局调试
if (process.env.NODE_ENV === 'development') {
  (window as any).debugTools = DebugTools;
  (window as any).performanceMonitor = PerformanceMonitor.getInstance();
}
```

---

## 🔍 代码审查检查清单

### 安全检查

- [ ] 所有用户输入都经过验证和清理
- [ ] 敏感信息不存储在localStorage中
- [ ] 错误信息不暴露系统内部信息
- [ ] API调用包含适当的错误处理
- [ ] 配置信息通过环境变量管理

### 性能检查

- [ ] 组件使用React.memo适当优化
- [ ] 图片使用OptimizedImage组件
- [ ] 长列表使用虚拟滚动
- [ ] 资源在组件卸载时正确清理
- [ ] 避免不必要的重新渲染

### 代码质量检查

- [ ] 函数和组件职责单一
- [ ] 使用TypeScript类型定义
- [ ] 遵循命名约定
- [ ] 包含适当的注释和文档
- [ ] 测试覆盖关键功能

这些最佳实践涵盖了安全、性能、架构和监控等各个方面，帮助开发团队构建高质量的前端应用。
