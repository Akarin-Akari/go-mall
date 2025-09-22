# Mall-Frontend 故障排除指南

## 📋 概述

本文档提供了Mall-Frontend项目中常见问题的诊断和解决方案，帮助开发者快速定位和修复问题。

---

## 🔧 常见问题诊断

### 1. 应用启动问题

#### 问题：应用无法启动或白屏

**症状：**

- 浏览器显示白屏
- 控制台出现JavaScript错误
- 应用卡在加载状态

**诊断步骤：**

```bash
# 1. 检查Node.js版本
node --version  # 应该 >= 18.0.0

# 2. 检查依赖安装
npm list --depth=0

# 3. 清理缓存重新安装
rm -rf node_modules package-lock.json
npm install

# 4. 检查环境变量
cat .env.local
```

**常见解决方案：**

```typescript
// 检查应用初始化
// app/layout.tsx
export default async function RootLayout({ children }) {
  try {
    // 确保服务正确初始化
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
  } catch (error) {
    console.error('应用初始化失败:', error);

    // 提供降级UI
    return (
      <html lang="zh-CN">
        <body>
          <div className="error-fallback">
            <h1>应用启动失败</h1>
            <p>请刷新页面重试</p>
            <button onClick={() => window.location.reload()}>
              刷新页面
            </button>
          </div>
        </body>
      </html>
    );
  }
}
```

#### 问题：模块导入错误

**症状：**

- `Cannot resolve module` 错误
- `Module not found` 错误
- TypeScript类型错误

**解决方案：**

```typescript
// 检查tsconfig.json路径映射
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"],
      "@/components/*": ["./src/components/*"],
      "@/utils/*": ["./src/utils/*"],
      "@/hooks/*": ["./src/hooks/*"]
    }
  }
}

// 检查next.config.ts
const nextConfig = {
  webpack: (config) => {
    config.resolve.alias = {
      ...config.resolve.alias,
      '@': path.resolve(__dirname, 'src'),
    };
    return config;
  },
};
```

### 2. 错误处理问题

#### 问题：ErrorHandler初始化失败

**症状：**

- `ErrorHandler.getInstance() is not a function`
- 错误处理不工作
- 错误信息丢失

**诊断代码：**

```typescript
// utils/errorHandler.ts 检查
export class ErrorHandler {
  private static instance: ErrorHandler | null = null;

  static getInstance(): ErrorHandler {
    if (!this.instance) {
      this.instance = new ErrorHandler();
    }
    return this.instance;
  }

  // 添加调试信息
  constructor() {
    console.log('ErrorHandler initialized');
    this.initialize();
  }
}

// 在应用中测试
const errorHandler = ErrorHandler.getInstance();
console.log('ErrorHandler status:', errorHandler.getStatus());
```

**解决方案：**

```typescript
// 确保正确的初始化顺序
// app/bootstrap.ts
export async function bootstrapApplication() {
  try {
    // 1. 首先初始化错误处理器
    const errorHandler = ErrorHandler.getInstance();
    await errorHandler.initialize();

    // 2. 然后初始化其他服务
    const configManager = ConfigManager.getInstance();
    await configManager.initialize();

    // 3. 验证初始化状态
    const healthChecks = await Promise.all([
      errorHandler.healthCheck(),
      configManager.healthCheck(),
    ]);

    if (!healthChecks.every(check => check === true)) {
      throw new Error('服务健康检查失败');
    }

    return { errorHandler, configManager };
  } catch (error) {
    console.error('应用启动失败:', error);
    throw error;
  }
}
```

#### 问题：错误信息不显示

**症状：**

- 错误发生但用户看不到提示
- 控制台有错误但UI无响应
- 错误边界不工作

**解决方案：**

```typescript
// 检查错误边界实现
class ErrorBoundary extends Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, errorInfo: null };
  }

  static getDerivedStateFromError(error) {
    console.log('ErrorBoundary caught error:', error);
    return { hasError: true };
  }

  componentDidCatch(error, errorInfo) {
    console.log('ErrorBoundary componentDidCatch:', { error, errorInfo });

    // 确保错误处理器存在
    try {
      const errorHandler = ErrorHandler.getInstance();
      errorHandler.handleError(error, {
        context: { errorBoundary: true, ...errorInfo }
      });
    } catch (handlerError) {
      console.error('ErrorHandler not available:', handlerError);
    }
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="error-boundary">
          <h2>出现了意外错误</h2>
          <details>
            <summary>错误详情</summary>
            <pre>{this.state.errorInfo}</pre>
          </details>
          <button onClick={() => window.location.reload()}>
            刷新页面
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}
```

### 3. 配置管理问题

#### 问题：配置值为undefined

**症状：**

- `configManager.get()` 返回 undefined
- 环境变量未加载
- 配置验证失败

**诊断步骤：**

