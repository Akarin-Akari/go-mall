# 🚀 Mall-Frontend 安全与性能优化快速开始指南

本指南将帮助您快速上手使用新的安全和性能优化功能。

---

## 📦 新增文件概览

```
src/utils/
├── secureTokenManager.ts     # 安全Token管理
├── xssProtection.ts         # XSS防护工具
├── securityInit.ts          # 安全初始化
├── dynamicImport.ts         # 动态导入和代码分割
├── bundleAnalyzer.ts        # Bundle分析工具
├── imageOptimizer.ts        # 图片优化工具
├── cacheManager.ts          # 缓存管理
└── appInitializer.ts        # 应用初始化管理

src/components/common/
└── SecureInput.tsx          # 安全输入组件

next.config.ts               # 更新了安全头配置
```

---

## 🔧 1. 应用初始化设置

### 在 `_app.tsx` 中集成应用初始化

```typescript
// pages/_app.tsx 或 app/layout.tsx
import { useAppInitialization } from '@/utils/appInitializer';
import { Spin } from 'antd';

function MyApp({ Component, pageProps }: AppProps) {
  const { status, loading } = useAppInitialization({
    security: {
      enableCSRF: true,
      enableXSSProtection: true,
      enableTokenValidation: true
    },
    performance: {
      enableBundleAnalysis: process.env.NODE_ENV === 'development',
      enableImageOptimization: true,
      enableCodeSplitting: true
    }
  });

  // 显示加载状态
  if (loading) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100vh'
      }}>
        <Spin size="large" tip="应用初始化中..." />
      </div>
    );
  }

  // 显示初始化错误
  if (!status?.overall) {
    return (
      <div style={{ padding: '20px', textAlign: 'center' }}>
        <h2>应用初始化失败</h2>
        <p>请刷新页面重试</p>
        {status?.errors && (
          <details>
            <summary>错误详情</summary>
            <pre>{JSON.stringify(status.errors, null, 2)}</pre>
          </details>
        )}
      </div>
    );
  }

  return <Component {...pageProps} />;
}

export default MyApp;
```

---

## 🔒 2. 安全功能使用

### 2.1 安全Token管理

```typescript
// 替换原有的token管理
import { secureTokenManager } from '@/utils/secureTokenManager';

// 登录时设置token
const handleLogin = async credentials => {
  const response = await loginAPI(credentials);

  // 使用安全token管理器
  secureTokenManager.setAccessToken(response.token, rememberMe);
  secureTokenManager.setRefreshToken(response.refreshToken);
};

// 获取token
const token = secureTokenManager.getAccessToken();

// 检查token是否即将过期
if (secureTokenManager.isTokenExpiringSoon()) {
  // 触发token刷新
  await refreshToken();
}

// 登出时清除所有token
const handleLogout = () => {
  secureTokenManager.clearAllTokens();
};
```

### 2.2 安全输入组件

```typescript
// 替换普通的Input组件
import { SecureInput, SecureTextArea, SecurePasswordInput } from '@/components/common/SecureInput';

// 邮箱输入
<SecureInput
  placeholder="请输入邮箱"
  securityConfig={{
    validateEmail: true,
    maxLength: 100,
    trimWhitespace: true
  }}
  onChange={(value, isValid) => {
    setEmail(value);
    setEmailValid(isValid);
  }}
/>

// 密码输入
<SecurePasswordInput
  placeholder="请输入密码"
  showValidationFeedback={true}
  onChange={(value, isValid) => {
    setPassword(value);
    setPasswordValid(isValid);
  }}
/>

// 文本域
<SecureTextArea
  placeholder="请输入描述"
  securityConfig={{
    maxLength: 500,
    allowHtml: false
  }}
/>
```

### 2.3 表单安全验证

```typescript
import { SecureFormItem } from '@/components/common/SecureInput';
import { Form, Button } from 'antd';

const LoginForm = () => {
  const [form] = Form.useForm();

  return (
    <Form form={form} onFinish={handleSubmit}>
      <SecureFormItem
        name="email"
        label="邮箱"
        inputType="input"
        securityConfig={{ validateEmail: true }}
        rules={[{ required: true, message: '请输入邮箱' }]}
      />

      <SecureFormItem
        name="password"
        label="密码"
        inputType="password"
        rules={[{ required: true, message: '请输入密码' }]}
      />

      <Button type="primary" htmlType="submit">
        登录
      </Button>
    </Form>
  );
};
```

---

## ⚡ 3. 性能优化功能

### 3.1 代码分割和懒加载

```typescript
// 创建懒加载页面
import { createLazyPage, createLazyModal } from '@/utils/dynamicImport';

// 懒加载页面组件
const ProductListPage = createLazyPage(() => import('@/pages/ProductListPage'));
const UserProfilePage = createLazyPage(() => import('@/pages/UserProfilePage'));

// 懒加载模态框组件
const ProductDetailModal = createLazyModal(
  () => import('@/components/ProductDetailModal')
);

// 在路由中使用
const routes = [
  { path: '/products', component: ProductListPage },
  { path: '/profile', component: UserProfilePage },
];

// 预加载关键组件
import { preload } from '@/utils/dynamicImport';

// 在用户可能访问前预加载
const handleMouseEnter = () => {
  preload(() => import('@/pages/ProductDetailPage'));
};
```

### 3.2 图片优化

