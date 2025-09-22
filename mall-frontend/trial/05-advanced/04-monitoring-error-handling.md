# 第4章：监控与错误处理 📊

> _"监控不是为了发现问题，而是为了预防问题！"_ 🛡️

## 📚 本章导览

监控与错误处理是保障生产环境稳定性的关键技术。在现代前端应用中，随着用户规模的增长和业务复杂度的提升，建立完善的可观测性体系已经成为必需品。本章将从监控理论基础出发，深入探讨各种监控工具、错误处理策略、日志管理最佳实践，结合Mall-Frontend项目的实际案例，构建完整的生产环境监控体系。

### 🎯 学习目标

通过本章学习，你将掌握：

- **监控理论基础** - 理解可观测性三大支柱和监控策略
- **错误追踪系统** - 掌握Sentry、LogRocket等工具的使用
- **性能监控** - 学会Real User Monitoring和Synthetic Monitoring
- **日志管理** - 掌握日志收集、存储、分析的最佳实践
- **告警系统** - 学会设计有效的告警策略和通知机制
- **可视化监控** - 掌握监控仪表板的设计和实现
- **故障排查** - 学会快速定位和解决生产问题
- **SRE实践** - 理解站点可靠性工程的核心理念

### 🛠️ 技术栈概览

```typescript
{
  "errorTracking": {
    "platforms": ["Sentry", "LogRocket", "Bugsnag", "Rollbar", "Airbrake"],
    "features": ["错误聚合", "性能监控", "用户会话", "源码映射", "告警通知"]
  },
  "performanceMonitoring": {
    "rum": ["Google Analytics", "New Relic", "DataDog", "Dynatrace"],
    "synthetic": ["Lighthouse CI", "SpeedCurve", "Pingdom", "GTmetrix"],
    "apm": ["New Relic", "DataDog", "AppDynamics", "Elastic APM"]
  },
  "logging": {
    "collection": ["Winston", "Pino", "Bunyan", "Log4js"],
    "aggregation": ["ELK Stack", "Fluentd", "Logstash", "Vector"],
    "storage": ["Elasticsearch", "Splunk", "CloudWatch", "Loki"],
    "analysis": ["Kibana", "Grafana", "Splunk", "DataDog"]
  },
  "metrics": {
    "collection": ["Prometheus", "StatsD", "InfluxDB", "CloudWatch"],
    "visualization": ["Grafana", "DataDog", "New Relic", "Kibana"],
    "alerting": ["AlertManager", "PagerDuty", "OpsGenie", "Slack"]
  },
  "infrastructure": {
    "monitoring": ["Prometheus", "Nagios", "Zabbix", "DataDog"],
    "tracing": ["Jaeger", "Zipkin", "AWS X-Ray", "DataDog APM"],
    "uptime": ["Pingdom", "UptimeRobot", "StatusCake", "Site24x7"]
  }
}
```

### 📖 本章目录

