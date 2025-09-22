# Mall-Frontend 使用示例

## 📋 概述

本文档提供了Mall-Frontend项目中安全和性能优化功能的实际使用示例，帮助开发者快速上手和集成这些功能。

---

## 🔐 安全功能使用示例

### 1. 统一错误处理

#### 在React组件中使用

```typescript
// components/UserProfile.tsx
import React, { useState, useEffect } from 'react';
import { ErrorHandler, ErrorType } from '@/utils/errorHandler';
import { useErrorHandler } from '@/hooks/useErrorHandler';

function UserProfile() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const { handleError, errors } = useErrorHandler();

  useEffect(() => {
    fetchUserProfile();
  }, []);

  const fetchUserProfile = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/user/profile');

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const userData = await response.json();
      setUser(userData);
    } catch (error) {
      handleError(error, {
        type: ErrorType.NETWORK,
        context: {
          component: 'UserProfile',
          action: 'fetchUserProfile',
          userId: user?.id
        }
      });
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <div>加载中...</div>;

  if (errors.length > 0) {
    return (
      <div className="error-container">
        <h3>出现错误</h3>
        {errors.map(error => (
          <div key={error.id} className="error-message">
            {error.message}
          </div>
        ))}
        <button onClick={fetchUserProfile}>重试</button>
      </div>
    );
  }

  return (
    <div className="user-profile">
      <h2>{user?.name}</h2>
      <p>{user?.email}</p>
    </div>
  );
}
```

#### 全局错误边界

```typescript
// components/ErrorBoundary.tsx
import React, { Component, ReactNode } from 'react';
import { ErrorHandler, ErrorType, ErrorLevel } from '@/utils/errorHandler';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  errorId?: string;
}

class ErrorBoundary extends Component<Props, State> {
  private errorHandler = ErrorHandler.getInstance();

  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    const errorResult = this.errorHandler.handleError(error, {
      type: ErrorType.RUNTIME,
      level: ErrorLevel.CRITICAL,
      context: {
        componentStack: errorInfo.componentStack,
        errorBoundary: true,
        timestamp: Date.now()
      }
    });

    this.setState({ errorId: errorResult.id });

    // 发送错误报告到监控服务
    this.reportError(errorResult);
  }

  private reportError = async (errorInfo: any) => {
    try {
      await fetch('/api/errors/report', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(errorInfo)
      });
    } catch (reportError) {
      console.error('Failed to report error:', reportError);
    }
  };

  render() {
    if (this.state.hasError) {
      return this.props.fallback || (
        <div className="error-boundary">
          <h2>出现了意外错误</h2>
          <p>错误ID: {this.state.errorId}</p>
          <button onClick={() => window.location.reload()}>
            刷新页面
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}

// App.tsx 中使用
function App() {
  return (
    <ErrorBoundary>
      <Router>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/profile" element={<UserProfile />} />
        </Routes>
      </Router>
    </ErrorBoundary>
  );
}
```

### 2. 配置管理系统

#### 应用配置初始化

```typescript
// config/appConfig.ts
import { ConfigManager } from '@/utils/configManager';

export async function initializeAppConfig() {
  const configManager = ConfigManager.getInstance();

  // 注册API配置组
  configManager.registerGroup('api', 'API相关配置');
  configManager.registerConfig('api', {
    baseUrl: {
      value: 'http://localhost:8080',
      description: 'API基础URL',
      type: 'string',
      required: true,
      envVar: 'NEXT_PUBLIC_API_BASE_URL',
      validator: (value: string) => value.startsWith('http'),
    },
    timeout: {
      value: 10000,
      description: '请求超时时间(ms)',
      type: 'number',
      envVar: 'NEXT_PUBLIC_API_TIMEOUT',
      validator: (value: number) => value > 0 && value <= 60000,
    },
    retryCount: {
      value: 3,
      description: '重试次数',
      type: 'number',
      validator: (value: number) => value >= 0 && value <= 10,
    },
  });

  // 注册UI配置组
  configManager.registerGroup('ui', 'UI相关配置');
  configManager.registerConfig('ui', {
    theme: {
      value: 'light',
      description: '主题模式',
      type: 'string',
      persistent: true,
      validator: (value: string) => ['light', 'dark'].includes(value),
    },
    language: {
      value: 'zh-CN',
      description: '界面语言',
      type: 'string',
      persistent: true,
      envVar: 'NEXT_PUBLIC_DEFAULT_LOCALE',
    },
    pageSize: {
      value: 20,
      description: '分页大小',
      type: 'number',
      persistent: true,
      validator: (value: number) => value > 0 && value <= 100,
    },
  });

  // 初始化配置
  await configManager.initialize();

  // 验证配置
  const apiValid = configManager.validateConfig('api');
  const uiValid = configManager.validateConfig('ui');

  if (!apiValid || !uiValid) {
    throw new Error('配置验证失败');
  }

  return configManager;
}
```

