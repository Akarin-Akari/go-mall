# 🔍 Mall-Frontend 技术诊断与修复指南

**诊断日期**：2025-11-01  
**项目**：mall-frontend（GoMall 电商平台前端）  
**调查范围**：代码架构、配置、功能完整性、技术债务  
**总体完成度**：~65%（以核心购买闭环权重计）

---

## 📋 项目概况

### 技术栈
| 组件 | 版本 | 说明 |
|------|------|------|
| React | 19.1.0 | 最新版本，需关注 antd 兼容性 |
| Next.js | 15.5.2 | App Router 架构 |
| TypeScript | 5.0+ | 类型安全 |
| Ant Design | 5.27.1 | UI 组件库 |
| Redux Toolkit | 2.8.2 | 状态管理 |
| TanStack Query | 5.85.5 | 数据获取 |
| Tailwind CSS | 4 | 样式框架 |

### 架构模式
- ✅ **Next.js App Router**：文件式路由，`src/app/` 目录结构
- ✅ **API 代理**：`/api/*` → `http://localhost:8081/api/*`
- ✅ **Redux 状态管理**：条件化 mock（dev）与真实 API（prod）
- ✅ **Token 管理**：Bearer 认证 + 自动刷新机制
- ⚠️ **CSP 安全头**：仅允许 localhost:8081，生产环境将拦截真实 API

### 当前完成度估算
```
认证系统：    66% ✅ (登录/注册 OK；忘记密码 ❌)
商品模块：    70% ⚠️ (数据 OK；交互落地缺失)
购物车：      80% ✅ (页面/切片 OK；详情页加购 ❌)
订单管理：    50% ❌ (列表 ❌；详情 ✅；创建 ✅)
支付流程：    10% 🔴 (常量有；路由缺失)
地址簿：      30% ❌ (使用本地 mock)
─────────────────────────
综合完成度：  ~65% ⚠️
```

---

## 🚨 问题清单（按优先级）

### 🔴 阻塞性问题（直接影响生产部署）

#### 1️⃣ 支付路由缺失 - 下单闭环中断
- **严重程度**：🔴 **CRITICAL**
- **影响范围**：结算页、订单详情、支付全流程
- **问题描述**：
  - `/payment`、`/payment/success`、`/payment/failed` 页面不存在
  - 结算与订单详情均跳转到 `/payment?order_id=xxx`，当前返回 404
  - 直接阻断"下单→支付→查看订单"的关键闭环

- **代码证据**：
  ```typescript
  // src/constants/index.ts - 常量已定义但页面缺失
  export const ROUTES = {
    PAYMENT: '/payment',
    PAYMENT_SUCCESS: '/payment/success',
    PAYMENT_FAILED: '/payment/failed',
  };
  
  // src/app/checkout/page.tsx (第 ~450 行)
  router.push(`/payment?order_id=${orderId}`); // 404
  
  // src/app/orders/[id]/page.tsx (第 ~200 行)
  router.push(`/payment?order_id=${currentOrder.id}`); // 404
  ```

- **修复方案**：
  1. 创建 `src/app/payment/page.tsx`（支付页）
  2. 创建 `src/app/payment/success/page.tsx`（成功页）
  3. 创建 `src/app/payment/failed/page.tsx`（失败页）
  4. 接入 `paymentAPI.create()` 与 `paymentAPI.query()`

- **预估工作量**：1.5–2.5 人日 | 复杂度：**M**

---

#### 2️⃣ CSP connect-src 限制 - 生产环境 API 全部拦截
- **严重程度**：🔴 **CRITICAL**
- **影响范围**：所有 API 请求（生产环境）
- **问题描述**：
  - `next.config.ts` 的 CSP 仅允许 `http://localhost:8081`
  - 生产环境若变更 baseURL，浏览器将直接拦截所有跨域请求
  - 导致生产环境无法调用任何后端 API

- **代码证据**：
  ```typescript
  // next.config.ts (第 ~80 行)
  "connect-src 'self' http://localhost:8081 ws://localhost:3000 ws://localhost:3001",
  ```

