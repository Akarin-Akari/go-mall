# Mall-Frontend API 文档

## 📋 概述

本文档详细介绍了Mall-Frontend项目中新增的安全和性能优化API的使用方法、配置选项和最佳实践。

---

## 🔐 安全管理 API

### ErrorHandler - 统一错误处理

#### 基本用法

```typescript
import { ErrorHandler } from '@/utils/errorHandler';

// 获取单例实例
const errorHandler = ErrorHandler.getInstance();

// 处理错误
const errorInfo = errorHandler.handleError(new Error('Something went wrong'), {
  type: ErrorType.NETWORK,
  level: ErrorLevel.ERROR,
  context: { userId: '123', action: 'fetchProducts' },
});
```

#### API 方法

##### `handleError(error, options?)`

处理错误并返回标准化的错误信息。

**参数:**

- `error: Error | string` - 错误对象或错误消息
- `options?: ErrorOptions` - 可选的错误处理选项

**返回值:**

```typescript
interface ErrorInfo {
  id: string;
  message: string;
  type: ErrorType;
  level: ErrorLevel;
  timestamp: number;
  stack?: string;
  context?: Record<string, any>;
  code?: string;
}
```

##### `addListener(listener)`

添加错误监听器。

**参数:**

- `listener: (error: ErrorInfo) => void` - 错误监听回调函数

**返回值:**

- `() => void` - 移除监听器的函数

##### `getErrorStats()`

获取错误统计信息。

**返回值:**

```typescript
interface ErrorStats {
  totalErrors: number;
  errorsByType: Record<ErrorType, number>;
  errorsByLevel: Record<ErrorLevel, number>;
  recentErrors: ErrorInfo[];
}
```

#### 配置选项

```typescript
interface ErrorHandlerConfig {
  enableConsoleLogging: boolean;
  enableRemoteReporting: boolean;
  remoteEndpoint?: string;
  maxQueueSize: number;
  batchSize: number;
  flushInterval: number;
}

// 更新配置
await errorHandler.updateConfig({
  enableRemoteReporting: true,
  remoteEndpoint: 'https://api.example.com/errors',
});
```

#### 使用示例

```typescript
// 基础错误处理
try {
  await fetchUserData();
} catch (error) {
  errorHandler.handleError(error, {
    context: { component: 'UserProfile', action: 'fetchData' },
  });
}

// 添加错误监听器
const removeListener = errorHandler.addListener(error => {
  if (error.level === ErrorLevel.CRITICAL) {
    // 发送紧急通知
    notifyAdmin(error);
  }
});

// 重试机制
const result = await errorHandler.withRetry(
  async () => {
    return await apiCall();
  },
  { maxRetries: 3, delay: 1000 }
);
```

---

## ⚙️ 配置管理 API

### ConfigManager - 统一配置管理

#### 基本用法

```typescript
import { ConfigManager } from '@/utils/configManager';

const configManager = ConfigManager.getInstance();

// 注册配置组
configManager.registerGroup('api', 'API相关配置');

// 注册配置项
configManager.registerConfig('api', {
  baseUrl: {
    value: 'http://localhost:8080',
    description: 'API基础URL',
    type: 'string',
    required: true,
    envVar: 'NEXT_PUBLIC_API_BASE_URL',
  },
  timeout: {
    value: 10000,
    description: '请求超时时间',
    type: 'number',
    validator: (value: number) => value > 0 && value < 60000,
  },
});
```

#### API 方法

##### `registerGroup(groupName, description?)`

注册配置组。

**参数:**

- `groupName: string` - 配置组名称
- `description?: string` - 配置组描述

##### `registerConfig(groupName, configs)`

注册配置项。

**参数:**

- `groupName: string` - 配置组名称
- `configs: Record<string, ConfigDefinition>` - 配置定义

```typescript
interface ConfigDefinition {
  value: any;
  description?: string;
  type: 'string' | 'number' | 'boolean' | 'object';
  required?: boolean;
  persistent?: boolean;
  envVar?: string;
  validator?: (value: any) => boolean;
}
```