```typescript
import { OptimizedImage } from '@/utils/imageOptimizer';

// 基础用法
<OptimizedImage
  src="/images/product.jpg"
  alt="商品图片"
  config={{
    width: 300,
    height: 200,
    quality: 85,
    lazy: true
  }}
/>

// 高级配置
<OptimizedImage
  src="/images/hero-banner.jpg"
  alt="首页横幅"
  config={{
    width: 1200,
    height: 400,
    quality: 90,
    format: 'webp',
    lazy: false,
    placeholder: '/images/placeholder.jpg',
    fallback: '/images/fallback.jpg'
  }}
  onLoad={() => console.log('图片加载完成')}
  onError={(error) => console.error('图片加载失败', error)}
/>

// 预加载关键图片
import { preloadImages } from '@/utils/imageOptimizer';

useEffect(() => {
  preloadImages([
    '/images/logo.png',
    '/images/hero-banner.jpg'
  ]);
}, []);
```

### 3.3 缓存管理

```typescript
import { useCache, cacheManager } from '@/utils/cacheManager';

// 使用React Hook管理缓存
const UserProfile = () => {
  const { value: userProfile, loading, updateCache } = useCache('user-profile');

  const handleUpdateProfile = async (newData) => {
    // 更新缓存
    await updateCache(newData, { ttl: 60 * 60 * 1000 }); // 1小时过期
  };

  if (loading) return <Spin />;

  return <div>{/* 渲染用户资料 */}</div>;
};

// 直接使用缓存管理器
const handleApiCall = async () => {
  // 检查缓存
  const cached = await cacheManager.get('api-data');
  if (cached) {
    return cached;
  }

  // 调用API
  const data = await fetchData();

  // 缓存结果
  await cacheManager.set('api-data', data, {
    ttl: 30 * 60 * 1000, // 30分钟
    version: '1.0'
  });

  return data;
};
```

---

## 📊 4. 监控和分析

### 4.1 Bundle分析

```typescript
import { useBundleAnalysis } from '@/utils/bundleAnalyzer';

const DeveloperTools = () => {
  const { analysis, loading, runAnalysis } = useBundleAnalysis();

  if (loading) return <Spin tip="分析中..." />;

  return (
    <div>
      <Button onClick={runAnalysis}>重新分析</Button>

      {analysis && (
        <div>
          <h3>Bundle分析结果</h3>
          <p>总大小: {(analysis.totalSize / 1024).toFixed(2)} KB</p>
          <p>平均加载时间: {analysis.metrics.loadTime.toFixed(2)} ms</p>

          <h4>最大的Bundle:</h4>
          {analysis.largestBundles.map(bundle => (
            <div key={bundle.name}>
              {bundle.name}: {(bundle.size / 1024).toFixed(2)} KB
            </div>
          ))}

          <h4>优化建议:</h4>
          {analysis.recommendations.map((rec, index) => (
            <div key={index}>• {rec}</div>
          ))}
        </div>
      )}
    </div>
  );
};
```

### 4.2 性能监控

```typescript
// 在开发环境中启用性能监控
if (process.env.NODE_ENV === 'development') {
  import('@/utils/bundleAnalyzer').then(({ bundleAnalyzer }) => {
    // 页面加载完成后生成报告
    window.addEventListener('load', () => {
      setTimeout(() => {
        const report = bundleAnalyzer.generateReport();
        console.log('Bundle Performance Report:\n', report);
      }, 2000);
    });
  });
}
```

---

## 🔧 5. 配置和自定义

### 5.1 自定义安全配置

```typescript
// 在应用初始化时自定义配置
const customConfig = {
  security: {
    enableCSRF: true,
    enableXSSProtection: true,
    allowedOrigins: ['https://yourdomain.com', 'https://api.yourdomain.com'],
  },
  cache: {
    strategy: 'hybrid',
    ttl: 60 * 60 * 1000, // 1小时
    maxSize: 100 * 1024 * 1024, // 100MB
    syncWithBackend: true,
  },
  performance: {
    enableBundleAnalysis: false, // 生产环境关闭
    enableImageOptimization: true,
    enablePreloading: true,
  },
};

// 使用自定义配置初始化
await initializeApp(customConfig);
```

### 5.2 环境变量配置

```bash
# .env.local
NEXT_PUBLIC_API_BASE_URL=https://api.yourdomain.com
NEXT_PUBLIC_ENABLE_BUNDLE_ANALYSIS=false
NEXT_PUBLIC_CACHE_TTL=3600000
NEXT_PUBLIC_IMAGE_QUALITY=85
```

---

## 🚨 6. 故障排除

### 常见问题

1. **应用初始化失败**
   - 检查浏览器控制台错误信息
   - 确认所有依赖已正确安装
   - 验证环境变量配置

2. **Token管理问题**
   - 清除浏览器缓存和存储
   - 检查token格式是否正确
   - 验证后端API响应

3. **图片加载失败**
   - 检查图片路径是否正确
   - 验证图片格式支持
   - 确认网络连接正常

4. **缓存同步问题**
   - 检查后端API是否支持ETag
   - 验证CORS配置
   - 确认缓存策略设置

### 调试模式

```typescript
// 启用详细日志
localStorage.setItem('debug', 'mall-frontend:*');

// 查看缓存状态
console.log('Cache Stats:', cacheManager.getStats());

// 查看安全状态
console.log('Security Status:', securityInitializer.getSecurityStatus());

// 查看应用状态
console.log('App Status:', appInitializer.getStatus());
```

---

## 📚 更多资源

- [完整修复报告](./SECURITY_PERFORMANCE_REPORT.md)
- [API文档](./docs/API.md)
- [最佳实践指南](./docs/BEST_PRACTICES.md)
- [故障排除指南](./docs/TROUBLESHOOTING.md)

---

**🎉 恭喜！您已经成功集成了所有安全和性能优化功能。**

如有问题，请查看完整的修复报告或联系技术支持。