#### 在组件中使用配置

```typescript
// components/ProductList.tsx
import React, { useState, useEffect } from 'react';
import { useConfig } from '@/hooks/useConfig';

function ProductList() {
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(false);

  // 使用配置Hook
  const apiConfig = useConfig('api');
  const uiConfig = useConfig('ui');

  const fetchProducts = async (page = 1) => {
    setLoading(true);
    try {
      const response = await fetch(
        `${apiConfig.baseUrl}/api/products?page=${page}&size=${uiConfig.pageSize}`,
        {
          timeout: apiConfig.timeout,
          headers: {
            'Accept-Language': uiConfig.language
          }
        }
      );

      const data = await response.json();
      setProducts(data.products);
    } catch (error) {
      console.error('获取商品列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProducts();
  }, [apiConfig.baseUrl, uiConfig.pageSize]);

  return (
    <div className={`product-list theme-${uiConfig.theme}`}>
      {loading ? (
        <div>加载中...</div>
      ) : (
        <div className="products">
          {products.map(product => (
            <ProductCard key={product.id} product={product} />
          ))}
        </div>
      )}
    </div>
  );
}
```

### 3. 资源管理和内存泄漏防护

#### 自定义Hook封装

```typescript
// hooks/useResourceManager.ts
import { useEffect, useRef, useCallback } from 'react';
import { ResourceManager } from '@/utils/resourceManager';

export function useResourceManager() {
  const resourceManager = ResourceManager.getInstance();
  const resourceIds = useRef<string[]>([]);

  const addResource = useCallback((id: string) => {
    resourceIds.current.push(id);
  }, []);

  const registerTimer = useCallback(
    (callback: () => void, delay: number, description?: string) => {
      const id = resourceManager.registerTimer(callback, delay, description);
      addResource(id);
      return id;
    },
    [addResource]
  );

  const registerInterval = useCallback(
    (callback: () => void, interval: number, description?: string) => {
      const id = resourceManager.registerInterval(
        callback,
        interval,
        description
      );
      addResource(id);
      return id;
    },
    [addResource]
  );

  const registerEventListener = useCallback(
    (
      element: EventTarget,
      event: string,
      listener: EventListener,
      options?: AddEventListenerOptions,
      description?: string
    ) => {
      const id = resourceManager.registerEventListener(
        element,
        event,
        listener,
        options,
        description
      );
      addResource(id);
      return id;
    },
    [addResource]
  );

  // 组件卸载时自动清理
  useEffect(() => {
    return () => {
      resourceIds.current.forEach(id => {
        resourceManager.cleanup(id);
      });
      resourceIds.current = [];
    };
  }, []);

  return {
    registerTimer,
    registerInterval,
    registerEventListener,
    cleanup: resourceManager.cleanup.bind(resourceManager),
    getStats: resourceManager.getResourceStats.bind(resourceManager),
  };
}
```

#### 在复杂组件中使用