##### `get(groupName, key)`

获取配置值。

**参数:**

- `groupName: string` - 配置组名称
- `key: string` - 配置键名

**返回值:**

- `any` - 配置值

##### `set(groupName, key, value)`

设置配置值。

**参数:**

- `groupName: string` - 配置组名称
- `key: string` - 配置键名
- `value: any` - 配置值

**返回值:**

- `boolean` - 设置是否成功

#### 配置监听

```typescript
// 添加配置变更监听器
const removeListener = configManager.addListener(
  'baseUrl',
  (key, newValue, oldValue) => {
    console.log(`配置 ${key} 从 ${oldValue} 变更为 ${newValue}`);
  }
);

// 移除监听器
removeListener();
```

#### 使用示例

```typescript
// 初始化配置
await configManager.initialize();

// 获取API配置
const apiConfig = {
  baseUrl: configManager.get('api', 'baseUrl'),
  timeout: configManager.get('api', 'timeout'),
};

// 动态更新配置
configManager.set('api', 'timeout', 15000);

// 验证配置
const isValid = configManager.validateConfig('api');
if (!isValid) {
  console.error('API配置验证失败');
}

// 重置配置
configManager.resetToDefaults('api');
```

---

## 🧹 资源管理 API

### ResourceManager - 内存泄漏防护

#### 基本用法

```typescript
import { ResourceManager } from '@/utils/resourceManager';

const resourceManager = ResourceManager.getInstance();

// 注册定时器
const timerId = resourceManager.registerTimer(
  () => {
    console.log('定时任务执行');
  },
  1000,
  '心跳检测'
);

// 注册事件监听器
const listenerId = resourceManager.registerEventListener(
  element,
  'click',
  handleClick,
  { passive: true },
  '按钮点击监听'
);

// 注册观察器
const observerId = resourceManager.registerObserver(
  new MutationObserver(callback),
  'DOM变化监听'
);
```

#### API 方法

##### `registerTimer(callback, delay, description?)`

注册定时器。

**参数:**

- `callback: () => void` - 回调函数
- `delay: number` - 延迟时间（毫秒）
- `description?: string` - 描述信息

**返回值:**

- `string` - 资源ID

##### `registerInterval(callback, interval, description?)`

注册间隔定时器。

**参数:**

- `callback: () => void` - 回调函数
- `interval: number` - 间隔时间（毫秒）
- `description?: string` - 描述信息

**返回值:**

- `string` - 资源ID

##### `registerEventListener(element, event, listener, options?, description?)`

注册事件监听器。

**参数:**

- `element: EventTarget` - 目标元素
- `event: string` - 事件类型
- `listener: EventListener` - 监听器函数
- `options?: AddEventListenerOptions` - 监听选项
- `description?: string` - 描述信息

**返回值:**

- `string` - 资源ID

##### `cleanup(resourceId)`

清理指定资源。

**参数:**

- `resourceId: string` - 资源ID

##### `cleanupAll()`

清理所有资源。

#### 资源组管理

```typescript
// 创建资源组
const groupId = resourceManager.createGroup('userProfile', true);

// 添加资源到组
resourceManager.addToGroup(groupId, timerId);
resourceManager.addToGroup(groupId, listenerId);

// 清理整个组
resourceManager.cleanupGroup(groupId);
```

#### 使用示例

```typescript
// React Hook 中的使用
function useResourceCleanup() {
  const resourceManager = ResourceManager.getInstance();
  const resourceIds = useRef<string[]>([]);

  const addResource = useCallback((id: string) => {
    resourceIds.current.push(id);
  }, []);

  useEffect(() => {
    return () => {
      // 组件卸载时清理所有资源
      resourceIds.current.forEach(id => {
        resourceManager.cleanup(id);
      });
    };
  }, []);

  return { addResource };
}

// 在组件中使用
function MyComponent() {
  const { addResource } = useResourceCleanup();

  useEffect(() => {
    const timerId = resourceManager.registerTimer(() => {
      // 定时任务
    }, 1000);

    addResource(timerId);
  }, []);

  return <div>My Component</div>;
}
```

