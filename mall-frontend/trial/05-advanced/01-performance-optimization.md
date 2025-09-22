# 第1章：前端性能优化策略 ⚡

> _"性能不是功能，而是用户体验的基础！"_ 🚀

## 📚 本章导览

前端性能优化是现代Web开发的核心技能之一。随着用户对体验要求的不断提高，以及移动设备和网络环境的多样化，性能优化已经从"锦上添花"变成了"必备技能"。本章将从性能优化的基础理论出发，深入探讨各种优化技术和策略，结合Mall-Frontend项目的实际案例，提供完整的性能优化解决方案。

### 🎯 学习目标

通过本章学习，你将掌握：

- **性能指标体系** - 理解Core Web Vitals和性能测量方法
- **加载性能优化** - 掌握资源加载和首屏优化技术
- **运行时性能优化** - 学会React性能优化和内存管理
- **网络性能优化** - 理解HTTP缓存和CDN优化策略
- **构建优化** - 掌握Webpack/Vite的构建优化技巧
- **图像优化** - 学会现代图像格式和懒加载技术
- **监控与分析** - 构建完整的性能监控体系
- **移动端优化** - 掌握移动设备的特殊优化策略

### 🛠️ 技术栈概览

```typescript
{
  "performanceOptimization": {
    "metrics": ["Core Web Vitals", "FCP", "LCP", "FID", "CLS", "TTFB"],
    "tools": ["Lighthouse", "WebPageTest", "Chrome DevTools", "Performance API"],
    "bundling": ["Webpack", "Vite", "Rollup", "Parcel", "esbuild"],
    "caching": ["Service Worker", "HTTP Cache", "CDN", "Browser Cache"]
  },
  "monitoring": {
    "realUserMonitoring": ["Google Analytics", "New Relic", "DataDog"],
    "syntheticMonitoring": ["Lighthouse CI", "SpeedCurve", "Pingdom"],
    "errorTracking": ["Sentry", "LogRocket", "Bugsnag"],
    "analytics": ["Google PageSpeed Insights", "GTmetrix", "WebPageTest"]
  },
  "optimization": {
    "images": ["WebP", "AVIF", "Lazy Loading", "Responsive Images"],
    "fonts": ["Font Display", "Preload", "Subset", "Variable Fonts"],
    "javascript": ["Code Splitting", "Tree Shaking", "Minification", "Compression"],
    "css": ["Critical CSS", "CSS Modules", "Purge CSS", "Atomic CSS"]
  }
}
```

### 📖 本章目录