```typescript
// components/Dashboard.tsx
import React, { useState, useEffect, useRef } from 'react';
import { useResourceManager } from '@/hooks/useResourceManager';

function Dashboard() {
  const [data, setData] = useState(null);
  const [notifications, setNotifications] = useState([]);
  const chartRef = useRef<HTMLCanvasElement>(null);

  const { registerTimer, registerInterval, registerEventListener } = useResourceManager();

  useEffect(() => {
    // 注册定时刷新数据
    registerInterval(() => {
      fetchDashboardData();
    }, 30000, '仪表板数据刷新');

    // 注册通知检查
    registerTimer(() => {
      checkNotifications();
    }, 5000, '通知检查');

    // 注册窗口大小变化监听
    registerEventListener(window, 'resize', handleResize, { passive: true }, '窗口大小变化');

    // 注册键盘快捷键
    registerEventListener(document, 'keydown', handleKeyDown, undefined, '键盘快捷键');

    // 初始化图表
    initializeChart();
  }, []);

  const fetchDashboardData = async () => {
    try {
      const response = await fetch('/api/dashboard');
      const newData = await response.json();
      setData(newData);
    } catch (error) {
      console.error('获取仪表板数据失败:', error);
    }
  };

  const checkNotifications = async () => {
    try {
      const response = await fetch('/api/notifications');
      const newNotifications = await response.json();
      setNotifications(prev => [...prev, ...newNotifications]);
    } catch (error) {
      console.error('检查通知失败:', error);
    }
  };

  const handleResize = () => {
    // 重新调整图表大小
    if (chartRef.current) {
      resizeChart(chartRef.current);
    }
  };

  const handleKeyDown = (event: KeyboardEvent) => {
    // 处理快捷键
    if (event.ctrlKey && event.key === 'r') {
      event.preventDefault();
      fetchDashboardData();
    }
  };

  const initializeChart = () => {
    // 初始化图表逻辑
    if (chartRef.current) {
      // 图表初始化代码
    }
  };

  return (
    <div className="dashboard">
      <div className="dashboard-header">
        <h1>仪表板</h1>
        <div className="notifications">
          {notifications.map(notification => (
            <div key={notification.id} className="notification">
              {notification.message}
            </div>
          ))}
        </div>
      </div>

      <div className="dashboard-content">
        <canvas ref={chartRef} className="chart" />
        {data && (
          <div className="data-summary">
            <div>总用户: {data.totalUsers}</div>
            <div>总订单: {data.totalOrders}</div>
            <div>总收入: ¥{data.totalRevenue}</div>
          </div>
        )}
      </div>
    </div>
  );
}
```

---

## 🖼️ 图片优化使用示例

### 1. 基础图片组件使用

```typescript
// components/ProductImage.tsx
import React from 'react';
import { OptimizedImage } from '@/utils/imageOptimizer';

interface ProductImageProps {
  product: {
    id: number;
    name: string;
    image: string;
  };
  size?: 'small' | 'medium' | 'large';
  priority?: boolean;
}

function ProductImage({ product, size = 'medium', priority = false }: ProductImageProps) {
  const sizeConfig = {
    small: { width: 150, height: 150 },
    medium: { width: 300, height: 300 },
    large: { width: 600, height: 600 }
  };

  const config = {
    ...sizeConfig[size],
    quality: priority ? 95 : 85,
    format: 'auto' as const,
    lazy: !priority,
    placeholder: '/images/product-placeholder.svg',
    fallback: '/images/product-error.svg',
    retryCount: 3,
    retryDelay: 1000
  };

  return (
    <OptimizedImage
      src={product.image}
      alt={product.name}
      config={config}
      className={`product-image product-image--${size}`}
      onLoad={() => console.log(`产品图片加载完成: ${product.name}`)}
      onError={(error) => console.error(`产品图片加载失败: ${product.name}`, error)}
    />
  );
}

// 使用示例
function ProductCard({ product }) {
  return (
    <div className="product-card">
      <ProductImage
        product={product}
        size="medium"
        priority={product.featured}
      />
      <h3>{product.name}</h3>
      <p>¥{product.price}</p>
    </div>
  );
}
```

### 2. 图片预加载策略

```typescript
// hooks/useImagePreload.ts
import { useState, useEffect } from 'react';
import { useImagePreload as useImagePreloadHook } from '@/utils/imageOptimizer';

export function useProductImagePreload(products: Product[]) {
  const imageUrls = products.map(product => product.image);

  const {
    loadedImages,
    failedImages,
    loading,
    progress,
    preloadImages
  } = useImagePreloadHook(imageUrls, {
    priority: 'high',
    batchSize: 5,
    delay: 200
  });

  useEffect(() => {
    // 自动开始预加载
    preloadImages();
  }, [preloadImages]);

  return {
    isLoading: loading,
    progress,
    loadedCount: loadedImages.size,
    failedCount: failedImages.size,
    totalCount: imageUrls.length,
    isComplete: !loading && loadedImages.size + failedImages.size === imageUrls.length
  };
}

// 在组件中使用
function ProductGallery({ products }: { products: Product[] }) {
  const {
    isLoading,
    progress,
    loadedCount,
    totalCount,
    isComplete
  } = useProductImagePreload(products);

  return (
    <div className="product-gallery">
      {isLoading && (
        <div className="preload-progress">
          <div className="progress-bar">
            <div
              className="progress-fill"
              style={{ width: `${progress}%` }}
            />
          </div>
          <span>预加载图片: {loadedCount}/{totalCount}</span>
        </div>
      )}

      <div className="products-grid">
        {products.map(product => (
          <ProductCard
            key={product.id}
            product={product}
            imageLoaded={isComplete}
          />
        ))}
      </div>
    </div>
  );
}
```