- **修复方案**：
  ```typescript
  // 改为环境感知配置
  const connectSrc = process.env.NODE_ENV === 'production'
    ? `'self' ${process.env.NEXT_PUBLIC_API_BASE_URL || 'https://api.example.com'}`
    : "'self' http://localhost:8081 ws://localhost:3000 ws://localhost:3001";
  ```

- **预估工作量**：0.5 人日 | 复杂度：**S**

---

#### 3️⃣ 账号恢复与注册校验 - 伪 API 无法上线
- **严重程度**：🔴 **CRITICAL**
- **影响范围**：忘记密码页、注册页实时校验
- **问题描述**：
  - 忘记密码：使用 `Promise + fetch.catch(() => ({ ok: true }))`，总是返回成功
  - 注册唯一性校验：硬编码黑名单 `['admin', 'test', 'user']`，无法真实验证
  - 路径不符合后端规范（缺少 `/v1`）

- **代码证据**：
  ```typescript
  // src/app/forgot-password/page.tsx (第 66–100 行)
  const handleSendCode = async (values) => {
    await new Promise(r => setTimeout(r, 1500));
    const response = await fetch('/api/auth/forgot-password', {
      method: 'POST',
      body: JSON.stringify({ email: values.email }),
    }).catch(() => ({ ok: true })); // ❌ 总是成功
  };
  
  // src/app/register/page.tsx (第 96–132 行)
  const validateUsername = async (username: string) => {
    await new Promise(r => setTimeout(r, 800));
    const isAvailable = !['admin', 'test', 'user'].includes(username);
    // ❌ 硬编码黑名单，无法真实验证
  };
  ```

- **修复方案**：
  1. 忘记密码改用 `authAPI.sendResetCode()` 与 `authAPI.resetPassword()`
  2. 注册校验改用 `authAPI.checkUsernameAvailable()` 与 `authAPI.checkEmailAvailable()`
  3. 统一使用 `request.ts` 的 axios 实例（自动处理错误、token 等）

- **预估工作量**：1–1.5 人日 | 复杂度：**M**

---

### 🟠 重要问题（影响功能完整性）

#### 4️⃣ 订单列表页使用硬编码 mock - 无法查看真实订单
- **严重程度**：🟠 **HIGH**
- **影响范围**：`/orders` 页面
- **问题描述**：
  - 使用本地 `mockOrders` 数组，未接入 Redux `orderSlice.fetchOrdersAsync`
  - 无分页、无筛选、无真实数据

- **代码证据**：
  ```typescript
  // src/app/orders/page.tsx (第 68–117 行)
  const mockOrders: Order[] = [
    { id: 1, order_no: 'ORD20240101001', ... },
    // ... 硬编码数据
  ];
  
  const loadOrders = useCallback(async () => {
    setLoading(true);
    try {
      setOrders(mockOrders); // ❌ 仅加载 mock
    } finally { setLoading(false); }
  }, []);
  ```

- **修复方案**：
  ```typescript
  // 改为调用真实 API
  const loadOrders = useCallback(async () => {
    const result = await dispatch(fetchOrdersAsync({
      page: currentPage,
      pageSize: 10,
    }));
    if (result.meta.requestStatus === 'fulfilled') {
      setOrders(result.payload.data);
    }
  }, [dispatch, currentPage]);
  ```

- **预估工作量**：0.8–1.2 人日 | 复杂度：**M**

---

#### 5️⃣ 商品列表/详情页交互未实现 - 加购/收藏无效
- **严重程度**：🟠 **HIGH**
- **影响范围**：`/products`、`/products/[id]` 页面
- **问题描述**：
  - 加入购物车：仅 `console.log`，未调用 Redux `addToCartAsync`
  - 收藏：仅 `console.log`，无 API 接入
  - 立即购买：使用 `setTimeout` 模拟，未真实创建订单