```typescript
// 调试配置管理器
const configManager = ConfigManager.getInstance();

// 1. 检查配置组是否注册
console.log('已注册的配置组:', configManager.getAllConfigs());

// 2. 检查特定配置
console.log('API配置:', configManager.get('api', 'baseUrl'));

// 3. 检查环境变量
console.log('环境变量:', {
  NODE_ENV: process.env.NODE_ENV,
  NEXT_PUBLIC_API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL,
});

// 4. 验证配置
const isValid = configManager.validateConfig('api');
console.log('API配置验证:', isValid);
```

**解决方案：**

```typescript
// 确保配置正确注册
export async function initializeConfig() {
  const configManager = ConfigManager.getInstance();

  // 注册配置前先检查
  if (!configManager.hasGroup('api')) {
    configManager.registerGroup('api', 'API相关配置');
  }

  configManager.registerConfig('api', {
    baseUrl: {
      value: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080',
      description: 'API基础URL',
      type: 'string',
      required: true,
      envVar: 'NEXT_PUBLIC_API_BASE_URL',
    },
  });

  await configManager.initialize();

  // 验证关键配置
  const apiUrl = configManager.get('api', 'baseUrl');
  if (!apiUrl) {
    throw new Error('API基础URL配置缺失');
  }

  return configManager;
}
```

### 4. 资源管理问题

#### 问题：内存泄漏

**症状：**

- 页面切换后内存持续增长
- 定时器未清理
- 事件监听器累积

**诊断工具：**

```typescript
// 内存泄漏检测工具
export class MemoryLeakDetector {
  private static timers = new Set<number>();
  private static intervals = new Set<number>();
  private static listeners = new Map<
    EventTarget,
    Map<string, EventListener[]>
  >();

  static wrapSetTimeout(originalSetTimeout: typeof setTimeout) {
    return function (callback: Function, delay: number, ...args: any[]) {
      const id = originalSetTimeout(() => {
        MemoryLeakDetector.timers.delete(id);
        callback(...args);
      }, delay);

      MemoryLeakDetector.timers.add(id);
      return id;
    };
  }

  static wrapSetInterval(originalSetInterval: typeof setInterval) {
    return function (callback: Function, delay: number, ...args: any[]) {
      const id = originalSetInterval(callback, delay, ...args);
      MemoryLeakDetector.intervals.add(id);
      return id;
    };
  }

  static getLeakReport() {
    return {
      activeTimers: this.timers.size,
      activeIntervals: this.intervals.size,
      activeListeners: Array.from(this.listeners.entries()).map(
        ([target, events]) => ({
          target: target.constructor.name,
          events: Array.from(events.keys()),
        })
      ),
    };
  }
}

// 在开发环境中启用
if (process.env.NODE_ENV === 'development') {
  window.setTimeout = MemoryLeakDetector.wrapSetTimeout(window.setTimeout);
  window.setInterval = MemoryLeakDetector.wrapSetInterval(window.setInterval);

  // 定期报告
  setInterval(() => {
    console.log('内存泄漏检测:', MemoryLeakDetector.getLeakReport());
  }, 10000);
}
```

**解决方案：**

```typescript
// 使用ResourceManager确保清理
function useProperResourceManagement() {
  const resourceManager = ResourceManager.getInstance();
  const resourceIds = useRef<string[]>([]);

  const addResource = useCallback((id: string) => {
    resourceIds.current.push(id);
  }, []);

  // 组件卸载时清理
  useEffect(() => {
    return () => {
      resourceIds.current.forEach(id => {
        try {
          resourceManager.cleanup(id);
        } catch (error) {
          console.warn('资源清理失败:', id, error);
        }
      });
      resourceIds.current = [];
    };
  }, []);

  return { addResource };
}
```

### 5. 图片优化问题

#### 问题：图片加载失败

**症状：**

- 图片显示为破损图标
- 加载时间过长
- 格式不支持

**诊断代码：**

```typescript
// 图片加载诊断工具
export class ImageDiagnostics {
  static async testImageLoad(src: string): Promise<{
    success: boolean;
    loadTime: number;
    error?: string;
    dimensions?: { width: number; height: number };
  }> {
    const startTime = performance.now();

    return new Promise(resolve => {
      const img = new Image();

      img.onload = () => {
        resolve({
          success: true,
          loadTime: performance.now() - startTime,
          dimensions: { width: img.naturalWidth, height: img.naturalHeight },
        });
      };

      img.onerror = error => {
        resolve({
          success: false,
          loadTime: performance.now() - startTime,
          error: error.toString(),
        });
      };

      img.src = src;
    });
  }

  static async testFormatSupport(): Promise<{
    webp: boolean;
    avif: boolean;
    jpeg: boolean;
    png: boolean;
  }> {
    const canvas = document.createElement('canvas');
    canvas.width = 1;
    canvas.height = 1;

    return {
      webp: canvas.toDataURL('image/webp').indexOf('data:image/webp') === 0,
      avif: canvas.toDataURL('image/avif').indexOf('data:image/avif') === 0,
      jpeg: true,
      png: true,
    };
  }
}

// 使用诊断工具
const diagnostics = await ImageDiagnostics.testImageLoad('/images/test.jpg');
console.log('图片加载诊断:', diagnostics);

const formatSupport = await ImageDiagnostics.testFormatSupport();
console.log('格式支持:', formatSupport);
```