### 3. 响应式图片处理

```typescript
// components/ResponsiveImage.tsx
import React, { useState, useEffect } from 'react';
import { OptimizedImage } from '@/utils/imageOptimizer';

interface ResponsiveImageProps {
  src: string;
  alt: string;
  className?: string;
  breakpoints?: {
    mobile: number;
    tablet: number;
    desktop: number;
  };
}

function ResponsiveImage({
  src,
  alt,
  className,
  breakpoints = { mobile: 375, tablet: 768, desktop: 1200 }
}: ResponsiveImageProps) {
  const [screenWidth, setScreenWidth] = useState(0);

  useEffect(() => {
    const updateScreenWidth = () => {
      setScreenWidth(window.innerWidth);
    };

    updateScreenWidth();
    window.addEventListener('resize', updateScreenWidth);

    return () => {
      window.removeEventListener('resize', updateScreenWidth);
    };
  }, []);

  const getImageConfig = () => {
    if (screenWidth <= breakpoints.mobile) {
      return { width: breakpoints.mobile, quality: 75 };
    } else if (screenWidth <= breakpoints.tablet) {
      return { width: breakpoints.tablet, quality: 80 };
    } else {
      return { width: breakpoints.desktop, quality: 85 };
    }
  };

  const config = {
    ...getImageConfig(),
    format: 'auto' as const,
    lazy: true
  };

  return (
    <OptimizedImage
      src={src}
      alt={alt}
      config={config}
      className={className}
    />
  );
}

// 使用示例
function HeroBanner() {
  return (
    <div className="hero-banner">
      <ResponsiveImage
        src="/images/hero-banner.jpg"
        alt="商城首页横幅"
        className="hero-image"
        breakpoints={{
          mobile: 375,
          tablet: 768,
          desktop: 1920
        }}
      />
      <div className="hero-content">
        <h1>欢迎来到我们的商城</h1>
        <p>发现最新的产品和优惠</p>
      </div>
    </div>
  );
}
```

---

## 🔧 依赖注入使用示例

### 1. 服务定义和注册

```typescript
// services/UserService.ts
export interface IUserService {
  getCurrentUser(): Promise<User>;
  updateUser(user: Partial<User>): Promise<User>;
  logout(): Promise<void>;
}

export class UserService implements IUserService {
  constructor(
    private apiClient: IApiClient,
    private configManager: IConfigManager,
    private errorHandler: IErrorHandler
  ) {}

  async getCurrentUser(): Promise<User> {
    try {
      const response = await this.apiClient.get('/api/user/profile');
      return response.data;
    } catch (error) {
      this.errorHandler.handleError(error, {
        context: { service: 'UserService', method: 'getCurrentUser' },
      });
      throw error;
    }
  }

  async updateUser(user: Partial<User>): Promise<User> {
    try {
      const response = await this.apiClient.put('/api/user/profile', user);
      return response.data;
    } catch (error) {
      this.errorHandler.handleError(error, {
        context: { service: 'UserService', method: 'updateUser' },
      });
      throw error;
    }
  }

  async logout(): Promise<void> {
    try {
      await this.apiClient.post('/api/auth/logout');
      // 清理本地状态
    } catch (error) {
      this.errorHandler.handleError(error, {
        context: { service: 'UserService', method: 'logout' },
      });
      throw error;
    }
  }
}
```

### 2. 模块化服务注册

```typescript
// modules/UserModule.ts
import { IModule, IApplicationBuilder } from '@/container';
import { UserService, IUserService } from '@/services/UserService';
import { ApiClient, IApiClient } from '@/services/ApiClient';

export class UserModule implements IModule {
  configure(builder: IApplicationBuilder): void {
    // 注册API客户端
    builder.services.register<IApiClient>(
      'apiClient',
      container => {
        const configManager = container.resolve('configManager');
        const errorHandler = container.resolve('errorHandler');
        return new ApiClient(configManager, errorHandler);
      },
      {
        lifecycle: 'singleton',
        dependencies: ['configManager', 'errorHandler'],
      }
    );

    // 注册用户服务
    builder.services.register<IUserService>(
      'userService',
      container => {
        const apiClient = container.resolve<IApiClient>('apiClient');
        const configManager = container.resolve('configManager');
        const errorHandler = container.resolve('errorHandler');
        return new UserService(apiClient, configManager, errorHandler);
      },
      {
        lifecycle: 'singleton',
        dependencies: ['apiClient', 'configManager', 'errorHandler'],
      }
    );
  }
}

// app/bootstrap.ts
import { ApplicationBuilder } from '@/container';
import { CoreModule } from '@/modules/CoreModule';
import { UserModule } from '@/modules/UserModule';
import { SecurityModule } from '@/modules/SecurityModule';

export async function bootstrapApplication() {
  const app = new ApplicationBuilder()
    .addModule(new CoreModule())
    .addModule(new SecurityModule())
    .addModule(new UserModule())
    .build();

  // 启动应用
  await app.start();

  // 验证服务
  const healthChecks = await app.validateServices();
  if (!healthChecks.every(check => check.healthy)) {
    throw new Error('服务验证失败');
  }

  return app;
}
```

