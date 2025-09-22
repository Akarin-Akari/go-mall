/**
 * 浏览器兼容性配置
 */

export interface BrowserSupport {
  name: string;
  minVersion: string;
  supported: boolean;
  features: string[];
  polyfillsRequired: string[];
}

/**
 * 支持的浏览器列表
 */
export const SUPPORTED_BROWSERS: BrowserSupport[] = [
  {
    name: 'Chrome',
    minVersion: '88.0',
    supported: true,
    features: [
      'ES2020',
      'WebP',
      'AVIF',
      'IntersectionObserver',
      'ResizeObserver',
      'fetch',
      'Promise',
      'async/await',
      'CSS Grid',
      'Flexbox',
      'CSS Custom Properties',
    ],
    polyfillsRequired: [],
  },
  {
    name: 'Firefox',
    minVersion: '85.0',
    supported: true,
    features: [
      'ES2020',
      'WebP',
      'IntersectionObserver',
      'ResizeObserver',
      'fetch',
      'Promise',
      'async/await',
      'CSS Grid',
      'Flexbox',
      'CSS Custom Properties',
    ],
    polyfillsRequired: ['AVIF'],
  },
  {
    name: 'Safari',
    minVersion: '14.0',
    supported: true,
    features: [
      'ES2020',
      'WebP',
      'IntersectionObserver',
      'ResizeObserver',
      'fetch',
      'Promise',
      'async/await',
      'CSS Grid',
      'Flexbox',
      'CSS Custom Properties',
    ],
    polyfillsRequired: ['AVIF'],
  },
  {
    name: 'Edge',
    minVersion: '88.0',
    supported: true,
    features: [
      'ES2020',
      'WebP',
      'AVIF',
      'IntersectionObserver',
      'ResizeObserver',
      'fetch',
      'Promise',
      'async/await',
      'CSS Grid',
      'Flexbox',
      'CSS Custom Properties',
    ],
    polyfillsRequired: [],
  },
  {
    name: 'Chrome',
    minVersion: '70.0',
    supported: true,
    features: [
      'ES2018',
      'WebP',
      'IntersectionObserver',
      'fetch',
      'Promise',
      'async/await',
      'CSS Grid',
      'Flexbox',
    ],
    polyfillsRequired: ['ResizeObserver', 'AVIF', 'CSS Custom Properties'],
  },
  {
    name: 'Firefox',
    minVersion: '70.0',
    supported: true,
    features: [
      'ES2018',
      'IntersectionObserver',
      'fetch',
      'Promise',
      'async/await',
      'CSS Grid',
      'Flexbox',
    ],
    polyfillsRequired: [
      'ResizeObserver',
      'WebP',
      'AVIF',
      'CSS Custom Properties',
    ],
  },
  {
    name: 'Safari',
    minVersion: '12.0',
    supported: true,
    features: [
      'ES2018',
      'IntersectionObserver',
      'fetch',
      'Promise',
      'async/await',
      'CSS Grid',
      'Flexbox',
    ],
    polyfillsRequired: [
      'ResizeObserver',
      'WebP',
      'AVIF',
      'CSS Custom Properties',
    ],
  },
  {
    name: 'IE',
    minVersion: '11.0',
    supported: false,
    features: ['ES5', 'Flexbox (partial)'],
    polyfillsRequired: [
      'Promise',
      'fetch',
      'Map',
      'Set',
      'WeakMap',
      'WeakSet',
      'Symbol',
      'Object.assign',
      'Array.from',
      'Array.includes',
      'String.includes',
      'String.startsWith',
      'String.endsWith',
      'IntersectionObserver',
      'ResizeObserver',
      'requestAnimationFrame',
      'requestIdleCallback',
      'CSS Grid',
      'CSS Custom Properties',
      'WebP',
      'AVIF',
    ],
  },
];

/**
 * 功能特性配置
 */