**解决方案：**

```typescript
// 增强的图片组件错误处理
function RobustOptimizedImage({ src, alt, ...props }) {
  const [currentSrc, setCurrentSrc] = useState(src);
  const [retryCount, setRetryCount] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const handleError = useCallback((error: Error) => {
    console.error('图片加载失败:', { src: currentSrc, error });

    if (retryCount < 3) {
      // 重试机制
      setTimeout(() => {
        setRetryCount(prev => prev + 1);
        setCurrentSrc(`${src}?retry=${retryCount + 1}`);
      }, 1000 * Math.pow(2, retryCount)); // 指数退避
    } else {
      // 使用fallback图片
      setError(error.message);
      setCurrentSrc('/images/fallback.svg');
    }
  }, [src, retryCount]);

  const handleLoad = useCallback(() => {
    setError(null);
    setRetryCount(0);
  }, []);

  if (error && retryCount >= 3) {
    return (
      <div className="image-error">
        <img src="/images/error-placeholder.svg" alt={alt} />
        <span className="error-message">图片加载失败</span>
      </div>
    );
  }

  return (
    <OptimizedImage
      src={currentSrc}
      alt={alt}
      onLoad={handleLoad}
      onError={handleError}
      {...props}
    />
  );
}
```

---

## 🔍 调试工具和技巧

### 1. 开发者工具集成

```typescript
// 开发环境调试工具
if (process.env.NODE_ENV === 'development') {
  // 全局调试对象
  (window as any).__DEBUG__ = {
    errorHandler: ErrorHandler.getInstance(),
    configManager: ConfigManager.getInstance(),
    resourceManager: ResourceManager.getInstance(),
    imageOptimizer: ImageOptimizer.getInstance(),

    // 快速诊断函数
    async diagnose() {
      const results = {
        errorHandler: await this.errorHandler.healthCheck(),
        configManager: await this.configManager.healthCheck(),
        resourceManager: await this.resourceManager.healthCheck(),
        imageOptimizer: await this.imageOptimizer.healthCheck(),

        stats: {
          errors: this.errorHandler.getErrorStats(),
          resources: this.resourceManager.getResourceStats(),
          images: this.imageOptimizer.getOptimizationStats(),
        },
      };

      console.table(results);
      return results;
    },

    // 清理所有资源
    cleanup() {
      this.resourceManager.cleanupAll();
      this.imageOptimizer.clearCache();
      this.errorHandler.resetErrorCount();
    },
  };

  console.log('🔧 调试工具已加载，使用 window.__DEBUG__ 访问');
}
```

### 2. 性能分析

```typescript
// 性能分析工具
export class PerformanceProfiler {
  private marks = new Map<string, number>();
  private measures = new Map<string, number>();

  mark(name: string): void {
    this.marks.set(name, performance.now());
    performance.mark(name);
  }

  measure(name: string, startMark: string, endMark?: string): number {
    if (endMark) {
      performance.measure(name, startMark, endMark);
    } else {
      this.mark(`${name}_end`);
      performance.measure(name, startMark, `${name}_end`);
    }

    const measure = performance.getEntriesByName(name, 'measure')[0];
    const duration = measure.duration;
    this.measures.set(name, duration);

    return duration;
  }

  getReport(): Record<string, number> {
    return Object.fromEntries(this.measures);
  }

  clear(): void {
    this.marks.clear();
    this.measures.clear();
    performance.clearMarks();
    performance.clearMeasures();
  }
}

// 使用示例
const profiler = new PerformanceProfiler();

// 在组件中使用
function ProfiledComponent() {
  useEffect(() => {
    profiler.mark('component_mount_start');

    return () => {
      profiler.measure('component_mount_duration', 'component_mount_start');
      console.log('组件性能报告:', profiler.getReport());
    };
  }, []);

  return <div>Profiled Component</div>;
}
```

---

## 📞 获取帮助

### 1. 日志收集

```bash
# 收集系统信息
npm run diagnose

# 导出错误日志
npm run export-logs

# 生成性能报告
npm run performance-report
```

### 2. 问题报告模板

```markdown
## 问题描述

[简要描述遇到的问题]

## 复现步骤

1. [第一步]
2. [第二步]
3. [第三步]

## 预期行为

[描述期望的正确行为]

## 实际行为

[描述实际发生的行为]

## 环境信息

- Node.js版本: [版本号]
- npm版本: [版本号]
- 浏览器: [浏览器及版本]
- 操作系统: [操作系统及版本]

## 错误信息
```

[粘贴完整的错误信息和堆栈跟踪]

```

## 相关日志
```

[粘贴相关的控制台日志]

```

## 已尝试的解决方案
[列出已经尝试过的解决方法]
```

### 3. 联系方式

- 📧 技术支持: support@example.com
- 📖 文档: [项目文档链接]
- 🐛 问题报告: [GitHub Issues链接]
- 💬 社区讨论: [讨论区链接]