---

## 🖼️ 图片优化 API

### ImageOptimizer - 图片性能优化

#### 基本用法

```typescript
import { ImageOptimizer, OptimizedImage } from '@/utils/imageOptimizer';

// 使用优化图片组件
<OptimizedImage
  src="/images/product.jpg"
  alt="商品图片"
  config={{
    width: 300,
    height: 200,
    quality: 85,
    format: 'webp',
    lazy: true,
    placeholder: '/images/placeholder.jpg'
  }}
  onLoad={() => console.log('图片加载完成')}
  onError={(error) => console.error('图片加载失败', error)}
/>
```

#### API 方法

##### `generateOptimizedUrl(src, config?)`

生成优化的图片URL。

**参数:**

- `src: string` - 原始图片URL
- `config?: ImageConfig` - 优化配置

```typescript
interface ImageConfig {
  width?: number;
  height?: number;
  quality?: number;
  format?: 'auto' | 'webp' | 'avif' | 'jpeg' | 'png';
  lazy?: boolean;
  placeholder?: string;
  fallback?: string;
  retryCount?: number;
  retryDelay?: number;
}
```

**返回值:**

- `string` - 优化后的图片URL

##### `loadImage(src, config?)`

预加载图片。

**参数:**

- `src: string` - 图片URL
- `config?: ImageConfig` - 加载配置

**返回值:**

- `Promise<HTMLImageElement>` - 图片元素

##### `preload(urls, options?)`

批量预加载图片。

**参数:**

- `urls: string[]` - 图片URL数组
- `options?: PreloadOptions` - 预加载选项

```typescript
interface PreloadOptions {
  priority?: 'high' | 'low';
  batchSize?: number;
  delay?: number;
}
```

#### 性能监控

```typescript
// 获取优化统计
const stats = imageOptimizer.getOptimizationStats();
console.log('图片优化统计:', stats);

// 重置统计
imageOptimizer.resetStats();
```

#### 使用示例

```typescript
// 手动优化图片URL
const optimizedUrl = imageOptimizer.generateOptimizedUrl('/images/hero.jpg', {
  width: 1200,
  height: 600,
  quality: 90,
  format: 'webp',
});

// 预加载关键图片
await imageOptimizer.preload(
  ['/images/hero.jpg', '/images/featured-1.jpg', '/images/featured-2.jpg'],
  {
    priority: 'high',
    batchSize: 2,
  }
);

// 使用React Hook
function useImagePreload(urls: string[]) {
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    imageOptimizer.preload(urls).then(() => {
      setLoaded(true);
    });
  }, [urls]);

  return loaded;
}
```

---

## 🔧 依赖注入 API

### ServiceContainer - 依赖注入容器

#### 基本用法

```typescript
import { ServiceContainer, useService } from '@/container';

// 注册服务
const container = new ServiceContainer();

container.register('apiClient', () => new ApiClient(), {
  lifecycle: 'singleton',
  tags: ['api', 'http'],
});

container.register(
  'userService',
  container => {
    const apiClient = container.resolve('apiClient');
    return new UserService(apiClient);
  },
  {
    lifecycle: 'transient',
    dependencies: ['apiClient'],
  }
);
```

#### React Hook 集成

```typescript
// 在组件中使用服务
function UserProfile() {
  const userService = useService<UserService>('userService');
  const [user, setUser] = useState(null);

  useEffect(() => {
    userService.getCurrentUser().then(setUser);
  }, [userService]);

  return <div>{user?.name}</div>;
}

// 使用服务状态
function ServiceStatus() {
  const status = useServiceStatus('userService');

  if (status === 'loading') {
    return <div>服务加载中...</div>;
  }

  if (status === 'error') {
    return <div>服务加载失败</div>;
  }

  return <div>服务就绪</div>;
}
```

#### 模块系统