- **代码证据**：
  ```typescript
  // src/app/products/page.tsx (第 255–260 行)
  const handleAddToCart = useCallback(async (product: Product) => {
    console.log('Add to cart:', product); // ❌ 仅日志
    await new Promise(r => setTimeout(r, 500));
  }, []);
  
  // src/app/products/[id]/page.tsx (第 143–173 行)
  const handleAddToCart = useCallback(async () => {
    await new Promise(r => setTimeout(r, 1000)); // ❌ 模拟延迟
    message.success('已添加到购物车');
  }, []);
  ```

- **修复方案**：
  ```typescript
  // 改为调用 Redux
  const handleAddToCart = useCallback(async (product: Product) => {
    const result = await dispatch(addToCartAsync({
      product_id: product.id,
      quantity: 1,
    }));
    if (result.meta.requestStatus === 'fulfilled') {
      message.success('已添加到购物车');
    }
  }, [dispatch]);
  ```

- **预估工作量**：1–1.5 人日 | 复杂度：**M**

---

#### 6️⃣ 环境变量与端口不一致 - 开发文档误导
- **严重程度**：🟠 **MEDIUM**
- **影响范围**：开发环境一致性、新人上手
- **问题描述**：
  - `.env.example` 写 `NEXT_PUBLIC_API_BASE_URL=http://localhost:8080`
  - 实际代码使用 `8081`
  - 缺少 `.env.local` 与 `.env.production` 文件

- **代码证据**：
  ```ini
  # .env.example
  NEXT_PUBLIC_API_BASE_URL=http://localhost:8080 # ❌ 错误端口
  
  # src/utils/request.ts (第 ~20 行)
  const baseURL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8081'; # ✅ 正确端口
  ```

- **修复方案**：
  1. 修正 `.env.example` 为 `8081`
  2. 创建 `.env.local` 与 `.env.production`
  3. 更新 README 中的环境配置说明

- **预估工作量**：0.3 人日 | 复杂度：**S**

---

### 🟡 次要问题（代码质量与一致性）

#### 7️⃣ package.json 脚本键重复
- **问题**：`test:performance` 定义两次，后者覆盖前者
- **位置**：`package.json` 第 ~45–50 行
- **修复**：删除重复项，保留一个

#### 8️⃣ 结算页地址使用本地 mock
- **问题**：`src/app/checkout/page.tsx` 第 47–68 行使用 `mockAddresses`
- **修复**：接入真实地址簿 API（或占位 Redux Slice）

#### 9️⃣ 控制台日志未清理
- **位置**：
  - `src/utils/request.ts` 第 120、138、183 行
  - `src/app/login/page.tsx` 第 194 行
  - `src/app/register/page.tsx` 第 260 行
- **修复**：删除或改为 `process.env.NODE_ENV === 'development'` 条件输出

---

## 📊 功能完整性评估

### Mock 数据使用情况统计

| 模块 | 页面 | 状态 | 说明 |
|------|------|------|------|
| 首页 | `/` | ✅ | Redux 条件化（dev=mock, prod=real） |
| 商品列表 | `/products` | ⚠️ | 数据 OK；交互未实现 |
| 商品详情 | `/products/[id]` | ⚠️ | 数据 OK；加购/立即购买未实现 |
| 购物车 | `/cart` | ✅ | 完整 Redux 流程 |
| 结算 | `/checkout` | ⚠️ | 订单创建 OK；地址使用 mock |
| 订单列表 | `/orders` | ❌ | 硬编码 mock，未接真实 API |
| 订单详情 | `/orders/[id]` | ✅ | 真实 API |
| 登录 | `/login` | ✅ | 真实 API |
| 注册 | `/register` | ⚠️ | 注册 OK；唯一性校验 mock |
| 忘记密码 | `/forgot-password` | ❌ | 全链路 mock/伪 API |
| 支付 | `/payment*` | ❌ | 路由缺失 |

### 按钮点击无响应的具体位置

| 页面 | 按钮 | 状态 | 行号 |
|------|------|------|------|
| `/products` | 加入购物车 | ❌ | 255–260 |
| `/products` | 收藏 | ❌ | 263–268 |
| `/products/[id]` | 加入购物车 | ❌ | 143–173 |
| `/products/[id]` | 立即购买 | ❌ | 176–207 |
| `/forgot-password` | 发送验证码 | ❌ | 66–100 |
| `/forgot-password` | 重置密码 | ❌ | 123–157 |
| `/checkout` | 新增地址 | ⚠️ | 仅内存，不持久化 |