export const FEATURE_CONFIG = {
  // 核心JavaScript特性
  javascript: {
    Promise: {
      required: true,
      polyfillAvailable: true,
      fallback: 'callback-based async',
    },
    fetch: {
      required: true,
      polyfillAvailable: true,
      fallback: 'XMLHttpRequest',
    },
    Map: {
      required: false,
      polyfillAvailable: true,
      fallback: 'Object',
    },
    Set: {
      required: false,
      polyfillAvailable: true,
      fallback: 'Array',
    },
    Symbol: {
      required: false,
      polyfillAvailable: true,
      fallback: 'string keys',
    },
    Proxy: {
      required: false,
      polyfillAvailable: false,
      fallback: 'direct object access',
    },
  },

  // Web APIs
  webAPIs: {
    IntersectionObserver: {
      required: true,
      polyfillAvailable: true,
      fallback: 'scroll event listeners',
    },
    ResizeObserver: {
      required: false,
      polyfillAvailable: true,
      fallback: 'window resize events',
    },
    MutationObserver: {
      required: false,
      polyfillAvailable: false,
      fallback: 'manual DOM monitoring',
    },
    requestAnimationFrame: {
      required: true,
      polyfillAvailable: true,
      fallback: 'setTimeout',
    },
    requestIdleCallback: {
      required: false,
      polyfillAvailable: true,
      fallback: 'setTimeout',
    },
  },

  // 存储APIs
  storage: {
    localStorage: {
      required: false,
      polyfillAvailable: false,
      fallback: 'memory storage',
    },
    sessionStorage: {
      required: false,
      polyfillAvailable: false,
      fallback: 'memory storage',
    },
    IndexedDB: {
      required: false,
      polyfillAvailable: true,
      fallback: 'localStorage',
    },
  },

  // 图片格式
  imageFormats: {
    WebP: {
      required: false,
      polyfillAvailable: false,
      fallback: 'JPEG/PNG',
    },
    AVIF: {
      required: false,
      polyfillAvailable: false,
      fallback: 'WebP/JPEG',
    },
  },

  // CSS特性
  css: {
    'CSS Grid': {
      required: true,
      polyfillAvailable: false,
      fallback: 'Flexbox layout',
    },
    Flexbox: {
      required: true,
      polyfillAvailable: false,
      fallback: 'float layout',
    },
    'CSS Custom Properties': {
      required: false,
      polyfillAvailable: true,
      fallback: 'SCSS variables',
    },
    'CSS Modules': {
      required: false,
      polyfillAvailable: false,
      fallback: 'global CSS',
    },
  },
};

/**
 * 降级策略配置
 */
export const FALLBACK_STRATEGIES = {
  // 图片加载降级
  imageLoading: {
    modern: ['AVIF', 'WebP', 'JPEG'],
    legacy: ['JPEG', 'PNG'],
    placeholder: '/images/placeholder.svg',
  },

  // 动画降级
  animations: {
    modern: 'CSS animations + requestAnimationFrame',
    legacy: 'CSS transitions only',
    disabled: 'no animations (prefers-reduced-motion)',
  },

  // 布局降级
  layout: {
    modern: 'CSS Grid + Flexbox',
    intermediate: 'Flexbox only',
    legacy: 'Float + Table layout',
  },

  // 交互降级
  interactions: {
    modern: 'Touch + Mouse + Keyboard',
    legacy: 'Mouse + Keyboard only',
    minimal: 'Click events only',
  },
};

/**
 * 性能优化配置
 */
export const PERFORMANCE_CONFIG = {
  // 代码分割
  codeSplitting: {
    enabled: true,
    chunkSize: {
      small: 50 * 1024, // 50KB
      medium: 200 * 1024, // 200KB
      large: 500 * 1024, // 500KB
    },
  },

  // 预加载策略
  preloading: {
    critical: ['fonts', 'above-fold-images'],
    important: ['hero-images', 'primary-scripts'],
    optional: ['below-fold-images', 'secondary-scripts'],
  },

  // 缓存策略
  caching: {
    static: '1y', // 静态资源缓存1年
    dynamic: '1h', // 动态内容缓存1小时
    api: '5m', // API响应缓存5分钟
  },
};

/**
 * 错误处理配置
 */
export const ERROR_HANDLING_CONFIG = {
  // 兼容性错误
  compatibility: {
    logLevel: 'warn',
    reportToAnalytics: true,
    showUserMessage: false,
  },

  // 功能降级错误
  fallback: {
    logLevel: 'info',
    reportToAnalytics: true,
    showUserMessage: false,
  },

  // 严重错误
  critical: {
    logLevel: 'error',
    reportToAnalytics: true,
    showUserMessage: true,
    fallbackPage: '/error',
  },
};

/**
 * 开发环境配置
 */
export const DEVELOPMENT_CONFIG = {
  // 兼容性警告
  warnings: {
    enabled: true,
    showInConsole: true,
    showInUI: false,
  },

  // 调试工具
  debugging: {
    compatibilityReport: true,
    performanceMetrics: true,
    featureDetection: true,
  },

  // 测试配置
  testing: {
    browserStack: false,
    localBrowsers: ['Chrome', 'Firefox', 'Safari', 'Edge'],
    mobileDevices: ['iPhone', 'Android'],
  },
};

/**
 * 生产环境配置
 */
export const PRODUCTION_CONFIG = {
  // 监控配置
  monitoring: {
    errorTracking: true,
    performanceTracking: true,
    userAnalytics: true,
  },

  // 优化配置
  optimization: {
    minification: true,
    compression: true,
    treeshaking: true,
    bundleAnalysis: false,
  },

  // 安全配置
  security: {
    csp: true,
    hsts: true,
    xssProtection: true,
    contentTypeNosniff: true,
  },
};