### 3. React组件中使用服务

```typescript
// components/UserProfile.tsx
import React, { useState, useEffect } from 'react';
import { useService, useServiceContainer } from '@/hooks/useServiceContainer';
import { IUserService } from '@/services/UserService';

function UserProfile() {
  const userService = useService<IUserService>('userService');
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);

  useEffect(() => {
    loadUser();
  }, [userService]);

  const loadUser = async () => {
    try {
      setLoading(true);
      const userData = await userService.getCurrentUser();
      setUser(userData);
    } catch (error) {
      console.error('加载用户信息失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async (updatedUser: Partial<User>) => {
    try {
      const savedUser = await userService.updateUser(updatedUser);
      setUser(savedUser);
      setEditing(false);
    } catch (error) {
      console.error('保存用户信息失败:', error);
    }
  };

  const handleLogout = async () => {
    try {
      await userService.logout();
      // 重定向到登录页
      window.location.href = '/login';
    } catch (error) {
      console.error('退出登录失败:', error);
    }
  };

  if (loading) {
    return <div>加载中...</div>;
  }

  if (!user) {
    return <div>用户信息加载失败</div>;
  }

  return (
    <div className="user-profile">
      <div className="profile-header">
        <h2>用户资料</h2>
        <button onClick={() => setEditing(!editing)}>
          {editing ? '取消' : '编辑'}
        </button>
        <button onClick={handleLogout}>退出登录</button>
      </div>

      {editing ? (
        <UserEditForm
          user={user}
          onSave={handleSave}
          onCancel={() => setEditing(false)}
        />
      ) : (
        <UserDisplayInfo user={user} />
      )}
    </div>
  );
}

// 服务状态监控组件
function ServiceStatus() {
  const container = useServiceContainer();
  const [services, setServices] = useState<any[]>([]);

  useEffect(() => {
    const updateServiceStatus = () => {
      const serviceList = container.getRegisteredServices();
      setServices(serviceList);
    };

    updateServiceStatus();
    const interval = setInterval(updateServiceStatus, 5000);

    return () => clearInterval(interval);
  }, [container]);

  return (
    <div className="service-status">
      <h3>服务状态</h3>
      {services.map(service => (
        <div key={service.name} className="service-item">
          <span>{service.name}</span>
          <span className={`status ${service.healthy ? 'healthy' : 'unhealthy'}`}>
            {service.healthy ? '正常' : '异常'}
          </span>
        </div>
      ))}
    </div>
  );
}
```

---

## 🚀 完整应用示例

### 应用入口点

```typescript
// app/layout.tsx
import React from 'react';
import { bootstrapApplication } from './bootstrap';
import { ServiceProvider } from '@/container/ServiceProvider';
import ErrorBoundary from '@/components/ErrorBoundary';

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  // 启动应用服务
  const app = await bootstrapApplication();

  return (
    <html lang="zh-CN">
      <body>
        <ServiceProvider app={app}>
          <ErrorBoundary>
            {children}
          </ErrorBoundary>
        </ServiceProvider>
      </body>
    </html>
  );
}
```

### 主页面组件

```typescript
// app/page.tsx
import React from 'react';
import { ProductGallery } from '@/components/ProductGallery';
import { UserProfile } from '@/components/UserProfile';
import { ServiceStatus } from '@/components/ServiceStatus';

export default function HomePage() {
  return (
    <div className="home-page">
      <header className="page-header">
        <h1>Mall Frontend</h1>
        <UserProfile />
      </header>

      <main className="page-content">
        <ProductGallery />
      </main>

      <aside className="page-sidebar">
        <ServiceStatus />
      </aside>
    </div>
  );
}
```

这些示例展示了如何在实际项目中使用Mall-Frontend的安全和性能优化功能。每个示例都包含了错误处理、资源管理、配置管理和依赖注入等核心功能的集成使用。