### 前后端接口对接完成度

```
认证系统：
  ✅ 登录 (authAPI.login)
  ✅ 注册 (authAPI.register)
  ❌ 忘记密码 (authAPI.resetPassword)
  ❌ 用户名可用性检查 (authAPI.checkUsernameAvailable)
  ❌ 邮箱可用性检查 (authAPI.checkEmailAvailable)

商品系统：
  ✅ 获取商品列表 (productAPI.getProducts)
  ✅ 获取商品详情 (productAPI.getProductDetail)
  ❌ 收藏商品 (favoriteAPI.addFavorite)
  ❌ 取消收藏 (favoriteAPI.removeFavorite)

购物车系统：
  ✅ 获取购物车 (cartAPI.getCart)
  ✅ 添加到购物车 (cartAPI.addItem)
  ✅ 更新购物车 (cartAPI.updateItem)
  ✅ 删除购物车项 (cartAPI.removeItem)

订单系统：
  ✅ 创建订单 (orderAPI.createOrder)
  ✅ 获取订单详情 (orderAPI.getOrderDetail)
  ❌ 获取订单列表 (orderAPI.getOrders) - 页面未接入
  ✅ 取消订单 (orderAPI.cancelOrder)

支付系统：
  ❌ 创建支付 (paymentAPI.create)
  ❌ 查询支付状态 (paymentAPI.query)

地址系统：
  ❌ 获取地址列表 (addressAPI.getAddresses)
  ❌ 创建地址 (addressAPI.createAddress)
  ❌ 更新地址 (addressAPI.updateAddress)
  ❌ 删除地址 (addressAPI.deleteAddress)

综合完成度：~65%
```

---

## 🔧 技术债务分析

### 配置问题
- ❌ CSP connect-src 硬编码 localhost:8081（生产拦截）
- ❌ 环境变量文件缺失（.env.local、.env.production）
- ❌ .env.example 端口错误（8080 vs 8081）
- ⚠️ package.json 脚本键重复

### 架构问题
- ❌ 支付路由完全缺失（3 个页面）
- ❌ 地址簿 API 未接入
- ❌ 收藏功能无 Redux Slice
- ⚠️ 忘记密码流程使用伪 API

### 代码质量问题
- ❌ 订单列表硬编码 mock（7 条记录）
- ❌ 商品页面加购/收藏仅 console.log
- ❌ 注册校验硬编码黑名单
- ⚠️ 控制台日志未清理（5 处）
- ⚠️ 部分页面 dev/prod 行为不一致

---

## 🎯 修复行动计划（优先级排序）

### 第一阶段：关键闭环修复（1 周）

#### 任务 1️⃣：修复支付完整流程
- **优先级**：🔴 **P0**
- **工作量**：1.5–2.5 人日
- **复杂度**：**M**（中等）
- **依赖**：无
- **任务清单**：
  - [ ] 创建 `src/app/payment/page.tsx`（支付页骨架）
  - [ ] 创建 `src/app/payment/success/page.tsx`（成功页）
  - [ ] 创建 `src/app/payment/failed/page.tsx`（失败页）
  - [ ] 接入 `paymentAPI.create()` 与 `paymentAPI.query()`
  - [ ] 结算页与订单详情统一跳转 `ROUTES.PAYMENT`
  - [ ] 支付页实现轮询机制（查询订单支付状态）

- **修复前后对比**：
  ```typescript
  // ❌ 修复前：404
  router.push(`/payment?order_id=${orderId}`);
  
  // ✅ 修复后：正常跳转到支付页
  router.push(`${ROUTES.PAYMENT}?order_id=${orderId}`);
  // 支付页接收 order_id，调用 paymentAPI.query() 轮询状态
  ```

---