```typescript
// 定义模块
export class CoreModule implements IModule {
  configure(builder: IApplicationBuilder): void {
    builder.services.register('logger', () => new Logger());
    builder.services.register('config', () => new ConfigService());
  }
}

// 应用启动
const app = new ApplicationBuilder()
  .addModule(new CoreModule())
  .addModule(new SecurityModule())
  .build();

await app.start();
```

---

## 📊 性能监控

### 性能指标收集

```typescript
// 收集性能指标
const performanceMetrics = {
  errorStats: errorHandler.getErrorStats(),
  resourceStats: resourceManager.getResourceStats(),
  imageStats: imageOptimizer.getOptimizationStats(),
  configStats: configManager.getAllConfigs(),
};

// 发送到监控服务
await fetch('/api/metrics', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(performanceMetrics),
});
```

### 健康检查

```typescript
// 系统健康检查
const healthChecks = await Promise.all([
  errorHandler.healthCheck(),
  configManager.healthCheck(),
  resourceManager.healthCheck(),
  imageOptimizer.healthCheck(),
]);

const isHealthy = healthChecks.every(check => check === true);
console.log('系统健康状态:', isHealthy ? '正常' : '异常');
```

---

## 🚀 最佳实践

### 1. 错误处理最佳实践

```typescript
// 统一错误边界
class ErrorBoundary extends React.Component {
  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    errorHandler.handleError(error, {
      context: {
        component: errorInfo.componentStack,
        errorBoundary: true,
      },
    });
  }
}

// API错误处理
const apiClient = axios.create({
  baseURL: configManager.get('api', 'baseUrl'),
});

apiClient.interceptors.response.use(
  response => response,
  error => {
    errorHandler.handleError(error, {
      type: ErrorType.NETWORK,
      context: { url: error.config?.url },
    });
    return Promise.reject(error);
  }
);
```

### 2. 资源管理最佳实践

```typescript
// 自定义Hook封装
function useTimer(callback: () => void, delay: number) {
  const resourceManager = ResourceManager.getInstance();

  useEffect(() => {
    const timerId = resourceManager.registerTimer(callback, delay);
    return () => resourceManager.cleanup(timerId);
  }, [callback, delay]);
}

// 页面级资源组管理
function PageComponent() {
  const resourceManager = ResourceManager.getInstance();
  const groupId = useMemo(
    () => resourceManager.createGroup('page-component', true),
    []
  );

  useEffect(() => {
    return () => resourceManager.cleanupGroup(groupId);
  }, [groupId]);

  // 页面逻辑...
}
```

### 3. 配置管理最佳实践

```typescript
// 环境特定配置
const envConfig = {
  development: {
    apiUrl: 'http://localhost:8080',
    debug: true,
  },
  production: {
    apiUrl: 'https://api.production.com',
    debug: false,
  },
};

// 配置验证
const configSchema = {
  apiUrl: { type: 'string', required: true },
  timeout: { type: 'number', min: 1000, max: 30000 },
};

configManager.registerConfig('api', envConfig[process.env.NODE_ENV]);
```

---

## 🔍 故障排除

### 常见问题

1. **错误处理器初始化失败**
   - 检查是否正确调用了 `initialize()` 方法
   - 确认没有循环依赖问题

2. **配置值获取为 undefined**
   - 确认配置组和键名是否正确注册
   - 检查环境变量是否正确设置

3. **资源清理不生效**
   - 确认资源ID是否正确保存
   - 检查组件卸载时是否调用了清理函数

4. **图片优化不工作**
   - 检查图片URL格式是否正确
   - 确认浏览器是否支持WebP格式

### 调试工具

```typescript
// 开启调试模式
if (process.env.NODE_ENV === 'development') {
  // 错误处理调试
  errorHandler.addListener(console.error);

  // 资源管理调试
  setInterval(() => {
    console.log('资源统计:', resourceManager.getResourceStats());
  }, 10000);

  // 配置管理调试
  configManager.addListener('*', (key, newValue, oldValue) => {
    console.log(`配置变更: ${key}`, { oldValue, newValue });
  });
}
```