- [性能指标与测量](#性能指标与测量)
- [加载性能优化](#加载性能优化)
- [运行时性能优化](#运行时性能优化)
- [网络性能优化](#网络性能优化)
- [构建优化策略](#构建优化策略)
- [图像与媒体优化](#图像与媒体优化)
- [字体优化](#字体优化)
- [CSS性能优化](#css性能优化)
- [JavaScript性能优化](#javascript性能优化)
- [移动端性能优化](#移动端性能优化)
- [性能监控体系](#性能监控体系)
- [Mall-Frontend性能优化实践](#mall-frontend性能优化实践)
- [面试常考知识点](#面试常考知识点)
- [实战练习](#实战练习)

---

## 📊 性能指标与测量

### Core Web Vitals核心指标

Google的Core Web Vitals是衡量用户体验的关键指标：

```typescript
// Core Web Vitals指标定义
interface CoreWebVitals {
  // 最大内容绘制 (Largest Contentful Paint)
  LCP: {
    description: '页面主要内容加载完成的时间';
    goodThreshold: '≤ 2.5秒';
    needsImprovement: '2.5秒 - 4.0秒';
    poor: '> 4.0秒';
    optimization: [
      '优化服务器响应时间',
      '消除阻塞渲染的资源',
      '优化CSS和JavaScript',
      '使用CDN加速资源加载',
    ];
  };

  // 首次输入延迟 (First Input Delay)
  FID: {
    description: '用户首次与页面交互到浏览器响应的时间';
    goodThreshold: '≤ 100毫秒';
    needsImprovement: '100毫秒 - 300毫秒';
    poor: '> 300毫秒';
    optimization: [
      '减少JavaScript执行时间',
      '拆分长任务',
      '使用Web Workers',
      '优化第三方代码',
    ];
  };

  // 累积布局偏移 (Cumulative Layout Shift)
  CLS: {
    description: '页面生命周期内所有意外布局偏移的累积分数';
    goodThreshold: '≤ 0.1';
    needsImprovement: '0.1 - 0.25';
    poor: '> 0.25';
    optimization: [
      '为图像和视频设置尺寸属性',
      '避免在现有内容上方插入内容',
      '使用transform动画而非改变布局的属性',
      '预留广告和嵌入内容的空间',
    ];
  };
}

// 其他重要性能指标
interface AdditionalMetrics {
  // 首次内容绘制 (First Contentful Paint)
  FCP: {
    description: '浏览器首次绘制任何文本、图像或非白色canvas的时间';
    goodThreshold: '≤ 1.8秒';
    measurement: 'Performance API';
  };

  // 首次有意义绘制 (First Meaningful Paint)
  FMP: {
    description: '页面主要内容对用户可见的时间';
    deprecated: true;
    replacedBy: 'LCP';
  };

  // 可交互时间 (Time to Interactive)
  TTI: {
    description: '页面完全可交互的时间';
    goodThreshold: '≤ 3.8秒';
    factors: [
      '页面显示有用内容',
      '事件处理程序已注册',
      '页面在50ms内响应用户交互',
    ];
  };

  // 首字节时间 (Time to First Byte)
  TTFB: {
    description: '从请求开始到接收到第一个字节的时间';
    goodThreshold: '≤ 600毫秒';
    optimization: ['优化服务器配置', '使用CDN', '减少重定向', '启用HTTP/2'];
  };

  // 速度指数 (Speed Index)
  SI: {
    description: '页面内容可见填充的速度';
    goodThreshold: '≤ 3.4秒';
    calculation: '基于视觉进度的积分';
  };
}

// 性能测量实现
class PerformanceMonitor {
  private observer: PerformanceObserver | null = null;
  private metrics: Map<string, number> = new Map();

  constructor() {
    this.initializeObserver();
    this.measureCoreWebVitals();
  }

  // 初始化性能观察器
  private initializeObserver(): void {
    if ('PerformanceObserver' in window) {
      this.observer = new PerformanceObserver(list => {
        for (const entry of list.getEntries()) {
          this.processPerformanceEntry(entry);
        }
      });

      // 观察各种性能指标
      try {
        this.observer.observe({
          entryTypes: [
            'navigation',
            'paint',
            'largest-contentful-paint',
            'first-input',
            'layout-shift',
          ],
        });
      } catch (error) {
        console.warn('Performance Observer not fully supported:', error);
      }
    }
  }

  // 处理性能条目
  private processPerformanceEntry(entry: PerformanceEntry): void {
    switch (entry.entryType) {
      case 'navigation':
        this.handleNavigationEntry(entry as PerformanceNavigationTiming);
        break;
      case 'paint':
        this.handlePaintEntry(entry as PerformancePaintTiming);
        break;
      case 'largest-contentful-paint':
        this.handleLCPEntry(entry as any);
        break;
      case 'first-input':
        this.handleFIDEntry(entry as any);
        break;
      case 'layout-shift':
        this.handleCLSEntry(entry as any);
        break;
    }
  }

  // 处理导航性能
  private handleNavigationEntry(entry: PerformanceNavigationTiming): void {
    const ttfb = entry.responseStart - entry.requestStart;
    const domContentLoaded =
      entry.domContentLoadedEventEnd - entry.navigationStart;
    const loadComplete = entry.loadEventEnd - entry.navigationStart;

    this.metrics.set('TTFB', ttfb);
    this.metrics.set('DOMContentLoaded', domContentLoaded);
    this.metrics.set('LoadComplete', loadComplete);

    this.reportMetric('TTFB', ttfb);
    this.reportMetric('DOMContentLoaded', domContentLoaded);
    this.reportMetric('LoadComplete', loadComplete);
  }

  // 处理绘制性能
  private handlePaintEntry(entry: PerformancePaintTiming): void {
    const value = entry.startTime;
    this.metrics.set(entry.name, value);

    if (entry.name === 'first-contentful-paint') {
      this.reportMetric('FCP', value);
    }
  }

  // 处理LCP
  private handleLCPEntry(entry: any): void {
    const lcp = entry.startTime;
    this.metrics.set('LCP', lcp);
    this.reportMetric('LCP', lcp);
  }

  // 处理FID
  private handleFIDEntry(entry: any): void {
    const fid = entry.processingStart - entry.startTime;
    this.metrics.set('FID', fid);
    this.reportMetric('FID', fid);
  }

  // 处理CLS
  private handleCLSEntry(entry: any): void {
    if (!entry.hadRecentInput) {
      const currentCLS = this.metrics.get('CLS') || 0;
      const newCLS = currentCLS + entry.value;
      this.metrics.set('CLS', newCLS);
      this.reportMetric('CLS', newCLS);
    }
  }

  // 测量Core Web Vitals
  private measureCoreWebVitals(): void {
    // 使用web-vitals库进行精确测量
    if (typeof window !== 'undefined') {
      import('web-vitals').then(
        ({ getCLS, getFID, getFCP, getLCP, getTTFB }) => {
          getCLS(this.onCLS.bind(this));
          getFID(this.onFID.bind(this));
          getFCP(this.onFCP.bind(this));
          getLCP(this.onLCP.bind(this));
          getTTFB(this.onTTFB.bind(this));
        }
      );
    }
  }

  // Core Web Vitals回调函数
  private onCLS(metric: any): void {
    this.reportMetric('CLS', metric.value);
  }

  private onFID(metric: any): void {
    this.reportMetric('FID', metric.value);
  }

  private onFCP(metric: any): void {
    this.reportMetric('FCP', metric.value);
  }

  private onLCP(metric: any): void {
    this.reportMetric('LCP', metric.value);
  }

  private onTTFB(metric: any): void {
    this.reportMetric('TTFB', metric.value);
  }

  // 上报性能指标
  private reportMetric(name: string, value: number): void {
    // 发送到分析服务
    if (typeof gtag !== 'undefined') {
      gtag('event', name, {
        event_category: 'Web Vitals',
        value: Math.round(value),
        non_interaction: true,
      });
    }

    // 发送到自定义分析服务
    this.sendToAnalytics(name, value);
  }

  // 发送到分析服务
  private sendToAnalytics(name: string, value: number): void {
    fetch('/api/analytics/performance', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        metric: name,
        value: value,
        timestamp: Date.now(),
        url: window.location.href,
        userAgent: navigator.userAgent,
      }),
    }).catch(error => {
      console.warn('Failed to send performance metric:', error);
    });
  }

  // 获取所有指标
  public getMetrics(): Map<string, number> {
    return new Map(this.metrics);
  }

  // 销毁监控器
  public destroy(): void {
    if (this.observer) {
      this.observer.disconnect();
      this.observer = null;
    }
  }
}

// 使用示例
const performanceMonitor = new PerformanceMonitor();

// 在应用卸载时清理
window.addEventListener('beforeunload', () => {
  performanceMonitor.destroy();
});
```

---

## 🚀 加载性能优化

### 资源加载优化策略

```typescript
// 资源加载优化配置
const loadingOptimization = {
  // 1. 关键资源优化
  criticalResources: {
    // 关键CSS内联
    inlineCriticalCSS: {
      description: '将首屏CSS内联到HTML中',
      implementation: `
        <!-- 内联关键CSS -->
        <style>
          /* 首屏样式 */
          .header { background: #fff; height: 60px; }
          .hero { min-height: 400px; background: #f5f5f5; }
          .loading { display: flex; justify-content: center; }
        </style>

        <!-- 异步加载非关键CSS -->
        <link rel="preload" href="/css/non-critical.css" as="style" onload="this.onload=null;this.rel='stylesheet'">
        <noscript><link rel="stylesheet" href="/css/non-critical.css"></noscript>
      `,
      tools: ['Critical', 'Critters', 'PurgeCSS'],
    },

    // 资源预加载
    resourcePreloading: {
      description: '预加载关键资源',
      strategies: {
        preload: {
          usage: '预加载当前页面需要的关键资源',
          example: `
            <!-- 预加载关键字体 -->
            <link rel="preload" href="/fonts/inter.woff2" as="font" type="font/woff2" crossorigin>

            <!-- 预加载关键图片 -->
            <link rel="preload" href="/images/hero.webp" as="image">

            <!-- 预加载关键脚本 -->
            <link rel="preload" href="/js/critical.js" as="script">
          `,
        },

        prefetch: {
          usage: '预获取下一页面可能需要的资源',
          example: `
            <!-- 预获取下一页面的资源 -->
            <link rel="prefetch" href="/js/product-detail.js">
            <link rel="prefetch" href="/css/product-detail.css">

            <!-- 预获取API数据 -->
            <link rel="prefetch" href="/api/products/trending">
          `,
        },

        preconnect: {
          usage: '预连接到外部域名',
          example: `
            <!-- 预连接到CDN -->
            <link rel="preconnect" href="https://cdn.example.com">

            <!-- 预连接到字体服务 -->
            <link rel="preconnect" href="https://fonts.googleapis.com">
            <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
          `,
        },
      },
    },
  },

  // 2. 代码分割策略
  codeSplitting: {
    // 路由级分割
    routeBasedSplitting: {
      description: '按路由拆分代码',
      implementation: `
        // React Router + Lazy Loading
        import { lazy, Suspense } from 'react';
        import { Routes, Route } from 'react-router-dom';

        // 懒加载页面组件
        const HomePage = lazy(() => import('./pages/HomePage'));
        const ProductPage = lazy(() => import('./pages/ProductPage'));
        const CartPage = lazy(() => import('./pages/CartPage'));
        const CheckoutPage = lazy(() => import('./pages/CheckoutPage'));

        // 加载中组件
        const PageLoader = () => (
          <div className="page-loader">
            <div className="spinner" />
            <p>正在加载页面...</p>
          </div>
        );

        const App = () => (
          <Suspense fallback={<PageLoader />}>
            <Routes>
              <Route path="/" element={<HomePage />} />
              <Route path="/products/*" element={<ProductPage />} />
              <Route path="/cart" element={<CartPage />} />
              <Route path="/checkout" element={<CheckoutPage />} />
            </Routes>
          </Suspense>
        );
      `,
    },

    // 组件级分割
    componentBasedSplitting: {
      description: '按组件拆分代码',
      implementation: `
        // 懒加载重型组件
        const HeavyChart = lazy(() => import('./components/HeavyChart'));
        const VideoPlayer = lazy(() => import('./components/VideoPlayer'));
        const RichTextEditor = lazy(() => import('./components/RichTextEditor'));

        // 条件懒加载
        const ConditionalComponent = ({ showChart }) => {
          const [ChartComponent, setChartComponent] = useState(null);

          useEffect(() => {
            if (showChart && !ChartComponent) {
              import('./components/HeavyChart').then(module => {
                setChartComponent(() => module.default);
              });
            }
          }, [showChart, ChartComponent]);

          return (
            <div>
              {showChart && ChartComponent && (
                <Suspense fallback={<div>加载图表中...</div>}>
                  <ChartComponent />
                </Suspense>
              )}
            </div>
          );
        };
      `,
    },

    // 第三方库分割
    vendorSplitting: {
      description: '分离第三方库代码',
      webpackConfig: `
        // webpack.config.js
        module.exports = {
          optimization: {
            splitChunks: {
              chunks: 'all',
              cacheGroups: {
                // React相关库
                react: {
                  test: /[\\/]node_modules[\\/](react|react-dom|react-router)[\\/]/,
                  name: 'react',
                  chunks: 'all',
                },

                // UI库
                ui: {
                  test: /[\\/]node_modules[\\/](@mui|antd|@chakra-ui)[\\/]/,
                  name: 'ui',
                  chunks: 'all',
                },

                // 工具库
                utils: {
                  test: /[\\/]node_modules[\\/](lodash|moment|date-fns)[\\/]/,
                  name: 'utils',
                  chunks: 'all',
                },

                // 其他第三方库
                vendor: {
                  test: /[\\/]node_modules[\\/]/,
                  name: 'vendor',
                  chunks: 'all',
                  priority: -10,
                },
              },
            },
          },
        };
      `,
    },
  },

  // 3. 懒加载实现
  lazyLoading: {
    // 图片懒加载
    imageLazyLoading: {
      description: '延迟加载图片资源',
      implementation: `
        // 使用Intersection Observer实现图片懒加载
        const LazyImage = ({ src, alt, className, ...props }) => {
          const [imageSrc, setImageSrc] = useState('');
          const [isLoaded, setIsLoaded] = useState(false);
          const imgRef = useRef(null);

          useEffect(() => {
            const observer = new IntersectionObserver(
              (entries) => {
                entries.forEach(entry => {
                  if (entry.isIntersecting) {
                    setImageSrc(src);
                    observer.unobserve(entry.target);
                  }
                });
              },
              { threshold: 0.1 }
            );

            if (imgRef.current) {
              observer.observe(imgRef.current);
            }

            return () => observer.disconnect();
          }, [src]);

          return (
            <div ref={imgRef} className={className}>
              {imageSrc && (
                <img
                  src={imageSrc}
                  alt={alt}
                  onLoad={() => setIsLoaded(true)}
                  style={{
                    opacity: isLoaded ? 1 : 0,
                    transition: 'opacity 0.3s ease'
                  }}
                  {...props}
                />
              )}
              {!isLoaded && (
                <div className="image-placeholder">
                  <div className="skeleton" />
                </div>
              )}
            </div>
          );
        };

        // 使用现代浏览器的loading属性
        const ModernLazyImage = ({ src, alt, ...props }) => (
          <img
            src={src}
            alt={alt}
            loading="lazy"
            decoding="async"
            {...props}
          />
        );
      `,
    },

    // 内容懒加载
    contentLazyLoading: {
      description: '延迟加载页面内容',
      implementation: `
        // 虚拟滚动实现
        const VirtualList = ({ items, itemHeight, containerHeight }) => {
          const [scrollTop, setScrollTop] = useState(0);
          const containerRef = useRef(null);

          // 计算可见范围
          const startIndex = Math.floor(scrollTop / itemHeight);
          const endIndex = Math.min(
            startIndex + Math.ceil(containerHeight / itemHeight) + 1,
            items.length
          );

          const visibleItems = items.slice(startIndex, endIndex);

          const handleScroll = (e) => {
            setScrollTop(e.target.scrollTop);
          };

          return (
            <div
              ref={containerRef}
              style={{ height: containerHeight, overflow: 'auto' }}
              onScroll={handleScroll}
            >
              <div style={{ height: items.length * itemHeight, position: 'relative' }}>
                {visibleItems.map((item, index) => (
                  <div
                    key={startIndex + index}
                    style={{
                      position: 'absolute',
                      top: (startIndex + index) * itemHeight,
                      height: itemHeight,
                      width: '100%'
                    }}
                  >
                    <ItemComponent item={item} />
                  </div>
                ))}
              </div>
            </div>
          );
        };
      `,
    },
  },

  // 4. Service Worker缓存
  serviceWorkerCaching: {
    description: '使用Service Worker实现离线缓存',
    implementation: `
      // sw.js - Service Worker
      const CACHE_NAME = 'mall-frontend-v1';
      const STATIC_ASSETS = [
        '/',
        '/static/css/main.css',
        '/static/js/main.js',
        '/manifest.json'
      ];

      // 安装事件 - 缓存静态资源
      self.addEventListener('install', (event) => {
        event.waitUntil(
          caches.open(CACHE_NAME)
            .then(cache => cache.addAll(STATIC_ASSETS))
            .then(() => self.skipWaiting())
        );
      });

      // 激活事件 - 清理旧缓存
      self.addEventListener('activate', (event) => {
        event.waitUntil(
          caches.keys().then(cacheNames => {
            return Promise.all(
              cacheNames
                .filter(cacheName => cacheName !== CACHE_NAME)
                .map(cacheName => caches.delete(cacheName))
            );
          }).then(() => self.clients.claim())
        );
      });

      // 拦截请求 - 缓存策略
      self.addEventListener('fetch', (event) => {
        const { request } = event;

        // 静态资源：缓存优先
        if (request.destination === 'script' || request.destination === 'style') {
          event.respondWith(
            caches.match(request).then(response => {
              return response || fetch(request).then(fetchResponse => {
                const responseClone = fetchResponse.clone();
                caches.open(CACHE_NAME).then(cache => {
                  cache.put(request, responseClone);
                });
                return fetchResponse;
              });
            })
          );
        }

        // API请求：网络优先
        else if (request.url.includes('/api/')) {
          event.respondWith(
            fetch(request).then(response => {
              const responseClone = response.clone();
              caches.open(CACHE_NAME).then(cache => {
                cache.put(request, responseClone);
              });
              return response;
            }).catch(() => {
              return caches.match(request);
            })
          );
        }

        // 其他请求：默认处理
        else {
          event.respondWith(
            caches.match(request).then(response => {
              return response || fetch(request);
            })
          );
        }
      });
    `,
  },
};
```

---

## ⚡ 运行时性能优化

### React性能优化策略

```typescript
// React性能优化技术
const reactPerformanceOptimization = {
  // 1. 组件优化
  componentOptimization: {
    // React.memo优化
    memoization: {
      description: '使用React.memo防止不必要的重渲染',
      implementation: `
        // 基础memo使用
        const ProductCard = React.memo(({ product, onAddToCart }) => {
          return (
            <div className="product-card">
              <img src={product.image} alt={product.name} />
              <h3>{product.name}</h3>
              <p>{product.price}</p>
              <button onClick={() => onAddToCart(product)}>
                加入购物车
              </button>
            </div>
          );
        });

        // 自定义比较函数
        const ProductCardOptimized = React.memo(({ product, onAddToCart }) => {
          // 组件实现
        }, (prevProps, nextProps) => {
          // 自定义比较逻辑
          return (
            prevProps.product.id === nextProps.product.id &&
            prevProps.product.price === nextProps.product.price &&
            prevProps.product.name === nextProps.product.name
          );
        });

        // 使用useMemo优化复杂计算
        const ProductList = ({ products, filters }) => {
          const filteredProducts = useMemo(() => {
            return products.filter(product => {
              return Object.entries(filters).every(([key, value]) => {
                if (!value) return true;
                return product[key] === value;
              });
            }).sort((a, b) => {
              // 复杂排序逻辑
              return a.price - b.price;
            });
          }, [products, filters]);

          return (
            <div className="product-list">
              {filteredProducts.map(product => (
                <ProductCard key={product.id} product={product} />
              ))}
            </div>
          );
        };
      `,
    },

    // useCallback优化
    callbackOptimization: {
      description: '使用useCallback优化事件处理函数',
      implementation: `
        const ProductManager = () => {
          const [products, setProducts] = useState([]);
          const [selectedCategory, setSelectedCategory] = useState('');

          // 优化事件处理函数
          const handleAddToCart = useCallback((product) => {
            // 添加到购物车的逻辑
            addToCart(product);

            // 发送分析事件
            analytics.track('add_to_cart', {
              product_id: product.id,
              product_name: product.name,
              price: product.price
            });
          }, []); // 空依赖数组，函数不会重新创建

          const handleCategoryChange = useCallback((category) => {
            setSelectedCategory(category);

            // 发送分析事件
            analytics.track('category_filter', {
              category: category
            });
          }, []); // 空依赖数组

          const handleProductUpdate = useCallback((productId, updates) => {
            setProducts(prevProducts =>
              prevProducts.map(product =>
                product.id === productId
                  ? { ...product, ...updates }
                  : product
              )
            );
          }, []); // 使用函数式更新，无需依赖products

          return (
            <div>
              <CategoryFilter
                selectedCategory={selectedCategory}
                onCategoryChange={handleCategoryChange}
              />
              <ProductList
                products={products}
                onAddToCart={handleAddToCart}
                onProductUpdate={handleProductUpdate}
              />
            </div>
          );
        };
      `,
    },

    // 状态优化
    stateOptimization: {
      description: '优化状态管理减少重渲染',
      implementation: `
        // 状态分离 - 避免单一大状态对象
        const useProductState = () => {
          const [products, setProducts] = useState([]);
          const [loading, setLoading] = useState(false);
          const [error, setError] = useState(null);

          return {
            products,
            setProducts,
            loading,
            setLoading,
            error,
            setError
          };
        };

        const useUIState = () => {
          const [selectedCategory, setSelectedCategory] = useState('');
          const [sortBy, setSortBy] = useState('name');
          const [viewMode, setViewMode] = useState('grid');

          return {
            selectedCategory,
            setSelectedCategory,
            sortBy,
            setSortBy,
            viewMode,
            setViewMode
          };
        };

        // 使用useReducer管理复杂状态
        const cartReducer = (state, action) => {
          switch (action.type) {
            case 'ADD_ITEM':
              const existingItem = state.items.find(item => item.id === action.payload.id);
              if (existingItem) {
                return {
                  ...state,
                  items: state.items.map(item =>
                    item.id === action.payload.id
                      ? { ...item, quantity: item.quantity + 1 }
                      : item
                  )
                };
              }
              return {
                ...state,
                items: [...state.items, { ...action.payload, quantity: 1 }]
              };

            case 'REMOVE_ITEM':
              return {
                ...state,
                items: state.items.filter(item => item.id !== action.payload.id)
              };

            case 'UPDATE_QUANTITY':
              return {
                ...state,
                items: state.items.map(item =>
                  item.id === action.payload.id
                    ? { ...item, quantity: action.payload.quantity }
                    : item
                )
              };

            default:
              return state;
          }
        };

        const useCart = () => {
          const [state, dispatch] = useReducer(cartReducer, {
            items: [],
            total: 0
          });

          // 计算总价
          const total = useMemo(() => {
            return state.items.reduce((sum, item) => sum + item.price * item.quantity, 0);
          }, [state.items]);

          return {
            ...state,
            total,
            dispatch
          };
        };
      `,
    },
  },

  // 2. 列表优化
  listOptimization: {
    // 虚拟滚动
    virtualScrolling: {
      description: '大列表虚拟滚动优化',
      implementation: `
        // 自定义虚拟滚动Hook
        const useVirtualScroll = ({
          items,
          itemHeight,
          containerHeight,
          overscan = 5
        }) => {
          const [scrollTop, setScrollTop] = useState(0);

          const startIndex = Math.max(0, Math.floor(scrollTop / itemHeight) - overscan);
          const endIndex = Math.min(
            items.length,
            Math.ceil((scrollTop + containerHeight) / itemHeight) + overscan
          );

          const visibleItems = items.slice(startIndex, endIndex);
          const totalHeight = items.length * itemHeight;
          const offsetY = startIndex * itemHeight;

          return {
            visibleItems,
            totalHeight,
            offsetY,
            startIndex,
            endIndex,
            setScrollTop
          };
        };

        // 虚拟滚动组件
        const VirtualProductList = ({ products }) => {
          const containerRef = useRef(null);
          const [containerHeight, setContainerHeight] = useState(600);

          const {
            visibleItems,
            totalHeight,
            offsetY,
            setScrollTop
          } = useVirtualScroll({
            items: products,
            itemHeight: 120,
            containerHeight
          });

          const handleScroll = useCallback((e) => {
            setScrollTop(e.target.scrollTop);
          }, [setScrollTop]);

          useEffect(() => {
            const updateHeight = () => {
              if (containerRef.current) {
                setContainerHeight(containerRef.current.clientHeight);
              }
            };

            updateHeight();
            window.addEventListener('resize', updateHeight);
            return () => window.removeEventListener('resize', updateHeight);
          }, []);

          return (
            <div
              ref={containerRef}
              className="virtual-list-container"
              style={{ height: '100%', overflow: 'auto' }}
              onScroll={handleScroll}
            >
              <div style={{ height: totalHeight, position: 'relative' }}>
                <div
                  style={{
                    transform: \`translateY(\${offsetY}px)\`,
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    right: 0
                  }}
                >
                  {visibleItems.map((product, index) => (
                    <ProductCard
                      key={product.id}
                      product={product}
                      style={{ height: 120 }}
                    />
                  ))}
                </div>
              </div>
            </div>
          );
        };
      `,
    },

    // 分页和无限滚动
    infiniteScrolling: {
      description: '无限滚动实现',
      implementation: `
        // 无限滚动Hook
        const useInfiniteScroll = (fetchMore) => {
          const [isFetching, setIsFetching] = useState(false);

          useEffect(() => {
            const handleScroll = () => {
              if (window.innerHeight + document.documentElement.scrollTop !== document.documentElement.offsetHeight || isFetching) return;
              setIsFetching(true);
            };

            window.addEventListener('scroll', handleScroll);
            return () => window.removeEventListener('scroll', handleScroll);
          }, [isFetching]);

          useEffect(() => {
            if (!isFetching) return;
            fetchMoreData();
          }, [isFetching]);

          const fetchMoreData = async () => {
            await fetchMore();
            setIsFetching(false);
          };

          return [isFetching, setIsFetching];
        };

        // 无限滚动产品列表
        const InfiniteProductList = () => {
          const [products, setProducts] = useState([]);
          const [hasMore, setHasMore] = useState(true);
          const [page, setPage] = useState(1);

          const fetchMoreProducts = useCallback(async () => {
            try {
              const response = await fetch(\`/api/products?page=\${page}&limit=20\`);
              const newProducts = await response.json();

              if (newProducts.length === 0) {
                setHasMore(false);
              } else {
                setProducts(prev => [...prev, ...newProducts]);
                setPage(prev => prev + 1);
              }
            } catch (error) {
              console.error('Failed to fetch products:', error);
            }
          }, [page]);

          const [isFetching] = useInfiniteScroll(fetchMoreProducts);

          return (
            <div>
              <div className="product-grid">
                {products.map(product => (
                  <ProductCard key={product.id} product={product} />
                ))}
              </div>

              {isFetching && hasMore && (
                <div className="loading-indicator">
                  <div className="spinner" />
                  <p>加载更多产品...</p>
                </div>
              )}

              {!hasMore && (
                <div className="end-indicator">
                  <p>已加载全部产品</p>
                </div>
              )}
            </div>
          );
        };
      `,
    },
  },
};
```

---

## 🎯 面试常考知识点

### 1. 性能优化策略

**Q: 前端性能优化有哪些主要策略？**

**A: 前端性能优化策略分类：**

```typescript
// 性能优化策略分类
const performanceOptimizationStrategies = {
  // 加载时性能优化
  loadingPerformance: {
    strategies: [
      '减少HTTP请求数量',
      '压缩和合并资源',
      '使用CDN加速',
      '启用浏览器缓存',
      '代码分割和懒加载',
      '关键资源优先加载',
      '预加载和预获取',
    ],

    techniques: {
      bundleOptimization: {
        description: '打包优化',
        methods: [
          'Tree Shaking移除无用代码',
          'Code Splitting按需加载',
          'Vendor Splitting分离第三方库',
          'Dynamic Import动态导入',
        ],
      },

      resourceOptimization: {
        description: '资源优化',
        methods: [
          '图片压缩和格式优化',
          'CSS和JS压缩混淆',
          'Gzip/Brotli压缩',
          '字体优化和子集化',
        ],
      },

      cacheOptimization: {
        description: '缓存优化',
        methods: [
          'HTTP缓存策略',
          'Service Worker缓存',
          'CDN缓存',
          '浏览器缓存',
        ],
      },
    },
  },

  // 运行时性能优化
  runtimePerformance: {
    strategies: [
      '减少DOM操作',
      '优化重排和重绘',
      '使用虚拟滚动',
      '防抖和节流',
      '内存泄漏防护',
      '长任务拆分',
      'Web Workers使用',
    ],

    reactOptimization: {
      description: 'React性能优化',
      methods: [
        'React.memo防止不必要重渲染',
        'useMemo缓存计算结果',
        'useCallback缓存函数引用',
        '状态提升和下沉',
        '组件懒加载',
        'Context优化',
      ],
    },
  },

  // 感知性能优化
  perceivedPerformance: {
    strategies: [
      '骨架屏显示',
      '渐进式加载',
      '优化动画性能',
      '减少布局偏移',
      '优化字体加载',
      '预加载关键资源',
    ],
  },
};
```

### 2. Core Web Vitals优化

**Q: 如何优化Core Web Vitals指标？**

**A: Core Web Vitals优化方法：**

```typescript
const coreWebVitalsOptimization = {
  // LCP优化
  LCP: {
    targetMetric: '≤ 2.5秒',
    optimizationMethods: [
      '优化服务器响应时间',
      '移除阻塞渲染的资源',
      '优化CSS加载',
      '预加载关键资源',
      '使用CDN',
      '压缩图片和文本',
    ],

    implementation: `
      // 预加载LCP元素
      <link rel="preload" href="/images/hero.webp" as="image">

      // 优化关键CSS
      <style>
        /* 内联关键CSS */
        .hero { background-image: url('/images/hero.webp'); }
      </style>

      // 异步加载非关键CSS
      <link rel="preload" href="/css/non-critical.css" as="style"
            onload="this.onload=null;this.rel='stylesheet'">
    `,
  },

  // FID优化
  FID: {
    targetMetric: '≤ 100毫秒',
    optimizationMethods: [
      '减少JavaScript执行时间',
      '拆分长任务',
      '移除无用代码',
      '使用Web Workers',
      '优化第三方脚本',
      '延迟非关键JavaScript',
    ],

    implementation: `
      // 拆分长任务
      function processLargeArray(array) {
        return new Promise(resolve => {
          const batchSize = 1000;
          let index = 0;

          function processBatch() {
            const endIndex = Math.min(index + batchSize, array.length);

            for (let i = index; i < endIndex; i++) {
              // 处理数组项
              processItem(array[i]);
            }

            index = endIndex;

            if (index < array.length) {
              // 让出主线程
              setTimeout(processBatch, 0);
            } else {
              resolve();
            }
          }

          processBatch();
        });
      }

      // 使用Web Workers
      const worker = new Worker('/js/heavy-computation.js');
      worker.postMessage(data);
      worker.onmessage = (e) => {
        updateUI(e.data);
      };
    `,
  },

  // CLS优化
  CLS: {
    targetMetric: '≤ 0.1',
    optimizationMethods: [
      '为图片和视频设置尺寸',
      '预留广告空间',
      '避免在现有内容上方插入内容',
      '使用transform动画',
      '优化字体加载',
      '避免无尺寸元素',
    ],

    implementation: `
      <!-- 为图片设置尺寸 -->
      <img src="/image.jpg" width="400" height="300" alt="Product">

      <!-- 使用aspect-ratio -->
      <div style="aspect-ratio: 16/9;">
        <img src="/image.jpg" alt="Product" style="width: 100%; height: 100%;">
      </div>

      <!-- 预留广告空间 -->
      <div class="ad-container" style="min-height: 250px;">
        <!-- 广告内容 -->
      </div>

      <!-- 使用transform动画而非改变布局 -->
      .slide-in {
        transform: translateX(-100%);
        transition: transform 0.3s ease;
      }

      .slide-in.active {
        transform: translateX(0);
      }
    `,
  },
};
```

---

## 📚 实战练习

### 练习1：性能监控系统

**任务**: 为Mall-Frontend构建完整的性能监控系统。

**要求**:

- 监控Core Web Vitals指标
- 实现Real User Monitoring
- 构建性能数据分析面板
- 设置性能预警机制

### 练习2：加载性能优化

**任务**: 优化Mall-Frontend的加载性能。

**要求**:

- 实现代码分割和懒加载
- 优化图片和字体加载
- 配置Service Worker缓存
- 优化首屏加载时间

### 练习3：运行时性能优化

**任务**: 优化Mall-Frontend的运行时性能。

**要求**:

- 实现虚拟滚动
- 优化React组件性能
- 处理内存泄漏问题
- 优化动画性能

---

## 📚 本章总结

通过本章学习，我们全面掌握了前端性能优化的核心技术：

### 🎯 核心收获

1. **性能指标体系精通** 📊
   - 掌握了Core Web Vitals的测量和优化
   - 理解了性能监控的实现方法
   - 学会了性能数据的分析和应用

2. **加载性能优化** 🚀
   - 掌握了资源加载优化策略
   - 学会了代码分割和懒加载技术
   - 理解了缓存策略和CDN优化

3. **运行时性能优化** ⚡
   - 掌握了React性能优化技巧
   - 学会了虚拟滚动和无限滚动
   - 理解了内存管理和长任务优化

4. **工程化性能优化** 🔧
   - 掌握了构建工具的性能优化
   - 学会了图像和字体优化技术
   - 理解了现代浏览器的性能特性

5. **监控与分析能力** 📈
   - 掌握了性能监控系统的构建
   - 学会了性能问题的诊断方法
   - 理解了性能优化的持续改进

### 🚀 技术进阶

- **下一步学习**: 测试策略与质量保证
- **实践建议**: 在项目中建立性能优化流程
- **深入方向**: 现代浏览器性能API和优化技术

性能优化是前端开发的永恒主题，掌握系统性的性能优化方法是高级前端工程师的必备技能！ 🎉

---

_下一章我们将学习《测试策略与质量保证》，探索现代前端应用的测试体系！_ 🚀

```

```