#### 任务 2️⃣：订单列表页接入真实 API
- **优先级**：🔴 **P0**
- **工作量**：0.8–1.2 人日
- **复杂度**：**M**
- **依赖**：无
- **任务清单**：
  - [ ] 删除 `mockOrders` 数组
  - [ ] 改用 `dispatch(fetchOrdersAsync())`
  - [ ] 补齐分页、loading/error 展示
  - [ ] 与订单详情保持数据模型一致

- **修复前后对比**：
  ```typescript
  // ❌ 修复前
  const mockOrders = [...];
  setOrders(mockOrders);
  
  // ✅ 修复后
  const result = await dispatch(fetchOrdersAsync({
    page: currentPage,
    pageSize: 10,
  }));
  if (result.meta.requestStatus === 'fulfilled') {
    setOrders(result.payload.data);
  }
  ```

---

#### 任务 3️⃣：商品列表/详情页事件处理落地
- **优先级**：🔴 **P0**
- **工作量**：1–1.5 人日
- **复杂度**：**M**
- **依赖**：无
- **任务清单**：
  - [ ] `/products` 页面 `handleAddToCart` 改为 `dispatch(addToCartAsync)`
  - [ ] `/products` 页面 `handleToggleFavorite` 接入 favorites API（或占位 Slice）
  - [ ] `/products/[id]` 页面 `handleAddToCart` 改为 `dispatch(addToCartAsync)`
  - [ ] `/products/[id]` 页面 `handleBuyNow` 最小化实现（加入选中项后跳转结算）

- **修复前后对比**：
  ```typescript
  // ❌ 修复前
  const handleAddToCart = async (product) => {
    console.log('Add to cart:', product);
    await new Promise(r => setTimeout(r, 500));
  };
  
  // ✅ 修复后
  const handleAddToCart = async (product) => {
    const result = await dispatch(addToCartAsync({
      product_id: product.id,
      quantity: 1,
    }));
    if (result.meta.requestStatus === 'fulfilled') {
      message.success('已添加到购物车');
    } else {
      message.error('添加失败，请重试');
    }
  };
  ```

---

### 第二阶段：认证与配置修复（3–4 天）

#### 任务 4️⃣：认证相关改造
- **优先级**：🟠 **P1**
- **工作量**：1–1.5 人日
- **复杂度**：**M**
- **依赖**：无
- **任务清单**：
  - [ ] 注册页用户名/邮箱可用性校验改为 `authAPI` 接口
  - [ ] 忘记密码页改为使用 `authAPI.sendResetCode()` 与 `authAPI.resetPassword()`
  - [ ] 替换裸 `fetch` 为 `request.ts` 的 axios 实例
  - [ ] 统一 API 路径为 `/api/v1/auth/*` 规范

- **修复前后对比**：
  ```typescript
  // ❌ 修复前：伪 API
  const response = await fetch('/api/auth/forgot-password')
    .catch(() => ({ ok: true }));
  
  // ✅ 修复后：真实 API
  const response = await authAPI.sendResetCode({ email });
  if (response.code === 0) {
    message.success('验证码已发送');
  }
  ```

---

#### 任务 5️⃣：环境与安全配置统一
- **优先级**：🟠 **P1**
- **工作量**：0.5 人日
- **复杂度**：**S**（小）
- **依赖**：无
- **任务清单**：
  - [ ] 创建 `.env.local` 与 `.env.production`
  - [ ] 修正 `.env.example` 端口为 `8081`
  - [ ] 修改 `next.config.ts` CSP connect-src 为环境感知配置
  - [ ] 清理 `package.json` 重复脚本键
  - [ ] 删除控制台日志（5 处）

- **修复前后对比**：
  ```typescript
  // ❌ 修复前：硬编码
  "connect-src 'self' http://localhost:8081 ws://localhost:3000",
  
  // ✅ 修复后：环境感知
  const connectSrc = process.env.NODE_ENV === 'production'
    ? `'self' ${process.env.NEXT_PUBLIC_API_BASE_URL}`
    : "'self' http://localhost:8081 ws://localhost:3000";
  ```

---

### 第三阶段：增强功能（可选，后续迭代）