- [可观测性理论基础](#可观测性理论基础)
- [错误追踪系统对比](#错误追踪系统对比)
- [性能监控实践](#性能监控实践)
- [日志管理体系](#日志管理体系)
- [指标收集与分析](#指标收集与分析)
- [告警系统设计](#告警系统设计)
- [监控仪表板](#监控仪表板)
- [故障排查流程](#故障排查流程)
- [SRE最佳实践](#sre最佳实践)
- [安全监控](#安全监控)
- [成本优化监控](#成本优化监控)
- [Mall-Frontend监控体系](#mall-frontend监控体系)
- [面试常考知识点](#面试常考知识点)
- [实战练习](#实战练习)

---

## 🎯 可观测性理论基础

### 可观测性三大支柱

可观测性（Observability）是现代系统监控的核心理念：

```typescript
// 可观测性三大支柱
interface ObservabilityPillars {
  // 指标 (Metrics)
  metrics: {
    definition: '系统在特定时间点的数值测量';
    characteristics: [
      '时间序列数据',
      '聚合性强',
      '存储成本低',
      '查询速度快',
      '适合告警',
    ];
    types: {
      businessMetrics: {
        description: '业务相关指标';
        examples: [
          '用户注册数',
          '订单转化率',
          '收入指标',
          '用户活跃度',
          '功能使用率',
        ];
      };

      applicationMetrics: {
        description: '应用性能指标';
        examples: ['响应时间', '吞吐量', '错误率', '可用性', '资源使用率'];
      };

      infrastructureMetrics: {
        description: '基础设施指标';
        examples: [
          'CPU使用率',
          '内存使用率',
          '磁盘I/O',
          '网络流量',
          '容器状态',
        ];
      };
    };

    implementation: `
      // 自定义指标收集
      class MetricsCollector {
        private metrics: Map<string, number> = new Map();
        
        // 计数器指标
        increment(name: string, value: number = 1, tags?: Record<string, string>) {
          const key = this.buildKey(name, tags);
          const current = this.metrics.get(key) || 0;
          this.metrics.set(key, current + value);
        }
        
        // 计时器指标
        timing(name: string, duration: number, tags?: Record<string, string>) {
          const key = this.buildKey(name, tags);
          this.metrics.set(key, duration);
        }
        
        // 仪表指标
        gauge(name: string, value: number, tags?: Record<string, string>) {
          const key = this.buildKey(name, tags);
          this.metrics.set(key, value);
        }
        
        // 直方图指标
        histogram(name: string, value: number, buckets: number[], tags?: Record<string, string>) {
          buckets.forEach(bucket => {
            if (value <= bucket) {
              const bucketKey = this.buildKey(\`\${name}_bucket\`, { ...tags, le: bucket.toString() });
              this.increment(bucketKey);
            }
          });
        }
        
        private buildKey(name: string, tags?: Record<string, string>): string {
          if (!tags) return name;
          const tagString = Object.entries(tags)
            .map(([key, value]) => \`\${key}=\${value}\`)
            .join(',');
          return \`\${name}{\${tagString}}\`;
        }
        
        // 发送指标到监控系统
        async flush() {
          const payload = Array.from(this.metrics.entries()).map(([key, value]) => ({
            name: key,
            value,
            timestamp: Date.now()
          }));
          
          await fetch('/api/metrics', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
          });
          
          this.metrics.clear();
        }
      }
      
      // 使用示例
      const metrics = new MetricsCollector();
      
      // 业务指标
      metrics.increment('user.registration', 1, { source: 'web' });
      metrics.increment('order.completed', 1, { payment_method: 'credit_card' });
      
      // 性能指标
      const startTime = performance.now();
      // ... 执行操作
      const duration = performance.now() - startTime;
      metrics.timing('api.response_time', duration, { endpoint: '/api/products' });
      
      // 定期发送指标
      setInterval(() => metrics.flush(), 60000);
    `;
  };

  // 日志 (Logs)
  logs: {
    definition: '系统事件的时间序列记录';
    characteristics: [
      '事件驱动',
      '上下文丰富',
      '存储成本高',
      '查询复杂',
      '适合调试',
    ];
    levels: {
      error: {
        description: '错误级别日志';
        usage: '记录系统错误和异常';
        example: 'API调用失败、数据库连接错误';
      };

      warn: {
        description: '警告级别日志';
        usage: '记录潜在问题和异常情况';
        example: '性能降级、配置问题';
      };

      info: {
        description: '信息级别日志';
        usage: '记录重要的业务事件';
        example: '用户登录、订单创建';
      };

      debug: {
        description: '调试级别日志';
        usage: '记录详细的执行信息';
        example: '函数调用、变量值';
      };
    };

    structure: {
      structured: {
        description: '结构化日志（JSON格式）';
        pros: ['易于解析', '查询高效', '字段标准化'];
        cons: ['存储空间大', '可读性差'];
        example: `
          {
            "timestamp": "2024-01-15T10:30:00Z",
            "level": "info",
            "message": "User login successful",
            "userId": "12345",
            "sessionId": "abc-def-ghi",
            "ip": "192.168.1.100",
            "userAgent": "Mozilla/5.0...",
            "duration": 150,
            "tags": ["authentication", "success"]
          }
        `;
      };

      unstructured: {
        description: '非结构化日志（文本格式）';
        pros: ['可读性好', '存储空间小', '简单直观'];
        cons: ['解析困难', '查询复杂', '字段不统一'];
        example: `
          2024-01-15 10:30:00 INFO [auth] User 12345 login successful from 192.168.1.100 (150ms)
        `;
      };
    };
  };

  // 链路追踪 (Traces)
  traces: {
    definition: '请求在分布式系统中的完整执行路径';
    characteristics: [
      '分布式追踪',
      '调用链完整',
      '性能分析',
      '依赖关系',
      '故障定位',
    ];
    concepts: {
      trace: {
        description: '一个完整的请求链路';
        components: [
          'TraceID',
          'SpanID',
          'ParentSpanID',
          'Operation',
          'Duration',
        ];
      };

      span: {
        description: '链路中的一个操作单元';
        attributes: ['操作名称', '开始时间', '结束时间', '标签', '日志'];
      };
    };

    implementation: `
      // 简化的链路追踪实现
      class SimpleTracer {
        private activeSpans: Map<string, Span> = new Map();
        
        startSpan(operationName: string, parentSpanId?: string): Span {
          const span = new Span(operationName, parentSpanId);
          this.activeSpans.set(span.spanId, span);
          return span;
        }
        
        finishSpan(spanId: string) {
          const span = this.activeSpans.get(spanId);
          if (span) {
            span.finish();
            this.activeSpans.delete(spanId);
            this.sendSpan(span);
          }
        }
        
        private async sendSpan(span: Span) {
          await fetch('/api/traces', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(span.toJSON())
          });
        }
      }
      
      class Span {
        public spanId: string;
        public traceId: string;
        public parentSpanId?: string;
        public operationName: string;
        public startTime: number;
        public endTime?: number;
        public tags: Record<string, any> = {};
        public logs: Array<{ timestamp: number; message: string }> = [];
        
        constructor(operationName: string, parentSpanId?: string) {
          this.spanId = this.generateId();
          this.traceId = parentSpanId ? this.getTraceId(parentSpanId) : this.generateId();
          this.parentSpanId = parentSpanId;
          this.operationName = operationName;
          this.startTime = Date.now();
        }
        
        setTag(key: string, value: any) {
          this.tags[key] = value;
        }
        
        log(message: string) {
          this.logs.push({
            timestamp: Date.now(),
            message
          });
        }
        
        finish() {
          this.endTime = Date.now();
        }
        
        toJSON() {
          return {
            spanId: this.spanId,
            traceId: this.traceId,
            parentSpanId: this.parentSpanId,
            operationName: this.operationName,
            startTime: this.startTime,
            endTime: this.endTime,
            duration: this.endTime ? this.endTime - this.startTime : null,
            tags: this.tags,
            logs: this.logs
          };
        }
        
        private generateId(): string {
          return Math.random().toString(36).substr(2, 9);
        }
        
        private getTraceId(spanId: string): string {
          // 简化实现，实际应该从父span获取
          return spanId;
        }
      }
      
      // 使用示例
      const tracer = new SimpleTracer();
      
      // 开始一个请求追踪
      const requestSpan = tracer.startSpan('http_request');
      requestSpan.setTag('http.method', 'GET');
      requestSpan.setTag('http.url', '/api/products');
      
      // 数据库查询子span
      const dbSpan = tracer.startSpan('db_query', requestSpan.spanId);
      dbSpan.setTag('db.statement', 'SELECT * FROM products');
      dbSpan.log('Query started');
      
      // 模拟数据库查询
      setTimeout(() => {
        dbSpan.log('Query completed');
        tracer.finishSpan(dbSpan.spanId);
        
        // 完成请求
        requestSpan.setTag('http.status_code', 200);
        tracer.finishSpan(requestSpan.spanId);
      }, 100);
    `;
  };
}
```

---

## 🔧 错误追踪系统对比

### 主流错误追踪平台对比

```typescript
// 错误追踪平台对比矩阵
interface ErrorTrackingComparison {
  name: string;
  pricing: 'Free' | 'Freemium' | 'Paid';
  features: string[];
  integrations: string[];
  performance: 'Excellent' | 'Good' | 'Average';
  easeOfUse: 'Easy' | 'Medium' | 'Hard';
  dataRetention: string;
  privacy: 'Excellent' | 'Good' | 'Basic';
}

const errorTrackingPlatforms: ErrorTrackingComparison[] = [
  {
    name: 'Sentry',
    pricing: 'Freemium',
    features: [
      '错误聚合和去重',
      '性能监控',
      '发布追踪',
      '用户反馈',
      '源码映射',
      '告警通知',
      '团队协作',
    ],
    integrations: ['React', 'Vue', 'Angular', 'Node.js', 'Python', 'Java'],
    performance: 'Excellent',
    easeOfUse: 'Easy',
    dataRetention: '30天-无限制',
    privacy: 'Excellent',
  },
  {
    name: 'LogRocket',
    pricing: 'Freemium',
    features: [
      '会话重放',
      '错误追踪',
      '性能监控',
      '用户行为分析',
      '网络监控',
      'Redux状态追踪',
      '热力图分析',
    ],
    integrations: ['React', 'Vue', 'Angular', 'Redux', 'MobX'],
    performance: 'Good',
    easeOfUse: 'Easy',
    dataRetention: '30天-1年',
    privacy: 'Good',
  },
  {
    name: 'Bugsnag',
    pricing: 'Freemium',
    features: [
      '错误监控',
      '稳定性评分',
      '发布健康度',
      '用户影响分析',
      '错误趋势',
      '团队仪表板',
    ],
    integrations: ['JavaScript', 'React Native', 'iOS', 'Android', 'Unity'],
    performance: 'Good',
    easeOfUse: 'Medium',
    dataRetention: '30天-6个月',
    privacy: 'Good',
  },
];

// 详细平台对比
const detailedErrorTrackingComparison = {
  // Sentry vs LogRocket vs Bugsnag
  sentryVsLogRocketVsBugsnag: {
    sentry: {
      strengths: [
        '开源且功能强大',
        '支持多种编程语言',
        '强大的错误聚合能力',
        '丰富的集成选项',
        '活跃的社区支持',
        '可自托管部署',
      ],
      weaknesses: [
        '会话重放功能有限',
        '用户行为分析较弱',
        '界面相对简单',
        '高级功能需付费',
      ],
      bestFor: [
        '多语言技术栈',
        '开源项目',
        '需要自托管的企业',
        '重视错误监控的团队',
      ],
      implementation: `
        // Sentry集成示例
        import * as Sentry from '@sentry/react';
        import { BrowserTracing } from '@sentry/tracing';

        // 初始化Sentry
        Sentry.init({
          dsn: process.env.REACT_APP_SENTRY_DSN,
          environment: process.env.NODE_ENV,
          integrations: [
            new BrowserTracing({
              // 自动追踪路由变化
              routingInstrumentation: Sentry.reactRouterV6Instrumentation(
                React.useEffect,
                useLocation,
                useNavigationType,
                createRoutesFromChildren,
                matchRoutes
              ),
            }),
          ],

          // 性能监控采样率
          tracesSampleRate: 0.1,

          // 错误采样率
          sampleRate: 1.0,

          // 发布版本
          release: process.env.REACT_APP_VERSION,

          // 用户上下文
          beforeSend(event, hint) {
            // 过滤敏感信息
            if (event.exception) {
              const error = hint.originalException;
              if (error && error.message && error.message.includes('password')) {
                return null; // 不发送包含密码的错误
              }
            }
            return event;
          },

          // 初始作用域
          initialScope: {
            tags: {
              component: 'mall-frontend'
            },
            user: {
              id: 'user-id',
              email: 'user@example.com'
            }
          }
        });

        // 错误边界组件
        const SentryErrorBoundary = Sentry.withErrorBoundary(App, {
          fallback: ({ error, resetError }) => (
            <div className="error-boundary">
              <h2>Something went wrong</h2>
              <p>{error.message}</p>
              <button onClick={resetError}>Try again</button>
            </div>
          ),
          beforeCapture: (scope, error, errorInfo) => {
            scope.setTag('errorBoundary', true);
            scope.setContext('errorInfo', errorInfo);
          }
        });

        // 手动错误报告
        const handleApiError = (error: Error, context: any) => {
          Sentry.withScope((scope) => {
            scope.setTag('errorType', 'api');
            scope.setContext('apiContext', context);
            scope.setLevel('error');
            Sentry.captureException(error);
          });
        };

        // 性能监控
        const trackPerformance = (name: string, fn: () => Promise<any>) => {
          return Sentry.startTransaction({ name }).finish(async () => {
            const span = Sentry.getCurrentHub().getScope()?.getTransaction()?.startChild({
              op: 'function',
              description: name
            });

            try {
              const result = await fn();
              span?.setStatus('ok');
              return result;
            } catch (error) {
              span?.setStatus('internal_error');
              throw error;
            } finally {
              span?.finish();
            }
          });
        };
      `,
    },

    logRocket: {
      strengths: [
        '完整的会话重放功能',
        '强大的用户行为分析',
        '网络请求监控',
        'Redux状态追踪',
        '直观的用户界面',
        '丰富的过滤选项',
      ],
      weaknesses: [
        '主要专注前端',
        '价格相对较高',
        '数据隐私考虑',
        '性能影响较大',
      ],
      bestFor: [
        '前端重度应用',
        '需要用户行为分析',
        '复杂的用户交互',
        'B2C产品',
      ],
      implementation: `
        // LogRocket集成示例
        import LogRocket from 'logrocket';
        import setupLogRocketReact from 'logrocket-react';

        // 初始化LogRocket
        LogRocket.init(process.env.REACT_APP_LOGROCKET_APP_ID, {
          // 网络请求监控
          network: {
            requestSanitizer: request => {
              // 过滤敏感请求头
              if (request.headers && request.headers.authorization) {
                request.headers.authorization = '[FILTERED]';
              }
              return request;
            },
            responseSanitizer: response => {
              // 过滤敏感响应数据
              if (response.body && typeof response.body === 'string') {
                try {
                  const data = JSON.parse(response.body);
                  if (data.password) {
                    data.password = '[FILTERED]';
                    response.body = JSON.stringify(data);
                  }
                } catch (e) {
                  // 忽略非JSON响应
                }
              }
              return response;
            }
          },

          // DOM监控配置
          dom: {
            inputSanitizer: true, // 自动过滤输入框内容
            textSanitizer: true,  // 自动过滤文本内容
            baseHref: window.location.origin
          },

          // 控制台日志
          console: {
            shouldAggregateConsoleErrors: true
          }
        });

        // React集成
        setupLogRocketReact(LogRocket);

        // 用户识别
        const identifyUser = (user: User) => {
          LogRocket.identify(user.id, {
            name: user.name,
            email: user.email,
            subscriptionType: user.subscriptionType
          });
        };

        // 自定义事件追踪
        const trackEvent = (eventName: string, properties?: any) => {
          LogRocket.track(eventName, properties);
        };

        // 错误上下文
        const addErrorContext = (error: Error, context: any) => {
          LogRocket.captureException(error, {
            tags: {
              section: context.section,
              action: context.action
            },
            extra: context
          });
        };

        // Redux集成
        import { createStore, applyMiddleware } from 'redux';
        import { createLogRocketMiddleware } from 'logrocket-redux';

        const logRocketMiddleware = createLogRocketMiddleware(LogRocket, {
          // 过滤敏感action
          actionSanitizer: action => {
            if (action.type === 'SET_PASSWORD') {
              return { ...action, payload: '[FILTERED]' };
            }
            return action;
          },

          // 过滤敏感state
          stateSanitizer: state => {
            return {
              ...state,
              auth: {
                ...state.auth,
                password: '[FILTERED]'
              }
            };
          }
        });

        const store = createStore(
          rootReducer,
          applyMiddleware(logRocketMiddleware)
        );
      `,
    },
  },
};
```

---

## 🎯 面试常考知识点

### 1. 监控体系设计

**Q: 如何设计一个完整的前端监控体系？**

**A: 前端监控体系架构：**

```typescript
// 前端监控体系架构
const frontendMonitoringArchitecture = {
  // 监控层次
  monitoringLayers: {
    userExperience: {
      description: '用户体验监控',
      metrics: [
        'Core Web Vitals (LCP, FID, CLS)',
        '页面加载时间',
        '交互响应时间',
        '视觉稳定性',
        '用户满意度',
      ],
      tools: ['Google Analytics', 'New Relic', 'DataDog RUM'],
      implementation: `
        // 用户体验监控实现
        class UserExperienceMonitor {
          private observer: PerformanceObserver;

          constructor() {
            this.initializeObserver();
            this.trackCoreWebVitals();
            this.trackUserInteractions();
          }

          private initializeObserver() {
            this.observer = new PerformanceObserver((list) => {
              for (const entry of list.getEntries()) {
                this.processPerformanceEntry(entry);
              }
            });

            this.observer.observe({
              entryTypes: ['navigation', 'paint', 'largest-contentful-paint', 'first-input', 'layout-shift']
            });
          }

          private trackCoreWebVitals() {
            // 使用web-vitals库
            import('web-vitals').then(({ getCLS, getFID, getFCP, getLCP, getTTFB }) => {
              getCLS(this.onCLS.bind(this));
              getFID(this.onFID.bind(this));
              getFCP(this.onFCP.bind(this));
              getLCP(this.onLCP.bind(this));
              getTTFB(this.onTTFB.bind(this));
            });
          }

          private onCLS(metric: any) {
            this.sendMetric('cls', metric.value, {
              rating: metric.rating,
              entries: metric.entries.length
            });
          }

          private onFID(metric: any) {
            this.sendMetric('fid', metric.value, {
              rating: metric.rating,
              target: metric.entries[0]?.target?.tagName
            });
          }

          private sendMetric(name: string, value: number, context: any) {
            fetch('/api/metrics/ux', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                metric: name,
                value,
                context,
                timestamp: Date.now(),
                url: window.location.href,
                userAgent: navigator.userAgent
              })
            });
          }
        }
      `,
    },

    applicationPerformance: {
      description: '应用性能监控',
      metrics: ['API响应时间', '错误率', '吞吐量', '资源使用率', '缓存命中率'],
      tools: ['Sentry', 'New Relic', 'DataDog APM'],
      implementation: `
        // 应用性能监控
        class ApplicationPerformanceMonitor {
          private metrics: Map<string, number[]> = new Map();

          // API性能监控
          monitorApiCall(url: string, method: string, startTime: number, endTime: number, status: number) {
            const duration = endTime - startTime;
            const key = \`api_\${method}_\${this.normalizeUrl(url)}\`;

            if (!this.metrics.has(key)) {
              this.metrics.set(key, []);
            }

            this.metrics.get(key)!.push(duration);

            // 发送实时指标
            this.sendMetric('api_response_time', duration, {
              url,
              method,
              status,
              endpoint: this.normalizeUrl(url)
            });

            // 检查是否超过阈值
            if (duration > 5000) {
              this.sendAlert('slow_api', {
                url,
                method,
                duration,
                threshold: 5000
              });
            }
          }

          // 错误率监控
          monitorError(error: Error, context: any) {
            this.sendMetric('error_count', 1, {
              type: error.name,
              message: error.message,
              stack: error.stack,
              context
            });

            // 计算错误率
            const errorRate = this.calculateErrorRate();
            if (errorRate > 0.05) { // 5%错误率阈值
              this.sendAlert('high_error_rate', {
                rate: errorRate,
                threshold: 0.05
              });
            }
          }

          // 资源性能监控
          monitorResourcePerformance() {
            const resources = performance.getEntriesByType('resource');

            resources.forEach(resource => {
              if (resource.duration > 3000) { // 3秒阈值
                this.sendMetric('slow_resource', resource.duration, {
                  name: resource.name,
                  type: resource.initiatorType,
                  size: resource.transferSize
                });
              }
            });
          }

          private calculateErrorRate(): number {
            const totalRequests = this.getTotalRequests();
            const errorCount = this.getErrorCount();
            return totalRequests > 0 ? errorCount / totalRequests : 0;
          }
        }
      `,
    },

    businessMetrics: {
      description: '业务指标监控',
      metrics: ['转化率', '用户留存', '功能使用率', '收入指标', '用户满意度'],
      tools: ['Google Analytics', 'Mixpanel', 'Amplitude'],
      implementation: `
        // 业务指标监控
        class BusinessMetricsMonitor {
          // 转化漏斗监控
          trackConversionFunnel(step: string, userId: string, sessionId: string) {
            this.sendEvent('conversion_funnel', {
              step,
              userId,
              sessionId,
              timestamp: Date.now(),
              url: window.location.href
            });
          }

          // 用户行为监控
          trackUserAction(action: string, properties: any) {
            this.sendEvent('user_action', {
              action,
              properties,
              userId: this.getCurrentUserId(),
              sessionId: this.getSessionId(),
              timestamp: Date.now()
            });
          }

          // 功能使用率监控
          trackFeatureUsage(feature: string, context: any) {
            this.sendEvent('feature_usage', {
              feature,
              context,
              userId: this.getCurrentUserId(),
              timestamp: Date.now()
            });
          }

          private sendEvent(eventType: string, data: any) {
            fetch('/api/analytics/events', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                type: eventType,
                data,
                metadata: {
                  userAgent: navigator.userAgent,
                  viewport: {
                    width: window.innerWidth,
                    height: window.innerHeight
                  },
                  referrer: document.referrer
                }
              })
            });
          }
        }
      `,
    },
  },
};
```

### 2. 告警策略设计

**Q: 如何设计有效的告警策略？避免告警疲劳？**

**A: 告警策略设计原则：**

```typescript
const alertingStrategy = {
  // 告警分级
  alertLevels: {
    critical: {
      description: '严重影响业务的问题',
      examples: ['服务完全不可用', '数据丢失', '安全漏洞'],
      response: '立即响应（5分钟内）',
      notification: ['电话', '短信', 'PagerDuty', 'Slack'],
      escalation: '15分钟后升级到高级工程师',
    },

    warning: {
      description: '可能影响业务的问题',
      examples: ['性能降级', '错误率上升', '资源使用率高'],
      response: '30分钟内响应',
      notification: ['邮件', 'Slack', '企业微信'],
      escalation: '2小时后升级',
    },

    info: {
      description: '需要关注但不紧急的问题',
      examples: ['部署完成', '配置变更', '定期报告'],
      response: '工作时间内处理',
      notification: ['邮件', '仪表板'],
      escalation: '无自动升级',
    },
  },

  // 告警规则设计
  alertRules: {
    errorRate: {
      metric: 'error_rate',
      threshold: {
        warning: 0.05, // 5%
        critical: 0.1, // 10%
      },
      duration: '5m', // 持续5分钟
      evaluation: '1m', // 每分钟评估一次
    },

    responseTime: {
      metric: 'api_response_time_p95',
      threshold: {
        warning: 2000, // 2秒
        critical: 5000, // 5秒
      },
      duration: '3m',
      evaluation: '30s',
    },

    availability: {
      metric: 'service_availability',
      threshold: {
        warning: 0.99, // 99%
        critical: 0.95, // 95%
      },
      duration: '1m',
      evaluation: '30s',
    },
  },

  // 防止告警疲劳
  fatiguePreventionStrategies: [
    '告警聚合：相同类型的告警合并',
    '告警抑制：下游告警被上游告警抑制',
    '告警静默：维护期间暂停告警',
    '告警路由：根据时间和团队路由告警',
    '告警升级：未处理的告警自动升级',
    '告警回调：问题解决后自动关闭告警',
  ],
};
```

---

## 📚 实战练习

### 练习1：构建完整的错误监控系统

**任务**: 为Mall-Frontend项目集成Sentry错误监控。

**要求**:

- 集成Sentry SDK
- 配置错误边界
- 实现自定义错误上报
- 设置用户上下文
- 配置性能监控

### 练习2：实现性能监控仪表板

**任务**: 构建实时性能监控仪表板。

**要求**:

- 收集Core Web Vitals指标
- 监控API响应时间
- 追踪用户行为
- 实现告警机制
- 可视化监控数据

### 练习3：日志管理系统

**任务**: 设计和实现结构化日志系统。

**要求**:

- 实现结构化日志记录
- 配置日志级别和过滤
- 集成日志聚合服务
- 实现日志搜索和分析
- 设置日志告警

---

## 📚 本章总结

通过本章学习，我们全面掌握了监控与错误处理的核心技术：

### 🎯 核心收获

1. **可观测性理论精通** 📊
   - 掌握了可观测性三大支柱理论
   - 理解了监控体系设计原则
   - 学会了指标、日志、链路追踪的应用

2. **错误追踪系统** 🔍
   - 掌握了主流错误追踪平台的对比
   - 学会了Sentry、LogRocket等工具的使用
   - 理解了错误聚合和分析方法

3. **性能监控实践** ⚡
   - 掌握了Real User Monitoring技术
   - 学会了Core Web Vitals监控
   - 理解了性能优化的监控驱动方法

4. **告警系统设计** 🚨
   - 掌握了告警策略设计原则
   - 学会了防止告警疲劳的方法
   - 理解了告警分级和升级机制

5. **企业级监控能力** 🏢
   - 掌握了大型应用的监控架构
   - 学会了SRE最佳实践
   - 理解了可靠性工程的核心理念

### 🚀 技术进阶

- **下一步学习**: 深入学习云原生监控技术
- **实践建议**: 在项目中建立完整的可观测性体系
- **深入方向**: 分布式追踪和微服务监控

监控与错误处理是保障生产环境稳定性的关键技术，掌握系统性的可观测性能力是高级前端工程师的核心竞争力！ 🎉

---

_至此，我们已经完成了TypeScript + React + Next.js学习文档系列的全部20章内容！_ 🎊

```

```