- 地址簿 API 接入（1 人日）
- 收藏功能完整实现（0.8 人日）
- 支付成功后回写订单状态（0.5 人日）
- 错误/结果页实现（0.5 人日）

---

## 📋 附录

### A. 关键代码文件清单

| 文件 | 行数 | 说明 |
|------|------|------|
| `src/utils/request.ts` | 564 | axios 实例、拦截器、token 刷新 |
| `src/services/api.ts` | 481 | API 端点定义 |
| `src/constants/index.ts` | 318 | 常量（路由、端点、状态码） |
| `src/store/slices/productSlice.ts` | 351 | 商品状态管理 |
| `src/store/slices/cartSlice.ts` | 381 | 购物车状态管理 |
| `src/store/slices/orderSlice.ts` | 282 | 订单状态管理 |
| `src/store/slices/authSlice.ts` | 270 | 认证状态管理 |
| `src/app/page.tsx` | 430 | 首页 |
| `src/app/products/page.tsx` | 578 | 商品列表 |
| `src/app/products/[id]/page.tsx` | 859 | 商品详情 |
| `src/app/orders/page.tsx` | 198 | 订单列表 ❌ |
| `src/app/checkout/page.tsx` | 660 | 结算页 |
| `src/app/forgot-password/page.tsx` | 421 | 忘记密码 ❌ |
| `next.config.ts` | ~120 | Next.js 配置 |

### B. API 端点对照表

| 功能 | 后端端点 | 前端方法 | 状态 |
|------|---------|---------|------|
| 用户登录 | `POST /api/v1/auth/login` | `authAPI.login()` | ✅ |
| 用户注册 | `POST /api/v1/auth/register` | `authAPI.register()` | ✅ |
| 发送重置码 | `POST /api/v1/auth/send-reset-code` | `authAPI.sendResetCode()` | ❌ |
| 重置密码 | `POST /api/v1/auth/reset-password` | `authAPI.resetPassword()` | ❌ |
| 检查用户名 | `GET /api/v1/auth/check-username` | `authAPI.checkUsernameAvailable()` | ❌ |
| 检查邮箱 | `GET /api/v1/auth/check-email` | `authAPI.checkEmailAvailable()` | ❌ |
| 获取商品列表 | `GET /api/v1/products` | `productAPI.getProducts()` | ✅ |
| 获取商品详情 | `GET /api/v1/products/:id` | `productAPI.getProductDetail()` | ✅ |
| 添加到购物车 | `POST /api/v1/cart/items` | `cartAPI.addItem()` | ✅ |
| 获取购物车 | `GET /api/v1/cart` | `cartAPI.getCart()` | ✅ |
| 创建订单 | `POST /api/v1/orders` | `orderAPI.createOrder()` | ✅ |
| 获取订单列表 | `GET /api/v1/orders` | `orderAPI.getOrders()` | ❌ |
| 获取订单详情 | `GET /api/v1/orders/:id` | `orderAPI.getOrderDetail()` | ✅ |
| 创建支付 | `POST /api/v1/payments` | `paymentAPI.create()` | ❌ |
| 查询支付状态 | `GET /api/v1/payments/:id` | `paymentAPI.query()` | ❌ |

### C. 环境配置检查清单

- [ ] `.env.local` 存在且包含 `NEXT_PUBLIC_API_BASE_URL=http://localhost:8081`
- [ ] `.env.production` 存在且包含生产环境 API 地址
- [ ] `.env.example` 端口为 `8081`
- [ ] `next.config.ts` CSP connect-src 包含环境变量
- [ ] `package.json` 无重复脚本键
- [ ] 控制台日志已清理（5 处）
- [ ] 支付路由已创建（3 个页面）
- [ ] 订单列表已接入真实 API
- [ ] 商品页面交互已实现
- [ ] 认证流程已改为真实 API

---

## 📞 后续支持

如有问题或需要澄清，请参考：
- 后端 API 文档：`docs/api.md`
- Redux 状态管理指南：`docs/redux.md`
- 开发环境配置：`README.md`

**报告生成时间**：2025-11-01  
**下一次审查建议**：修复完成后进行集成测试

